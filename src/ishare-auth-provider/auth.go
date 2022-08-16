package main

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

/**
* Name of the files containing the signing key. Needs to be the same as used by the configuration-service.
 */
const keyfile = "key.pem"

/**
* Name of the files containing the certificate chain. Needs to be the same as used by the configuration-service.
 */
const certChainFile = "cert.cer"

/**
* Struct for holding the required auth info.
 */
type AuthInfo struct {
	AuthType         string `json:"authType"`
	IShareIdpAddress string `json:"iShareIdpAddress"`
	RequestGrantType string `json:"requestGrantType"`
	IShareClientID   string `json:"iShareClientId"`
	IShareIdpID      string `json:"iShareIdpId"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

/**
* Struct to contain headers to be returned by the auth provider.
 */
type HeadersList []Header

var errEmptyDomain error = errors.New("empty_domain")
var errEmptyPath error = errors.New("empty_path")
var errNoResponseBody = errors.New("no_response_body")
var errCertDecode = errors.New("cert_decode_failed")

// auth getter interface to improve testability
type AuthGetterInterface interface {
	getAuthInfo(domain string, path string) (authInfo AuthInfo, err error)
	getSigningKey(credentialsFolderPath string) (key *rsa.PrivateKey, err error)
	getCertificate(credentialsFolderPath string) (encodedCert string, err error)
}

type AuthGetter struct{}

func (AuthGetter) getAuthInfo(domain string, path string) (authInfo AuthInfo, err error) {
	return getAuthInformation(domain, path)
}

func (AuthGetter) getSigningKey(credentialsFolderPath string) (key *rsa.PrivateKey, err error) {
	return getSigningKey(credentialsFolderPath)
}

func (AuthGetter) getCertificate(credentialsFolderPath string) (encodedCert string, err error) {
	return getEncodedCertificate(credentialsFolderPath)
}

var authGetter AuthGetterInterface = &AuthGetter{}

/**
* Route implementation for auth retrieval
 */
func getAuth(c *gin.Context) {

	domain := c.Query("domain")
	if domain == "" {
		logger.Warn("Empty domain was requested.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	path := c.Query("path")
	if path == "" {
		logger.Warn("Empty path was requested.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logger.Info("Get auth for " + domain + " - " + path)

	authInfo, err := authGetter.getAuthInfo(domain, path)
	if err != nil {
		logger.Warn("Was not able to retrieve auth-info. ", err)
		c.String(http.StatusBadGateway, "Was not able to retrieve auth info from the config-service.")
		return
	}

	// the files are stored in folders namend by the clientId
	credentialsFolderPath := buildCredentialsFolderPath(authInfo.IShareClientID)

	logger.Info("CredentialsFolderPath: " + credentialsFolderPath)

	randomUuid, err := uuid.NewRandom()

	if err != nil {
		logger.Warn("Was not able to generate a uuid.", err)
		c.String(http.StatusInternalServerError, "Failed to generate a uuid.")
		return
	}

	// prepare token headers
	now := time.Now().Unix()
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"jti": randomUuid.String(),
		"iss": authInfo.IShareClientID,
		"sub": authInfo.IShareClientID,
		"aud": authInfo.IShareIdpID,
		"iat": now,
		"exp": now + 30,
	})

	key, err := authGetter.getSigningKey(credentialsFolderPath)
	if err != nil {
		logger.Warn("Was not able to read the signing key.")
		c.String(http.StatusInternalServerError, "Error reading the signingKey.")
		return
	}
	if key == nil {
		logger.Warn("Was not able to read a valid signing key.")
		c.String(http.StatusInternalServerError, "Error reading the signingKey.")
		return
	}

	cert, err := authGetter.getCertificate(credentialsFolderPath)
	if err != nil {
		logger.Warn("Was not able to read the certificate.")
		c.String(http.StatusInternalServerError, "Error reading the certificateChain.")
		return
	}

	x5cCerts := [1]string{cert}
	jwtToken.Header["x5c"] = x5cCerts

	// sign the token
	signedToken, err := jwtToken.SignedString(key)
	if err != nil {
		logger.Warn("Was not able to sign the jwt.", err)
		c.String(http.StatusInternalServerError, "Error signing the request jwt.")
		return
	}

	// prepare the form-body
	data := url.Values{
		"grant_type":            {authInfo.RequestGrantType},
		"scope":                 {"iSHARE"},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {signedToken},
		"client_id":             {authInfo.IShareClientID},
	}

	// get the token
	resp, err := globalHttpClient.PostForm(authInfo.IShareIdpAddress, data)
	if err != nil {
		logger.Warn("Was not able to get the token from the idp.", err)
		c.String(http.StatusBadGateway, "Was not able to get the token from the idp.")
		return
	}

	if resp.Body == nil {
		logger.Warn("Did not receive a valid body from the idp.")
		c.AbortWithStatus(http.StatusBadGateway)
		return
	}

	// decode and return
	var res map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		logger.Warnf("Was not able to decode idp response. Err: %v", err)
		c.AbortWithStatus(http.StatusBadGateway)
		return
	}

	if res == nil || res["access_token"] == nil {
		logger.Warnf("Did not receive an access token from the idp. Resp: %v", res)
		c.AbortWithStatus(http.StatusBadGateway)
		return
	}

	header := Header{"Authorization", res["access_token"].(string)}
	headersList := HeadersList{header}

	// Ishare tokens are defined to expire after max 30s. Thus, they should be cached for a little less time.
	c.Header("Cache-Control", "max-age=25")
	c.JSON(http.StatusOK, headersList)
}

/**
* Retrieve auth information from the config service
 */
func getAuthInformation(domain string, path string) (authInfo AuthInfo, err error) {

	req, err := http.NewRequest("GET", configurationServiceUrl+"/auth", nil)
	if err != nil {
		logger.Warn("Was not able to build a request. Invalid configuration server url.", err)
	}

	if domain == "" {
		logger.Warn("Did not receive a domain.")
		return authInfo, errEmptyDomain
	}

	if path == "" {
		logger.Warn("Did not receive a path.")
		return authInfo, errEmptyPath
	}

	q := req.URL.Query()
	q.Add("domain", domain)
	q.Add("path", path)
	req.URL.RawQuery = q.Encode()
	resp, err := globalHttpClient.Get(req.URL.String())
	if err != nil {
		logger.Warn("Was not able to get authInfo. Err: ", err)
		return authInfo, err
	}

	if resp.Body == nil {
		logger.Warn("Did not receive an response body.")
		return authInfo, errNoResponseBody
	}

	// decode and return
	err = json.NewDecoder(resp.Body).Decode(&authInfo)
	if err != nil {
		logger.Warn("Was not able to decode the auth response.", err)
		return authInfo, err
	}
	return authInfo, err
}

/**
* Read siging key from local filesystem
 */
func getSigningKey(credentialsFolderPath string) (key *rsa.PrivateKey, err error) {
	// read key file
	priv, err := globalFileAccessor.read(credentialsFolderPath + keyfile)
	if err != nil {
		logger.Warn("Was not able to read the key file. ", err)
		return key, err
	}

	// parse key file
	key, err = jwt.ParseRSAPrivateKeyFromPEM(priv)
	if err != nil {
		logger.Warn("Was not able to parse the key.", err)
		return key, err
	}

	return key, err
}

/**
* Read and encode(base64) certificate from file system
 */
func getEncodedCertificate(credentialsFolderPath string) (encodedCert string, err error) {
	// read certificate file and set it in the token header
	cert, err := globalFileAccessor.read(credentialsFolderPath + certChainFile)
	if err != nil {
		logger.Warn("Was not able to read the certificateChain file.", err)
		return encodedCert, err
	}
	certCer, _ := pem.Decode(cert)
	if certCer == nil {
		logger.Warn("Was not able to decode certificate.")
		return encodedCert, errCertDecode
	}
	encodedCert = base64.StdEncoding.EncodeToString(certCer.Bytes)
	return encodedCert, err
}

/**
* Build the path to the credentials folder for the given domain/path combination. It will include the trailing /
 */
func buildCredentialsFolderPath(authFolder string) string {

	credentialsFolder := credentialsBaseFolder + "/" + authFolder
	if string(credentialsFolder[len(credentialsFolder)-1:]) != "/" {
		return credentialsFolder + "/"
	}
	return credentialsFolder
}

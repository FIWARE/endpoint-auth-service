package main

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
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
* Global var to held the basefolder to the credentials for all domain/path combinations.
 */
var credentialsBaseFolder string

/**
* URL of the configuration service.
 */
var configurationServiceUrl string

/**
* Struct for holding the required auth info.
 */
type AuthInfo struct {
	AuthType          string `json:"authType"`
	IShareIdpAddress  string `json:"iShareIdpAddress"`
	RequestGrantType  string `json:"requestGrantType"`
	IShareClientID    string `json:"iShareClientId"`
	CredentialsFolder string `json:"credentialsFolder"`
	IShareIdpID       string `json:"iShareIdpId"`
}

/**
* Struct to contain headers to be returned by the auth provider.
 */
type HeadersList struct {
	Array []string
}

func main() {

	router := gin.Default()
	router.GET("/auth", getAuth)

	serverPort := os.Getenv("SERVER_PORT")
	configurationServiceUrl = os.Getenv("CONFIGURATION_SERVICE_URL")
	credentialsBaseFolder = os.Getenv("CERTIFICATE_FOLDER")

	if serverPort == "" {
		log.Fatal("No server port was provided.")
	}
	if configurationServiceUrl == "" {
		log.Fatal("No URL for the configuration service was provided.")
	}

	if credentialsBaseFolder == "" {
		log.Fatal("No credentials base folder was provided.")
	}

	router.Run("0.0.0.0:" + serverPort)
}

func getAuthInformation(domain string, path string) (authInfo AuthInfo, err error) {

	req, err := http.NewRequest("GET", configurationServiceUrl+"/auth", nil)

	if err != nil {
		log.Warn("Was not able to build a request. Invalid configuration server url.", err)
	}

	q := req.URL.Query()
	q.Add("domain", domain)
	q.Add("path", path)
	req.URL.RawQuery = q.Encode()
	resp, err := http.Get(req.URL.String())
	if err != nil {
		log.Warn("Was not able to get authInfo.", err)
		return authInfo, err
	}

	// decode and return
	err = json.NewDecoder(resp.Body).Decode(&authInfo)
	if err != nil {
		log.Warn("Was not able to decode that auth response.", err)
		return authInfo, err
	}
	return authInfo, nil
}

func getAuth(c *gin.Context) {

	domain := c.Query("domain")
	path := c.Query("path")

	authInfo, err := getAuthInformation(domain, path)
	if err != nil {
		log.Warn("Was not able to retrieve auth-info.", err)
		c.String(http.StatusBadGateway, "Was not able to retrieve auth info from the config-service.")
		return
	}

	credentialsFolderPath := buildCredentialsFolderPath(authInfo.CredentialsFolder)

	var randomUuid uuid.UUID

	randomUuid, err = uuid.NewRandom()

	if err != nil {
		log.Warn("Was not able to generate a uuid.", err)
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

	// parse key file
	key, err := getSigningKey(credentialsFolderPath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error reading the signingKey.")
		return
	}

	cert, err := getEncodedCertificate(credentialsFolderPath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error reading the certificateChain.")
		return
	}

	x5cCerts := [1]string{cert}
	jwtToken.Header["x5c"] = x5cCerts

	// sign the token
	signedToken, err := jwtToken.SignedString(key)
	if err != nil {
		log.Warn("Was not able to sign the jwt.", err)
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
	resp, err := http.PostForm(authInfo.IShareIdpAddress, data)
	if err != nil {
		log.Warn("Was not able to get the token from the idp.", err)
		c.String(http.StatusBadGateway, "Was not able to get the token from the idp.")
		return
	}

	// decode and return
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	b, _ := io.ReadAll(resp.Body)

	fmt.Println(string(b))

	headersList := &HeadersList{Array: []string{"Authorization", res["access_token"].(string)}}
	encjsonHeaders, err := json.Marshal(headersList)
	if err != nil {
		log.Warn("Was not able to build a headerList to return.", err)
		c.String(http.StatusInternalServerError, "Error building the response.")
		return
	}
	c.JSON(http.StatusOK, encjsonHeaders)
}

func getSigningKey(credentialsFolderPath string) (key *rsa.PrivateKey, err error) {
	// read key file
	priv, err := ioutil.ReadFile(credentialsFolderPath + keyfile)
	if err != nil {
		log.Warn("Was not able to read the key file.", err)
		return key, err
	}

	// parse key file
	key, err = jwt.ParseRSAPrivateKeyFromPEM(priv)
	if err != nil {
		log.Warn("Was not able to parse the key.", err)
		return key, err
	}

	return key, err
}

func getEncodedCertificate(credentialsFolderPath string) (encodedCert string, err error) {
	// read certificate file and set it in the token header
	cert, err := ioutil.ReadFile(credentialsFolderPath + certChainFile)
	if err != nil {
		log.Warn("Was not able to read the certificateChain file.", err)
		return encodedCert, err
	}
	certCer, _ := pem.Decode(cert)
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

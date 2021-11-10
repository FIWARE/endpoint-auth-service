package main

import (
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"os"

	"github.com/gin-gonic/gin"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func main() {
	router := gin.Default()
	router.GET("/token", getToken)
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		log.Fatal("No server port was provided")
	}
	router.Run("0.0.0.0:" + serverPort)
}

func getToken(c *gin.Context) {

	var randomUuid uuid.UUID
	var err error
	randomUuid, err = uuid.NewRandom()

	idpURL := os.Getenv("IDP_URL")
	ipdId := os.Getenv("IDP_ID")
	clientId := os.Getenv("CLIENT_ID")

	if err != nil {
		log.Fatal(err)
	}
	// prepare token headers
	now := time.Now().Unix()
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"jti": randomUuid.String(),
		"iss": clientId,
		"sub": clientId,
		"aud": ipdId,
		"iat": now,
		"exp": now + 30,
	})

	// read key file
	priv, err := ioutil.ReadFile("./certs/token.pem")
	if err != nil {
		log.Fatal(err)
	}

	// parse key file
	key, err := jwt.ParseRSAPrivateKeyFromPEM(priv)
	if err != nil {
		log.Fatal(err)
	}

	// read certificate file and set it in the token header
	cert, err := ioutil.ReadFile("./certs/token.cer")
	if err != nil {
		log.Fatal(err)
	}
	certCer, _ := pem.Decode(cert)
	encodedCert := base64.StdEncoding.EncodeToString(certCer.Bytes)
	x5cCerts := [1]string{encodedCert}
	jwtToken.Header["x5c"] = x5cCerts

	// sign the token
	signedToken, err := jwtToken.SignedString(key)
	if err != nil {
		log.Fatal(err)
	}
	// prepare the form-body
	data := url.Values{
		"grant_type":            {"client_credentials"},
		"scope":                 {"iSHARE"},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {signedToken},
		"client_id":             {clientId},
	}

	// get the token
	resp, err := http.PostForm(idpURL, data)
	if err != nil {
		log.Fatal(err)
	}

	// decode and return
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	c.String(http.StatusOK, res["access_token"].(string))
}

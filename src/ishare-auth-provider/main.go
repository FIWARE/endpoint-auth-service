package main

import (
	"os"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

/**
* Global var to held the basefolder to the credentials for all domain/path combinations.
 */
var credentialsBaseFolder string

/**
 * URL of the configuration service.
 */
var configurationServiceUrl string

/**
* Startup method to run the gin-server.
 */
func main() {

	router := gin.Default()
	// auth api
	router.GET("/ISHARE/auth", getAuth)

	// credentials management api
	router.GET("/credentials", getCredentialsList)
	router.DELETE("/credentials/:clientId", deleteCredentials)
	router.POST("/credentials/:clientId", postCredentials)
	router.PUT("/credentials/:clientId/certificateChain", putCertificateChain)
	router.PUT("/credentials/:clientId/signingKey", putSigningKey)

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

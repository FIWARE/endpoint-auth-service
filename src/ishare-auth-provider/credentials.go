package main

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type CredentialsType int

const (
	certificateChain CredentialsType = iota
	signingKey       CredentialsType = iota
)

type Credentials struct {
	CertificateChain string `json:"certificateChain"`
	SigningKey       string `json:"signingKey"`
}

func getCredentialsList(c *gin.Context) {
	folders, err := ioutil.ReadDir(credentialsBaseFolder)
	if err != nil {
		log.Warn("Was not able to read credentials folder.", err)
		c.String(http.StatusInternalServerError, "Was not able to read credentials folder.")
		return
	}

	credentialsList := []string{}

	for _, folder := range folders {
		if folder.IsDir() {
			credentialsList = append(credentialsList, folder.Name())
		}
	}

	c.JSON(http.StatusOK, credentialsList)
}

func postCredentials(c *gin.Context) {
	c.SetAccepted("application/json")
	var credentials Credentials
	c.BindJSON(&credentials)
	clientId := c.Param("clientId")

	// the files are stored in folders namend by the clientId
	credentialsFolderPath := buildCredentialsFolderPath(clientId)

	// on post, we dont allow override
	if _, err := os.Stat(credentialsFolderPath); err == nil {
		log.Warn("Credentials for " + clientId + " already exist.")
		c.String(http.StatusConflict, "Credentials for the requested client already exist.")
		return
	}

	err := os.MkdirAll(credentialsFolderPath, os.ModePerm)
	if err != nil {
		log.Warn("Was not able to create folder: "+credentialsFolderPath, err)
		c.String(http.StatusInternalServerError, "Was not able to store the credentials.")
		return
	}

	err = ioutil.WriteFile(credentialsFolderPath+keyfile, []byte(credentials.SigningKey), 0666)
	if err != nil {
		log.Warn("Was not able to store signingKey for: "+clientId, err)
		c.String(http.StatusInternalServerError, "Was not able to store the credentials.")
		os.RemoveAll(credentialsFolderPath)
		return
	}

	err = ioutil.WriteFile(credentialsFolderPath+certChainFile, []byte(credentials.CertificateChain), 0666)
	if err != nil {
		log.Warn("Was not able to store certificate for: "+clientId, err)
		c.String(http.StatusInternalServerError, "Was not able to store the credentials.")
		os.RemoveAll(credentialsFolderPath)
		return
	}

	c.Status(http.StatusCreated)
}

func putCertificateChain(c *gin.Context) {
	storeCredential(c, certificateChain)
}

func putSigningKey(c *gin.Context) {
	storeCredential(c, signingKey)
}

func deleteCredentials(c *gin.Context) {
	clientId := c.Param("clientId")
	// the files are stored in folders namend by the clientId
	credentialsFolderPath := buildCredentialsFolderPath(clientId)

	_, err := os.Stat(credentialsFolderPath)

	if errors.Is(err, os.ErrNotExist) {
		log.Warn("No credentials for "+clientId+" exist.", err)
		c.String(http.StatusNotFound, "No such client exists.")
		return
	}

	err = os.RemoveAll(credentialsFolderPath)
	if err != nil {
		log.Warn("Was not able to delete the credentials for: "+clientId, err)
		c.String(http.StatusInternalServerError, "Was not able to delete credentials.")
		return
	}
	c.Status(http.StatusNoContent)
}

func storeCredential(c *gin.Context, credentialsType CredentialsType) {
	c.SetAccepted("text/plain")
	clientId := c.Param("clientId")

	credential, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Warn("Was not able to read the request body.", err)
		c.String(http.StatusBadRequest, "Was not able to read the body.")
		return
	}

	// the files are stored in folders namend by the clientId
	credentialsFolderPath := buildCredentialsFolderPath(clientId)

	_, err = os.Stat(credentialsFolderPath)

	if errors.Is(err, os.ErrNotExist) {
		log.Warn("No credentials for "+clientId+" exist.", err)
		c.String(http.StatusNotFound, "No such client exists.")
		return
	}

	var filePath string
	var errorMsg string
	if credentialsType == certificateChain {
		filePath = credentialsFolderPath + certChainFile
		errorMsg = "certrificate"
	} else {
		filePath = credentialsFolderPath + keyfile
		errorMsg = "signingKey"
	}
	err = ioutil.WriteFile(filePath, []byte(credential), 0666)
	if err != nil {
		log.Warn("Was not able to store "+errorMsg+" for: "+clientId, err)
		c.String(http.StatusInternalServerError, "Was not able to store the "+errorMsg+".")
		return
	}
	c.Status(http.StatusNoContent)
}

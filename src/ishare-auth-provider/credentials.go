package main

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
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

// route implementations
func getCredentialsList(c *gin.Context) {

	folders, err := globalFolderAccessor.get(credentialsBaseFolder)

	if err != nil {
		logger.Warn("Was not able to read credentials folder.", err)
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
	err := c.BindJSON(&credentials)
	if err != nil {
		logger.Warn("Was not able to read credentials to json.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	clientId := c.Param("clientId")
	if clientId == "" {
		logger.Warn("No clientId present.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// the files are stored in folders namend by the clientId
	credentialsFolderPath := buildCredentialsFolderPath(clientId)

	// on post, we dont allow override
	if _, err := diskFs.Stat(credentialsFolderPath); err == nil {
		logger.Warn("Credentials for " + clientId + " already exist.")
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	err = diskFs.MkdirAll(credentialsFolderPath, os.ModePerm)
	if err != nil {
		logger.Warn("Was not able to create folder: "+credentialsFolderPath, err)
		c.String(http.StatusInternalServerError, "Was not able to store the credentials.")
		return
	}

	err = globalFileAccessor.write(credentialsFolderPath+keyfile, []byte(credentials.SigningKey), 0666)
	if err != nil {
		logger.Warn("Was not able to store signingKey for: "+clientId, err)
		c.String(http.StatusInternalServerError, "Was not able to store the credentials.")
		diskFs.RemoveAll(credentialsFolderPath)
		return
	}

	err = globalFileAccessor.write(credentialsFolderPath+certChainFile, []byte(credentials.CertificateChain), 0666)
	if err != nil {
		logger.Warn("Was not able to store certificate for: "+clientId, err)
		c.String(http.StatusInternalServerError, "Was not able to store the credentials.")
		diskFs.RemoveAll(credentialsFolderPath)
		return
	}

	c.AbortWithStatus(http.StatusCreated)
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

	_, err := diskFs.Stat(credentialsFolderPath)

	if errors.Is(err, os.ErrNotExist) {
		logger.Warn("No credentials for "+clientId+" exist.", err)
		c.String(http.StatusNotFound, "No such client exists.")
		return
	}

	err = diskFs.RemoveAll(credentialsFolderPath)
	if err != nil {
		logger.Warn("Was not able to delete the credentials for: "+clientId, err)
		c.String(http.StatusInternalServerError, "Was not able to delete credentials.")
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}

func storeCredential(c *gin.Context, credentialsType CredentialsType) {
	c.SetAccepted("text/plain")
	clientId := c.Param("clientId")
	if clientId == "" {
		logger.Warn("No clientId present.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	credential, err := io.ReadAll(c.Request.Body)
	if err != nil || bytes.Equal([]byte(credential), []byte{}) {
		logger.Warn("Was not able to read the request body.", err)
		c.String(http.StatusBadRequest, "Was not able to read the body.")
		return
	}

	// the files are stored in folders namend by the clientId
	credentialsFolderPath := buildCredentialsFolderPath(clientId)

	_, err = diskFs.Stat(credentialsFolderPath)

	if errors.Is(err, os.ErrNotExist) {
		logger.Warn("No credentials for "+clientId+" exist.", err)
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
	err = globalFileAccessor.write(filePath, []byte(credential), 0666)
	if err != nil {
		logger.Warn("Was not able to store "+errorMsg+" for: "+clientId, err)
		c.String(http.StatusInternalServerError, "Was not able to store the "+errorMsg+".")
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}

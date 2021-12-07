package main

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
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

/**
* Interfaces for accessing the file system.
 */
type FolderContentGetter func(path string) (folders []fs.FileInfo, err error)
type Folder struct {
	get FolderContentGetter
}

type FileWriter func(path string, content []byte, fileMode fs.FileMode) (err error)
type File struct {
	write FileWriter
}

type fileSystem interface {
	Open(name string) (file, error)
	Stat(name string) (os.FileInfo, error)
	MkdirAll(path string, perm fs.FileMode) error
	RemoveAll(path string) error
}

type file interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	Stat() (os.FileInfo, error)
}

// osFS implements fileSystem using the local disk.
type osFS struct{}

func (osFS) Open(name string) (file, error)               { return os.Open(name) }
func (osFS) Stat(name string) (os.FileInfo, error)        { return os.Stat(name) }
func (osFS) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }
func (osFS) RemoveAll(path string) error                  { return os.RemoveAll(path) }

func getFolderContent(path string) (folders []fs.FileInfo, err error) {
	return ioutil.ReadDir(path)
}

func writeFile(path string, content []byte, fileMode fs.FileMode) (err error) {
	return ioutil.WriteFile(path, content, fileMode)
}

// route mappings
func getCredentialsListRoute(c *gin.Context) {
	getCredentialsList(c, &Folder{getFolderContent})
}

func postCredentialsRoute(c *gin.Context) {
	postCredentials(c, &File{writeFile}, &osFS{})
}

// route implementations
func getCredentialsList(c *gin.Context, f *Folder) {

	folders, err := f.get(credentialsBaseFolder)

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

func postCredentials(c *gin.Context, f *File, fs fileSystem) {
	c.SetAccepted("application/json")
	var credentials Credentials
	err := c.BindJSON(&credentials)
	if err != nil {
		log.Warn("Was not able to read credentials to json.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	clientId := c.Param("clientId")
	if clientId == "" {
		log.Warn("No clientId present.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// the files are stored in folders namend by the clientId
	credentialsFolderPath := buildCredentialsFolderPath(clientId)

	// on post, we dont allow override
	if _, err := fs.Stat(credentialsFolderPath); err == nil {
		log.Warn("Credentials for " + clientId + " already exist.")
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	err = fs.MkdirAll(credentialsFolderPath, os.ModePerm)
	if err != nil {
		log.Warn("Was not able to create folder: "+credentialsFolderPath, err)
		c.String(http.StatusInternalServerError, "Was not able to store the credentials.")
		return
	}

	err = f.write(credentialsFolderPath+keyfile, []byte(credentials.SigningKey), 0666)
	if err != nil {
		log.Warn("Was not able to store signingKey for: "+clientId, err)
		c.String(http.StatusInternalServerError, "Was not able to store the credentials.")
		fs.RemoveAll(credentialsFolderPath)
		return
	}

	err = f.write(credentialsFolderPath+certChainFile, []byte(credentials.CertificateChain), 0666)
	if err != nil {
		log.Warn("Was not able to store certificate for: "+clientId, err)
		c.String(http.StatusInternalServerError, "Was not able to store the credentials.")
		fs.RemoveAll(credentialsFolderPath)
		return
	}

	c.AbortWithStatus(http.StatusCreated)
}

func putCertificateChain(c *gin.Context) {
	storeCredential(c, certificateChain, &File{writeFile}, &osFS{})
}

func putSigningKey(c *gin.Context) {
	storeCredential(c, signingKey, &File{writeFile}, &osFS{})
}

func deleteCredentialsRoute(c *gin.Context) {
	deleteCredentials(c, &osFS{})
}

func deleteCredentials(c *gin.Context, fs fileSystem) {
	clientId := c.Param("clientId")
	// the files are stored in folders namend by the clientId
	credentialsFolderPath := buildCredentialsFolderPath(clientId)

	_, err := fs.Stat(credentialsFolderPath)

	if errors.Is(err, os.ErrNotExist) {
		log.Warn("No credentials for "+clientId+" exist.", err)
		c.String(http.StatusNotFound, "No such client exists.")
		return
	}

	err = fs.RemoveAll(credentialsFolderPath)
	if err != nil {
		log.Warn("Was not able to delete the credentials for: "+clientId, err)
		c.String(http.StatusInternalServerError, "Was not able to delete credentials.")
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}

func storeCredential(c *gin.Context, credentialsType CredentialsType, f *File, fs fileSystem) {
	c.SetAccepted("text/plain")
	clientId := c.Param("clientId")
	if clientId == "" {
		log.Warn("No clientId present.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	credential, err := io.ReadAll(c.Request.Body)
	if err != nil || bytes.Equal([]byte(credential), []byte{}) {
		log.Warn("Was not able to read the request body.", err)
		c.String(http.StatusBadRequest, "Was not able to read the body.")
		return
	}

	// the files are stored in folders namend by the clientId
	credentialsFolderPath := buildCredentialsFolderPath(clientId)

	_, err = fs.Stat(credentialsFolderPath)

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
	err = f.write(filePath, []byte(credential), 0666)
	if err != nil {
		log.Warn("Was not able to store "+errorMsg+" for: "+clientId, err)
		c.String(http.StatusInternalServerError, "Was not able to store the "+errorMsg+".")
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}

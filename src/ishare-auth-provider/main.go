package main

import (
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

/**
* Global logger
 */
var logger = logrus.New()

/**
* Global var to held the basefolder to the credentials for all domain/path combinations.
 */
var credentialsBaseFolder string

/**
 * URL of the configuration service.
 */
var configurationServiceUrl string

/**
* Global filesystem accessor
 */
var diskFs fileSystem = &osFS{}

/**
* Global folder accessor
 */
var globalFolderAccessor folderAccessor = folderAccessor{getFolderContent}

/**
* Global file accessor
 */
var globalFileAccessor fileAccessor = fileAccessor{writeFile, readFile}

/**
* Global http client
 */
var globalHttpClient httpClient = &http.Client{}

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
	enableJsonLogging, err := strconv.ParseBool(os.Getenv("JSON_LOGGING_ENABLED"))

	if err != nil {
		logger.Warnf("Json log env-var not readable. Use default logging. %v", err)
		enableJsonLogging = false
	}

	if enableJsonLogging {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	if serverPort == "" {
		logger.Fatal("No server port was provided.")
	}
	if configurationServiceUrl == "" {
		logger.Fatal("No URL for the configuration service was provided.")
	}

	if credentialsBaseFolder == "" {
		logger.Fatal("No credentials base folder was provided.")
	}

	logger.Info("Start router at " + serverPort)
	router.Run("0.0.0.0:" + serverPort)
}

// Interfaces for accessing the file system.
// Introduced to improve testability

type folderContentGetter func(path string) (folders []fs.FileInfo, err error)
type folderAccessor struct {
	get folderContentGetter
}

type fileWriter func(path string, content []byte, fileMode fs.FileMode) (err error)
type fileReader func(filename string) (content []byte, err error)

type fileAccessor struct {
	write fileWriter
	read  fileReader
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

func readFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

// Interface to the http-client
type httpClient interface {
	Get(url string) (*http.Response, error)
	PostForm(url string, data url.Values) (*http.Response, error)
}

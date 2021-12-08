package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var contentMock map[string][]byte
var mockReadErr error

func mock_noop_write(path string, content []byte, fileMode fs.FileMode) (err error) { return err }

func mock_read_content(filename string) (content []byte, err error) {
	if mockReadErr != nil {
		return content, mockReadErr
	}
	return contentMock[filename], err
}

type mockHttpClient struct {
	mockGetResponse  *http.Response
	mockPostResponse *http.Response
	mockGetError     error
	mockPostError    error
}

func (mhc mockHttpClient) Get(url string) (response *http.Response, err error) {
	return mhc.mockGetResponse, mhc.mockGetError
}

func (mhc mockHttpClient) PostForm(url string, data url.Values) (response *http.Response, err error) {
	return mhc.mockPostResponse, mhc.mockPostError
}

type mockAuthGetter struct {
	mockAuthInfo AuthInfo
	infoGetError error
	mockKey      *rsa.PrivateKey
	keyGetError  error
	mockCert     string
	certGetError error
}

func (mag mockAuthGetter) getAuthInfo(domain string, path string) (authInfo AuthInfo, err error) {
	return mag.mockAuthInfo, mag.infoGetError
}
func (mag mockAuthGetter) getSigningKey(credentialsFolderPath string) (key *rsa.PrivateKey, err error) {
	return mag.mockKey, mag.keyGetError
}
func (mag mockAuthGetter) getCertificate(credentialsFolderPath string) (encodedCert string, err error) {
	return mag.mockCert, mag.certGetError
}

func TestGetEncodedCertificate(t *testing.T) {
	testFolder := "myFolder/"
	type test struct {
		testName     string
		testCert     []byte
		mockError    error
		expectedCert string
		expectError  error
	}

	globalFileAccessor = fileAccessor{mock_noop_write, mock_read_content}

	notReadableError := errors.New("no_readable_cert")

	tests := []test{
		{testName: "Successfully retrive cert", testCert: getPemEncoded("myCert"), expectedCert: "bXlDZXJ0"},
		{testName: "Cert not pem encoded", testCert: []byte("myCert"), expectError: errCertDecode},
		{testName: "Cert not readable", mockError: notReadableError, expectError: notReadableError},
	}

	for _, tc := range tests {
		log.Info("TestGetEncodedCertificate +++++++++++++++++++++ Running test: " + tc.testName)
		contentMock = map[string][]byte{testFolder + certChainFile: tc.testCert}
		mockReadErr = tc.mockError

		cert, err := getEncodedCertificate(testFolder)

		if tc.expectedCert != cert {
			t.Errorf(tc.testName + ": Did not receive the expected cert. Exoected: " + tc.expectedCert + " Actual: " + cert)
		}
		if !errors.Is(err, tc.expectError) {
			t.Errorf(tc.testName + ": Did not receive the expected error. Exoected: " + fmt.Sprint(tc.expectError) + " Actual: " + fmt.Sprint(err))
		}
	}
}

func TestGetSigningKey(t *testing.T) {

	testFolder := "myFolder/"

	type test struct {
		testName           string
		testKey            []byte
		mockError          error
		expectKey          bool
		expectError        error
		expectGenericError bool
	}

	globalFileAccessor = fileAccessor{mock_noop_write, mock_read_content}

	notReadableError := errors.New("no_readable_keyfile")

	tests := []test{
		{testName: "Successfully retrive key", testKey: getValidKeyBytes(), expectKey: true},
		{testName: "Keyfile not readable", mockError: notReadableError, expectError: notReadableError, expectKey: false},
		{testName: "Keyfile not parseable", testKey: []byte("something invalid"), expectGenericError: true, expectKey: false},
	}

	for _, tc := range tests {
		log.Info("TestGetSigningKey +++++++++++++++++++++ Running test: " + tc.testName)
		contentMock = map[string][]byte{testFolder + keyfile: tc.testKey}
		mockReadErr = tc.mockError

		key, err := getSigningKey(testFolder)

		if tc.expectKey && key == nil {
			t.Errorf(tc.testName + ": Was not able to retrieve the key as expected.")
		}

		if !tc.expectGenericError && tc.expectError != nil {
			if !errors.Is(err, tc.expectError) {
				t.Errorf(tc.testName + ": Did not receive the expected error.")
			}
		}
		if tc.expectGenericError && err == nil {
			t.Errorf(tc.testName + ": Did not receive the expected generic error.")
		}
	}
}

func TestGetAuthInformation(t *testing.T) {

	successfullResponse := &http.Response{Body: io.NopCloser(strings.NewReader(
		"{\"authType\":\"iShare\"," +
			"\"iShareIdpAddress\": \"http://my-idp\"," +
			"\"requestGrantType\": \"client_credentials\"," +
			"\"iShareClientId\": \"clientId\"," +
			"\"iShareIdpId\": \"idpId\"}"))}

	matchingAuthInfo := AuthInfo{AuthType: "iShare", IShareIdpAddress: "http://my-idp", RequestGrantType: "client_credentials", IShareClientID: "clientId", IShareIdpID: "idpId"}

	type test struct {
		testName           string
		testDomain         string
		testPath           string
		mockResponse       *http.Response
		mockError          error
		expectedInfo       AuthInfo
		expectedError      error
		expectGenericError bool
	}

	mockError := errors.New("something_went_wrong")
	tests := []test{
		{testName: "Successfull retrieval", testDomain: "https://test.domain", testPath: "/auth", mockResponse: successfullResponse, expectedInfo: matchingAuthInfo},
		{testName: "Empty domain error", testDomain: "", testPath: "/path", expectedError: errEmptyDomain},
		{testName: "Empty path error", testDomain: "https://test.domain", testPath: "", expectedError: errEmptyPath},
		{testName: "Error from config service", testDomain: "https://test.domain", testPath: "/auth", mockError: mockError, expectedError: mockError},
		{testName: "Error from config service - invalid json", testDomain: "https://test.domain", testPath: "/auth", mockResponse: &http.Response{Body: io.NopCloser(strings.NewReader("no-json"))}, expectGenericError: true},
		{testName: "Error from config service - empty body", testDomain: "https://test.domain", testPath: "/auth", mockResponse: &http.Response{}, expectedError: errNoResponseBody},
	}

	for _, tc := range tests {
		log.Info("TestGetAuthInformation +++++++++++++++++++++ Running test: " + tc.testName)
		globalHttpClient = &mockHttpClient{mockGetResponse: tc.mockResponse, mockGetError: tc.mockError}

		authInfo, err := getAuthInformation(tc.testDomain, tc.testPath)

		if authInfo != tc.expectedInfo {
			t.Errorf(tc.testName + ": Expected auth info: " + fmt.Sprint(tc.expectedInfo) + " but got: " + fmt.Sprint(authInfo))
		}
		if err == nil && tc.expectGenericError {
			t.Errorf(tc.testName + ": An error should have been thrown")
		}
		if !tc.expectGenericError && err != nil && !errors.Is(err, tc.expectedError) {
			t.Errorf(tc.testName + ": Expected error: " + fmt.Sprint(tc.expectedError) + " but got: " + fmt.Sprint(err))
		}
	}

}

func TestGetAuthRoute(t *testing.T) {
	var ginContext *gin.Context
	var recorder *httptest.ResponseRecorder

	type test struct {
		testName          string
		testDomain        string
		testPath          string
		mockKey           *rsa.PrivateKey
		mockCert          string
		mockAuthInfo      AuthInfo
		mockAuthInfoError error
		mockKeyReadError  error
		mockCertReadError error
		mockIdpResponse   *http.Response
		mockIdpError      error
		expectedCode      int
		expectedHeader    string
	}

	validKey, _ := getValidKey()
	validAuthInfo := AuthInfo{AuthType: "iShare", IShareIdpAddress: "http://ishare.de", RequestGrantType: "client_credentials", IShareClientID: "clientId", IShareIdpID: "idpId"}
	accesTokenResponse := &http.Response{Body: io.NopCloser(strings.NewReader("{\"access_token\":\"myToken\"}"))}

	tests := []test{
		{testName: "Successful auth retrieval", testDomain: "test.domain", testPath: "/", mockIdpResponse: accesTokenResponse, mockKey: validKey, mockCert: "cert", mockAuthInfo: validAuthInfo, expectedHeader: "myToken", expectedCode: 200},
		{testName: "502: No body returned from idp", testDomain: "test.domain", testPath: "/", mockIdpResponse: &http.Response{}, mockKey: validKey, mockCert: "cert", mockAuthInfo: validAuthInfo, expectedCode: 502},
		{testName: "502: Invalid body returned from idp", testDomain: "test.domain", testPath: "/", mockIdpResponse: &http.Response{Body: io.NopCloser(strings.NewReader("myToken"))}, mockKey: validKey, mockCert: "cert", mockAuthInfo: validAuthInfo, expectedCode: 502},
		{testName: "502: Json body withou token returned from idp", testDomain: "test.domain", testPath: "/", mockIdpResponse: &http.Response{Body: io.NopCloser(strings.NewReader("{\"valid\":\"json\"}"))}, mockKey: validKey, mockCert: "cert", mockAuthInfo: validAuthInfo, expectedCode: 502},
		{testName: "502: Error on config-service", testDomain: "test.domain", testPath: "/", mockAuthInfoError: errors.New("service_error"), expectedCode: 502},
		{testName: "500: Error reading signing key", testDomain: "test.domain", testPath: "/", mockKeyReadError: errors.New("read_error"), expectedCode: 500},
		{testName: "500: Error reading certificate", testDomain: "test.domain", testPath: "/", mockCertReadError: errors.New("read_error"), expectedCode: 500},
		{testName: "500: Signing error - nil key", testDomain: "test.domain", testPath: "/", mockIdpResponse: accesTokenResponse, mockCert: "cert", mockAuthInfo: validAuthInfo, expectedCode: 500},
		{testName: "400: Empty path received", testDomain: "test.domain", testPath: "", expectedCode: 400},
	}

	for _, tc := range tests {
		log.Info("TestGetAuth +++++++++++++++++++++ Running test: " + tc.testName)
		recorder = httptest.NewRecorder()
		ginContext, _ = gin.CreateTestContext(recorder)
		ginContext.Request, _ = http.NewRequest(http.MethodGet, "http://auth.domain/auth?domain="+tc.testDomain+"&path="+tc.testPath, nil)

		globalHttpClient = &mockHttpClient{mockPostResponse: tc.mockIdpResponse, mockPostError: tc.mockIdpError}
		authGetter = &mockAuthGetter{mockKey: tc.mockKey, mockCert: tc.mockCert, mockAuthInfo: tc.mockAuthInfo, infoGetError: tc.mockAuthInfoError, keyGetError: tc.mockKeyReadError, certGetError: tc.mockCertReadError}

		getAuth(ginContext)

		if recorder.Code != tc.expectedCode {
			t.Errorf(tc.testName + ": Did not receive the correct code. Expected: " + fmt.Sprint(tc.expectedCode) + " Actual: " + fmt.Sprint(recorder.Code))
		}

		if tc.expectedHeader != "" && recorder.Body == nil {
			t.Errorf(tc.testName + ": Did receive a nil body.")
		}

		if tc.expectedHeader != "" && recorder.Body != nil {
			var result HeadersList
			if json.NewDecoder(recorder.Body).Decode(&result) != nil {
				t.Errorf(tc.testName + ": Did receive invalid body.")
				continue
			}
			if len(result) != 1 {
				t.Errorf(tc.testName + ": Did receive wrong number of headers. Was: " + fmt.Sprint(len(result)))
				continue
			}
			if result[0].Name != "Authorization" {
				t.Errorf(tc.testName + ": Did receive wrong headers. Was: " + fmt.Sprint(result[0]))
				continue
			}
			if result[0].Value != tc.expectedHeader {
				t.Errorf(tc.testName + ": Did receive wrong headers. Expected: " + tc.expectedHeader + " Actual: " + result[0].Value)
			}
		}
	}
}

func getValidKeyBytes() []byte {
	privateKey, _ := getValidKey()
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	return pem.EncodeToMemory(privateKeyBlock)
}

func getValidKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

func getPemEncoded(cert string) []byte {
	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: []byte(cert),
	}
	return pem.EncodeToMemory(certBlock)
}

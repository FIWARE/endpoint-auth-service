package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var mockFolders []fs.FileInfo
var mockError error
var fileWriteRecord []FileWriteRecord
var filesDeleted bool
var pathErrors map[string]error

type FileWriteRecord struct {
	content string
	path    string
}

type MockedFileInfo struct {
	FileName    string
	IsDirectory bool
}

func (mfi MockedFileInfo) Name() string       { return mfi.FileName }
func (mfi MockedFileInfo) Size() int64        { return int64(8) }
func (mfi MockedFileInfo) Mode() os.FileMode  { return os.ModePerm }
func (mfi MockedFileInfo) ModTime() time.Time { return time.Now() }
func (mfi MockedFileInfo) IsDir() bool        { return mfi.IsDirectory }
func (mfi MockedFileInfo) Sys() interface{}   { return nil }

type mockFS struct {
	mockErrRead         error
	mockErrCreateFolder error
	mockErrDelete       error
	mockFile            *file
	fileInfo            MockedFileInfo
}

func (mfs mockFS) Open(name string) (f file, err error)         { return *mfs.mockFile, mfs.mockErrRead }
func (mfs mockFS) Stat(name string) (fi os.FileInfo, err error) { return mfs.fileInfo, mfs.mockErrRead }
func (mfs mockFS) MkdirAll(path string, perm fs.FileMode) error { return mfs.mockErrCreateFolder }
func (mfs mockFS) RemoveAll(path string) error {
	if mfs.mockErrDelete == nil {
		filesDeleted = true
	}
	return mfs.mockErrDelete
}

func mock_get_folder(path string) (folders []fs.FileInfo, err error) {
	return mockFolders, mockError
}

func mock_path_based_write(path string, content []byte, fileMode fs.FileMode) (err error) {
	log.Info("Store to " + path)
	if pathErrors[path] != nil {
		return pathErrors[path]
	}
	fileWriteRecord = append(fileWriteRecord, FileWriteRecord{string(content), path})
	return err
}

func emptyMockFolders() (folders []fs.FileInfo) {
	return folders
}

func singleMockFolder() (folders []fs.FileInfo) {
	fileInfo := &MockedFileInfo{FileName: "myClient", IsDirectory: true}
	folders = []fs.FileInfo{fileInfo}
	return folders
}

func multipleMockFolders() (folders []fs.FileInfo) {

	folders = []fs.FileInfo{
		&MockedFileInfo{FileName: "myClient1", IsDirectory: true},
		&MockedFileInfo{FileName: "myClient2", IsDirectory: true},
		&MockedFileInfo{FileName: "myClient3", IsDirectory: true},
	}
	return folders
}

func multipleWithFile() (folders []fs.FileInfo) {

	folders = []fs.FileInfo{
		&MockedFileInfo{FileName: "myClient1", IsDirectory: true},
		&MockedFileInfo{FileName: "myClient2", IsDirectory: true},
		&MockedFileInfo{FileName: "myFile", IsDirectory: false},
	}
	return folders
}

func TestGetCredentialsList(t *testing.T) {

	type test struct {
		testName     string
		mockFolders  []fs.FileInfo
		mockError    error
		expectedCode int
		expectedBody string
	}

	tests := []test{
		{"Get empty list", emptyMockFolders(), nil, 200, "[]"},
		{"500: cannot read folder.", nil, errors.New("Cannot read folder"), 500, "Was not able to read credentials folder."},
		{"Get single client.", singleMockFolder(), nil, 200, "[\"myClient\"]"},
		{"Get multiple clients.", multipleMockFolders(), nil, 200, "[\"myClient1\",\"myClient2\",\"myClient3\"]"},
		{"Get multiple clients with file in folder.", multipleWithFile(), nil, 200, "[\"myClient1\",\"myClient2\"]"},
	}

	folderMock := &Folder{mock_get_folder}
	var ginContext *gin.Context
	var recorder *httptest.ResponseRecorder

	credentialsBaseFolder = "test/credentials"
	for _, tc := range tests {
		mockError = tc.mockError
		mockFolders = tc.mockFolders

		recorder = httptest.NewRecorder()

		ginContext, _ = gin.CreateTestContext(recorder)
		getCredentialsList(ginContext, folderMock)

		if recorder.Code != tc.expectedCode {
			t.Fatalf("Expected to get" + fmt.Sprint(tc.expectedCode) + ", but got " + fmt.Sprint(recorder.Code))
		}

		body, _ := ioutil.ReadAll(recorder.Body)
		if string(body) != tc.expectedBody {
			t.Fatalf("Expected list" + tc.expectedBody + " should have been returned, but was " + string(body))
		}
	}
}

func TestPostCredentials(t *testing.T) {

	credentialsBaseFolder = "test/credentials"

	type test struct {
		testName            string
		mockRequestContent  string
		clientId            string
		expectedCode        int
		expectStored        bool
		mockErrRead         error
		mockErrCreateFolder error
		mockErrDelete       error
		mockErrWrite        map[string]error
	}

	expectedKeyFile := FileWriteRecord{"key", "test/credentials/testClient/key.pem"}
	expectedCertFile := FileWriteRecord{"cert", "test/credentials/testClient/cert.cer"}
	reqBody := "{\"certificateChain\":\"cert\",\"signingKey\":\"key\"}"

	tests := []test{
		{testName: "Successfull creation.", mockRequestContent: reqBody, clientId: "testClient", expectedCode: 201, mockErrRead: errors.New("No such folder."), expectStored: true},
		{testName: "No request body.", expectedCode: 400, expectStored: false},
		{testName: "Credentials already exist.", mockRequestContent: reqBody, clientId: "testClient", mockErrRead: nil, expectedCode: 409, expectStored: false},
		{testName: "500: cannot create folder.", mockRequestContent: reqBody, clientId: "testClient", expectedCode: 500, mockErrRead: errors.New("No such folder."), mockErrCreateFolder: errors.New("Cannot create folder."), expectStored: false},
		{testName: "500: cannot store key.", mockRequestContent: reqBody, clientId: "testClient", mockErrWrite: map[string]error{"test/credentials/testClient/key.pem": errors.New("Err")}, mockErrRead: errors.New("No such folder."), expectStored: false, expectedCode: 500},
		{testName: "500: cannot store cert.", mockRequestContent: reqBody, clientId: "testClient", mockErrWrite: map[string]error{"test/credentials/testClient/cert.cer": errors.New("Err")}, mockErrRead: errors.New("No such folder."), expectStored: false, expectedCode: 500},
	}

	var ginContext *gin.Context
	var recorder *httptest.ResponseRecorder
	var fileSystemMock fileSystem
	fileMock := &File{mock_path_based_write}

	for _, tc := range tests {
		log.Info("+++++++++++++++++++++Running test: " + tc.testName)

		pathErrors = tc.mockErrWrite
		fileWriteRecord = []FileWriteRecord{}
		filesDeleted = false
		recorder = httptest.NewRecorder()
		ginContext, _ = gin.CreateTestContext(recorder)
		ginContext.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(tc.mockRequestContent)))
		ginContext.Params = []gin.Param{{Key: "clientId", Value: tc.clientId}}
		fileSystemMock = &mockFS{mockErrRead: tc.mockErrRead, mockErrCreateFolder: tc.mockErrCreateFolder, mockErrDelete: tc.mockErrDelete}

		postCredentials(ginContext, fileMock, fileSystemMock)

		if recorder.Code != tc.expectedCode {
			t.Fatalf("Should have been " + fmt.Sprint(tc.expectedCode) + ", but was " + fmt.Sprint(recorder.Code))
		}

		if tc.expectStored && !contains(fileWriteRecord, expectedKeyFile) {
			log.Warn(fmt.Sprint(fileWriteRecord))
			t.Fatalf("Key was not stored correctly")
		}

		if tc.expectStored && !contains(fileWriteRecord, expectedCertFile) {
			t.Fatalf("Cert was not stored correctly")
		}

		if !tc.expectStored && contains(fileWriteRecord, expectedKeyFile) {
			// rollback check
			if !filesDeleted {
				t.Fatalf("Key should not have been stored.")
			}
		}

		if !tc.expectStored && contains(fileWriteRecord, expectedCertFile) {
			// rollback check
			if !filesDeleted {
				t.Fatalf("Cert should not have been stored.")
			}
		}
	}
}

//used by both put methods, thus only this method has an own test
func TestStoreCredentials(t *testing.T) {
	credentialsBaseFolder = "test/credentials"

	type test struct {
		testName           string
		mockRequestContent string
		clientId           string
		expectedCode       int
		expectStored       bool
		mockErrRead        error
		mockErrWrite       map[string]error
		credentialsType    CredentialsType
	}

	expectedKeyFile := FileWriteRecord{"newKey", "test/credentials/testClient/key.pem"}
	expectedCertFile := FileWriteRecord{"newCert", "test/credentials/testClient/cert.cer"}

	mockKey := "newKey"
	mockCert := "newCert"

	tests := []test{
		{testName: "Update key.", mockRequestContent: mockKey, clientId: "testClient", expectedCode: 204, expectStored: true, mockErrRead: nil, credentialsType: signingKey},
		{testName: "Update cert.", mockRequestContent: mockCert, clientId: "testClient", expectedCode: 204, expectStored: true, mockErrRead: nil, credentialsType: certificateChain},
		{testName: "No body for key.", clientId: "testClient", expectedCode: 400, expectStored: false, credentialsType: signingKey},
		{testName: "No body for cert.", clientId: "testClient", expectedCode: 400, expectStored: false, credentialsType: certificateChain},
		{testName: "No credentials exist for key.", clientId: "testClient", mockRequestContent: mockKey, expectedCode: 404, mockErrRead: fs.ErrNotExist, expectStored: false, credentialsType: signingKey},
		{testName: "No credentials exist for cert.", clientId: "testClient", mockRequestContent: mockCert, expectedCode: 404, mockErrRead: fs.ErrNotExist, expectStored: false, credentialsType: certificateChain},
		{testName: "500: cannot store key", clientId: "testClient", mockRequestContent: mockKey, expectedCode: 500, mockErrWrite: map[string]error{"test/credentials/testClient/key.pem": errors.New("Err")}, expectStored: false, credentialsType: signingKey},
		{testName: "500: cannot store cert", clientId: "testClient", mockRequestContent: mockCert, expectedCode: 500, mockErrWrite: map[string]error{"test/credentials/testClient/cert.cer": errors.New("Err")}, expectStored: false, credentialsType: certificateChain},
	}

	var ginContext *gin.Context
	var recorder *httptest.ResponseRecorder
	var fileSystemMock fileSystem
	fileMock := &File{mock_path_based_write}

	for _, tc := range tests {
		log.Info("+++++++++++++++++++++Running test: " + tc.testName)
		pathErrors = tc.mockErrWrite
		fileWriteRecord = []FileWriteRecord{}
		recorder = httptest.NewRecorder()
		ginContext, _ = gin.CreateTestContext(recorder)
		ginContext.Request, _ = http.NewRequest(http.MethodPut, "/", bytes.NewBuffer([]byte(tc.mockRequestContent)))
		ginContext.Params = []gin.Param{{Key: "clientId", Value: tc.clientId}}
		fileSystemMock = &mockFS{mockErrRead: tc.mockErrRead}

		storeCredential(ginContext, tc.credentialsType, fileMock, fileSystemMock)

		if recorder.Code != tc.expectedCode {
			t.Fatalf("Should have been " + fmt.Sprint(tc.expectedCode) + ", but was " + fmt.Sprint(recorder.Code))
		}

		if tc.credentialsType == signingKey && tc.expectStored && !contains(fileWriteRecord, expectedKeyFile) {
			t.Fatalf("Key was not stored correctly")
		}

		if tc.credentialsType == certificateChain && tc.expectStored && !contains(fileWriteRecord, expectedCertFile) {
			t.Fatalf("Cert was not stored correctly")
		}

		if !tc.expectStored && contains(fileWriteRecord, expectedKeyFile) {
			t.Fatalf("Key should not have been stored.")

		}

		if !tc.expectStored && contains(fileWriteRecord, expectedCertFile) {
			t.Fatalf("Cert should not have been stored.")
		}
	}

}

func TestDeleteCredentials(t *testing.T) {

	credentialsBaseFolder = "test/credentials"

	type test struct {
		testName      string
		clientId      string
		expectedCode  int
		expectDelete  bool
		mockErrRead   error
		mockErrDelete error
	}

	tests := []test{
		{testName: "Delete credentials", clientId: "testClient", expectedCode: 204, expectDelete: true, mockErrRead: nil},
		{testName: "No such credentials", clientId: "testClient", expectedCode: 404, expectDelete: false, mockErrRead: fs.ErrNotExist},
		{testName: "No such credentials", clientId: "testClient", expectedCode: 404, expectDelete: false, mockErrRead: fs.ErrNotExist},
		{testName: "No such credentials", clientId: "testClient", expectedCode: 500, expectDelete: false, mockErrDelete: errors.New("Was not able to delete")},
	}

	var ginContext *gin.Context
	var recorder *httptest.ResponseRecorder
	var fileSystemMock fileSystem

	for _, tc := range tests {
		log.Info("+++++++++++++++++++++Running test: " + tc.testName)
		filesDeleted = false
		fileWriteRecord = []FileWriteRecord{}
		recorder = httptest.NewRecorder()
		ginContext, _ = gin.CreateTestContext(recorder)
		ginContext.Request, _ = http.NewRequest(http.MethodDelete, "/", nil)
		ginContext.Params = []gin.Param{{Key: "clientId", Value: tc.clientId}}
		fileSystemMock = &mockFS{mockErrRead: tc.mockErrRead, mockErrDelete: tc.mockErrDelete}

		deleteCredentials(ginContext, fileSystemMock)

		if recorder.Code != tc.expectedCode {
			t.Fatalf("Should have been " + fmt.Sprint(tc.expectedCode) + ", but was " + fmt.Sprint(recorder.Code))
		}

		if tc.expectDelete != filesDeleted {
			t.Fatalf("The files should have been deleted " + fmt.Sprint(tc.expectDelete))
		}
	}

}

func contains(s []FileWriteRecord, e FileWriteRecord) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

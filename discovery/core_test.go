package discovery

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/pureinsights/pdp-cli/internal/fileutils"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Test_newLabelsClient tests the constructor of newLabelsClient
func Test_newLabelsClient(t *testing.T) {
	c := newClient("http://localhost:12010/v2/v2", "Api Key")
	lc := newLabelsClient(c.client.BaseURL, c.ApiKey)

	assert.Equal(t, c.ApiKey, lc.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/label", lc.client.client.BaseURL)
}

// Test_newSecretsClient tests the constructor of newSecretsClient
func Test_newSecretsClient(t *testing.T) {
	c := newClient("http://localhost:12010/v2/v2", "Api Key")
	sc := newSecretsClient(c.client.BaseURL, c.ApiKey)

	assert.Equal(t, c.ApiKey, sc.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/secret", sc.client.client.BaseURL)
}

// Test_newCredentialsClient tests the constructor of newCredentialsClient
func Test_newCredentialsClient(t *testing.T) {
	c := newClient("http://localhost:12010/v2/v2", "Api Key")
	cc := newCredentialsClient(c.client.BaseURL, c.ApiKey)

	assert.Equal(t, c.ApiKey, cc.crud.client.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/credential", cc.crud.client.client.BaseURL)
	assert.Equal(t, c.ApiKey, cc.cloner.client.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/credential", cc.cloner.client.client.BaseURL)
	assert.Equal(t, c.ApiKey, cc.searcher.client.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/credential", cc.searcher.client.client.BaseURL)
}

// Test_newServersClient tests the constructor of newServersClient
func Test_newServersClient(t *testing.T) {
	c := newClient("http://localhost:12010/v2/v2", "Api Key")
	sc := newServersClient(c.client.BaseURL, c.ApiKey)

	assert.Equal(t, c.ApiKey, sc.crud.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/server", sc.crud.client.client.BaseURL)
	assert.Equal(t, c.ApiKey, sc.cloner.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/server", sc.cloner.client.client.BaseURL)
	assert.Equal(t, c.ApiKey, sc.searcher.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/server", sc.searcher.client.client.BaseURL)
}

// Test_newFilesClient tests the constructor of newFilesClient
func Test_newFilesClient(t *testing.T) {
	c := newClient("http://localhost:12010/v2/v2", "Api Key")
	fc := newFilesClient(c.client.BaseURL, c.ApiKey)

	assert.Equal(t, c.ApiKey, fc.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/file", fc.client.client.BaseURL)
}

// Test_newMaintenanceClient tests the constructor of newMaintenanceClient
func Test_newMaintenanceClient(t *testing.T) {
	c := newClient("http://localhost:12010/v2/v2", "Api Key")
	mc := newMaintenanceClient(c.client.BaseURL, c.ApiKey)

	assert.Equal(t, c.ApiKey, mc.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/maintenance", mc.client.client.BaseURL)
}

// Test_serversClient_Ping tests the Ping method.
func Test_serversClient_Ping(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		response         string
		expectedResponse gjson.Result
		err              error
	}{
		// Working case
		{
			name:       "Ping returns acknowledged true",
			method:     http.MethodGet,
			path:       "/server/f6950327-3175-4a98-a570-658df852424a/ping",
			statusCode: http.StatusOK,
			response: `{
			"acknowledged": true
			}`,
			expectedResponse: gjson.Parse(`{
			"acknowledged": true
			}`),
			err: nil,
		},

		// Error case
		{
			name:       "Ping returns a 502 error",
			method:     http.MethodGet,
			path:       "/server/f6950327-3175-4a98-a570-658df852424a/ping",
			statusCode: http.StatusBadGateway,
			response: `{
					"status": 502,
					"code": 8002,
					"messages": [
							"An error occurred while pinging the Mongo client."
					],
					"timestamp": "2025-08-26T20:42:26.372708600Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusBadGateway, Body: gjson.Parse(`{
					"status": 502,
					"code": 8002,
					"messages": [
							"An error occurred while pinging the Mongo client."
					],
					"timestamp": "2025-08-26T20:42:26.372708600Z"
			}`)},
		},
		{
			name:       "Ping returns a 400 error",
			method:     http.MethodGet,
			path:       "/server/f6950327-3175-4a98-a570-658df852424a/ping",
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
					"Failed to convert argument [id] for value [notuuid] due to: Invalid UUID string: notuuid"
			],
			"timestamp": "2025-09-30T15:35:00.121829500Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
					"Failed to convert argument [id] for value [notuuid] due to: Invalid UUID string: notuuid"
			],
			"timestamp": "2025-09-30T15:35:00.121829500Z"
			}`)},
		},
		{
			name:       "Ping returns a 422 error",
			method:     http.MethodGet,
			path:       "/server/f6950327-3175-4a98-a570-658df852424a/ping",
			statusCode: http.StatusUnprocessableEntity,
			response: `{
			"status": 422,
			"code": 4001,
			"messages": [
					"Client of type openai cannot be validated"
			],
			"timestamp": "2025-09-30T15:35:00.121829500Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusUnprocessableEntity, Body: gjson.Parse(`{
			"status": 422,
			"code": 4001,
			"messages": [
					"Client of type openai cannot be validated"
			],
			"timestamp": "2025-09-30T15:35:00.121829500Z"
			}`)},
		},
		{
			name:       "Ping returns a 404 error",
			method:     http.MethodGet,
			path:       "/server/f6950327-3175-4a98-a570-658df852424a/ping",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1003,
			"messages": [
				"Entity not found: f6950327-3175-4a98-a570-658df852424a"
			],
			"timestamp": "2025-09-30T15:38:42.885125200Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Entity not found: f6950327-3175-4a98-a570-658df852424a"
			],
			"timestamp": "2025-09-30T15:38:42.885125200Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, tc.path, r.URL.Path)
				}))
			defer srv.Close()

			serverClient := newServersClient(srv.URL, "")
			id := uuid.MustParse("f6950327-3175-4a98-a570-658df852424a")
			response, err := serverClient.Ping(id)
			assert.Equal(t, tc.expectedResponse, response)
			if tc.err == nil {
				require.NoError(t, err)
				assert.True(t, response.IsObject())
			} else {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

// Test_filesClient_Upload tests the fileClient.Upload() function.
func Test_filesClient_Upload(t *testing.T) {
	fileContent := "This is a test file"
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		response         string
		expectedResponse gjson.Result
		fileName         string
		err              error
	}{
		// Working case
		{
			name:       "Upload returns acknowledged true",
			method:     http.MethodPut,
			path:       "/file/testFile.txt",
			statusCode: http.StatusOK,
			response: `{
			"acknowledged": true
			}`,
			expectedResponse: gjson.Parse(`{
			"acknowledged": true
			}`),
			fileName: "testFile.txt",
			err:      nil,
		},
		// Error case
		{
			name:       "Upload uses a file that does not exist",
			method:     http.MethodPut,
			path:       "/file/doesNotExist.txt",
			statusCode: http.StatusBadRequest,
			response: `{
			"acknowledged": true
			}`,
			fileName:         "doesNotExist.txt",
			expectedResponse: gjson.Result{},
			err:              fmt.Errorf("open doesNotExist.txt"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, tc.path, r.URL.Path)
				body, _ := io.ReadAll(r.Body)
				assert.Contains(t, string(body), fileContent)
			}))
			defer srv.Close()

			filesClient := newFilesClient(srv.URL, "")

			if tc.err == nil {
				tmpFile, err := fileutils.CreateTemporaryFile(t.TempDir(), tc.fileName, fileContent)
				require.NoError(t, err)
				defer os.Remove(tmpFile)
				response, err := filesClient.Upload(tc.fileName, tmpFile)
				assert.Equal(t, tc.expectedResponse, response)
				require.NoError(t, err)
				assert.True(t, response.IsObject())
			} else {
				response, err := filesClient.Upload(tc.fileName, tc.fileName)
				assert.Equal(t, tc.expectedResponse, response)
				assert.Contains(t, err.Error(), tc.err.Error())
			}
		})
	}
}

// Test_filesClient_Retrieve tests the filesClient.Retrieve() function
func Test_filesClient_Retrieve(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		response   string
		fileName   string
		err        error
	}{
		// Working case
		{
			name:       "Retrieve returns the file's data",
			method:     http.MethodGet,
			path:       "/file/testFile",
			fileName:   "testFile",
			statusCode: http.StatusOK,
			response:   `This is a test file.`,
			err:        nil,
		},
		{
			name:       "Retrieve returns an empty file",
			method:     http.MethodGet,
			path:       "/file/empty",
			fileName:   "empty",
			statusCode: http.StatusOK,
			response:   ``,
			err:        nil,
		},
		// Error case
		{
			name:       "Retrieve returns 404 Not found",
			method:     http.MethodGet,
			path:       "/file/testFile",
			fileName:   "testFile",
			statusCode: http.StatusNotFound,
			response:   ``,
			err:        Error{Status: http.StatusNotFound, Body: gjson.Result{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/octet-stream", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, tc.path, r.URL.Path)
			}))
			defer srv.Close()

			filesClient := newFilesClient(srv.URL, "")
			response, err := filesClient.Retrieve(tc.fileName)
			assert.Equal(t, tc.response, string(response))
			if tc.err == nil {
				require.NoError(t, err)
			} else {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

// Test_filesClient_List tests the filesClient.List() function
func Test_filesClient_List(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		response         string
		expectedResponse []string
		err              error
	}{
		// Working case
		{
			name:       "List returns the list of files",
			method:     http.MethodGet,
			path:       "/file",
			statusCode: http.StatusOK,
			response: `[
				"Credential.ndjson",
				"Server.ndjson",
				"buildContextPrompt.js",
				"buildSimplePrompt.js",
				"constructPrompt.js",
				"constructSuggestedPrompt.js",
				"elastic-extraction.py",
				"extractReference.groovy",
				"extractReferenceAtlas.groovy",
				"formatAnalysisResponse.js",
				"formatAutocompleteResponse.js",
				"formatChunksResponse.js",
				"formatKeywordResponse.js",
				"formatKeywordResponseAtlas.js",
				"formatKeywordSearch.js",
				"formatQuestionsResponse.js",
				"formatSearchResponse.js",
				"formatSearchResponseAtlas.js",
				"formatSemanticResponse.js",
				"formatSuggestionsResponse.js",
				"keywordSearchTemplateAtlas.json",
				"searchTemplate.json",
				"searchTemplateAtlas.json"
				]`,
			expectedResponse: []string{
				"Credential.ndjson",
				"Server.ndjson",
				"buildContextPrompt.js",
				"buildSimplePrompt.js",
				"constructPrompt.js",
				"constructSuggestedPrompt.js",
				"elastic-extraction.py",
				"extractReference.groovy",
				"extractReferenceAtlas.groovy",
				"formatAnalysisResponse.js",
				"formatAutocompleteResponse.js",
				"formatChunksResponse.js",
				"formatKeywordResponse.js",
				"formatKeywordResponseAtlas.js",
				"formatKeywordSearch.js",
				"formatQuestionsResponse.js",
				"formatSearchResponse.js",
				"formatSearchResponseAtlas.js",
				"formatSemanticResponse.js",
				"formatSuggestionsResponse.js",
				"keywordSearchTemplateAtlas.json",
				"searchTemplate.json",
				"searchTemplateAtlas.json",
			},
		},
		{
			name:             "List returns no content",
			method:           http.MethodGet,
			path:             "/file",
			statusCode:       http.StatusNoContent,
			response:         ``,
			expectedResponse: []string{},
		},
		// Error case
		{
			name:             "List returns a response that cannot be marshalled into a []string",
			method:           http.MethodGet,
			path:             "/file",
			statusCode:       http.StatusOK,
			response:         `{"message"} : "This cannot be marshalled."`,
			expectedResponse: []string(nil),
			err:              fmt.Errorf("invalid character '}' after object key"),
		},
		{
			name:             "List returns a internal server error",
			method:           http.MethodGet,
			path:             "/file",
			statusCode:       http.StatusInternalServerError,
			response:         ``,
			expectedResponse: []string(nil),
			err:              Error{Status: http.StatusInternalServerError, Body: gjson.Result{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, tc.path, r.URL.Path)
			}))
			defer srv.Close()

			filesClient := newFilesClient(srv.URL, "")
			response, err := filesClient.List()
			assert.Equal(t, tc.expectedResponse, response)
			if tc.err == nil {
				require.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

// Test_filesClient_Delete tests de filesClient.Delete() function
func Test_filesClient_Delete(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		response         string
		expectedResponse gjson.Result
		err              error
	}{
		// Working case
		{
			name:       "Delete returns acknowledged true",
			method:     http.MethodDelete,
			path:       "/file/testFile",
			statusCode: http.StatusOK,
			response: `{
			"acknowledged": true
			}`,
			expectedResponse: gjson.Parse(`{
			"acknowledged": true
			}`),
			err: nil,
		},
		{
			name:       "Delete returns acknowledged false",
			method:     http.MethodDelete,
			path:       "/file/testFile",
			statusCode: http.StatusOK,
			response: `{
			"acknowledged": false
			}`,
			expectedResponse: gjson.Parse(`{
			"acknowledged": false
			}`),
			err: nil,
		},
		// Error case
		{
			name:             "Delete returns internal server error",
			method:           http.MethodDelete,
			path:             "/file/testFile",
			statusCode:       http.StatusInternalServerError,
			response:         `{"error": "Internal Server Error"}`,
			expectedResponse: gjson.Result{},
			err:              Error{Status: http.StatusInternalServerError, Body: gjson.Parse(`{"error": "Internal Server Error"}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, tc.path, r.URL.Path)
			}))
			defer srv.Close()

			filesClient := newFilesClient(srv.URL, "")
			response, err := filesClient.Delete("testFile")
			assert.Equal(t, tc.expectedResponse, response)
			if tc.err == nil {
				require.NoError(t, err)
				assert.True(t, response.IsObject())
			} else {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

// Test_maintenanceClient_Log tests the Log method
func Test_maintenanceClient_Log(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		componentName    string
		logLevel         LogLevel
		loggerName       string
		statusCode       int
		response         string
		expectedResponse gjson.Result
		err              error
	}{
		// Working case
		{
			name:          "Log returns acknowledged true",
			method:        http.MethodPost,
			path:          "/maintenance/log",
			componentName: "ingestion-api",
			logLevel:      LevelInfo,
			loggerName:    "api",
			statusCode:    http.StatusOK,
			response: `{
			"acknowledged": true
			}`,
			expectedResponse: gjson.Parse(`{
			"acknowledged": true
			}`),
			err: nil,
		},
		// Error cases
		{
			name:             "Log uses a log level that does not exist",
			method:           http.MethodPost,
			path:             "/maintenance/log",
			componentName:    "ingestion-api",
			logLevel:         "DOESNOTEXIST",
			loggerName:       "",
			statusCode:       http.StatusBadRequest,
			response:         `{"status":400,"code":3002,"messages":["Failed to convert argument [level] for value [DOESNOTEXIST] due to: No enum constant org.slf4j.event.Level.DOESNOTEXIST"],"timestamp":"2025-08-27T23:51:45.872308300Z"}`,
			expectedResponse: gjson.Result{},
			err:              Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{"status":400,"code":3002,"messages":["Failed to convert argument [level] for value [DOESNOTEXIST] due to: No enum constant org.slf4j.event.Level.DOESNOTEXIST"],"timestamp":"2025-08-27T23:51:45.872308300Z"}`)},
		},
		{
			name:             "Log sends an empty component name",
			method:           http.MethodPost,
			path:             "/maintenance/log",
			componentName:    "",
			logLevel:         LevelInfo,
			loggerName:       "",
			statusCode:       http.StatusBadRequest,
			response:         `{"status":400,"code":3002,"messages":["Required QueryValue [componentName] not specified"],"timestamp":"2025-08-28T00:03:31.103683200Z"}`,
			expectedResponse: gjson.Result{},
			err:              Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{"status":400,"code":3002,"messages":["Required QueryValue [componentName] not specified"],"timestamp":"2025-08-28T00:03:31.103683200Z"}`)},
		},
		{
			name:          "Log returns 422 error because the maintenance service is disabled.",
			method:        http.MethodPost,
			path:          "/maintenance/log",
			componentName: "",
			logLevel:      LevelInfo,
			loggerName:    "",
			statusCode:    http.StatusUnprocessableEntity,
			response: `{
			"status": 422,
			"code": 4001,
			"messages": [
				"The maintenance service is disabled per application configuration"
			],
			"timestamp": "2025-09-30T15:57:00.226830700Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusUnprocessableEntity, Body: gjson.Parse(`{
			"status": 422,
			"code": 4001,
			"messages": [
				"The maintenance service is disabled per application configuration"
			],
			"timestamp": "2025-09-30T15:57:00.226830700Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, tc.path, r.URL.Path)
				assert.Equal(t, tc.componentName, r.URL.Query().Get("componentName"))
				assert.Equal(t, string(tc.logLevel), r.URL.Query().Get("level"))
				assert.Equal(t, tc.loggerName, r.URL.Query().Get("loggerName"))
			}))
			defer srv.Close()

			maintenanceClient := newMaintenanceClient(srv.URL, "")
			response, err := maintenanceClient.Log(tc.componentName, tc.logLevel, tc.loggerName)
			assert.Equal(t, tc.expectedResponse, response)
			if tc.err == nil {
				require.NoError(t, err)
				assert.True(t, response.IsObject())
			} else {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

// Test_core_Labels tests the core.Labels() function
func Test_core_Labels(t *testing.T) {
	c := NewCore("http://localhost:12010", "Api Key")
	lc := c.Labels()

	assert.Equal(t, c.ApiKey, lc.ApiKey)
	assert.Equal(t, c.Url+"/label", lc.client.client.BaseURL)
}

// Test_core_Secrets tests the core.Secrets() function
func Test_core_Secrets(t *testing.T) {
	c := NewCore("http://localhost:12010", "Api Key")
	sc := c.Secrets()

	assert.Equal(t, c.ApiKey, sc.ApiKey)
	assert.Equal(t, c.Url+"/secret", sc.client.client.BaseURL)
}

// Test_core_Credentials tests the core.Credentials() function
func Test_core_Credentials(t *testing.T) {
	c := NewCore("http://localhost:12010", "Api Key")
	cc := c.Credentials()

	assert.Equal(t, c.ApiKey, cc.crud.ApiKey)
	assert.Equal(t, c.Url+"/credential", cc.crud.client.client.BaseURL)
	assert.Equal(t, c.ApiKey, cc.cloner.client.ApiKey)
	assert.Equal(t, c.Url+"/credential", cc.cloner.client.client.BaseURL)
	assert.Equal(t, c.ApiKey, cc.searcher.client.ApiKey)
	assert.Equal(t, c.Url+"/credential", cc.searcher.client.client.BaseURL)
}

// Test_core_Servers tests the core.Servers() function
func Test_core_Servers(t *testing.T) {
	c := NewCore("http://localhost:12010", "Api Key")
	sc := c.Servers()

	assert.Equal(t, c.ApiKey, sc.crud.ApiKey)
	assert.Equal(t, c.Url+"/server", sc.crud.client.client.BaseURL)
	assert.Equal(t, c.ApiKey, sc.cloner.ApiKey)
	assert.Equal(t, c.Url+"/server", sc.cloner.client.client.BaseURL)
	assert.Equal(t, c.ApiKey, sc.searcher.ApiKey)
	assert.Equal(t, c.Url+"/server", sc.searcher.client.client.BaseURL)
}

// Test_core_Files tests the core.Files() function
func Test_core_Files(t *testing.T) {
	c := NewCore("http://localhost:12010", "Api Key")
	fc := c.Files()

	assert.Equal(t, c.ApiKey, fc.ApiKey)
	assert.Equal(t, c.Url+"/file", fc.client.client.BaseURL)
}

// Test_core_Maintenance tests the core.Maintenance() function
func Test_core_Maintenance(t *testing.T) {
	c := NewCore("http://localhost:12010", "Api Key")
	mc := c.Maintenance()

	assert.Equal(t, c.ApiKey, mc.ApiKey)
	assert.Equal(t, c.Url+"/maintenance", mc.client.client.BaseURL)
}

// Test_core_BackupRestore tests the core.BackupRestore() function
func Test_core_BackupRestore(t *testing.T) {
	c := NewCore("http://localhost:12010", "Api Key")
	bc := c.BackupRestore()

	assert.Equal(t, c.ApiKey, bc.ApiKey)
	assert.Equal(t, c.Url, bc.client.client.BaseURL)
}

// Test_NewCore_UrlAndAPIKey tests the function to create a new core client.
// It verifies that the API Key and base URL correctly match.
func Test_NewCore_UrlAndAPIKey(t *testing.T) {
	url := "http://localhost:12010"
	apiKey := "secret-key"
	c := NewCore(url, apiKey)

	assert.Equal(t, apiKey, c.ApiKey, "ApiKey should be stored")
	assert.Equal(t, url+"/v2", c.Url, "BaseURL should match server URL")
}

package discovery

import (
	"fmt"
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
	c := newClient("http://localhost:8080/v2", "Api Key")
	lc := newLabelsClient(c.client.BaseURL, c.ApiKey)

	assert.Equal(t, c.ApiKey, lc.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/label", lc.client.client.BaseURL)
}

// Test_newSecretsClient tests the constructor of newSecretsClient
func Test_newSecretsClient(t *testing.T) {
	c := newClient("http://localhost:8080/v2", "Api Key")
	sc := newSecretsClient(c.client.BaseURL, c.ApiKey)

	assert.Equal(t, c.ApiKey, sc.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/secret", sc.client.client.BaseURL)
}

// Test_newCredentialsClient tests the constructor of newCredentialsClient
func Test_newCredentialsClient(t *testing.T) {
	c := newClient("http://localhost:8080/v2", "Api Key")
	cc := newCredentialsClient(c.client.BaseURL, c.ApiKey)

	assert.Equal(t, c.ApiKey, cc.crud.client.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/credential", cc.crud.client.client.BaseURL)
	assert.Equal(t, c.ApiKey, cc.cloner.client.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/credential", cc.cloner.client.client.BaseURL)
}

// Test_newServersClient tests the constructor of newServersClient
func Test_newServersClient(t *testing.T) {
	c := newClient("http://localhost:8080/v2", "Api Key")
	sc := newServersClient(c.client.BaseURL, c.ApiKey)

	assert.Equal(t, c.ApiKey, sc.crud.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/server", sc.crud.client.client.BaseURL)
	assert.Equal(t, c.ApiKey, sc.cloner.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/server", sc.cloner.client.client.BaseURL)
}

// Test_newFilesClient tests the constructor of newFilesClient
func Test_newFilesClient(t *testing.T) {
	c := newClient("http://localhost:8080/v2", "Api Key")
	fc := newFilesClient(c.client.BaseURL, c.ApiKey)

	assert.Equal(t, c.ApiKey, fc.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/file", fc.client.client.BaseURL)
}

// Test_newMaintenanceClient tests the constructor of newMaintenanceClient
func Test_newMaintenanceClient(t *testing.T) {
	c := newClient("http://localhost:8080/v2", "Api Key")
	mc := newMaintenanceClient(c.client.BaseURL, c.ApiKey)

	assert.Equal(t, c.ApiKey, mc.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/maintenance", mc.client.client.BaseURL)
}

// Test_serversClient_Ping tests the Ping method.
func Test_serversClient_Ping(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		response   string
		testFunc   func(t *testing.T, sc serversClient)
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
			testFunc: func(t *testing.T, sc serversClient) {
				id := uuid.MustParse("f6950327-3175-4a98-a570-658df852424a")
				response, err := sc.Ping(id)
				require.NoError(t, err)
				assert.True(t, response.Get("acknowledged").Bool())
			},
		},

		// Error case
		{
			name:       "Ping returns an error",
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
			testFunc: func(t *testing.T, sc serversClient) {
				id := uuid.MustParse("f6950327-3175-4a98-a570-658df852424a")
				response, err := sc.Ping(id)
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusBadGateway, `{
					"status": 502,
					"code": 8002,
					"messages": [
							"An error occurred while pinging the Mongo client."
					],
					"timestamp": "2025-08-26T20:42:26.372708600Z"
			}`))
			},
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

			c := newClient(srv.URL, "")
			serverClient := newServersClient(c.client.BaseURL, c.ApiKey)
			tc.testFunc(t, serverClient)
		})
	}
}

// Test_filesClient_CRUD tests all of the file CRUD operations.
func Test_filesClient_CRUD(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		response   string
		testFunc   func(t *testing.T, fc filesClient)
	}{
		// Working cases
		{
			name:       "Upload returns acknowledged true",
			method:     http.MethodPut,
			path:       "/file/testFile",
			statusCode: http.StatusOK,
			response: `{
			"acknowledged": true
			}`,
			testFunc: func(t *testing.T, fc filesClient) {
				tmpFile, err := fileutils.CreateTemporaryFile("", "testFile", "This is a test file")
				if err != nil {
					t.Fatalf("Failed to create file")
				}
				defer os.Remove(tmpFile)
				response, err := fc.Upload("testFile", tmpFile)
				require.NoError(t, err)
				assert.True(t, response.Get("acknowledged").Bool())
			},
		},
		{
			name:       "Retrieve returns the file's data",
			method:     http.MethodGet,
			path:       "/file/testFile",
			statusCode: http.StatusOK,
			response:   `This is a test file.`,
			testFunc: func(t *testing.T, fc filesClient) {
				response, err := fc.Retrieve("testFile")
				require.NoError(t, err)
				assert.Equal(t, `This is a test file.`, string(response))
			},
		},
		{
			name:       "Retrieve returns an empty file",
			method:     http.MethodGet,
			path:       "/file/empty",
			statusCode: http.StatusOK,
			response:   ``,
			testFunc: func(t *testing.T, fc filesClient) {
				response, err := fc.Retrieve("empty")
				require.NoError(t, err)
				assert.Empty(t, string(response))
			},
		},
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
			testFunc: func(t *testing.T, fc filesClient) {
				response, err := fc.List()
				require.NoError(t, err)
				assert.Len(t, response, 23)
			},
		},
		{
			name:       "List returns no content",
			method:     http.MethodGet,
			path:       "/file",
			statusCode: http.StatusNoContent,
			response:   ``,
			testFunc: func(t *testing.T, fc filesClient) {
				response, err := fc.List()
				require.NoError(t, err)
				assert.Len(t, response, 0)
			},
		},
		{
			name:       "Delete returns acknowledged true",
			method:     http.MethodDelete,
			path:       "/file/testFile",
			statusCode: http.StatusOK,
			response: `{
			"acknowledged": true
			}`,
			testFunc: func(t *testing.T, fc filesClient) {
				response, err := fc.Delete("testFile")
				require.NoError(t, err)
				assert.True(t, response.Get("acknowledged").Bool())
			},
		},
		{
			name:       "Delete returns acknowledged false",
			method:     http.MethodDelete,
			path:       "/file/testFile",
			statusCode: http.StatusOK,
			response: `{
			"acknowledged": false
			}`,
			testFunc: func(t *testing.T, fc filesClient) {
				response, err := fc.Delete("testFile")
				require.NoError(t, err)
				assert.False(t, response.Get("acknowledged").Bool())
			},
		},
		// Error cases
		{
			name:       "List returns a response that cannot be marshalled into an []string",
			method:     http.MethodGet,
			path:       "/file",
			statusCode: http.StatusOK,
			response:   `{"message"} : "This cannot be marshalled."`,
			testFunc: func(t *testing.T, fc filesClient) {
				response, err := fc.List()
				assert.Equal(t, []string{}, response)
				assert.EqualError(t, err, "invalid character '}' after object key")
			},
		},
		{
			name:       "List returns a internal server error",
			method:     http.MethodGet,
			path:       "/file",
			statusCode: http.StatusInternalServerError,
			response:   ``,
			testFunc: func(t *testing.T, fc filesClient) {
				response, err := fc.List()
				assert.Equal(t, []string{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusInternalServerError, ``))
			},
		},
		{
			name:       "Upload uses a file that does not exist",
			method:     http.MethodPut,
			path:       "/file/testFile",
			statusCode: http.StatusBadRequest,
			response: `{
			"acknowledged": true
			}`,
			testFunc: func(t *testing.T, fc filesClient) {
				response, err := fc.Upload("testFile", "doesNotExist.txt")
				assert.Equal(t, gjson.Result{}, response)
				assert.Contains(t, err.Error(), "open doesNotExist.txt")
			},
		},
		{
			name:       "Retrieve returns 404 Not found",
			method:     http.MethodGet,
			path:       "/file/testFile",
			statusCode: http.StatusNotFound,
			response:   ``,
			testFunc: func(t *testing.T, fc filesClient) {
				response, err := fc.Retrieve("testFile")
				assert.Equal(t, []byte{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusNotFound, ``))
			},
		},
		{
			name:       "Delete returns internal server error",
			method:     http.MethodDelete,
			path:       "/file/testFile",
			statusCode: http.StatusInternalServerError,
			response:   `{"error": "Internal Server Error"}`,
			testFunc: func(t *testing.T, fc filesClient) {
				response, err := fc.Delete("testFile")
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusInternalServerError, `{"error": "Internal Server Error"}`))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, tc.path, r.URL.Path)
			}))
			defer srv.Close()

			c := newClient(srv.URL, "")
			filesClient := newFilesClient(c.client.BaseURL, c.ApiKey)
			tc.testFunc(t, filesClient)
		})
	}
}

// Test_maintenanceClient_Log tests the Log method
func Test_maintenanceClient_Log(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		response   string
		testFunc   func(t *testing.T, fc maintenanceClient)
	}{
		// Working case
		{
			name:       "Log returns acknowledged true",
			method:     http.MethodPost,
			path:       "/maintenance/log",
			statusCode: http.StatusOK,
			response: `{
			"acknowledged": true
			}`,
			testFunc: func(t *testing.T, mc maintenanceClient) {
				response, err := mc.Log("ingestion-api", "INFO", "")
				require.NoError(t, err)
				assert.True(t, response.Get("acknowledged").Bool())
			},
		},
		// Error cases
		{
			name:       "Log uses a log level that does not exist",
			method:     http.MethodPost,
			path:       "/maintenance/log",
			statusCode: http.StatusBadRequest,
			response:   `{"status":400,"code":3002,"messages":["Failed to convert argument [level] for value [DOESNOTEXIST] due to: No enum constant org.slf4j.event.Level.DOESNOTEXIST"],"timestamp":"2025-08-27T23:51:45.872308300Z"}`,
			testFunc: func(t *testing.T, mc maintenanceClient) {
				response, err := mc.Log("ingestion-api", "DOESNOTEXIST", "")
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusBadRequest, `{"status":400,"code":3002,"messages":["Failed to convert argument [level] for value [DOESNOTEXIST] due to: No enum constant org.slf4j.event.Level.DOESNOTEXIST"],"timestamp":"2025-08-27T23:51:45.872308300Z"}`))
			},
		},
		{
			name:       "Log sends an empty component name",
			method:     http.MethodPost,
			path:       "/maintenance/log",
			statusCode: http.StatusBadRequest,
			response:   `{"status":400,"code":3002,"messages":["Required QueryValue [componentName] not specified"],"timestamp":"2025-08-28T00:03:31.103683200Z"}`,
			testFunc: func(t *testing.T, mc maintenanceClient) {
				response, err := mc.Log("", "INFO", "")
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusBadRequest, `{"status":400,"code":3002,"messages":["Required QueryValue [componentName] not specified"],"timestamp":"2025-08-28T00:03:31.103683200Z"}`))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, tc.path, r.URL.Path)
			}))
			defer srv.Close()

			c := newClient(srv.URL, "")
			maintenanceClient := newMaintenanceClient(c.client.BaseURL, c.ApiKey)
			tc.testFunc(t, maintenanceClient)
		})
	}
}

// Test_core_Labels tests the core.Labels() function
func Test_core_Labels(t *testing.T) {
	c := NewCore("http://localhost:8080/v2", "Api Key")
	lc := c.Labels()

	assert.Equal(t, c.ApiKey, lc.ApiKey)
	assert.Equal(t, c.Url+"/label", lc.client.client.BaseURL)
}

// Test_core_Secrets tests the core.Secrets() function
func Test_core_Secrets(t *testing.T) {
	c := NewCore("http://localhost:8080/v2", "Api Key")
	sc := c.Secrets()

	assert.Equal(t, c.ApiKey, sc.ApiKey)
	assert.Equal(t, c.Url+"/secret", sc.client.client.BaseURL)
}

// Test_core_Credentials tests the core.Credentials() function
func Test_core_Credentials(t *testing.T) {
	c := NewCore("http://localhost:8080/v2", "Api Key")
	cc := c.Credentials()

	assert.Equal(t, c.ApiKey, cc.ApiKey)
	assert.Equal(t, c.Url+"/credential", cc.client.client.BaseURL)
	assert.Equal(t, c.ApiKey, cc.cloner.client.ApiKey)
	assert.Equal(t, c.Url+"/credential", cc.cloner.client.client.BaseURL)
}

// Test_core_Servers tests the core.Servers() function
func Test_core_Servers(t *testing.T) {
	c := NewCore("http://localhost:8080/v2", "Api Key")
	sc := c.Servers()

	assert.Equal(t, c.ApiKey, sc.ApiKey)
	assert.Equal(t, c.Url+"/server", sc.client.client.BaseURL)
	assert.Equal(t, c.ApiKey, sc.cloner.ApiKey)
	assert.Equal(t, c.Url+"/server", sc.cloner.client.client.BaseURL)
}

// Test_core_Files tests the core.Files() function
func Test_core_Files(t *testing.T) {
	c := NewCore("http://localhost:8080/v2", "Api Key")
	fc := c.Files()

	assert.Equal(t, c.ApiKey, fc.ApiKey)
	assert.Equal(t, c.Url+"/file", fc.client.client.BaseURL)
}

// Test_core_Maintenance tests the core.Maintenance() function
func Test_core_Maintenance(t *testing.T) {
	c := NewCore("http://localhost:8080/v2", "Api Key")
	mc := c.Maintenance()

	assert.Equal(t, c.ApiKey, mc.ApiKey)
	assert.Equal(t, c.Url+"/maintenance", mc.client.client.BaseURL)
}

// Test_NewCore_UrlAndAPIKey tests the function to create a new core client.
// It verifies that the API Key and base URL correctly match.
func Test_NewCore_UrlAndAPIKey(t *testing.T) {
	url := "http://localhost:8080/v2"
	apiKey := "secret-key"
	c := NewCore(url, apiKey)

	assert.Equal(t, apiKey, c.ApiKey, "ApiKey should be stored")
	assert.Equal(t, url, c.Url, "BaseURL should match server URL")
}

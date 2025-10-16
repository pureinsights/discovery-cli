package discovery

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/pureinsights/pdp-cli/internal/fileutils"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Test_newClient_BaseURLAndAPIKey tests the function to create a new client.
// It verifies that the API Key and base URL correctly match.
func Test_newClient_BaseURLAndAPIKey(t *testing.T) {
	url := "http://localhost:12010/v2"
	apiKey := "secret-key"
	c := newClient(url, apiKey)

	assert.Equal(t, apiKey, c.ApiKey, "ApiKey should be stored")
	assert.Equal(t, url, c.client.BaseURL, "BaseURL should match server URL")
}

// Test_newSubClient_BaseURLJoin tests if the newSubClient function correctly handles edge cases when joining a URL and a path.
func Test_newSubClient_BaseURLJoin(t *testing.T) {
	tests := []struct {
		name string
		base string
		path string
		want string
	}{
		{
			name: "no slashes",
			base: "http://localhost",
			path: "api",
			want: "http://localhost/api",
		},
		{
			name: "base has trailing slash",
			base: "http://localhost/",
			path: "api",
			want: "http://localhost/api",
		},
		{
			name: "path has leading slash",
			base: "http://localhost",
			path: "/api",
			want: "http://localhost/api",
		},
		{
			name: "both have slashes",
			base: "http://localhost/",
			path: "/api",
			want: "http://localhost/api",
		},
		{
			name: "collapse multiple slashes and keep trailing on path",
			base: "http://localhost///",
			path: "/api/",
			want: "http://localhost/api",
		},
		{
			name: "nested base path",
			base: "http://localhost/v2",
			path: "api",
			want: "http://localhost/v2/api",
		},
		{
			name: "localhost without scheme and extra slashes",
			base: "localhost///",
			path: "/v2",
			want: "localhost/v2",
		},
		{
			name: "empty path keeps base",
			base: "http://localhost/",
			path: "",
			want: "http://localhost",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parent := newClient(tc.base, "apiKey")
			got := newSubClient(parent, tc.path)

			assert.Equalf(t, parent.ApiKey, got.ApiKey, "API Key not inherited")
			assert.Equalf(t, tc.want, got.client.BaseURL, "Base URL is different:\n")
		})
	}
}

// Test_client_execute_SendsAPIKeyReturnsBody tests when execute() sets the API key and returns the response's body.
func Test_client_execute_SendsAPIKeyReturnsBody(t *testing.T) {
	const apiKey = "api-key"

	srv := httptest.NewServer(testutils.HttpHandler(t, http.StatusOK, "application/json", `{"ok":true}`,
		func(t *testing.T, r *http.Request) {
			assert.Equal(t, "/seed", r.URL.Path)
			assert.Equal(t, apiKey, r.Header.Get("X-API-Key"))
		}))
	t.Cleanup(srv.Close)

	c := newClient(srv.URL, apiKey)

	res, err := c.execute(http.MethodGet, "/seed")
	require.NoError(t, err)

	assert.IsType(t, []byte(nil), res)

	assert.Equal(t, `{"ok":true}`, string(res))
}

// Test_client_execute_HTTPErrorTypedError tests when the response is an error.
func Test_client_execute_HTTPErrorTypedError(t *testing.T) {
	type tc struct {
		name        string
		status      int
		contentType string
		body        string
		expectBody  string
	}

	tests := []tc{
		{
			name:        "403 with JSON string body",
			status:      http.StatusForbidden,
			contentType: "text/plain",
			body:        `"Forbidden"`,
			expectBody:  "Forbidden",
		},
		{
			name:        "404 with JSON object",
			status:      http.StatusNotFound,
			contentType: "application/json",
			body:        `{"message":"missing"}`,
			expectBody:  `{"message":"missing"}`,
		},
		{
			name:        "415 with JSON array",
			status:      http.StatusUnsupportedMediaType,
			contentType: "application/json",
			body:        `["a","b"]`,
			expectBody:  `["a","b"]`,
		},
		{
			name:        "500 with empty body",
			status:      http.StatusInternalServerError,
			contentType: "application/json",
			body:        "",
			expectBody:  "",
		},
		{
			name:        "400 with invalid JSON/plain text",
			status:      http.StatusBadRequest,
			contentType: "text/plain",
			body:        `Forbidden`,
			expectBody:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tt.status, tt.contentType, tt.body, nil))
			t.Cleanup(srv.Close)

			c := newClient(srv.URL, "")

			res, err := c.execute(http.MethodGet, "/fail")
			assert.Nil(t, res, "result should be nil on response error")
			require.Error(t, err, "expected an error")

			assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", tt.status, tt.expectBody))
		})
	}
}

// Test_client_execute_RestyReturnsError tests when the Resty Execute function returns an error.
func Test_client_execute_RestyReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	base := srv.URL
	srv.Close()

	c := newClient(base, "")
	res, err := c.execute(http.MethodGet, "/down")
	require.Error(t, err)
	assert.Nil(t, res, "result should be nil on execute error")
	assert.Contains(t, err.Error(), base+"/down")
}

// Test_client_execute_NoContent tests the client.execute() function when it receives a No Content Response.
func Test_client_execute_NoContent(t *testing.T) {
	srv := httptest.NewServer(testutils.HttpNoContentHandler(t, nil))
	defer srv.Close()

	c := newClient(srv.URL, "")
	response, err := c.execute(http.MethodGet, "")
	require.NoError(t, err)
	assert.Len(t, response, 0)
}

// Test_client_execute_FunctionalOptionsFail tests when one of the functional options returns an error.
func Test_client_execute_FunctionalOptionsFail(t *testing.T) {
	failingOption := func(r *resty.Request) error {
		return errors.New("The option failed")
	}
	srv := httptest.NewServer(http.NotFoundHandler())
	base := srv.URL
	srv.Close()

	c := newClient(base, "")
	res, err := c.execute(http.MethodGet, "/down", failingOption)
	assert.EqualError(t, err, "The option failed")
	assert.Nil(t, res, "result should be nil on execute error")
}

// TestWithQueryParameters tests the WithQueryParameters() options.
// It tests WithQueryParameters() with a single value and an array.
func TestWithQueryParameters(t *testing.T) {
	srv := httptest.NewServer(
		testutils.HttpHandler(t,
			http.StatusOK, "application/json", `{"ok":true}`,
			func(t *testing.T, r *http.Request) {
				assert.Equal(t, "Google", r.URL.Query().Get("q"))
				assert.Equal(t, []string{"item1", "item2", "item3"}, r.URL.Query()["items"])
			}))
	t.Cleanup(srv.Close)

	c := newClient(srv.URL, "")
	response, err := c.execute("POST", "", WithQueryParameters(map[string][]string{"q": {"Google"}, "items": {"item1", "item2", "item3"}}))
	require.NoError(t, err)
	require.True(t, gjson.Parse(string(response)).Get("ok").Bool())
}

// TestWithJSONBody tests the WithJSONBody() option.
// It verifies that the sent body is a JSON and with the correct content.
func TestWithJSONBody(t *testing.T) {
	srv := httptest.NewServer(
		testutils.HttpHandler(t,
			http.StatusOK, "application/json", `{"ok":true}`,
			func(t *testing.T, r *http.Request) {
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				body, _ := io.ReadAll(r.Body)
				bodyJSON := gjson.Parse(string(body))
				assert.Equal(t, "test-secret", bodyJSON.Get("name").String())
				assert.True(t, bodyJSON.Get("active").Bool())
			}))
	t.Cleanup(srv.Close)

	c := newClient(srv.URL, "")
	response, err := c.execute("POST", "",
		WithJSONBody(`{
		"name": "test-secret",
		"active": true
		}`))
	require.NoError(t, err)
	require.True(t, gjson.Parse(string(response)).Get("ok").Bool())
}

// TestWithFile tests the WithFile() option.
func TestWithFile(t *testing.T) {
	srv := httptest.NewServer(testutils.HttpHandler(t, http.StatusOK, "application/json", `{"ok":true}`,
		func(t *testing.T, r *http.Request) {
			assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")
			body, _ := io.ReadAll(r.Body)
			assert.Contains(t, string(body), "This is a test file")
		}))
	t.Cleanup(srv.Close)

	tmpFile, err := fileutils.CreateTemporaryFile(t.TempDir(), "testFile.txt", "This is a test file")
	require.NoError(t, err)

	defer os.Remove(tmpFile)

	c := newClient(srv.URL, "")
	response, err := c.execute("PUT", "", WithFile(tmpFile))
	require.NoError(t, err)
	require.True(t, gjson.Parse(string(response)).Get("ok").Bool())
}

// Test_execute_ParsedResult tests the execute() function when gjson correctly parses the response.
func Test_execute_ParsedResult(t *testing.T) {
	srv := httptest.NewServer(
		testutils.HttpHandler(t, http.StatusOK, "application/json", `{
		"name": "test-secret",
		"active": true,
		"content": { 
			"username": "user"
		}
		}`, nil))
	t.Cleanup(srv.Close)

	c := newClient(srv.URL, "")
	response, err := execute(c, "GET", "")
	require.NoError(t, err)
	assert.Equal(t, "test-secret", response.Get("name").String())
	assert.Equal(t, "user", response.Get("content.username").String())
}

// Test_execute_NoContent tests the execute() function when it receives a No Content Response.
func Test_execute_NoContent(t *testing.T) {
	srv := httptest.NewServer(testutils.HttpNoContentHandler(t, nil))
	defer srv.Close()

	c := newClient(srv.URL, "")
	response, err := execute(c, http.MethodGet, "")
	require.NoError(t, err)
	assert.Equal(t, gjson.Null, response.Type)
	assert.Equal(t, "", response.Raw)
}

// Test_execute_HTTPError tests the execute function when the response is an error.
func Test_execute_HTTPError(t *testing.T) {
	srv := httptest.NewServer(testutils.HttpHandler(t,
		http.StatusNotFound, "application/json", `{"message":"missing"}`, nil))
	t.Cleanup(srv.Close)

	c := newClient(srv.URL, "")
	response, err := execute(c, "GET", "")
	assert.Equal(t, response, gjson.Result{})
	assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusNotFound, []byte(`{"message":"missing"}`)))
}

// Test_execute_RestyReturnsError tests the execute function when Resty returns an error.
func Test_execute_RestyReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	base := srv.URL
	srv.Close()

	c := newClient(base, "")
	response, err := execute(c, http.MethodGet, "/down")
	require.Error(t, err)
	assert.Equal(t, response, gjson.Result{})
	assert.Contains(t, err.Error(), base+"/down")
}

// Test_executeWithPagination_HTTPResponseCases tests how the executeWithPagination() function behaves with various HTTP responses.
func Test_executeWithPagination_HTTPResponseCases(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		statusCode  int
		response    string
		expectedLen int
		err         error
	}{
		// Working cases
		{
			name:       "executeWithPagination returns array",
			method:     http.MethodGet,
			path:       "/",
			statusCode: http.StatusOK,
			response: `{
			"content": [
				{
				"type": "mongo",
				"name": "MongoDB text processor 4",
				"labels": [],
				"active": true,
				"id": "3393f6d9-94c1-4b70-ba02-5f582727d998",
				"creationTimestamp": "2025-08-21T17:57:16Z",
				"lastUpdatedTimestamp": "2025-08-21T17:57:16Z"
				},
				{
				"type": "mongo",
				"name": "MongoDB text processor",
				"labels": [],
				"active": true,
				"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
				},
				{
				"type": "script",
				"name": "Script processor",
				"labels": [],
				"active": true,
				"id": "86e7f920-a4e4-4b64-be84-5437a7673db8",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-14T18:02:38Z"
				}
			],
			"pageable": {
				"page": 0,
				"size": 3,
				"sort": []
			},
			"totalSize": 12,
			"totalPages": 4,
			"empty": false,
			"size": 3,
			"offset": 0,
			"numberOfElements": 3,
			"pageNumber": 0
			}`,
			expectedLen: 12,
			err:         nil,
		},
		{
			name:        "executeWithPagination returns no content",
			method:      http.MethodGet,
			path:        "/",
			statusCode:  http.StatusNoContent,
			response:    `{"content": []}`,
			expectedLen: 0,
			err:         nil,
		},
		{
			name:        "executeWithPagination has no content field",
			method:      http.MethodGet,
			path:        "/",
			statusCode:  http.StatusNoContent,
			response:    ``,
			expectedLen: 0,
			err:         nil,
		},

		// Error cases
		{
			name:       "executeWithPagination returns a 401 Unauthorized",
			method:     http.MethodGet,
			path:       "/",
			statusCode: http.StatusUnauthorized,
			response:   `{"error":"unauthorized"}`,
			err:        Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)},
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
			results, err := executeWithPagination(c, tc.method, "")
			if tc.err == nil {
				require.NoError(t, err)
				assert.Len(t, results, tc.expectedLen)
			} else {
				assert.Equal(t, []gjson.Result(nil), results)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

// Test_executeWithPagination_ErrorInSecondPage tests when executeWithPagination fails in a request while trying to get every content from every page.
func Test_executeWithPagination_ErrorInSecondPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/getall", r.URL.Path)
			pageNumber, _ := strconv.Atoi(r.URL.Query().Get("page"))
			w.Header().Set("Content-Type", "application/json")
			if pageNumber > 0 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"Internal Server Error"}`))
			} else {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{
			"content": [
				{
				"type": "mongo",
				"name": "MongoDB text processor 4",
				"labels": [],
				"active": true,
				"id": "3393f6d9-94c1-4b70-ba02-5f582727d998",
				"creationTimestamp": "2025-08-21T17:57:16Z",
				"lastUpdatedTimestamp": "2025-08-21T17:57:16Z"
				},
				{
				"type": "mongo",
				"name": "MongoDB text processor",
				"labels": [],
				"active": true,
				"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
				},
				{
				"type": "script",
				"name": "Script processor",
				"labels": [],
				"active": true,
				"id": "86e7f920-a4e4-4b64-be84-5437a7673db8",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-14T18:02:38Z"
				}
			],
			"pageable": {
				"page": 0,
				"size": 3,
				"sort": []
			},
			"totalSize": 12,
			"totalPages": 4,
			"empty": false,
			"size": 3,
			"offset": 0,
			"numberOfElements": 3,
			"pageNumber": 0
			}`))
			}
		}))
	t.Cleanup(srv.Close)

	c := newClient(srv.URL, "")
	response, err := executeWithPagination(c, http.MethodGet, "/getall")
	assert.Equal(t, []gjson.Result(nil), response)
	var errStruct Error
	require.ErrorAs(t, err, &errStruct)
	assert.EqualError(t, err, Error{Status: http.StatusInternalServerError, Body: gjson.Parse(`{"error":"Internal Server Error"}`)}.Error())
}

// Test_executeWithPagination_RestyReturnsError tests what happens when the Resty client fails to execute the request.
func Test_executeWithPagination_RestyReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	base := srv.URL
	srv.Close()

	c := newClient(base, "")
	response, err := executeWithPagination(c, http.MethodGet, "/down")
	require.Error(t, err)
	assert.Equal(t, response, []gjson.Result(nil))
	assert.Contains(t, err.Error(), base+"/down")
}

// Test_executeWithPagination_ContentInSecondPage tests that the executeWithPagination() function
// can successfully get all content when there are two pages with content in them
func Test_executeWithPagination_ContentInSecondPage(t *testing.T) {
	body := `{
	"equals": {
		"field": "type",
		"value": "mongo",
		"normalize": true
	}
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		requestBody, _ := io.ReadAll(r.Body)
		assert.Equal(t, gjson.Parse(body), gjson.Parse(string(requestBody)))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "/getall", r.URL.Path)
		pageNumber, _ := strconv.Atoi(r.URL.Query().Get("page"))
		w.Header().Set("Content-Type", "application/json")
		if pageNumber > 0 {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			"content": [
				{
					"type": "openai",
					"name": "OpenAI Chat Processor",
					"labels": [],
					"active": true,
					"id": "8a399b1c-95fc-406c-a220-7d321aaa7b0e",
					"creationTimestamp": "2025-08-14T18:02:38Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:38Z"
				},
				{
					"type": "mongo",
					"name": "MongoDB vector processor",
					"labels": [],
					"active": true,
					"id": "a5ee116b-bd95-474e-9d50-db7be988b196",
					"creationTimestamp": "2025-08-14T18:02:38Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:38Z"
				},
				{
					"type": "openai",
					"name": "OpenAI embeddings processor",
					"labels": [],
					"active": true,
					"id": "b5c25cd3-e7c9-4fd2-b7e6-2bcf6e2caf89",
					"creationTimestamp": "2025-08-14T18:02:38Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:38Z"
				}
			],
			"pageable": {
				"page": 0,
				"size": 3,
				"sort": []
			},
			"totalSize": 6,
			"totalPages": 2,
			"empty": false,
			"size": 3,
			"offset": 0,
			"numberOfElements": 3,
			"pageNumber": 1
			}`))
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			"content": [
				{
				"type": "mongo",
				"name": "MongoDB text processor 4",
				"labels": [],
				"active": true,
				"id": "3393f6d9-94c1-4b70-ba02-5f582727d998",
				"creationTimestamp": "2025-08-21T17:57:16Z",
				"lastUpdatedTimestamp": "2025-08-21T17:57:16Z"
				},
				{
				"type": "mongo",
				"name": "MongoDB text processor",
				"labels": [],
				"active": true,
				"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
				},
				{
				"type": "script",
				"name": "Script processor",
				"labels": [],
				"active": true,
				"id": "86e7f920-a4e4-4b64-be84-5437a7673db8",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-14T18:02:38Z"
				}
			],
			"pageable": {
				"page": 0,
				"size": 3,
				"sort": []
			},
			"totalSize": 6,
			"totalPages": 2,
			"empty": false,
			"size": 3,
			"offset": 0,
			"numberOfElements": 3,
			"pageNumber": 0
			}`))
		}
	}))
	t.Cleanup(srv.Close)

	c := newClient(srv.URL, "")
	response, err := executeWithPagination(c, http.MethodPost, "/getall", WithJSONBody(body))
	require.NoError(t, err)
	assert.Len(t, response, 6)
}

// Test_executeWithPagination_NoContentInSecondPage tests what happens if one of the later pages returns No Content
func Test_executeWithPagination_NoContentInSecondPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/getall", r.URL.Path)
		pageNumber, _ := strconv.Atoi(r.URL.Query().Get("page"))
		w.Header().Set("Content-Type", "application/json")
		if pageNumber > 0 {
			w.WriteHeader(http.StatusNoContent)
			_, _ = w.Write([]byte(`[]`))
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			"content": [
				{
				"type": "mongo",
				"name": "MongoDB text processor 4",
				"labels": [],
				"active": true,
				"id": "3393f6d9-94c1-4b70-ba02-5f582727d998",
				"creationTimestamp": "2025-08-21T17:57:16Z",
				"lastUpdatedTimestamp": "2025-08-21T17:57:16Z"
				},
				{
				"type": "mongo",
				"name": "MongoDB text processor",
				"labels": [],
				"active": true,
				"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
				},
				{
				"type": "script",
				"name": "Script processor",
				"labels": [],
				"active": true,
				"id": "86e7f920-a4e4-4b64-be84-5437a7673db8",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-14T18:02:38Z"
				}
			],
			"pageable": {
				"page": 0,
				"size": 3,
				"sort": []
			},
			"totalSize": 6,
			"totalPages": 2,
			"empty": false,
			"size": 3,
			"offset": 0,
			"numberOfElements": 3,
			"pageNumber": 0
			}`))
		}
	}))
	t.Cleanup(srv.Close)

	c := newClient(srv.URL, "")
	response, err := executeWithPagination(c, http.MethodGet, "/getall")
	require.NoError(t, err)
	assert.Len(t, response, 3)
}

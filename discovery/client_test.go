package discovery

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Test_newClient_BaseURLAndAPIKey tests the function to create a new client.
// It verifies that the API Key and base URL correctly match.
func Test_newClient_BaseURLAndAPIKey(t *testing.T) {
	url := "http://localhost:8080"
	apiKey := "secret-key"
	c := newClient(url, apiKey)

	assert.Equal(t, apiKey, c.ApiKey, "ApiKey should be stored")
	assert.Equal(t, url, c.client.BaseURL, "BaseURL should match server URL")
}

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

			if got.ApiKey != parent.ApiKey {
				t.Fatalf("API Key not inherited: got %q want %q", got.ApiKey, parent.ApiKey)
			}

			if got.client.BaseURL != tc.want {
				t.Fatalf("Base URL is different:\n  base=%q path=%q\n  got =%q\n  want=%q",
					tc.base, tc.path, got.client.BaseURL, tc.want)
			}
		})
	}
}

// Test_client_execute_SendsAPIKeyReturnsBody tests when execute() sets the API key and returns the response's body.
func Test_client_execute_SendsAPIKeyReturnsBody(t *testing.T) {
	const apiKey = "api-key"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/seed", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
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
		body        []byte
		expectBody  string
	}

	tests := []tc{
		{
			name:        "403 with JSON string body",
			status:      http.StatusForbidden,
			contentType: "text/plain",
			body:        []byte(`"Forbidden"`),
			expectBody:  "Forbidden",
		},
		{
			name:        "404 with JSON object",
			status:      http.StatusNotFound,
			contentType: "application/json",
			body:        []byte(`{"message":"missing"}`),
			expectBody:  `{"message":"missing"}`,
		},
		{
			name:        "415 with JSON array",
			status:      http.StatusUnsupportedMediaType,
			contentType: "application/json",
			body:        []byte(`["a","b"]`),
			expectBody:  `["a","b"]`,
		},
		{
			name:        "500 with empty body",
			status:      http.StatusInternalServerError,
			contentType: "application/json",
			body:        nil,
			expectBody:  "",
		},
		{
			name:        "400 with invalid JSON/plain text",
			status:      http.StatusBadRequest,
			contentType: "text/plain",
			body:        []byte(`Forbidden`),
			expectBody:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tt.contentType != "" {
					w.Header().Set("Content-Type", tt.contentType)
				}
				w.WriteHeader(tt.status)
				if tt.body != nil {
					_, _ = w.Write(tt.body)
				}
			}))
			t.Cleanup(srv.Close)

			c := newClient(srv.URL, "")

			res, err := c.execute(http.MethodGet, "/fail")
			assert.Nil(t, res, "result should be nil on response error")
			require.Error(t, err, "expected an error")

			assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", tt.status, tt.expectBody))
		})
	}
}

// TestExecute_RestyReturnsError tests when the Resty Execute function returns an error.
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

// Test_client_execute_FunctionalOptionsFail tests when one of the functional options returns an error.
func Test_client_execute_FunctionalOptionsFail(t *testing.T) {
	failingOption := func(r *resty.Request) error {
		return Error{
			Status: http.StatusBadRequest,
			Body:   gjson.Parse(`{"error": "RequestOption Failed"}`),
		}
	}
	srv := httptest.NewServer(http.NotFoundHandler())
	base := srv.URL
	srv.Close()

	c := newClient(base, "")
	res, err := c.execute(http.MethodGet, "/down", failingOption)
	assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusBadRequest, []byte(`{"error": "RequestOption Failed"}`)))
	assert.Nil(t, res, "result should be nil on execute error")
}

// TestRequestOption_FunctionalOptions tests the WithQueryParameters() and WithJSONBody() options.
// It tests WithQueryParameters() with a single value and an array.
func TestRequestOption_FunctionalOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Google", r.URL.Query().Get("q"))
		assert.Equal(t, []string{"item1", "item2", "item3"}, r.URL.Query()["items"])
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, "test-secret", gjson.Parse(string(body)).Get("name").String())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	c := newClient(srv.URL, "")
	response, err := c.execute("POST", "", WithQueryParameters(map[string][]string{"q": {"Google"}, "items": {"item1", "item2", "item3"}}),
		WithJSONBody(`{
		"name": "test-secret",
		"active": true
		}`))
	require.NoError(t, err)
	require.True(t, gjson.Parse(string(response)).Get("ok").Bool())
}

// TestRequestOption_FileOption tests the WithFile() option.
func TestRequestOption_FileOption(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "Esto es un archivo de prueba")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	c := newClient(srv.URL, "")
	response, err := c.execute("PUT", "", WithFile("test_files/testFile.txt"))
	require.NoError(t, err)
	require.True(t, gjson.Parse(string(response)).Get("ok").Bool())
}

// Tests the execute() function when gjson correctly parses the response.
func Test_execute_ParsedResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
		"name": "test-secret",
		"active": true,
		"content": { 
			"username": "user"
		}
		}`))
	}))
	t.Cleanup(srv.Close)

	c := newClient(srv.URL, "")
	response, err := execute(c, "GET", "")
	require.NoError(t, err)
	assert.Equal(t, "test-secret", response.Get("name").String())
	assert.Equal(t, "user", response.Get("content.username").String())
}

// Test_execute_HTTPError tests the execute function when the response is an error.
func Test_execute_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"missing"}`))
	}))
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

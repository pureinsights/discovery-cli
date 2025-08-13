package discovery

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewClient_BaseURLAndAPIKey tests the function to create a new client.
// It verifies that the ALI Key and base URL correctly match.
func TestNewClient_BaseURLAndAPIKey(t *testing.T) {
	url := "http://localhost:8080"
	apiKey := "secret-key"
	c := newClient(url, apiKey)

	assert.Equal(t, apiKey, c.ApiKey, "ApiKey should be stored")
	assert.Equal(t, url, c.client.BaseURL, "BaseURL should match server URL")
}

// TestNewSubClient_AppendsPath tests if the sub client is correctly adding its path to the parent URL.
func TestNewSubClient_AppendsPath(t *testing.T) {
	url := "http://localhost:8080"
	path := "/seed"

	parent := newClient(url, "key")
	sub := newSubClient(parent, path)

	assert.Equal(t, parent.ApiKey, sub.ApiKey, "subclient should inherit ApiKey")

	want := parent.client.BaseURL + path
	assert.Equal(t, want, sub.client.BaseURL, "subclient BaseURL should append path")
}

// TestExecute_SendsAPIKeyReturnsBody tests when execute() sets the API key and returns the response's body.
func TestExecute_SendsAPIKeyReturnsBody(t *testing.T) {
	const apiKey = "api-key"
	var sawHeader bool

	// Sets up a mock server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/seed", r.URL.Path)     // Verifies that the client sent the request to the correct path.
		if r.Header.Get("X-API-Key") == apiKey { // Verifies that the API key was set.
			sawHeader = true
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	c := newClient(srv.URL, apiKey)

	res, err := c.execute(http.MethodGet, "/seed")
	require.NoError(t, err)
	assert.True(t, sawHeader, "X-API-Key header should be sent")

	body, ok := res.([]byte)
	require.True(t, ok, "expected []byte from execute result")
	assert.Equal(t, `{"ok":true}`, strings.TrimSpace(string(body)))
}

// TestExecute_HTTPErrorTypedError tests when the response is an error.
func TestExecute_HTTPErrorTypedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden) // Sets the response as an error.
		_, _ = w.Write([]byte(`{"error":"Forbidden"}`))
	}))
	t.Cleanup(srv.Close)

	c := newClient(srv.URL, "")

	res, err := c.execute(http.MethodGet, "/fail")
	assert.Nil(t, res, "result should be nil on response error")
	require.Error(t, err, "expected an error")

	e, ok := err.(Error)
	require.True(t, ok, "error should be of type Error")
	// Verifies that the Error struct has the correct values.
	assert.Equal(t, http.StatusForbidden, e.Status)
	assert.Equal(t, "Forbidden", e.Body.Get("error").String())
}

// TestExecute_RestyReturnsError tests when the Resty Execute function returns an error.
func TestExecute_RestyReturnsError(t *testing.T) {
	c := newClient("http://fakeserver", "")

	res, err := c.execute(http.MethodGet, "/down") // Resty will not be able to send the request to the server.
	require.Error(t, err, "expect an error when the server is unreachable")
	assert.Nil(t, res, "result should be nil on execute error")

	// Assert that the returned error is of type Error and has Status 500: Internal Server Error
	var execErr Error
	if assert.ErrorAs(t, err, &execErr) {
		assert.Equal(t, http.StatusInternalServerError, execErr.Status, "status should be internal server error when execute returns an error")
		assert.NotEmpty(t, execErr.Body.String(), "error body should contain the error message")
	}
}

// TestStruct is an auxiliary struct in order to do a test with the Resty SetResult() function.
type TestStruct struct {
	Test bool   `json:"test"`
	Text string `string:"text"`
}

// TestExecute_ReturnsTypedResultWhenResultIsSet tests how Execute behaves when the client's
// Resty client had a SetResult() specified.
func TestExecute_ReturnsTypedResultWhenResultIsSet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/test", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"test":true,"text":"Text struct"}`))
	}))
	t.Cleanup(srv.Close)

	// Create client and configure a Result target via Resty's SetResult()
	c := newClient(srv.URL, "")
	c.client.OnBeforeRequest(func(_ *resty.Client, r *resty.Request) error {
		r.SetResult(&TestStruct{})
		return nil
	})

	res, err := c.execute(http.MethodGet, "/test")
	require.NoError(t, err)
	require.NotNil(t, res)

	test, ok := res.(*TestStruct)
	require.True(t, ok, "Expected the response to be marshaled automatically")
	require.True(t, test.Test, "Expected Test=true unmarshaled from JSON")
	assert.Equal(t, test.Text, "Text struct", "Expected Text=Text Struct")
}

package testutils

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHttpHandler tests that the HttpHandler function actually sets the information and runs the assertions.
func TestHttpHandler(t *testing.T) {
	expectedBody := `{"ok":true}`
	expectedContentType := "application/json"
	expectedStatus := http.StatusOK

	assertions := func(r *http.Request) {
		assert.Equal(t, "/test", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
	}

	handler := HttpHandler(assertions, expectedStatus, expectedContentType, expectedBody)
	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	response := responseRecorder.Result()
	actualBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, expectedStatus, response.StatusCode)
	assert.Equal(t, expectedContentType, response.Header.Get("Content-Type"))
	body := string(actualBody)
	assert.Equal(t, expectedBody, body)
}

// TestHttpNoContentHandler tests that the function actually returns a No Content response.
func TestHttpNoContentHandler(t *testing.T) {
	assertions := []func(*testing.T, *http.Request){
		func(t *testing.T, r *http.Request) {
			assert.Equal(t, "/nocontent", r.URL.Path)
		},
		func(t *testing.T, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
		},
	}

	handler := HttpNoContentHandler(t, assertions)
	request := httptest.NewRequest(http.MethodDelete, "/nocontent", nil)
	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	response := responseRecorder.Result()
	actualBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
	assert.Equal(t, 0, len(actualBody), "NoContent should not have a body")
}

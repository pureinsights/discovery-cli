package testutils

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHttpHandler_Table tests the HttpHandler function using table-driven tests.
func TestHttpHandler_Table(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		expectedCT     string
		expectedBody   string
		reqMethod      string
		reqPath        string
		assertions     func(test *testing.T, r *http.Request)
	}{
		{
			name:           "OK JSON response with assertions",
			expectedStatus: http.StatusOK,
			expectedCT:     "application/json",
			expectedBody:   `{"ok":true}`,
			reqMethod:      http.MethodGet,
			reqPath:        "/test",
			assertions: func(test *testing.T, r *http.Request) {
				assert.Equal(test, "/test", r.URL.Path)
				assert.Equal(test, http.MethodGet, r.Method)
			},
		},
		{
			name:           "OK JSON response with nil assertions",
			expectedStatus: http.StatusOK,
			expectedCT:     "application/json",
			expectedBody:   `{"ok":true}`,
			reqMethod:      http.MethodGet,
			reqPath:        "/test",
			assertions:     nil,
		},
		{
			name:           "Error JSON response with nil assertions",
			expectedStatus: http.StatusNotFound,
			expectedCT:     "application/json",
			expectedBody:   `{"error": "Not found"}`,
			reqMethod:      http.MethodGet,
			reqPath:        "/test",
			assertions:     nil,
		},
		{
			name:           "String response with assertions",
			expectedStatus: http.StatusNotFound,
			expectedCT:     "text/plain",
			expectedBody:   `This is a test response.`,
			reqMethod:      http.MethodGet,
			reqPath:        "/test",
			assertions: func(test *testing.T, r *http.Request) {
				assert.Equal(test, "/test", r.URL.Path)
				assert.Equal(test, http.MethodGet, r.Method)
			},
		},
		{
			name:           "Empty string response with no assertions",
			expectedStatus: http.StatusNoContent,
			expectedCT:     "text/plain",
			expectedBody:   ``,
			reqMethod:      http.MethodGet,
			reqPath:        "/test",
			assertions:     nil,
		},
		{
			name:           "String response with application/json content type and assertions",
			expectedStatus: http.StatusOK,
			expectedCT:     "application/json",
			expectedBody:   `This is a test response.`,
			reqMethod:      http.MethodGet,
			reqPath:        "/test",
			assertions: func(test *testing.T, r *http.Request) {
				assert.Equal(test, "/test", r.URL.Path)
				assert.Equal(test, http.MethodGet, r.Method)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := HttpHandler(t, tc.expectedStatus, tc.expectedCT, tc.expectedBody, tc.assertions)

			req := httptest.NewRequest(tc.reqMethod, tc.reqPath, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			actualBody, _ := io.ReadAll(res.Body)

			assert.Equal(t, tc.expectedStatus, res.StatusCode)
			assert.Equal(t, tc.expectedCT, res.Header.Get("Content-Type"))
			assert.Equal(t, tc.expectedBody, string(actualBody))
		})
	}
}

// TestHttpNoContentHandler tests that the function actually returns a No Content response.
func TestHttpNoContentHandler(t *testing.T) {
	assertions := func(test *testing.T, r *http.Request) {
		assert.Equal(t, "/nocontent", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)
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

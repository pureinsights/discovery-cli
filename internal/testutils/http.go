package testutils

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// HttpHandler returns a handler that performs given assertions and responds
// with the provided status, content type, and body.
func HttpHandler(
	t *testing.T,
	statusCode int,
	contentType string,
	body string,
	assertions func(*testing.T, *http.Request),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if assertions != nil {
			assertions(t, r)
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	}
}

func HttpHandlerWithContentDisposition(
	t *testing.T,
	apiKey string,
	statusCode int,
	method string,
	path string,
	contentType string,
	contentDisposition string,
	body []byte,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if apiKey != "" {
			assert.Equal(t, apiKey, r.Header.Get("X-API-KEY"))
		}
		assert.Equal(t, method, r.Method)
		assert.Equal(t, path, r.URL.Path)
		w.Header().Set("Content-Type", contentType)
		w.Header().Set(
			"Content-Disposition",
			contentDisposition,
		)
		w.WriteHeader(statusCode)
		w.Write(body)
	}
}

// HttpNoContentHandler returns a handler that performs given assertions and responds
// with a NoContent response.
func HttpNoContentHandler(
	t *testing.T,
	assertions func(*testing.T, *http.Request),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if assertions != nil {
			assertions(t, r)
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// MockResponse mocks an HTTP response from Discovery
type MockResponse struct {
	StatusCode  int
	ContentType string
	Body        string
	Assertions  func(*testing.T, *http.Request)
}

// HttpMultiResponseHandler returns different HTTP responses depending on the received path
func HttpMultiResponseHandler(
	t *testing.T,
	responses map[string]MockResponse,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for path, response := range responses {
			if r.Method+":"+r.URL.Path == path {
				if response.Assertions != nil {
					response.Assertions(t, r)
				}
				w.Header().Set("Content-Type", response.ContentType)
				w.WriteHeader(response.StatusCode)
				w.Write([]byte(response.Body))
				return
			}
		}
	}
}

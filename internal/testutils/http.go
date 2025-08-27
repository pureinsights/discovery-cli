package testutils

import (
	"net/http"
	"testing"
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

package discovery

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestCore has table-driven tests to test the cloner methods.
func TestCore(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		response   string
		testFunc   func(t *testing.T, c cloner)
	}{
		// Working case
		{
			name:       "Ping returns acknowledged true",
			method:     http.MethodGet,
			path:       "/v2/server/f6950327-3175-4a98-a570-658df852424a/ping",
			statusCode: http.StatusOK,
			response: `{
			"acknowledged": true
			}`,
			testFunc: func(t *testing.T, c cloner) {
				id := uuid.MustParse("f6950327-3175-4a98-a570-658df852424a")
				response, err := c.Clone(id, map[string][]string{"name": {"mongo2"}})
				require.NoError(t, err)
				assert.Equal(t, "mongo2", response.Get("name").String())
				assert.Equal(t, "mongo-secret", response.Get("secret").String())
			},
		},

		// Error case
		{
			name:       "Ping returns an error",
			method:     http.MethodGet,
			path:       "/v2/server/f6950327-3175-4a98-a570-658df852424a/ping",
			statusCode: http.StatusBadGateway,
			response: `{
					"status": 502,
					"code": 8002,
					"messages": [
							"An error occurred while pinging the Mongo client."
					],
					"timestamp": "2025-08-26T20:42:26.372708600Z"
			}`,
			testFunc: func(t *testing.T, c cloner) {
				id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
				response, err := c.Clone(id, map[string][]string{"name": {"mongo2"}})
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusNotFound, []byte(`{
					"status": 502,
					"code": 8002,
					"messages": [
							"An error occurred while pinging the Mongo client."
					],
					"timestamp": "2025-08-26T20:42:26.372708600Z"
			}`)))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(
				testutils.HttpHandler(func(r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, tc.path, r.URL.Path)
				},
					tc.statusCode, "application/json", tc.response)))
			defer srv.Close()

			c := cloner{client: newClient(srv.URL, "")}
			tc.testFunc(t, c)
		})
	}
}

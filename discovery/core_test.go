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

// TestCore has table-driven tests to test the core methods.
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
			path:       "/v2/server/f6950327-3175-4a98-a570-658df852424a/ping",
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
				testutils.HttpHandler(t,
					tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
						assert.Equal(t, tc.method, r.Method)
						assert.Equal(t, tc.path, r.URL.Path)
					}))
			defer srv.Close()

			coreClient := newClient(srv.URL, "")
			serverClient := newServersClient(coreClient)
			tc.testFunc(t, serverClient)
		})
	}
}

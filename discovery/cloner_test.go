package discovery

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestCloner has table-driven tests to test the cloner methods.
func TestCloner(t *testing.T) {
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
			name:       "Clone returns object",
			method:     http.MethodPost,
			path:       "/5f125024-1e5e-4591-9fee-365dc20eeeed/clone",
			statusCode: http.StatusOK,
			response: `{
			"type": "mongo",
			"name": "mongo2",
			"labels": [],
			"active": true,
			"id": "c77caced-b1d4-49de-b690-09bac3bc80a7",
			"creationTimestamp": "2025-08-21T15:19:37.980898Z",
			"lastUpdatedTimestamp": "2025-08-21T15:19:37.980898Z",
			"secret": "mongo-secret"
			}`,
			testFunc: func(t *testing.T, c cloner) {
				id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
				response, err := c.Clone(id, map[string][]string{"name": {"mongo2"}})
				require.NoError(t, err)
				assert.Equal(t, "mongo2", response.Get("name").String())
				assert.Equal(t, "mongo-secret", response.Get("secret").String())
			},
		},

		// Error cases
		{
			name:       "Get by ID returns 404 Not Found",
			method:     http.MethodPost,
			path:       "/5f125024-1e5e-4591-9fee-365dc20eeeed/clone",
			statusCode: http.StatusNotFound,
			response:   `{"messages": ["Entity not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`,
			testFunc: func(t *testing.T, c cloner) {
				id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
				response, err := c.Clone(id, map[string][]string{"name": {"mongo2"}})
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusNotFound, []byte(`{"messages": ["Entity not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`)))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, tc.path, r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.response))
			}))
			defer srv.Close()

			c := cloner{client: newClient(srv.URL, "")}
			tc.testFunc(t, c)
		})
	}
}

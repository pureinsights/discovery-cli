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

// TestEnabler has table-driven tests to test the enabler methods.
func TestEnabler(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		response   string
		testFunc   func(t *testing.T, e enabler)
	}{
		// Working cases
		{
			name:       "Enable returns true",
			method:     http.MethodPatch,
			path:       "/5f125024-1e5e-4591-9fee-365dc20eeeed/enable",
			statusCode: http.StatusOK,
			response: `{
			"acknowledged": true
			}`,
			testFunc: func(t *testing.T, e enabler) {
				id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
				response, err := e.Enable(id)
				require.NoError(t, err)
				assert.True(t, response.Get("acknowledged").Bool())
			},
		},
		{
			name:       "Enable returns true",
			method:     http.MethodPatch,
			path:       "/5f125024-1e5e-4591-9fee-365dc20eeeed/disable",
			statusCode: http.StatusOK,
			response: `{
			"acknowledged": true
			}`,
			testFunc: func(t *testing.T, e enabler) {
				id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
				response, err := e.Disable(id)
				require.NoError(t, err)
				assert.True(t, response.Get("acknowledged").Bool())
			},
		},

		// Error case
		{
			name:       "Enable returns 404 Not Found",
			method:     http.MethodPatch,
			path:       "/5f125024-1e5e-4591-9fee-365dc20eeeed/enable",
			statusCode: http.StatusNotFound,
			response:   `{"messages": ["Entity not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`,
			testFunc: func(t *testing.T, e enabler) {
				id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
				response, err := e.Enable(id)
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusNotFound, []byte(`{"messages": ["Entity not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`)))
			},
		},
		{
			name:       "Disable returns 404 Not Found",
			method:     http.MethodPatch,
			path:       "/5f125024-1e5e-4591-9fee-365dc20eeeed/disable",
			statusCode: http.StatusNotFound,
			response:   `{"messages": ["Entity not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`,
			testFunc: func(t *testing.T, e enabler) {
				id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
				response, err := e.Disable(id)
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusNotFound, []byte(`{"messages": ["Entity not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`)))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, tc.path, r.URL.Path)
				}))
			defer srv.Close()

			e := enabler{client: newClient(srv.URL, "")}
			tc.testFunc(t, e)
		})
	}
}

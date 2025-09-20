package discovery

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Test_enabler_Enable has table-driven tests to test the enabler.Enable() method.
func Test_enabler_Enable(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		response         string
		expectedResponse gjson.Result
		err              error
		testFunc         func(t *testing.T, e enabler)
	}{
		// Working case
		{
			name:       "Enable returns true",
			method:     http.MethodPatch,
			path:       "/5f125024-1e5e-4591-9fee-365dc20eeeed/enable",
			statusCode: http.StatusOK,
			response: `{
			"acknowledged": true
			}`,
			expectedResponse: gjson.Parse(`{
			"acknowledged": true
			}`),
			err: nil,
		},

		// Error case
		{
			name:             "Enable returns 404 Not Found",
			method:           http.MethodPatch,
			path:             "/5f125024-1e5e-4591-9fee-365dc20eeeed/enable",
			statusCode:       http.StatusNotFound,
			response:         `{"messages": ["Entity not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`,
			expectedResponse: gjson.Result{},
			err:              Error{Status: http.StatusNotFound, Body: gjson.Parse(`{"messages": ["Entity not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`)},
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
			id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
			response, err := e.Enable(id)
			assert.Equal(t, tc.expectedResponse, response)
			if tc.err == nil {
				require.NoError(t, err)
				assert.True(t, response.IsObject())
			} else {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

// Test_enabler_Disable has table-driven tests to test the enabler.Disable() method.
func Test_enabler_Disable(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		response         string
		testFunc         func(t *testing.T, e enabler)
		expectedResponse gjson.Result
		err              error
	}{
		// Working case
		{
			name:       "Disable returns true",
			method:     http.MethodPatch,
			path:       "/5f125024-1e5e-4591-9fee-365dc20eeeed/disable",
			statusCode: http.StatusOK,
			response: `{
			"acknowledged": true
			}`,
			expectedResponse: gjson.Parse(`{
			"acknowledged": true
			}`),
			err: nil,
		},

		// Error case
		{
			name:             "Disable returns 404 Not Found",
			method:           http.MethodPatch,
			path:             "/5f125024-1e5e-4591-9fee-365dc20eeeed/disable",
			statusCode:       http.StatusNotFound,
			response:         `{"messages": ["Entity not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`,
			expectedResponse: gjson.Result{},
			err:              Error{Status: http.StatusNotFound, Body: gjson.Parse(`{"messages": ["Entity not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`)},
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
			id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
			response, err := e.Disable(id)
			assert.Equal(t, tc.expectedResponse, response)
			if tc.err == nil {
				require.NoError(t, err)
				assert.True(t, response.IsObject())
			} else {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

package discovery

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Test_cloner_Clone has table-driven tests to test the cloner.Clone() method.
func Test_cloner_Clone(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		response         string
		expectedResponse gjson.Result
		queryParameters  map[string][]string
		err              error
	}{
		// Working cases
		{
			name:       "Clone with query parameters returns object",
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
			expectedResponse: gjson.Parse(`{
			"type": "mongo",
			"name": "mongo2",
			"labels": [],
			"active": true,
			"id": "c77caced-b1d4-49de-b690-09bac3bc80a7",
			"creationTimestamp": "2025-08-21T15:19:37.980898Z",
			"lastUpdatedTimestamp": "2025-08-21T15:19:37.980898Z",
			"secret": "mongo-secret"
			}`),
			queryParameters: map[string][]string{"name": {"mongo2"}},
			err:             nil,
		},
		{
			name:       "Clone with no query parameters returns object",
			method:     http.MethodPost,
			path:       "/5f125024-1e5e-4591-9fee-365dc20eeeed/clone",
			statusCode: http.StatusOK,
			response: `{
			"type": "mongo",
			"name": "mongo-clone",
			"labels": [],
			"active": true,
			"id": "c77caced-b1d4-49de-b690-09bac3bc80a7",
			"creationTimestamp": "2025-08-21T15:19:37.980898Z",
			"lastUpdatedTimestamp": "2025-08-21T15:19:37.980898Z",
			"secret": "mongo-secret"
			}`,
			expectedResponse: gjson.Parse(`{
			"type": "mongo",
			"name": "mongo-clone",
			"labels": [],
			"active": true,
			"id": "c77caced-b1d4-49de-b690-09bac3bc80a7",
			"creationTimestamp": "2025-08-21T15:19:37.980898Z",
			"lastUpdatedTimestamp": "2025-08-21T15:19:37.980898Z",
			"secret": "mongo-secret"
			}`),
			queryParameters: map[string][]string(nil),
			err:             nil,
		},

		// Error cases
		{
			name:       "Clone with query parameters returns 409 Conflict",
			method:     http.MethodPost,
			path:       "/5f125024-1e5e-4591-9fee-365dc20eeeed/clone",
			statusCode: http.StatusConflict,
			response: `{
			"status": 409,
			"code": 2001,
			"messages": [
				"Duplicate entry for field(s): name"
			],
			"timestamp": "2025-09-22T14:51:10.223790900Z"
			}`,
			expectedResponse: gjson.Result{},
			queryParameters:  map[string][]string{"name": {"mongo2"}},
			err: Error{Status: http.StatusConflict, Body: gjson.Parse(`{
			"status": 409,
			"code": 2001,
			"messages": [
				"Duplicate entry for field(s): name"
			],
			"timestamp": "2025-09-22T14:51:10.223790900Z"
			}`)},
		},
		{
			name:             "Clone with no query parameters returns 404 Not Found",
			method:           http.MethodPost,
			path:             "/5f125024-1e5e-4591-9fee-365dc20eeeed/clone",
			statusCode:       http.StatusNotFound,
			response:         `{"messages": ["Entity not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`,
			expectedResponse: gjson.Result{},
			queryParameters:  map[string][]string(nil),
			err:              Error{Status: http.StatusNotFound, Body: gjson.Parse(`{"messages": ["Entity not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, tc.path, r.URL.Path)
				for k, v := range tc.queryParameters {
					assert.Equal(t, v, r.URL.Query()[k])
				}
			}))
			defer srv.Close()

			c := cloner{client: newClient(srv.URL, "")}
			id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
			response, err := c.Clone(id, tc.queryParameters)
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

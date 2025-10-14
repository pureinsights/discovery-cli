package cli

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

// GetEntity obtains the entity with the given ID using the given client and then prints out the result using the received printer or the JSON printer.
func Test_discovery_GetEntity(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		response         string
		expectedResponse gjson.Result
		err              error
	}{
		// Working case
		{
			name:             "Get by ID returns object",
			method:           http.MethodGet,
			path:             "/5f125024-1e5e-4591-9fee-365dc20eeeed",
			statusCode:       http.StatusOK,
			response:         `{"id":"5f125024-1e5e-4591-9fee-365dc20eeeed","name":"test-secret"}`,
			expectedResponse: gjson.Parse(`{"id":"5f125024-1e5e-4591-9fee-365dc20eeeed","name":"test-secret"}`),
			err:              nil,
		},

		// Error case
		{
			name:             "Get by ID returns 404 Not Found",
			method:           http.MethodGet,
			path:             "/5f125024-1e5e-4591-9fee-365dc20eeeed",
			statusCode:       http.StatusNotFound,
			response:         `{"messages": ["Secret not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`,
			expectedResponse: gjson.Result{},
			err:              errors.New("fail"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, tc.path, r.URL.Path)
			}))

			defer srv.Close()

			// c := crud{getter{newClient(srv.URL, "")}}
			// id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
			// response, err := c.Get(id)
			// assert.Equal(t, tc.expectedResponse, response)
			// if tc.err == nil {
			// 	require.NoError(t, err)
			// 	assert.True(t, response.IsObject())
			// } else {
			// 	var errStruct Error
			// 	require.ErrorAs(t, err, &errStruct)
			// 	assert.EqualError(t, err, tc.err.Error())
			// }
		})
	}
}

// GetEntities obtains all of the entities using the given client and then prints out the result using the received printer or the JSON array printer.
func Test_discovery_GetEntities(t *testing.T) {

}

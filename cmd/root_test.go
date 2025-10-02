package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/spf13/viper"
)

func Test_newRootCommand(t *testing.T) {
	in := strings.NewReader("In Reader")
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: out,
		Err: errBuf,
	}

	dir := t.TempDir()
	d := cli.NewDiscovery(&ios, viper.New(), dir)
	discoveryCmd := newRootCommand(d)

}

func TestRun(t *testing.T) {
	// tests := []struct {
	// 	name             string
	// 	method           string
	// 	path             string
	// 	statusCode       int
	// 	response         string
	// 	expectedResponse gjson.Result
	// 	err              error
	// }{
	// 	// Working case
	// 	{
	// 		name:       "Ping returns acknowledged true",
	// 		method:     http.MethodGet,
	// 		path:       "/server/f6950327-3175-4a98-a570-658df852424a/ping",
	// 		statusCode: http.StatusOK,
	// 		response: `{
	// 		"acknowledged": true
	// 		}`,
	// 		expectedResponse: gjson.Parse(`{
	// 		"acknowledged": true
	// 		}`),
	// 		err: nil,
	// 	},

	// 	// Error case
	// 	{
	// 		name:       "Ping returns a 502 error",
	// 		method:     http.MethodGet,
	// 		path:       "/server/f6950327-3175-4a98-a570-658df852424a/ping",
	// 		statusCode: http.StatusBadGateway,
	// 		response: `{
	// 				"status": 502,
	// 				"code": 8002,
	// 				"messages": [
	// 						"An error occurred while pinging the Mongo client."
	// 				],
	// 				"timestamp": "2025-08-26T20:42:26.372708600Z"
	// 		}`,
	// 		expectedResponse: gjson.Result{},
	// 		err: Error{Status: http.StatusBadGateway, Body: gjson.Parse(`{
	// 				"status": 502,
	// 				"code": 8002,
	// 				"messages": [
	// 						"An error occurred while pinging the Mongo client."
	// 				],
	// 				"timestamp": "2025-08-26T20:42:26.372708600Z"
	// 		}`)},
	// 	},
	// 	{
	// 		name:       "Ping returns a 400 error",
	// 		method:     http.MethodGet,
	// 		path:       "/server/f6950327-3175-4a98-a570-658df852424a/ping",
	// 		statusCode: http.StatusBadRequest,
	// 		response: `{
	// 		"status": 400,
	// 		"code": 3002,
	// 		"messages": [
	// 				"Failed to convert argument [id] for value [notuuid] due to: Invalid UUID string: notuuid"
	// 		],
	// 		"timestamp": "2025-09-30T15:35:00.121829500Z"
	// 		}`,
	// 		expectedResponse: gjson.Result{},
	// 		err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
	// 		"status": 400,
	// 		"code": 3002,
	// 		"messages": [
	// 				"Failed to convert argument [id] for value [notuuid] due to: Invalid UUID string: notuuid"
	// 		],
	// 		"timestamp": "2025-09-30T15:35:00.121829500Z"
	// 		}`)},
	// 	},
	// 	{
	// 		name:       "Ping returns a 422 error",
	// 		method:     http.MethodGet,
	// 		path:       "/server/f6950327-3175-4a98-a570-658df852424a/ping",
	// 		statusCode: http.StatusUnprocessableEntity,
	// 		response: `{
	// 		"status": 422,
	// 		"code": 4001,
	// 		"messages": [
	// 				"Client of type openai cannot be validated"
	// 		],
	// 		"timestamp": "2025-09-30T15:35:00.121829500Z"
	// 		}`,
	// 		expectedResponse: gjson.Result{},
	// 		err: Error{Status: http.StatusUnprocessableEntity, Body: gjson.Parse(`{
	// 		"status": 422,
	// 		"code": 4001,
	// 		"messages": [
	// 				"Client of type openai cannot be validated"
	// 		],
	// 		"timestamp": "2025-09-30T15:35:00.121829500Z"
	// 		}`)},
	// 	},
	// 	{
	// 		name:       "Ping returns a 404 error",
	// 		method:     http.MethodGet,
	// 		path:       "/server/f6950327-3175-4a98-a570-658df852424a/ping",
	// 		statusCode: http.StatusNotFound,
	// 		response: `{
	// 		"status": 404,
	// 		"code": 1003,
	// 		"messages": [
	// 			"Entity not found: f6950327-3175-4a98-a570-658df852424a"
	// 		],
	// 		"timestamp": "2025-09-30T15:38:42.885125200Z"
	// 		}`,
	// 		expectedResponse: gjson.Result{},
	// 		err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
	// 		"status": 404,
	// 		"code": 1003,
	// 		"messages": [
	// 			"Entity not found: f6950327-3175-4a98-a570-658df852424a"
	// 		],
	// 		"timestamp": "2025-09-30T15:38:42.885125200Z"
	// 		}`)},
	// 	},
	// }

	// for _, tc := range tests {
	// 	t.Run(tc.name, func(t *testing.T) {
	// 		srv := httptest.NewServer(
	// 			testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
	// 				assert.Equal(t, tc.method, r.Method)
	// 				assert.Equal(t, tc.path, r.URL.Path)
	// 			}))
	// 		defer srv.Close()

	// 		serverClient := newServersClient(srv.URL, "")
	// 		id := uuid.MustParse("f6950327-3175-4a98-a570-658df852424a")
	// 		response, err := serverClient.Ping(id)
	// 		assert.Equal(t, tc.expectedResponse, response)
	// 		if tc.err == nil {
	// 			require.NoError(t, err)
	// 			assert.True(t, response.IsObject())
	// 		} else {
	// 			var errStruct Error
	// 			require.ErrorAs(t, err, &errStruct)
	// 			assert.EqualError(t, err, tc.err.Error())
	// 		}
	// 	})
	// }
}

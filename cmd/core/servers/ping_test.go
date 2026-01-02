package servers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestNewPingCommand tests the NewPingCommand function.
func TestNewPingCommand(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		url       bool
		apiKey    string
		outGolden string
		errGolden string
		outBytes  []byte
		errBytes  []byte
		responses map[string]testutils.MockResponse
		err       error
	}{
		// Working case
		{
			name:      "Ping by ID returns an acknowledged true",
			args:      []string{"3d51beef-8b90-40aa-84b5-033241dc6239"},
			url:       true,
			apiKey:    "",
			outGolden: "NewPingCommand_Out_PingByIdReturnsObject",
			errGolden: "NewPingCommand_Err_PingByIdReturnsObject",
			outBytes:  testutils.Read(t, "NewPingCommand_Out_PingByIdReturnsObject"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/server/search": {
					StatusCode: http.StatusNoContent,
					Body: `{
			"content": [],
			"pageable": {
				"page": 0,
				"size": 25,
				"sort": []
			},
			"totalSize": 1,
			"totalPages": 1,
			"empty": false,
			"size": 25,
			"offset": 0,
			"numberOfElements": 0,
			"pageNumber": 0
			}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/server/search", r.URL.Path)
					},
				},
				"GET:/v2/server/3d51beef-8b90-40aa-84b5-033241dc6239": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "my-server",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "3d51beef-8b90-40aa-84b5-033241dc6239",
					"creationTimestamp": "2025-10-17T22:37:57Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:57Z"
				}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/server/3d51beef-8b90-40aa-84b5-033241dc6239", r.URL.Path)
					},
				},
				"GET:/v2/server/3d51beef-8b90-40aa-84b5-033241dc6239/ping": {
					StatusCode: http.StatusOK,
					Body: `{
				"acknowledged": true
			}`,

					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/server/3d51beef-8b90-40aa-84b5-033241dc6239/ping", r.URL.Path)
					},
				},
			},
			err: nil,
		},
		{
			name:      "Ping by name returns an acknowledged true",
			args:      []string{"my-server"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewPingCommand_Out_PingByNameReturnsObject",
			errGolden: "NewPingCommand_Err_PingByNameReturnsObject",
			outBytes:  testutils.Read(t, "NewPingCommand_Out_PingByNameReturnsObject"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/server/search": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "my-server",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "3d51beef-8b90-40aa-84b5-033241dc6239",
					"creationTimestamp": "2025-10-17T22:37:57Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:57Z"
				},
				"highlight": {}
			],
			"pageable": {
				"page": 0,
				"size": 25,
				"sort": []
			},
			"totalSize": 18,
			"totalPages": 1,
			"empty": false,
			"size": 25,
			"offset": 0,
			"numberOfElements": 18,
			"pageNumber": 0
			}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/server/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"GET:/v2/server/3d51beef-8b90-40aa-84b5-033241dc6239": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "my-server",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "3d51beef-8b90-40aa-84b5-033241dc6239",
					"creationTimestamp": "2025-10-17T22:37:57Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:57Z"
				}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/server/3d51beef-8b90-40aa-84b5-033241dc6239", r.URL.Path)
					},
				},
				"GET:/v2/server/3d51beef-8b90-40aa-84b5-033241dc6239/ping": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
				"acknowledged": true
			}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/server/3d51beef-8b90-40aa-84b5-033241dc6239/ping", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			args:      []string{"my-server"},
			outGolden: "NewPingCommand_Out_NoURL",
			errGolden: "NewPingCommand_Err_NoURL",
			outBytes:  testutils.Read(t, "NewPingCommand_Out_NoURL"),
			errBytes:  testutils.Read(t, "NewPingCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:      "sent name does not exist",
			args:      []string{"test"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewPingCommand_Out_NameDoesNotExist",
			errGolden: "NewPingCommand_Err_NameDoesNotExist",
			outBytes:  testutils.Read(t, "NewPingCommand_Out_NameDoesNotExist"),
			errBytes:  testutils.Read(t, "NewPingCommand_Err_NameDoesNotExist"),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/server/search": {
					StatusCode:  http.StatusNoContent,
					Body:        ``,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/server/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name "test" does not exist"
	]
}`),
			}, "Could not get server ID."),
		},
		{
			name:      "Printing JSON object fails",
			args:      []string{"my-server"},
			outGolden: "NewPingCommand_Out_PrintJSONFails",
			errGolden: "NewPingCommand_Err_PrintJSONFails",
			outBytes:  testutils.Read(t, "NewPingCommand_Out_PrintJSONFails"),
			errBytes:  testutils.Read(t, "NewPingCommand_Err_PrintJSONFails"),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/server/search": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "my-server",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "3d51beef-8b90-40aa-84b5-033241dc6239",
					"creationTimestamp": "2025-10-17T22:37:57Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:57Z"
				},
				"highlight": {}
			],
			"pageable": {
				"page": 0,
				"size": 25,
				"sort": []
			},
			"totalSize": 1,
			"totalPages": 1,
			"empty": false,
			"size": 25,
			"offset": 0,
			"numberOfElements": 1,
			"pageNumber": 0
			}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/server/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"GET:/v2/server/3d51beef-8b90-40aa-84b5-033241dc6239": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "my-server",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "3d51beef-8b90-40aa-84b5-033241dc6239",
					"creationTimestamp": "2025-10-17T22:37:57Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:57Z"
				}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/server/3d51beef-8b90-40aa-84b5-033241dc6239", r.URL.Path)
					},
				},
				"GET:/v2/server/3d51beef-8b90-40aa-84b5-033241dc6239/ping": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
				"acknowledged: true
			}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/server/3d51beef-8b90-40aa-84b5-033241dc6239/ping", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character '\\n' in string literal"), "Could not print JSON object"),
		},
		{
			name:      "Ping by ID fails",
			args:      []string{"3d51beef-8b90-40aa-84b5-033241dc6239"},
			url:       true,
			apiKey:    "",
			outGolden: "NewPingCommand_Out_PingByIdFails",
			errGolden: "NewPingCommand_Err_PingByIdFails",
			outBytes:  testutils.Read(t, "NewPingCommand_Out_PingByIdFails"),
			errBytes:  testutils.Read(t, "NewPingCommand_Err_PingByIdFails"),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/server/search": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [],
			"pageable": {
				"page": 0,
				"size": 25,
				"sort": []
			},
			"totalSize": 1,
			"totalPages": 1,
			"empty": false,
			"size": 25,
			"offset": 0,
			"numberOfElements": 0,
			"pageNumber": 0
			}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/server/search", r.URL.Path)
					},
				},
				"GET:/v2/server/3d51beef-8b90-40aa-84b5-033241dc6239": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "my-server",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "3d51beef-8b90-40aa-84b5-033241dc6239",
					"creationTimestamp": "2025-10-17T22:37:57Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:57Z"
				}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/server/3d51beef-8b90-40aa-84b5-033241dc6239", r.URL.Path)
					},
				},
				"GET:/v2/server/3d51beef-8b90-40aa-84b5-033241dc6239/ping": {
					StatusCode: http.StatusBadGateway,
					Body: `{
  "status": 502,
  "code": 8002,
  "messages": [
    "An error occurred while pinging the Mongo client."
  ],
  "timestamp": "2025-12-18T17:03:39.578033300Z"
}`,

					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/server/3d51beef-8b90-40aa-84b5-033241dc6239/ping", r.URL.Path)
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusBadGateway,
				Body: gjson.Parse(`{
  "status": 502,
  "code": 8002,
  "messages": [
    "An error occurred while pinging the Mongo client."
  ],
  "timestamp": "2025-12-18T17:03:39.578033300Z"
}`),
			}, "Could not ping server with id \"3d51beef-8b90-40aa-84b5-033241dc6239\""),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpMultiResponseHandler(t, tc.responses))

			defer srv.Close()

			in := strings.NewReader("")
			out := &bytes.Buffer{}

			errBuf := &bytes.Buffer{}
			ios := iostreams.IOStreams{
				In:  in,
				Out: out,
				Err: errBuf,
			}

			vpr := viper.New()
			vpr.Set("profile", "default")
			vpr.Set("output", "pretty-json")
			if tc.url {
				vpr.Set("default.core_url", srv.URL)
			}
			if tc.apiKey != "" {
				vpr.Set("default.core_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, t.TempDir())

			pingCmd := NewPingCommand(d)

			pingCmd.SetIn(ios.In)
			pingCmd.SetOut(ios.Out)
			pingCmd.SetErr(ios.Err)

			pingCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			pingCmd.SetArgs(tc.args)

			err := pingCmd.Execute()
			if tc.err != nil {
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
				testutils.CompareBytes(t, tc.errGolden, tc.errBytes, errBuf.Bytes())
			} else {
				require.NoError(t, err)
			}

			testutils.CompareBytes(t, tc.outGolden, tc.outBytes, out.Bytes())
		})
	}
}

// TestNewPingCommand_NoProfileFlag tests the NewPingCommand when the profile flag was not defined.
func TestNewPingCommand_NoProfileFlag(t *testing.T) {
	in := strings.NewReader("")
	out := &bytes.Buffer{}

	errBuf := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: out,
		Err: errBuf,
	}

	vpr := viper.New()
	vpr.Set("profile", "default")
	vpr.Set("output", "pretty-json")

	vpr.Set("default.core_url", "test")
	vpr.Set("default.core_key", "test")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	pingCmd := NewPingCommand(d)

	pingCmd.SetIn(ios.In)
	pingCmd.SetOut(ios.Out)
	pingCmd.SetErr(ios.Err)

	pingCmd.SetArgs([]string{"my-server"})

	err := pingCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewPingCommand_Out_NoProfile", testutils.Read(t, "NewPingCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewPingCommand_Err_NoProfile", testutils.Read(t, "NewPingCommand_Err_NoProfile"), errBuf.Bytes())
}

// TestNewPingCommand_NotExactly1Arg tests the NewPingCommand function when it does not receive exactly one argument.
func TestNewPingCommand_NotExactly1Arg(t *testing.T) {
	in := strings.NewReader("")
	out := &bytes.Buffer{}

	errBuf := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: out,
		Err: errBuf,
	}

	vpr := viper.New()
	vpr.Set("profile", "default")
	vpr.Set("output", "pretty-json")

	vpr.Set("default.core_url", "test")
	vpr.Set("default.core_key", "test")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	pingCmd := NewPingCommand(d)

	pingCmd.SetIn(ios.In)
	pingCmd.SetOut(ios.Out)
	pingCmd.SetErr(ios.Err)

	pingCmd.SetArgs([]string{})

	err := pingCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")

	testutils.CompareBytes(t, "NewPingCommand_Out_NotExactly1Arg", testutils.Read(t, "NewPingCommand_Out_NotExactly1Arg"), out.Bytes())
	testutils.CompareBytes(t, "NewPingCommand_Err_NotExactly1Arg", testutils.Read(t, "NewPingCommand_Err_NotExactly1Arg"), errBuf.Bytes())
}

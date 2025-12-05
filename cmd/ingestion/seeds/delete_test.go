package seeds

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

// TestNewDeleteCommand tests the NewDeleteCommand function
func TestNewDeleteCommand(t *testing.T) {
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
			name:      "Delete by ID returns an acknowledged true",
			args:      []string{"my-seed"},
			url:       true,
			apiKey:    "",
			outGolden: "NewDeleteCommand_Out_DeleteByIdReturnsObject",
			errGolden: "NewDeleteCommand_Err_DeleteByIdReturnsObject",
			outBytes:  testutils.Read(t, "NewDeleteCommand_Out_DeleteByIdReturnsObject"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "my-seed",
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
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
					},
				},
				"GET:/v2/seed/3d51beef-8b90-40aa-84b5-033241dc6239": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "my-seed",
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
						assert.Equal(t, "/v2/seed/3d51beef-8b90-40aa-84b5-033241dc6239", r.URL.Path)
					},
				},
				"DELETE:/v2/seed/3d51beef-8b90-40aa-84b5-033241dc6239": {
					StatusCode: http.StatusOK,
					Body: `{
				"acknowledged": true
			}`,

					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodDelete, r.Method)
						assert.Equal(t, "/v2/seed/3d51beef-8b90-40aa-84b5-033241dc6239", r.URL.Path)
					},
				},
			},
			err: nil,
		},
		{
			name:      "Delete by name returns an acknowledged true",
			args:      []string{"my-seed"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewDeleteCommand_Out_DeleteByIdReturnsObject",
			errGolden: "NewDeleteCommand_Err_DeleteByIdReturnsObject",
			outBytes:  testutils.Read(t, "NewDeleteCommand_Out_DeleteByIdReturnsObject"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "my-seed",
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
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"GET:/v2/seed/3d51beef-8b90-40aa-84b5-033241dc6239": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "my-seed",
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
						assert.Equal(t, "/v2/seed/3d51beef-8b90-40aa-84b5-033241dc6239", r.URL.Path)
					},
				},
				"DELETE:/v2/seed/3d51beef-8b90-40aa-84b5-033241dc6239": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
				"acknowledged": true
			}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodDelete, r.Method)
						assert.Equal(t, "/v2/seed/3d51beef-8b90-40aa-84b5-033241dc6239", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			args:      []string{"my-seed"},
			outGolden: "NewDeleteCommand_Out_NoURL",
			errGolden: "NewDeleteCommand_Err_NoURL",
			outBytes:  testutils.Read(t, "NewDeleteCommand_Out_NoURL"),
			errBytes:  testutils.Read(t, "NewDeleteCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Ingestion URL is missing for profile \"default\".\nTo set the URL for the Discovery Ingestion API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery ingestion config --profile \"default\""),
		},
		{
			name:      "sent name does not exist",
			args:      []string{"test"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewDeleteCommand_Out_NameDoesNotExist",
			errGolden: "NewDeleteCommand_Err_NameDoesNotExist",
			outBytes:  testutils.Read(t, "NewDeleteCommand_Out_NameDoesNotExist"),
			errBytes:  testutils.Read(t, "NewDeleteCommand_Err_NameDoesNotExist"),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
					StatusCode:  http.StatusNoContent,
					Body:        ``,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
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
			}, "Could not search for entity with name \"test\""),
		},
		{
			name:      "Printing JSON object fails",
			args:      []string{"my-seed"},
			outGolden: "NewDeleteCommand_Out_PrintJSONFails",
			errGolden: "NewDeleteCommand_Err_PrintJSONFails",
			outBytes:  testutils.Read(t, "NewDeleteCommand_Out_PrintJSONFails"),
			errBytes:  testutils.Read(t, "NewDeleteCommand_Err_PrintJSONFails"),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "my-seed",
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
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"GET:/v2/seed/3d51beef-8b90-40aa-84b5-033241dc6239": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "my-seed",
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
						assert.Equal(t, "/v2/seed/3d51beef-8b90-40aa-84b5-033241dc6239", r.URL.Path)
					},
				},
				"DELETE:/v2/seed/3d51beef-8b90-40aa-84b5-033241dc6239": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
				"acknowledged: true
			}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodDelete, r.Method)
						assert.Equal(t, "/v2/seed/3d51beef-8b90-40aa-84b5-033241dc6239", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character '\\n' in string literal"), "Could not print JSON object"),
		},
		{
			name:      "Search returns invalid UUID error",
			args:      []string{"test"},
			outGolden: "NewDeleteCommand_Out_InvalidUUID",
			errGolden: "NewDeleteCommand_Err_InvalidUUID",
			outBytes:  testutils.Read(t, "NewDeleteCommand_Out_InvalidUUID"),
			errBytes:  testutils.Read(t, "NewDeleteCommand_Err_InvalidUUID"),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "test",
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
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"GET:/v2/seed/3d51beef-8b90-40aa-84b5-033241dc6239": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "test",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "test",
					"creationTimestamp": "2025-10-17T22:37:57Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:57Z"
				}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/3d51beef-8b90-40aa-84b5-033241dc6239", r.URL.Path)
					},
				},
				"DELETE:/v2/seed/3d51beef-8b90-40aa-84b5-033241dc6239": {
					StatusCode: http.StatusOK,
					Body: `{
				"acknowledged": true
			}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodDelete, r.Method)
						assert.Equal(t, "/v2/seed/3d51beef-8b90-40aa-84b5-033241dc6239", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid UUID length: 4"), "Could not delete entity with name \"test\""),
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
				vpr.Set("default.ingestion_url", srv.URL)
			}
			if tc.apiKey != "" {
				vpr.Set("default.ingestion_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, t.TempDir())

			deleteCmd := NewDeleteCommand(d)

			deleteCmd.SetIn(ios.In)
			deleteCmd.SetOut(ios.Out)
			deleteCmd.SetErr(ios.Err)

			deleteCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			deleteCmd.SetArgs(tc.args)

			err := deleteCmd.Execute()
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

// TestNewDeleteCommand_NoProfileFlag tests the NewDeleteCommand when the profile flag was not defined.
func TestNewDeleteCommand_NoProfileFlag(t *testing.T) {
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

	vpr.Set("default.ingestion_url", "test")
	vpr.Set("default.ingestion_key", "test")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	deleteCmd := NewDeleteCommand(d)

	deleteCmd.SetIn(ios.In)
	deleteCmd.SetOut(ios.Out)
	deleteCmd.SetErr(ios.Err)

	deleteCmd.SetArgs([]string{"my-seed"})

	err := deleteCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewDeleteCommand_Out_NoProfile", testutils.Read(t, "NewDeleteCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewDeleteCommand_Err_NoProfile", testutils.Read(t, "NewDeleteCommand_Err_NoProfile"), errBuf.Bytes())
}

// TestNewDeleteCommand_NotExactly1Arg tests the NewDeleteCommand function when it does not receive exactly one argument.
func TestNewDeleteCommand_NotExactly1Arg(t *testing.T) {
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

	vpr.Set("default.ingestion_url", "test")
	vpr.Set("default.ingestion_key", "test")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	deleteCmd := NewDeleteCommand(d)

	deleteCmd.SetIn(ios.In)
	deleteCmd.SetOut(ios.Out)
	deleteCmd.SetErr(ios.Err)

	deleteCmd.SetArgs([]string{})

	err := deleteCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")

	testutils.CompareBytes(t, "NewDeleteCommand_Out_NotExactly1Arg", testutils.Read(t, "NewDeleteCommand_Out_NotExactly1Arg"), out.Bytes())
	testutils.CompareBytes(t, "NewDeleteCommand_Err_NotExactly1Arg", testutils.Read(t, "NewDeleteCommand_Err_NotExactly1Arg"), errBuf.Bytes())
}

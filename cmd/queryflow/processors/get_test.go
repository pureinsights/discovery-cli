package processors

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

// TestNewGetCommand tests the NewGetCommand() function.
func TestNewGetCommand(t *testing.T) {
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
			name:      "Search by name returns an array of which the first object is returned",
			args:      []string{"my-processor"},
			url:       true,
			apiKey:    "",
			outGolden: "NewGetCommand_Out_SearchByNameReturnsObject",
			errGolden: "NewGetCommand_Err_SearchByNameReturnsObject",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_SearchByNameReturnsObject"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/processor/search": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "my-processor",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "3b32e410-2f33-412d-9fb8-17970131921c",
					"creationTimestamp": "2025-10-17T22:37:57Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:57Z"
				},
				"highlight": {}
				"score": 1.4854797
				},
				{
				"source": {
					"type": "mongo",
					"name": "my-processor",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "4957145b-6192-4862-a5da-e97853974e9f",
					"creationTimestamp": "2025-10-17T22:37:53Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
				},
				"highlight": {
					"name": [
					"<em>label</em> <em>test</em> 1 <em>clone</em>"
					]
				},
				"score": 0.3980717
				}
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
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/processor/search", r.URL.Path)
					},
				},
				"GET:/v2/processor/3b32e410-2f33-412d-9fb8-17970131921c": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "my-processor",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "4957145b-6192-4862-a5da-e97853974e9f",
					"creationTimestamp": "2025-10-17T22:37:53Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
				}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/processor/3b32e410-2f33-412d-9fb8-17970131921c", r.URL.Path)
					},
				},
			},
			err: nil,
		},
		{
			name:      "Get with no args returns an array",
			args:      []string{},
			outGolden: "NewGetCommand_Out_GetAllReturnsArray",
			errGolden: "NewGetCommand_Err_GetAllReturnsArray",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_GetAllReturnsArray"),
			errBytes:  []byte(nil),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"GET:/v2/processor": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
				"type": "mongo",
				"name": "my-processor",
				"labels": [
					{
					"key": "A",
					"value": "A"
					}
				],
				"active": true,
				"id": "3b32e410-2f33-412d-9fb8-17970131921c",
				"creationTimestamp": "2025-10-17T22:37:57Z",
				"lastUpdatedTimestamp": "2025-10-17T22:37:57Z"
				},
				{
				"type": "openai",
				"name": "OpenAI processor",
				"labels": [],
				"active": true,
				"id": "5c09589e-b643-41aa-a766-3b7fc3660473",
				"creationTimestamp": "2025-10-17T22:38:12Z",
				"lastUpdatedTimestamp": "2025-10-17T22:38:12Z"
				},
			],
			"pageable": {
				"page": 0,
				"size": 25,
				"sort": []
			},
			"totalSize": 2,
			"totalPages": 1,
			"empty": false,
			"size": 25,
			"offset": 0,
			"numberOfElements": 2,
			"pageNumber": 0
			}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/processor", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},
		{
			name:      "Get with args returns a search array",
			args:      []string{"--filter", "type=mongo"},
			outGolden: "NewGetCommand_Out_SearchWithFiltersReturnsArray",
			errGolden: "NewGetCommand_Err_SearchWithFiltersReturnsArray",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_SearchWithFiltersReturnsArray"),
			errBytes:  []byte(nil),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/processor/search": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "processor-2",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "8c243a1d-9384-421d-8f99-4ef28d4e0ab0",
					"creationTimestamp": "2025-10-17T15:33:58Z",
					"lastUpdatedTimestamp": "2025-10-17T15:33:58Z"
				},
				"highlight": {},
				"score": 0.15534057
				},
				{
				"source": {
					"type": "mongo",
					"name": "my-processor",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "4957145b-6192-4862-a5da-e97853974e9f",
					"creationTimestamp": "2025-10-17T22:37:53Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
				},
				"highlight": {},
				"score": 0.15534057
				}
			],
			"pageable": {
				"page": 0,
				"size": 25,
				"sort": []
			},
			"totalSize": 13,
			"totalPages": 1,
			"empty": false,
			"size": 25,
			"offset": 0,
			"numberOfElements": 13,
			"pageNumber": 0
			}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/processor/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			args:      []string{},
			outGolden: "NewGetCommand_Out_NoURL",
			errGolden: "NewGetCommand_Err_NoURL",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_NoURL"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery QueryFlow URL is missing for profile \"default\".\nTo set the URL for the Discovery QueryFlow API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery queryflow config --profile \"default\""),
		},
		{
			name:      "user sends a name that does not exist",
			args:      []string{"test"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewGetCommand_Out_NameDoesNotExist",
			errGolden: "NewGetCommand_Err_NameDoesNotExist",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_NameDoesNotExist"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_NameDoesNotExist"),
			responses: map[string]testutils.MockResponse{
				"/v2/processor/search": {
					StatusCode:  http.StatusNoContent,
					Body:        ``,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/processor/search", r.URL.Path)
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
			}, "Could not search for entity with id \"test\""),
		},
		{
			name:      "Search By Name returns HTTP error",
			args:      []string{"3b32e410-2F33-412d-9fb8-17970131921c"},
			outGolden: "NewGetCommand_Out_SearchByNameHTTPError",
			errGolden: "NewGetCommand_Err_SearchByNameHTTPError",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_SearchByNameHTTPError"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_SearchByNameHTTPError"),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/processor/search": {
					StatusCode: http.StatusInternalServerError,
					Body: `{
			"status": 500,
			"code": 1003,
			"messages": [
				"Internal server error"
			],
			"timestamp": "2025-10-16T17:46:45.386963700Z"
			}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/processor/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusInternalServerError, Body: gjson.Parse(`{
			"status": 500,
			"code": 1003,
			"messages": [
				"Internal server error"
			],
			"timestamp": "2025-10-16T17:46:45.386963700Z"
			}`)}, "Could not search for entity with id \"3b32e410-2F33-412d-9fb8-17970131921c\""),
		},
		{
			name:      "GetEntities returns HTTP error",
			args:      []string{},
			outGolden: "NewGetCommand_Out_GetEntitiesHTTPError",
			errGolden: "NewGetCommand_Err_GetEntitiesHTTPError",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_GetEntitiesHTTPError"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_GetEntitiesHTTPError"),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"GET:/v2/processor": {
					StatusCode:  http.StatusUnauthorized,
					Body:        `{"error": "unauthorized"}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/processor", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error": "unauthorized"}`)}, "Could not get all entities"),
		},
		{
			name:      "SearchEntities returns HTTP error",
			args:      []string{"--filter", "label=A"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewGetCommand_Out_SearchHTTPError",
			errGolden: "NewGetCommand_Err_SearchHTTPError",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_SearchHTTPError"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_SearchHTTPError"),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/processor/search": {
					StatusCode: http.StatusUnauthorized,
					Body: `{
	"status": 401,
	"code": 1003,
	"messages": [
		"user is unauthorized"
	],
	"timestamp": "2025-09-30T15:38:42.885125200Z"
}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/processor/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusUnauthorized,
				Body: gjson.Parse(`{
	"status": 401,
	"code": 1003,
	"messages": [
		"user is unauthorized"
	],
	"timestamp": "2025-09-30T15:38:42.885125200Z"
}`),
			}, "Could not search for the entities"),
		},
		{
			name:      "Filter does not exist",
			args:      []string{"--filter", "gte=field:1"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewGetCommand_Out_FilterDoesNotExist",
			errGolden: "NewGetCommand_Err_FilterDoesNotExist",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_FilterDoesNotExist"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_FilterDoesNotExist"),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/processor/search": {
					StatusCode:  http.StatusBadRequest,
					Body:        ``,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/processor/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewError(cli.ErrorExitCode, "Filter type \"gte\" does not exist"),
		},
		{
			name:      "Printing JSON object fails",
			args:      []string{"test"},
			outGolden: "NewGetCommand_Out_PrintJSONFails",
			errGolden: "NewGetCommand_Err_PrintJSONFails",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_PrintJSONFails"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_PrintJSONFails"),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/processor/search": {
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
						assert.Equal(t, "/v2/processor/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"GET:/v2/processor/3d51beef-8b90-40aa-84b5-033241dc6239": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
					"type": "mongo",
					"name": "my-processor",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active: true,
					"id": "3d51beef-8b90-40aa-84b5-033241dc6239",
					"creationTimestamp": "2025-10-17T22:37:53Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
				}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/processor/3d51beef-8b90-40aa-84b5-033241dc6239", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character '\\n' in string literal"), "Could not print JSON object"),
		},
		{
			name:      "Printing JSON array fails",
			args:      []string{},
			outGolden: "NewGetCommand_Out_PrintArrayFails",
			errGolden: "NewGetCommand_Err_PrintArrayFails",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_PrintArrayFails"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_PrintArrayFails"),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"GET:/v2/processor": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [{"source":{"active":true,"creationTimestamp":"2025-08-21T17:57:16Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","processors":[],"lastUpdatedTimestamp":"2025-08-21T17:57:16Z","name":"test","type":"mongo"}},     
			{"source":{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"5f125024-1e5e-4591-9fee-365dc20eeeed","processors":[],"lastUpdatedTimestamp":"2025-08-18T20:55:43Z","name":"MongoDB text processor","type":mongo}},       
			{"source":{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"86e7f920-a4e4-4b64-be84-5437a7673db8","processors":[],"lastUpdatedTimestamp":"2025-08-14T18:02:38Z","name":"Script processor","type":"script"}}
			],
			"pageable": {
				"page": 0,
				"size": 3,
				"sort": []
			},
			"totalSize": 12,
			"totalPages": 4,
			"empty": false,
			"size": 3,
			"offset": 0,
			"numberOfElements": 3,
			"pageNumber": 0
			}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/processor", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character 'm' looking for beginning of value"), "Could not print JSON Array"),
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

			if tc.url {
				vpr.Set("default.queryflow_url", srv.URL)
			}
			if tc.apiKey != "" {
				vpr.Set("default.queryflow_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, t.TempDir())

			getCmd := NewGetCommand(d)

			getCmd.SetIn(ios.In)
			getCmd.SetOut(ios.Out)
			getCmd.SetErr(ios.Err)

			getCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			getCmd.SetArgs(tc.args)

			err := getCmd.Execute()
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

// TestNewGetCommand_NoProfileFlag tests the NewGetCommand when the profile flag was not defined.
func TestNewGetCommand_NoProfileFlag(t *testing.T) {
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
	vpr.Set("output", "json")

	vpr.Set("default.queryflow_url", "test")
	vpr.Set("default.queryflow_key", "test")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	getCmd := NewGetCommand(d)

	getCmd.SetIn(ios.In)
	getCmd.SetOut(ios.Out)
	getCmd.SetErr(ios.Err)

	getCmd.SetArgs([]string{"test"})

	err := getCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewGetCommand_Out_NoProfile", testutils.Read(t, "NewGetCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewGetCommand_Err_NoProfile", testutils.Read(t, "NewGetCommand_Err_NoProfile"), errBuf.Bytes())
}

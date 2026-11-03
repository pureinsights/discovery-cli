package serviceaccounts

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/pureinsights/discovery-cli/v2/internal/iostreams"
	"github.com/pureinsights/discovery-cli/v2/internal/testutils"
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
			args:      []string{"ci"},
			url:       true,
			apiKey:    "",
			outGolden: "NewGetCommand_Out_SearchByNameReturnsObject",
			errGolden: "NewGetCommand_Err_SearchByNameReturnsObject",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_SearchByNameReturnsObject"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/account/service/search": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
					"source": {
						"name": "ci",
						"description": "Continuous Integration account",
						"labels": [
							{
								"key": "keyA",
								"value": "valueA"
							},
							{
								"key": "keyB",
								"value": "valueB"
							}
						],
						"active": true,
						"id": "06f52112-a2bb-4490-97c2-171fb7d54d5b",
						"creationTimestamp": "2026-10-29T15:37:47Z",
						"lastUpdatedTimestamp": "2026-11-03T11:50:39Z"
					},
					"highlight": {},
					"score": 1
				}
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
						assert.Equal(t, "/v2/account/service/search", r.URL.Path)
					},
				},
				"GET:/v2/account/service/06f52112-a2bb-4490-97c2-171fb7d54d5b": {
					StatusCode: http.StatusOK,
					Body: `{
					"name": "ci",
					"description": "Continuous Integration account",
					"labels": [
						{
							"key": "keyA",
							"value": "valueA"
						},
						{
							"key": "keyB",
							"value": "valueB"
						}
					],
					"active": true,
					"id": "06f52112-a2bb-4490-97c2-171fb7d54d5b",
					"creationTimestamp": "2026-10-29T15:37:47Z",
					"lastUpdatedTimestamp": "2026-11-03T11:50:39Z",
					"role": "OPERATE"
				}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/account/service/06f52112-a2bb-4490-97c2-171fb7d54d5b", r.URL.Path)
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
				"GET:/v2/account/service": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
					"name": "ci",
					"description": "Continuous Integration account",
					"labels": [
						{
							"key": "keyA",
							"value": "valueA"
						},
						{
							"key": "keyB",
							"value": "valueB"
						}
					],
					"active": true,
					"id": "06f52112-a2bb-4490-97c2-171fb7d54d5b",
					"creationTimestamp": "2026-10-29T15:37:47Z",
					"lastUpdatedTimestamp": "2026-11-03T11:50:39Z"
				},
				{
					"name": "test-account",
					"labels": [],
					"active": true,
					"id": "0e35e9ed-9fa0-4397-91e7-c36ea2a798db",
					"creationTimestamp": "2026-11-03T11:51:08Z",
					"lastUpdatedTimestamp": "2026-11-03T11:51:08Z"
				},
				{
					"name": "admin",
					"description": "Admin account",
					"labels": [
						{
							"key": "keyA",
							"value": "valueA"
						}
					],
					"active": true,
					"id": "d74aef4e-38b0-4823-9aee-8521ff19bf5b",
					"creationTimestamp": "2026-10-29T15:27:02Z",
					"lastUpdatedTimestamp": "2026-10-29T15:57:06Z"
				}
			],
			"pageable": {
				"page": 0,
				"size": 25,
				"sort": []
			},
			"totalSize": 3,
			"totalPages": 1,
			"empty": false,
			"size": 25,
			"offset": 0,
			"numberOfElements": 3,
			"pageNumber": 0
		}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/account/service", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},
		{
			name:      "Get with a label filter returns a search array",
			args:      []string{"--filter", "label=keyA:valueA"},
			outGolden: "NewGetCommand_Out_SearchWithFiltersReturnsArray",
			errGolden: "NewGetCommand_Err_SearchWithFiltersReturnsArray",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_SearchWithFiltersReturnsArray"),
			errBytes:  []byte(nil),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/account/service/search": {
					StatusCode: http.StatusOK,
					Body: `{
				"content": [
					{
						"source": {
							"name": "ci",
							"description": "Continuous Integration account",
							"labels": [
								{
									"key": "keyA",
									"value": "valueA"
								},
								{
									"key": "keyB",
									"value": "valueB"
								}
							],
							"active": true,
							"id": "99aff895-36fd-4843-831b-6c64242a7b34",
							"creationTimestamp": "2026-11-03T16:27:10Z",
							"lastUpdatedTimestamp": "2026-11-03T16:27:10Z"
						},
						"highlight": {},
						"score": 1.4100635
					},
					{
						"source": {
							"name": "ci2",
							"description": "Continuous Integration account",
							"labels": [
								{
									"key": "keyA",
									"value": "valueA"
								},
								{
									"key": "keyB",
									"value": "valueB"
								}
							],
							"active": true,
							"id": "c1fb512f-c8d1-49e9-8a93-ac20dd692abe",
							"creationTimestamp": "2026-11-03T16:27:24Z",
							"lastUpdatedTimestamp": "2026-11-03T16:27:24Z"
						},
						"highlight": {},
						"score": 1.4100635
					}
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
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/account/service/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
						body, err := io.ReadAll(r.Body)
						require.NoError(t, err)
						assert.JSONEq(t, `{"and":[{"equals":{"field":"labels.key","value":"keyA"}},{"equals":{"field":"labels.value","value":"valueA"}}]}`, string(body))
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
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:      "user sends a name that does not exist",
			args:      []string{"test"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewGetCommand_Out_NameDoesNotExist",
			errGolden: "NewGetCommand_Err_NameDoesNotExist",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_NameDoesNotExist"),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/account/service/search": {
					StatusCode:  http.StatusNoContent,
					Body:        ``,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/account/service/search", r.URL.Path)
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
			args:      []string{"06f52112-a2bb-4490-97c2-171fb7d54d5b"},
			outGolden: "NewGetCommand_Out_SearchByNameHTTPError",
			errGolden: "NewGetCommand_Err_SearchByNameHTTPError",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_SearchByNameHTTPError"),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/account/service/search": {
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
						assert.Equal(t, "/v2/account/service/search", r.URL.Path)
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
			}`)}, "Could not search for entity with id \"06f52112-a2bb-4490-97c2-171fb7d54d5b\""),
		},
		{
			name:      "GetEntities returns HTTP error",
			args:      []string{},
			outGolden: "NewGetCommand_Out_GetEntitiesHTTPError",
			errGolden: "NewGetCommand_Err_GetEntitiesHTTPError",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_GetEntitiesHTTPError"),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"GET:/v2/account/service": {
					StatusCode:  http.StatusUnauthorized,
					Body:        `{"error": "unauthorized"}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/account/service", r.URL.Path)
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
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_SearchHTTPError"),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/account/service/search": {
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
						assert.Equal(t, "/v2/account/service/search", r.URL.Path)
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
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_FilterDoesNotExist"),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/account/service/search": {
					StatusCode:  http.StatusBadRequest,
					Body:        ``,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/account/service/search", r.URL.Path)
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
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_PrintJSONFails"),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/account/service/search": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
			"content": [
				{
				"source": {
					"name": "test",
					"description": "Test account",
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
						assert.Equal(t, "/v2/account/service/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"GET:/v2/account/service/3d51beef-8b90-40aa-84b5-033241dc6239": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
					"name": "test",
					"description": "Test account",
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
						assert.Equal(t, "/v2/account/service/3d51beef-8b90-40aa-84b5-033241dc6239", r.URL.Path)
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
				"GET:/v2/account/service": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [{"source":{"active":true,"creationTimestamp":"2025-08-21T17:57:16Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","labels":[],"lastUpdatedTimestamp":"2025-08-21T17:57:16Z","name":"test","role":"OPERATE"}},
			{"source":{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"5f125024-1e5e-4591-9fee-365dc20eeeed","labels":[],"lastUpdatedTimestamp":"2025-08-18T20:55:43Z","name":"ci","role":OPERATE}},
			{"source":{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"86e7f920-a4e4-4b64-be84-5437a7673db8","labels":[],"lastUpdatedTimestamp":"2025-08-14T18:02:38Z","name":"admin","role":"ADMIN"}}
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
						assert.Equal(t, "/v2/account/service", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character 'O' looking for beginning of value"), "Could not print JSON Array"),
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
				vpr.Set("default.core_url", srv.URL)
			}
			if tc.apiKey != "" {
				vpr.Set("default.core_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, t.TempDir())

			getCmd := NewGetCommand(d)

			getCmd.SilenceUsage = true
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

			if tc.outBytes != nil {
				testutils.CompareBytes(t, tc.outGolden, tc.outBytes, out.Bytes())
			}
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

	vpr.Set("default.core_url", "test")
	vpr.Set("default.core_key", "test")

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

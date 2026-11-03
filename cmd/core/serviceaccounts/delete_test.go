package serviceaccounts

import (
	"bytes"
	"errors"
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

// TestNewDeleteCommand tests the NewDeleteCommand() function.
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
			args:      []string{"06f52112-a2bb-4490-97c2-171fb7d54d5b"},
			url:       true,
			apiKey:    "",
			outGolden: "NewDeleteCommand_Out_DeleteByIdReturnsObject",
			errGolden: "NewDeleteCommand_Err_DeleteByIdReturnsObject",
			outBytes:  testutils.Read(t, "NewDeleteCommand_Out_DeleteByIdReturnsObject"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/account/service/search": {
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
				"DELETE:/v2/account/service/06f52112-a2bb-4490-97c2-171fb7d54d5b": {
					StatusCode: http.StatusOK,
					Body: `{
				"acknowledged": true
			}`,

					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodDelete, r.Method)
						assert.Equal(t, "/v2/account/service/06f52112-a2bb-4490-97c2-171fb7d54d5b", r.URL.Path)
					},
				},
			},
			err: nil,
		},
		{
			name:      "Delete by name returns an acknowledged true",
			args:      []string{"ci3"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewDeleteCommand_Out_DeleteByNameReturnsObject",
			errGolden: "NewDeleteCommand_Err_DeleteByNameReturnsObject",
			outBytes:  testutils.Read(t, "NewDeleteCommand_Out_DeleteByNameReturnsObject"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/account/service/search": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
				"content": [
					{
						"source": {
							"name": "ci3",
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
							"active": false,
							"id": "65370d8b-d83c-4245-b702-3fea4ef32814",
							"creationTimestamp": "2026-11-03T16:48:04Z",
							"lastUpdatedTimestamp": "2026-11-03T16:48:04Z"
						},
						"highlight": {
							"name": [
								"<em>ci3</em>"
							]
						},
						"score": 2.3932114
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
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/account/service/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"GET:/v2/account/service/65370d8b-d83c-4245-b702-3fea4ef32814": {
					StatusCode: http.StatusOK,
					Body: `{
					"name": "ci3",
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
					"active": false,
					"id": "65370d8b-d83c-4245-b702-3fea4ef32814",
					"creationTimestamp": "2026-11-03T16:48:04Z",
					"lastUpdatedTimestamp": "2026-11-03T16:48:04Z",
					"role": "OPERATE"
				}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/account/service/65370d8b-d83c-4245-b702-3fea4ef32814", r.URL.Path)
					},
				},
				"DELETE:/v2/account/service/65370d8b-d83c-4245-b702-3fea4ef32814": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
				"acknowledged": true
			}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodDelete, r.Method)
						assert.Equal(t, "/v2/account/service/65370d8b-d83c-4245-b702-3fea4ef32814", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			args:      []string{"ci"},
			outGolden: "NewDeleteCommand_Out_NoURL",
			errGolden: "NewDeleteCommand_Err_NoURL",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewDeleteCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:      "sent name does not exist",
			args:      []string{"test"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewDeleteCommand_Out_NameDoesNotExist",
			errGolden: "NewDeleteCommand_Err_NameDoesNotExist",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewDeleteCommand_Err_NameDoesNotExist"),
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
			}, "Could not search for entity with name \"test\""),
		},
		{
			name:      "Printing JSON object fails",
			args:      []string{"ci"},
			outGolden: "NewDeleteCommand_Out_PrintJSONFails",
			errGolden: "NewDeleteCommand_Err_PrintJSONFails",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewDeleteCommand_Err_PrintJSONFails"),
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
					"name": "ci",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "06f52112-a2bb-4490-97c2-171fb7d54d5b",
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
				"GET:/v2/account/service/06f52112-a2bb-4490-97c2-171fb7d54d5b": {
					StatusCode: http.StatusOK,
					Body: `{
					"name": "ci",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "06f52112-a2bb-4490-97c2-171fb7d54d5b",
					"creationTimestamp": "2025-10-17T22:37:57Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:57Z"
				}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/account/service/06f52112-a2bb-4490-97c2-171fb7d54d5b", r.URL.Path)
					},
				},
				"DELETE:/v2/account/service/06f52112-a2bb-4490-97c2-171fb7d54d5b": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
				"acknowledged: true
			}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodDelete, r.Method)
						assert.Equal(t, "/v2/account/service/06f52112-a2bb-4490-97c2-171fb7d54d5b", r.URL.Path)
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
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewDeleteCommand_Err_InvalidUUID"),
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
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "06f52112-a2bb-4490-97c2-171fb7d54d5b",
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
				"GET:/v2/account/service/06f52112-a2bb-4490-97c2-171fb7d54d5b": {
					StatusCode: http.StatusOK,
					Body: `{
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
						assert.Equal(t, "/v2/account/service/06f52112-a2bb-4490-97c2-171fb7d54d5b", r.URL.Path)
					},
				},
				"DELETE:/v2/account/service/06f52112-a2bb-4490-97c2-171fb7d54d5b": {
					StatusCode: http.StatusOK,
					Body: `{
				"acknowledged": true
			}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodDelete, r.Method)
						assert.Equal(t, "/v2/account/service/06f52112-a2bb-4490-97c2-171fb7d54d5b", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid UUID length: 4"), "Could not delete entity with name \"test\""),
		},
		{
			name:      "Delete returns HTTP error",
			args:      []string{"06f52112-a2bb-4490-97c2-171fb7d54d5b"},
			outGolden: "NewDeleteCommand_Out_DeleteHTTPError",
			errGolden: "NewDeleteCommand_Err_DeleteHTTPError",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewDeleteCommand_Err_DeleteHTTPError"),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/account/service/search": {
					StatusCode:  http.StatusNoContent,
					Body:        ``,
					ContentType: "application/json",
				},
				"GET:/v2/account/service/06f52112-a2bb-4490-97c2-171fb7d54d5b": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body:        `{"name":"ci","active":true,"id":"06f52112-a2bb-4490-97c2-171fb7d54d5b","role":"OPERATE"}`,
				},
				"DELETE:/v2/account/service/06f52112-a2bb-4490-97c2-171fb7d54d5b": {
					StatusCode:  http.StatusNotFound,
					ContentType: "application/json",
					Body: `{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: 06f52112-a2bb-4490-97c2-171fb7d54d5b"
	],
	"timestamp": "2026-11-03T16:50:12.000000000Z"
}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodDelete, r.Method)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: 06f52112-a2bb-4490-97c2-171fb7d54d5b"
	],
	"timestamp": "2026-11-03T16:50:12.000000000Z"
}`)}, "Could not delete entity with id \"06f52112-a2bb-4490-97c2-171fb7d54d5b\""),
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

			deleteCmd := NewDeleteCommand(d)

			deleteCmd.SilenceUsage = true
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

			if tc.outBytes != nil {
				testutils.CompareBytes(t, tc.outGolden, tc.outBytes, out.Bytes())
			}
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

	vpr.Set("default.core_url", "test")
	vpr.Set("default.core_key", "test")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	deleteCmd := NewDeleteCommand(d)

	deleteCmd.SetIn(ios.In)
	deleteCmd.SetOut(ios.Out)
	deleteCmd.SetErr(ios.Err)

	deleteCmd.SetArgs([]string{"ci"})

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

	vpr.Set("default.core_url", "test")
	vpr.Set("default.core_key", "test")

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

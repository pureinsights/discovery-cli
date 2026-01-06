package credentials

import (
	"bytes"
	"errors"
	"fmt"
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

// TestNewStoreCommand tests the NewStoreCommand function.
func TestNewStoreCommand(t *testing.T) {
	tests := []struct {
		name         string
		url          bool
		apiKey       string
		outGolden    string
		errGolden    string
		outBytes     []byte
		errBytes     []byte
		data         string
		file         string
		abortOnError bool
		responses    map[string]testutils.MockResponse
		err          error
	}{
		// Working case
		{
			name:      "Store receives a single JSON",
			url:       true,
			apiKey:    "",
			outGolden: "NewStoreCommand_Out_StoreSingleJSON",
			errGolden: "NewStoreCommand_Err_StoreSingleJSON",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_StoreSingleJSON"),
			errBytes:  []byte(nil),
			data: `{
			"type": "mongo",
			"name": "MongoDB credential",
			"labels": [],
			"active": true,
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret"
			}`,
			file:         "",
			abortOnError: false,
			responses: map[string]testutils.MockResponse{
				"POST:/v2/credential": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "MongoDB credential",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
					"creationTimestamp": "2025-08-14T18:02:11Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
					"secret": "mongo-secret"
					}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/credential", r.URL.Path)
					},
				},
			},
			err: nil,
		},
		{
			name:      "Store receives a JSON array of configs with creates, failures, and updates with abort on error false",
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewStoreCommand_Out_StoreArrayNoAbort",
			errGolden: "NewStoreCommand_Err_StoreArrayNoAbort",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_StoreArrayNoAbort"),
			errBytes:  []byte(nil),
			data: `[{
			"type": "mongo",
			"name": "MongoDB credential",
			"labels": [],
			"active": true,
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret"
			},
			{
			"type": "mongo",
			"name": "MongoDB credential 2",
			"labels": [],
			"active": true,
			"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4d",
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret-2"
			},
			{
			"type": "openai",
			"name": "OpenAI credential 3",
			"labels": [],
			"active": true,
			"id": "9ababe08-0b74-4672-bb7c-e7a8227d6dad",
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "openai-secret"
			}
			]`,
			file:         "",
			abortOnError: false,
			responses: map[string]testutils.MockResponse{
				"POST:/v2/credential": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "MongoDB credential",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
					"creationTimestamp": "2025-08-14T18:02:11Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
					"secret": "mongo-secret"
					}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/credential", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"PUT:/v2/credential/9ababe08-0b74-4672-bb7c-e7a8227d6d4d": {
					StatusCode: http.StatusNotFound,
					Body: `{
					"status": 404,
					"code": 1003,
					"messages": [
						"Entity not found: 9ababe08-0b74-4672-bb7c-e7a8227d6d4d"
					],
					"timestamp": "2025-10-29T23:12:08.002244700Z"
					}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPut, r.Method)
						assert.Equal(t, "/v2/credential/9ababe08-0b74-4672-bb7c-e7a8227d6d4d", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"PUT:/v2/credential/9ababe08-0b74-4672-bb7c-e7a8227d6dad": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "openai",
					"name": "OpenAI credential",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6dad",
					"creationTimestamp": "2025-08-14T18:02:11Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
					"secret": "openai-secret"
					}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPut, r.Method)
						assert.Equal(t, "/v2/credential/9ababe08-0b74-4672-bb7c-e7a8227d6dad", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},
		{
			name:         "Store receives a file with a valid JSON array",
			url:          true,
			apiKey:       "apiKey123",
			outGolden:    "NewStoreCommand_Out_StoreFile",
			errGolden:    "NewStoreCommand_Err_StoreFile",
			outBytes:     testutils.Read(t, "NewStoreCommand_Out_StoreFile"),
			errBytes:     []byte(nil),
			data:         "",
			file:         "testdata/StoreCommand_JSONFile.json",
			abortOnError: false,
			responses: map[string]testutils.MockResponse{
				"POST:/v2/credential": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "MongoDB credential",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
					"creationTimestamp": "2025-08-14T18:02:11Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
					"secret": "mongo-secret"
					}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/credential", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"PUT:/v2/credential/9ababe08-0b74-4672-bb7c-e7a8227d6d4d": {
					StatusCode: http.StatusNotFound,
					Body: `{
					"status": 404,
					"code": 1003,
					"messages": [
						"Entity not found: 9ababe08-0b74-4672-bb7c-e7a8227d6d4d"
					],
					"timestamp": "2025-10-29T23:12:08.002244700Z"
					}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPut, r.Method)
						assert.Equal(t, "/v2/credential/9ababe08-0b74-4672-bb7c-e7a8227d6d4d", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"PUT:/v2/credential/9ababe08-0b74-4672-bb7c-e7a8227d6dad": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "openai",
					"name": "OpenAI credential",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6dad",
					"creationTimestamp": "2025-08-14T18:02:11Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
					"secret": "openai-secret"
					}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPut, r.Method)
						assert.Equal(t, "/v2/credential/9ababe08-0b74-4672-bb7c-e7a8227d6dad", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			outGolden: "NewStoreCommand_Out_NoURL",
			errGolden: "NewStoreCommand_Err_NoURL",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_NoURL"),
			errBytes:  testutils.Read(t, "NewStoreCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			data: `{
			"type": "mongo",
			"name": "MongoDB credential",
			"labels": [],
			"active": true,
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret"
			}`,
			file:         "",
			abortOnError: false,
			err:          cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:      "Store receives a JSON array of configs with creates, failures, and updates with abort on error true",
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewStoreCommand_Out_StoreArrayAbort",
			errGolden: "NewStoreCommand_Err_StoreArrayAbort",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_StoreArrayAbort"),
			errBytes:  testutils.Read(t, "NewStoreCommand_Err_StoreArrayAbort"),
			data: `[{
			"type": "mongo",
			"name": "MongoDB credential",
			"labels": [],
			"active": true,
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret"
			},
			{
			"type": "mongo",
			"name": "MongoDB credential 2",
			"labels": [],
			"active": true,
			"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4d",
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret-2"
			},
			{
			"type": "openai",
			"name": "OpenAI credential 3",
			"labels": [],
			"active": true,
			"id": "9ababe08-0b74-4672-bb7c-e7a8227d6dad",
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "openai-secret"
			}
			]`,
			file:         "",
			abortOnError: true,
			responses: map[string]testutils.MockResponse{
				"POST:/v2/credential": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "MongoDB credential",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
					"creationTimestamp": "2025-08-14T18:02:11Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
					"secret": "mongo-secret"
					}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/credential", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"PUT:/v2/credential/9ababe08-0b74-4672-bb7c-e7a8227d6d4d": {
					StatusCode: http.StatusNotFound,
					Body: `{
					"status": 404,
					"code": 1003,
					"messages": [
						"Entity not found: 9ababe08-0b74-4672-bb7c-e7a8227d6d4d"
					]
					}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPut, r.Method)
						assert.Equal(t, "/v2/credential/9ababe08-0b74-4672-bb7c-e7a8227d6d4d", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"PUT:/v2/credential/9ababe08-0b74-4672-bb7c-e7a8227d6dad": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "openai",
					"name": "OpenAI credential",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6dad",
					"creationTimestamp": "2025-08-14T18:02:11Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
					"secret": "openai-secret"
					}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPut, r.Method)
						assert.Equal(t, "/v2/credential/9ababe08-0b74-4672-bb7c-e7a8227d6dad", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
					"status": 404,
					"code": 1003,
					"messages": [
						"Entity not found: 9ababe08-0b74-4672-bb7c-e7a8227d6d4d"
					]
					}`)}, "Could not store entities"),
		},
		{
			name:         "StoreCommand reads an empty file",
			url:          true,
			apiKey:       "apiKey123",
			outGolden:    "NewStoreCommand_Out_StoreEmptyFile",
			errGolden:    "NewStoreCommand_Err_StoreEmptyFile",
			outBytes:     testutils.Read(t, "NewStoreCommand_Out_StoreEmptyFile"),
			errBytes:     testutils.Read(t, "NewStoreCommand_Err_StoreEmptyFile"),
			data:         "",
			file:         "testdata/StoreCommand_EmptyFile.json",
			abortOnError: false,
			err:          cli.NewError(cli.ErrorExitCode, "Data cannot be empty"),
		},
		{
			name:      "StoreCommand tries to read a file that does not exist",
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewStoreCommand_Out_StoreFileNotExists",
			errGolden: "NewStoreCommand_Err_StoreFileNotExists",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_StoreFileNotExists"),
			data:      "",
			file:      "doesnotexist",
			errBytes:  testutils.Read(t, "NewStoreCommand_Err_StoreFileNotExists"),
			err:       cli.NewErrorWithCause(cli.ErrorExitCode, fmt.Errorf("file does not exist: %s", "doesnotexist"), "Could not read file \"doesnotexist\""),
		},
		{
			name:         "StoreCommand gets empty data",
			url:          true,
			apiKey:       "apiKey123",
			outGolden:    "NewStoreCommand_Out_StoreEmptyData",
			errGolden:    "NewStoreCommand_Err_StoreEmptyData",
			outBytes:     testutils.Read(t, "NewStoreCommand_Out_StoreEmptyData"),
			errBytes:     testutils.Read(t, "NewStoreCommand_Err_StoreEmptyData"),
			data:         "",
			file:         "",
			abortOnError: false,
			err:          cli.NewError(cli.ErrorExitCode, "Data cannot be empty"),
		},
		{
			name:      "Printing JSON Array fails",
			outGolden: "NewStoreCommand_Out_PrintJSONFails",
			errGolden: "NewStoreCommand_Err_PrintJSONFails",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_PrintJSONFails"),
			errBytes:  testutils.Read(t, "NewStoreCommand_Err_PrintJSONFails"),
			url:       true,
			apiKey:    "apiKey123",
			data: `{
			"type": "mongo",
			"name": "MongoDB credential",
			"labels": [],
			"active": true,
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret"
			}`,
			file:         "",
			abortOnError: false,
			responses: map[string]testutils.MockResponse{
				"POST:/v2/credential": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
			"type": "mongo",
			"name": "MongoDB credential",
			"labels": [],
			"active": true,
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret
			}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/credential", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character '\\n' in string literal"), "Could not print JSON Array"),
		},
		{
			name:      "user sends invalid UUID error",
			outGolden: "NewStoreCommand_Out_InvalidUUID",
			errGolden: "NewStoreCommand_Err_InvalidUUID",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_InvalidUUID"),
			errBytes:  testutils.Read(t, "NewStoreCommand_Err_InvalidUUID"),
			url:       true,
			apiKey:    "apiKey123",
			data: `{
			"type": "mongo",
			"name": "MongoDB credential",
			"labels": [],
			"active": true,
			"id": "test",
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret"
			}`,
			file:         "",
			abortOnError: true,
			responses: map[string]testutils.MockResponse{
				"POST:/v2/credential": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
			"type": "mongo",
			"name": "MongoDB credential",
			"labels": [],
			"active": true,
			"id": "test",
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret
			}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/credential", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid UUID length: 4"), "Could not store entities"),
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
			vpr.Set("output", "json")
			if tc.url {
				vpr.Set("default.core_url", srv.URL)
			}
			if tc.apiKey != "" {
				vpr.Set("default.core_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, t.TempDir())

			storeCmd := NewStoreCommand(d)

			storeCmd.SetIn(ios.In)
			storeCmd.SetOut(ios.Out)
			storeCmd.SetErr(ios.Err)

			storeCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			args := []string{}
			if tc.data != "" || tc.file == "" {
				args = append(args, "--data")
				args = append(args, tc.data)
			}

			if tc.file != "" {
				args = append(args, tc.file)
			}

			args = append(args, fmt.Sprintf("--abort-on-error=%t", tc.abortOnError))
			storeCmd.SetArgs(args)

			err := storeCmd.Execute()
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

// TestNewStoreCommand_MultipleFiles tests what happens if the user sends multiple files as arguments.
func TestNewStoreCommand_MultipleFiles(t *testing.T) {
	responses := map[string]testutils.MockResponse{
		"POST:/v2/credential": {
			StatusCode: http.StatusOK,
			Body: `{
			"type": "mongo",
			"name": "MongoDB credential",
			"labels": [],
			"active": true,
			"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret"
			}`,
			ContentType: "application/json",
			Assertions: func(t *testing.T, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/v2/credential", r.URL.Path)
				assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
			},
		},
		"PUT:/v2/credential/9ababe08-0b74-4672-bb7c-e7a8227d6d4d": {
			StatusCode: http.StatusNotFound,
			Body: `{
			"status": 404,
			"code": 1003,
			"messages": [
				"Entity not found: 9ababe08-0b74-4672-bb7c-e7a8227d6d4d"
			],
			"timestamp": "2025-10-29T23:12:08.002244700Z"
			}`,
			ContentType: "application/json",
			Assertions: func(t *testing.T, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Equal(t, "/v2/credential/9ababe08-0b74-4672-bb7c-e7a8227d6d4d", r.URL.Path)
				assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
			},
		},
		"PUT:/v2/credential/9ababe08-0b74-4672-bb7c-e7a8227d6dad": {
			StatusCode: http.StatusOK,
			Body: `{
			"type": "openai",
			"name": "OpenAI credential",
			"labels": [],
			"active": true,
			"id": "9ababe08-0b74-4672-bb7c-e7a8227d6dad",
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "openai-secret"
			}`,
			ContentType: "application/json",
			Assertions: func(t *testing.T, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Equal(t, "/v2/credential/9ababe08-0b74-4672-bb7c-e7a8227d6dad", r.URL.Path)
				assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
			},
		},
		"PUT:/v2/credential/2fba75cd-e5dc-454f-b028-a505f0c690b2": {
			StatusCode: http.StatusOK,
			Body: `{
				"type": "mongo",
				"name": "MongoDB credential 4",
				"labels": [],
				"active": true,
				"id": "2fba75cd-e5dc-454f-b028-a505f0c690b2",
				"creationTimestamp": "2025-08-14T18:02:11Z",
				"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
				"secret": "mongo-secret-2"
			}`,
			ContentType: "application/json",
			Assertions: func(t *testing.T, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Equal(t, "/v2/credential/2fba75cd-e5dc-454f-b028-a505f0c690b2", r.URL.Path)
				assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
			},
		},
		"PUT:/v2/credential/77cd87e5-5c29-4960-9058-4a2eda0c513f": {
			StatusCode: http.StatusOK,
			Body: `{
				"type": "openai",
				"name": "OpenAI credential 5",
				"labels": [],
				"active": true,
				"id": "77cd87e5-5c29-4960-9058-4a2eda0c513f",
				"creationTimestamp": "2025-08-14T18:02:11Z",
				"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
				"secret": "openai-secret"
			}`,
			ContentType: "application/json",
			Assertions: func(t *testing.T, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Equal(t, "/v2/credential/77cd87e5-5c29-4960-9058-4a2eda0c513f", r.URL.Path)
				assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
			},
		},
	}

	srv := httptest.NewServer(testutils.HttpMultiResponseHandler(t, responses))

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
	vpr.Set("output", "json")
	vpr.Set("default.core_url", srv.URL)
	vpr.Set("default.core_key", "apiKey123")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	storeCmd := NewStoreCommand(d)

	storeCmd.SetIn(ios.In)
	storeCmd.SetOut(ios.Out)
	storeCmd.SetErr(ios.Err)

	storeCmd.PersistentFlags().StringP(
		"profile",
		"p",
		d.Config().GetString("profile"),
		"configuration profile to use",
	)

	args := []string{"testdata/StoreCommand_JSONFile.json", "testdata/StoreCommand_JSONFile2.json"}

	storeCmd.SetArgs(args)

	err := storeCmd.Execute()
	require.NoError(t, err)

	testutils.CompareBytes(t, "NewStoreCommand_Out_MultipleFiles", testutils.Read(t, "NewStoreCommand_Out_MultipleFiles"), out.Bytes())
}

// TestNewStoreCommand_NoProfileFlag tests the NewStoreCommand when the profile flag was not defined.
func TestNewStoreCommand_NoProfileFlag(t *testing.T) {
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

	storeCmd := NewStoreCommand(d)

	storeCmd.SetIn(ios.In)
	storeCmd.SetOut(ios.Out)
	storeCmd.SetErr(ios.Err)

	storeCmd.SetArgs([]string{"testdata/StoreCommand_JSONFile.json"})

	err := storeCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewStoreCommand_Out_NoProfile", testutils.Read(t, "NewStoreCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewStoreCommand_Err_NoProfile", testutils.Read(t, "NewStoreCommand_Err_NoProfile"), errBuf.Bytes())
}

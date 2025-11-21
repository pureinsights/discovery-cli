package seeds

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// NewStoreCommand tests the NewStoreCommand function
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
			"name": "MongoDB seed",
			"labels": [],
			"active": true,
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret"
			}`,
			file:         "",
			abortOnError: false,
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "MongoDB seed",
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
						assert.Equal(t, "/v2/seed", r.URL.Path)
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
			"name": "MongoDB seed",
			"labels": [],
			"active": true,
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret"
			},
			{
			"type": "mongo",
			"name": "MongoDB seed 2",
			"labels": [],
			"active": true,
			"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4d",
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret-2"
			},
			{
			"type": "openai",
			"name": "OpenAI seed 3",
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
				"POST:/v2/seed": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "MongoDB seed",
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
						assert.Equal(t, "/v2/seed", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"PUT:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4d": {
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
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4d", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"PUT:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6dad": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "openai",
					"name": "OpenAI seed",
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
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6dad", r.URL.Path)
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
				"POST:/v2/seed": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "MongoDB seed",
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
						assert.Equal(t, "/v2/seed", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"PUT:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4d": {
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
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4d", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"PUT:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6dad": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "openai",
					"name": "OpenAI seed",
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
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6dad", r.URL.Path)
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
			"name": "MongoDB seed",
			"labels": [],
			"active": true,
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret"
			}`,
			file:         "",
			abortOnError: false,
			err:          cli.NewError(cli.ErrorExitCode, "The Discovery Ingestion URL is missing for profile \"default\".\nTo set the URL for the Discovery Ingestion API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery ingestion config --profile \"default\""),
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
			"name": "MongoDB seed",
			"labels": [],
			"active": true,
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret"
			},
			{
			"type": "mongo",
			"name": "MongoDB seed 2",
			"labels": [],
			"active": true,
			"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4d",
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret-2"
			},
			{
			"type": "openai",
			"name": "OpenAI seed 3",
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
				"POST:/v2/seed": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "MongoDB seed",
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
						assert.Equal(t, "/v2/seed", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"PUT:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4d": {
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
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4d", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"PUT:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6dad": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "openai",
					"name": "OpenAI seed",
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
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6dad", r.URL.Path)
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
			name:         "StoreCommand tries to read a file that does not exist",
			url:          true,
			apiKey:       "apiKey123",
			outGolden:    "NewStoreCommand_Out_StoreFileNotExists",
			errGolden:    "NewStoreCommand_Err_StoreFileNotExists",
			outBytes:     testutils.Read(t, "NewStoreCommand_Out_StoreFileNotExists"),
			errBytes:     []byte(nil),
			data:         "",
			file:         "doesnotexist",
			abortOnError: false,
			err:          cli.NewErrorWithCause(cli.ErrorExitCode, fs.ErrNotExist, "Could not read file \"doesnotexist\""),
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
			"name": "MongoDB seed",
			"labels": [],
			"active": true,
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret"
			}`,
			file:         "",
			abortOnError: false,
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
			"type": "mongo",
			"name": "MongoDB seed",
			"labels": [],
			"active": true,
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret
			}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed", r.URL.Path)
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
			"name": "MongoDB seed",
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
				"POST:/v2/seed": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
			"type": "mongo",
			"name": "MongoDB seed",
			"labels": [],
			"active": true,
			"id": "test",
			"creationTimestamp": "2025-08-14T18:02:11Z",
			"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
			"secret": "mongo-secret
			}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed", r.URL.Path)
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
				vpr.Set("default.ingestion_url", srv.URL)
			}
			if tc.apiKey != "" {
				vpr.Set("default.ingestion_key", tc.apiKey)
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
				args = append(args, "--file")
				args = append(args, tc.file)
			}

			args = append(args, fmt.Sprintf("--abort-on-error=%t", tc.abortOnError))
			storeCmd.SetArgs(args)

			err := storeCmd.Execute()
			if tc.err != nil {
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				cliError, _ := tc.err.(cli.Error)
				if !errors.Is(cliError.Cause, fs.ErrNotExist) {
					assert.EqualError(t, err, tc.err.Error())
					testutils.CompareBytes(t, tc.errGolden, tc.errBytes, errBuf.Bytes())
				} else {
					assert.Equal(t, cliError.ExitCode, errStruct.ExitCode)
					assert.Equal(t, cliError.Message, errStruct.Message)
				}
			} else {
				require.NoError(t, err)
			}

			testutils.CompareBytes(t, tc.outGolden, tc.outBytes, out.Bytes())
		})
	}
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

	vpr.Set("default.ingestion_url", "test")
	vpr.Set("default.ingestion_key", "test")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	storeCmd := NewStoreCommand(d)

	storeCmd.SetIn(ios.In)
	storeCmd.SetOut(ios.Out)
	storeCmd.SetErr(ios.Err)

	storeCmd.SetArgs([]string{"--file", "testdata/StoreCommand_JSONFile.json"})

	err := storeCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewStoreCommand_Out_NoProfile", testutils.Read(t, "NewStoreCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewStoreCommand_Err_NoProfile", testutils.Read(t, "NewStoreCommand_Err_NoProfile"), errBuf.Bytes())
}

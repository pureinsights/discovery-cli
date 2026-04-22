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

// TestNewStatusCommand tests the NewStatusCommand() function.
func TestNewStatusCommand(t *testing.T) {
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
			name:      "The user gets the status of a seed execution by its id and with details",
			args:      []string{"3b32e410-2f33-412d-9fb8-17970131921c", "--execution", "f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3", "--details"},
			url:       true,
			apiKey:    "",
			outGolden: "NewStatusCommand_Out_StatusExecutionByIdDetails",
			errGolden: "NewStatusCommand_Err_StatusExecutionByIdDetails",
			outBytes:  testutils.Read(t, "NewStatusCommand_Out_StatusExecutionByIdDetails"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
					StatusCode:  http.StatusNoContent,
					Body:        ``,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
					},
				},
				"GET:/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c": {
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
							"id": "3b32e410-2F33-412D-9fb8-17970131921c",
							"creationTimestamp": "2025-10-17T22:37:53Z",
							"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
						}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c", r.URL.Path)
					},
				},
				"GET:/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/execution/f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3": {
					StatusCode: http.StatusOK,
					Body: `{
							"id": "f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3",
							"creationTimestamp": "2025-10-10T19:48:31Z",
							"lastUpdatedTimestamp": "2025-10-10T19:48:31Z",
							"triggerType": "MANUAL",
							"status": "RUNNING",
							"scanType": "FULL",
							"properties": {
								"stagingBucket": "testBucket"
							},
							"stages": ["BEFORE_HOOKS","INGEST"]
							}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/execution/f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3", r.URL.Path)
					},
				},
				"GET:/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/execution/f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3/audit": {
					StatusCode: http.StatusOK,
					Body: `{
					"content": [
						{"timestamp":"2025-09-05T20:09:22.543Z","status":"CREATED","stages":[]},
						{"timestamp":"2025-09-05T20:09:26.621Z","status":"RUNNING","stages":[]},
						{"timestamp":"2025-09-05T20:09:37.592Z","status":"RUNNING","stages":["BEFORE_HOOKS"]},
						{"timestamp":"2025-09-05T20:13:26.602Z","status":"RUNNING","stages":["BEFORE_HOOKS","INGEST"]}
					],
					"pageable": {
						"page": 0,
						"size": 25,
						"sort": []
					},
					"totalSize": 4,
					"totalPages": 1,
					"empty": false,
					"size": 25,
					"offset": 0,
					"numberOfElements": 4,
					"pageNumber": 0
					}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/execution/f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3/audit", r.URL.Path)
					},
				},
				"GET:/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/execution/f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3/record/summary": {
					StatusCode:  http.StatusOK,
					Body:        `{"PROCESSING":4,"DONE": 4}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/execution/f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3/record/summary", r.URL.Path)
					},
				},
				"GET:/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/execution/f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3/job/summary": {
					StatusCode:  http.StatusOK,
					Body:        `{"DONE":5,"RUNNING":3}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/execution/f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3/job/summary", r.URL.Path)
					},
				},
			},
			err: nil,
		},
		{
			name:      "The user gets the status of a seed execution by its id, but with no details",
			args:      []string{"3b32e410-2f33-412d-9fb8-17970131921c", "--execution", "f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3"},
			url:       true,
			apiKey:    "",
			outGolden: "NewStatusCommand_Out_StatusExecutionByIdNoDetails",
			errGolden: "NewStatusCommand_Err_StatusExecutionByIdNoDetails",
			outBytes:  testutils.Read(t, "NewStatusCommand_Out_StatusExecutionByIdNoDetails"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
					StatusCode:  http.StatusNoContent,
					Body:        ``,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
					},
				},
				"GET:/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c": {
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
							"id": "3b32e410-2F33-412D-9fb8-17970131921c",
							"creationTimestamp": "2025-10-17T22:37:53Z",
							"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
						}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c", r.URL.Path)
					},
				},
				"GET:/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/execution/f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3": {
					StatusCode: http.StatusOK,
					Body: `{
							"id": "f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3",
							"creationTimestamp": "2025-10-10T19:48:31Z",
							"lastUpdatedTimestamp": "2025-10-10T19:48:31Z",
							"triggerType": "MANUAL",
							"status": "RUNNING",
							"scanType": "FULL",
							"properties": {
								"stagingBucket": "testBucket"
							},
							"stages": ["BEFORE_HOOKS","INGEST"]
							}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/execution/f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3", r.URL.Path)
					},
				},
			},
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			args:      []string{"3b32e410-2f33-412d-9fb8-17970131921c"},
			outGolden: "NewStatusCommand_Out_NoURL",
			errGolden: "NewStatusCommand_Err_NoURL",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewStatusCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Ingestion URL is missing for profile \"default\".\nTo set the URL for the Discovery Ingestion API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery ingestion config --profile \"default\""),
		},
		{
			name:      "user sends a name that does not exist",
			args:      []string{"test", "--execution", "3b32e410-2f33-412d-9fb8-17970131921c"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewStatusCommand_Out_NameDoesNotExist",
			errGolden: "NewStatusCommand_Err_NameDoesNotExist",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewStatusCommand_Err_NameDoesNotExist"),
			responses: map[string]testutils.MockResponse{
				"/v2/seed/search": {
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
			}, "Could not search for entity with id \"test\""),
		},
		{
			name:      "Search By Name returns HTTP error",
			args:      []string{"3b32e410-2F33-412d-9fb8-17970131921c", "--execution", "f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3"},
			outGolden: "NewStatusCommand_Out_SearchByNameHTTPError",
			errGolden: "NewStatusCommand_Err_SearchByNameHTTPError",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewStatusCommand_Err_SearchByNameHTTPError"),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
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
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
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
			name:      "The user tries to get the status of a seed execution by its id, but gets Not Found",
			args:      []string{"3b32e410-2f33-412d-9fb8-17970131921c", "--execution", "f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3"},
			url:       true,
			apiKey:    "",
			outGolden: "NewStatusCommand_Out_StatusExecutionNotFound",
			errGolden: "NewStatusCommand_Err_StatusExecutionNotFound",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewStatusCommand_Err_StatusExecutionNotFound"),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
					StatusCode:  http.StatusNoContent,
					Body:        ``,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
					},
				},
				"GET:/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c": {
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
							"id": "3b32e410-2F33-412D-9fb8-17970131921c",
							"creationTimestamp": "2025-10-17T22:37:53Z",
							"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
						}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c", r.URL.Path)
					},
				},
				"GET:/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/execution/f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3": {
					StatusCode: http.StatusNotFound,
					Body: `{
		  "status": 404,
		  "code": 1003,
		  "messages": [
		    "Seed execution not found: f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3"
		  ],
		  "timestamp": "2025-11-18T01:26:12.946825800Z"
		}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/execution/f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3", r.URL.Path)
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
		  "status": 404,
		  "code": 1003,
		  "messages": [
		    "Seed execution not found: f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3"
		  ],
		  "timestamp": "2025-11-18T01:26:12.946825800Z"
		}`),
			}, "Could not get seed execution with id \"f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3\""),
		},
		{
			name:      "The user sends an invalid execution id",
			args:      []string{"3b32e410-2f33-412d-9fb8-17970131921c", "--execution", "test"},
			url:       true,
			apiKey:    "",
			outGolden: "NewStatusCommand_Out_StatusExecutionInvalidUUID",
			errGolden: "NewStatusCommand_Err_StatusExecutionInvalidUUID",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewStatusCommand_Err_StatusExecutionInvalidUUID"),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
					StatusCode:  http.StatusNoContent,
					Body:        ``,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
					},
				},
				"GET:/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c": {
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
							"id": "3b32e410-2F33-412D-9fb8-17970131921c",
							"creationTimestamp": "2025-10-17T22:37:53Z",
							"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
						}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c", r.URL.Path)
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid UUID length: 4"), "Could not get seed execution id"),
		},
		{
			name:      "The seed has an invalid UUID",
			args:      []string{"3b32e410-2f33-412d-9fb8-17970131921c", "--execution", "f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3"},
			url:       true,
			apiKey:    "",
			outGolden: "NewStatusCommand_Out_SeedWithInvalidUUID",
			errGolden: "NewStatusCommand_Err_SeedWithInvalidUUID",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewStatusCommand_Err_SeedWithInvalidUUID"),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
					StatusCode:  http.StatusNoContent,
					Body:        ``,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
					},
				},
				"GET:/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c": {
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
							"id": "test",
							"creationTimestamp": "2025-10-17T22:37:53Z",
							"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
						}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c", r.URL.Path)
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid UUID length: 4"), "Could not get seed id"),
		},
		{
			name:      "Printing JSON object fails",
			args:      []string{"3b32e410-2f33-412d-9fb8-17970131921c", "--execution", "f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3"},
			outGolden: "NewStatusCommand_Out_PrintJSONFails",
			errGolden: "NewStatusCommand_Err_PrintJSONFails",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewStatusCommand_Err_PrintJSONFails"),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
					StatusCode:  http.StatusNoContent,
					Body:        ``,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
					},
				},
				"GET:/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c": {
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
							"id": "3b32e410-2F33-412D-9fb8-17970131921c",
							"creationTimestamp": "2025-10-17T22:37:53Z",
							"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
						}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c", r.URL.Path)
					},
				},
				"GET:/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/execution/f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3": {
					StatusCode: http.StatusOK,
					Body: `{
							"id": "f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3",
							"creationTimestamp": "2025-10-10T19:48:31Z",
							"lastUpdatedTimestamp": "2025-10-10T19:48:31Z",
							"triggerType": "MANUAL",
							"status: "RUNNING",
							"scanType": "FULL",
							"properties": {
								"stagingBucket": "testBucket"
							},
							"stages": ["BEFORE_HOOKS","INGEST"]
							}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/execution/f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3", r.URL.Path)
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character 'R' after object key"), "Could not print JSON object"),
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
				vpr.Set("default.ingestion_url", srv.URL)
			}
			if tc.apiKey != "" {
				vpr.Set("default.ingestion_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, t.TempDir())

			statusCmd := NewStatusCommand(d)

			statusCmd.SilenceUsage = true
			statusCmd.SetIn(ios.In)
			statusCmd.SetOut(ios.Out)
			statusCmd.SetErr(ios.Err)

			statusCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			statusCmd.SetArgs(tc.args)

			err := statusCmd.Execute()
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

// TestNewStatusCommand_NoProfileFlag tests the NewStatusCommand when the profile flag was not defined.
func TestNewStatusCommand_NoProfileFlag(t *testing.T) {
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

	statusCmd := NewStatusCommand(d)

	statusCmd.SetIn(ios.In)
	statusCmd.SetOut(ios.Out)
	statusCmd.SetErr(ios.Err)

	statusCmd.SetArgs([]string{"test"})

	err := statusCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewStatusCommand_Out_NoProfile", testutils.Read(t, "NewStatusCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewStatusCommand_Err_NoProfile", testutils.Read(t, "NewStatusCommand_Err_NoProfile"), errBuf.Bytes())
}

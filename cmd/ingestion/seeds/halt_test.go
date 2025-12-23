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

// TestNewHaltCommand tests the NewHaltCommand() function.
func TestNewHaltCommand(t *testing.T) {
	tests := []struct {
		name      string
		url       bool
		apiKey    string
		outGolden string
		errGolden string
		outBytes  []byte
		errBytes  []byte
		execution string
		responses map[string]testutils.MockResponse
		err       error
	}{
		// Working case
		{
			name:      "Halt works without execution",
			url:       true,
			apiKey:    "",
			outGolden: "NewHaltCommand_Out_NoExecution",
			errGolden: "NewHaltCommand_Err_NoExecution",
			outBytes:  testutils.Read(t, "NewHaltCommand_Out_NoExecution"),
			errBytes:  []byte(nil),
			execution: "",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB seed",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
					"creationTimestamp": "2025-08-14T18:02:11Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
					"seed": "mongo-seed"
				},
				"highlight": {},
				"singestion": 0.15534057
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
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
					},
				},
				"GET:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "MongoDB seed",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
					"creationTimestamp": "2025-10-17T22:37:53Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
				}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c", r.URL.Path)
					},
				},
				"POST:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c/halt": {
					StatusCode:  http.StatusOK,
					Body:        `[{"id":"cb48ab6b-577a-4354-8edf-981e1b0c9acb","status":202}]`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c/halt", r.URL.Path)
					},
				},
			},
			err: nil,
		},
		{
			name:      "Halt works with execution",
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewHaltCommand_Out_YesExecution",
			errGolden: "NewHaltCommand_Err_YesExecution",
			outBytes:  testutils.Read(t, "NewHaltCommand_Out_YesExecution"),
			errBytes:  []byte(nil),
			execution: "f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB seed",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
					"creationTimestamp": "2025-08-14T18:02:11Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
					"seed": "mongo-seed"
				},
				"highlight": {},
				"singestion": 0.15534057
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
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"GET:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "MongoDB seed",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
					"creationTimestamp": "2025-10-17T22:37:53Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
				}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"POST:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c/execution/f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36/halt": {
					StatusCode:  http.StatusOK,
					Body:        `{"acknowledged":true}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c/execution/f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36/halt", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			outGolden: "NewHaltCommand_Out_NoURL",
			errGolden: "NewHaltCommand_Err_NoURL",
			outBytes:  testutils.Read(t, "NewHaltCommand_Out_NoURL"),
			errBytes:  testutils.Read(t, "NewHaltCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			execution: "",
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Ingestion URL is missing for profile \"default\".\nTo set the URL for the Discovery Ingestion API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery ingestion config --profile \"default\""),
		},
		{
			name:      "Halt fails because the execution is already halting",
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewHaltCommand_Out_ErrorExecutionIsHalting",
			errGolden: "NewHaltCommand_Err_ErrorExecutionIsHalting",
			outBytes:  testutils.Read(t, "NewHaltCommand_Out_ErrorExecutionIsHalting"),
			errBytes:  testutils.Read(t, "NewHaltCommand_Err_ErrorExecutionIsHalting"),
			execution: "f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB seed",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
					"creationTimestamp": "2025-08-14T18:02:11Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
					"seed": "mongo-seed"
				},
				"highlight": {},
				"singestion": 0.15534057
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
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"GET:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "MongoDB seed",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
					"creationTimestamp": "2025-10-17T22:37:53Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
				}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"POST:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c/execution/f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36/halt": {
					StatusCode:  http.StatusConflict,
					Body:        `{"status":409,"code":4001,"messages":["Action HALT cannot be applied to seed execution f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36 because of its current status: HALTING"],"timestamp":"2025-11-05T21:22:43.927371900Z"}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c/execution/f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36/halt", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{"status":409,"code":4001,"messages":["Action HALT cannot be applied to seed execution f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36 because of its current status: HALTING"],"timestamp":"2025-11-05T21:22:43.927371900Z"}`)}, "Could not halt the seed execution with id \"f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36\""),
		},
		{
			name:      "Halt fails because seed was not found",
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewHaltCommand_Out_ErrorSeedNotFound",
			errGolden: "NewHaltCommand_Err_ErrorSeedNotFound",
			outBytes:  testutils.Read(t, "NewHaltCommand_Out_ErrorSeedNotFound"),
			errBytes:  testutils.Read(t, "NewHaltCommand_Err_ErrorSeedNotFound"),
			execution: "",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
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
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
					"status": 404,
					"code": 1003,
					"messages": [
						"Entity not found: 9ababe08-0b74-4672-bb7c-e7a8227d6d4d"
					],
					"timestamp": "2025-10-29T23:12:08.002244700Z"
					}`)}, "Could not get seed ID to halt execution."),
		},
		{
			name:      "Halt fails because the execution not found",
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewHaltCommand_Out_ErrorExecutionNotFound",
			errGolden: "NewHaltCommand_Err_ErrorExecutionNotFound",
			outBytes:  testutils.Read(t, "NewHaltCommand_Out_ErrorExecutionNotFound"),
			errBytes:  testutils.Read(t, "NewHaltCommand_Err_ErrorExecutionNotFound"),
			execution: "f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB seed",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
					"creationTimestamp": "2025-08-14T18:02:11Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
					"seed": "mongo-seed"
				},
				"highlight": {},
				"singestion": 0.15534057
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
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"GET:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "MongoDB seed",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
					"creationTimestamp": "2025-10-17T22:37:53Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
				}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"POST:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c/execution/f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36/halt": {
					StatusCode:  http.StatusNotFound,
					Body:        `{"status":404,"code":1003,"messages":["Seed execution not found: f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36"],"timestamp":"2025-11-05T21:24:31.858049700Z"}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c/execution/f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36/halt", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{"status":404,"code":1003,"messages":["Seed execution not found: f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36"],"timestamp":"2025-11-05T21:24:31.858049700Z"}`)}, "Could not halt the seed execution with id \"f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36\""),
		},
		{
			name:      "Halt fails because the execution is not a UUID",
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewHaltCommand_Out_ErrorExecutionNotUUID",
			errGolden: "NewHaltCommand_Err_ErrorExecutionNotUUID",
			outBytes:  testutils.Read(t, "NewHaltCommand_Out_ErrorExecutionNotUUID"),
			errBytes:  testutils.Read(t, "NewHaltCommand_Err_ErrorExecutionNotUUID"),
			execution: "test",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB seed",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
					"creationTimestamp": "2025-08-14T18:02:11Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
					"seed": "mongo-seed"
				},
				"highlight": {},
				"singestion": 0.15534057
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
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"GET:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "MongoDB seed",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
					"creationTimestamp": "2025-10-17T22:37:53Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
				}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid UUID length: 4"), "Failed to convert the execution ID to UUID"),
		},
		{
			name:      "Printing JSON Object fails",
			outGolden: "NewHaltCommand_Out_PrintJSONFails",
			errGolden: "NewHaltCommand_Err_PrintJSONFails",
			outBytes:  testutils.Read(t, "NewHaltCommand_Out_PrintJSONFails"),
			errBytes:  testutils.Read(t, "NewHaltCommand_Err_PrintJSONFails"),
			url:       true,
			apiKey:    "apiKey123",
			execution: "f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/seed/search": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB seed",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
					"creationTimestamp": "2025-08-14T18:02:11Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
					"seed": "mongo-seed"
				},
				"highlight": {},
				"singestion": 0.15534057
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
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"GET:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "MongoDB seed",
					"labels": [],
					"active": true,
					"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
					"creationTimestamp": "2025-10-17T22:37:53Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
				}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"POST:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c/execution/f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36/halt": {
					StatusCode:  http.StatusOK,
					Body:        `{"acknowledged:true}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c/execution/f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36/halt", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("unexpected end of JSON input"), "Could not print JSON object"),
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

			haltCmd := NewHaltCommand(d)

			haltCmd.SetIn(ios.In)
			haltCmd.SetOut(ios.Out)
			haltCmd.SetErr(ios.Err)

			haltCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			args := []string{"MongoDB seed"}
			if tc.execution != "" {
				args = append(args, "--execution")
				args = append(args, tc.execution)
			}

			haltCmd.SetArgs(args)

			err := haltCmd.Execute()
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

// TestNewHaltCommand_NoProfileFlag tests the NewHaltCommand when the profile flag was not defined.
func TestNewHaltCommand_NoProfileFlag(t *testing.T) {
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

	haltCmd := NewHaltCommand(d)

	haltCmd.SetIn(ios.In)
	haltCmd.SetOut(ios.Out)
	haltCmd.SetErr(ios.Err)

	haltCmd.SetArgs([]string{"MongoDB seed"})

	err := haltCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewHaltCommand_Out_NoProfile", testutils.Read(t, "NewHaltCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewHaltCommand_Err_NoProfile", testutils.Read(t, "NewHaltCommand_Err_NoProfile"), errBuf.Bytes())
}

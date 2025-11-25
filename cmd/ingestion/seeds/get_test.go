package seeds

import (
	"bytes"
	"errors"
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

// NewGetCommand creates the seed get command
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
			args:      []string{"my-seed"},
			url:       true,
			apiKey:    "",
			outGolden: "NewGetCommand_Out_SearchByNameReturnsObject",
			errGolden: "NewGetCommand_Err_SearchByNameReturnsObject",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_SearchByNameReturnsObject"),
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
					"name": "my-seed",
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
				"GET:/v2/seed": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
				"type": "mongo",
				"name": "my-seed",
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
				"name": "OpenAI seed",
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
						assert.Equal(t, "/v2/seed", r.URL.Path)
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
				"POST:/v2/seed/search": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "seed-2",
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
					"name": "my-seed",
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
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},
		{
			name:      "The user gets a record by its id",
			args:      []string{"3b32e410-2f33-412d-9fb8-17970131921c", "--record", "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0="},
			url:       true,
			apiKey:    "",
			outGolden: "NewGetCommand_Out_GetRecordById",
			errGolden: "NewGetCommand_Err_GetRecordById",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_GetRecordById"),
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
				"GET:/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/record/A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=": {
					StatusCode: http.StatusOK,
					Body: `{
						"id": {
							"plain": "4e7c8a47efd829ef7f710d64da661786",
							"hash": "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0="
						},
						"creationTimestamp": "2025-09-04T21:05:25Z",
						"lastUpdatedTimestamp": "2025-09-04T21:05:25Z",
						"status": "SUCCESS"
					}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/record/A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=", r.URL.Path)
					},
				},
			},
			err: nil,
		},
		{
			name:      "The user gets a seed execution by its id and with details",
			args:      []string{"3b32e410-2f33-412d-9fb8-17970131921c", "--execution", "f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3", "--details"},
			url:       true,
			apiKey:    "",
			outGolden: "NewGetCommand_Out_GetExecutionByIdDetails",
			errGolden: "NewGetCommand_Err_GetExecutionByIdDetails",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_GetExecutionByIdDetails"),
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
			name:      "The user gets a seed execution by its id, but with no details",
			args:      []string{"3b32e410-2f33-412d-9fb8-17970131921c", "--execution", "f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3"},
			url:       true,
			apiKey:    "",
			outGolden: "NewGetCommand_Out_GetExecutionByIdNoDetails",
			errGolden: "NewGetCommand_Err_GetExecutionByIdNoDetails",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_GetExecutionByIdNoDetails"),
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
			args:      []string{"3b32e410-2f33-412d-9fb8-17970131921c", "--record", "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0="},
			outGolden: "NewGetCommand_Out_NoURL",
			errGolden: "NewGetCommand_Err_NoURL",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_NoURL"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Ingestion URL is missing for profile \"default\".\nTo set the URL for the Discovery Ingestion API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery ingestion config --profile \"default\""),
		},
		{
			name:      "The user sends no seed",
			args:      []string{"--record", "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0="},
			outGolden: "NewGetCommand_Out_NoSeed",
			errGolden: "NewGetCommand_Err_NoSeed",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_NoSeed"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_NoSeed"),
			url:       true,
			apiKey:    "apiKey123",
			err:       cli.NewError(cli.ErrorExitCode, "Missing the seed"),
		},
		{
			name:      "user sends a name that does not exist",
			args:      []string{"test", "--record", "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0="},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewGetCommand_Out_NameDoesNotExist",
			errGolden: "NewGetCommand_Err_NameDoesNotExist",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_NameDoesNotExist"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_NameDoesNotExist"),
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
			args:      []string{"3b32e410-2F33-412d-9fb8-17970131921c"},
			outGolden: "NewGetCommand_Out_SearchByNameHTTPError",
			errGolden: "NewGetCommand_Err_SearchByNameHTTPError",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_SearchByNameHTTPError"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_SearchByNameHTTPError"),
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
			name:      "GetEntities returns HTTP error",
			args:      []string{},
			outGolden: "NewGetCommand_Out_GetEntitiesHTTPError",
			errGolden: "NewGetCommand_Err_GetEntitiesHTTPError",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_GetEntitiesHTTPError"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_GetEntitiesHTTPError"),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"GET:/v2/seed": {
					StatusCode:  http.StatusUnauthorized,
					Body:        `{"error": "unauthorized"}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed", r.URL.Path)
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
				"POST:/v2/seed/search": {
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
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
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
				"POST:/v2/seed/search": {
					StatusCode:  http.StatusBadRequest,
					Body:        ``,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/search", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewError(cli.ErrorExitCode, "Filter type \"gte\" does not exist"),
		},
		{
			name:      "The user gets a record by its id fails",
			args:      []string{"3b32e410-2f33-412d-9fb8-17970131921c", "--record", "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0="},
			url:       true,
			apiKey:    "",
			outGolden: "NewGetCommand_Out_GetRecordByIdFails",
			errGolden: "NewGetCommand_Err_GetRecordByIdFails",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_GetRecordByIdFails"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_GetRecordByIdFails"),
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
				"GET:/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/record/A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=": {
					StatusCode: http.StatusNotFound,
					Body: `{
  "status": 404,
  "code": 1003,
  "messages": [
    "Entity not found: SeedRecordId(seed=Seed(super=AbstractComponentConfigEntity(super=AbstractJsonConfigEntity(super=AbstractTypedConfigEntity(super=AbstractConfigEntity(super=AbstractUpdatableEntity(super=AbstractCoreEntity(id=2acd0a61-852c-4f38-af2b-9c84e152873e), creationTimestamp=2025-08-21T21:52:03Z, lastUpdatedTimestamp=2025-08-21T21:52:03Z), name=Search seed, description=null, active=true), type=staging), config={\"action\":\"scroll\",\"bucket\":\"blogs\"})), properties=null, labels=[], recordOptions=SeedRecordPolicy[timeoutPolicy=TimeoutPolicy[slice=PT1H], errorPolicy=FATAL, outboundPolicy=OutboundPolicy[idPolicy=IdPolicy[generator=null], batchPolicy=BatchPolicy[maxCount=25, flushAfter=PT1M]]], hooks=[], beforeHooksOptions=null, afterHooksOptions=null), recordId=[3, 113, -45, 12, 72, 2, 107, -82, 65, 21, -101, 26, 115, -44, -56, -100, 88, -84, -66, 90, 17, -108, -67, -52, -25, 72, -93, 9, 99, 66, 43, 31])"
  ],
  "timestamp": "2025-11-10T17:01:44.254941300Z"
}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/seed/3b32e410-2f33-412d-9fb8-17970131921c/record/A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=", r.URL.Path)
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
  "status": 404,
  "code": 1003,
  "messages": [
    "Entity not found: SeedRecordId(seed=Seed(super=AbstractComponentConfigEntity(super=AbstractJsonConfigEntity(super=AbstractTypedConfigEntity(super=AbstractConfigEntity(super=AbstractUpdatableEntity(super=AbstractCoreEntity(id=2acd0a61-852c-4f38-af2b-9c84e152873e), creationTimestamp=2025-08-21T21:52:03Z, lastUpdatedTimestamp=2025-08-21T21:52:03Z), name=Search seed, description=null, active=true), type=staging), config={\"action\":\"scroll\",\"bucket\":\"blogs\"})), properties=null, labels=[], recordOptions=SeedRecordPolicy[timeoutPolicy=TimeoutPolicy[slice=PT1H], errorPolicy=FATAL, outboundPolicy=OutboundPolicy[idPolicy=IdPolicy[generator=null], batchPolicy=BatchPolicy[maxCount=25, flushAfter=PT1M]]], hooks=[], beforeHooksOptions=null, afterHooksOptions=null), recordId=[3, 113, -45, 12, 72, 2, 107, -82, 65, 21, -101, 26, 115, -44, -56, -100, 88, -84, -66, 90, 17, -108, -67, -52, -25, 72, -93, 9, 99, 66, 43, 31])"
  ],
  "timestamp": "2025-11-10T17:01:44.254941300Z"
}`),
			}, "Could not get record with id \"A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=\""),
		},
		{
			name:      "The user tries to get a seed execution by its id, but gets Not Found",
			args:      []string{"3b32e410-2f33-412d-9fb8-17970131921c", "--execution", "f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3"},
			url:       true,
			apiKey:    "",
			outGolden: "NewGetCommand_Out_GetExecutionNotFound",
			errGolden: "NewGetCommand_Err_GetExecutionNotFound",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_GetExecutionNotFound"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_GetExecutionNotFound"),
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
			outGolden: "NewGetCommand_Out_GetExecutionInvalidUUID",
			errGolden: "NewGetCommand_Err_GetExecutionInvalidUUID",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_GetExecutionInvalidUUID"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_GetExecutionInvalidUUID"),
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
			args:      []string{"3b32e410-2f33-412d-9fb8-17970131921c", "--record", "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0="},
			url:       true,
			apiKey:    "",
			outGolden: "NewGetCommand_Out_SeedWithInvalidUUID",
			errGolden: "NewGetCommand_Err_SeedWithInvalidUUID",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_SeedWithInvalidUUID"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_SeedWithInvalidUUID"),
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
			args:      []string{"test"},
			outGolden: "NewGetCommand_Out_PrintJSONFails",
			errGolden: "NewGetCommand_Err_PrintJSONFails",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_PrintJSONFails"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_PrintJSONFails"),
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
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
					"type": "mongo",
					"name": "my-seed",
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
						assert.Equal(t, "/v2/seed/3d51beef-8b90-40aa-84b5-033241dc6239", r.URL.Path)
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
				"GET:/v2/seed": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [{"source":{"active":true,"creationTimestamp":"2025-08-21T17:57:16Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","seeds":[],"lastUpdatedTimestamp":"2025-08-21T17:57:16Z","name":"test","type":"mongo"}},     
			{"source":{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"5f125024-1e5e-4591-9fee-365dc20eeeed","seeds":[],"lastUpdatedTimestamp":"2025-08-18T20:55:43Z","name":"MongoDB text seed","type":mongo}},       
			{"source":{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"86e7f920-a4e4-4b64-be84-5437a7673db8","seeds":[],"lastUpdatedTimestamp":"2025-08-14T18:02:38Z","name":"Script seed","type":"script"}}
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
						assert.Equal(t, "/v2/seed", r.URL.Path)
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
			vpr.Set("output", "json")
			if tc.url {
				vpr.Set("default.ingestion_url", srv.URL)
			}
			if tc.apiKey != "" {
				vpr.Set("default.ingestion_key", tc.apiKey)
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

	vpr.Set("default.ingestion_url", "test")
	vpr.Set("default.ingestion_key", "test")

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

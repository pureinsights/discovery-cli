package servers

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

// NewGetCommand creates the server get command
func TestNewGetCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		url        bool
		apiKey     string
		outGolden  string
		errGolden  string
		method     string
		path       string
		statusCode int
		response   string
		err        error
	}{
		// Working case
		{
			name:       "Search by name returns an array of which the first object is returned",
			args:       []string{"MongoDB Atlas Server Clone 2"},
			url:        true,
			apiKey:     "apiKey123",
			outGolden:  "NewGetCommand_Out_SearchByNameReturnsObject",
			errGolden:  "NewGetCommand_Err_SearchByNameReturnsObject",
			method:     http.MethodPost,
			path:       "/server/search",
			statusCode: http.StatusOK,
			response: `{
			"content": [
				{
				"source": {
				"type": "mongo",
				"name": "MongoDB Atlas Server Clone 2",
				"labels": [
					{
						"key": "A",
						"value": "A"
					}
				],
				"active": true,
				"id": "21029da3-041c-43b5-a67e-870251f2f6a6",
				"creationTimestamp": "2025-09-29T15:50:19Z",
				"lastUpdatedTimestamp": "2025-09-29T15:50:19Z"
				},
				"highlight": {},
				"score": 0.30723327
			},
			{
				"source": {
				"type": "mongo",
				"name": "MongoDB Atlas server clone 4",
				"labels": [
					{
						"key": "A",
						"value": "A"
					}
				],
				"active": true,
				"id": "a798cd5b-aa7a-4fc5-9292-1de6fe8e8b7f",
				"creationTimestamp": "2025-09-29T15:50:21Z",
				"lastUpdatedTimestamp": "2025-09-29T15:50:21Z"
				},
				"highlight": {},
				"score": 0.30723327
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
			err: nil,
		},
		{
			name:       "Get with no args returns an array",
			args:       []string{},
			outGolden:  "NewGetCommand_Out_GetAllReturnsArray",
			errGolden:  "NewGetCommand_Err_GetAllReturnsArray",
			url:        true,
			apiKey:     "apiKey123",
			method:     http.MethodGet,
			path:       "/server",
			statusCode: http.StatusOK,
			response: `{
			"content": [
			{
				"type": "mongo",
				"name": "MongoDB Atlas server clone 6",
				"labels": [],
				"active": true,
				"id": "226e8a0b-5016-4ebe-9963-1461edd39d0a",
				"creationTimestamp": "2025-09-29T15:50:22Z",
				"lastUpdatedTimestamp": "2025-09-29T15:50:22Z"
			},
			{
				"type": "mongo",
				"name": "MongoDB Atlas server clone 9",
				"labels": [],
				"active": true,
				"id": "2b839453-ddad-4ced-8e13-2c7860af60a7",
				"creationTimestamp": "2025-09-29T15:50:26Z",
				"lastUpdatedTimestamp": "2025-09-29T15:50:26Z"
			},
			{
				"type": "mongo",
				"name": "MongoDB Atlas server clone 3",
				"labels": [],
				"active": true,
				"id": "3a0214a4-72cc-4eee-ad0c-9e3af9b08a6c",
				"creationTimestamp": "2025-09-29T15:50:20Z",
				"lastUpdatedTimestamp": "2025-09-29T15:50:20Z"
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
			err: nil,
		},
		{
			name:       "Get with args returns a search array",
			args:       []string{"--filter", "type=mongo"},
			outGolden:  "NewGetCommand_Out_SearchWithFiltersReturnsArray",
			errGolden:  "NewGetCommand_Err_SearchWithFiltersReturnsArray",
			url:        true,
			apiKey:     "apiKey123",
			method:     http.MethodPost,
			path:       "/server/search",
			statusCode: http.StatusOK,
			response: `{
			"content": [
			{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone",
					"labels": [],
					"active": true,
					"id": "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
					"creationTimestamp": "2025-09-29T15:50:17Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:17Z"
				},
				"highlight": {},
				"score": 0.20970252
			},
			{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone 1",
					"labels": [],
					"active": true,
					"id": "8f14c11c-bb66-49d3-aa2a-dedff4608c17",
					"creationTimestamp": "2025-09-29T15:50:19Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:19Z"
				},
				"highlight": {},
				"score": 0.20970252
			},
            {
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone 3",
					"labels": [],
					"active": true,
					"id": "3a0214a4-72cc-4eee-ad0c-9e3af9b08a6c",
					"creationTimestamp": "2025-09-29T15:50:20Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:20Z"
				},
				"highlight": {},
				"score": 0.20970252
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
			err: nil,
		},

		// Error case
		{
			name:       "No URL",
			args:       []string{},
			outGolden:  "NewGetCommand_Out_NoURL",
			errGolden:  "NewGetCommand_Err_NoURL",
			url:        false,
			apiKey:     "apiKey123",
			method:     http.MethodPost,
			path:       "/server/search",
			statusCode: http.StatusOK,
			response:   ``,
			err:        cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile {profile}\n      discovery core config --profile {profile}"),
		},
		{
			name:       "No API key",
			args:       []string{},
			outGolden:  "NewGetCommand_Out_NoAPIKey",
			errGolden:  "NewGetCommand_Err_NoAPIKey",
			url:        true,
			apiKey:     "",
			method:     http.MethodPost,
			path:       "/server",
			statusCode: http.StatusNotFound,
			response:   ``,
			err:        cli.NewError(cli.ErrorExitCode, "The Discovery Core API key is missing for profile \"default\".\nTo set the API key for the Discovery Core API, run any of the following commands:\n      discovery config  --profile {profile}\n      discovery core config --profile {profile}"),
		},
		{
			name:       "user sends a name that does not exist",
			args:       []string{"test"},
			url:        true,
			apiKey:     "apiKey123",
			outGolden:  "NewGetCommand_Out_NameDoesNotExist",
			errGolden:  "NewGetCommand_Err_NameDoesNotExist",
			method:     http.MethodPost,
			path:       "/server/search",
			statusCode: http.StatusNoContent,
			response:   ``,
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name "test" does not exist"
	],
	"timestamp": "2025-09-30T15:38:42.885125200Z"
}`),
			}, "Could not search for entity with id \"test\""),
		},

		{
			name:       "Search By Name returns HTTP error",
			args:       []string{"3b32e410-2F33-412d-9fb8-17970131921c"},
			outGolden:  "NewGetCommand_Out_SearchByNameHTTPError",
			errGolden:  "NewGetCommand_Err_SearchByNameHTTPError",
			url:        true,
			apiKey:     "apiKey123",
			method:     http.MethodPost,
			path:       "/server/search",
			statusCode: http.StatusInternalServerError,
			response: `{
			"status": 500,
			"code": 1003,
			"messages": [
				"Internal server error"
			],
			"timestamp": "2025-10-16T17:46:45.386963700Z"
			}`,
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
			name:       "GetEntities returns HTTP error",
			args:       []string{},
			outGolden:  "NewGetCommand_Out_GetEntitiesHTTPError",
			errGolden:  "NewGetCommand_Err_GetEntitiesHTTPError",
			url:        true,
			apiKey:     "apiKey123",
			method:     http.MethodGet,
			path:       "/server",
			statusCode: http.StatusUnauthorized,
			response:   `{"error": "unauthorized"}`,
			err:        cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error": "unauthorized"}`)}, "Could not get all entities"),
		},
		{
			name:       "SearchEntities returns HTTP error",
			args:       []string{"--filter", "label=A"},
			url:        true,
			apiKey:     "apiKey123",
			outGolden:  "NewGetCommand_Out_SearchHTTPError",
			errGolden:  "NewGetCommand_Err_SearchHTTPError",
			method:     http.MethodPost,
			path:       "/server/search",
			statusCode: http.StatusUnauthorized,
			response: `{
	"status": 401,
	"code": 1003,
	"messages": [
		"user is unauthorized"
	],
	"timestamp": "2025-09-30T15:38:42.885125200Z"
}`,
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
			name:       "Filter does not exist",
			args:       []string{"--filter", "gte=field:1"},
			url:        true,
			apiKey:     "apiKey123",
			outGolden:  "NewGetCommand_Out_FilterDoesNotExist",
			errGolden:  "NewGetCommand_Err_FilterDoesNotExist",
			method:     http.MethodPost,
			path:       "/server/search",
			statusCode: http.StatusBadRequest,
			response:   ``,
			err:        cli.NewError(cli.ErrorExitCode, "Filter type \"gte\" does not exist"),
		},
		{
			name:       "Printing JSON object fails",
			args:       []string{"test"},
			outGolden:  "NewGetCommand_Out_PrintJSONFails",
			errGolden:  "NewGetCommand_Err_PrintJSONFails",
			url:        true,
			apiKey:     "apiKey123",
			method:     http.MethodPost,
			path:       "/server/search",
			statusCode: http.StatusOK,
			response: `{
			"content": [{"source":{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"5f125024-1e5e-4591-9fee-365dc20eeeed","credentials":[],"lastUpdatedTimestamp":"2025-08-18T20:55:43Z","name":"test","type":mongo"}},       
			{{"source":{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"86e7f920-a4e4-4b64-be84-5437a7673db8","credentials":[],"lastUpdatedTimestamp":"2025-08-14T18:02:38Z","name":"Script processor","type":"script"}],
			"pageable": {
				"page": 0,
				"size": 3,
				"sort": []
			}},
			"totalSize": 2,
			"totalPages": 4,
			"empty": false,
			"size": 3,
			"offset": 0,
			"numberOfElements": 2,
			"pageNumber": 0
			}`,
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character 'm' looking for beginning of value"), "Could not print JSON object"),
		},
		{
			name:       "Printing JSON array fails",
			args:       []string{},
			outGolden:  "NewGetCommand_Out_PrintArrayFails",
			errGolden:  "NewGetCommand_Err_PrintArrayFails",
			url:        true,
			apiKey:     "apiKey123",
			method:     http.MethodGet,
			path:       "/server",
			statusCode: http.StatusOK,
			response: `{
			"content": [{"source":{"active":true,"creationTimestamp":"2025-08-21T17:57:16Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","credentials":[],"lastUpdatedTimestamp":"2025-08-21T17:57:16Z","name":"test","type":"mongo"}},     
			{"source":{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"5f125024-1e5e-4591-9fee-365dc20eeeed","credentials":[],"lastUpdatedTimestamp":"2025-08-18T20:55:43Z","name":"MongoDB text processor","type":mongo}},       
			{"source":{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"86e7f920-a4e4-4b64-be84-5437a7673db8","credentials":[],"lastUpdatedTimestamp":"2025-08-14T18:02:38Z","name":"Script processor","type":"script"}}
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
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character 'm' looking for beginning of value"), "Could not print JSON array"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, tc.path, r.URL.Path)
				assert.Equal(t, tc.apiKey, r.Header.Get("X-API-Key"))
			}))

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
			} else {
				require.NoError(t, err)
			}

			testutils.CompareBytes(t, tc.outGolden, out.Bytes())
			testutils.CompareBytes(t, tc.errGolden, errBuf.Bytes())
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

	testutils.CompareBytes(t, "NewGetCommand_Out_NoProfile", out.Bytes())
	testutils.CompareBytes(t, "NewGetCommand_Err_NoProfile", errBuf.Bytes())
}

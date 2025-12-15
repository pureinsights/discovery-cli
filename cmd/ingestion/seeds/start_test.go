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

// TestNewStartCommand_WithProfileFlag tests the NewStartCommand function.
func TestNewStartCommand_WithProfileFlag(t *testing.T) {
	tests := []struct {
		name                string
		url                 bool
		apiKey              string
		outGolden           string
		errGolden           string
		outBytes            []byte
		errBytes            []byte
		scanType            string
		executionProperties string
		responses           map[string]testutils.MockResponse
		err                 error
	}{
		// Working case
		{
			name:                "Start works without executionProperties and with scanType",
			url:                 true,
			apiKey:              "",
			outGolden:           "NewStartCommand_Out_NoPropertiesYesScan",
			errGolden:           "NewStartCommand_Err_NoPropertiesYesScan",
			outBytes:            testutils.Read(t, "NewStartCommand_Out_NoPropertiesYesScan"),
			errBytes:            []byte(nil),
			scanType:            "FULL",
			executionProperties: "",
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
				"POST:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c": {
					StatusCode:  http.StatusOK,
					Body:        `{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","creationTimestamp":"2025-09-04T19:29:41.119013Z","lastUpdatedTimestamp":"2025-09-04T19:29:41.119013Z","triggerType":"MANUAL","status":"CREATED","scanType":"FULL"}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c", r.URL.Path)
					},
				},
			},
			err: nil,
		},
		{
			name:                "Start works with executionProperties and without scan type",
			url:                 true,
			apiKey:              "apiKey123",
			outGolden:           "NewStartCommand_Out_YesPropertiesNoScan",
			errGolden:           "NewStartCommand_Err_YesPropertiesNoScan",
			outBytes:            testutils.Read(t, "NewStartCommand_Out_YesPropertiesNoScan"),
			errBytes:            []byte(nil),
			scanType:            "",
			executionProperties: "{\"stagingBucket\":\"testBucket\"}",
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
				"POST:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c": {
					StatusCode:  http.StatusOK,
					Body:        `{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","creationTimestamp":"2025-09-04T19:29:41.119013Z","lastUpdatedTimestamp":"2025-09-04T19:29:41.119013Z","triggerType":"MANUAL","status":"CREATED","scanType":"INCREMENTAL","properties":{"stagingBucket":"testBucket"}}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},

		// Error case
		{
			name:                "No URL",
			outGolden:           "NewStartCommand_Out_NoURL",
			errGolden:           "NewStartCommand_Err_NoURL",
			outBytes:            testutils.Read(t, "NewStartCommand_Out_NoURL"),
			errBytes:            testutils.Read(t, "NewStartCommand_Err_NoURL"),
			url:                 false,
			apiKey:              "apiKey123",
			executionProperties: "",
			scanType:            "FULL",
			err:                 cli.NewError(cli.ErrorExitCode, "The Discovery Ingestion URL is missing for profile \"default\".\nTo set the URL for the Discovery Ingestion API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery ingestion config --profile \"default\""),
		},
		{
			name:                "Start fails because it already has an execution",
			url:                 true,
			apiKey:              "apiKey123",
			outGolden:           "NewStartCommand_Out_ErrorHasActiveExecutions",
			errGolden:           "NewStartCommand_Err_ErrorHasActiveExecutions",
			outBytes:            testutils.Read(t, "NewStartCommand_Out_ErrorHasActiveExecutions"),
			errBytes:            testutils.Read(t, "NewStartCommand_Err_ErrorHasActiveExecutions"),
			scanType:            "FULL",
			executionProperties: "",
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
				"POST:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c": {
					StatusCode: http.StatusConflict,
					Body: `{
			"status": 409,
			"code": 4001,
			"messages": [
				"The seed has 1 executions: 0c309dbb-0402-4710-8659-2c75f5d649b6"
			],
			"timestamp": "2025-09-04T20:17:00.116546400Z"
			}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
			"status": 409,
			"code": 4001,
			"messages": [
				"The seed has 1 executions: 0c309dbb-0402-4710-8659-2c75f5d649b6"
			],
			"timestamp": "2025-09-04T20:17:00.116546400Z"
			}`)}, "Could not start seed execution for seed with id \"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\""),
		},
		{
			name:                "Start fails because seed was not found",
			url:                 true,
			apiKey:              "apiKey123",
			outGolden:           "NewStartCommand_Out_ErrorNotFound",
			errGolden:           "NewStartCommand_Err_ErrorNotFound",
			outBytes:            testutils.Read(t, "NewStartCommand_Out_ErrorNotFound"),
			errBytes:            testutils.Read(t, "NewStartCommand_Err_ErrorNotFound"),
			scanType:            "FULL",
			executionProperties: "",
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
					}`)}, "Could not get seed ID to start execution."),
		},
		{
			name:                "Printing JSON Object fails",
			outGolden:           "NewStartCommand_Out_PrintJSONFails",
			errGolden:           "NewStartCommand_Err_PrintJSONFails",
			outBytes:            testutils.Read(t, "NewStartCommand_Out_PrintJSONFails"),
			errBytes:            testutils.Read(t, "NewStartCommand_Err_PrintJSONFails"),
			url:                 true,
			apiKey:              "apiKey123",
			scanType:            "FULL",
			executionProperties: "",
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
				"POST:/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c": {
					StatusCode:  http.StatusOK,
					Body:        `{"id:"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","creationTimestamp":"2025-09-04T19:29:41.119013Z","lastUpdatedTimestamp":"2025-09-04T19:29:41.119013Z","triggerType":"MANUAL","status":"CREATED","scanType":"INCREMENTAL","properties":{"stagingBucket":"testBucket"}}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/seed/9ababe08-0b74-4672-bb7c-e7a8227d6d4c", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character 'a' after object key"), "Could not print JSON object"),
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

			startCmd := NewStartCommand(d)

			startCmd.SetIn(ios.In)
			startCmd.SetOut(ios.Out)
			startCmd.SetErr(ios.Err)

			startCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			args := []string{"MongoDB seed"}
			if tc.executionProperties != "" {
				args = append(args, "--properties")
				args = append(args, tc.executionProperties)
			}

			if tc.scanType != "" {
				args = append(args, "--scan-type")
				args = append(args, tc.scanType)
			}

			startCmd.SetArgs(args)

			err := startCmd.Execute()
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

// TestNewStartCommand_NoProfileFlag tests the NewStartCommand when the profile flag was not defined.
func TestNewStartCommand_NoProfileFlag(t *testing.T) {
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

	startCmd := NewStartCommand(d)

	startCmd.SetIn(ios.In)
	startCmd.SetOut(ios.Out)
	startCmd.SetErr(ios.Err)

	startCmd.SetArgs([]string{"MongoDB seed"})

	err := startCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewStartCommand_Out_NoProfile", testutils.Read(t, "NewStartCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewStartCommand_Err_NoProfile", testutils.Read(t, "NewStartCommand_Err_NoProfile"), errBuf.Bytes())
}

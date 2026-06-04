package buckets

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
		url       bool
		apiKey    string
		outGolden string
		errGolden string
		outBytes  []byte
		errBytes  []byte
		args      []string
		filters   []string
		responses map[string]testutils.MockResponse
		err       error
	}{
		// Working cases
		{
			name:      "Get all buckets",
			url:       true,
			apiKey:    "",
			outGolden: "NewGetCommand_Out_GetAllBuckets",
			errGolden: "NewGetCommand_Err_GetAllBuckets",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_GetAllBuckets"),
			errBytes:  []byte(nil),
			args:      []string{},
			filters:   []string{},
			responses: map[string]testutils.MockResponse{
				"GET:/v2/bucket": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
					"name": "my-bucket-b",
					"description": "description",
					"labels": [
						{
							"key": "type",
							"value": "test"
						}
					],
					"active": false,
					"id": "2831800d-289a-4a06-a65b-c3f2499547f8",
					"creationTimestamp": "2026-06-04T22:07:56Z",
					"lastUpdatedTimestamp": "2026-06-04T22:20:23Z"
				},
				{
					"name": "my-bucket",
					"description": "description",
					"labels": [],
					"active": true,
					"id": "69eeb20b-8ded-478f-937f-64caa0a3e8c0",
					"creationTimestamp": "2026-06-04T22:06:02Z",
					"lastUpdatedTimestamp": "2026-06-04T22:06:02Z"
				}
			],
			"pageable": {
				"page": 0,
				"size": 20,
				"sort": []
			},
			"totalSize": 2,
			"totalPages": 1,
			"empty": false,
			"size": 20,
			"offset": 0,
			"numberOfElements": 2,
			"pageNumber": 0
		}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/bucket", r.URL.Path)
					},
				},
			},
			err: nil,
		},
		{
			name:      "Search by name returns an array of which the first object is returned",
			args:      []string{"my-bucket"},
			url:       true,
			apiKey:    "",
			outGolden: "NewGetCommand_Out_SearchByNameReturnsObject",
			errGolden: "NewGetCommand_Err_SearchByNameReturnsObject",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_SearchByNameReturnsObject"),
			errBytes:  []byte(nil),
			filters:   []string{},
			responses: map[string]testutils.MockResponse{
				"POST:/v2/bucket/search": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
				"source": {
					"name": "my-bucket",
					"description": "description",
					"labels": [],
					"active": true,
					"id": "69eeb20b-8ded-478f-937f-64caa0a3e8c0",
					"creationTimestamp": "2026-06-04T22:06:02Z",
					"lastUpdatedTimestamp": "2026-06-04T22:06:02Z"
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
						assert.Equal(t, "/v2/bucket/search", r.URL.Path)
					},
				},
				"GET:/v2/bucket/69eeb20b-8ded-478f-937f-64caa0a3e8c0": {
					StatusCode: http.StatusOK,
					Body: `{
			"name": "my-bucket",
			"documentCount": {},
			"indices": [
				{
				"name": "myIndexA",
				"fields": [
					{
					"fieldName": "ASC"
					}
				],
				"unique": false
				}
			]
			}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/bucket/69eeb20b-8ded-478f-937f-64caa0a3e8c0", r.URL.Path)
					},
				},
			},
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			url:       false,
			apiKey:    "apiKey123",
			outGolden: "NewGetCommand_Out_NoURL",
			errGolden: "NewGetCommand_Err_NoURL",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_NoURL"),
			args:      []string{},
			filters:   []string{},
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Staging URL is missing for profile \"default\".\nTo set the URL for the Discovery Staging API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery staging config --profile \"default\""),
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
			filters:   []string{},
			responses: map[string]testutils.MockResponse{
				"/v2/bucket/search": {
					StatusCode:  http.StatusNoContent,
					Body:        ``,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/bucket/search", r.URL.Path)
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
				vpr.Set("default.staging_url", srv.URL)
			}
			if tc.apiKey != "" {
				vpr.Set("default.staging_key", tc.apiKey)
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

			args := tc.args
			for _, f := range tc.filters {
				args = append(args, "--filter", f)
			}
			getCmd.SetArgs(args)

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
	vpr.Set("default.staging_url", "test")
	vpr.Set("default.staging_key", "test")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	getCmd := NewGetCommand(d)

	getCmd.SetIn(ios.In)
	getCmd.SetOut(ios.Out)
	getCmd.SetErr(ios.Err)

	getCmd.SetArgs([]string{"my-bucket"})

	err := getCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewGetCommand_Out_NoProfile", testutils.Read(t, "NewGetCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewGetCommand_Err_NoProfile", testutils.Read(t, "NewGetCommand_Err_NoProfile"), errBuf.Bytes())
}

package buckets

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery/v2"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestNewCountCommand_ErrorCases tests the NewCountCommand() function's error cases.
func TestNewCountCommand_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		url       bool
		apiKey    string
		errGolden string
		errBytes  []byte
		responses map[string]testutils.MockResponse
		err       error
	}{
		// Error case
		{
			name:      "No URL",
			args:      []string{"my-bucket"},
			errGolden: "NewCountCommand_Err_NoURL",
			errBytes:  testutils.Read(t, "NewCountCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Staging URL is missing for profile \"default\".\nTo set the URL for the Discovery Staging API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery staging config --profile \"default\""),
		},
		{
			name:      "sent name does not exist",
			args:      []string{"my-bucket"},
			url:       true,
			apiKey:    "apiKey123",
			errGolden: "NewCountCommand_Err_NameDoesNotExist",
			errBytes:  testutils.Read(t, "NewCountCommand_Err_NameDoesNotExist"),
			responses: map[string]testutils.MockResponse{},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name "my-bucket" does not exist"
	]
}`),
			}, "Could not find bucket with name or id \"my-bucket\""),
		},
		{
			name:      "count fails",
			args:      []string{"my-bucket"},
			url:       true,
			apiKey:    "apiKey123",
			errGolden: "NewCountCommand_Err_CountFails",
			errBytes:  testutils.Read(t, "NewCountCommand_Err_CountFails"),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/bucket/search": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
	"content": [
		{
			"source": {
				"id": "fbe3e8ab-44a7-4b8f-b696-cbfc528d9bb0",
				"name": "my-bucket",
				"active": true
			},
			"highlight": {},
			"score": 1.0
		}
	],
	"empty": false
}`,
				},
				"GET:/v2/bucket/fbe3e8ab-44a7-4b8f-b696-cbfc528d9bb0": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
	"id": "fbe3e8ab-44a7-4b8f-b696-cbfc528d9bb0",
	"name": "my-bucket",
	"active": true
}`,
				},
				"POST:/v2/content/my-bucket/count": {
					StatusCode:  http.StatusInternalServerError,
					ContentType: "application/json",
					Body:        `{"status": 500, "code": 5000, "messages": ["Internal server error"]}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/content/my-bucket/count", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusInternalServerError,
				Body:   gjson.Parse(`{"status": 500, "code": 5000, "messages": ["Internal server error"]}`),
			}, "Could not count the bucket with name \"my-bucket\"."),
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

			countCmd := NewCountCommand(d)

			countCmd.SilenceUsage = true
			countCmd.SetIn(ios.In)
			countCmd.SetOut(ios.Out)
			countCmd.SetErr(ios.Err)

			countCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			countCmd.SetArgs(tc.args)

			err := countCmd.Execute()
			if tc.err != nil {
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
				testutils.CompareBytes(t, tc.errGolden, tc.errBytes, errBuf.Bytes())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestNewCountCommand_WorkingCase tests the Count command with a working count.
func TestNewCountCommand_WorkingCase(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/v2/bucket/search":
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"content": [
					{
						"source": {
							"id": "fbe3e8ab-44a7-4b8f-b696-cbfc528d9bb0",
							"name": "my-bucket",
							"active": true
						},
						"highlight": {},
						"score": 1.0
					}
				],
				"empty": false
			}`))
		case "/v2/bucket/fbe3e8ab-44a7-4b8f-b696-cbfc528d9bb0":
			assert.Equal(t, http.MethodGet, r.Method)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"id": "fbe3e8ab-44a7-4b8f-b696-cbfc528d9bb0",
				"name": "my-bucket",
				"active": true
			}`))
		case "/v2/content/my-bucket/count":
			assert.Equal(t, http.MethodPost, r.Method)
			body := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(body)
			assert.JSONEq(t, `{
	"equals": {
		"field": "author",
		"value": "Martin Bayton",
		"normalize": true
	}
}`, string(body))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"total": 278930}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

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
	vpr.Set("default.staging_url", srv.URL)
	vpr.Set("default.staging_key", "")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	countCmd := NewCountCommand(d)

	countCmd.SilenceUsage = true
	countCmd.SetIn(ios.In)
	countCmd.SetOut(ios.Out)
	countCmd.SetErr(ios.Err)

	countCmd.PersistentFlags().StringP(
		"profile",
		"p",
		d.Config().GetString("profile"),
		"configuration profile to use",
	)

	countCmd.SetArgs([]string{"my-bucket", "--filter", `{
	"equals": {
		"field": "author",
		"value": "Martin Bayton",
		"normalize": true
	}
}`})

	err := countCmd.Execute()
	require.NoError(t, err)
	testutils.CompareBytes(t, "NewCountCommand_Out_WorkingCount", testutils.Read(t, "NewCountCommand_Out_WorkingCount"), out.Bytes())
}

// TestNewCountCommand_NoProfileFlag tests the NewCountCommand when the profile flag was not defined.
func TestNewCountCommand_NoProfileFlag(t *testing.T) {
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

	vpr.Set("default.staging_url", "test")
	vpr.Set("default.staging_key", "test")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	countCmd := NewCountCommand(d)

	countCmd.SetIn(ios.In)
	countCmd.SetOut(ios.Out)
	countCmd.SetErr(ios.Err)

	countCmd.SetArgs([]string{"my-bucket"})

	err := countCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewCountCommand_Out_NoProfile", testutils.Read(t, "NewCountCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewCountCommand_Err_NoProfile", testutils.Read(t, "NewCountCommand_Err_NoProfile"), errBuf.Bytes())
}

// TestNewCountCommand_NotExactly1Arg tests the NewCountCommand function when it does not receive exactly one argument.
func TestNewCountCommand_NotExactly1Arg(t *testing.T) {
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

	vpr.Set("default.staging_url", "test")
	vpr.Set("default.staging_key", "test")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	countCmd := NewCountCommand(d)

	countCmd.SetIn(ios.In)
	countCmd.SetOut(ios.Out)
	countCmd.SetErr(ios.Err)

	countCmd.SetArgs([]string{})

	err := countCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")

	testutils.CompareBytes(t, "NewCountCommand_Out_NotExactly1Arg", testutils.Read(t, "NewCountCommand_Out_NotExactly1Arg"), out.Bytes())
	testutils.CompareBytes(t, "NewCountCommand_Err_NotExactly1Arg", testutils.Read(t, "NewCountCommand_Err_NotExactly1Arg"), errBuf.Bytes())
}

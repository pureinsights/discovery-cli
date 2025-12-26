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

// TestNewDumpCommand tests the NewDumpCommand function.
func TestNewDumpCommand_ErrorCases(t *testing.T) {
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
		// Error case
		{
			name:      "No URL",
			args:      []string{"my-bucket"},
			outGolden: "NewDumpCommand_Out_NoURL",
			errGolden: "NewDumpCommand_Err_NoURL",
			outBytes:  testutils.Read(t, "NewDumpCommand_Out_NoURL"),
			errBytes:  testutils.Read(t, "NewDumpCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Staging URL is missing for profile \"default\".\nTo set the URL for the Discovery Staging API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery staging config --profile \"default\""),
		},
		{
			name:      "sent name does not exist",
			args:      []string{"my-bucket"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewDumpCommand_Out_NameDoesNotExist",
			errGolden: "NewDumpCommand_Err_NameDoesNotExist",
			outBytes:  testutils.Read(t, "NewDumpCommand_Out_NameDoesNotExist"),
			errBytes:  testutils.Read(t, "NewDumpCommand_Err_NameDoesNotExist"),
			responses: map[string]testutils.MockResponse{
				"DUMP:/v2/bucket/my-bucket": {
					StatusCode:  http.StatusNotFound,
					ContentType: "application/json",
					Body: `{
  "status": 404,
  "code": 1002,
  "messages": [
    "The bucket 'my-bucket' was not found."
  ],
  "timestamp": "2025-12-23T14:53:32.321524600Z"
}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/bucket/my-bucket", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
  "status": 404,
  "code": 1002,
  "messages": [
    "The bucket 'my-bucket' was not found."
  ],
  "timestamp": "2025-12-23T14:53:32.321524600Z"
}`),
			}, "Could not dump the bucket with name \"my-bucket\"."),
		},
		{
			name:      "Printing JSON object fails",
			args:      []string{"my-bucket"},
			outGolden: "NewDumpCommand_Out_PrintJSONFails",
			errGolden: "NewDumpCommand_Err_PrintJSONFails",
			outBytes:  testutils.Read(t, "NewDumpCommand_Out_PrintJSONFails"),
			errBytes:  testutils.Read(t, "NewDumpCommand_Err_PrintJSONFails"),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"DUMP:/v2/bucket/my-bucket": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
				"acknowledged: true
			}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/bucket/my-bucket", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character '\\n' in string literal"), "Could not print JSON object"),
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

			dumpCmd := NewDumpCommand(d)

			dumpCmd.SetIn(ios.In)
			dumpCmd.SetOut(ios.Out)
			dumpCmd.SetErr(ios.Err)

			dumpCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			dumpCmd.SetArgs(tc.args)

			err := dumpCmd.Execute()
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

// TestNewDumpCommand_WorkingCase tests the Dump command with a working scroll.
func TestNewDumpCommand_WorkingCase(t *testing.T) {

}

// TestNewDumpCommand_NoProfileFlag tests the NewDumpCommand when the profile flag was not defined.
func TestNewDumpCommand_NoProfileFlag(t *testing.T) {
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

	dumpCmd := NewDumpCommand(d)

	dumpCmd.SetIn(ios.In)
	dumpCmd.SetOut(ios.Out)
	dumpCmd.SetErr(ios.Err)

	dumpCmd.SetArgs([]string{"my-bucket"})

	err := dumpCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewDumpCommand_Out_NoProfile", testutils.Read(t, "NewDumpCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewDumpCommand_Err_NoProfile", testutils.Read(t, "NewDumpCommand_Err_NoProfile"), errBuf.Bytes())
}

// TestNewDumpCommand_NotExactly1Arg tests the NewDumpCommand function when it does not receive exactly one argument.
func TestNewDumpCommand_NotExactly1Arg(t *testing.T) {
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

	dumpCmd := NewDumpCommand(d)

	dumpCmd.SetIn(ios.In)
	dumpCmd.SetOut(ios.Out)
	dumpCmd.SetErr(ios.Err)

	dumpCmd.SetArgs([]string{})

	err := dumpCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")

	testutils.CompareBytes(t, "NewDumpCommand_Out_NotExactly1Arg", testutils.Read(t, "NewDumpCommand_Out_NotExactly1Arg"), out.Bytes())
	testutils.CompareBytes(t, "NewDumpCommand_Err_NotExactly1Arg", testutils.Read(t, "NewDumpCommand_Err_NotExactly1Arg"), errBuf.Bytes())
}

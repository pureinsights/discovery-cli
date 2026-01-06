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

// TestNewDeleteCommand tests the NewDeleteCommand function.
func TestNewDeleteCommand(t *testing.T) {
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
			name:      "Delete bucket returns an acknowledged true",
			args:      []string{"my-bucket"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewDeleteCommand_Out_DeleteBucketReturnsTrue",
			errGolden: "NewDeleteCommand_Err_DeleteBucketReturnsTrue",
			outBytes:  testutils.Read(t, "NewDeleteCommand_Out_DeleteBucketReturnsTrue"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"DELETE:/v2/bucket/my-bucket": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
	"acknowledged": true
}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodDelete, r.Method)
						assert.Equal(t, "/v2/bucket/my-bucket", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			args:      []string{"my-bucket"},
			outGolden: "NewDeleteCommand_Out_NoURL",
			errGolden: "NewDeleteCommand_Err_NoURL",
			outBytes:  testutils.Read(t, "NewDeleteCommand_Out_NoURL"),
			errBytes:  testutils.Read(t, "NewDeleteCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Staging URL is missing for profile \"default\".\nTo set the URL for the Discovery Staging API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery staging config --profile \"default\""),
		},
		{
			name:      "sent name does not exist",
			args:      []string{"my-bucket"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewDeleteCommand_Out_NameDoesNotExist",
			errGolden: "NewDeleteCommand_Err_NameDoesNotExist",
			outBytes:  testutils.Read(t, "NewDeleteCommand_Out_NameDoesNotExist"),
			errBytes:  testutils.Read(t, "NewDeleteCommand_Err_NameDoesNotExist"),
			responses: map[string]testutils.MockResponse{
				"DELETE:/v2/bucket/my-bucket": {
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
						assert.Equal(t, http.MethodDelete, r.Method)
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
			}, "Could not delete the bucket with name \"my-bucket\"."),
		},
		{
			name:      "Printing JSON object fails",
			args:      []string{"my-bucket"},
			outGolden: "NewDeleteCommand_Out_PrintJSONFails",
			errGolden: "NewDeleteCommand_Err_PrintJSONFails",
			outBytes:  testutils.Read(t, "NewDeleteCommand_Out_PrintJSONFails"),
			errBytes:  testutils.Read(t, "NewDeleteCommand_Err_PrintJSONFails"),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"DELETE:/v2/bucket/my-bucket": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body: `{
				"acknowledged: true
			}`,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodDelete, r.Method)
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

			deleteCmd := NewDeleteCommand(d)

			deleteCmd.SetIn(ios.In)
			deleteCmd.SetOut(ios.Out)
			deleteCmd.SetErr(ios.Err)

			deleteCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			deleteCmd.SetArgs(tc.args)

			err := deleteCmd.Execute()
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

// TestNewDeleteCommand_NoProfileFlag tests the NewDeleteCommand when the profile flag was not defined.
func TestNewDeleteCommand_NoProfileFlag(t *testing.T) {
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

	deleteCmd := NewDeleteCommand(d)

	deleteCmd.SetIn(ios.In)
	deleteCmd.SetOut(ios.Out)
	deleteCmd.SetErr(ios.Err)

	deleteCmd.SetArgs([]string{"my-bucket"})

	err := deleteCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewDeleteCommand_Out_NoProfile", testutils.Read(t, "NewDeleteCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewDeleteCommand_Err_NoProfile", testutils.Read(t, "NewDeleteCommand_Err_NoProfile"), errBuf.Bytes())
}

// TestNewDeleteCommand_NotExactly1Arg tests the NewDeleteCommand function when it does not receive exactly one argument.
func TestNewDeleteCommand_NotExactly1Arg(t *testing.T) {
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

	deleteCmd := NewDeleteCommand(d)

	deleteCmd.SetIn(ios.In)
	deleteCmd.SetOut(ios.Out)
	deleteCmd.SetErr(ios.Err)

	deleteCmd.SetArgs([]string{})

	err := deleteCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")

	testutils.CompareBytes(t, "NewDeleteCommand_Out_NotExactly1Arg", testutils.Read(t, "NewDeleteCommand_Out_NotExactly1Arg"), out.Bytes())
	testutils.CompareBytes(t, "NewDeleteCommand_Err_NotExactly1Arg", testutils.Read(t, "NewDeleteCommand_Err_NotExactly1Arg"), errBuf.Bytes())
}

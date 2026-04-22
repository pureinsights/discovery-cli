package file

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

// TestNewDeleteCommand tests the NewDeleteCommand() function.
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
			name:      "Delete by key returns an acknowledged true",
			args:      []string{"my-file"},
			url:       true,
			apiKey:    "",
			outGolden: "NewDeleteCommand_Out_DeleteByKeyReturnsTrue",
			errGolden: "NewDeleteCommand_Err_DeleteByKeyReturnsTrue",
			outBytes:  testutils.Read(t, "NewDeleteCommand_Out_DeleteByKeyReturnsTrue"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"DELETE:/v2/file/my-file": {
					StatusCode: http.StatusOK,
					Body: `{
				"acknowledged": true
			}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodDelete, r.Method)
						assert.Equal(t, "/v2/file/my-file", r.URL.Path)
					},
				},
			},
			err: nil,
		},
		{
			name:      "Delete by key returns an acknowledged false",
			args:      []string{"my-file"},
			url:       true,
			apiKey:    "",
			outGolden: "NewDeleteCommand_Out_DeleteByKeyReturnsFalse",
			errGolden: "NewDeleteCommand_Err_DeleteByKeyReturnsFalse",
			outBytes:  testutils.Read(t, "NewDeleteCommand_Out_DeleteByKeyReturnsFalse"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"DELETE:/v2/file/my-file": {
					StatusCode: http.StatusOK,
					Body: `{
				"acknowledged": false
			}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodDelete, r.Method)
						assert.Equal(t, "/v2/file/my-file", r.URL.Path)
					},
				},
			},
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			args:      []string{"my-file"},
			outGolden: "NewDeleteCommand_Out_NoURL",
			errGolden: "NewDeleteCommand_Err_NoURL",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewDeleteCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:      "delete returns an error",
			args:      []string{"my-file"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewDeleteCommand_Out_DeleteError",
			errGolden: "NewDeleteCommand_Err_DeleteError",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewDeleteCommand_Err_DeleteError"),
			responses: map[string]testutils.MockResponse{
				"DELETE:/v2/file/my-file": {
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
						assert.Equal(t, http.MethodDelete, r.Method)
						assert.Equal(t, "/v2/file/my-file", r.URL.Path)
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
}`)}, "Could not delete file with key \"my-file\""),
		},
		{
			name:      "Printing JSON object fails",
			args:      []string{"my-file"},
			outGolden: "NewDeleteCommand_Out_PrintJSONFails",
			errGolden: "NewDeleteCommand_Err_PrintJSONFails",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewDeleteCommand_Err_PrintJSONFails"),
			url:       true,
			apiKey:    "apiKey123",
			responses: map[string]testutils.MockResponse{
				"DELETE:/v2/file/my-file": {
					StatusCode: http.StatusOK,
					Body: `{
				"acknowledged: true
			}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodDelete, r.Method)
						assert.Equal(t, "/v2/file/my-file", r.URL.Path)
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
				vpr.Set("default.core_url", srv.URL)
			}
			if tc.apiKey != "" {
				vpr.Set("default.core_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, t.TempDir())

			deleteCmd := NewDeleteCommand(d)

			deleteCmd.SilenceUsage = true
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
			*testutils.Update = true
			err := deleteCmd.Execute()
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

	vpr.Set("default.core_url", "test")
	vpr.Set("default.core_key", "test")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	deleteCmd := NewDeleteCommand(d)

	deleteCmd.SetIn(ios.In)
	deleteCmd.SetOut(ios.Out)
	deleteCmd.SetErr(ios.Err)

	deleteCmd.SetArgs([]string{"my-file"})

	err := deleteCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewDeleteCommand_Out_NoProfile", testutils.Read(t, "NewDeleteCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewDeleteCommand_Err_NoProfile", testutils.Read(t, "NewDeleteCommand_Err_NoProfile"), errBuf.Bytes())
}

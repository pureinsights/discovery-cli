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

// NewDeleteCommand tests the NewDeleteCommand function
func TestNewDeleteCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		url        bool
		apiKey     string
		outGolden  string
		errGolden  string
		outBytes   []byte
		errBytes   []byte
		method     string
		path       string
		statusCode int
		response   string
		err        error
	}{
		// Working case
		{
			name:       "Delete by ID returns an acknowledged true",
			args:       []string{"3d51beef-8b90-40aa-84b5-033241dc6239"},
			url:        true,
			apiKey:     "apiKey123",
			outGolden:  "NewDeleteCommand_Out_DeleteByIdReturnsObject",
			errGolden:  "NewDeleteCommand_Err_DeleteByIdReturnsObject",
			outBytes:   testutils.Read(t, "NewDeleteCommand_Out_DeleteByIdReturnsObject"),
			errBytes:   []byte(nil),
			method:     http.MethodDelete,
			statusCode: http.StatusOK,
			response: `{
				"acknowledged": true
			}`,
			err: nil,
		},
		{
			name:       "Delete by name returns an acknowledged true",
			args:       []string{""},
			url:        true,
			apiKey:     "apiKey123",
			outGolden:  "NewDeleteCommand_Out_DeleteByIdReturnsObject",
			errGolden:  "NewDeleteCommand_Err_DeleteByIdReturnsObject",
			outBytes:   testutils.Read(t, "NewDeleteCommand_Out_DeleteByIdReturnsObject"),
			errBytes:   []byte(nil),
			method:     http.MethodDelete,
			statusCode: http.StatusOK,
			response: `{
				"acknowledged": true
			}`,
			err: nil,
		},

		// Error case
		{
			name:       "No URL",
			args:       []string{"3d51beef-8b90-40aa-84b5-033241dc6239"},
			outGolden:  "NewDeleteCommand_Out_NoURL",
			errGolden:  "NewDeleteCommand_Err_NoURL",
			outBytes:   testutils.Read(t, "NewDeleteCommand_Out_NoURL"),
			errBytes:   testutils.Read(t, "NewDeleteCommand_Err_NoURL"),
			url:        false,
			apiKey:     "apiKey123",
			method:     http.MethodDelete,
			statusCode: http.StatusOK,
			response:   ``,
			err:        cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile {profile}\n      discovery core config --profile {profile}"),
		},
		{
			name:       "No API key",
			args:       []string{"3d51beef-8b90-40aa-84b5-033241dc6239"},
			outGolden:  "NewDeleteCommand_Out_NoAPIKey",
			errGolden:  "NewDeleteCommand_Err_NoAPIKey",
			outBytes:   testutils.Read(t, "NewDeleteCommand_Out_NoAPIKey"),
			errBytes:   testutils.Read(t, "NewDeleteCommand_Err_NoAPIKey"),
			url:        true,
			apiKey:     "",
			method:     http.MethodDelete,
			statusCode: http.StatusNotFound,
			response:   ``,
			err:        cli.NewError(cli.ErrorExitCode, "The Discovery Core API key is missing for profile \"default\".\nTo set the API key for the Discovery Core API, run any of the following commands:\n      discovery config  --profile {profile}\n      discovery core config --profile {profile}"),
		},
		// {
		// 	name:       "user does not send a UUID",
		// 	args:       []string{"test"},
		// 	url:        true,
		// 	apiKey:     "apiKey123",
		// 	outGolden:  "NewDeleteCommand_Out_NotUUID",
		// 	errGolden:  "NewDeleteCommand_Err_NotUUID",
		// 	method:     http.MethodDelete,
		// 	statusCode: http.StatusOK,
		// 	response:   ``,
		// 	err:        cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid UUID length: 4"), "Could not convert given id \"test\" to UUID. This command does not support referencing an entity by name."),
		// },
		{
			name:       "Printing JSON object fails",
			args:       []string{"3d51beef-8b90-40aa-84b5-033241dc6239"},
			outGolden:  "NewDeleteCommand_Out_PrintJSONFails",
			errGolden:  "NewDeleteCommand_Err_PrintJSONFails",
			outBytes:   testutils.Read(t, "NewDeleteCommand_Out_PrintJSONFails"),
			errBytes:   testutils.Read(t, "NewDeleteCommand_Err_PrintJSONFails"),
			url:        true,
			apiKey:     "apiKey123",
			method:     http.MethodDelete,
			statusCode: http.StatusOK,
			response:   `{"messages": {{}`,
			err:        cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character '{' looking for beginning of object key string"), "Could not print JSON object"),
		},
		{
			name:       "DeleteEntity returns HTTP error",
			args:       []string{"3d51beef-8b90-40aa-84b5-033241dc6239"},
			outGolden:  "NewDeleteCommand_Out_DeleteEntityHTTPError",
			errGolden:  "NewDeleteCommand_Err_DeleteEntityHTTPError",
			url:        true,
			apiKey:     "apiKey123",
			method:     http.MethodDelete,
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [id] for value [test] due to: Invalid UUID string: test"
			],
			"timestamp": "2025-10-23T22:35:38.345647200Z"
			}`,
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [id] for value [test] due to: Invalid UUID string: test"
			],
			"timestamp": "2025-10-23T22:35:38.345647200Z"
			}`)}, "Could not delete entity with id \"3d51beef-8b90-40aa-84b5-033241dc6239\""),
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
	vpr.Set("output", "json")

	vpr.Set("default.core_url", "test")
	vpr.Set("default.core_key", "test")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	deleteCmd := NewDeleteCommand(d)

	deleteCmd.SetIn(ios.In)
	deleteCmd.SetOut(ios.Out)
	deleteCmd.SetErr(ios.Err)

	deleteCmd.SetArgs([]string{"3d51beef-8b90-40aa-84b5-033241dc6239"})

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
	vpr.Set("output", "json")

	vpr.Set("default.core_url", "test")
	vpr.Set("default.core_key", "test")

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

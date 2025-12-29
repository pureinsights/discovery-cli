package statuscheck

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

// TestNewStatusCommand tests the NewStatusCommand() function.
func TestNewStatusCommand(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		httpStatus int
		url        bool
		apiKey     string
		outGolden  string
		errGolden  string
		outBytes   []byte
		errBytes   []byte
		err        error
	}{
		// Working case
		{
			name: "Status returns an UP result",
			response: `{
    "status": "UP"
}`,
			url:        true,
			apiKey:     "",
			httpStatus: http.StatusOK,
			outGolden:  "NewStatusCommand_Out_UpStatus",
			errGolden:  "NewStatusCommand_Err_UpStatus",
			outBytes:   testutils.Read(t, "NewStatusCommand_Out_UpStatus"),
			errBytes:   []byte(nil),
			err:        nil,
		},
		{
			name: "Status returns a down result",
			response: `{
    "status": "DOWN"
}`,
			url:        true,
			apiKey:     "",
			outGolden:  "NewStatusCommand_Out_DownStatus",
			errGolden:  "NewStatusCommand_Err_DownStatus",
			httpStatus: http.StatusOK,
			outBytes:   testutils.Read(t, "NewStatusCommand_Out_DownStatus"),
			errBytes:   []byte(nil),
			err:        nil,
		},

		// Error case
		{
			name:       "No URL",
			outGolden:  "NewStatusCommand_Out_NoURL",
			errGolden:  "NewStatusCommand_Err_NoURL",
			outBytes:   testutils.Read(t, "NewStatusCommand_Out_NoURL"),
			errBytes:   testutils.Read(t, "NewStatusCommand_Err_NoURL"),
			url:        false,
			apiKey:     "apiKey123",
			httpStatus: http.StatusOK,
			err:        cli.NewError(cli.ErrorExitCode, "The Discovery Ingestion URL is missing for profile \"default\".\nTo set the URL for the Discovery Ingestion API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery ingestion config --profile \"default\""),
		},
		{
			name:       "Status returns error",
			outGolden:  "NewStatusCommand_Out_StatusError",
			errGolden:  "NewStatusCommand_Err_StatusError",
			outBytes:   testutils.Read(t, "NewStatusCommand_Out_StatusError"),
			errBytes:   testutils.Read(t, "NewStatusCommand_Err_StatusError"),
			httpStatus: http.StatusInternalServerError,
			response:   "{\"error\": \"Internal server error\"}",
			url:        true,
			apiKey:     "",
			err:        cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusInternalServerError, Body: gjson.Parse("{\"error\": \"Internal server error\"}")}, "Could not check the status of Discovery Ingestion."),
		},
		{
			name:       "Printing JSON object fails",
			outGolden:  "NewStatusCommand_Out_PrintJSONFails",
			errGolden:  "NewStatusCommand_Err_PrintJSONFails",
			outBytes:   testutils.Read(t, "NewStatusCommand_Out_PrintJSONFails"),
			errBytes:   testutils.Read(t, "NewStatusCommand_Err_PrintJSONFails"),
			url:        true,
			apiKey:     "apiKey123",
			httpStatus: http.StatusOK,
			response: `{
    "status": "UP
}`,
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character '\\n' in string literal"), "Could not print JSON object"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.httpStatus, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/health", r.URL.Path)
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

			if tc.url {
				vpr.Set("default.ingestion_url", srv.URL)
			}
			if tc.apiKey != "" {
				vpr.Set("default.ingestion_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, t.TempDir())

			statusCmd := NewStatusCommand(d)

			statusCmd.SetIn(ios.In)
			statusCmd.SetOut(ios.Out)
			statusCmd.SetErr(ios.Err)

			statusCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			statusCmd.SetArgs([]string{})

			err := statusCmd.Execute()
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

// TestNewStatusCommand_NoProfileFlag tests the NewStatusCommand when the profile flag was not defined.
func TestNewStatusCommand_NoProfileFlag(t *testing.T) {
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

	statusCmd := NewStatusCommand(d)

	statusCmd.SetIn(ios.In)
	statusCmd.SetOut(ios.Out)
	statusCmd.SetErr(ios.Err)

	statusCmd.SetArgs([]string{})
	err := statusCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewStatusCommand_Out_NoProfile", testutils.Read(t, "NewStatusCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewStatusCommand_Err_NoProfile", testutils.Read(t, "NewStatusCommand_Err_NoProfile"), errBuf.Bytes())
}

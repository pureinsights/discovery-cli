package tokens

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/pureinsights/discovery-cli/v2/internal/iostreams"
	"github.com/pureinsights/discovery-cli/v2/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestNewEnableCommand tests the NewEnableCommand() function.
func TestNewEnableCommand(t *testing.T) {
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
			name:      "Enable by ID acknowledges the token",
			args:      []string{tokenId, "--service-account", "ci"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewEnableCommand_Out_EnableByIdReturnsObject",
			errGolden: "NewEnableCommand_Err_EnableByIdReturnsObject",
			outBytes:  testutils.Read(t, "NewEnableCommand_Out_EnableByIdReturnsObject"),
			errBytes:  []byte(nil),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByIdPhase(),
				map[string]testutils.MockResponse{
					"PATCH:" + tokenPath + "/" + tokenId + "/enable": mock(http.StatusOK, `{"acknowledged": true}`, http.MethodPatch, tokenPath+"/"+tokenId+"/enable"),
				},
			),
			err: nil,
		},
		{
			name:      "Enable by name acknowledges the token",
			args:      []string{"My Token", "--service-account", "ci"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewEnableCommand_Out_EnableByNameReturnsObject",
			errGolden: "NewEnableCommand_Err_EnableByNameReturnsObject",
			outBytes:  testutils.Read(t, "NewEnableCommand_Out_EnableByNameReturnsObject"),
			errBytes:  []byte(nil),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByNamePhase(),
				map[string]testutils.MockResponse{
					"PATCH:" + tokenPath + "/" + tokenId + "/enable": mock(http.StatusOK, `{"acknowledged": true}`, http.MethodPatch, tokenPath+"/"+tokenId+"/enable"),
				},
			),
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			args:      []string{"My Token", "--service-account", "ci"},
			outGolden: "NewEnableCommand_Out_NoURL",
			errGolden: "NewEnableCommand_Err_NoURL",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewEnableCommand_Err_NoURL"),
			url:       false,
			apiKey:    apiKey,
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:      "sent service account does not exist",
			args:      []string{"My Token", "--service-account", "test"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewEnableCommand_Out_ServiceAccountDoesNotExist",
			errGolden: "NewEnableCommand_Err_ServiceAccountDoesNotExist",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewEnableCommand_Err_ServiceAccountDoesNotExist"),
			responses: serviceAccountNotFoundPhase(),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body:   gjson.Parse(fmt.Sprintf(discoveryPackage.NotFoundError, "test")),
			}, "Could not find service account \"test\""),
		},
		{
			name:      "sent token name does not exist",
			args:      []string{"test", "--service-account", "ci"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewEnableCommand_Out_NameDoesNotExist",
			errGolden: "NewEnableCommand_Err_NameDoesNotExist",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewEnableCommand_Err_NameDoesNotExist"),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				tokenNotFoundPhase(),
			),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body:   gjson.Parse(fmt.Sprintf(discoveryPackage.NotFoundError, "test")),
			}, "Could not get entity ID."),
		},
		{
			name:      "Enable returns HTTP error",
			args:      []string{tokenId, "--service-account", "ci"},
			outGolden: "NewEnableCommand_Out_EnableHTTPError",
			errGolden: "NewEnableCommand_Err_EnableHTTPError",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewEnableCommand_Err_EnableHTTPError"),
			url:       true,
			apiKey:    apiKey,
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByIdPhase(),
				map[string]testutils.MockResponse{
					"PATCH:" + tokenPath + "/" + tokenId + "/enable": mock(http.StatusInternalServerError, `{
	"status": 500,
	"code": 1003,
	"messages": [
		"Internal server error"
	],
	"timestamp": "2026-11-03T17:02:11.000000000Z"
}`, http.MethodPatch, tokenPath+"/"+tokenId+"/enable"),
				},
			),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusInternalServerError, Body: gjson.Parse(`{
	"status": 500,
	"code": 1003,
	"messages": [
		"Internal server error"
	],
	"timestamp": "2026-11-03T17:02:11.000000000Z"
}`)}, "Could not enable entity with id \""+tokenId+"\""),
		},
		{
			name:      "Printing JSON object fails",
			args:      []string{tokenId, "--service-account", "ci"},
			outGolden: "NewEnableCommand_Out_PrintJSONFails",
			errGolden: "NewEnableCommand_Err_PrintJSONFails",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewEnableCommand_Err_PrintJSONFails"),
			url:       true,
			apiKey:    apiKey,
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByIdPhase(),
				map[string]testutils.MockResponse{
					"PATCH:" + tokenPath + "/" + tokenId + "/enable": mock(http.StatusOK, `{
				"acknowledged: true
			}`, http.MethodPatch, tokenPath+"/"+tokenId+"/enable"),
				},
			),
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

			enableCmd := NewEnableCommand(d)

			enableCmd.SilenceUsage = true
			enableCmd.SetIn(ios.In)
			enableCmd.SetOut(ios.Out)
			enableCmd.SetErr(ios.Err)

			enableCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			enableCmd.SetArgs(tc.args)

			err := enableCmd.Execute()
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

// TestNewEnableCommand_NoServiceAccountFlag tests the NewEnableCommand when the required service account flag is not sent.
func TestNewEnableCommand_NoServiceAccountFlag(t *testing.T) {
	ios := iostreams.IOStreams{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	d := cli.NewDiscovery(&ios, viper.New(), t.TempDir())

	enableCmd := NewEnableCommand(d)
	enableCmd.SetIn(ios.In)
	enableCmd.SetOut(ios.Out)
	enableCmd.SetErr(ios.Err)
	enableCmd.SetArgs([]string{"My Token"})

	err := enableCmd.Execute()
	assert.EqualError(t, err, `required flag(s) "service-account" not set`)
}

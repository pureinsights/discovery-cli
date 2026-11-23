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

// TestNewRevokeCommand tests the NewRevokeCommand() function.
func TestNewRevokeCommand(t *testing.T) {
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
			name:      "Revoke by ID returns an acknowledged true",
			args:      []string{tokenId, "--service-account", "ci"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewRevokeCommand_Out_RevokeByIdReturnsObject",
			errGolden: "NewRevokeCommand_Err_RevokeByIdReturnsObject",
			outBytes:  testutils.Read(t, "NewRevokeCommand_Out_RevokeByIdReturnsObject"),
			errBytes:  []byte(nil),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByIdPhase(),
				map[string]testutils.MockResponse{
					"DELETE:" + tokenPath + "/" + tokenId: mock(http.StatusOK, `{"acknowledged": true}`, http.MethodDelete, tokenPath+"/"+tokenId),
				},
			),
			err: nil,
		},
		{
			name:      "Revoke by name returns an acknowledged true",
			args:      []string{"My Token", "--service-account", "ci"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewRevokeCommand_Out_RevokeByNameReturnsObject",
			errGolden: "NewRevokeCommand_Err_RevokeByNameReturnsObject",
			outBytes:  testutils.Read(t, "NewRevokeCommand_Out_RevokeByNameReturnsObject"),
			errBytes:  []byte(nil),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByNamePhase(),
				map[string]testutils.MockResponse{
					"DELETE:" + tokenPath + "/" + tokenId: mock(http.StatusOK, `{"acknowledged": true}`, http.MethodDelete, tokenPath+"/"+tokenId),
				},
			),
			err: nil,
		},
		{
			name:      "Revoke a token that no longer exists returns an acknowledged false",
			args:      []string{tokenId, "--service-account", "ci"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewRevokeCommand_Out_RevokeNotFoundReturnsAcknowledgedFalse",
			errGolden: "NewRevokeCommand_Err_RevokeNotFoundReturnsAcknowledgedFalse",
			outBytes:  testutils.Read(t, "NewRevokeCommand_Out_RevokeNotFoundReturnsAcknowledgedFalse"),
			errBytes:  []byte(nil),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByIdPhase(),
				map[string]testutils.MockResponse{
					"DELETE:" + tokenPath + "/" + tokenId: mock(http.StatusOK, `{"acknowledged": false}`, http.MethodDelete, tokenPath+"/"+tokenId),
				},
			),
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			args:      []string{"My Token", "--service-account", "ci"},
			outGolden: "NewRevokeCommand_Out_NoURL",
			errGolden: "NewRevokeCommand_Err_NoURL",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewRevokeCommand_Err_NoURL"),
			url:       false,
			apiKey:    apiKey,
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:      "sent service account does not exist",
			args:      []string{"My Token", "--service-account", "test"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewRevokeCommand_Out_ServiceAccountDoesNotExist",
			errGolden: "NewRevokeCommand_Err_ServiceAccountDoesNotExist",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewRevokeCommand_Err_ServiceAccountDoesNotExist"),
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
			outGolden: "NewRevokeCommand_Out_NameDoesNotExist",
			errGolden: "NewRevokeCommand_Err_NameDoesNotExist",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewRevokeCommand_Err_NameDoesNotExist"),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				tokenNotFoundPhase(),
			),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body:   gjson.Parse(fmt.Sprintf(discoveryPackage.NotFoundError, "test")),
			}, "Could not search for entity with name \"test\""),
		},
		{
			name:      "Printing JSON object fails",
			args:      []string{"My Token", "--service-account", "ci"},
			outGolden: "NewRevokeCommand_Out_PrintJSONFails",
			errGolden: "NewRevokeCommand_Err_PrintJSONFails",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewRevokeCommand_Err_PrintJSONFails"),
			url:       true,
			apiKey:    apiKey,
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByNamePhase(),
				map[string]testutils.MockResponse{
					"DELETE:" + tokenPath + "/" + tokenId: mock(http.StatusOK, `{
				"acknowledged: true
			}`, http.MethodDelete, tokenPath+"/"+tokenId),
				},
			),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character '\\n' in string literal"), "Could not print JSON object"),
		},
		{
			name:      "Search returns invalid UUID error",
			args:      []string{"My Token", "--service-account", "ci"},
			outGolden: "NewRevokeCommand_Out_InvalidUUID",
			errGolden: "NewRevokeCommand_Err_InvalidUUID",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewRevokeCommand_Err_InvalidUUID"),
			url:       true,
			apiKey:    apiKey,
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByNamePhase(),
				map[string]testutils.MockResponse{
					"GET:" + tokenPath + "/" + tokenId: mock(http.StatusOK, strings.Replace(tokenBody, tokenId, "test", 1), http.MethodGet, tokenPath+"/"+tokenId),
				},
			),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid UUID length: 4"), "Could not delete entity with name \"My Token\""),
		},
		{
			name:      "Revoke returns HTTP error",
			args:      []string{tokenId, "--service-account", "ci"},
			outGolden: "NewRevokeCommand_Out_RevokeHTTPError",
			errGolden: "NewRevokeCommand_Err_RevokeHTTPError",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewRevokeCommand_Err_RevokeHTTPError"),
			url:       true,
			apiKey:    apiKey,
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByIdPhase(),
				map[string]testutils.MockResponse{
					"DELETE:" + tokenPath + "/" + tokenId: mock(http.StatusInternalServerError, `{
	"status": 500,
	"code": 1003,
	"messages": [
		"Internal server error"
	],
	"timestamp": "2026-11-03T16:50:12.000000000Z"
}`, http.MethodDelete, tokenPath+"/"+tokenId),
				},
			),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusInternalServerError, Body: gjson.Parse(`{
	"status": 500,
	"code": 1003,
	"messages": [
		"Internal server error"
	],
	"timestamp": "2026-11-03T16:50:12.000000000Z"
}`)}, "Could not delete entity with id \""+tokenId+"\""),
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

			revokeCmd := NewRevokeCommand(d)

			revokeCmd.SilenceUsage = true
			revokeCmd.SetIn(ios.In)
			revokeCmd.SetOut(ios.Out)
			revokeCmd.SetErr(ios.Err)

			revokeCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			revokeCmd.SetArgs(tc.args)

			err := revokeCmd.Execute()
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

// TestNewRevokeCommand_NoServiceAccountFlag tests the NewRevokeCommand when the required service account flag is not sent.
func TestNewRevokeCommand_NoServiceAccountFlag(t *testing.T) {
	ios := iostreams.IOStreams{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	d := cli.NewDiscovery(&ios, viper.New(), t.TempDir())

	revokeCmd := NewRevokeCommand(d)
	revokeCmd.SetIn(ios.In)
	revokeCmd.SetOut(ios.Out)
	revokeCmd.SetErr(ios.Err)
	revokeCmd.SetArgs([]string{"My Token"})

	err := revokeCmd.Execute()
	assert.EqualError(t, err, `required flag(s) "service-account" not set`)
}

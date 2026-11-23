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

// TestNewRenewCommand tests the NewRenewCommand() function.
func TestNewRenewCommand(t *testing.T) {
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
			name:      "Renew by ID returns the token with its new value",
			args:      []string{tokenId, "--service-account", "ci", "--ttl", "60d"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewRenewCommand_Out_RenewByIdReturnsObject",
			errGolden: "NewRenewCommand_Err_RenewByIdReturnsObject",
			outBytes:  testutils.Read(t, "NewRenewCommand_Out_RenewByIdReturnsObject"),
			errBytes:  []byte(nil),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByIdPhase(),
				map[string]testutils.MockResponse{
					"PATCH:" + tokenPath + "/" + tokenId + "/renew": {
						StatusCode:  http.StatusOK,
						Body:        renewedTokenBody,
						ContentType: "application/json",
						Assertions: func(t *testing.T, r *http.Request) {
							assert.Equal(t, http.MethodPatch, r.Method)
							assert.Equal(t, tokenPath+"/"+tokenId+"/renew", r.URL.Path)
							assert.Equal(t, "60d", r.URL.Query().Get("ttl"))
						},
					},
				},
			),
			err: nil,
		},
		{
			name:      "Renew by name with a non-expiring TTL returns the token with its new value",
			args:      []string{"My Token", "--service-account", "ci", "--ttl", "-1"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewRenewCommand_Out_RenewByNameReturnsObject",
			errGolden: "NewRenewCommand_Err_RenewByNameReturnsObject",
			outBytes:  testutils.Read(t, "NewRenewCommand_Out_RenewByNameReturnsObject"),
			errBytes:  []byte(nil),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByNamePhase(),
				map[string]testutils.MockResponse{
					"PATCH:" + tokenPath + "/" + tokenId + "/renew": {
						StatusCode:  http.StatusOK,
						Body:        strings.Replace(renewedTokenBody, `"expirationTimestamp": "2027-01-18T10:00:00Z",`, ``, 1),
						ContentType: "application/json",
						Assertions: func(t *testing.T, r *http.Request) {
							assert.Equal(t, http.MethodPatch, r.Method)
							assert.Equal(t, tokenPath+"/"+tokenId+"/renew", r.URL.Path)
							assert.Equal(t, "-1", r.URL.Query().Get("ttl"))
							assert.Equal(t, apiKey, r.Header.Get("X-API-Key"))
						},
					},
				},
			),
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			args:      []string{"My Token", "--service-account", "ci", "--ttl", "60d"},
			outGolden: "NewRenewCommand_Out_NoURL",
			errGolden: "NewRenewCommand_Err_NoURL",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewRenewCommand_Err_NoURL"),
			url:       false,
			apiKey:    apiKey,
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:      "sent service account does not exist",
			args:      []string{"My Token", "--service-account", "test", "--ttl", "60d"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewRenewCommand_Out_ServiceAccountDoesNotExist",
			errGolden: "NewRenewCommand_Err_ServiceAccountDoesNotExist",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewRenewCommand_Err_ServiceAccountDoesNotExist"),
			responses: serviceAccountNotFoundPhase(),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body:   gjson.Parse(fmt.Sprintf(discoveryPackage.NotFoundError, "test")),
			}, "Could not find service account \"test\""),
		},
		{
			name:      "sent token name does not exist",
			args:      []string{"test", "--service-account", "ci", "--ttl", "60d"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewRenewCommand_Out_NameDoesNotExist",
			errGolden: "NewRenewCommand_Err_NameDoesNotExist",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewRenewCommand_Err_NameDoesNotExist"),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				tokenNotFoundPhase(),
			),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body:   gjson.Parse(fmt.Sprintf(discoveryPackage.NotFoundError, "test")),
			}, "Could not get the token ID."),
		},
		{
			name:      "Renew with an invalid TTL returns HTTP error",
			args:      []string{tokenId, "--service-account", "ci", "--ttl", "0"},
			outGolden: "NewRenewCommand_Out_InvalidTTL",
			errGolden: "NewRenewCommand_Err_InvalidTTL",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewRenewCommand_Err_InvalidTTL"),
			url:       true,
			apiKey:    apiKey,
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByIdPhase(),
				map[string]testutils.MockResponse{
					"PATCH:" + tokenPath + "/" + tokenId + "/renew": {
						StatusCode:  http.StatusBadRequest,
						ContentType: "application/json",
						Body: `{
	"status": 400,
	"code": 3002,
	"messages": [
		"ttl: must match \"-1|[1-9]\\d{0,8}[smhd]?\""
	],
	"timestamp": "2026-11-19T10:00:00.000000000Z"
}`,
						Assertions: func(t *testing.T, r *http.Request) {
							assert.Equal(t, http.MethodPatch, r.Method)
							assert.Equal(t, "0", r.URL.Query().Get("ttl"))
							assert.Equal(t, apiKey, r.Header.Get("X-API-Key"))
						},
					},
				},
			),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
	"status": 400,
	"code": 3002,
	"messages": [
		"ttl: must match \"-1|[1-9]\\d{0,8}[smhd]?\""
	],
	"timestamp": "2026-11-19T10:00:00.000000000Z"
}`)}, "Could not renew token with id \""+tokenId+"\""),
		},
		{
			name:      "Printing JSON object fails",
			args:      []string{tokenId, "--service-account", "ci", "--ttl", "60d"},
			outGolden: "NewRenewCommand_Out_PrintJSONFails",
			errGolden: "NewRenewCommand_Err_PrintJSONFails",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewRenewCommand_Err_PrintJSONFails"),
			url:       true,
			apiKey:    apiKey,
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByIdPhase(),
				map[string]testutils.MockResponse{
					"PATCH:" + tokenPath + "/" + tokenId + "/renew": mock(http.StatusOK, strings.Replace(renewedTokenBody, `"active": true`, `"active: true`, 1), http.MethodPatch, tokenPath+"/"+tokenId+"/renew"),
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

			renewCmd := NewRenewCommand(d)

			renewCmd.SilenceUsage = true
			renewCmd.SetIn(ios.In)
			renewCmd.SetOut(ios.Out)
			renewCmd.SetErr(ios.Err)

			renewCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			renewCmd.SetArgs(tc.args)

			err := renewCmd.Execute()
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

// TestNewRenewCommand_RequiredFlags tests the NewRenewCommand when the required flags are not sent.
func TestNewRenewCommand_RequiredFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		err  string
	}{
		{name: "no flags", args: []string{"My Token"}, err: `required flag(s) "service-account", "ttl" not set`},
		{name: "no TTL", args: []string{"My Token", "--service-account", "ci"}, err: `required flag(s) "ttl" not set`},
		{name: "no service account", args: []string{"My Token", "--ttl", "60d"}, err: `required flag(s) "service-account" not set`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ios := iostreams.IOStreams{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
			d := cli.NewDiscovery(&ios, viper.New(), t.TempDir())

			renewCmd := NewRenewCommand(d)
			renewCmd.SetIn(ios.In)
			renewCmd.SetOut(ios.Out)
			renewCmd.SetErr(ios.Err)
			renewCmd.SetArgs(tc.args)

			assert.EqualError(t, renewCmd.Execute(), tc.err)
		})
	}
}

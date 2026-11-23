package tokens

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
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

// TestNewUpdateCommand tests the NewUpdateCommand() function.
func TestNewUpdateCommand(t *testing.T) {
	_, parseBoolErr := strconv.ParseBool("maybe")

	tests := []struct {
		name      string
		args      []string
		input     string
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
			name: "Add and remove scopes by name",
			args: []string{
				"My Token",
				"--service-account", "ci",
				"--add-scopes", "staging:bucket:create,staging:content:*",
				"--remove-scopes", "core:*"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewUpdateCommand_Out_AddRemoveScopes",
			errGolden: "NewUpdateCommand_Err_AddRemoveScopes",
			outBytes:  testutils.Read(t, "NewUpdateCommand_Out_AddRemoveScopes"),
			errBytes:  []byte(nil),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByNamePhase(),
				map[string]testutils.MockResponse{
					"PUT:" + tokenPath + "/" + tokenId: {
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
						Body:        strings.Replace(tokenBody, `"scopes": ["core:*"]`, `"scopes": ["staging:bucket:create", "staging:content:*"]`, 1),
						Assertions: func(t *testing.T, r *http.Request) {
							assert.Equal(t, http.MethodPut, r.Method)
							assert.Equal(t, tokenPath+"/"+tokenId, r.URL.Path)
							assert.Equal(t, apiKey, r.Header.Get("X-API-Key"))
							body, err := io.ReadAll(r.Body)
							require.NoError(t, err)
							assert.JSONEq(t, strings.Replace(tokenBody, `"scopes": ["core:*"]`, `"scopes": ["staging:bucket:create", "staging:content:*"]`, 1), string(body))
						},
					},
				},
			),
			err: nil,
		},
		{
			name:      "Replace the scopes by ID",
			args:      []string{tokenId, "--service-account", "ci", "--scopes", "queryflow:*"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewUpdateCommand_Out_ReplaceScopes",
			errGolden: "NewUpdateCommand_Err_ReplaceScopes",
			outBytes:  testutils.Read(t, "NewUpdateCommand_Out_ReplaceScopes"),
			errBytes:  []byte(nil),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByIdPhase(),
				map[string]testutils.MockResponse{
					"PUT:" + tokenPath + "/" + tokenId: {
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
						Body:        strings.Replace(tokenBody, `"scopes": ["core:*"]`, `"scopes": ["queryflow:*"]`, 1),
						Assertions: func(t *testing.T, r *http.Request) {
							assert.Equal(t, http.MethodPut, r.Method)
							assert.Equal(t, tokenPath+"/"+tokenId, r.URL.Path)
							assert.Equal(t, apiKey, r.Header.Get("X-API-Key"))
							body, err := io.ReadAll(r.Body)
							require.NoError(t, err)
							assert.JSONEq(t, strings.Replace(tokenBody, `"scopes": ["core:*"]`, `"scopes": ["queryflow:*"]`, 1), string(body))
						},
					},
				},
			),
			err: nil,
		},
		{
			name:      "Disable the token with the enabled flag",
			args:      []string{"My Token", "--service-account", "ci", "--enabled=false"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewUpdateCommand_Out_EnabledFlag",
			errGolden: "NewUpdateCommand_Err_EnabledFlag",
			outBytes:  testutils.Read(t, "NewUpdateCommand_Out_EnabledFlag"),
			errBytes:  []byte(nil),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByNamePhase(),
				map[string]testutils.MockResponse{
					"PUT:" + tokenPath + "/" + tokenId: {
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
						Body:        strings.Replace(tokenBody, `"active": true`, `"active": false`, 1),
						Assertions: func(t *testing.T, r *http.Request) {
							assert.Equal(t, http.MethodPut, r.Method)
							assert.Equal(t, tokenPath+"/"+tokenId, r.URL.Path)
							assert.Equal(t, apiKey, r.Header.Get("X-API-Key"))
							body, err := io.ReadAll(r.Body)
							require.NoError(t, err)
							assert.JSONEq(t, strings.Replace(tokenBody, `"active": true`, `"active": false`, 1), string(body))
						},
					},
				},
			),
			err: nil,
		},
		{
			name:      "Update interactively",
			args:      []string{"My Token", "--service-account", "ci"},
			input:     "Renamed Token\n\nfalse\n",
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewUpdateCommand_Out_Interactive",
			errGolden: "NewUpdateCommand_Err_Interactive",
			outBytes:  testutils.Read(t, "NewUpdateCommand_Out_Interactive"),
			errBytes:  []byte(nil),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByNamePhase(),
				map[string]testutils.MockResponse{
					"PUT:" + tokenPath + "/" + tokenId: {
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
						Body:        strings.Replace(strings.Replace(tokenBody, `"My Token"`, `"Renamed Token"`, 1), `"active": true`, `"active": false`, 1),
						Assertions: func(t *testing.T, r *http.Request) {
							assert.Equal(t, http.MethodPut, r.Method)
							assert.Equal(t, tokenPath+"/"+tokenId, r.URL.Path)
							assert.Equal(t, apiKey, r.Header.Get("X-API-Key"))
							body, err := io.ReadAll(r.Body)
							require.NoError(t, err)
							assert.JSONEq(t, strings.Replace(strings.Replace(tokenBody, `"My Token"`, `"Renamed Token"`, 1), `"active": true`, `"active": false`, 1), string(body))
						},
					},
				},
			),
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			args:      []string{"My Token", "--service-account", "ci", "--scopes", "core:*"},
			outGolden: "NewUpdateCommand_Out_NoURL",
			errGolden: "NewUpdateCommand_Err_NoURL",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewUpdateCommand_Err_NoURL"),
			url:       false,
			apiKey:    apiKey,
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:      "sent service account does not exist",
			args:      []string{"My Token", "--service-account", "test", "--scopes", "core:*"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewUpdateCommand_Out_ServiceAccountDoesNotExist",
			errGolden: "NewUpdateCommand_Err_ServiceAccountDoesNotExist",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewUpdateCommand_Err_ServiceAccountDoesNotExist"),
			responses: serviceAccountNotFoundPhase(),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body:   gjson.Parse(fmt.Sprintf(discoveryPackage.NotFoundError, "test")),
			}, "Could not find service account \"test\""),
		},
		{
			name:      "sent token name does not exist",
			args:      []string{"test", "--service-account", "ci", "--scopes", "core:*"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewUpdateCommand_Out_NameDoesNotExist",
			errGolden: "NewUpdateCommand_Err_NameDoesNotExist",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewUpdateCommand_Err_NameDoesNotExist"),
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
			name:      "Interactive update with an invalid active value",
			args:      []string{"My Token", "--service-account", "ci"},
			input:     "\n\nmaybe\n",
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewUpdateCommand_Out_InvalidActive",
			errGolden: "NewUpdateCommand_Err_InvalidActive",
			outBytes:  testutils.Read(t, "NewUpdateCommand_Out_InvalidActive"),
			errBytes:  testutils.Read(t, "NewUpdateCommand_Err_InvalidActive"),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByNamePhase(),
			),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, parseBoolErr, "Invalid value for active: \"maybe\""),
		},
		{
			name:      "Update returns HTTP error",
			args:      []string{"My Token", "--service-account", "ci", "--scopes", "Invalid Scope"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewUpdateCommand_Out_UpdateHTTPError",
			errGolden: "NewUpdateCommand_Err_UpdateHTTPError",
			outBytes:  testutils.Read(t, "NewUpdateCommand_Out_UpdateHTTPError"),
			errBytes:  testutils.Read(t, "NewUpdateCommand_Err_UpdateHTTPError"),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByNamePhase(),
				map[string]testutils.MockResponse{
					"PUT:" + tokenPath + "/" + tokenId: {
						StatusCode:  http.StatusBadRequest,
						ContentType: "application/json",
						Body: `{
	"status": 400,
	"code": 3002,
	"messages": [
		"scopes[]: Scopes must be product:entity:permission, where entity or permission may be '*'"
	],
	"timestamp": "2026-11-19T10:00:00.000000000Z"
}`,
						Assertions: func(t *testing.T, r *http.Request) {
							assert.Equal(t, http.MethodPut, r.Method)
							assert.Equal(t, tokenPath+"/"+tokenId, r.URL.Path)
							assert.Equal(t, apiKey, r.Header.Get("X-API-Key"))
							body, err := io.ReadAll(r.Body)
							require.NoError(t, err)
							assert.JSONEq(t, strings.Replace(tokenBody, `"scopes": ["core:*"]`, `"scopes": ["Invalid Scope"]`, 1), string(body))
						},
					},
				},
			),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
	"status": 400,
	"code": 3002,
	"messages": [
		"scopes[]: Scopes must be product:entity:permission, where entity or permission may be '*'"
	],
	"timestamp": "2026-11-19T10:00:00.000000000Z"
}`)}, "Could not store entities"),
		},
		{
			name:      "Printing JSON array fails",
			args:      []string{"My Token", "--service-account", "ci", "--scopes", "core:*"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewUpdateCommand_Out_PrintJSONFails",
			errGolden: "NewUpdateCommand_Err_PrintJSONFails",
			outBytes:  testutils.Read(t, "NewUpdateCommand_Out_PrintJSONFails"),
			errBytes:  testutils.Read(t, "NewUpdateCommand_Err_PrintJSONFails"),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByNamePhase(),
				map[string]testutils.MockResponse{
					"PUT:" + tokenPath + "/" + tokenId: {
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
						Body:        strings.Replace(tokenBody, `"active": true`, `"active: true`, 1),
						Assertions: func(t *testing.T, r *http.Request) {
							assert.Equal(t, http.MethodPut, r.Method)
							assert.Equal(t, tokenPath+"/"+tokenId, r.URL.Path)
							assert.Equal(t, apiKey, r.Header.Get("X-API-Key"))
							body, err := io.ReadAll(r.Body)
							require.NoError(t, err)
							assert.JSONEq(t, tokenBody, string(body))
						},
					},
				},
			),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character '\\n' in string literal"), "Could not print JSON Array"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpMultiResponseHandler(t, tc.responses))

			defer srv.Close()

			in := strings.NewReader(tc.input)
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

			updateCmd := NewUpdateCommand(d)

			updateCmd.SilenceUsage = true
			updateCmd.SetIn(ios.In)
			updateCmd.SetOut(ios.Out)
			updateCmd.SetErr(ios.Err)

			updateCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			updateCmd.SetArgs(tc.args)

			err := updateCmd.Execute()
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

// TestNewUpdateCommand_Flags tests the flags of the NewUpdateCommand.
func TestNewUpdateCommand_Flags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		err  string
	}{
		{name: "no service account", args: []string{"My Token", "--scopes", "core:*"}, err: `required flag(s) "service-account" not set`},
		{name: "the TTL can only change by renewing", args: []string{"My Token", "--service-account", "ci", "--ttl", "60d"}, err: `unknown flag: --ttl`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ios := iostreams.IOStreams{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
			d := cli.NewDiscovery(&ios, viper.New(), t.TempDir())

			updateCmd := NewUpdateCommand(d)
			updateCmd.SetIn(ios.In)
			updateCmd.SetOut(ios.Out)
			updateCmd.SetErr(ios.Err)
			updateCmd.SetArgs(tc.args)

			assert.EqualError(t, updateCmd.Execute(), tc.err)
		})
	}
}

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

// TestNewGetCommand tests the NewGetCommand() function.
func TestNewGetCommand(t *testing.T) {
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
			name:      "Search by name returns an array of which the first object is returned",
			args:      []string{"My Token", "--service-account", "ci"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewGetCommand_Out_SearchByNameReturnsObject",
			errGolden: "NewGetCommand_Err_SearchByNameReturnsObject",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_SearchByNameReturnsObject"),
			errBytes:  []byte(nil),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByNamePhase(),
			),
			err: nil,
		},
		{
			name:      "Get with no args returns an array",
			args:      []string{"--service-account", "ci"},
			outGolden: "NewGetCommand_Out_GetAllReturnsArray",
			errGolden: "NewGetCommand_Err_GetAllReturnsArray",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_GetAllReturnsArray"),
			errBytes:  []byte(nil),
			url:       true,
			apiKey:    apiKey,
			responses: mockResponses(
				serviceAccountFoundPhase(),
				map[string]testutils.MockResponse{
					"GET:" + tokenPath: mock(http.StatusOK, tokenGetAllBody, http.MethodGet, tokenPath),
				},
			),
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			args:      []string{"--service-account", "ci"},
			outGolden: "NewGetCommand_Out_NoURL",
			errGolden: "NewGetCommand_Err_NoURL",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_NoURL"),
			url:       false,
			apiKey:    apiKey,
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:      "user sends a service account that does not exist",
			args:      []string{"--service-account", "test"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewGetCommand_Out_ServiceAccountDoesNotExist",
			errGolden: "NewGetCommand_Err_ServiceAccountDoesNotExist",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_ServiceAccountDoesNotExist"),
			responses: serviceAccountNotFoundPhase(),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body:   gjson.Parse(fmt.Sprintf(discoveryPackage.NotFoundError, "test")),
			}, "Could not find service account \"test\""),
		},
		{
			name:      "user sends a token name that does not exist",
			args:      []string{"test", "--service-account", "ci"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewGetCommand_Out_NameDoesNotExist",
			errGolden: "NewGetCommand_Err_NameDoesNotExist",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_NameDoesNotExist"),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				tokenNotFoundPhase(),
			),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body:   gjson.Parse(fmt.Sprintf(discoveryPackage.NotFoundError, "test")),
			}, "Could not search for entity with id \"test\""),
		},
		{
			name:      "Search By Name returns HTTP error",
			args:      []string{tokenId, "--service-account", "ci"},
			outGolden: "NewGetCommand_Out_SearchByNameHTTPError",
			errGolden: "NewGetCommand_Err_SearchByNameHTTPError",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_SearchByNameHTTPError"),
			url:       true,
			apiKey:    apiKey,
			responses: mockResponses(
				serviceAccountFoundPhase(),
				map[string]testutils.MockResponse{
					"POST:" + tokenPath + "/search": mock(http.StatusInternalServerError, `{
			"status": 500,
			"code": 1003,
			"messages": [
				"Internal server error"
			],
			"timestamp": "2025-10-16T17:46:45.386963700Z"
			}`, http.MethodPost, tokenPath+"/search"),
				},
			),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusInternalServerError, Body: gjson.Parse(`{
			"status": 500,
			"code": 1003,
			"messages": [
				"Internal server error"
			],
			"timestamp": "2025-10-16T17:46:45.386963700Z"
			}`)}, "Could not search for entity with id \""+tokenId+"\""),
		},
		{
			name:      "GetEntities returns HTTP error",
			args:      []string{"--service-account", "ci"},
			outGolden: "NewGetCommand_Out_GetEntitiesHTTPError",
			errGolden: "NewGetCommand_Err_GetEntitiesHTTPError",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_GetEntitiesHTTPError"),
			url:       true,
			apiKey:    apiKey,
			responses: mockResponses(
				serviceAccountFoundPhase(),
				map[string]testutils.MockResponse{
					"GET:" + tokenPath: mock(http.StatusUnauthorized, `{"error": "unauthorized"}`, http.MethodGet, tokenPath),
				},
			),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error": "unauthorized"}`)}, "Could not get all entities"),
		},
		{
			name:      "Printing JSON object fails",
			args:      []string{"My Token", "--service-account", "ci"},
			outGolden: "NewGetCommand_Out_PrintJSONFails",
			errGolden: "NewGetCommand_Err_PrintJSONFails",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_PrintJSONFails"),
			url:       true,
			apiKey:    apiKey,
			responses: mockResponses(
				serviceAccountFoundPhase(),
				findTokenByNamePhase(),
				map[string]testutils.MockResponse{
					"GET:" + tokenPath + "/" + tokenId: mock(http.StatusOK, strings.Replace(tokenBody, `"active": true`, `"active: true`, 1), http.MethodGet, tokenPath+"/"+tokenId),
				},
			),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character '\\n' in string literal"), "Could not print JSON object"),
		},
		{
			name:      "Printing JSON array fails",
			args:      []string{"--service-account", "ci"},
			outGolden: "NewGetCommand_Out_PrintArrayFails",
			errGolden: "NewGetCommand_Err_PrintArrayFails",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_PrintArrayFails"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_PrintArrayFails"),
			url:       true,
			apiKey:    apiKey,
			responses: mockResponses(
				serviceAccountFoundPhase(),
				map[string]testutils.MockResponse{
					"GET:" + tokenPath: mock(http.StatusOK, strings.Replace(tokenGetAllBody, `"token_core_infinite"`, `TOKEN`, 1), http.MethodGet, tokenPath),
				},
			),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character 'T' looking for beginning of value"), "Could not print JSON Array"),
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

			if tc.url {
				vpr.Set("default.core_url", srv.URL)
			}
			if tc.apiKey != "" {
				vpr.Set("default.core_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, t.TempDir())

			getCmd := NewGetCommand(d)

			getCmd.SilenceUsage = true
			getCmd.SetIn(ios.In)
			getCmd.SetOut(ios.Out)
			getCmd.SetErr(ios.Err)

			getCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			getCmd.SetArgs(tc.args)

			err := getCmd.Execute()
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

// TestNewGetCommand_NoServiceAccountFlag tests the NewGetCommand when the required service account flag is not sent.
func TestNewGetCommand_NoServiceAccountFlag(t *testing.T) {
	ios := iostreams.IOStreams{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	d := cli.NewDiscovery(&ios, viper.New(), t.TempDir())

	getCmd := NewGetCommand(d)
	getCmd.SetIn(ios.In)
	getCmd.SetOut(ios.Out)
	getCmd.SetErr(ios.Err)
	getCmd.SetArgs([]string{"My Token"})

	err := getCmd.Execute()
	assert.EqualError(t, err, `required flag(s) "service-account" not set`)
}

// TestNewGetCommand_NoProfileFlag tests the NewGetCommand when the profile flag was not defined.
func TestNewGetCommand_NoProfileFlag(t *testing.T) {
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

	getCmd := NewGetCommand(d)

	getCmd.SetIn(ios.In)
	getCmd.SetOut(ios.Out)
	getCmd.SetErr(ios.Err)

	getCmd.SetArgs([]string{"test", "--service-account", "ci"})

	err := getCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewGetCommand_Out_NoProfile", testutils.Read(t, "NewGetCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewGetCommand_Err_NoProfile", testutils.Read(t, "NewGetCommand_Err_NoProfile"), errBuf.Bytes())
}

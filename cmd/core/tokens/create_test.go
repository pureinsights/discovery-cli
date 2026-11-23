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

// TestNewCreateCommand tests the NewCreateCommand() function.
func TestNewCreateCommand(t *testing.T) {
	_, parseBoolErr := strconv.ParseBool("maybe")

	tests := []struct {
		name      string
		args      []string
		stdin     string
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
			name: "Create with flags sends every field",
			args: []string{
				"create-token",
				"--service-account", "ci",
				"--scopes", "core:file:read,core:server:write",
				"--ttl", "1h",
				"--enabled=false"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewCreateCommand_Out_CreateWithFlags",
			errGolden: "NewCreateCommand_Err_CreateWithFlags",
			outBytes:  testutils.Read(t, "NewCreateCommand_Out_CreateWithFlags"),
			errBytes:  []byte(nil),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				map[string]testutils.MockResponse{
					"POST:" + tokenPath: {
						StatusCode:  http.StatusCreated,
						ContentType: "application/json",
						Body:        createdTokenBody("3132c311-885a-4116-be1e-25345f225541", "create-token", "core:file:read,core:server:write", "2026-12-18T16:48:42Z", false),
						Assertions: func(t *testing.T, r *http.Request) {
							assert.Equal(t, http.MethodPost, r.Method)
							assert.Equal(t, tokenPath, r.URL.Path)
							assert.Equal(t, apiKey, r.Header.Get("X-API-Key"))
							body, err := io.ReadAll(r.Body)
							require.NoError(t, err)
							assert.JSONEq(t, `{"name":"create-token","active":false,"scopes":["core:file:read","core:server:write"],"ttl":"1h"}`, string(body))
						},
					},
				},
			),
		},
		{
			name:      "Create with only the required flags",
			args:      []string{"create-token", "--service-account", "ci", "--scopes", "core:file:read, core:server:write"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewCreateCommand_Out_CreateWithRequiredFlags",
			errGolden: "NewCreateCommand_Err_CreateWithRequiredFlags",
			outBytes:  testutils.Read(t, "NewCreateCommand_Out_CreateWithRequiredFlags"),
			errBytes:  []byte(nil),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				map[string]testutils.MockResponse{
					"POST:" + tokenPath: {
						StatusCode:  http.StatusCreated,
						ContentType: "application/json",
						Body:        createdTokenBody("3132c311-885a-4116-be1e-25345f225541", "create-token", "core:file:read,core:server:write", "", true),
						Assertions: func(t *testing.T, r *http.Request) {
							assert.Equal(t, http.MethodPost, r.Method)
							assert.Equal(t, tokenPath, r.URL.Path)
							assert.Equal(t, apiKey, r.Header.Get("X-API-Key"))
							body, err := io.ReadAll(r.Body)
							require.NoError(t, err)
							assert.JSONEq(t, `{"name":"create-token","active":true,"scopes":["core:file:read","core:server:write"]}`, string(body))
						},
					},
				},
			),
		},
		{
			name:      "Create interactively",
			args:      []string{"--service-account", "ci"},
			stdin:     "create-token\ncore:file:read, core:server:write\n1h\n\n",
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewCreateCommand_Out_CreateInteractively",
			errGolden: "NewCreateCommand_Err_CreateInteractively",
			outBytes:  testutils.Read(t, "NewCreateCommand_Out_CreateInteractively"),
			errBytes:  []byte(nil),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				map[string]testutils.MockResponse{
					"POST:" + tokenPath: {
						StatusCode:  http.StatusCreated,
						ContentType: "application/json",
						Body:        createdTokenBody("3132c311-885a-4116-be1e-25345f225541", "create-token", "core:file:read,core:server:write", "", true),
						Assertions: func(t *testing.T, r *http.Request) {
							assert.Equal(t, http.MethodPost, r.Method)
							assert.Equal(t, tokenPath, r.URL.Path)
							assert.Equal(t, apiKey, r.Header.Get("X-API-Key"))
							body, err := io.ReadAll(r.Body)
							require.NoError(t, err)
							assert.JSONEq(t, `{"name":"create-token","active":true,"scopes":["core:file:read","core:server:write"],"ttl":"1h"}`, string(body))
						},
					},
				},
			),
		},

		// Error case
		{
			name:      "No URL",
			args:      []string{"--service-account=ci"},
			outGolden: "NewCreateCommand_Out_NoURL",
			errGolden: "NewCreateCommand_Err_NoURL",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewCreateCommand_Err_NoURL"),
			url:       false,
			apiKey:    apiKey,
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:      "Service Account does not exist",
			args:      []string{"create-token", "--service-account", "test", "--scopes", "core:*"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewCreateCommand_Out_ServiceAccountDoesNotExist",
			errGolden: "NewCreateCommand_Err_ServiceAccountDoesNotExist",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewCreateCommand_Err_ServiceAccountDoesNotExist"),
			responses: serviceAccountNotFoundPhase(),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body:   gjson.Parse(fmt.Sprintf(discoveryPackage.NotFoundError, "test")),
			}, "Could not find service account \"test\""),
		},
		{
			name:      "Invalid active value in interactive mode fails",
			args:      []string{"--service-account=ci"},
			stdin:     "token-2\ncore:*\n\nmaybe\n",
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewCreateCommand_Out_InvalidActive",
			errGolden: "NewCreateCommand_Err_InvalidActive",
			outBytes:  testutils.Read(t, "NewCreateCommand_Out_InvalidActive"),
			errBytes:  testutils.Read(t, "NewCreateCommand_Err_InvalidActive"),
			responses: mockResponses(serviceAccountFoundPhase()),
			err:       cli.NewErrorWithCause(cli.ErrorExitCode, parseBoolErr, "Invalid value for active: \"maybe\""),
		},
		{
			name:      "Create returns HTTP error",
			args:      []string{"create-token", "--service-account", "ci", "--scopes", "Invalid Scope"},
			url:       true,
			apiKey:    apiKey,
			outGolden: "NewCreateCommand_Out_CreateHTTPError",
			errGolden: "NewCreateCommand_Err_CreateHTTPError",
			outBytes:  testutils.Read(t, "NewCreateCommand_Out_CreateHTTPError"),
			errBytes:  testutils.Read(t, "NewCreateCommand_Err_CreateHTTPError"),
			responses: mockResponses(
				serviceAccountFoundPhase(),
				map[string]testutils.MockResponse{
					"POST:" + tokenPath: {
						StatusCode:  http.StatusBadRequest,
						ContentType: "application/json",
						Body: `{
	"status": 400,
	"code": 3002,
	"messages": [
		"dto.scopes[]: Scopes must be product:entity:permission, where entity or permission may be '*'"
	],
	"timestamp": "2026-11-23T10:00:00.000000000Z"
}`,
						Assertions: func(t *testing.T, r *http.Request) {
							assert.Equal(t, http.MethodPost, r.Method)
							assert.Equal(t, tokenPath, r.URL.Path)
							assert.Equal(t, apiKey, r.Header.Get("X-API-Key"))
							body, err := io.ReadAll(r.Body)
							require.NoError(t, err)
							assert.JSONEq(t, `{"name":"create-token","active":true,"scopes":["Invalid Scope"]}`, string(body))
						},
					},
				},
			),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
	"status": 400,
	"code": 3002,
	"messages": [
		"dto.scopes[]: Scopes must be product:entity:permission, where entity or permission may be '*'"
	],
	"timestamp": "2026-11-23T10:00:00.000000000Z"
}`)}, "Could not store entities"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpMultiResponseHandler(t, tc.responses))

			defer srv.Close()

			in := strings.NewReader(tc.stdin)
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

			createCmd := NewCreateCommand(d)

			createCmd.SilenceUsage = true
			createCmd.SetIn(ios.In)
			createCmd.SetOut(ios.Out)
			createCmd.SetErr(ios.Err)

			createCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			createCmd.SetArgs(tc.args)

			err := createCmd.Execute()
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

// TestNewCreateCommand_NoProfileFlag tests the NewCreateCommand when the profile flag was not defined.
func TestNewCreateCommand_NoProfileFlag(t *testing.T) {
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

	createCmd := NewCreateCommand(d)

	createCmd.SetIn(ios.In)
	createCmd.SetOut(ios.Out)
	createCmd.SetErr(ios.Err)

	createCmd.SetArgs([]string{"create-token", "--service-account", "ci"})

	err := createCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewCreateCommand_Out_NoProfile", testutils.Read(t, "NewCreateCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewCreateCommand_Err_NoProfile", testutils.Read(t, "NewCreateCommand_Err_NoProfile"), errBuf.Bytes())
}

// TestNewCreateCommand_TooManyArgs tests the NewCreateCommand function when it receives more than one argument.
func TestNewCreateCommand_TooManyArgs(t *testing.T) {
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

	createCmd := NewCreateCommand(d)

	createCmd.SetIn(ios.In)
	createCmd.SetOut(ios.Out)
	createCmd.SetErr(ios.Err)

	// --enabled false instead of --enabled=false, so "false" is taken as a second argument
	createCmd.SetArgs([]string{"token", "--service-account", "ci", "--enabled", "false"})

	err := createCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, "accepts at most 1 arg(s), received 2")

	testutils.CompareBytes(t, "NewCreateCommand_Out_TooManyArgs", testutils.Read(t, "NewCreateCommand_Out_TooManyArgs"), out.Bytes())
	testutils.CompareBytes(t, "NewCreateCommand_Err_TooManyArgs", testutils.Read(t, "NewCreateCommand_Err_TooManyArgs"), errBuf.Bytes())
}

// TestNewCreateCommand_NoServiceAccountFlag tests the NewCreateCommand when the required service account flag is not sent.
func TestNewCreateCommand_NoServiceAccountFlag(t *testing.T) {
	ios := iostreams.IOStreams{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	d := cli.NewDiscovery(&ios, viper.New(), t.TempDir())

	createCmd := NewCreateCommand(d)
	createCmd.SetIn(ios.In)
	createCmd.SetOut(ios.Out)
	createCmd.SetErr(ios.Err)
	createCmd.SetArgs([]string{"create-token", "--scopes", "core:*"})

	assert.EqualError(t, createCmd.Execute(), `required flag(s) "service-account" not set`)
}

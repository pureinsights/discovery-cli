package serviceaccounts

import (
	"bytes"
	"errors"
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

const createdAccount = `{
	"name": "ci",
	"description": "Continuous Integration account",
	"labels": [
		{
			"key": "keyA",
			"value": "valueA"
		},
		{
			"key": "keyB",
			"value": "valueB"
		}
	],
	"active": true,
	"id": "06f52112-a2bb-4490-97c2-171fb7d54d5b",
	"creationTimestamp": "2026-11-04T15:37:47Z",
	"lastUpdatedTimestamp": "2026-11-04T15:37:47Z",
	"role": "OPERATE"
}`

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
			name:      "Create with flags sends every field",
			args:      []string{"ci", "--description", "Continuous Integration account", "--role", "OPERATE", "--labels", "keyA:valueA,keyB:valueB", "--enabled"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewCreateCommand_Out_CreateWithFlags",
			errGolden: "NewCreateCommand_Err_CreateWithFlags",
			outBytes:  testutils.Read(t, "NewCreateCommand_Out_CreateWithFlags"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/account/service": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body:        createdAccount,
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/account/service", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
						body, err := io.ReadAll(r.Body)
						require.NoError(t, err)
						assert.JSONEq(t, `{"name":"ci","description":"Continuous Integration account","role":"OPERATE","labels":[{"key":"keyA","value":"valueA"},{"key":"keyB","value":"valueB"}],"active":true}`, string(body))
					},
				},
			},
			err: nil,
		},
		{
			name:      "Create with only the name and role sends an active account",
			args:      []string{"ci", "--role", "OPERATE"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewCreateCommand_Out_CreateOnlyName",
			errGolden: "NewCreateCommand_Err_CreateOnlyName",
			outBytes:  testutils.Read(t, "NewCreateCommand_Out_CreateOnlyName"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/account/service": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body:        `{"name":"ci","labels":[],"active":true,"id":"06f52112-a2bb-4490-97c2-171fb7d54d5b","role":"OPERATE"}`,
					Assertions: func(t *testing.T, r *http.Request) {
						body, err := io.ReadAll(r.Body)
						require.NoError(t, err)
						assert.JSONEq(t, `{"name":"ci","active":true,"role":"OPERATE"}`, string(body))
					},
				},
			},
			err: nil,
		},
		{
			name:      "Create a disabled account",
			args:      []string{"ci", "--role", "OPERATE", "--enabled=false"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewCreateCommand_Out_CreateDisabled",
			errGolden: "NewCreateCommand_Err_CreateDisabled",
			outBytes:  testutils.Read(t, "NewCreateCommand_Out_CreateDisabled"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/account/service": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body:        `{"name":"ci","labels":[],"active":false,"id":"06f52112-a2bb-4490-97c2-171fb7d54d5b","role":"OPERATE"}`,
					Assertions: func(t *testing.T, r *http.Request) {
						body, err := io.ReadAll(r.Body)
						require.NoError(t, err)
						assert.JSONEq(t, `{"name":"ci","role":"OPERATE","active":false}`, string(body))
					},
				},
			},
			err: nil,
		},
		{
			name:      "Create interactively asks for every field",
			args:      []string{},
			stdin:     "ci\nContinuous Integration account\nOPERATE\nkeyA:valueA,keyB:valueB\n\n",
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewCreateCommand_Out_CreateInteractive",
			errGolden: "NewCreateCommand_Err_CreateInteractive",
			outBytes:  testutils.Read(t, "NewCreateCommand_Out_CreateInteractive"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/account/service": {
					StatusCode:  http.StatusOK,
					ContentType: "application/json",
					Body:        createdAccount,
					Assertions: func(t *testing.T, r *http.Request) {
						body, err := io.ReadAll(r.Body)
						require.NoError(t, err)
						assert.JSONEq(t, `{"name":"ci","description":"Continuous Integration account","role":"OPERATE","labels":[{"key":"keyA","value":"valueA"},{"key":"keyB","value":"valueB"}],"active":true}`, string(body))
					},
				},
			},
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			args:      []string{"ci"},
			outGolden: "NewCreateCommand_Out_NoURL",
			errGolden: "NewCreateCommand_Err_NoURL",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewCreateCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:      "Flags without a name return Core's validation error",
			args:      []string{"--role", "OPERATE"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewCreateCommand_Out_NoName",
			errGolden: "NewCreateCommand_Err_NoName",
			outBytes:  testutils.Read(t, "NewCreateCommand_Out_NoName"),
			errBytes:  testutils.Read(t, "NewCreateCommand_Err_NoName"),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/account/service": {
					StatusCode:  http.StatusBadRequest,
					ContentType: "application/json",
					Body:        `{"status":400,"code":3002,"messages":["dto.name: must not be blank"],"timestamp":"2026-11-04T09:48:20.236954600Z"}`,
					Assertions: func(t *testing.T, r *http.Request) {
						body, err := io.ReadAll(r.Body)
						require.NoError(t, err)
						assert.JSONEq(t, `{"role":"OPERATE","active":true}`, string(body))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{"status":400,"code":3002,"messages":["dto.name: must not be blank"],"timestamp":"2026-11-04T09:48:20.236954600Z"}`)}, "Could not store entities"),
		},
		{
			name:      "Invalid active value in interactive mode fails",
			args:      []string{},
			stdin:     "ci\n\n\n\nmaybe\n",
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewCreateCommand_Out_InvalidActive",
			errGolden: "NewCreateCommand_Err_InvalidActive",
			outBytes:  testutils.Read(t, "NewCreateCommand_Out_InvalidActive"),
			errBytes:  testutils.Read(t, "NewCreateCommand_Err_InvalidActive"),
			err:       cli.NewErrorWithCause(cli.ErrorExitCode, parseBoolErr, "Invalid value for active: %q", "maybe"),
		},
		{
			name:      "Create returns HTTP error",
			args:      []string{"ci", "--role", "OPERATE"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewCreateCommand_Out_CreateHTTPError",
			errGolden: "NewCreateCommand_Err_CreateHTTPError",
			outBytes:  testutils.Read(t, "NewCreateCommand_Out_CreateHTTPError"),
			errBytes:  testutils.Read(t, "NewCreateCommand_Err_CreateHTTPError"),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/account/service": {
					StatusCode:  http.StatusConflict,
					ContentType: "application/json",
					Body: `{
	"status": 409,
	"code": 4001,
	"messages": [
		"Duplicated entity: ci"
	],
	"timestamp": "2026-11-04T15:37:47.000000000Z"
}`,
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
	"status": 409,
	"code": 4001,
	"messages": [
		"Duplicated entity: ci"
	],
	"timestamp": "2026-11-04T15:37:47.000000000Z"
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

	createCmd.SetArgs([]string{"ci"})

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

	// --enabled false instead of --enabled=false
	createCmd.SetArgs([]string{"ci", "--enabled", "false"})

	err := createCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, "accepts at most 1 arg(s), received 2")

	testutils.CompareBytes(t, "NewCreateCommand_Out_TooManyArgs", testutils.Read(t, "NewCreateCommand_Out_TooManyArgs"), out.Bytes())
	testutils.CompareBytes(t, "NewCreateCommand_Err_TooManyArgs", testutils.Read(t, "NewCreateCommand_Err_TooManyArgs"), errBuf.Bytes())
}

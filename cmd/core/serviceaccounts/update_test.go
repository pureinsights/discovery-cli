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

const existingAccount = `{
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
	"creationTimestamp": "2026-10-29T15:37:47Z",
	"lastUpdatedTimestamp": "2026-11-03T11:50:39Z",
	"role": "OPERATE"
}`

// existingAccountResponses returns the responses needed to find the "ci" service account by name, plus the response to its update.
// The update response asserts the request body is the expected JSON.
func existingAccountResponses(updateStatus int, updateBody, expectedRequest string) map[string]testutils.MockResponse {
	return map[string]testutils.MockResponse{
		"POST:/v2/account/service/search": {
			StatusCode:  http.StatusOK,
			ContentType: "application/json",
			Body:        `{"content":[{"source":` + existingAccount + `,"highlight":{},"score":1}],"totalSize":1,"totalPages":1,"empty":false,"numberOfElements":1,"pageNumber":0}`,
			Assertions: func(t *testing.T, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/v2/account/service/search", r.URL.Path)
			},
		},
		"GET:/v2/account/service/06f52112-a2bb-4490-97c2-171fb7d54d5b": {
			StatusCode:  http.StatusOK,
			ContentType: "application/json",
			Body:        existingAccount,
			Assertions: func(t *testing.T, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
			},
		},
		"PUT:/v2/account/service/06f52112-a2bb-4490-97c2-171fb7d54d5b": {
			StatusCode:  updateStatus,
			ContentType: "application/json",
			Body:        updateBody,
			Assertions: func(t *testing.T, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				assert.JSONEq(t, expectedRequest, string(body))
			},
		},
	}
}

// TestNewUpdateCommand tests the NewUpdateCommand() function.
func TestNewUpdateCommand(t *testing.T) {
	_, parseBoolErr := strconv.ParseBool("maybe")

	adminAccount := strings.Replace(existingAccount, `"OPERATE"`, `"ADMIN"`, 1)
	internalServerError := `{
	"status": 500,
	"code": 1003,
	"messages": [
		"Internal server error"
	],
	"timestamp": "2026-11-04T15:37:47.000000000Z"
}`

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
			name:      "Update with flags only changes the given fields",
			args:      []string{"ci", "--role", "ADMIN"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewUpdateCommand_Out_UpdateWithFlags",
			errGolden: "NewUpdateCommand_Err_UpdateWithFlags",
			outBytes:  testutils.Read(t, "NewUpdateCommand_Out_UpdateWithFlags"),
			errBytes:  []byte(nil),
			responses: existingAccountResponses(http.StatusOK, adminAccount, adminAccount),
			err:       nil,
		},
		{
			name:      "Update with flags replaces the labels and disables the account",
			args:      []string{"ci", "--labels", "keyC:valueC", "--enabled=false"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewUpdateCommand_Out_UpdateLabelsAndDisable",
			errGolden: "NewUpdateCommand_Err_UpdateLabelsAndDisable",
			outBytes:  testutils.Read(t, "NewUpdateCommand_Out_UpdateLabelsAndDisable"),
			errBytes:  []byte(nil),
			responses: func() map[string]testutils.MockResponse {
				expected := `{
	"name": "ci",
	"description": "Continuous Integration account",
	"labels": [{"key": "keyC", "value": "valueC"}],
	"active": false,
	"id": "06f52112-a2bb-4490-97c2-171fb7d54d5b",
	"creationTimestamp": "2026-10-29T15:37:47Z",
	"lastUpdatedTimestamp": "2026-11-03T11:50:39Z",
	"role": "OPERATE"
}`
				return existingAccountResponses(http.StatusOK, expected, expected)
			}(),
			err: nil,
		},
		{
			name:      "Update interactively shows the current values",
			args:      []string{"ci"},
			stdin:     "\n \nADMIN\n\n\n",
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewUpdateCommand_Out_UpdateInteractive",
			errGolden: "NewUpdateCommand_Err_UpdateInteractive",
			outBytes:  testutils.Read(t, "NewUpdateCommand_Out_UpdateInteractive"),
			errBytes:  []byte(nil),
			responses: func() map[string]testutils.MockResponse {
				expected := strings.Replace(adminAccount, `"description": "Continuous Integration account"`, `"description": ""`, 1)
				return existingAccountResponses(http.StatusOK, expected, expected)
			}(),
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			args:      []string{"ci", "--role", "ADMIN"},
			outGolden: "NewUpdateCommand_Out_NoURL",
			errGolden: "NewUpdateCommand_Err_NoURL",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewUpdateCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:      "sent name does not exist",
			args:      []string{"test", "--role", "ADMIN"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewUpdateCommand_Out_NameDoesNotExist",
			errGolden: "NewUpdateCommand_Err_NameDoesNotExist",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewUpdateCommand_Err_NameDoesNotExist"),
			responses: map[string]testutils.MockResponse{
				"POST:/v2/account/service/search": {
					StatusCode:  http.StatusNoContent,
					Body:        ``,
					ContentType: "application/json",
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name "test" does not exist"
	]
}`),
			}, "Could not get service account ID."),
		},
		{
			name:      "Invalid active value in interactive mode fails",
			args:      []string{"ci"},
			stdin:     "\n\n\n\nmaybe\n",
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewUpdateCommand_Out_InvalidActive",
			errGolden: "NewUpdateCommand_Err_InvalidActive",
			outBytes:  testutils.Read(t, "NewUpdateCommand_Out_InvalidActive"),
			errBytes:  testutils.Read(t, "NewUpdateCommand_Err_InvalidActive"),
			responses: existingAccountResponses(http.StatusOK, existingAccount, existingAccount),
			err:       cli.NewErrorWithCause(cli.ErrorExitCode, parseBoolErr, "Invalid value for active: %q", "maybe"),
		},
		{
			name:      "Update returns HTTP error",
			args:      []string{"ci", "--role", "ADMIN"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewUpdateCommand_Out_UpdateHTTPError",
			errGolden: "NewUpdateCommand_Err_UpdateHTTPError",
			outBytes:  testutils.Read(t, "NewUpdateCommand_Out_UpdateHTTPError"),
			errBytes:  testutils.Read(t, "NewUpdateCommand_Err_UpdateHTTPError"),
			responses: existingAccountResponses(http.StatusInternalServerError, internalServerError, adminAccount),
			err:       cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusInternalServerError, Body: gjson.Parse(internalServerError)}, "Could not store entities"),
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

// TestNewUpdateCommand_NoProfileFlag tests the NewUpdateCommand when the profile flag was not defined.
func TestNewUpdateCommand_NoProfileFlag(t *testing.T) {
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

	updateCmd := NewUpdateCommand(d)

	updateCmd.SetIn(ios.In)
	updateCmd.SetOut(ios.Out)
	updateCmd.SetErr(ios.Err)

	updateCmd.SetArgs([]string{"ci"})

	err := updateCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewUpdateCommand_Out_NoProfile", testutils.Read(t, "NewUpdateCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewUpdateCommand_Err_NoProfile", testutils.Read(t, "NewUpdateCommand_Err_NoProfile"), errBuf.Bytes())
}

// TestNewUpdateCommand_NotExactly1Arg tests the NewUpdateCommand function when it does not receive exactly one argument.
func TestNewUpdateCommand_NotExactly1Arg(t *testing.T) {
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

	updateCmd := NewUpdateCommand(d)

	updateCmd.SetIn(ios.In)
	updateCmd.SetOut(ios.Out)
	updateCmd.SetErr(ios.Err)

	updateCmd.SetArgs([]string{})

	err := updateCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")

	testutils.CompareBytes(t, "NewUpdateCommand_Out_NotExactly1Arg", testutils.Read(t, "NewUpdateCommand_Out_NotExactly1Arg"), out.Bytes())
	testutils.CompareBytes(t, "NewUpdateCommand_Err_NotExactly1Arg", testutils.Read(t, "NewUpdateCommand_Err_NotExactly1Arg"), errBuf.Bytes())
}

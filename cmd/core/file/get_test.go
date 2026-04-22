package file

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
			name: "Get returns an array",
			args:      []string{},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewGetCommand_Out_GetFilesReturnsArray",
			errGolden: "NewGetCommand_Err_GetFilesReturnsArray",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_GetFilesReturnsArray"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"GET:/v2/file": {
					StatusCode: http.StatusOK,
					Body: `
					[
						"Credential.ndjson",
						"Server.ndjson",
						"buildContextPrompt.js",
						"buildSimplePrompt.js",
						"constructPrompt.js",
						"constructSuggestedPrompt.js",
						"elastic-extraction.py",
						"extractReference.groovy",
						"extractReferenceAtlas.groovy",
						"format/formatAnalysisResponse.js",
						"format/formatAutocompleteResponse.js",
						"format/formatChunksResponse.js",
						"format/formatKeywordResponse.js",
						"format/formatKeywordResponseAtlas.js",
						"format/formatKeywordSearch.js",
						"format/formatQuestionsResponse.js",
						"format/formatSearchResponse.js",
						"format/formatSearchResponseAtlas.js",
						"format/formatSemanticResponse.js",
						"format/formatSuggestionsResponse.js",
						"templates/keywordSearchTemplateAtlas.json",
						"templates/searchTemplate.json",
						"templates/searchTemplateAtlas.json",
					]
					`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/file", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},
		{
			name: "Get returns nothing",
			args:      []string{},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewGetCommand_Out_GetFilesReturnsNothing",
			errGolden: "NewGetCommand_Err_GetFilesReturnsNothing",
			outBytes:  []byte(nil),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"GET:/v2/file": {
					StatusCode: http.StatusNoContent,
					Body: ``,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/file", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},
		// Error Case
		{
			name:      "No URL",
			args:      []string{},
			url:       false,
			apiKey:    "apiKey123",
			outGolden: "NewGetCommand_Out_NoURL",
			errGolden: "NewGetCommand_Err_NoURL",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_NoURL"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_NoURL"),
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name: "GetFileList returns HTTP error",
			args:      []string{},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewGetCommand_Out_GetFileListHTTPError",
			errGolden: "NewGetCommand_Err_GetFileListHTTPError",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_GetFileListHTTPError"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_GetFileListHTTPError"),
			responses: map[string]testutils.MockResponse{
				"GET:/v2/file": {
					StatusCode: http.StatusInternalServerError,
					Body: `{
			"status": 500,
			"code": 1003,
			"messages": [
				"Internal server error"
			],
			"timestamp": "2025-10-16T17:46:45.386963700Z"
			}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/file", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusInternalServerError, Body: gjson.Parse(`{
			"status": 500,
			"code": 1003,
			"messages": [
				"Internal server error"
			],
			"timestamp": "2025-10-16T17:46:45.386963700Z"
			}`)}, "Could not get file list"),
		},
		{
			name: "Printing JSON array fails",
			args:      []string{},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewGetCommand_Out_PrintArrayFails",
			errGolden: "NewGetCommand_Err_PrintArrayFails",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_PrintArrayFails"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_PrintArrayFails"),
			responses: map[string]testutils.MockResponse{
				"GET:/v2/file": {
					StatusCode: http.StatusOK,
					Body: `
					[
						"Credential.ndjson",
						"Server.ndjson",
						"buildContextPrompt.js",
						"buildSimplePrompt.js",
						constructPrompt.js,
						"constructSuggestedPrompt.js",
						"elastic-extraction.py",
						"extractReference.groovy",
						"extractReferenceAtlas.groovy",
						"format/formatAnalysisResponse.js",
						"format/formatAutocompleteResponse.js",
						"format/formatChunksResponse.js",
						"format/formatKeywordResponse.js",
						"format/formatKeywordResponseAtlas.js",
						"format/formatKeywordSearch.js",
						"format/formatQuestionsResponse.js",
						"format/formatSearchResponse.js",
						"format/formatSearchResponseAtlas.js",
						"format/formatSemanticResponse.js",
						"format/formatSuggestionsResponse.js",
						"templates/keywordSearchTemplateAtlas.json",
						"templates/searchTemplate.json",
						"templates/searchTemplateAtlas.json",
					]
					`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/file", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character 's' in literal null (expecting 'u')"), "Could not print JSON Array"),
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

			testutils.CompareBytes(t, tc.outGolden, tc.outBytes, out.Bytes())
		})
	}
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

	getCmd.SetArgs([]string{})

	err := getCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewGetCommand_Out_NoProfile", testutils.Read(t, "NewGetCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewGetCommand_Err_NoProfile", testutils.Read(t, "NewGetCommand_Err_NoProfile"), errBuf.Bytes())
}
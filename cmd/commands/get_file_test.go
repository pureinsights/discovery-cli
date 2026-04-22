package commands

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/pureinsights/discovery-cli/internal/testutils/mocks"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestGetCommand tests the GetCommand() function.
func TestGetFilesCommand(t *testing.T) {
	tests := []struct {
		name           string
		client         cli.CoreFileController
		args           []string
		url            string
		apiKey         string
		componentName  string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "GetFiles correctly prints the list of files correctly",
			url:            "http://localhost:12010",
			apiKey:         "core123",
			componentName:  "Core",
			args:           []string{},
			client:         new(mocks.WorkingFileClient),
			expectedOutput: "\"Credential.ndjson\"\n\"Server.ndjson\"\n\"buildContextPrompt.js\"\n\"buildSimplePrompt.js\"\n\"constructPrompt.js\"\n\"constructSuggestedPrompt.js\"\n\"elastic-extraction.py\"\n\"extractReference.groovy\"\n\"extractReferenceAtlas.groovy\"\n\"formatAnalysisResponse.js\"\n\"formatAutocompleteResponse.js\"\n\"formatChunksResponse.js\"\n\"formatKeywordResponse.js\"\n\"formatKeywordResponseAtlas.js\"\n\"formatKeywordSearch.js\"\n\"formatQuestionsResponse.js\"\n\"formatSearchResponse.js\"\n\"formatSearchResponseAtlas.js\"\n\"formatSemanticResponse.js\"\n\"formatSuggestionsResponse.js\"\n\"keywordSearchTemplateAtlas.json\"\n\"searchTemplate.json\"\n\"searchTemplateAtlas.json\"\n",
			err:            nil,
		},
		// Error case
		{
			name:          "CheckCredentials fails",
			client:        new(mocks.WorkingFileClient),
			url:           "",
			apiKey:        "core123",
			componentName: "Core",
			args:          []string{},
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
			err:           cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			var out io.Writer
			if tc.outWriter != nil {
				out = tc.outWriter
			} else {
				out = buf
			}

			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: out,
				Err: os.Stderr,
			}

			vpr := viper.New()
			vpr.Set("profile", "default")
			vpr.Set("output", "pretty-json")
			if tc.url != "" {
				vpr.Set("default.core_url", tc.url)
			}
			if tc.apiKey != "" {
				vpr.Set("default.core_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, "")
			err := GetFilesCommand(tc.args, d, tc.client, GetCommandConfig("default", "pretty-json", tc.componentName, "core_url"))

			if tc.err != nil {
				require.Error(t, err)
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}

// TestSearchCommand tests the SearchCommand() function.
func TestDownloadCommand(t *testing.T) {
	tests := []struct {
		name           string
		client         cli.CoreFileController
		args           []string
		filters        []string
		url            string
		apiKey         string
		output		   string
		componentName  string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "Download a file",
			args:           []string{"test"},
			url:            "http://localhost:12010",
			apiKey:         "apiKey123",
			componentName:  "Core",
			client:         new(mocks.WorkingFileClient),
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		{
			name:           "Download a file with output",
			args:           []string{"test"},
			output: 		"./test",
			url:            "http://localhost:12010",
			apiKey:         "apiKey123",
			componentName:  "Core",
			client:         new(mocks.WorkingFileClient),
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		// Error case
		{
			name:          "CheckCredentials fails",
			client:        new(mocks.FailingFileClient),
			apiKey:        "core123",
			componentName: "Core",
			args:          []string{"test"},
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
			err:           cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:           "User sends a key that does not exist",
			args:           []string{"test"},
			url:            "http://localhost:12010",
			apiKey:         "apiKey123",
			client:         new(mocks.FailingFileClient),
			expectedOutput: "",
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Result{},
			}, "Could not get file with key \"test\""),
		},
	}
	testutils.ChangeDirectoryHelper(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			var out io.Writer
			if tc.outWriter != nil {
				out = tc.outWriter
			} else {
				out = buf
			}

			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: out,
				Err: os.Stderr,
			}

			vpr := viper.New()
			vpr.Set("profile", "default")
			if tc.url != "" {
				vpr.Set("default.core_url", tc.url)
			}
			if tc.apiKey != "" {
				vpr.Set("default.core_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, "")
			err := DownloadCommand(tc.args, d, tc.client, GetCommandConfig("default", "pretty-json", tc.componentName, "core_url"),tc.output)

			if tc.err != nil {
				require.Error(t, err)
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}

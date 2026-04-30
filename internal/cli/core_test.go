package cli

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/pureinsights/discovery-cli/internal/testutils/mocks"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// function used to get the exact error of opening a directory that doesn't exist using the
// ReadDir() function or Stat() function
func getDirectoryDoesNotExistError(t *testing.T, path string, useReadDir bool) error {
	var err error
	if useReadDir {
		_, err = os.ReadDir(path)
	} else {
		_, err = os.Stat(path)
	}
	require.Error(t, err)
	return NormalizeReadFileError(path,err)
}

// TestRecursiveStore tests the recursiveStore() auxiliary function used by discovery.StoreFiles()
func TestRecursiveStore(t *testing.T) {
	tests := []struct {
		name           string
		client         CoreFileController
		path           string
		recursive      bool
		expectedOutput gjson.Result
		err            error
	}{
		//Working Case
		{
			name:           "Store a directory",
			path:           "./testdata",
			client:         new(mocks.WorkingFileClient),
			recursive:      false,
			expectedOutput: gjson.Parse(`{"acknowledged": true}`),
			err:            nil,
		},
		{
			name:           "Store a directory recursively",
			path:           "./",
			client:         new(mocks.WorkingFileClient),
			recursive:      true,
			expectedOutput: gjson.Parse(`{"acknowledged": true}`),
			err:            nil,
		},
		{
			name:           "Directory is empty",
			path:           t.TempDir(),
			client:         new(mocks.WorkingFileClient),
			expectedOutput: gjson.Parse(`{"acknowledged": false}`),
			err:            nil,
		},
		//Error Case
		{
			name:           "Directory does not exist",
			path:           "./test/",
			recursive:      false,
			client:         new(mocks.FailingFileClient),
			expectedOutput: gjson.Result{},
			err:            getDirectoryDoesNotExistError(t, "./test/", true),
		},
		{
			name:           "Invalid format error",
			path:           "./testdata",
			recursive:      false,
			client:         new(mocks.FailingFileClient),
			expectedOutput: gjson.Result{},
			err: discoveryPackage.Error{
				Status: http.StatusBadRequest,
				Body: gjson.Parse(`{
	"status": 400,
	"code": 3002,
	"messages": [
		"key: Invalid format for file path, use only alphanumeric symbols with a limit of 255 characters and a max of 10 path levels."
	],
	"timestamp": "2025-10-16T17:46:45.386963700Z"
}`)},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			response, err := recursiveStore(tc.client, tc.path, tc.path, tc.recursive)

			if tc.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, response)
			}
		})
	}
}

// Test_discovery_StoreFiles tests the discovery.StoreFiles() function.
func Test_discovery_StoreFiles(t *testing.T) {
	tests := []struct {
		name           string
		client         CoreFileController
		path           string
		recursive      bool
		expectedOutput string
		err            error
	}{
		//Working Case
		{
			name:           "Store a file",
			path:           "./testdata/discovery.zip",
			client:         new(mocks.WorkingFileClient),
			recursive:      false,
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		{
			name:           "Store a directory",
			path:           "./testdata",
			client:         new(mocks.WorkingFileClient),
			recursive:      false,
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		{
			name:           "Store a directory recursively",
			path:           "./",
			client:         new(mocks.WorkingFileClient),
			recursive:      true,
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		{
			name:           "Directory is empty",
			path:           t.TempDir(),
			client:         new(mocks.WorkingFileClient),
			expectedOutput: "{\n  \"acknowledged\": false\n}\n",
			err:            nil,
		},
		//Error Case
		{
			name:      "Directory does not exist",
			path:      "./test/",
			recursive: false,
			client:    new(mocks.FailingFileClient),
			err: NewErrorWithCause(
				ErrorExitCode,
				getDirectoryDoesNotExistError(t, "./test/", false),
				"The path \"./test/\" does not exist",
			),
		},
		{
			name:      "Invalid format error for directory",
			path:      "./testdata",
			recursive: false,
			client:    new(mocks.FailingFileClient),
			err: NewErrorWithCause(
				ErrorExitCode,
				discoveryPackage.Error{
					Status: http.StatusBadRequest,
					Body: gjson.Parse(`{
	"status": 400,
	"code": 3002,
	"messages": [
		"key: Invalid format for file path, use only alphanumeric symbols with a limit of 255 characters and a max of 10 path levels."
	],
	"timestamp": "2025-10-16T17:46:45.386963700Z"
}`)},
				"Could not store directory \"./testdata\"",
			),
		},
		{
			name:      "Invalid format error for file",
			path:      "./testdata/discovery.zip",
			recursive: false,
			client:    new(mocks.FailingFileClient),
			err: NewErrorWithCause(
				ErrorExitCode,
				discoveryPackage.Error{
					Status: http.StatusBadRequest,
					Body: gjson.Parse(`{
	"status": 400,
	"code": 3002,
	"messages": [
		"key: Invalid format for file path, use only alphanumeric symbols with a limit of 255 characters and a max of 10 path levels."
	],
	"timestamp": "2025-10-16T17:46:45.386963700Z"
}`)},
				"Could not store the file with path \"./testdata/discovery.zip\"",
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := strings.NewReader("")
			out := &bytes.Buffer{}

			errBuf := &bytes.Buffer{}
			ios := iostreams.IOStreams{
				In:  in,
				Out: out,
				Err: errBuf,
			}

			d := NewDiscovery(&ios, viper.New(), "")

			err := d.StoreFiles(tc.client, tc.path, tc.recursive, nil)

			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, out.String())
			}
		})
	}
}

// Test_discovery_GetFile tests the GetFile auxiliary function used in discovery.GetFiles().
func Test_discovery_GetFile(t *testing.T) {
	tests := []struct {
		name           string
		client         CoreFileController
		file           string
		fileContent    []byte
		fileOutput     string
		expectedOutput gjson.Result
		err            error
	}{
		//Working Case
		{
			name:   "GetFile Obtains an existing file",
			client: new(mocks.WorkingFileClient),
			file:   "script.py",
			fileContent: []byte(`
	def main():
		print("Hello, World!")

	if __name__ == "__main__":
		main()

	`),
			expectedOutput: gjson.Parse(`{"acknowledged": true}`),
			err:            nil,
		},
		{
			name:   "GetFile Obtains an existing file with output",
			client: new(mocks.WorkingFileClient),
			file:   "script.py",
			fileContent: []byte(`
	def main():
		print("Hello, World!")

	if __name__ == "__main__":
		main()

	`),
			fileOutput:     "./test",
			expectedOutput: gjson.Parse(`{"acknowledged": true}`),
			err:            nil,
		},
		//Error Case
		{
			name:           "Key does not exists",
			client:         new(mocks.FailingFileClient),
			file:           "script.py",
			expectedOutput: gjson.Result{},
			err: NewErrorWithCause(
				ErrorExitCode,
				discoveryPackage.Error{
					Status: http.StatusNotFound,
					Body:   gjson.Result{},
				},
				"Could not get file with key \"script.py\"",
			),
		},
	}
	testutils.ChangeDirectoryHelper(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var output string
			if tc.fileOutput == "" {
				output = "."
			} else {
				output = tc.fileOutput
			}

			response, err := GetFile(tc.client, tc.file, output)

			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, response)
				file, err := os.ReadFile(filepath.Join(".", tc.file))
				require.NoError(t, err)
				require.Equal(t, tc.fileContent, file)
			}
		})
	}
}

// Test_discovery_GetFiles tests the discovery.GetFiles function.
func Test_discovery_GetFiles(t *testing.T) {
	tests := []struct {
		name           string
		client         CoreFileController
		files          []string
		filesContent   [][]byte
		filesOutput    string
		expectedOutput string
		err            error
	}{
		//Working Case
		{
			name:   "GetFiles Obtains an existing file",
			client: new(mocks.WorkingFileClient),
			files:  []string{"script.py"},
			filesContent: [][]byte{[]byte(`
	def main():
		print("Hello, World!")

	if __name__ == "__main__":
		main()

	`)},
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		{
			name:   "GetFiles Obtains an existing file with output",
			client: new(mocks.WorkingFileClient),
			files:  []string{"script.py"},
			filesContent: [][]byte{[]byte(`
	def main():
		print("Hello, World!")

	if __name__ == "__main__":
		main()

	`)},
			filesOutput:    "./test",
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		//Error Case
		{
			name:   "Key does not exists",
			client: new(mocks.FailingFileClient),
			files:  []string{"script.py"},
			err: NewErrorWithCause(
				ErrorExitCode,
				discoveryPackage.Error{
					Status: http.StatusNotFound,
					Body:   gjson.Result{},
				},
				"Could not get file with key \"script.py\"",
			),
		},
	}
	testutils.ChangeDirectoryHelper(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var output string
			if tc.filesOutput == "" {
				output = "."
			} else {
				output = tc.filesOutput
			}

			in := strings.NewReader("")
			out := &bytes.Buffer{}

			errBuf := &bytes.Buffer{}
			ios := iostreams.IOStreams{
				In:  in,
				Out: out,
				Err: errBuf,
			}

			d := NewDiscovery(&ios, viper.New(), "")

			err := d.GetFiles(tc.client, tc.files, output, nil)

			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, out.String())
				for i, file := range tc.files {
					fileContent, err := os.ReadFile(filepath.Join(".", file))
					require.NoError(t, err)
					require.Equal(t, tc.filesContent[i], fileContent)
				}

			}
		})
	}
}

// Test_discovery_GetFileList tests the discovery.GetFileList function.
func Test_discovery_GetFileList(t *testing.T) {
	tests := []struct {
		name           string
		client         CoreFileController
		expectedOutput string
		err            error
	}{
		//Working Case
		{
			name:           "Prints list of files",
			client:         new(mocks.WorkingFileClient),
			expectedOutput: "\"Credential.ndjson\"\n\"Server.ndjson\"\n\"buildContextPrompt.js\"\n\"buildSimplePrompt.js\"\n\"constructPrompt.js\"\n\"constructSuggestedPrompt.js\"\n\"elastic-extraction.py\"\n\"extractReference.groovy\"\n\"extractReferenceAtlas.groovy\"\n\"formatAnalysisResponse.js\"\n\"formatAutocompleteResponse.js\"\n\"formatChunksResponse.js\"\n\"formatKeywordResponse.js\"\n\"formatKeywordResponseAtlas.js\"\n\"formatKeywordSearch.js\"\n\"formatQuestionsResponse.js\"\n\"formatSearchResponse.js\"\n\"formatSearchResponseAtlas.js\"\n\"formatSemanticResponse.js\"\n\"formatSuggestionsResponse.js\"\n\"keywordSearchTemplateAtlas.json\"\n\"searchTemplate.json\"\n\"searchTemplateAtlas.json\"\n",
			err:            nil,
		},
		//Error Case
		{
			name:           "GetFile List HTTP Error",
			client:         new(mocks.FailingFileClient),
			expectedOutput: "",
			err: NewErrorWithCause(
				ErrorExitCode,
				discoveryPackage.Error{
					Status: http.StatusInternalServerError,
					Body: gjson.Parse(`{
	"status": 500,
	"code": 1003,
	"messages": [
		"Internal server error"
	],
	"timestamp": "2025-10-16T17:46:45.386963700Z"
}`,
					),
				},
				"Could not get file list",
			),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := strings.NewReader("")
			out := &bytes.Buffer{}
			errBuf := &bytes.Buffer{}

			ios := iostreams.IOStreams{
				In:  in,
				Out: out,
				Err: errBuf,
			}

			d := NewDiscovery(&ios, viper.New(), "")
			err := d.GetFileList(tc.client, nil)

			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, out.String())
			}
		})
	}
}

// Test_discovery_PingServer tests the discovery.PingServer() function.
func Test_discovery_PingServer(t *testing.T) {
	tests := []struct {
		name           string
		client         ServerPinger
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "StartSeed correctly prints the received object with the pretty printer",
			client:         new(mocks.WorkingServerPinger),
			printer:        nil,
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		{
			name:           "StartSeed correctly prints the received object with JSON ugly printer",
			client:         new(mocks.WorkingServerPinger),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"acknowledged\":true}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "GetEntityById fails",
			client:         new(mocks.FailingServerPingerServerNotFound),
			printer:        nil,
			expectedOutput: "",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: 21029da3-041c-43b5-a67e-870251f2f6a6"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
			}, "Could not get server ID."),
		},
		{
			name:           "Ping fails",
			client:         new(mocks.FailingServerPingerPingFailed),
			printer:        nil,
			expectedOutput: "",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnprocessableEntity, Body: gjson.Parse(`{
	"status": 422,
	"code": 4001,
	"messages": [
		"Client of type openai cannot be validated."
	],
	"timestamp": "2025-10-23T22:35:38.345647200Z"
	}`)}, "Could not ping server with id \"21029da3-041c-43b5-a67e-870251f2f6a6\""),
		},
		{
			name:      "Printing fails",
			client:    new(mocks.WorkingServerPinger),
			printer:   nil,
			outWriter: testutils.ErrWriter{Err: errors.New("write failed")},
			err:       NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
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

			d := NewDiscovery(&ios, viper.New(), "")
			err := d.PingServer(tc.client, "Mongo Atlas server", tc.printer)

			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}

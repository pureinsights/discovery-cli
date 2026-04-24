package file

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func GetCurrentWorkingDirectory(t *testing.T) string {
	wd, err := os.Getwd()
	require.NoError(t, err)

	return wd
}

func ReadFile(t *testing.T, path string) []byte {
	bytes, err := os.ReadFile(path)
	assert.NoError(t, err)

	return bytes
}

// TestNewGetCommand tests the NewGetCommand() function.
func TestNewDownloadCommand(t *testing.T) {
	const filePrefix = "NewDownloadCommand"

	tests := []struct {
		name         string
		args         []string
		url          bool
		filesToCheck []string
		apiKey       string
		outGolden    string
		errGolden    string
		outBytes     []byte
		errBytes     []byte
		responses    map[string]testutils.MockResponse
		err          error
	}{
		// Working case
		{
			name:         "Downloads a file",
			args:         []string{"script.py"},
			url:          true,
			filesToCheck: []string{"script.py"},
			apiKey:       "apiKey123",
			outGolden:    "NewDownloadCommand_Out_DownloadsFile",
			errGolden:    "NewDownloadCommand_Err_DownloadsFile",
			outBytes:     testutils.Read(t, "NewDownloadCommand_Out_DownloadsFile"),
			errBytes:     []byte(nil),
			responses: map[string]testutils.MockResponse{
				"GET:/v2/file/script.py": {
					StatusCode:  http.StatusOK,
					Body:        string(ReadFile(t, filepath.Join("testdata", filePrefix+"_script.py"))),
					ContentType: "application/octet-stream",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/file/script.py", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},
		{
			name:         "Download multiple files",
			args:         []string{"script.py", "elastictemplate.json"},
			url:          true,
			filesToCheck: []string{"script.py", "elastictemplate.json"},
			apiKey:       "apiKey123",
			outGolden:    "NewDownloadCommand_Out_DownloadMultipleFiles",
			errGolden:    "NewDownloadCommand_Err_DownloadMultipleFiles",
			outBytes:     testutils.Read(t, "NewDownloadCommand_Out_DownloadMultipleFiles"),
			errBytes:     []byte(nil),
			responses: map[string]testutils.MockResponse{
				"GET:/v2/file/script.py": {
					StatusCode:  http.StatusOK,
					Body:        string(ReadFile(t, filepath.Join("testdata", filePrefix+"_script.py"))),
					ContentType: "application/octet-stream",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/file/script.py", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"GET:/v2/file/elastictemplate.json": {
					StatusCode:  http.StatusOK,
					Body:        string(ReadFile(t, filepath.Join("testdata", filePrefix+"_elastictemplate.json"))),
					ContentType: "application/octet-stream",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/file/elastictemplate.json", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},
		{
			name:         "Download multiple files with Output",
			args:         []string{"script.py", "elastictemplate.json", "-o", "NewDownloadCommandTest"},
			url:          true,
			filesToCheck: []string{"NewDownloadCommandTest/script.py", "NewDownloadCommandTest/elastictemplate.json"},
			apiKey:       "apiKey123",
			outGolden:    "NewDownloadCommand_Out_DownloadMultipleFilesOutput",
			errGolden:    "NewDownloadCommand_Err_DownloadMultipleFilesOutput",
			outBytes:     testutils.Read(t, "NewDownloadCommand_Out_DownloadMultipleFilesOutput"),
			errBytes:     []byte(nil),
			responses: map[string]testutils.MockResponse{
				"GET:/v2/file/script.py": {
					StatusCode:  http.StatusOK,
					Body:        string(ReadFile(t, filepath.Join("testdata", filePrefix+"_script.py"))),
					ContentType: "application/octet-stream",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/file/script.py", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"GET:/v2/file/elastictemplate.json": {
					StatusCode:  http.StatusOK,
					Body:        string(ReadFile(t, filepath.Join("testdata", filePrefix+"_elastictemplate.json"))),
					ContentType: "application/octet-stream",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/file/elastictemplate.json", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},
		// Error Case
		{
			name:      "Fails to download a file",
			args:      []string{"script.py"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewDownloadCommand_Out_FailDownloadFile",
			errGolden: "NewDownloadCommand_Err_FailDownloadFile",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewDownloadCommand_Err_FailDownloadFile"),
			responses: map[string]testutils.MockResponse{
				"GET:/v2/file/script.py": {
					StatusCode:  http.StatusNotFound,
					Body:        "",
					ContentType: "",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/file/script.py", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(``)},
				"Could not get file with key \"script.py\"",
			),
		},
		{
			name:         "Fails to download a file when downloading multiple",
			args:         []string{"script.py", "elastictemplate.json"},
			url:          true,
			filesToCheck: []string{"script.py"},
			apiKey:       "apiKey123",
			outGolden:    "NewDownloadCommand_Out_FailDownloadMultipleFiles",
			errGolden:    "NewDownloadCommand_Err_FailDownloadMultipleFiles",
			outBytes:     []byte(nil),
			errBytes:     testutils.Read(t, "NewDownloadCommand_Err_FailDownloadMultipleFiles"),
			responses: map[string]testutils.MockResponse{
				"GET:/v2/file/script.py": {
					StatusCode:  http.StatusOK,
					Body:        string(ReadFile(t, filepath.Join("testdata", filePrefix+"_script.py"))),
					ContentType: "application/octet-stream",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/file/script.py", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"GET:/v2/file/elastictemplate.json": {
					StatusCode:  http.StatusNotFound,
					Body:        "",
					ContentType: "",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/file/elastictemplate.json", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(``)},
				"Could not get file with key \"elastictemplate.json\"",
			),
		},
		{
			name:      "No URL",
			args:      []string{"script.py"},
			url:       false,
			apiKey:    "apiKey123",
			outGolden: "NewDownloadCommand_Out_NoURL",
			errGolden: "NewDownloadCommand_Err_NoURL",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewDownloadCommand_Err_NoURL"),
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:      "GetFiles returns HTTP error",
			args:      []string{"script.py"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewDownloadCommand_Out_GetFilesHTTPError",
			errGolden: "NewDownloadCommand_Err_GetFilesHTTPError",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewDownloadCommand_Err_GetFilesHTTPError"),
			responses: map[string]testutils.MockResponse{
				"GET:/v2/file/script.py": {
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
						assert.Equal(t, "/v2/file/script.py", r.URL.Path)
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
			}`)}, "Could not get file with key \"script.py\""),
		},
	}
	testDataDir := filepath.Join(GetCurrentWorkingDirectory(t), "testdata")
	tmpDir := testutils.ChangeDirectoryHelper(t)
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

			getCmd := NewDownloadCommand(d)

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

			for _, path := range tc.filesToCheck {

				fileName := filepath.Base(path)
				fullPath := filepath.Join(tmpDir, path)

				assert.FileExists(t, fullPath)

				testdata := ReadFile(t, filepath.Join(testDataDir, filePrefix+"_"+fileName))
				downloaded := ReadFile(t, fullPath)

				assert.Equal(t, testdata, downloaded)
			}

			if tc.outBytes != nil {
				testutils.CompareBytes(t, tc.outGolden, tc.outBytes, out.Bytes())
			}
		})
	}
}

// TestNewGetCommand_NoProfileFlag tests the NewGetCommand when the profile flag was not defined.
func TestNewDownloadCommand_NoProfileFlag(t *testing.T) {
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

	getCmd := NewDownloadCommand(d)

	getCmd.SilenceUsage = true
	getCmd.SetIn(ios.In)
	getCmd.SetOut(ios.Out)
	getCmd.SetErr(ios.Err)

	getCmd.SetArgs([]string{"my_test_file.json"})

	err := getCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewDownloadCommand_Err_NoProfile", testutils.Read(t, "NewDownloadCommand_Err_NoProfile"), errBuf.Bytes())
}

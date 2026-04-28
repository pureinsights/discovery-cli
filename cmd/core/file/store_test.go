package file

import (
	"os"
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

// function used to get the exact error of trying to get the Stat() of a directory or file that doesn't exist.
func getFileAttributesExError(t *testing.T, path string) error {
	_, err := os.Stat(path)
	require.Error(t, err)
	return err
}

// TestNewStoreCommand tests the NewStoreCommand() function.
func TestNewStoreCommand(t *testing.T) {
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
			name:      "Store a single file",
			args:      []string{"./testdata/NewDownloadCommand_script.py"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewStoreCommand_Out_StoreSingleFile",
			errGolden: "NewStoreCommand_Err_StoreSingleFile",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_StoreSingleFile"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"PUT:/v2/file/NewDownloadCommand_script.py": {
					StatusCode: http.StatusOK,
					Body: `
					{
						"acknowledged": true
					}
					`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPut, r.Method)
						assert.Equal(t, "/v2/file/NewDownloadCommand_script.py", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},
		{
			name:      "Stores a directory",
			args:      []string{"./testdata/NewStoreCommand"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewStoreCommand_Out_StoreDirectory",
			errGolden: "NewStoreCommand_Err_StoreDirectory",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_StoreDirectory"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"PUT:/v2/file/script.py": {
					StatusCode: http.StatusOK,
					Body: `
					{
						"acknowledged": true
					}
					`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPut, r.Method)
						assert.Equal(t, "/v2/file/script.py", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"PUT:/v2/file/other_script.py": {
					StatusCode: http.StatusOK,
					Body: `
					{
						"acknowledged": true
					}
					`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPut, r.Method)
						assert.Equal(t, "/v2/file/other_script.py", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},
		{
			name:      "Stores an empty directory",
			args:      []string{t.TempDir()},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewStoreCommand_Out_StoreEmptyDirectory",
			errGolden: "NewStoreCommand_Err_StoreEmptyDirectory",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_StoreEmptyDirectory"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{},
			err:       nil,
		},
		{
			name:      "Stores a directory recursively",
			args:      []string{"./testdata/NewStoreCommand", "--recursive"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewStoreCommand_Out_StoreDirectoryRecursively",
			errGolden: "NewStoreCommand_Err_StoreDirectoryRecursively",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_StoreDirectoryRecursively"),
			errBytes:  []byte(nil),
			responses: map[string]testutils.MockResponse{
				"PUT:/v2/file/script.py": {
					StatusCode: http.StatusOK,
					Body: `
					{
						"acknowledged": true
					}
					`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPut, r.Method)
						assert.Equal(t, "/v2/file/script.py", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"PUT:/v2/file/other_script.py": {
					StatusCode: http.StatusOK,
					Body: `
					{
						"acknowledged": true
					}
					`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPut, r.Method)
						assert.Equal(t, "/v2/file/other_script.py", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
				"PUT:/v2/file/TestSubFolder/script.py": {
					StatusCode: http.StatusOK,
					Body: `
					{
						"acknowledged": true
					}
					`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPut, r.Method)
						assert.Equal(t, "/v2/file/TestSubFolder/script.py", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: nil,
		},
		// Error Case
		{
			name:      "No URL",
			args:      []string{"test"},
			url:       false,
			apiKey:    "apiKey123",
			outGolden: "NewStoreCommand_Out_NoURL",
			errGolden: "NewStoreCommand_Err_NoURL",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewStoreCommand_Err_NoURL"),
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:      "Upload returns HTTP error",
			args:      []string{"./testdata/NewDownloadCommand_script.py"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewStoreCommand_Out_UploadHTTPError",
			errGolden: "NewStoreCommand_Err_UploadHTTPError",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewStoreCommand_Err_UploadHTTPError"),
			responses: map[string]testutils.MockResponse{
				"PUT:/v2/file/NewDownloadCommand_script.py": {
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
						assert.Equal(t, http.MethodPut, r.Method)
						assert.Equal(t, "/v2/file/NewDownloadCommand_script.py", r.URL.Path)
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
			}`)}, "Could not store the file with path \"./testdata/NewDownloadCommand_script.py\""),
		},
		{
			name:      "Printing JSON array fails",
			args:      []string{"./testdata/NewDownloadCommand_script.py"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewStoreCommand_Out_PrintArrayFails",
			errGolden: "NewStoreCommand_Err_PrintArrayFails",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewStoreCommand_Err_PrintArrayFails"),
			responses: map[string]testutils.MockResponse{
				"PUT:/v2/file/NewDownloadCommand_script.py": {
					StatusCode: http.StatusOK,
					Body: `
					{
						acknowledged: true
					}
					`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPut, r.Method)
						assert.Equal(t, "/v2/file/NewDownloadCommand_script.py", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character 'a' looking for beginning of object key string"), "Could not print JSON object"),
		},
		{
			name:      "Invalid key to store file",
			args:      []string{"./testdata/NewDownloadCommand_script.py"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewStoreCommand_Out_InvalidKey",
			errGolden: "NewStoreCommand_Err_InvalidKey",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewStoreCommand_Err_InvalidKey"),
			responses: map[string]testutils.MockResponse{
				"PUT:/v2/file/NewDownloadCommand_script.py": {
					StatusCode: http.StatusBadRequest,
					Body: `{
						"status": 400,
						"code": 3002,
						"messages": [
							"key: Invalid format for file path, use only alphanumeric symbols with a limit of 255 characters and a max of 10 path levels."
						],
						"timestamp": "2025-10-16T17:46:45.386963700Z"
					}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPut, r.Method)
						assert.Equal(t, "/v2/file/NewDownloadCommand_script.py", r.URL.Path)
						assert.Equal(t, "apiKey123", r.Header.Get("X-API-Key"))
					},
				},
			},
			err: cli.NewErrorWithCause(
				cli.ErrorExitCode,
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
				"Could not store the file with path \"./testdata/NewDownloadCommand_script.py\"",
			),
		},
		{
			name:      "Invalid key to store file",
			args:      []string{"./testdata/NewDownloadCommand_this_file_does_not_exist.py"},
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewStoreCommand_Out_FileDoesNotExist",
			errGolden: "NewStoreCommand_Err_FileDoesNotExist",
			outBytes:  []byte(nil),
			errBytes:  testutils.Read(t, "NewStoreCommand_Err_FileDoesNotExist"),
			err: cli.NewErrorWithCause(
				cli.ErrorExitCode,
				getFileAttributesExError(t,"./testdata/NewDownloadCommand_this_file_does_not_exist.py"),
				"The path \"./testdata/NewDownloadCommand_this_file_does_not_exist.py\" does not exist",
			),
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

			getCmd := NewStoreCommand(d)

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

// TestNewStoreCommand_NoProfileFlag tests the NewStoreCommand when the profile flag was not defined.
func TestNewStoreCommand_NoProfileFlag(t *testing.T) {
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

	getCmd := NewStoreCommand(d)

	getCmd.SilenceUsage = true
	getCmd.SetIn(ios.In)
	getCmd.SetOut(ios.Out)
	getCmd.SetErr(ios.Err)

	getCmd.SetArgs([]string{"test"})

	err := getCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewStoreCommand_Err_NoProfile", testutils.Read(t, "NewStoreCommand_Err_NoProfile"), errBuf.Bytes())
}

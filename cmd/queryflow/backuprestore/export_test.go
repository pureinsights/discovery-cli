package backuprestore

import (
	"bytes"
	"errors"
	"fmt"
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

// TestNewExportCommand_ProfileFlag tests the NewExportCommand() function when there is a profile flag.
func TestNewExportCommand(t *testing.T) {
	zipBytes, _ := os.ReadFile("testdata/queryflow-export.zip")
	tests := []struct {
		name               string
		url                bool
		apiKey             string
		outGolden          string
		errGolden          string
		outBytes           []byte
		errBytes           []byte
		method             string
		path               string
		statusCode         int
		response           []byte
		contentDisposition string
		file               string
		err                error
		compareOptions     []testutils.CompareBytesOption
	}{
		// Working case
		{
			name:               "Export returns acknowledged true",
			url:                true,
			apiKey:             "",
			outGolden:          "NewExportCommand_Out_ExportReturnsTrue",
			errGolden:          "NewExportCommand_Err_ExportReturnsTrue",
			outBytes:           testutils.Read(t, "NewExportCommand_Out_ExportReturnsTrue"),
			errBytes:           []byte(nil),
			method:             http.MethodGet,
			path:               "/v2/export",
			statusCode:         http.StatusOK,
			response:           zipBytes,
			contentDisposition: `attachment; filename="export-20251110T1455.zip"; filename*=utf-8''Export-20251110T1460.zip`,
			file:               filepath.Join(t.TempDir(), "queryflow-export.zip"),
			err:                nil,
		},

		// Error case
		{
			name:      "No URL",
			outGolden: "NewExportCommand_Out_NoURL",
			errGolden: "NewExportCommand_Err_NoURL",
			outBytes:  testutils.Read(t, "NewExportCommand_Out_NoURL"),
			errBytes:  testutils.Read(t, "NewExportCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery QueryFlow URL is missing for profile \"default\".\nTo set the URL for the Discovery QueryFlow API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery queryflow config --profile \"default\""),
		},
		{
			name:               "Export Fails because the sent directory does not exist",
			url:                true,
			apiKey:             "",
			outGolden:          "NewExportCommand_Out_DirectoryDoesNotExist",
			errGolden:          "NewExportCommand_Err_DirectoryDoesNotExist",
			outBytes:           testutils.Read(t, "NewExportCommand_Out_DirectoryDoesNotExist"),
			errBytes:           testutils.Read(t, "NewExportCommand_Err_DirectoryDoesNotExist"),
			method:             http.MethodGet,
			path:               "/v2/export",
			statusCode:         http.StatusOK,
			response:           zipBytes,
			contentDisposition: `attachment; filename="export-20251110T1455.zip"; filename*=utf-8''Export-20251110T1460.zip`,
			file:               filepath.Join("doesnotexist", "queryflow-export.zip"),
			err:                cli.NewErrorWithCause(cli.ErrorExitCode, fmt.Errorf("the given path does not exist: %s", filepath.Join("doesnotexist", "queryflow-export.zip")), "Could not export entities"),
			compareOptions:     []testutils.CompareBytesOption{testutils.WithNormalizePaths()},
		},
		{
			name:               "Export fails",
			url:                true,
			apiKey:             "",
			outGolden:          "NewExportCommand_Out_ExportFails",
			errGolden:          "NewExportCommand_Err_ExportFails",
			outBytes:           testutils.Read(t, "NewExportCommand_Out_ExportFails"),
			errBytes:           testutils.Read(t, "NewExportCommand_Err_ExportFails"),
			method:             http.MethodGet,
			path:               "/v2/export",
			statusCode:         http.StatusUnauthorized,
			response:           []byte(`{"error":"unauthorized"}`),
			contentDisposition: `attachment; filename="export-20251110T1455.zip"; filename*=utf-8''Export-20251110T1460.zip`,
			file:               filepath.Join(t.TempDir(), "queryflow-export.zip"),
			err:                cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}, "Could not export entities"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.apiKey != "" {
					assert.Equal(t, tc.apiKey, r.Header.Get("X-API-KEY"))
				}
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, tc.path, r.URL.Path)
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Header().Set(
					"Content-Disposition",
					tc.contentDisposition,
				)
				w.WriteHeader(tc.statusCode)
				w.Write(tc.response)
			}))

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
			vpr.Set("output", "json")
			if tc.url {
				vpr.Set("default.queryflow_url", srv.URL)
			}
			if tc.apiKey != "" {
				vpr.Set("default.queryflow_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, t.TempDir())

			exportCmd := NewExportCommand(d)

			exportCmd.SetIn(ios.In)
			exportCmd.SetOut(ios.Out)
			exportCmd.SetErr(ios.Err)

			exportCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			args := []string{}
			if tc.file != "" {
				args = append(args, "--file")
				args = append(args, tc.file)
			}

			exportCmd.SetArgs(args)

			err := exportCmd.Execute()
			if tc.err != nil {
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
				testutils.CompareBytes(t, tc.errGolden, tc.errBytes, errBuf.Bytes(), tc.compareOptions...)
			} else {
				require.NoError(t, err)
				readBytes, err := os.ReadFile(tc.file)
				require.NoError(t, err)
				assert.Equal(t, readBytes, tc.response)
			}

			testutils.CompareBytes(t, tc.outGolden, tc.outBytes, out.Bytes())
		})
	}
}

// TestNewExportCommand_NoProfileFlag tests the NewExportCommand when the profile flag was not defined.
func TestNewExportCommand_NoProfileFlag(t *testing.T) {
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

	vpr.Set("default.queryflow_url", "test")
	vpr.Set("default.queryflow_key", "test")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	getCmd := NewExportCommand(d)

	getCmd.SetIn(ios.In)
	getCmd.SetOut(ios.Out)
	getCmd.SetErr(ios.Err)

	getCmd.SetArgs([]string{})

	err := getCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewExportCommand_Out_NoProfile", testutils.Read(t, "NewExportCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewExportCommand_Err_NoProfile", testutils.Read(t, "NewExportCommand_Err_NoProfile"), errBuf.Bytes())
}

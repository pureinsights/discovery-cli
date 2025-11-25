package backuprestore

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ExportResponse is used to store the response of the export endpoint in the test cases.
type ExportResponse struct {
	StatusCode         int
	ContentType        string
	Body               []byte
	FileName           string
	ContentDisposition string
}

// TestNewExportCommand_ProfileFlag tests the NewExportCommand() function when there is a profile flag.
func TestNewExportCommand(t *testing.T) {
	coreBytes, _ := os.ReadFile("testdata/core-export.zip")
	ingestionBytes, _ := os.ReadFile("testdata/ingestion-export.zip")
	queryflowBytes, _ := os.ReadFile("testdata/queryflow-export.zip")
	tests := []struct {
		name           string
		coreUrl        bool
		ingestionUrl   bool
		queryflowUrl   bool
		apiKey         string
		outGolden      string
		errGolden      string
		outBytes       []byte
		errBytes       []byte
		method         string
		path           string
		responses      map[string]ExportResponse
		file           string
		err            error
		compareOptions []testutils.CompareBytesOption
	}{
		// Working case
		{
			name:         "Export returns acknowledged true in all components",
			coreUrl:      true,
			ingestionUrl: true,
			queryflowUrl: true,
			apiKey:       "",
			outGolden:    "NewExportCommand_Out_ExportReturnsTrue",
			errGolden:    "NewExportCommand_Err_ExportReturnsTrue",
			outBytes:     testutils.Read(t, "NewExportCommand_Out_ExportReturnsTrue"),
			errBytes:     []byte(nil),
			method:       http.MethodGet,
			path:         "/v2/export",
			responses: map[string]ExportResponse{
				"core": {
					StatusCode:         http.StatusOK,
					ContentType:        "application/octet-stream",
					Body:               coreBytes,
					FileName:           "export-20251110T1455.zip",
					ContentDisposition: `attachment; filename="export-20251110T1455.zip"; filename*=utf-8'export-20251110T1455.zip`,
				},
				"ingestion": {
					StatusCode:         http.StatusOK,
					ContentType:        "application/octet-stream",
					Body:               ingestionBytes,
					FileName:           "export-20251110T1455.zip",
					ContentDisposition: `attachment; filename="export-20251110T1455.zip"; filename*=utf-8'export-20251110T1455.zip`,
				},
				"queryflow": {
					StatusCode:         http.StatusOK,
					ContentType:        "application/octet-stream",
					Body:               queryflowBytes,
					FileName:           "export-20251110T1455.zip",
					ContentDisposition: `attachment; filename="export-20251110T1455.zip"; filename*=utf-8'export-20251110T1455.zip`,
				},
			},
			file: filepath.Join(t.TempDir(), "discovery-export.zip"),
			err:  nil,
		},
		{
			name:         "Export returns acknowledged true in Core and QueryFlow but not Ingestion",
			coreUrl:      true,
			ingestionUrl: true,
			queryflowUrl: true,
			apiKey:       "",
			outGolden:    "NewExportCommand_Out_ExportReturnsTrueIngestionFails",
			errGolden:    "NewExportCommand_Err_ExportReturnsTrueIngestionFails",
			outBytes:     testutils.Read(t, "NewExportCommand_Out_ExportReturnsTrueIngestionFails"),
			errBytes:     []byte(nil),
			method:       http.MethodGet,
			path:         "/v2/export",
			responses: map[string]ExportResponse{
				"core": {
					StatusCode:         http.StatusOK,
					ContentType:        "application/octet-stream",
					Body:               coreBytes,
					FileName:           "export-20251110T1455.zip",
					ContentDisposition: `attachment; filename="export-20251110T1455.zip"; filename*=utf-8'export-20251110T1455.zip`,
				},
				"ingestion": {
					StatusCode:         http.StatusUnauthorized,
					ContentType:        "application/json",
					Body:               []byte(`{"error":"unauthorized"}`),
					FileName:           "export-20251110T1455.zip",
					ContentDisposition: `attachment; filename="export-20251110T1455.zip"; filename*=utf-8'export-20251110T1455.zip`,
				},
				"queryflow": {
					StatusCode:         http.StatusOK,
					ContentType:        "application/octet-stream",
					Body:               queryflowBytes,
					FileName:           "export-20251110T1455.zip",
					ContentDisposition: `attachment; filename="export-20251110T1455.zip"; filename*=utf-8'export-20251110T1455.zip`,
				},
			},
			file: filepath.Join(t.TempDir(), "discovery-export.zip"),
			err:  nil,
		},

		// Error case
		{
			name:         "No Core URL",
			outGolden:    "NewExportCommand_Out_NoCoreURL",
			errGolden:    "NewExportCommand_Err_NoCoreURL",
			outBytes:     testutils.Read(t, "NewExportCommand_Out_NoCoreURL"),
			errBytes:     testutils.Read(t, "NewExportCommand_Err_NoCoreURL"),
			coreUrl:      false,
			ingestionUrl: true,
			queryflowUrl: true,
			apiKey:       "apiKey123",
			err:          cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:         "No Ingestion URL",
			outGolden:    "NewExportCommand_Out_NoIngestionURL",
			errGolden:    "NewExportCommand_Err_NoIngestionURL",
			outBytes:     testutils.Read(t, "NewExportCommand_Out_NoIngestionURL"),
			errBytes:     testutils.Read(t, "NewExportCommand_Err_NoIngestionURL"),
			coreUrl:      true,
			ingestionUrl: false,
			queryflowUrl: true,
			apiKey:       "",
			err:          cli.NewError(cli.ErrorExitCode, "The Discovery Ingestion URL is missing for profile \"default\".\nTo set the URL for the Discovery Ingestion API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery ingestion config --profile \"default\""),
		},
		{
			name:         "No QueryFlow URL",
			outGolden:    "NewExportCommand_Out_NoQueryFlowURL",
			errGolden:    "NewExportCommand_Err_NoQueryFlowURL",
			outBytes:     testutils.Read(t, "NewExportCommand_Out_NoQueryFlowURL"),
			errBytes:     testutils.Read(t, "NewExportCommand_Err_NoQueryFlowURL"),
			coreUrl:      true,
			ingestionUrl: true,
			queryflowUrl: false,
			apiKey:       "apiKey123",
			err:          cli.NewError(cli.ErrorExitCode, "The Discovery QueryFlow URL is missing for profile \"default\".\nTo set the URL for the Discovery QueryFlow API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery queryflow config --profile \"default\""),
		},
		{
			name:         "Export Fails",
			coreUrl:      true,
			ingestionUrl: true,
			queryflowUrl: true,
			apiKey:       "",
			outGolden:    "NewExportCommand_Out_ExportFails",
			errGolden:    "NewExportCommand_Err_ExportFails",
			outBytes:     testutils.Read(t, "NewExportCommand_Out_ExportFails"),
			errBytes:     testutils.Read(t, "NewExportCommand_Err_ExportFails"),
			method:       http.MethodGet,
			path:         "/v2/export",
			responses: map[string]ExportResponse{
				"core": {
					StatusCode:         http.StatusOK,
					ContentType:        "application/octet-stream",
					Body:               coreBytes,
					FileName:           "export-20251110T1455.zip",
					ContentDisposition: `attachment; filename="export-20251110T1455.zip"; filename*=utf-8'export-20251110T1455.zip`,
				},
				"ingestion": {
					StatusCode:         http.StatusOK,
					ContentType:        "application/octet-stream",
					Body:               ingestionBytes,
					FileName:           "export-20251110T1455.zip",
					ContentDisposition: `attachment; filename="export-20251110T1455.zip"; filename*=utf-8'export-20251110T1455.zip`,
				},
				"queryflow": {
					StatusCode:         http.StatusOK,
					ContentType:        "application/octet-stream",
					Body:               queryflowBytes,
					FileName:           "export-20251110T1455.zip",
					ContentDisposition: `attachment; filename="export-20251110T1455.zip"; filename*=utf-8'export-20251110T1455.zip`,
				},
			},
			file:           filepath.Join("doesnotexist", "discovery-export.zip"),
			err:            cli.NewErrorWithCause(cli.ErrorExitCode, fmt.Errorf("the given path does not exist: %s", filepath.Join("doesnotexist", "discovery-export.zip")), "Could not export entities"),
			compareOptions: []testutils.CompareBytesOption{testutils.WithNormalizePaths()},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			coreResponse := tc.responses["core"]
			coreServer := httptest.NewServer(http.HandlerFunc(
				testutils.HttpHandlerWithContentDisposition(
					t,
					coreResponse.StatusCode,
					coreResponse.ContentType,
					coreResponse.ContentDisposition,
					coreResponse.Body,
					func(t *testing.T, r *http.Request) {
						if tc.apiKey != "" {
							assert.Equal(t, tc.apiKey, r.Header.Get("X-API-KEY"))
						}
						assert.Equal(t, tc.method, r.Method)
						assert.Equal(t, tc.path, r.URL.Path)
					},
				)))
			defer coreServer.Close()

			ingestionResponse := tc.responses["ingestion"]
			ingestionServer := httptest.NewServer(http.HandlerFunc(
				testutils.HttpHandlerWithContentDisposition(
					t,
					ingestionResponse.StatusCode,
					ingestionResponse.ContentType,
					ingestionResponse.ContentDisposition,
					ingestionResponse.Body,
					func(t *testing.T, r *http.Request) {
						if tc.apiKey != "" {
							assert.Equal(t, tc.apiKey, r.Header.Get("X-API-KEY"))
						}
						assert.Equal(t, tc.method, r.Method)
						assert.Equal(t, tc.path, r.URL.Path)
					},
				)))
			defer ingestionServer.Close()

			queryflowResponse := tc.responses["queryflow"]
			queryflowServer := httptest.NewServer(http.HandlerFunc(
				testutils.HttpHandlerWithContentDisposition(
					t,
					queryflowResponse.StatusCode,
					queryflowResponse.ContentType,
					queryflowResponse.ContentDisposition,
					queryflowResponse.Body,
					func(t *testing.T, r *http.Request) {
						if tc.apiKey != "" {
							assert.Equal(t, tc.apiKey, r.Header.Get("X-API-KEY"))
						}
						assert.Equal(t, tc.method, r.Method)
						assert.Equal(t, tc.path, r.URL.Path)
					},
				)))
			defer queryflowServer.Close()

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
			if tc.coreUrl {
				vpr.Set("default.core_url", coreServer.URL)
			}
			if tc.ingestionUrl {
				vpr.Set("default.ingestion_url", ingestionServer.URL)
			}
			if tc.queryflowUrl {
				vpr.Set("default.queryflow_url", queryflowServer.URL)
			}
			if tc.apiKey != "" {
				vpr.Set("default.ingestion_key", tc.apiKey)
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
				zipFile, err := os.ReadFile(tc.file)
				require.NoError(t, err)
				zipReader, err := zip.NewReader(bytes.NewReader(zipFile), int64(len(zipFile)))
				require.NoError(t, err)
				files := make(map[string]*zip.File, len(zipReader.File))
				for _, f := range zipReader.File {
					files[f.Name] = f
				}
				for component, response := range tc.responses {
					if response.StatusCode < http.StatusBadRequest {
						exportedFile, ok := files[fmt.Sprintf("%s-%s", component, response.FileName)]
						require.True(t, ok)
						fileContent, err := exportedFile.Open()
						require.NoError(t, err)
						gotBytes, err := io.ReadAll(fileContent)
						require.NoError(t, err)
						fileContent.Close()
						assert.Equal(t, response.Body, gotBytes)
					}
				}
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

	vpr.Set("default.ingestion_url", "test")
	vpr.Set("default.ingestion_key", "test")

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

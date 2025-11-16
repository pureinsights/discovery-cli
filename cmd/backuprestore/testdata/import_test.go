package backuprestore

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
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

type ImportResponse struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

// TestNewImportCommand_ProfileFlag tests the NewImportCommand() function when there is a profile flag.
func TestNewImportCommand_ProfileFlag(t *testing.T) {
	coreImport, _ := os.ReadFile("testdata/core-import.json")
	ingestionImport, _ := os.ReadFile("testdata/ingestion-import.json")
	queryflowImport, _ := os.ReadFile("testdata/queryflow-import.json")
	tests := []struct {
		name         string
		coreUrl      bool
		ingestionUrl bool
		queryflowUrl bool
		apiKey       string
		outGolden    string
		errGolden    string
		outBytes     []byte
		errBytes     []byte
		method       string
		path         string
		responses    map[string]ImportResponse
		file         string
		err          error
	}{
		// Working case
		{
			name:         "Import returns the import results of in all products",
			coreUrl:      true,
			ingestionUrl: true,
			queryflowUrl: true,
			apiKey:       "",
			outGolden:    "NewImportCommand_Out_ImportReturnsResults",
			errGolden:    "NewImportCommand_Err_ImportReturnsResults",
			outBytes:     testutils.Read(t, "NewImportCommand_Out_ImportReturnsResults"),
			errBytes:     []byte(nil),
			method:       http.MethodGet,
			path:         "/v2/import",
			responses: map[string]ImportResponse{
				"core": {
					StatusCode:         http.StatusOK,
					ContentType:        "application/octet-stream",
					Body:               coreBytes,
					FileName:           "import-20251110T1455.zip",
					ContentDisposition: `attachment; filename="import-20251110T1455.zip"; filename*=utf-8'import-20251110T1455.zip`,
				},
				"ingestion": {
					StatusCode:         http.StatusOK,
					ContentType:        "application/octet-stream",
					Body:               ingestionBytes,
					FileName:           "import-20251110T1455.zip",
					ContentDisposition: `attachment; filename="import-20251110T1455.zip"; filename*=utf-8'import-20251110T1455.zip`,
				},
				"queryflow": {
					StatusCode:         http.StatusOK,
					ContentType:        "application/octet-stream",
					Body:               queryflowBytes,
					FileName:           "import-20251110T1455.zip",
					ContentDisposition: `attachment; filename="import-20251110T1455.zip"; filename*=utf-8'import-20251110T1455.zip`,
				},
			},
			file: filepath.Join(t.TempDir(), "discovery-import.json"),
			err:  nil,
		},
		{
			name:         "Import returns acknowledged true in Core and QueryFlow but not Ingestion",
			coreUrl:      true,
			ingestionUrl: true,
			queryflowUrl: true,
			apiKey:       "",
			outGolden:    "NewImportCommand_Out_ImportReturnsTrueIngestionFails",
			errGolden:    "NewImportCommand_Err_ImportReturnsTrueIngestionFails",
			outBytes:     testutils.Read(t, "NewImportCommand_Out_ImportReturnsTrueIngestionFails"),
			errBytes:     []byte(nil),
			method:       http.MethodGet,
			path:         "/v2/import",
			responses: map[string]ImportResponse{
				"core": {
					StatusCode:         http.StatusOK,
					ContentType:        "application/octet-stream",
					Body:               coreBytes,
					FileName:           "import-20251110T1455.zip",
					ContentDisposition: `attachment; filename="import-20251110T1455.zip"; filename*=utf-8'import-20251110T1455.zip`,
				},
				"ingestion": {
					StatusCode:         http.StatusUnauthorized,
					ContentType:        "application/json",
					Body:               []byte(`{"error":"unauthorized"}`),
					FileName:           "import-20251110T1455.zip",
					ContentDisposition: `attachment; filename="import-20251110T1455.zip"; filename*=utf-8'import-20251110T1455.zip`,
				},
				"queryflow": {
					StatusCode:         http.StatusOK,
					ContentType:        "application/octet-stream",
					Body:               queryflowBytes,
					FileName:           "import-20251110T1455.zip",
					ContentDisposition: `attachment; filename="import-20251110T1455.zip"; filename*=utf-8'import-20251110T1455.zip`,
				},
			},
			file: filepath.Join(t.TempDir(), "discovery-import.json"),
			err:  nil,
		},

		// Error case
		{
			name:         "No Core URL",
			outGolden:    "NewImportCommand_Out_NoCoreURL",
			errGolden:    "NewImportCommand_Err_NoCoreURL",
			outBytes:     testutils.Read(t, "NewImportCommand_Out_NoCoreURL"),
			errBytes:     testutils.Read(t, "NewImportCommand_Err_NoCoreURL"),
			coreUrl:      false,
			ingestionUrl: true,
			queryflowUrl: true,
			apiKey:       "apiKey123",
			err:          cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:         "No Ingestion URL",
			outGolden:    "NewImportCommand_Out_NoIngestionURL",
			errGolden:    "NewImportCommand_Err_NoIngestionURL",
			outBytes:     testutils.Read(t, "NewImportCommand_Out_NoIngestionURL"),
			errBytes:     testutils.Read(t, "NewImportCommand_Err_NoIngestionURL"),
			coreUrl:      true,
			ingestionUrl: false,
			queryflowUrl: true,
			apiKey:       "",
			err:          cli.NewError(cli.ErrorExitCode, "The Discovery Ingestion URL is missing for profile \"default\".\nTo set the URL for the Discovery Ingestion API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery ingestion config --profile \"default\""),
		},
		{
			name:         "No QueryFlow URL",
			outGolden:    "NewImportCommand_Out_NoQueryFlowURL",
			errGolden:    "NewImportCommand_Err_NoQueryFlowURL",
			outBytes:     testutils.Read(t, "NewImportCommand_Out_NoQueryFlowURL"),
			errBytes:     testutils.Read(t, "NewImportCommand_Err_NoQueryFlowURL"),
			coreUrl:      true,
			ingestionUrl: true,
			queryflowUrl: false,
			apiKey:       "apiKey123",
			err:          cli.NewError(cli.ErrorExitCode, "The Discovery QueryFlow URL is missing for profile \"default\".\nTo set the URL for the Discovery QueryFlow API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery queryflow config --profile \"default\""),
		},
		{
			name:         "Import Fails",
			coreUrl:      true,
			ingestionUrl: true,
			queryflowUrl: true,
			apiKey:       "",
			outGolden:    "NewImportCommand_Out_ImportFails",
			errGolden:    "NewImportCommand_Err_ImportFails",
			outBytes:     testutils.Read(t, "NewImportCommand_Out_ImportFails"),
			errBytes:     []byte(nil),
			method:       http.MethodGet,
			path:         "/v2/import",
			responses: map[string]ImportResponse{
				"core": {
					StatusCode:         http.StatusOK,
					ContentType:        "application/octet-stream",
					Body:               coreBytes,
					FileName:           "import-20251110T1455.zip",
					ContentDisposition: `attachment; filename="import-20251110T1455.zip"; filename*=utf-8'import-20251110T1455.zip`,
				},
				"ingestion": {
					StatusCode:         http.StatusOK,
					ContentType:        "application/octet-stream",
					Body:               ingestionBytes,
					FileName:           "import-20251110T1455.zip",
					ContentDisposition: `attachment; filename="import-20251110T1455.zip"; filename*=utf-8'import-20251110T1455.zip`,
				},
				"queryflow": {
					StatusCode:         http.StatusOK,
					ContentType:        "application/octet-stream",
					Body:               queryflowBytes,
					FileName:           "import-20251110T1455.zip",
					ContentDisposition: `attachment; filename="import-20251110T1455.zip"; filename*=utf-8'import-20251110T1455.zip`,
				},
			},
			file: filepath.Join("doesnotexist", "discovery-import.json"),
			err:  cli.NewErrorWithCause(cli.ErrorExitCode, fs.ErrNotExist, "Could not import entities"),
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

			importCmd := NewImportCommand(d)

			importCmd.SetIn(ios.In)
			importCmd.SetOut(ios.Out)
			importCmd.SetErr(ios.Err)

			importCmd.PersistentFlags().StringP(
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

			importCmd.SetArgs(args)

			err := importCmd.Execute()
			if tc.err != nil {
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				cliError, _ := tc.err.(cli.Error)
				if !errors.Is(cliError.Cause, fs.ErrNotExist) {
					assert.EqualError(t, err, tc.err.Error())
					testutils.CompareBytes(t, tc.errGolden, tc.errBytes, errBuf.Bytes())
				} else {
					assert.Equal(t, cliError.ExitCode, errStruct.ExitCode)
					assert.Equal(t, cliError.Message, errStruct.Message)
				}
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
						importedFile, ok := files[fmt.Sprintf("%s-%s", component, response.FileName)]
						require.True(t, ok)
						fileContent, err := importedFile.Open()
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

// TestNewImportCommand_NoProfileFlag tests the NewImportCommand when the profile flag was not defined.
func TestNewImportCommand_NoProfileFlag(t *testing.T) {
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

	getCmd := NewImportCommand(d)

	getCmd.SetIn(ios.In)
	getCmd.SetOut(ios.Out)
	getCmd.SetErr(ios.Err)

	getCmd.SetArgs([]string{})

	err := getCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewImportCommand_Out_NoProfile", testutils.Read(t, "NewImportCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewImportCommand_Err_NoProfile", testutils.Read(t, "NewImportCommand_Err_NoProfile"), errBuf.Bytes())
}

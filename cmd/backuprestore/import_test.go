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

	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ImportResponse stores the response of the Import command to use in the test cases.
type ImportResponse struct {
	StatusCode int
	Body       []byte
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
			name:         "Import returns the import results of all products",
			coreUrl:      true,
			ingestionUrl: true,
			queryflowUrl: true,
			apiKey:       "",
			outGolden:    "NewImportCommand_Out_ImportReturnsResults",
			errGolden:    "NewImportCommand_Err_ImportReturnsResults",
			outBytes:     testutils.Read(t, "NewImportCommand_Out_ImportReturnsResults"),
			errBytes:     []byte(nil),
			method:       http.MethodPost,
			path:         "/v2/import",
			responses: map[string]ImportResponse{
				"core": {
					StatusCode: http.StatusMultiStatus,
					Body:       coreImport,
				},
				"ingestion": {
					StatusCode: http.StatusMultiStatus,
					Body:       ingestionImport,
				},
				"queryflow": {
					StatusCode: http.StatusMultiStatus,
					Body:       queryflowImport,
				},
			},
			file: filepath.Join("testdata", "discovery.zip"),
			err:  nil,
		},
		{
			name:         "Import returns the results of only Core and QueryFlow",
			coreUrl:      true,
			ingestionUrl: true,
			queryflowUrl: true,
			apiKey:       "",
			outGolden:    "NewImportCommand_Out_ImportReturnsCoreQueryFlow",
			errGolden:    "NewImportCommand_Err_ImportReturnsCoreQueryFlow",
			outBytes:     testutils.Read(t, "NewImportCommand_Out_ImportReturnsCoreQueryFlow"),
			errBytes:     []byte(nil),
			method:       http.MethodPost,
			path:         "/v2/import",
			responses: map[string]ImportResponse{
				"core": {
					StatusCode: http.StatusMultiStatus,
					Body:       coreImport,
				},
				"queryflow": {
					StatusCode: http.StatusMultiStatus,
					Body:       queryflowImport,
				},
			},
			file: filepath.Join("testdata", "OnlyCoreQueryFlow.zip"),
			err:  nil,
		},

		// Error case
		{
			name:         "Import Fails because the sent file does not exist",
			coreUrl:      true,
			ingestionUrl: true,
			queryflowUrl: true,
			apiKey:       "",
			outGolden:    "NewImportCommand_Out_FileDoesNotExist",
			errGolden:    "NewImportCommand_Err_FileDoesNotExist",
			outBytes:     testutils.Read(t, "NewImportCommand_Out_FileDoesNotExist"),
			errBytes:     testutils.Read(t, "NewImportCommand_Err_FileDoesNotExist"),
			method:       http.MethodPost,
			path:         "/v2/import",
			responses:    map[string]ImportResponse{},
			file:         filepath.Join("doesnotexist", "discovery-export.zip"),
			err:          cli.NewErrorWithCause(cli.ErrorExitCode, fmt.Errorf("file does not exist: %s", filepath.Join("doesnotexist", "discovery-export.zip")), "Could not open the file with the entities"),
		},
		{
			name:         "Import Fails because the sent file has four entries when it should have at most three.",
			coreUrl:      true,
			ingestionUrl: true,
			queryflowUrl: true,
			apiKey:       "",
			outGolden:    "NewImportCommand_Out_ZipHasFourFiles",
			errGolden:    "NewImportCommand_Err_ZipHasFourFiles",
			outBytes:     testutils.Read(t, "NewImportCommand_Out_ZipHasFourFiles"),
			errBytes:     testutils.Read(t, "NewImportCommand_Err_ZipHasFourFiles"),
			method:       http.MethodPost,
			path:         "/v2/import",
			responses:    map[string]ImportResponse{},
			file:         filepath.Join("testdata", "4-files.zip"),
			err:          cli.NewError(cli.ErrorExitCode, "The sent file should only contain the Core, Ingestion, or QueryFlow export files."),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var coreServer *httptest.Server
			var ingestionServer *httptest.Server
			var queryflowServer *httptest.Server

			coreResponse, coreOk := tc.responses["core"]
			if coreOk {
				coreServer = httptest.NewServer(
					testutils.HttpHandler(t, coreResponse.StatusCode, "application/json", string(coreResponse.Body), func(t *testing.T, r *http.Request) {
						assert.Equal(t, tc.method, r.Method)
						assert.Equal(t, tc.path, r.URL.Path)
					}))
				defer coreServer.Close()
			}

			ingestionResponse, ingestionOk := tc.responses["ingestion"]
			if ingestionOk {
				ingestionServer = httptest.NewServer(
					testutils.HttpHandler(t, ingestionResponse.StatusCode, "application/json", string(ingestionResponse.Body), func(t *testing.T, r *http.Request) {
						assert.Equal(t, tc.method, r.Method)
						assert.Equal(t, tc.path, r.URL.Path)
					}))
				defer ingestionServer.Close()
			}

			queryflowResponse, queryflowOk := tc.responses["queryflow"]
			if queryflowOk {
				queryflowServer = httptest.NewServer(
					testutils.HttpHandler(t, queryflowResponse.StatusCode, "application/json", string(queryflowResponse.Body), func(t *testing.T, r *http.Request) {
						assert.Equal(t, tc.method, r.Method)
						assert.Equal(t, tc.path, r.URL.Path)
					}))
				defer queryflowServer.Close()
			}

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
				url := "http://localhost:12010"
				if coreServer != nil {
					url = coreServer.URL
				}
				vpr.Set("default.core_url", url)
			}
			if tc.ingestionUrl {
				url := "http://localhost:12030"
				if ingestionServer != nil {
					url = ingestionServer.URL
				}
				vpr.Set("default.ingestion_url", url)
			}
			if tc.queryflowUrl {
				url := "http://localhost:12040"
				if queryflowOk {
					url = queryflowServer.URL
				}
				vpr.Set("default.queryflow_url", url)
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
			args = append(args, "--file")
			args = append(args, tc.file)

			args = append(args, "--on-conflict")
			args = append(args, string(discoveryPackage.OnConflictUpdate))

			importCmd.SetArgs(args)

			err := importCmd.Execute()
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

	getCmd.SetArgs([]string{"--file", "testdata/discovery.zip"})

	err := getCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewImportCommand_Out_NoProfile", testutils.Read(t, "NewImportCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewImportCommand_Err_NoProfile", testutils.Read(t, "NewImportCommand_Err_NoProfile"), errBuf.Bytes())
}

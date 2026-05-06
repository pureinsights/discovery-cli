package deploy

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

// DeployResponse stores the response of the Deploy command to use in the test cases.
type DeployResponse struct {
	StatusCode int
	Body       []byte
}

// TestNewDeployCommand_ProfileFlag tests the NewDeployCommand() function when there is a profile flag.
func TestNewDeployCommand_ProfileFlag(t *testing.T) {
	coreDeploy, err := os.ReadFile("testdata/core-import.json")
	require.NoError(t, err)
	ingestionDeploy, err := os.ReadFile("testdata/ingestion-import.json")
	require.NoError(t, err)
	queryflowDeploy, _ := os.ReadFile("testdata/queryflow-import.json")
	require.NoError(t, err)
	tests := []struct {
		name               string
		coreUrl            bool
		ingestionUrl       bool
		queryflowUrl       bool
		apiKey             string
		outGolden          string
		errGolden          string
		outBytes           []byte
		errBytes           []byte
		method             string
		path               string
		responses          map[string]DeployResponse
		fileUploadResponse string
		fileUploadStatus   int
		directoryPath      string
		err                error
		compareOptions     []testutils.CompareBytesOption
	}{
		// Working case
		{
			name:         "Deploy returns the deploy results of all products",
			coreUrl:      true,
			ingestionUrl: true,
			queryflowUrl: true,
			apiKey:       "",
			outGolden:    "NewDeployCommand_Out_DeployReturnsResults",
			errGolden:    "NewDeployCommand_Err_DeployReturnsResults",
			outBytes:     testutils.Read(t, "NewDeployCommand_Out_DeployReturnsResults"),
			errBytes:     []byte(nil),
			method:       http.MethodPost,
			path:         "/v2/import",
			responses: map[string]DeployResponse{
				"core": {
					StatusCode: http.StatusMultiStatus,
					Body:       coreDeploy,
				},
				"ingestion": {
					StatusCode: http.StatusMultiStatus,
					Body:       ingestionDeploy,
				},
				"queryflow": {
					StatusCode: http.StatusMultiStatus,
					Body:       queryflowDeploy,
				},
			},
			directoryPath: filepath.Join("..", "..", "internal", "cli", "testdata", "deploy"),
			err:           nil,
			fileUploadResponse: `{
  "acknowledged": true
}`,
			fileUploadStatus: http.StatusOK,
		},
		{
			name:         "Deploy returns the results of only Core and QueryFlow",
			coreUrl:      true,
			ingestionUrl: true,
			queryflowUrl: true,
			apiKey:       "",
			outGolden:    "NewDeployCommand_Out_DeployReturnsCoreQueryFlow",
			errGolden:    "NewDeployCommand_Err_DeployReturnsCoreQueryFlow",
			outBytes:     testutils.Read(t, "NewDeployCommand_Out_DeployReturnsCoreQueryFlow"),
			errBytes:     []byte(nil),
			method:       http.MethodPost,
			path:         "/v2/import",
			responses: map[string]DeployResponse{
				"core": {
					StatusCode: http.StatusMultiStatus,
					Body:       coreDeploy,
				},
				"queryflow": {
					StatusCode: http.StatusMultiStatus,
					Body:       queryflowDeploy,
				},
			},
			directoryPath: filepath.Join("..", "..", "internal", "cli", "testdata", "deploy_OnlyCoreQueryFlow"),
			err:           nil,
			fileUploadResponse: `{
  "acknowledged": true
}`,
			fileUploadStatus: http.StatusOK,
		},

		// Error case
		{
			name:           "Deploy Fails because the sent directoryPath does not exist",
			coreUrl:        true,
			ingestionUrl:   true,
			queryflowUrl:   true,
			apiKey:         "",
			outGolden:      "NewDeployCommand_Out_DirectoryPathDoesNotExist",
			errGolden:      "NewDeployCommand_Err_DirectoryPathDoesNotExist",
			outBytes:       []byte(nil),
			errBytes:       testutils.Read(t, "NewDeployCommand_Err_DirectoryPathDoesNotExist"),
			method:         http.MethodPost,
			path:           "/v2/import",
			responses:      map[string]DeployResponse{},
			directoryPath:  "doesnotexist",
			err:            cli.NewErrorWithCause(cli.ErrorExitCode, fmt.Errorf("file does not exist: %s", "doesnotexist"), "Could not open the \"doesnotexist\" directory"),
			compareOptions: []testutils.CompareBytesOption{testutils.WithNormalizePaths()},
			fileUploadResponse: `{
  "acknowledged": true
}`,
			fileUploadStatus: http.StatusOK,
		},
		{
			name:         "Deploy Fails because file upload fails",
			coreUrl:      true,
			ingestionUrl: true,
			queryflowUrl: true,
			apiKey:       "",
			outGolden:    "NewDeployCommand_Out_FileUploadFails",
			errGolden:    "NewDeployCommand_Err_FileUploadFails",
			outBytes:     []byte(nil),
			errBytes:     testutils.Read(t, "NewDeployCommand_Err_FileUploadFails"),
			method:       http.MethodPost,
			path:         "/v2/import",
			responses: map[string]DeployResponse{
				"core": {
					StatusCode: http.StatusMultiStatus,
					Body:       coreDeploy,
				},
				"ingestion": {
					StatusCode: http.StatusMultiStatus,
					Body:       ingestionDeploy,
				},
				"queryflow": {
					StatusCode: http.StatusMultiStatus,
					Body:       queryflowDeploy,
				},
			},
			directoryPath: filepath.Join("..", "..", "internal", "cli", "testdata", "deploy"),
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusInternalServerError, Body: gjson.Parse(`{
			"status": 500,
			"code": 1003,
			"messages": [
				"Internal server error"
			],
			"timestamp": "2025-10-16T17:46:45.386963700Z"
			}`)}, "Could not create the temporary zips to import entities"),
			fileUploadResponse: `{
			"status": 500,
			"code": 1003,
			"messages": [
				"Internal server error"
			],
			"timestamp": "2025-10-16T17:46:45.386963700Z"
			}`,
			fileUploadStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var coreServer *httptest.Server
			var ingestionServer *httptest.Server
			var queryflowServer *httptest.Server

			coreResponse, coreOk := tc.responses["core"]
			if coreOk {
				coreServer = httptest.NewServer(testutils.HttpMultiResponseHandler(t, map[string]testutils.MockResponse{
					"POST:/v2/import": {
						StatusCode:  http.StatusMultiStatus,
						Body:        string(coreResponse.Body),
						ContentType: "application/json",
						Assertions: func(t *testing.T, r *http.Request) {
							assert.Equal(t, http.MethodPost, r.Method)
							assert.Equal(t, "/v2/import", r.URL.Path)
						},
					},
					"PUT:/v2/file/text.txt": {
						StatusCode:  tc.fileUploadStatus,
						Body:        tc.fileUploadResponse,
						ContentType: "application/json",
						Assertions: func(t *testing.T, r *http.Request) {
							assert.Equal(t, http.MethodPut, r.Method)
							assert.Equal(t, "/v2/file/text.txt", r.URL.Path)
						},
					},
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
			vpr.Set("output", "pretty-json")
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

			deployCmd := NewDeployCommand(d)

			deployCmd.SilenceUsage = true
			deployCmd.SetIn(ios.In)
			deployCmd.SetOut(ios.Out)
			deployCmd.SetErr(ios.Err)

			deployCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			args := []string{}
			args = append(args, tc.directoryPath)

			deployCmd.SetArgs(args)

			err := deployCmd.Execute()
			if tc.err != nil {
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
				testutils.CompareBytes(t, tc.errGolden, tc.errBytes, errBuf.Bytes(), tc.compareOptions...)
			} else {
				require.NoError(t, err)
			}

			if tc.outBytes != nil {
				testutils.CompareBytes(t, tc.outGolden, tc.outBytes, out.Bytes())
			}
		})
	}
}

// TestNewDeployCommand_NoProfileFlag tests the NewDeployCommand when the profile flag was not defined.
func TestNewDeployCommand_NoProfileFlag(t *testing.T) {
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

	deployCmd := NewDeployCommand(d)
	deployCmd.SilenceUsage = true
	deployCmd.SetIn(ios.In)
	deployCmd.SetOut(ios.Out)
	deployCmd.SetErr(ios.Err)

	deployCmd.SetArgs([]string{"testdata/discovery.zip"})

	err := deployCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewDeployCommand_Err_NoProfile", testutils.Read(t, "NewDeployCommand_Err_NoProfile"), errBuf.Bytes())
}

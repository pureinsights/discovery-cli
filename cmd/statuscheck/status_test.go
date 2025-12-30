package statuscheck

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// StatusResponse stores the response of the Status command to use in the test cases.
type StatusResponse struct {
	StatusCode int
	Body       string
}

// TestNewStatusCommand_ProfileFlag tests the NewStatusCommand() function when there is a profile flag.
func TestNewStatusCommand_ProfileFlag(t *testing.T) {
	const statusUp = `{
    "status": "UP"
}`
	const statusDown = `{
    "status": "DOWN"
}`
	tests := []struct {
		name           string
		coreUrl        bool
		ingestionUrl   bool
		queryflowUrl   bool
		stagingUrl     bool
		outGolden      string
		errGolden      string
		outBytes       []byte
		errBytes       []byte
		method         string
		path           string
		responses      map[string]StatusResponse
		file           string
		err            error
		compareOptions []testutils.CompareBytesOption
	}{
		// Working case
		{
			name:         "Status returns the status results of all products",
			coreUrl:      true,
			ingestionUrl: true,
			queryflowUrl: true,
			outGolden:    "NewStatusCommand_Out_StatusReturnsResults",
			errGolden:    "NewStatusCommand_Err_StatusReturnsResults",
			outBytes:     testutils.Read(t, "NewStatusCommand_Out_StatusReturnsResults"),
			errBytes:     []byte(nil),
			method:       http.MethodGet,
			path:         "/health",
			responses: map[string]StatusResponse{
				"core": {
					StatusCode: http.StatusOK,
					Body:       statusUp,
				},
				"ingestion": {
					StatusCode: http.StatusOK,
					Body:       statusUp,
				},
				"queryflow": {
					StatusCode: http.StatusServiceUnavailable,
					Body:       statusDown,
				},
				"staging": {
					StatusCode: http.StatusOK,
					Body:       statusUp,
				},
			},
			file: filepath.Join("testdata", "discovery.zip"),
			err:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			coreResponse := tc.responses["core"]
			coreServer := httptest.NewServer(
				testutils.HttpHandler(t, coreResponse.StatusCode, "application/json", string(coreResponse.Body), func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, tc.path, r.URL.Path)
				}))
			defer coreServer.Close()

			ingestionResponse := tc.responses["ingestion"]
			ingestionServer := httptest.NewServer(
				testutils.HttpHandler(t, ingestionResponse.StatusCode, "application/json", string(ingestionResponse.Body), func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, tc.path, r.URL.Path)
				}))
			defer ingestionServer.Close()

			queryflowResponse := tc.responses["queryflow"]
			queryflowServer := httptest.NewServer(
				testutils.HttpHandler(t, queryflowResponse.StatusCode, "application/json", string(queryflowResponse.Body), func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, tc.path, r.URL.Path)
				}))
			defer queryflowServer.Close()

			stagingResponse := tc.responses["staging"]
			stagingServer := httptest.NewServer(
				testutils.HttpHandler(t, stagingResponse.StatusCode, "application/json", string(stagingResponse.Body), func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, tc.path, r.URL.Path)
				}))
			defer stagingServer.Close()

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

			vpr.Set("default.core_url", coreServer.URL)

			vpr.Set("default.ingestion_url", ingestionServer.URL)

			vpr.Set("default.queryflow_url", queryflowServer.URL)

			vpr.Set("default.staging_url", stagingServer.URL)

			d := cli.NewDiscovery(&ios, vpr, t.TempDir())

			statusCmd := NewStatusCommand(d)

			statusCmd.SetIn(ios.In)
			statusCmd.SetOut(ios.Out)
			statusCmd.SetErr(ios.Err)

			statusCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			statusCmd.SetArgs([]string{})

			err := statusCmd.Execute()
			if tc.err != nil {
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
				testutils.CompareBytes(t, tc.errGolden, tc.errBytes, errBuf.Bytes(), tc.compareOptions...)
			} else {
				require.NoError(t, err)
			}

			testutils.CompareBytes(t, tc.outGolden, tc.outBytes, out.Bytes())
		})
	}
}

// TestNewStatusCommand_NoProfileFlag tests the NewStatusCommand when the profile flag was not defined.
func TestNewStatusCommand_NoProfileFlag(t *testing.T) {
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

	getCmd := NewStatusCommand(d)

	getCmd.SetIn(ios.In)
	getCmd.SetOut(ios.Out)
	getCmd.SetErr(ios.Err)

	getCmd.SetArgs([]string{})

	err := getCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewStatusCommand_Out_NoProfile", testutils.Read(t, "NewStatusCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewStatusCommand_Err_NoProfile", testutils.Read(t, "NewStatusCommand_Err_NoProfile"), errBuf.Bytes())
}

package commands

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/pureinsights/discovery-cli/internal/testutils/mocks"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestSearchDumpCommand(t *testing.T) {
	tests := []struct {
		name            string
		client          cli.Searcher
		id              string
		url             string
		apiKey          string
		contentProvider func(string) cli.StagingContentController
		config          cli.DumpConfig
		expectedOutput  string
		outWriter       io.Writer
		err             error
	}{
		// Working case
		{
			name:   "DumpCommand correctly prints acknowledged",
			url:    "http://localhost:12020/v2",
			apiKey: "staging123",
			id:     "my-bucket",
			client: new(mocks.WorkingSearcher),
			contentProvider: func(name string) cli.StagingContentController {
				return new(mocks.WorkingContentController)
			},
			config: cli.DumpConfig{Size: intPtr(-1), File: filepath.Join(t.TempDir(), "my-bucket.zip")},
			expectedOutput: `{
  "acknowledged": true
}
`,
			err: nil,
		},
		// Error cases
		{
			name:   "DumpCommand returns error when entity not found",
			url:    "http://localhost:12020/v2",
			apiKey: "staging123",
			id:     "falseBucket",
			client: new(mocks.FailingSearcher),
			contentProvider: func(name string) cli.StagingContentController {
				return new(mocks.WorkingContentController)
			},
			config: cli.DumpConfig{},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusBadRequest,
				Body: gjson.Parse(`{
	"status": 400,
	"code": 3002,
	"messages": [
		"Invalid JSON: Unexpected end-of-input:"
	],
	"timestamp": "2025-10-17T17:43:52.817308100Z"
	}`),
			}, "Could not find bucket with name or id \"falseBucket\""),
		},
		{
			name:   "DumpCommand returns error when scroll fails",
			url:    "http://localhost:12020/v2",
			apiKey: "staging123",
			id:     "my-bucket",
			client: new(mocks.WorkingSearcher),
			contentProvider: func(name string) cli.StagingContentController {
				return new(mocks.FailingContentController)
			},
			config: cli.DumpConfig{Size: intPtr(-1)},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusInternalServerError,
				Body:   gjson.Parse(`{"status": 500, "code": 5000, "messages": ["Internal server error"]}`),
			}, "Could not scroll the bucket with name \"MongoDB Atlas server\"."),
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
			if tc.url != "" {
				vpr.Set("default.staging_url", tc.url)
			}
			if tc.apiKey != "" {
				vpr.Set("default.staging_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, t.TempDir())
			err := SearchDumpCommand(tc.id, d, tc.client, tc.contentProvider, tc.config, nil)

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

func intPtr(i int) *int { return &i }

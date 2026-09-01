package commands

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery/v2"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/pureinsights/discovery-cli/internal/testutils/mocks"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestSearchCountCommand tests the SearchCountCommand() function.
func TestSearchCountCommand(t *testing.T) {
	tests := []struct {
		name            string
		client          cli.Searcher
		id              string
		url             string
		apiKey          string
		contentProvider func(string) cli.StagingContentController
		filters         gjson.Result
		expectedOutput  string
		outWriter       io.Writer
		err             error
	}{
		// Working case
		{
			name:   "CountCommand correctly prints the total",
			url:    "http://localhost:12020/v2",
			apiKey: "staging123",
			id:     "my-bucket",
			client: new(mocks.WorkingSearcher),
			contentProvider: func(name string) cli.StagingContentController {
				return new(mocks.WorkingContentController)
			},
			filters: gjson.Parse(`{"equals": {"field": "correct_index", "value": "hd_index"}}`),
			expectedOutput: `{
  "total": 10
}
`,
			err: nil,
		},
		// Error cases
		{
			name:   "CountCommand returns error when entity not found",
			url:    "http://localhost:12020/v2",
			apiKey: "staging123",
			id:     "falseBucket",
			client: new(mocks.FailingSearcherWorkingGetter),
			contentProvider: func(name string) cli.StagingContentController {
				return new(mocks.WorkingContentController)
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body:   gjson.Parse("{\n\t\"status\": 404,\n\t\"code\": 1003,\n\t\"messages\": [\n\t\t\"Entity not found: entity with name \"falseBucket\" does not exist\"\n\t]\n}"),
			}, "Could not find bucket with name or id \"falseBucket\""),
		},
		{
			name:   "CountCommand returns error when count fails",
			url:    "http://localhost:12020/v2",
			apiKey: "staging123",
			id:     "my-bucket",
			client: new(mocks.WorkingSearcher),
			contentProvider: func(name string) cli.StagingContentController {
				return new(mocks.FailingContentController)
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusInternalServerError,
				Body:   gjson.Parse(`{"status": 500, "code": 5000, "messages": ["Internal server error"]}`),
			}, "Could not count the bucket with name \"MongoDB Atlas server\"."),
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
			err := SearchCountCommand(tc.id, d, tc.client, tc.contentProvider, tc.filters, nil)

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

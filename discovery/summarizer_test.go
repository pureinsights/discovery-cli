package discovery

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestSummarizer has table-driven tests to test the Summarize method.
func TestSummarizer(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		response   string
		testFunc   func(t *testing.T, s summarizer)
	}{
		// Working case
		{
			name:       "Summarizer returns results",
			method:     http.MethodGet,
			path:       "/summary",
			statusCode: http.StatusOK,
			response: `{
			"DONE": 8
			}`,
			testFunc: func(t *testing.T, s summarizer) {
				response, err := s.Summarize()
				require.NoError(t, err)
				assert.Equal(t, 8, int(response.Get("DONE").Int()))
			},
		},

		// Error case
		{
			name:       "Summary fails",
			method:     http.MethodGet,
			path:       "/summary",
			statusCode: http.StatusNotFound,
			response:   `{"messages": ["Seed execution not found: 9ababe08-0b74-4672-bb7c-e7a8225d6d4"]}`,
			testFunc: func(t *testing.T, s summarizer) {
				response, err := s.Summarize()
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusNotFound, []byte(`{"messages": ["Seed execution not found: 9ababe08-0b74-4672-bb7c-e7a8225d6d4"]}`)))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, tc.path, r.URL.Path)
				}))
			defer srv.Close()

			s := summarizer{client: newClient(srv.URL, "")}
			tc.testFunc(t, s)
		})
	}
}

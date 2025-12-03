package discovery

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestSummarizer has table-driven tests to test the summarizer.Summarize() method.
func Test_summarizer_Summarize(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		response         string
		expectedResponse gjson.Result
		err              error
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
			expectedResponse: gjson.Parse(`{
			"DONE": 8
			}`),
			err: nil,
		},
		{
			name:             "Summarizer returns no content",
			method:           http.MethodGet,
			path:             "/summary",
			statusCode:       http.StatusNoContent,
			response:         ``,
			expectedResponse: gjson.Parse(``),
			err:              nil,
		},

		// Error case
		{
			name:             "Summary fails",
			method:           http.MethodGet,
			path:             "/summary",
			statusCode:       http.StatusNotFound,
			response:         `{"messages": ["Seed execution not found: 9ababe08-0b74-4672-bb7c-e7a8225d6d4"]}`,
			expectedResponse: gjson.Result{},
			err:              Error{Status: http.StatusNotFound, Body: gjson.Parse(`{"messages": ["Seed execution not found: 9ababe08-0b74-4672-bb7c-e7a8225d6d4"]}`)},
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
			response, err := s.Summarize()
			assert.Equal(t, tc.expectedResponse, response)
			if tc.err == nil {
				require.NoError(t, err)
			} else {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

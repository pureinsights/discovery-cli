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

// Test_newQueryFlowProcessorsClient test the queryFlowProcessorsClient's constructor
func Test_newContentClient(t *testing.T) {
	url := "http://localhost:8081/v2"
	apiKey := "Api Key"
	bucketName := "test-bucket"
	c := newContentClient(url, apiKey, bucketName)

	assert.Equal(t, apiKey, c.client.ApiKey)
	assert.Equal(t, url+"/content/"+bucketName, c.client.client.BaseURL)
}

func Test_newBucketsClient(t *testing.T) {
	url := "http://localhost:8081/v2"
	apiKey := "Api Key"
	c := newBucketsClient(url, apiKey)

	assert.Equal(t, apiKey, c.client.ApiKey)
	assert.Equal(t, url+"/bucket", c.client.client.BaseURL)
}

func Test_bucketsClient_Create(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		response   string
		testFunc   func(t *testing.T, response gjson.Result, err error)
	}{
		// Working case
		{
			name:       "Create works",
			method:     http.MethodPost,
			path:       "/",
			statusCode: http.StatusCreated,
			response:   `{"id":"5f125024-1e5e-4591-9fee-365dc20eeeed","name":"new-secret"}`,
			testFunc: func(t *testing.T, c crud) {
				config := gjson.Parse(`{"name":"new-secret"}`)
				response, err := c.Create(config)
				require.NoError(t, err)
				assert.Equal(t, "5f125024-1e5e-4591-9fee-365dc20eeeed", response.Get("id").String())
			},
		},

		// Error case
		{
			name:       "Create returns 403 Forbidden",
			method:     http.MethodPost,
			path:       "/",
			statusCode: http.StatusForbidden,
			response:   `{"error":"forbidden"}`,
			testFunc: func(t *testing.T, c crud) {
				config := gjson.Parse(`{"name":"new-secret"}`)
				response, err := c.Create(config)
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusForbidden, []byte(`{"error":"forbidden"}`)))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, tc.path, r.URL.Path)
			}))

			defer srv.Close()

			c := crud{getter{newClient(srv.URL, "")}}
			tc.testFunc(t, c)
		})
	}
}

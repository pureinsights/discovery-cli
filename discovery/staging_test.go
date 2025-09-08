package discovery

import (
	"fmt"
	"io"
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
		err        bool
		testFunc   func(t *testing.T, response gjson.Result, err error)
	}{
		// Working case
		{
			name:       "Create works",
			method:     http.MethodPost,
			path:       "/test-bucket",
			statusCode: http.StatusCreated,
			response:   `{"acknowledged": true}`,
			err:        false,
		},

		// Error case
		{
			name:       "Create returns 409 Conflict",
			method:     http.MethodPost,
			path:       "/test-bucket",
			statusCode: http.StatusConflict,
			response:   `{"acknowledged":false}`,
			err:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := `{
			"indices": [
				{
					"name": "indexTest",
					"fields": [{"author": "ASC"}],
					"unique": false
				}
			],
			"config": {}
			}`
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, "/bucket"+tc.path, r.URL.Path)
				body, _ := io.ReadAll(r.Body)
				json := gjson.Parse(string(body))
				assert.Equal(t, "indexTest", json.Get("indices.0.name").String())
				assert.Equal(t, "ASC", json.Get("indices.0.fields.0.author").String())
				assert.False(t, json.Get("indices.0.unique").Bool())
			}))

			defer srv.Close()

			bucketsClient := newBucketsClient(srv.URL, "")
			jsonBody := gjson.Parse(config)

			response, err := bucketsClient.Create("test-bucket", jsonBody)
			if !(tc.err) {
				require.NoError(t, err)
				assert.True(t, response.Get("acknowledged").Bool())
			} else {
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", tc.statusCode, tc.response))
			}
		})
	}
}

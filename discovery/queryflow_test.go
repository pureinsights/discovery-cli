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
func Test_newQueryFlowProcessorsClient(t *testing.T) {
	url := "http://localhost:8088/v2"
	apiKey := "Api Key"
	qpc := newQueryFlowProcessorsClient(url, apiKey)

	assert.Equal(t, apiKey, qpc.crud.client.ApiKey)
	assert.Equal(t, url+"/processor", qpc.crud.client.client.BaseURL)
	assert.Equal(t, apiKey, qpc.cloner.client.ApiKey)
	assert.Equal(t, url+"/processor", qpc.cloner.client.client.BaseURL)
}

// Test_newEndpointsClient tests the constructor of endpointsClients.
func Test_newEndpointsClient(t *testing.T) {
	url := "http://localhost:8088/v2"
	apiKey := "Api Key"
	qec := newEndpointsClient(url, apiKey)

	assert.Equal(t, apiKey, qec.crud.client.ApiKey)
	assert.Equal(t, url+"/endpoint", qec.crud.client.client.BaseURL)
	assert.Equal(t, apiKey, qec.cloner.client.ApiKey)
	assert.Equal(t, url+"/endpoint", qec.cloner.client.client.BaseURL)
	assert.Equal(t, apiKey, qec.enabler.client.ApiKey)
	assert.Equal(t, url+"/endpoint", qec.enabler.client.client.BaseURL)
}

// Test_queryFlow_Processors tests the queryFlow.Processors() function
func Test_queryFlow_Processors(t *testing.T) {
	q := NewQueryFlow("http://localhost:8080/v2", "Api Key")
	qpc := q.Processors()

	assert.Equal(t, q.ApiKey, qpc.crud.client.ApiKey)
	assert.Equal(t, q.Url+"/processor", qpc.crud.client.client.BaseURL)
	assert.Equal(t, q.ApiKey, qpc.cloner.client.ApiKey)
	assert.Equal(t, q.Url+"/processor", qpc.cloner.client.client.BaseURL)
}

// Test_queryFlow_Endpoints tests the queryFlow.Endpoints() function
func Test_queryFlow_Endpoints(t *testing.T) {
	q := NewQueryFlow("http://localhost:8080/v2", "Api Key")
	qec := q.Endpoints()

	assert.Equal(t, q.ApiKey, qec.crud.client.ApiKey)
	assert.Equal(t, q.Url+"/endpoint", qec.crud.client.client.BaseURL)
	assert.Equal(t, q.ApiKey, qec.cloner.client.ApiKey)
	assert.Equal(t, q.Url+"/endpoint", qec.cloner.client.client.BaseURL)
	assert.Equal(t, q.ApiKey, qec.enabler.client.ApiKey)
	assert.Equal(t, q.Url+"/endpoint", qec.enabler.client.client.BaseURL)
}

// Test_queryFlow_BackupRestore tests the core.BackupRestore() function
func Test_queryFlow_BackupRestore(t *testing.T) {
	q := NewQueryFlow("http://localhost:8088/v2", "Api Key")
	bc := q.BackupRestore()

	assert.Equal(t, q.ApiKey, bc.ApiKey)
	assert.Equal(t, q.Url, bc.client.client.BaseURL)
}

func Test_queryFlow_Invoke(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		response   string
		testFunc   func(t *testing.T, response string, err error)
	}{
		// Working case
		{
			name:       "Invoke returns a real response",
			method:     http.MethodGet,
			path:       "/api/blogs-search",
			statusCode: http.StatusOK,
			response: `[
				{
					"_id": "5625c64483bef0d48e9ad91aca9b2f94",
					"link": "https://pureinsights.com/blog/2024/pureinsights-named-mongodbs-2024-ai-partner-of-the-year/",
					"author": "Graham Gillen",
					"header": "Pureinsights Named MongoDB's 2024 AI Partner of the Year - Pureinsights: PRESS RELEASE - Pureinsights named MongoDB's Service AI Partner of the Year for 2024 and also joins the MongoDB AI Application Program (MAAP)."
				}
			]`,
			testFunc: func(t *testing.T, response string, err error) {
				json := gjson.Parse(response).Array()
				require.NoError(t, err)
				assert.Equal(t, "Graham Gillen", json[0].Get("author").String())
			},
		},

		// Error case
		{
			name:       "Invoking an endpoint returns an error",
			method:     http.MethodGet,
			path:       "/api/endpoint-false",
			statusCode: http.StatusNotFound,
			response: `{
				"status": 404,
				"code": 1001,
				"messages": [
					"The requested endpoint was not found or is inactive"
				],
				"timestamp": "2025-09-01T22:54:37.580046500Z"
			}`,
			testFunc: func(t *testing.T, response string, err error) {
				assert.Empty(t, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusNotFound, `{
				"status": 404,
				"code": 1001,
				"messages": [
					"The requested endpoint was not found or is inactive"
				],
				"timestamp": "2025-09-01T22:54:37.580046500Z"
			}`))
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

			c := newClient(srv.URL, "")
			serverClient := newServersClient(c.client.BaseURL, c.ApiKey)
			tc.testFunc(t, serverClient)
		})
	}
}

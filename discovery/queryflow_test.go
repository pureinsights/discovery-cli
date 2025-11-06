package discovery

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Test_newQueryFlowProcessorsClient test the queryFlowProcessorsClient's constructor
func Test_newQueryFlowProcessorsClient(t *testing.T) {
	url := "http://localhost:12040"
	apiKey := "Api Key"
	qpc := newQueryFlowProcessorsClient(url, apiKey)

	assert.Equal(t, apiKey, qpc.crud.client.ApiKey)
	assert.Equal(t, url+"/processor", qpc.crud.client.client.BaseURL)
	assert.Equal(t, apiKey, qpc.cloner.client.ApiKey)
	assert.Equal(t, url+"/processor", qpc.cloner.client.client.BaseURL)
	assert.Equal(t, apiKey, qpc.searcher.client.ApiKey)
	assert.Equal(t, url+"/processor", qpc.searcher.client.client.BaseURL)
}

// Test_newEndpointsClient tests the constructor of endpointsClients.
func Test_newEndpointsClient(t *testing.T) {
	url := "http://localhost:12040"
	apiKey := "Api Key"
	qec := newEndpointsClient(url, apiKey)

	assert.Equal(t, apiKey, qec.crud.client.ApiKey)
	assert.Equal(t, url+"/endpoint", qec.crud.client.client.BaseURL)
	assert.Equal(t, apiKey, qec.cloner.client.ApiKey)
	assert.Equal(t, url+"/endpoint", qec.cloner.client.client.BaseURL)
	assert.Equal(t, apiKey, qec.enabler.client.ApiKey)
	assert.Equal(t, url+"/endpoint", qec.enabler.client.client.BaseURL)
	assert.Equal(t, apiKey, qec.searcher.client.ApiKey)
	assert.Equal(t, url+"/endpoint", qec.searcher.client.client.BaseURL)
}

// Test_queryFlow_Processors tests the queryFlow.Processors() function
func Test_queryFlow_Processors(t *testing.T) {
	q := NewQueryFlow("http://localhost:12040", "Api Key")
	qpc := q.Processors()

	assert.Equal(t, q.ApiKey, qpc.crud.client.ApiKey)
	assert.Equal(t, q.Url+"/processor", qpc.crud.client.client.BaseURL)
	assert.Equal(t, q.ApiKey, qpc.cloner.client.ApiKey)
	assert.Equal(t, q.Url+"/processor", qpc.cloner.client.client.BaseURL)
	assert.Equal(t, q.ApiKey, qpc.searcher.client.ApiKey)
}

// Test_queryFlow_Endpoints tests the queryFlow.Endpoints() function
func Test_queryFlow_Endpoints(t *testing.T) {
	q := NewQueryFlow("http://localhost:12040", "Api Key")
	qec := q.Endpoints()

	assert.Equal(t, q.ApiKey, qec.crud.client.ApiKey)
	assert.Equal(t, q.Url+"/endpoint", qec.crud.client.client.BaseURL)
	assert.Equal(t, q.ApiKey, qec.cloner.client.ApiKey)
	assert.Equal(t, q.Url+"/endpoint", qec.cloner.client.client.BaseURL)
	assert.Equal(t, q.ApiKey, qec.enabler.client.ApiKey)
	assert.Equal(t, q.Url+"/endpoint", qec.enabler.client.client.BaseURL)
	assert.Equal(t, q.ApiKey, qec.searcher.client.ApiKey)
	assert.Equal(t, q.Url+"/endpoint", qec.searcher.client.client.BaseURL)
}

// Test_queryFlow_BackupRestore tests the queryFlow.BackupRestore() function
func Test_queryFlow_BackupRestore(t *testing.T) {
	q := NewQueryFlow("http://localhost:12040", "Api Key")
	bc := q.BackupRestore()

	assert.Equal(t, q.ApiKey, bc.ApiKey)
	assert.Equal(t, q.Url, bc.client.client.BaseURL)
}

// Test_queryFlow_Invoke tests the queryFlow.Invoke() function with different endpoint responses.
func Test_queryFlow_Invoke(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		response         string
		body             string
		queryParams      map[string][]string
		expectedResponse gjson.Result
		expectedPath     string
		err              error
	}{
		// Working case
		{
			name:         "Invoke returns a real response",
			method:       http.MethodGet,
			path:         "/blogs-search",
			expectedPath: "/v2/api/blogs-search",
			statusCode:   http.StatusOK,
			queryParams: map[string][]string{
				"q": {"Google"},
			},
			response: `[
				{
					"_id": "5625c64483bef0d48e9ad91aca9b2f94",
					"link": "https://pureinsights.com/blog/2024/pureinsights-named-mongodbs-2024-ai-partner-of-the-year/",
					"author": "Graham Gillen",
					"header": "Pureinsights Named MongoDB's 2024 AI Partner of the Year - Pureinsights: PRESS RELEASE - Pureinsights named MongoDB's Service AI Partner of the Year for 2024 and also joins the MongoDB AI Application Program (MAAP)."
				}
			]`,
			expectedResponse: gjson.Parse(`[
				{
					"_id": "5625c64483bef0d48e9ad91aca9b2f94",
					"link": "https://pureinsights.com/blog/2024/pureinsights-named-mongodbs-2024-ai-partner-of-the-year/",
					"author": "Graham Gillen",
					"header": "Pureinsights Named MongoDB's 2024 AI Partner of the Year - Pureinsights: PRESS RELEASE - Pureinsights named MongoDB's Service AI Partner of the Year for 2024 and also joins the MongoDB AI Application Program (MAAP)."
				}
			]`),
			err: nil,
		},
		{
			name:             "Invoke returns an empty array",
			method:           http.MethodGet,
			path:             "blogs-search",
			expectedPath:     "/v2/api/blogs-search",
			statusCode:       http.StatusOK,
			response:         `[]`,
			expectedResponse: gjson.Parse(`[]`),
			err:              nil,
		},
		{
			name:         "Invoke has a JSON Body",
			method:       http.MethodPost,
			path:         "/blogs-search",
			expectedPath: "/v2/api/blogs-search",
			statusCode:   http.StatusCreated,
			body: `{
				"_id": "5625c64483bef0d48e9ad91aca9b2f94",
				"link": "https://pureinsights.com/blog/2024/pureinsights-named-mongodbs-2024-ai-partner-of-the-year/",
				"author": "Graham Gillen",
				"header": "Pureinsights Named MongoDB's 2024 AI Partner of the Year - Pureinsights: PRESS RELEASE - Pureinsights named MongoDB's Service AI Partner of the Year for 2024 and also joins the MongoDB AI Application Program (MAAP)."
			}`,
			response:         `{"acknowledged": true}`,
			expectedResponse: gjson.Parse(`{"acknowledged": true}`),
			err:              nil,
		},
		{
			name:         "Invoke sends query parameters",
			method:       http.MethodDelete,
			path:         "/blogs-search",
			expectedPath: "/v2/api/blogs-search",
			statusCode:   http.StatusOK,
			queryParams: map[string][]string{
				"author": {"Graham Gillen"},
			},
			response:         `{"acknowledged": true}`,
			expectedResponse: gjson.Parse(`{"acknowledged": true}`),
			err:              nil,
		},
		{
			name:         "Invoke sends a body and query parameters",
			method:       http.MethodPut,
			path:         "/blogs-search",
			expectedPath: "/v2/api/blogs-search",
			statusCode:   http.StatusMultiStatus,
			queryParams: map[string][]string{
				"author": {"Graham Gillen"},
			},
			body: `{
				"_id": "5625c64483bef0d48e9ad91aca9b2f94",
				"link": "https://pureinsights.com/blog/2024/pureinsights-named-mongodbs-2024-ai-partner-of-the-year/",
				"author": "Graham Gillen",
				"header": "Pureinsights Named MongoDB's 2024 AI Partner of the Year - Pureinsights: PRESS RELEASE - Pureinsights named MongoDB's Service AI Partner of the Year for 2024 and also joins the MongoDB AI Application Program (MAAP)."
			}`,
			response: `[
				{
					"5625c64483bef0d48e9ad91aca9b2f94": 204,
					"04f8787b-c352-4c0c-aee4-6b537d873007": 204,
					"c26dc995-2ecf-4f6f-b78e-93a79e3c07a8": 204,
					"fb627ac3-e3c7-4452-aac3-3349c6492d60": 204
				}
			]`,
			expectedResponse: gjson.Parse(`[
				{
					"5625c64483bef0d48e9ad91aca9b2f94": 204,
					"04f8787b-c352-4c0c-aee4-6b537d873007": 204,
					"c26dc995-2ecf-4f6f-b78e-93a79e3c07a8": 204,
					"fb627ac3-e3c7-4452-aac3-3349c6492d60": 204
				}
			]`),
			err: nil,
		},

		// Error case
		{
			name:         "Invoking an endpoint returns an error",
			method:       http.MethodGet,
			path:         "///endpoint-false",
			expectedPath: "/v2/api///endpoint-false",
			statusCode:   http.StatusNotFound,
			response: `{
				"status": 404,
				"code": 1001,
				"messages": [
					"The requested endpoint was not found or is inactive"
				],
				"timestamp": "2025-09-01T22:54:37.580046500Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
				"status": 404,
				"code": 1001,
				"messages": [
					"The requested endpoint was not found or is inactive"
				],
				"timestamp": "2025-09-01T22:54:37.580046500Z"
			}`)},
		},
		{
			name:         "Invoking an endpoint returns unprocessable entity",
			method:       http.MethodGet,
			path:         "/blogs-search",
			expectedPath: "/v2/api/blogs-search",
			statusCode:   http.StatusUnprocessableEntity,
			response: `{
				"status": 422,
				"code": 4001,
				"messages": [
					"Timed out after 30000 ms while waiting for a server that matches ReadPreferenceServerSelector{readPreference=primary}. Client view of cluster state is {type=REPLICA_SET, servers=[{address=cluster0-shard-00-00.dleud.mongodb.net:27017, type=UNKNOWN, state=CONNECTING, exception={com.mongodb.MongoSocketWriteException: Exception sending message}, caused by {javax.net.ssl.SSLException: Received fatal alert: internal_error}}, {address=cluster0-shard-00-01.dleud.mongodb.net:27017, type=UNKNOWN, state=CONNECTING, exception={com.mongodb.MongoSocketWriteException: Exception sending message}, caused by {javax.net.ssl.SSLException: Received fatal alert: internal_error}}, {address=cluster0-shard-00-02.dleud.mongodb.net:27017, type=UNKNOWN, state=CONNECTING, exception={com.mongodb.MongoSocketWriteException: Exception sending message}, caused by {javax.net.ssl.SSLException: Received fatal alert: internal_error}}]"
				],
				"timestamp": "2025-09-01T22:59:38.272579400Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusUnprocessableEntity, Body: gjson.Parse(`{
				"status": 422,
				"code": 4001,
				"messages": [
					"Timed out after 30000 ms while waiting for a server that matches ReadPreferenceServerSelector{readPreference=primary}. Client view of cluster state is {type=REPLICA_SET, servers=[{address=cluster0-shard-00-00.dleud.mongodb.net:27017, type=UNKNOWN, state=CONNECTING, exception={com.mongodb.MongoSocketWriteException: Exception sending message}, caused by {javax.net.ssl.SSLException: Received fatal alert: internal_error}}, {address=cluster0-shard-00-01.dleud.mongodb.net:27017, type=UNKNOWN, state=CONNECTING, exception={com.mongodb.MongoSocketWriteException: Exception sending message}, caused by {javax.net.ssl.SSLException: Received fatal alert: internal_error}}, {address=cluster0-shard-00-02.dleud.mongodb.net:27017, type=UNKNOWN, state=CONNECTING, exception={com.mongodb.MongoSocketWriteException: Exception sending message}, caused by {javax.net.ssl.SSLException: Received fatal alert: internal_error}}]"
				],
				"timestamp": "2025-09-01T22:59:38.272579400Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, tc.expectedPath, r.URL.Path)
					if tc.body != "" {
						body, _ := io.ReadAll(r.Body)
						assert.Equal(t, tc.body, string(body))
					}

					if tc.queryParams != nil {
						for k, v := range tc.queryParams {
							assert.Equal(t, v, r.URL.Query()[k])
						}
					}
				}))
			defer srv.Close()

			q := NewQueryFlow(srv.URL, "")

			requestOptions := []RequestOption{}
			if tc.body != "" {
				requestOptions = append(requestOptions, WithJSONBody(tc.body))
			}

			if tc.queryParams != nil {
				requestOptions = append(requestOptions, WithQueryParameters(tc.queryParams))
			}
			response, err := q.Invoke(tc.method, tc.path, requestOptions...)
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

// Test_queryFlow_Debug tests the queryFlow.Debug() function with different endpoint responses.
func Test_queryFlow_Debug(t *testing.T) {
	realResponse, err := os.ReadFile(filepath.Join("testdata", "queryFlow_Debug_RealResponse.golden"))
	require.NoError(t, err)
	tests := []struct {
		name             string
		method           string
		path             string
		expectedPath     string
		statusCode       int
		response         string
		body             string
		queryParams      map[string][]string
		expectedResponse gjson.Result
		err              error
	}{
		// Working case
		{
			name:             "Debug returns a real response",
			method:           http.MethodGet,
			path:             "/blogs-search",
			expectedPath:     "/v2/debug/blogs-search",
			statusCode:       http.StatusOK,
			response:         string(realResponse),
			expectedResponse: gjson.ParseBytes(realResponse),
			err:              nil,
		},
		{
			name:         "Debug returns an empty array",
			method:       http.MethodGet,
			path:         "/blogs-search",
			expectedPath: "/v2/debug/blogs-search",
			statusCode:   http.StatusOK,
			response: `{
				"duration": 2147,
				"execution": [
					{
						"state": "searchState",
						"result": [
							{
								"processor": "5f125024-1e5e-4591-9fee-365dc20eeeed",
								"output": {
									"mongo": []
								},
								"duration": 2132
							}
						]
					}
				]
			}`,
			expectedResponse: gjson.Parse(`{
				"duration": 2147,
				"execution": [
					{
						"state": "searchState",
						"result": [
							{
								"processor": "5f125024-1e5e-4591-9fee-365dc20eeeed",
								"output": {
									"mongo": []
								},
								"duration": 2132
							}
						]
					}
				]
			}`),
			err: nil,
		},
		{
			name:         "Debug has a JSON Body",
			method:       http.MethodPost,
			path:         "/blogs-search",
			expectedPath: "/v2/debug/blogs-search",
			statusCode:   http.StatusCreated,
			body: `{
				"_id": "5625c64483bef0d48e9ad91aca9b2f94",
				"link": "https://pureinsights.com/blog/2024/pureinsights-named-mongodbs-2024-ai-partner-of-the-year/",
				"author": "Graham Gillen",
				"header": "Pureinsights Named MongoDB's 2024 AI Partner of the Year - Pureinsights: PRESS RELEASE - Pureinsights named MongoDB's Service AI Partner of the Year for 2024 and also joins the MongoDB AI Application Program (MAAP)."
			}`,
			response:         string(realResponse),
			expectedResponse: gjson.ParseBytes(realResponse),
			err:              nil,
		},
		{
			name:         "Debug sends query parameters",
			method:       http.MethodDelete,
			path:         "/blogs-search",
			expectedPath: "/v2/debug/blogs-search",
			statusCode:   http.StatusOK,
			queryParams: map[string][]string{
				"author": {"Graham Gillen"},
			},
			response:         `{"acknowledged": true}`,
			expectedResponse: gjson.Parse(`{"acknowledged": true}`),
			err:              nil,
		},
		{
			name:         "Debug sends a body and query parameters",
			method:       http.MethodPut,
			path:         "/blogs-search",
			expectedPath: "/v2/debug/blogs-search",
			statusCode:   http.StatusMultiStatus,
			queryParams: map[string][]string{
				"author": {"Graham Gillen"},
			},
			body: `{
				"_id": "5625c64483bef0d48e9ad91aca9b2f94",
				"link": "https://pureinsights.com/blog/2024/pureinsights-named-mongodbs-2024-ai-partner-of-the-year/",
				"author": "Graham Gillen",
				"header": "Pureinsights Named MongoDB's 2024 AI Partner of the Year - Pureinsights: PRESS RELEASE - Pureinsights named MongoDB's Service AI Partner of the Year for 2024 and also joins the MongoDB AI Application Program (MAAP)."
			}`,
			response:         string(realResponse),
			expectedResponse: gjson.ParseBytes(realResponse),
			err:              nil,
		},

		// Error case
		{
			name:         "Debugging an endpoint returns an error",
			method:       http.MethodGet,
			path:         "///endpoint-false",
			expectedPath: "/v2/debug///endpoint-false",
			statusCode:   http.StatusNotFound,
			response: `{
				"status": 404,
				"code": 1001,
				"messages": [
					"The requested endpoint was not found or is inactive"
				],
				"timestamp": "2025-09-01T22:54:37.580046500Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
				"status": 404,
				"code": 1001,
				"messages": [
					"The requested endpoint was not found or is inactive"
				],
				"timestamp": "2025-09-01T22:54:37.580046500Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, tc.expectedPath, r.URL.Path)
					if tc.body != "" {
						body, _ := io.ReadAll(r.Body)
						assert.Equal(t, tc.body, string(body))
					}

					if tc.queryParams != nil {
						for k, v := range tc.queryParams {
							assert.Equal(t, v, r.URL.Query()[k])
						}
					}
				}))
			defer srv.Close()

			q := NewQueryFlow(srv.URL, "")
			requestOptions := []RequestOption{}
			if tc.body != "" {
				requestOptions = append(requestOptions, WithJSONBody(tc.body))
			}

			if tc.queryParams != nil {
				requestOptions = append(requestOptions, WithQueryParameters(tc.queryParams))
			}
			response, err := q.Debug(tc.method, tc.path, requestOptions...)
			assert.Equal(t, tc.expectedResponse, response)
			if tc.err == nil {
				require.NoError(t, err)
				assert.True(t, response.IsObject())
			} else {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

// Test_NewQueryFlow tests the QueryFlow constructor.
func Test_NewQueryFlow(t *testing.T) {
	i := NewQueryFlow("http://localhost:12040", "Api Key")

	assert.Equal(t, "http://localhost:12040/v2", i.Url)
	assert.Equal(t, "Api Key", i.ApiKey)
}

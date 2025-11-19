package discovery

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Test_newSeedExecutionsClient test the constructor of seedExecutionsClient
func Test_newSeedExecutionsClient(t *testing.T) {
	url := "http://localhost:12030"
	apiKey := "Api Key"
	seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
	require.NoError(t, err)
	ingestionSeedsClient := newSeedsClient(url, apiKey)
	ingestionSeedExecutionsClient := newSeedExecutionsClient(ingestionSeedsClient, seedId)

	assert.Equal(t, apiKey, ingestionSeedExecutionsClient.getter.client.ApiKey)
	assert.Equal(t, url+"/seed/"+seedId.String()+"/execution", ingestionSeedExecutionsClient.getter.client.client.BaseURL)
}

// Test_newSeedRecordsClient tests the constructor of seedRecordsClient.
func Test_newSeedRecordsClient(t *testing.T) {
	url := "http://localhost:12030"
	apiKey := "Api Key"
	seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
	require.NoError(t, err)
	ingestionSeedsClient := newSeedsClient(url, apiKey)
	ingestionSeedRecordsClient := newSeedRecordsClient(ingestionSeedsClient, seedId)

	assert.Equal(t, apiKey, ingestionSeedRecordsClient.summarizer.client.ApiKey)
	assert.Equal(t, url+"/seed/"+seedId.String()+"/record", ingestionSeedRecordsClient.summarizer.client.client.BaseURL)
}

// Test_newSeedExecutionRecordsClient tests the constructor of seedExecutionRecordsClient
func Test_newSeedExecutionRecordsClient(t *testing.T) {
	url := "http://localhost:12030"
	apiKey := "Api Key"
	seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
	require.NoError(t, err)
	ingestionSeedsClient := newSeedsClient(url, apiKey)
	ingestionSeedExecutionsClient := newSeedExecutionsClient(ingestionSeedsClient, seedId)

	executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
	require.NoError(t, err)
	ingestionSeedExecutionRecordsClient := newSeedExecutionRecordsClient(ingestionSeedExecutionsClient, executionId)

	assert.Equal(t, apiKey, ingestionSeedExecutionRecordsClient.client.ApiKey)
	assert.Equal(t, url+"/seed/"+seedId.String()+"/execution/"+executionId.String()+"/record", ingestionSeedExecutionRecordsClient.client.client.BaseURL)
}

// Test_newSeedExecutionJobsClient tests the constructor of seedExecutionJobsClient.
func Test_newSeedExecutionJobsClient(t *testing.T) {
	url := "http://localhost:12030"
	apiKey := "Api Key"
	seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
	require.NoError(t, err)
	ingestionSeedsClient := newSeedsClient(url, apiKey)
	ingestionSeedExecutionsClient := newSeedExecutionsClient(ingestionSeedsClient, seedId)

	executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
	require.NoError(t, err)
	ingestionSeedExecutionJobClient := newSeedExecutionJobsClient(ingestionSeedExecutionsClient, executionId)

	assert.Equal(t, apiKey, ingestionSeedExecutionJobClient.client.ApiKey)
	assert.Equal(t, url+"/seed/"+seedId.String()+"/execution/"+executionId.String()+"/job", ingestionSeedExecutionJobClient.client.client.BaseURL)
}

// Test_newIngestionProcessorsClient tests the constructor of ingestionProcessorsClient.
func Test_newIngestionProcessorsClient(t *testing.T) {
	url := "http://localhost:12030"
	apiKey := "Api Key"
	processorsClient := newIngestionProcessorsClient(url, apiKey)

	assert.Equal(t, apiKey, processorsClient.crud.getter.client.ApiKey)
	assert.Equal(t, url+"/processor", processorsClient.crud.getter.client.client.BaseURL)
	assert.Equal(t, apiKey, processorsClient.cloner.client.ApiKey)
	assert.Equal(t, url+"/processor", processorsClient.cloner.client.client.BaseURL)
	assert.Equal(t, apiKey, processorsClient.searcher.client.ApiKey)
	assert.Equal(t, url+"/processor", processorsClient.searcher.client.client.BaseURL)
}

// Test_newPipelinesClient tests the constructor of pipelinesClient.
func Test_newPipelinesClient(t *testing.T) {
	url := "http://localhost:12030"
	apiKey := "Api Key"
	pipelineClient := newPipelinesClient(url, apiKey)

	assert.Equal(t, apiKey, pipelineClient.crud.getter.client.ApiKey)
	assert.Equal(t, url+"/pipeline", pipelineClient.crud.getter.client.client.BaseURL)
	assert.Equal(t, apiKey, pipelineClient.cloner.client.ApiKey)
	assert.Equal(t, url+"/pipeline", pipelineClient.cloner.client.client.BaseURL)
	assert.Equal(t, apiKey, pipelineClient.searcher.client.ApiKey)
	assert.Equal(t, url+"/pipeline", pipelineClient.searcher.client.client.BaseURL)
}

// Test_newSeedsClient tests the constructor of seedsClient.
func Test_newSeedsClient(t *testing.T) {
	url := "http://localhost:12030"
	apiKey := "Api Key"
	seedClient := newSeedsClient(url, apiKey)

	assert.Equal(t, apiKey, seedClient.crud.getter.client.ApiKey)
	assert.Equal(t, url+"/seed", seedClient.crud.getter.client.client.BaseURL)
	assert.Equal(t, apiKey, seedClient.cloner.client.ApiKey)
	assert.Equal(t, url+"/seed", seedClient.cloner.client.client.BaseURL)
}

// Test_seedExecutionsClient_Seed tests the seedExecutionsClient.Seed() function
func Test_seedExecutionsClient_Seed(t *testing.T) {
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
			name:             "Seed returns a real response",
			method:           http.MethodGet,
			path:             "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/seed",
			statusCode:       http.StatusOK,
			response:         `{"id":"2acd0a61-852c-4f38-af2b-9c84e152873e","name":"Search seed","type":"staging","active":true,"config":{"action":"scroll","bucket":"blogs"},"labels":[],"pipeline":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","recordPolicy":{"errorPolicy":"FATAL","timeoutPolicy":{"slice":"PT1H"},"outboundPolicy":{"idPolicy":{},"batchPolicy":{"maxCount":25,"flushAfter":"PT1M"}}},"creationTimestamp":"2025-08-21T21:52:03Z","lastUpdatedTimestamp":"2025-08-21T21:52:03Z"}`,
			expectedResponse: gjson.Parse(`{"id":"2acd0a61-852c-4f38-af2b-9c84e152873e","name":"Search seed","type":"staging","active":true,"config":{"action":"scroll","bucket":"blogs"},"labels":[],"pipeline":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","recordPolicy":{"errorPolicy":"FATAL","timeoutPolicy":{"slice":"PT1H"},"outboundPolicy":{"idPolicy":{},"batchPolicy":{"maxCount":25,"flushAfter":"PT1M"}}},"creationTimestamp":"2025-08-21T21:52:03Z","lastUpdatedTimestamp":"2025-08-21T21:52:03Z"}`),
			err:              nil,
		},
		// Error cases
		{
			name:       "Seed config returns execution not found",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/seed",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1003,
			"messages": [
				"Seed execution not found: 6b7f0b69-126f-49ab-b2ff-0a876f42e5ed"
			],
			"timestamp": "2025-09-03T17:44:01.557816Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Seed execution not found: 6b7f0b69-126f-49ab-b2ff-0a876f42e5ed"
			],
			"timestamp": "2025-09-03T17:44:01.557816Z"
			}`)},
		},
		{
			name:       "Seed config returns bad request",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/seed",
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [executionId] for value [notseed] due to: Invalid UUID string: notseed"
			],
			"timestamp": "2025-10-01T19:21:49.817539200Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [executionId] for value [notseed] due to: Invalid UUID string: notseed"
			],
			"timestamp": "2025-10-01T19:21:49.817539200Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, "/seed/2acd0a61-852c-4f38-af2b-9c84e152873e/execution/"+strings.TrimLeft(tc.path, "/"), r.URL.Path)
				}))
			defer srv.Close()

			apiKey := "Api Key"
			seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
			require.NoError(t, err)
			ingestionSeedsClient := newSeedsClient(srv.URL, apiKey)
			ingestionSeedExecutionsClient := newSeedExecutionsClient(ingestionSeedsClient, seedId)

			executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
			require.NoError(t, err)
			response, err := ingestionSeedExecutionsClient.Seed(executionId)
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

// Test_seedExecutionsClient_Pipeline tests the seedExecutionsClient.Pipeline() function
func Test_seedExecutionsClient_Pipeline(t *testing.T) {
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
			name:             "Pipeline returns a real response",
			method:           http.MethodGet,
			path:             "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/pipeline/9a74bf3a-eb2a-4334-b803-c92bf1bc45fe",
			statusCode:       http.StatusOK,
			response:         `{"id":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","name":"Search pipeline","active":true,"labels":[],"states":{"ingestionState":{"type":"processor","processors":[{"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","active":true,"outputField":"header"},{"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8","active":true}]}},"initialState":"ingestionState","recordPolicy":{"idPolicy":{},"errorPolicy":"FAIL","retryPolicy":{"active":true,"maxRetries":3},"timeoutPolicy":{"record":"PT1M"},"outboundPolicy":{"batchPolicy":{"maxCount":25,"flushAfter":"PT1M"}}},"creationTimestamp":"2025-08-21T21:52:02Z","lastUpdatedTimestamp":"2025-08-21T21:52:02Z"}`,
			expectedResponse: gjson.Parse(`{"id":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","name":"Search pipeline","active":true,"labels":[],"states":{"ingestionState":{"type":"processor","processors":[{"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","active":true,"outputField":"header"},{"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8","active":true}]}},"initialState":"ingestionState","recordPolicy":{"idPolicy":{},"errorPolicy":"FAIL","retryPolicy":{"active":true,"maxRetries":3},"timeoutPolicy":{"record":"PT1M"},"outboundPolicy":{"batchPolicy":{"maxCount":25,"flushAfter":"PT1M"}}},"creationTimestamp":"2025-08-21T21:52:02Z","lastUpdatedTimestamp":"2025-08-21T21:52:02Z"}`),
			err:              nil,
		},
		// Error case
		{
			name:       "pipeline config returns not found",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/pipeline/9a74bf3a-eb2a-4334-b803-c92bf1bc45fe",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1003,
			"messages": [
				"Pipeline not found: 9a74bf3a-eb2a-4334-b803-c92bf1bc45fe"
			],
			"timestamp": "2025-09-03T17:44:01.557816Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Pipeline not found: 9a74bf3a-eb2a-4334-b803-c92bf1bc45fe"
			],
			"timestamp": "2025-09-03T17:44:01.557816Z"
			}`)},
		},
		{
			name:       "pipeline config returns bad request",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/pipeline/9a74bf3a-eb2a-4334-b803-c92bf1bc45fe",
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [executionId] for value [-] due to: Invalid UUID string: -"
			],
			"timestamp": "2025-10-01T19:26:07.794214300Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [executionId] for value [-] due to: Invalid UUID string: -"
			],
			"timestamp": "2025-10-01T19:26:07.794214300Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, "/seed/2acd0a61-852c-4f38-af2b-9c84e152873e/execution/"+strings.TrimLeft(tc.path, "/"), r.URL.Path)
				}))
			defer srv.Close()

			apiKey := "Api Key"
			seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
			require.NoError(t, err)
			ingestionSeedsClient := newSeedsClient(srv.URL, apiKey)
			ingestionSeedExecutionsClient := newSeedExecutionsClient(ingestionSeedsClient, seedId)

			executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
			require.NoError(t, err)

			pipelineId, err := uuid.Parse("9a74bf3a-eb2a-4334-b803-c92bf1bc45fe")
			require.NoError(t, err)
			response, err := ingestionSeedExecutionsClient.Pipeline(executionId, pipelineId)
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

// Test_seedExecutionsClient_Processor tests the seedExecutionsClient.Processor() function.
func Test_seedExecutionsClient_Processor(t *testing.T) {
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
			name:             "Processor returns a real response",
			method:           http.MethodGet,
			path:             "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/processor/aa0186f1-746f-4b20-b1b0-313bd79e78b8",
			statusCode:       http.StatusOK,
			response:         `{"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8","name":"MongoDB store processor","type":"mongo","active":true,"config":{"data":{"link":"#{ data('/reference') }","author":"#{ data('/author') }","header":"#{ data('/header') }"},"action":"hydrate","database":"pureinsights","collection":"blogs"},"labels":[],"server":{"id":"f6950327-3175-4a98-a570-658df852424a","credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c"},"creationTimestamp":"2025-08-21T21:52:02Z","lastUpdatedTimestamp":"2025-08-21T21:52:02Z"}`,
			expectedResponse: gjson.Parse(`{"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8","name":"MongoDB store processor","type":"mongo","active":true,"config":{"data":{"link":"#{ data('/reference') }","author":"#{ data('/author') }","header":"#{ data('/header') }"},"action":"hydrate","database":"pureinsights","collection":"blogs"},"labels":[],"server":{"id":"f6950327-3175-4a98-a570-658df852424a","credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c"},"creationTimestamp":"2025-08-21T21:52:02Z","lastUpdatedTimestamp":"2025-08-21T21:52:02Z"}`),
			err:              nil,
		},
		// Error case
		{
			name:       "processor config returns not found",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/processor/aa0186f1-746f-4b20-b1b0-313bd79e78b8",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1003,
			"messages": [
				"Processor not found: aa0186f1-746f-4b20-b1b0-313bd79e78b8"
			],
			"timestamp": "2025-09-03T17:44:01.557816Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Processor not found: aa0186f1-746f-4b20-b1b0-313bd79e78b8"
			],
			"timestamp": "2025-09-03T17:44:01.557816Z"
			}`)},
		},
		{
			name:       "processor config returns bad request",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/processor/aa0186f1-746f-4b20-b1b0-313bd79e78b8",
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [executionId] for value [-] due to: Invalid UUID string: -"
			],
			"timestamp": "2025-10-01T19:26:07.794214300Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [executionId] for value [-] due to: Invalid UUID string: -"
			],
			"timestamp": "2025-10-01T19:26:07.794214300Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, "/seed/2acd0a61-852c-4f38-af2b-9c84e152873e/execution/"+strings.TrimLeft(tc.path, "/"), r.URL.Path)
				}))
			defer srv.Close()

			apiKey := "Api Key"
			seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
			require.NoError(t, err)
			ingestionSeedsClient := newSeedsClient(srv.URL, apiKey)
			ingestionSeedExecutionsClient := newSeedExecutionsClient(ingestionSeedsClient, seedId)

			executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
			require.NoError(t, err)

			processorId, err := uuid.Parse("aa0186f1-746f-4b20-b1b0-313bd79e78b8")
			require.NoError(t, err)
			response, err := ingestionSeedExecutionsClient.Processor(executionId, processorId)
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

// Test_seedExecutionsClient_Server tests the seedExecutionsClient.Server() function.
func Test_seedExecutionsClient_Server(t *testing.T) {
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
			name:             "Server returns a real response",
			method:           http.MethodGet,
			path:             "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/server/f6950327-3175-4a98-a570-658df852424a",
			statusCode:       http.StatusOK,
			response:         `{"id":"f6950327-3175-4a98-a570-658df852424a","name":"MongoDB store server","type":"mongo","active":true,"config":{"data":{"link":"#{ data('/reference') }","author":"#{ data('/author') }","header":"#{ data('/header') }"},"action":"hydrate","database":"pureinsights","collection":"blogs"},"labels":[],"server":{"id":"f6950327-3175-4a98-a570-658df852424a","credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c"},"creationTimestamp":"2025-08-21T21:52:02Z","lastUpdatedTimestamp":"2025-08-21T21:52:02Z"}`,
			expectedResponse: gjson.Parse(`{"id":"f6950327-3175-4a98-a570-658df852424a","name":"MongoDB store server","type":"mongo","active":true,"config":{"data":{"link":"#{ data('/reference') }","author":"#{ data('/author') }","header":"#{ data('/header') }"},"action":"hydrate","database":"pureinsights","collection":"blogs"},"labels":[],"server":{"id":"f6950327-3175-4a98-a570-658df852424a","credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c"},"creationTimestamp":"2025-08-21T21:52:02Z","lastUpdatedTimestamp":"2025-08-21T21:52:02Z"}`),
			err:              nil,
		},
		// Error case
		{
			name:       "server config returns not found",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/server/f6950327-3175-4a98-a570-658df852424a",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: f6950327-3175-4a98-a570-658df852424a"
			],
			"timestamp": "2025-09-03T17:44:01.557816Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: f6950327-3175-4a98-a570-658df852424a"
			],
			"timestamp": "2025-09-03T17:44:01.557816Z"
			}`)},
		},
		{
			name:       "server config returns bad request",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/server/f6950327-3175-4a98-a570-658df852424a",
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [executionId] for value [-] due to: Invalid UUID string: -"
			],
			"timestamp": "2025-10-01T19:26:07.794214300Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [executionId] for value [-] due to: Invalid UUID string: -"
			],
			"timestamp": "2025-10-01T19:26:07.794214300Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, "/seed/2acd0a61-852c-4f38-af2b-9c84e152873e/execution/"+strings.TrimLeft(tc.path, "/"), r.URL.Path)
				}))
			defer srv.Close()

			apiKey := "Api Key"
			seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
			require.NoError(t, err)
			ingestionSeedsClient := newSeedsClient(srv.URL, apiKey)
			ingestionSeedExecutionsClient := newSeedExecutionsClient(ingestionSeedsClient, seedId)

			executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
			require.NoError(t, err)

			serverId, err := uuid.Parse("f6950327-3175-4a98-a570-658df852424a")
			require.NoError(t, err)
			response, err := ingestionSeedExecutionsClient.Server(executionId, serverId)
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

// Test_seedExecutionsClient_Credential tests the seedExecutionsClient.Credential() function.
func Test_seedExecutionsClient_Credential(t *testing.T) {
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
			name:             "Credential returns a real response",
			method:           http.MethodGet,
			path:             "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/credential/9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
			statusCode:       http.StatusOK,
			response:         `{"id":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","name":"MongoDB credential","type":"mongo","active":true,"labels":[],"secret":"mongo-secret","creationTimestamp":"2025-08-14T18:02:11Z","lastUpdatedTimestamp":"2025-08-14T18:02:11Z"}`,
			expectedResponse: gjson.Parse(`{"id":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","name":"MongoDB credential","type":"mongo","active":true,"labels":[],"secret":"mongo-secret","creationTimestamp":"2025-08-14T18:02:11Z","lastUpdatedTimestamp":"2025-08-14T18:02:11Z"}`),
			err:              nil,
		},
		// Error case
		{
			name:       "credential config returns not found",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/credential/9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1003,
			"messages": [
				"Credential not found: 9ababe08-0b74-4672-bb7c-e7a8227d6d4c"
			],
			"timestamp": "2025-09-03T17:44:01.557816Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Credential not found: 9ababe08-0b74-4672-bb7c-e7a8227d6d4c"
			],
			"timestamp": "2025-09-03T17:44:01.557816Z"
			}`)},
		},
		{
			name:       "credential config returns bad request",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/credential/9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [executionId] for value [-] due to: Invalid UUID string: -"
			],
			"timestamp": "2025-10-01T19:26:07.794214300Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [executionId] for value [-] due to: Invalid UUID string: -"
			],
			"timestamp": "2025-10-01T19:26:07.794214300Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, "/seed/2acd0a61-852c-4f38-af2b-9c84e152873e/execution/"+strings.TrimLeft(tc.path, "/"), r.URL.Path)
				}))
			defer srv.Close()

			apiKey := "Api Key"
			seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
			require.NoError(t, err)
			ingestionSeedsClient := newSeedsClient(srv.URL, apiKey)
			ingestionSeedExecutionsClient := newSeedExecutionsClient(ingestionSeedsClient, seedId)

			executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
			require.NoError(t, err)

			credentialId, err := uuid.Parse("9ababe08-0b74-4672-bb7c-e7a8227d6d4c")
			require.NoError(t, err)
			response, err := ingestionSeedExecutionsClient.Credential(executionId, credentialId)
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

// Test_seedExecutionsClient_Halt tests the seedExecutionsClient.Halt() function.
func Test_seedExecutionsClient_Halt(t *testing.T) {
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
			name:             "Halt works correctly",
			method:           http.MethodPost,
			path:             "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/halt",
			statusCode:       http.StatusOK,
			response:         `{"acknowledged":true}`,
			expectedResponse: gjson.Parse(`{"acknowledged":true}`),
			err:              nil,
		},
		// Error cases
		{
			name:       "Halt fails because it is already halting or in a state that does not allow it.",
			method:     http.MethodPost,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/halt",
			statusCode: http.StatusConflict,
			response: `{
			"status": 409,
			"code": 4001,
			"messages": [
				"Action HALT cannot be applied to seed execution cc89b714-d00a-4774-9c45-9497b5d9f8ef because of its current status: HALTING"
			],
			"timestamp": "2025-09-03T21:05:21.861757200Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusConflict, Body: gjson.Parse(`{
			"status": 409,
			"code": 4001,
			"messages": [
				"Action HALT cannot be applied to seed execution cc89b714-d00a-4774-9c45-9497b5d9f8ef because of its current status: HALTING"
			],
			"timestamp": "2025-09-03T21:05:21.861757200Z"
			}`)},
		},
		{
			name:       "Halt fails because the execution was not found.",
			method:     http.MethodPost,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/halt",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1003,
			"messages": [
				"Seed execution not found: cc89b714-d00a-4774-9c45-9497b5d9f8e3"
			],
			"timestamp": "2025-09-03T21:37:21.871825500Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Seed execution not found: cc89b714-d00a-4774-9c45-9497b5d9f8e3"
			],
			"timestamp": "2025-09-03T21:37:21.871825500Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, "/seed/2acd0a61-852c-4f38-af2b-9c84e152873e/execution/"+strings.TrimLeft(tc.path, "/"), r.URL.Path)
				}))
			defer srv.Close()

			apiKey := "Api Key"
			seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
			require.NoError(t, err)
			ingestionSeedsClient := newSeedsClient(srv.URL, apiKey)
			ingestionSeedExecutionsClient := newSeedExecutionsClient(ingestionSeedsClient, seedId)

			executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
			require.NoError(t, err)

			response, err := ingestionSeedExecutionsClient.Halt(executionId)
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

// Test_seedExecutionsClient_Records tests the seedExecutionsClient.Records() function.
func Test_seedExecutionsClient_Records(t *testing.T) {
	url := "http://localhost:12030"
	apiKey := "Api Key"
	seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
	require.NoError(t, err)
	ingestionSeedsClient := newSeedsClient(url, apiKey)
	ingestionSeedExecutionsClient := newSeedExecutionsClient(ingestionSeedsClient, seedId)

	executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
	require.NoError(t, err)

	ingestionSeedExecutionRecordsClient := ingestionSeedExecutionsClient.Records(executionId)

	assert.Equal(t, apiKey, ingestionSeedExecutionRecordsClient.client.ApiKey)
	assert.Equal(t, url+"/seed/"+seedId.String()+"/execution/"+executionId.String()+"/record", ingestionSeedExecutionRecordsClient.client.client.BaseURL)
}

// Test_seedExecutionsClient_Jobs tests the seedExecutionsClient.Jobs() function.
func Test_seedExecutionsClient_Jobs(t *testing.T) {
	url := "http://localhost:12030"
	apiKey := "Api Key"
	seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
	require.NoError(t, err)
	ingestionSeedsClient := newSeedsClient(url, apiKey)
	ingestionSeedExecutionsClient := newSeedExecutionsClient(ingestionSeedsClient, seedId)

	executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
	require.NoError(t, err)

	ingestionSeedExecutionJobClient := ingestionSeedExecutionsClient.Jobs(executionId)

	assert.Equal(t, apiKey, ingestionSeedExecutionJobClient.client.ApiKey)
	assert.Equal(t, url+"/seed/"+seedId.String()+"/execution/"+executionId.String()+"/job", ingestionSeedExecutionJobClient.client.client.BaseURL)
}

// Test_seedExecutionsClient_Audit_HTTPResponseCases tests how the seedExecutionsClient.Audit() function behaves when receiving different HTTP responses and errors.
// It does not test if reading all the pages works.
func Test_seedExecutionsClient_Audit_HTTPResponseCases(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		statusCode  int
		response    string
		expectedLen int
		err         error
	}{
		// Working cases
		{
			name:       "Audit returns array",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/audit",
			statusCode: http.StatusOK,
			response: `{
			"content": [
				{
				"timestamp": "2025-09-03T19:58:09.495Z",
				"status": "CREATED",
				"stages": []
				},
				{
				"timestamp": "2025-09-03T19:58:18.379Z",
				"status": "RUNNING",
				"stages": []
				},
				{
				"timestamp": "2025-09-03T19:58:25.277Z",
				"status": "RUNNING",
				"stages": [
					"BEFORE_HOOKS"
				]
				},
				{
				"timestamp": "2025-09-03T20:02:33.461Z",
				"status": "RUNNING",
				"stages": [
					"BEFORE_HOOKS",
					"INGEST"
				]
				},
				{
				"timestamp": "2025-09-03T20:02:48.877Z",
				"status": "RUNNING",
				"stages": [
					"BEFORE_HOOKS",
					"INGEST",
					"RETRY"
				]
				},
				{
				"timestamp": "2025-09-03T20:02:49.097Z",
				"status": "RUNNING",
				"stages": [
					"BEFORE_HOOKS",
					"INGEST",
					"RETRY",
					"AFTER_HOOKS"
				]
				},
				{
				"timestamp": "2025-09-03T20:02:56.812Z",
				"status": "DONE",
				"stages": [
					"BEFORE_HOOKS",
					"INGEST",
					"RETRY",
					"AFTER_HOOKS"
				]
				}
			],
			"pageable": {
				"page": 0,
				"size": 25,
				"sort": []
			},
			"totalSize": 7,
			"totalPages": 1,
			"empty": false,
			"size": 25,
			"offset": 0,
			"numberOfElements": 7,
			"pageNumber": 0
			}`,
			expectedLen: 7,
			err:         nil,
		},
		{
			name:        "Audit returns no content",
			method:      http.MethodGet,
			path:        "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/audit",
			statusCode:  http.StatusNoContent,
			response:    `{"content": []}`,
			expectedLen: 0,
			err:         nil,
		},
		{
			name:        "Audit has no content field",
			method:      http.MethodGet,
			path:        "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/audit",
			statusCode:  http.StatusNoContent,
			response:    ``,
			expectedLen: 0,
			err:         nil,
		},

		// Error cases
		{
			name:       "Audit returns a 400 Bad request",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/audit",
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [executionId] for value [nouuid] due to: Invalid UUID string: nouuid"
			],
			"timestamp": "2025-09-03T22:09:32.940650300Z"
			}`,
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [executionId] for value [nouuid] due to: Invalid UUID string: nouuid"
			],
			"timestamp": "2025-09-03T22:09:32.940650300Z"
			}`)},
		},
		{
			name:       "Audit returns a 404 Not found",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/audit",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1003,
			"messages": [
				"Seed execution not found: 2acd0a61-852c-4f38-af2b-9c84e152873e"
			],
			"timestamp": "2025-09-03T22:43:49.251888500Z"
			}`,
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Seed execution not found: 2acd0a61-852c-4f38-af2b-9c84e152873e"
			],
			"timestamp": "2025-09-03T22:43:49.251888500Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, "/seed/2acd0a61-852c-4f38-af2b-9c84e152873e/execution/"+strings.TrimLeft(tc.path, "/"), r.URL.Path)
				}))
			defer srv.Close()

			apiKey := "Api Key"
			seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
			require.NoError(t, err)
			ingestionSeedsClient := newSeedsClient(srv.URL, apiKey)
			ingestionSeedExecutionsClient := newSeedExecutionsClient(ingestionSeedsClient, seedId)

			executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
			require.NoError(t, err)

			response, err := ingestionSeedExecutionsClient.Audit(executionId)
			if tc.err == nil {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedLen, len(response))
			} else {
				assert.Equal(t, []gjson.Result(nil), response)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

// Test_seedRecordsClient_Get tests the seedRecordsClient.Get() function.
func Test_seedRecordsClient_Get(t *testing.T) {
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
			name:       "Get by ID returns object",
			method:     http.MethodGet,
			path:       "/seed/2acd0a61-852c-4f38-af2b-9c84e152873e/record/A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=",
			statusCode: http.StatusOK,
			response: `{
				"id": {
					"plain": "4e7c8a47efd829ef7f710d64da661786",
					"hash": "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0="
				},
				"creationTimestamp": "2025-09-03T21:02:54Z",
				"lastUpdatedTimestamp": "2025-09-03T21:02:54Z",
				"status": "SUCCESS"
			}`,
			expectedResponse: gjson.Parse(`{
				"id": {
					"plain": "4e7c8a47efd829ef7f710d64da661786",
					"hash": "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0="
				},
				"creationTimestamp": "2025-09-03T21:02:54Z",
				"lastUpdatedTimestamp": "2025-09-03T21:02:54Z",
				"status": "SUCCESS"
			}`),
			err: nil,
		},

		// Error case
		{
			name:       "Get by ID returns 404 Not Found",
			method:     http.MethodGet,
			path:       "/seed/2acd0a61-852c-4f38-af2b-9c84e152873e/record/A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1003,
			"messages": [
				"Entity not found: SeedRecordId(seed=Seed(super=AbstractComponentConfigEntity(super=AbstractJsonConfigEntity(super=AbstractTypedConfigEntity(super=AbstractConfigEntity(super=AbstractUpdatableEntity(super=AbstractCoreEntity(id=2acd0a61-852c-4f38-af2b-9c84e152873e), creationTimestamp=2025-08-21T21:52:03Z, lastUpdatedTimestamp=2025-08-21T21:52:03Z), name=Search seed, description=null, active=true), type=staging), config={\"action\":\"scroll\",\"bucket\":\"blogs\"})), properties=null, labels=[], recordOptions=SeedRecordPolicy[timeoutPolicy=TimeoutPolicy[slice=PT1H], errorPolicy=FATAL, outboundPolicy=OutboundPolicy[idPolicy=IdPolicy[generator=null], batchPolicy=BatchPolicy[maxCount=25, flushAfter=PT1M]]], hooks=[], beforeHooksOptions=null, afterHooksOptions=null), recordId=[3, 113, -45, 12, 72, 2, 107, -82, 65, 21, -101, 26, 115, -44, -56, -100, 88, -84, -66, 90, 17, -108, -67, -52, -25, 72, -93, 9, 99])"
			],
			"timestamp": "2025-09-04T14:07:13.759984600Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Entity not found: SeedRecordId(seed=Seed(super=AbstractComponentConfigEntity(super=AbstractJsonConfigEntity(super=AbstractTypedConfigEntity(super=AbstractConfigEntity(super=AbstractUpdatableEntity(super=AbstractCoreEntity(id=2acd0a61-852c-4f38-af2b-9c84e152873e), creationTimestamp=2025-08-21T21:52:03Z, lastUpdatedTimestamp=2025-08-21T21:52:03Z), name=Search seed, description=null, active=true), type=staging), config={\"action\":\"scroll\",\"bucket\":\"blogs\"})), properties=null, labels=[], recordOptions=SeedRecordPolicy[timeoutPolicy=TimeoutPolicy[slice=PT1H], errorPolicy=FATAL, outboundPolicy=OutboundPolicy[idPolicy=IdPolicy[generator=null], batchPolicy=BatchPolicy[maxCount=25, flushAfter=PT1M]]], hooks=[], beforeHooksOptions=null, afterHooksOptions=null), recordId=[3, 113, -45, 12, 72, 2, 107, -82, 65, 21, -101, 26, 115, -44, -56, -100, 88, -84, -66, 90, 17, -108, -67, -52, -25, 72, -93, 9, 99])"
			],
			"timestamp": "2025-09-04T14:07:13.759984600Z"
			}`)},
		},
		{
			name:       "Get by ID returns 400 Not Found",
			method:     http.MethodGet,
			path:       "/seed/2acd0a61-852c-4f38-af2b-9c84e152873e/record/A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=",
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [seedId] for value [notaseed] due to: Invalid UUID string: notaseed"
			],
			"timestamp": "2025-10-01T19:43:37.753734900Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [seedId] for value [notaseed] due to: Invalid UUID string: notaseed"
			],
			"timestamp": "2025-10-01T19:43:37.753734900Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, tc.path, r.URL.Path)
			}))

			defer srv.Close()

			apiKey := "Api Key"
			seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
			require.NoError(t, err)
			ingestionSeedsClient := newSeedsClient(srv.URL, apiKey)
			ingestionSeedRecordsClient := newSeedRecordsClient(ingestionSeedsClient, seedId)

			response, err := ingestionSeedRecordsClient.Get("A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=")
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

// Test_seedRecordsClient_GetAll tests the seedRecords.GetAll() function
func Test_seedRecordsClient_GetAll(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		statusCode  int
		response    string
		expectedLen int
		err         error
	}{
		// Working cases
		{
			name:       "GetAll returns array",
			method:     http.MethodGet,
			path:       "/seed/2acd0a61-852c-4f38-af2b-9c84e152873e/record",
			statusCode: http.StatusOK,
			response: `{
			"content": [
				{"id":{"plain":"4e7c8a47efd829ef7f710d64da661786","hash":"A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0="},"creationTimestamp":"2025-09-05T20:13:47Z","lastUpdatedTimestamp":"2025-09-05T20:13:47Z","status":"SUCCESS"},
				{"id":{"plain":"8148e6a7b952a3b2964f706ced8c6885","hash":"IJeF-losyj33EAuqjgGW2G7sT-eE7poejQ5HokerZio="},"creationTimestamp":"2025-09-05T20:13:47Z","lastUpdatedTimestamp":"2025-09-05T20:13:47Z","status":"SUCCESS"},
				{"id":{"plain":"b1e3e4f42c0818b1580e306eb776d4a1","hash":"N2lubqCWTqEEaymQVntpdP5dqKDP-LYk81C_PCr6btQ="},"creationTimestamp":"2025-09-05T20:13:47Z","lastUpdatedTimestamp":"2025-09-05T20:13:47Z","status":"SUCCESS"},
				{"id":{"plain":"5625c64483bef0d48e9ad91aca9b2f94","hash":"d5_RAnEgZUtF8FjQTXFYFFfwvLuBrS0SAt-USRn3g4g="},"creationTimestamp":"2025-09-05T20:13:47Z","lastUpdatedTimestamp":"2025-09-05T20:13:47Z","status":"SUCCESS"},
				{"id":{"plain":"768b0a3bcee501dc624484ba8a0d7f6d","hash":"hKScM7isXQIHr7ctMf-qQzWU58PYz0BSjIPLU2ksdnw="},"creationTimestamp":"2025-09-05T20:13:47Z","lastUpdatedTimestamp":"2025-09-05T20:13:47Z","status":"SUCCESS"},
				{"id":{"plain":"c28db957887e1aae75e7ab1dd0fd34e9","hash":"loMWIdFvJkPiFvIbFA21woPHx2fgDt1cAwmIVv8eS-I="},"creationTimestamp":"2025-09-05T20:13:47Z","lastUpdatedTimestamp":"2025-09-05T20:13:47Z","status":"SUCCESS"},
				{"id":{"plain":"d758c733466967ea6f13b20bcbfcebb5","hash":"reWGUUVo0_ziOCyCKOZmovXFD0pAVizu7gcrGB33Jxg="},"creationTimestamp":"2025-09-05T20:13:47Z","lastUpdatedTimestamp":"2025-09-05T20:13:47Z","status":"SUCCESS"},
				{"id":{"plain":"232638a332048c4cb159f8cf6636507f","hash":"xk-6yAQIaIyHiJRUFa3va-QJbbix9QUq77cgzzFIVdA="},"creationTimestamp":"2025-09-05T20:13:47Z","lastUpdatedTimestamp":"2025-09-05T20:13:47Z","status":"SUCCESS"}
			],
			"pageable": {
				"page": 0,
				"size": 25,
				"sort": []
			},
			"totalSize": 8,
			"totalPages": 1,
			"empty": false,
			"size": 25,
			"offset": 0,
			"numberOfElements": 8,
			"pageNumber": 0
			}`,
			expectedLen: 8,
			err:         nil,
		},
		{
			name:        "GetAll returns no content",
			method:      http.MethodGet,
			path:        "/seed/2acd0a61-852c-4f38-af2b-9c84e152873e/record",
			statusCode:  http.StatusNoContent,
			response:    `{"content": []}`,
			expectedLen: 0,
			err:         nil,
		},
		{
			name:        "GetAll has no content field",
			method:      http.MethodGet,
			path:        "/seed/2acd0a61-852c-4f38-af2b-9c84e152873e/record",
			statusCode:  http.StatusNoContent,
			response:    ``,
			expectedLen: 0,
			err:         nil,
		},

		// Error cases
		{
			name:       "GetAll returns a 400 Bad request",
			method:     http.MethodGet,
			path:       "/seed/2acd0a61-852c-4f38-af2b-9c84e152873e/record",
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [executionId] for value [nouuid] due to: Invalid UUID string: nouuid"
			],
			"timestamp": "2025-09-03T22:09:32.940650300Z"
			}`,
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [executionId] for value [nouuid] due to: Invalid UUID string: nouuid"
			],
			"timestamp": "2025-09-03T22:09:32.940650300Z"
			}`)},
		},
		{
			name:       "GetAll returns a 404 Not found",
			method:     http.MethodGet,
			path:       "/seed/2acd0a61-852c-4f38-af2b-9c84e152873e/record",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1003,
			"messages": [
				"Seed not found: 2acd0a61-852c-4f38-af2b-9c84e152873e"
			],
			"timestamp": "2025-09-03T22:43:49.251888500Z"
			}`,
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Seed not found: 2acd0a61-852c-4f38-af2b-9c84e152873e"
			],
			"timestamp": "2025-09-03T22:43:49.251888500Z"
			}`)},
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

			apiKey := "Api Key"
			seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
			require.NoError(t, err)
			ingestionSeedsClient := newSeedsClient(srv.URL, apiKey)
			ingestionSeedRecordsClient := newSeedRecordsClient(ingestionSeedsClient, seedId)

			response, err := ingestionSeedRecordsClient.GetAll()
			if tc.err == nil {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedLen, len(response))
			} else {
				assert.Equal(t, []gjson.Result(nil), response)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

// Test_seedsClient_Start tests the seedsClient.Start() function
func Test_seedsClient_Start(t *testing.T) {
	tests := []struct {
		name                string
		method              string
		path                string
		statusCode          int
		response            string
		scan                ScanType
		expectedResponse    gjson.Result
		executionProperties gjson.Result
		err                 error
	}{
		// Working case
		{
			name:                "Start works correctly without executionProperties",
			method:              http.MethodPost,
			path:                "/2acd0a61-852c-4f38-af2b-9c84e152873e",
			statusCode:          http.StatusAccepted,
			response:            `{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","creationTimestamp":"2025-09-04T19:29:41.119013Z","lastUpdatedTimestamp":"2025-09-04T19:29:41.119013Z","triggerType":"MANUAL","status":"CREATED","scanType":"FULL"}`,
			scan:                ScanFull,
			expectedResponse:    gjson.Parse(`{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","creationTimestamp":"2025-09-04T19:29:41.119013Z","lastUpdatedTimestamp":"2025-09-04T19:29:41.119013Z","triggerType":"MANUAL","status":"CREATED","scanType":"FULL"}`),
			executionProperties: gjson.Result{},
			err:                 nil,
		},
		{
			name:                "Start works correctly with executionProperties",
			method:              http.MethodPost,
			path:                "/2acd0a61-852c-4f38-af2b-9c84e152873e",
			statusCode:          http.StatusAccepted,
			scan:                ScanIncremental,
			response:            `{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","creationTimestamp":"2025-09-04T19:29:41.119013Z","lastUpdatedTimestamp":"2025-09-04T19:29:41.119013Z","triggerType":"MANUAL","status":"CREATED","scanType":"INCREMENTAL","properties":{"stagingBucket":"testBucket"}}`,
			expectedResponse:    gjson.Parse(`{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","creationTimestamp":"2025-09-04T19:29:41.119013Z","lastUpdatedTimestamp":"2025-09-04T19:29:41.119013Z","triggerType":"MANUAL","status":"CREATED","scanType":"INCREMENTAL","properties":{"stagingBucket":"testBucket"}}`),
			executionProperties: gjson.Parse(`{"stagingBucket":"testBucket"}`),
			err:                 nil,
		},
		{
			name:                "Start works correctly with OK",
			method:              http.MethodPost,
			path:                "/2acd0a61-852c-4f38-af2b-9c84e152873e",
			statusCode:          http.StatusOK,
			scan:                ScanFull,
			response:            `{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","creationTimestamp":"2025-09-04T19:29:41.119013Z","lastUpdatedTimestamp":"2025-09-04T19:29:41.119013Z","triggerType":"MANUAL","status":"CREATED","scanType":"FULL"}`,
			expectedResponse:    gjson.Parse(`{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","creationTimestamp":"2025-09-04T19:29:41.119013Z","lastUpdatedTimestamp":"2025-09-04T19:29:41.119013Z","triggerType":"MANUAL","status":"CREATED","scanType":"FULL"}`),
			executionProperties: gjson.Result{},
			err:                 nil,
		},
		// Error cases
		{
			name:       "Start fails because the seed already has active executions.",
			method:     http.MethodPost,
			path:       "/2acd0a61-852c-4f38-af2b-9c84e152873e",
			statusCode: http.StatusConflict,
			scan:       ScanFull,
			response: `{
			"status": 409,
			"code": 4001,
			"messages": [
				"The seed has 1 executions: 0c309dbb-0402-4710-8659-2c75f5d649b6"
			],
			"timestamp": "2025-09-04T20:17:00.116546400Z"
			}`,
			expectedResponse:    gjson.Result{},
			executionProperties: gjson.Result{},
			err: Error{Status: http.StatusConflict, Body: gjson.Parse(`{
			"status": 409,
			"code": 4001,
			"messages": [
				"The seed has 1 executions: 0c309dbb-0402-4710-8659-2c75f5d649b6"
			],
			"timestamp": "2025-09-04T20:17:00.116546400Z"
			}`)},
		},
		{
			name:       "start fails because the seed was not found.",
			method:     http.MethodPost,
			path:       "/2acd0a61-852c-4f38-af2b-9c84e152873e",
			statusCode: http.StatusNotFound,
			scan:       ScanFull,
			response: `{
			"status": 404,
			"code": 1003,
			"messages": [
				"Entity not found: 2acd0a61-852c-4f38-af2b-9c84e152873e"
			],
			"timestamp": "2025-09-04T20:20:47.326270700Z"
			}`,
			expectedResponse:    gjson.Result{},
			executionProperties: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Entity not found: 2acd0a61-852c-4f38-af2b-9c84e152873e"
			],
			"timestamp": "2025-09-04T20:20:47.326270700Z"
			}`)},
		},
		{
			name:       "start fails because of bad request",
			method:     http.MethodPost,
			path:       "/2acd0a61-852c-4f38-af2b-9c84e152873e",
			statusCode: http.StatusBadRequest,
			scan:       ScanFull,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [id] for value [notuuid] due to: Invalid UUID string: notuuid"
			],
			"timestamp": "2025-10-01T20:01:34.588816100Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [id] for value [notuuid] due to: Invalid UUID string: notuuid"
			],
			"timestamp": "2025-10-01T20:01:34.588816100Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, "/seed"+tc.path, r.URL.Path)
					if tc.executionProperties.Exists() {
						body, _ := io.ReadAll(r.Body)
						assert.Equal(t, tc.executionProperties.Raw, string(body))
					}
					assert.Equal(t, string(tc.scan), r.URL.Query().Get("scanType"))
				}))
			defer srv.Close()

			apiKey := "Api Key"
			seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
			require.NoError(t, err)
			ingestionSeedsClient := newSeedsClient(srv.URL, apiKey)
			response, err := ingestionSeedsClient.Start(seedId, tc.scan, tc.executionProperties)
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

// Test_seedsClient_Halt tests the seedsClient.Halt() function
func Test_seedsClient_Halt(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		response         string
		expectedResponse []gjson.Result
		err              error
	}{
		// Working case
		{
			name:             "Halt works correctly",
			method:           http.MethodPost,
			path:             "/2acd0a61-852c-4f38-af2b-9c84e152873e/halt",
			statusCode:       http.StatusMultiStatus,
			response:         `[{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","status":202}]`,
			expectedResponse: gjson.Parse(`[{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","status":202}]`).Array(),
			err:              nil,
		},
		{
			name:             "Halt returns an empty array",
			method:           http.MethodPost,
			path:             "/2acd0a61-852c-4f38-af2b-9c84e152873e/halt",
			statusCode:       http.StatusMultiStatus,
			response:         `[]`,
			expectedResponse: []gjson.Result{},
			err:              nil,
		},
		{
			name:             "Halt returns an empty array with OK",
			method:           http.MethodPost,
			path:             "/2acd0a61-852c-4f38-af2b-9c84e152873e/halt",
			statusCode:       http.StatusOK,
			response:         `[]`,
			expectedResponse: []gjson.Result{},
			err:              nil,
		},
		// Error case
		{
			name:       "halt fails because the seed was not found.",
			method:     http.MethodPost,
			path:       "/2acd0a61-852c-4f38-af2b-9c84e152873e/halt",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1003,
			"messages": [
				"Entity not found: 2acd0a61-852c-4f38-af2b-9c84e152873e"
			],
			"timestamp": "2025-09-04T20:20:47.326270700Z"
			}`,
			expectedResponse: []gjson.Result(nil),
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Entity not found: 2acd0a61-852c-4f38-af2b-9c84e152873e"
			],
			"timestamp": "2025-09-04T20:20:47.326270700Z"
			}`)},
		},
		{
			name:       "halt fails because of bad request",
			method:     http.MethodPost,
			path:       "/2acd0a61-852c-4f38-af2b-9c84e152873e/halt",
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [id] for value [notuuid] due to: Invalid UUID string: notuuid"
			],
			"timestamp": "2025-10-01T20:01:34.588816100Z"
			}`,
			expectedResponse: []gjson.Result(nil),
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [id] for value [notuuid] due to: Invalid UUID string: notuuid"
			],
			"timestamp": "2025-10-01T20:01:34.588816100Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, "/seed"+tc.path, r.URL.Path)
				}))
			defer srv.Close()

			apiKey := "Api Key"
			seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
			require.NoError(t, err)
			ingestionSeedsClient := newSeedsClient(srv.URL, apiKey)
			response, err := ingestionSeedsClient.Halt(seedId)
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

// Test_seedsClient_Reset tests the seedsClient.Reset() function
func Test_seedsClient_Reset(t *testing.T) {
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
			name:             "Reset works correctly",
			method:           http.MethodPost,
			path:             "/2acd0a61-852c-4f38-af2b-9c84e152873e/reset",
			statusCode:       http.StatusMultiStatus,
			response:         `{"acknowledged":true}`,
			expectedResponse: gjson.Parse(`{"acknowledged":true}`),
			err:              nil,
		},
		// Error case
		{
			name:       "Reset fails because the seed has active executions.",
			method:     http.MethodPost,
			path:       "/2acd0a61-852c-4f38-af2b-9c84e152873e/reset",
			statusCode: http.StatusConflict,
			response: `{
			"status": 409,
			"code": 4001,
			"messages": [
				"Can not reset the seed '2acd0a61-852c-4f38-af2b-9c84e152873e' because it has active executions."
			],
			"timestamp": "2025-09-04T21:00:41.928010Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusConflict, Body: gjson.Parse(`{
			"status": 409,
			"code": 4001,
			"messages": [
				"Can not reset the seed '2acd0a61-852c-4f38-af2b-9c84e152873e' because it has active executions."
			],
			"timestamp": "2025-09-04T21:00:41.928010Z"
			}`)},
		},
		{
			name:       "Reset fails because the seed was not found.",
			method:     http.MethodPost,
			path:       "/2acd0a61-852c-4f38-af2b-9c84e152873e/reset",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1003,
			"messages": [
				"Entity not found: 2acd0a61-852c-4f38-af2b-9c84e152873e"
			],
			"timestamp": "2025-09-04T20:20:47.326270700Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Entity not found: 2acd0a61-852c-4f38-af2b-9c84e152873e"
			],
			"timestamp": "2025-09-04T20:20:47.326270700Z"
			}`)},
		},
		{
			name:       "reset fails because of bad request",
			method:     http.MethodPost,
			path:       "/2acd0a61-852c-4f38-af2b-9c84e152873e/reset",
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [id] for value [notuuid] due to: Invalid UUID string: notuuid"
			],
			"timestamp": "2025-10-01T20:01:34.588816100Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [id] for value [notuuid] due to: Invalid UUID string: notuuid"
			],
			"timestamp": "2025-10-01T20:01:34.588816100Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, "/seed"+tc.path, r.URL.Path)
				}))
			defer srv.Close()

			apiKey := "Api Key"
			seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
			require.NoError(t, err)
			ingestionSeedsClient := newSeedsClient(srv.URL, apiKey)
			response, err := ingestionSeedsClient.Reset(seedId)
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

// Test_seedsClient_Records tests the seedsClient.Records() function.
func Test_seedsClient_Records(t *testing.T) {
	url := "http://localhost:12030"
	apiKey := "Api Key"
	seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
	require.NoError(t, err)
	ingestionSeedsClient := newSeedsClient(url, apiKey)
	ingestionSeedRecordsClient := ingestionSeedsClient.Records(seedId)

	assert.Equal(t, apiKey, ingestionSeedRecordsClient.summarizer.client.ApiKey)
	assert.Equal(t, url+"/seed/"+seedId.String()+"/record", ingestionSeedRecordsClient.summarizer.client.client.BaseURL)
}

// Test_newSeedsClient tests the seedClient.Executions() function
func Test_seedsClient_Executions(t *testing.T) {
	url := "http://localhost:12030"
	apiKey := "Api Key"
	seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
	require.NoError(t, err)
	ingestionSeedsClient := newSeedsClient(url, apiKey)
	ingestionSeedExecutionsClient := ingestionSeedsClient.Executions(seedId)

	assert.Equal(t, apiKey, ingestionSeedExecutionsClient.getter.client.ApiKey)
	assert.Equal(t, url+"/seed/"+seedId.String()+"/execution", ingestionSeedExecutionsClient.getter.client.client.BaseURL)
}

// Test_ingestion_Processors tests the ingestion.Processors() function
func Test_ingestion_Processors(t *testing.T) {
	i := NewIngestion("http://localhost:12030", "Api Key")
	ipc := i.Processors()

	assert.Equal(t, i.ApiKey, ipc.crud.client.ApiKey)
	assert.Equal(t, i.Url+"/processor", ipc.crud.client.client.BaseURL)
	assert.Equal(t, i.ApiKey, ipc.cloner.client.ApiKey)
	assert.Equal(t, i.Url+"/processor", ipc.cloner.client.client.BaseURL)
	assert.Equal(t, i.ApiKey, ipc.searcher.client.ApiKey)
	assert.Equal(t, i.Url+"/processor", ipc.searcher.client.client.BaseURL)
}

// Test_ingestion_Pipelines tests the ingestion.Pipelines() function
func Test_ingestion_Pipelines(t *testing.T) {
	i := NewIngestion("http://localhost:12030", "Api Key")
	ipc := i.Pipelines()

	assert.Equal(t, i.ApiKey, ipc.crud.client.ApiKey)
	assert.Equal(t, i.Url+"/pipeline", ipc.crud.client.client.BaseURL)
	assert.Equal(t, i.ApiKey, ipc.cloner.client.ApiKey)
	assert.Equal(t, i.Url+"/pipeline", ipc.cloner.client.client.BaseURL)
	assert.Equal(t, i.ApiKey, ipc.searcher.client.ApiKey)
	assert.Equal(t, i.Url+"/pipeline", ipc.searcher.client.client.BaseURL)
}

// Test_ingestion_Seeds test the ingestion.Seeds() function.
func Test_ingestion_Seeds(t *testing.T) {
	i := NewIngestion("http://localhost:12030", "Api Key")
	ipc := i.Seeds()

	assert.Equal(t, i.ApiKey, ipc.crud.client.ApiKey)
	assert.Equal(t, i.Url+"/seed", ipc.crud.client.client.BaseURL)
	assert.Equal(t, i.ApiKey, ipc.cloner.client.ApiKey)
	assert.Equal(t, i.Url+"/seed", ipc.cloner.client.client.BaseURL)
}

// Test_ingestion_BackupRestore tests the ingestion.BackupRestore() function
func Test_ingestion_BackupRestore(t *testing.T) {
	i := NewIngestion("http://localhost:12030", "Api Key")
	bc := i.BackupRestore()

	assert.Equal(t, i.ApiKey, bc.ApiKey)
	assert.Equal(t, i.Url, bc.client.client.BaseURL)
}

// Test_NewIngestion tests the ingestion constructor
func Test_NewIngestion(t *testing.T) {
	i := NewIngestion("http://localhost:12030", "Api Key")

	assert.Equal(t, "http://localhost:12030/v2", i.Url)
	assert.Equal(t, "Api Key", i.ApiKey)
}

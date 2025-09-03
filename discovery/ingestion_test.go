package discovery

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Test_newQueryFlowProcessorsClient test the queryFlowProcessorsClient's constructor
func Test_newSeedExecutionsClient(t *testing.T) {
	url := "http://localhost:8083/v2"
	apiKey := "Api Key"
	seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
	if err != nil {
		fmt.Println("UUID conversion failed: " + err.Error())
		return
	}
	ingestionSeedExecutionsClient := newSeedExecutionsClient(url, apiKey, seedId)

	assert.Equal(t, apiKey, ingestionSeedExecutionsClient.getter.client.ApiKey)
	assert.Equal(t, url+"/seed/"+seedId.String()+"/execution", ingestionSeedExecutionsClient.getter.client.client.BaseURL)
}

// Test_newEndpointsClient tests the constructor of endpointsClients.
func Test_newSeedRecordsClient(t *testing.T) {
	url := "http://localhost:8083/v2"
	apiKey := "Api Key"
	seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
	if err != nil {
		fmt.Println("UUID conversion failed: " + err.Error())
		return
	}
	ingestionSeedRecordsClient := newSeedRecordsClient(url, apiKey, seedId)

	assert.Equal(t, apiKey, ingestionSeedRecordsClient.getter.client.ApiKey)
	assert.Equal(t, url+"/seed/"+seedId.String()+"/record", ingestionSeedRecordsClient.getter.client.client.BaseURL)
	assert.Equal(t, apiKey, ingestionSeedRecordsClient.summarizer.client.ApiKey)
	assert.Equal(t, url+"/seed/"+seedId.String()+"/record", ingestionSeedRecordsClient.summarizer.client.client.BaseURL)
}

func Test_newSeedExecutionRecordsClient(t *testing.T) {
	url := "http://localhost:8083/v2"
	apiKey := "Api Key"
	seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
	if err != nil {
		fmt.Println("UUID conversion failed: " + err.Error())
		return
	}
	ingestionSeedExecutionsClient := newSeedExecutionsClient(url, apiKey, seedId)

	executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
	if err != nil {
		fmt.Println("UUID conversion failed: " + err.Error())
		return
	}
	ingestionSeedExecutionRecordsClient := newSeedExecutionRecordsClient(ingestionSeedExecutionsClient, executionId)

	assert.Equal(t, apiKey, ingestionSeedExecutionRecordsClient.client.ApiKey)
	assert.Equal(t, url+"/seed/"+seedId.String()+"/execution/"+executionId.String()+"/record", ingestionSeedExecutionRecordsClient.client.client.BaseURL)
}

func Test_newSeedExecutionJobsClient(t *testing.T) {
	url := "http://localhost:8083/v2"
	apiKey := "Api Key"
	seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
	if err != nil {
		fmt.Println("UUID conversion failed: " + err.Error())
		return
	}
	ingestionSeedExecutionsClient := newSeedExecutionsClient(url, apiKey, seedId)

	executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
	if err != nil {
		fmt.Println("UUID conversion failed: " + err.Error())
		return
	}
	ingestionSeedExecutionJobClient := newSeedExecutionJobsClient(ingestionSeedExecutionsClient, executionId)

	assert.Equal(t, apiKey, ingestionSeedExecutionJobClient.client.ApiKey)
	assert.Equal(t, url+"/seed/"+seedId.String()+"/execution/"+executionId.String()+"/job", ingestionSeedExecutionJobClient.client.client.BaseURL)
}

func Test_seedExecutionsClient_Seed(t *testing.T) {
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
			name:       "Seed returns a real response",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/seed",
			statusCode: http.StatusOK,
			response:   `{"id":"2acd0a61-852c-4f38-af2b-9c84e152873e","name":"Search seed","type":"staging","active":true,"config":{"action":"scroll","bucket":"blogs"},"labels":[],"pipeline":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","recordPolicy":{"errorPolicy":"FATAL","timeoutPolicy":{"slice":"PT1H"},"outboundPolicy":{"idPolicy":{},"batchPolicy":{"maxCount":25,"flushAfter":"PT1M"}}},"creationTimestamp":"2025-08-21T21:52:03Z","lastUpdatedTimestamp":"2025-08-21T21:52:03Z"}`,
			testFunc: func(t *testing.T, response gjson.Result, err error) {
				require.NoError(t, err)
				assert.Equal(t, "Search seed", response.Get("name").String())
				assert.Equal(t, "2acd0a61-852c-4f38-af2b-9c84e152873e", response.Get("id").String())
			},
		},
		// Error case
		{
			name:       "Seed config returns not found",
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
			testFunc: func(t *testing.T, response gjson.Result, err error) {
				assert.Equal(t, gjson.Result{}, response)
				require.Error(t, err)
				errorStruct, ok := err.(Error)
				if ok {
					assert.Equal(t, http.StatusNotFound, errorStruct.Status)
					assert.Equal(t, "Seed execution not found: 6b7f0b69-126f-49ab-b2ff-0a876f42e5ed", errorStruct.Body.Get("messages.0").String())
				}
			},
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
			if err != nil {
				fmt.Println("UUID conversion failed: " + err.Error())
				return
			}
			ingestionSeedExecutionsClient := newSeedExecutionsClient(srv.URL, apiKey, seedId)

			executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
			if err != nil {
				fmt.Println("UUID conversion failed: " + err.Error())
				return
			}
			response, err := ingestionSeedExecutionsClient.Seed(executionId)
			tc.testFunc(t, response, err)
		})
	}
}

func Test_seedExecutionsClient_Pipeline(t *testing.T) {
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
			name:       "Pipeline returns a real response",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/pipeline/9a74bf3a-eb2a-4334-b803-c92bf1bc45fe",
			statusCode: http.StatusOK,
			response:   `{"id":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","name":"Search pipeline","active":true,"labels":[],"states":{"ingestionState":{"type":"processor","processors":[{"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","active":true,"outputField":"header"},{"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8","active":true}]}},"initialState":"ingestionState","recordPolicy":{"idPolicy":{},"errorPolicy":"FAIL","retryPolicy":{"active":true,"maxRetries":3},"timeoutPolicy":{"record":"PT1M"},"outboundPolicy":{"batchPolicy":{"maxCount":25,"flushAfter":"PT1M"}}},"creationTimestamp":"2025-08-21T21:52:02Z","lastUpdatedTimestamp":"2025-08-21T21:52:02Z"}`,
			testFunc: func(t *testing.T, response gjson.Result, err error) {
				require.NoError(t, err)
				assert.Equal(t, "Search pipeline", response.Get("name").String())
				assert.Equal(t, "9a74bf3a-eb2a-4334-b803-c92bf1bc45fe", response.Get("id").String())
			},
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
			testFunc: func(t *testing.T, response gjson.Result, err error) {
				assert.Equal(t, gjson.Result{}, response)
				require.Error(t, err)
				errorStruct, ok := err.(Error)
				if ok {
					assert.Equal(t, http.StatusNotFound, errorStruct.Status)
					assert.Equal(t, "Pipeline not found: 9a74bf3a-eb2a-4334-b803-c92bf1bc45fe", errorStruct.Body.Get("messages.0").String())
				}
			},
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
			if err != nil {
				fmt.Println("UUID conversion failed: " + err.Error())
				return
			}
			ingestionSeedExecutionsClient := newSeedExecutionsClient(srv.URL, apiKey, seedId)

			executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
			if err != nil {
				fmt.Println("UUID conversion failed: " + err.Error())
				return
			}

			pipelineId, err := uuid.Parse("9a74bf3a-eb2a-4334-b803-c92bf1bc45fe")
			if err != nil {
				fmt.Println("UUID conversion failed: " + err.Error())
				return
			}
			response, err := ingestionSeedExecutionsClient.Pipeline(executionId, pipelineId)
			tc.testFunc(t, response, err)
		})
	}
}

func Test_seedExecutionsClient_Processor(t *testing.T) {
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
			name:       "Processor returns a real response",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/processor/aa0186f1-746f-4b20-b1b0-313bd79e78b8",
			statusCode: http.StatusOK,
			response:   `{"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8","name":"MongoDB store processor","type":"mongo","active":true,"config":{"data":{"link":"#{ data('/reference') }","author":"#{ data('/author') }","header":"#{ data('/header') }"},"action":"hydrate","database":"pureinsights","collection":"blogs"},"labels":[],"server":{"id":"f6950327-3175-4a98-a570-658df852424a","credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c"},"creationTimestamp":"2025-08-21T21:52:02Z","lastUpdatedTimestamp":"2025-08-21T21:52:02Z"}`,
			testFunc: func(t *testing.T, response gjson.Result, err error) {
				require.NoError(t, err)
				assert.Equal(t, "MongoDB store processor", response.Get("name").String())
				assert.Equal(t, "aa0186f1-746f-4b20-b1b0-313bd79e78b8", response.Get("id").String())
			},
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
			testFunc: func(t *testing.T, response gjson.Result, err error) {
				assert.Equal(t, gjson.Result{}, response)
				require.Error(t, err)
				errorStruct, ok := err.(Error)
				if ok {
					assert.Equal(t, http.StatusNotFound, errorStruct.Status)
					assert.Equal(t, "Processor not found: aa0186f1-746f-4b20-b1b0-313bd79e78b8", errorStruct.Body.Get("messages.0").String())
				}
			},
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
			if err != nil {
				fmt.Println("UUID conversion failed: " + err.Error())
				return
			}
			ingestionSeedExecutionsClient := newSeedExecutionsClient(srv.URL, apiKey, seedId)

			executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
			if err != nil {
				fmt.Println("UUID conversion failed: " + err.Error())
				return
			}

			processorId, err := uuid.Parse("aa0186f1-746f-4b20-b1b0-313bd79e78b8")
			if err != nil {
				fmt.Println("UUID conversion failed: " + err.Error())
				return
			}
			response, err := ingestionSeedExecutionsClient.Processor(executionId, processorId)
			tc.testFunc(t, response, err)
		})
	}
}

func Test_seedExecutionsClient_Server(t *testing.T) {
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
			name:       "Server returns a real response",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/server/f6950327-3175-4a98-a570-658df852424a",
			statusCode: http.StatusOK,
			response:   `{"id":"f6950327-3175-4a98-a570-658df852424a","name":"MongoDB store server","type":"mongo","active":true,"config":{"data":{"link":"#{ data('/reference') }","author":"#{ data('/author') }","header":"#{ data('/header') }"},"action":"hydrate","database":"pureinsights","collection":"blogs"},"labels":[],"server":{"id":"f6950327-3175-4a98-a570-658df852424a","credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c"},"creationTimestamp":"2025-08-21T21:52:02Z","lastUpdatedTimestamp":"2025-08-21T21:52:02Z"}`,
			testFunc: func(t *testing.T, response gjson.Result, err error) {
				require.NoError(t, err)
				assert.Equal(t, "MongoDB store server", response.Get("name").String())
				assert.Equal(t, "f6950327-3175-4a98-a570-658df852424a", response.Get("id").String())
			},
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
			testFunc: func(t *testing.T, response gjson.Result, err error) {
				assert.Equal(t, gjson.Result{}, response)
				require.Error(t, err)
				errorStruct, ok := err.(Error)
				if ok {
					assert.Equal(t, http.StatusNotFound, errorStruct.Status)
					assert.Equal(t, "Server not found: f6950327-3175-4a98-a570-658df852424a", errorStruct.Body.Get("messages.0").String())
				}
			},
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
			if err != nil {
				fmt.Println("UUID conversion failed: " + err.Error())
				return
			}
			ingestionSeedExecutionsClient := newSeedExecutionsClient(srv.URL, apiKey, seedId)

			executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
			if err != nil {
				fmt.Println("UUID conversion failed: " + err.Error())
				return
			}

			serverId, err := uuid.Parse("f6950327-3175-4a98-a570-658df852424a")
			if err != nil {
				fmt.Println("UUID conversion failed: " + err.Error())
				return
			}
			response, err := ingestionSeedExecutionsClient.Server(executionId, serverId)
			tc.testFunc(t, response, err)
		})
	}
}

func Test_seedExecutionsClient_Credential(t *testing.T) {
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
			name:       "Credential returns a real response",
			method:     http.MethodGet,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/config/credential/9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
			statusCode: http.StatusOK,
			response:   `{"id":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","name":"MongoDB credential","type":"mongo","active":true,"labels":[],"secret":"mongo-secret","creationTimestamp":"2025-08-14T18:02:11Z","lastUpdatedTimestamp":"2025-08-14T18:02:11Z"}`,
			testFunc: func(t *testing.T, response gjson.Result, err error) {
				require.NoError(t, err)
				assert.Equal(t, "MongoDB credential", response.Get("name").String())
				assert.Equal(t, "9ababe08-0b74-4672-bb7c-e7a8227d6d4c", response.Get("id").String())
			},
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
			testFunc: func(t *testing.T, response gjson.Result, err error) {
				assert.Equal(t, gjson.Result{}, response)
				require.Error(t, err)
				errorStruct, ok := err.(Error)
				if ok {
					assert.Equal(t, http.StatusNotFound, errorStruct.Status)
					assert.Equal(t, "Credential not found: 9ababe08-0b74-4672-bb7c-e7a8227d6d4c", errorStruct.Body.Get("messages.0").String())
				}
			},
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
			if err != nil {
				fmt.Println("UUID conversion failed: " + err.Error())
				return
			}
			ingestionSeedExecutionsClient := newSeedExecutionsClient(srv.URL, apiKey, seedId)

			executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
			if err != nil {
				fmt.Println("UUID conversion failed: " + err.Error())
				return
			}

			credentialId, err := uuid.Parse("9ababe08-0b74-4672-bb7c-e7a8227d6d4c")
			if err != nil {
				fmt.Println("UUID conversion failed: " + err.Error())
				return
			}
			response, err := ingestionSeedExecutionsClient.Credential(executionId, credentialId)
			tc.testFunc(t, response, err)
		})
	}
}

func Test_seedExecutionsClient_Halt(t *testing.T) {
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
			name:       "Halt works correctly",
			method:     http.MethodPost,
			path:       "/6b7f0b69-126f-49ab-b2ff-0a876f42e5ed/halt",
			statusCode: http.StatusOK,
			response:   `{"acknowledged":true}`,
			testFunc: func(t *testing.T, response gjson.Result, err error) {
				require.NoError(t, err)
				assert.True(t, response.Get("acknowledged").Bool())
			},
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
			testFunc: func(t *testing.T, response gjson.Result, err error) {
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusConflict, `{
			"status": 409,
			"code": 4001,
			"messages": [
				"Action HALT cannot be applied to seed execution cc89b714-d00a-4774-9c45-9497b5d9f8ef because of its current status: HALTING"
			],
			"timestamp": "2025-09-03T21:05:21.861757200Z"
			}`))
			},
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
			testFunc: func(t *testing.T, response gjson.Result, err error) {
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusNotFound, `{
			"status": 404,
			"code": 1003,
			"messages": [
				"Seed execution not found: cc89b714-d00a-4774-9c45-9497b5d9f8e3"
			],
			"timestamp": "2025-09-03T21:37:21.871825500Z"
			}`))
			},
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
			if err != nil {
				fmt.Println("UUID conversion failed: " + err.Error())
				return
			}
			ingestionSeedExecutionsClient := newSeedExecutionsClient(srv.URL, apiKey, seedId)

			executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
			if err != nil {
				fmt.Println("UUID conversion failed: " + err.Error())
				return
			}

			response, err := ingestionSeedExecutionsClient.Halt(executionId)
			tc.testFunc(t, response, err)
		})
	}
}

// Test_newEndpointsClient tests the constructor of endpointsClients.
func Test_seedExecutionsClient_Records(t *testing.T) {
	url := "http://localhost:8083/v2"
	apiKey := "Api Key"
	seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
	if err != nil {
		fmt.Println("UUID conversion failed: " + err.Error())
		return
	}
	ingestionSeedExecutionsClient := newSeedExecutionsClient(url, apiKey, seedId)

	executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
	if err != nil {
		fmt.Println("UUID conversion failed: " + err.Error())
		return
	}

	ingestionSeedExecutionRecordsClient := ingestionSeedExecutionsClient.Records(executionId)

	assert.Equal(t, apiKey, ingestionSeedExecutionRecordsClient.client.ApiKey)
	assert.Equal(t, url+"/seed/"+seedId.String()+"/execution/"+executionId.String()+"/record", ingestionSeedExecutionRecordsClient.client.client.BaseURL)
}

func Test_seedExecutionsClient_Jobs(t *testing.T) {
	url := "http://localhost:8083/v2"
	apiKey := "Api Key"
	seedId, err := uuid.Parse("2acd0a61-852c-4f38-af2b-9c84e152873e")
	if err != nil {
		fmt.Println("UUID conversion failed: " + err.Error())
		return
	}
	ingestionSeedExecutionsClient := newSeedExecutionsClient(url, apiKey, seedId)

	executionId, err := uuid.Parse("6b7f0b69-126f-49ab-b2ff-0a876f42e5ed")
	if err != nil {
		fmt.Println("UUID conversion failed: " + err.Error())
		return
	}

	ingestionSeedExecutionJobClient := ingestionSeedExecutionsClient.Jobs(executionId)

	assert.Equal(t, apiKey, ingestionSeedExecutionJobClient.client.ApiKey)
	assert.Equal(t, url+"/seed/"+seedId.String()+"/execution/"+executionId.String()+"/job", ingestionSeedExecutionJobClient.client.client.BaseURL)
}

// Test_seedExecutionsClient_Audit_HTTPResponseCases tests how the getter.GetAll() function behaves when receiving different HTTP responses and errors.
// It does not test if reading all the pages works.
func Test_seedExecutionsClient_Audit_HTTPResponseCases(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		response   string
		testFunc   func(t *testing.T, c crud)
	}{
		// Working cases
		{
			name:       "GetAll returns array",
			method:     http.MethodGet,
			path:       "/",
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
			testFunc: func(t *testing.T, c crud) {
				results, err := c.GetAll()
				require.NoError(t, err)
				assert.Len(t, results, 7)
			},
		},
		{
			name:       "GetAll returns no content",
			method:     http.MethodGet,
			path:       "/",
			statusCode: http.StatusNoContent,
			response:   `{"content": []}`,
			testFunc: func(t *testing.T, c crud) {
				results, err := c.GetAll()
				require.NoError(t, err)
				assert.Equal(t, []gjson.Result{}, results)
				assert.Len(t, results, 0)
			},
		},

		// Error cases
		{
			name:       "GetAll returns a 401 Unauthorized",
			method:     http.MethodGet,
			path:       "/",
			statusCode: http.StatusUnauthorized,
			response:   `{"error":"unauthorized"}`,
			testFunc: func(t *testing.T, c crud) {
				response, err := c.GetAll()
				assert.Equal(t, []gjson.Result(nil), response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusUnauthorized, []byte(`{"error":"unauthorized"}`)))
			},
		},
		{
			name:       "GetAll has no content field",
			method:     http.MethodGet,
			path:       "/",
			statusCode: http.StatusNoContent,
			response:   ``,
			testFunc: func(t *testing.T, c crud) {
				results, err := c.GetAll()
				require.NoError(t, err)
				assert.Len(t, results, 0)
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

// Test_seedExecutionsClient_Audit_ErrorInSecondPage tests when GetAll fails in a request while trying to get every content from every page.
func Test_seedExecutionsClient_Audit_ErrorInSecondPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			pageNumber, _ := strconv.Atoi(r.URL.Query().Get("page"))
			w.Header().Set("Content-Type", "application/json")
			if pageNumber > 0 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"Internal Server Error"}`))
			} else {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{
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
		],
		"pageable": {
			"page": 0,
			"size": 3,
			"sort": []
		},
		"totalSize": 7,
		"totalPages": 2,
		"empty": false,
		"size": 3,
		"offset": 0,
		"numberOfElements": 3,
		"pageNumber": 0
		}`))
			}
		}))
	t.Cleanup(srv.Close)

	c := crud{getter{newClient(srv.URL, "")}}

	response, err := c.GetAll()
	assert.Equal(t, []gjson.Result(nil), response)
	assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusInternalServerError, []byte(`{"error":"Internal Server Error"}`)))
}

// Test_seedExecutionsClient_Audit_NoContentInSecondPage tests what happens if one of the later pages returns No Content
func Test_seedExecutionsClient_Audit_NoContentInSecondPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		pageNumber, _ := strconv.Atoi(r.URL.Query().Get("page"))
		w.Header().Set("Content-Type", "application/json")
		if pageNumber > 0 {
			w.WriteHeader(http.StatusNoContent)
			_, _ = w.Write([]byte(`[]`))
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
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
		],
		"pageable": {
			"page": 0,
			"size": 3,
			"sort": []
		},
		"totalSize": 7,
		"totalPages": 2,
		"empty": false,
		"size": 3,
		"offset": 0,
		"numberOfElements": 3,
		"pageNumber": 0
		}`))
		}
	}))
	t.Cleanup(srv.Close)

	c := crud{getter{newClient(srv.URL, "")}}

	response, err := c.GetAll()
	require.NoError(t, err)
	assert.Len(t, response, 3)
}

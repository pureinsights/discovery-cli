package discovery

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Test_newBucketsClient tests the bucketsClient constructor.
func Test_newBucketsClient(t *testing.T) {
	url := "http://localhost:12020"
	apiKey := "Api Key"
	c := newBucketsClient(url, apiKey)

	assert.Equal(t, apiKey, c.client.ApiKey)
	assert.Equal(t, url+"/bucket", c.client.client.BaseURL)
}

// Test_bucketsClient_Create tests the bucketsClient.Create() function.
func Test_bucketsClient_Create(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		config           gjson.Result
		response         string
		expectedResponse gjson.Result
		err              error
	}{
		// Working case
		{
			name:       "Create works with full config",
			method:     http.MethodPost,
			path:       "/testBucket",
			statusCode: http.StatusCreated,
			config: gjson.Parse(`{
			"indices": [
				{
					"name": "indexTest",
					"fields": [{"author": "ASC"}],
					"unique": false
				}
			],
			"config": {}
			}`),
			response:         `{"acknowledged": true}`,
			expectedResponse: gjson.Parse(`{"acknowledged": true}`),
			err:              nil,
		},
		{
			name:             "Create works with empty config",
			method:           http.MethodPost,
			path:             "/testBucket",
			statusCode:       http.StatusCreated,
			config:           gjson.Parse(``),
			response:         `{"acknowledged": true}`,
			expectedResponse: gjson.Parse(`{"acknowledged": true}`),
			err:              nil,
		},

		// Error case
		{
			name:       "Create returns 409 Conflict",
			method:     http.MethodPost,
			path:       "/testBucket",
			statusCode: http.StatusConflict,
			config: gjson.Parse(`{
			"indices": [
				{
					"name": "indexTest",
					"fields": [{"author": "ASC"}],
					"unique": false
				}
			],
			"config": {}
			}`),
			response:         `{"acknowledged":false}`,
			expectedResponse: gjson.Result{},
			err:              Error{Status: http.StatusConflict, Body: gjson.Parse(`{"acknowledged":false}`)},
		},
		{
			name:       "Create returns 400 Bad Request",
			method:     http.MethodPost,
			path:       "/testBucket",
			statusCode: http.StatusBadRequest,
			config: gjson.Parse(`{
			"indices": [
				{
					"fields": [{"author": "ASC"}],
					"unique": false
				}
			]
			}`),
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"options.indices[0].name: must not be blank"
			],
			"timestamp": "2025-09-22T16:40:06.261655900Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"options.indices[0].name: must not be blank"
			],
			"timestamp": "2025-09-22T16:40:06.261655900Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, "/bucket"+tc.path, r.URL.Path)
				body, _ := io.ReadAll(r.Body)
				json := gjson.Parse(string(body))
				assert.Equal(t, tc.config, json)
			}))

			defer srv.Close()

			bucketsClient := newBucketsClient(srv.URL, "")

			response, err := bucketsClient.Create("testBucket", tc.config)
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

// Test_bucketsClient_Get tests the bucketsClient.Get() function.
func Test_bucketsClient_Get(t *testing.T) {
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
			name:             "Get works",
			method:           http.MethodGet,
			path:             "/testBucket",
			statusCode:       http.StatusOK,
			response:         `{"name":"testBucket","documentCount":{"STORE":3},"indices":[{"name":"authorIndex","fields":[{"author":"DESC"}],"unique":false}]}`,
			expectedResponse: gjson.Parse(`{"name":"testBucket","documentCount":{"STORE":3},"indices":[{"name":"authorIndex","fields":[{"author":"DESC"}],"unique":false}]}`),
			err:              nil,
		},

		// Error case
		{
			name:       "Get returns 404 Not Found",
			method:     http.MethodGet,
			path:       "/testBucket",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1002,
			"messages": [
				"The bucket 'testBucket' was not found."
			],
			"timestamp": "2025-09-08T23:05:32.202752400Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1002,
			"messages": [
				"The bucket 'testBucket' was not found."
			],
			"timestamp": "2025-09-08T23:05:32.202752400Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, "/bucket"+tc.path, r.URL.Path)
			}))

			defer srv.Close()

			bucketsClient := newBucketsClient(srv.URL, "")

			response, err := bucketsClient.Get("testBucket")
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

// Test_bucketsClient_GetAll tests the bucketsClient.GetAll() function.
func Test_bucketsClient_GetAll(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		response         string
		expectedResponse []string
		err              error
	}{
		// Working case
		{
			name:       "GetAll works",
			method:     http.MethodGet,
			path:       "",
			statusCode: http.StatusOK,
			response: `[
			"blogs",
			"blogsq",
			"wikis"
			]`,
			expectedResponse: []string{
				"blogs",
				"blogsq",
				"wikis",
			},
			err: nil,
		}, {
			name:             "GetAll returns no content",
			method:           http.MethodGet,
			path:             "",
			statusCode:       http.StatusNoContent,
			response:         ``,
			expectedResponse: []string{},
			err:              nil,
		},
		// Error case
		{
			name:             "GetAll returns an internal server error",
			method:           http.MethodGet,
			path:             "",
			statusCode:       http.StatusInternalServerError,
			response:         ``,
			expectedResponse: []string(nil),
			err:              Error{Status: http.StatusInternalServerError, Body: gjson.Result{}},
		},
		{
			name:             "GetAll returns a response that cannot be marshalled into an []string",
			method:           http.MethodGet,
			path:             "",
			statusCode:       http.StatusOK,
			response:         `{"message"} : "This cannot be marshalled."`,
			expectedResponse: []string(nil),
			err:              fmt.Errorf("invalid character '}' after object key"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, "/bucket"+tc.path, r.URL.Path)
			}))

			defer srv.Close()

			bucketsClient := newBucketsClient(srv.URL, "")

			response, err := bucketsClient.GetAll()
			assert.Equal(t, tc.expectedResponse, response)
			if tc.err == nil {
				require.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

// Test_bucketsClient_Delete tests the bucketsClient.Delete() function.
func Test_bucketsClient_Delete(t *testing.T) {
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
			name:             "Delete works",
			method:           http.MethodDelete,
			path:             "/testBucket",
			statusCode:       http.StatusOK,
			response:         `{"acknowledged":true}`,
			expectedResponse: gjson.Parse(`{"acknowledged":true}`),
			err:              nil,
		},

		// Error case
		{
			name:       "Delete returns 404 Not Found",
			method:     http.MethodDelete,
			path:       "/testBucket",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1002,
			"messages": [
				"The bucket 'testBucket' was not found."
			],
			"timestamp": "2025-09-08T23:05:32.202752400Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1002,
			"messages": [
				"The bucket 'testBucket' was not found."
			],
			"timestamp": "2025-09-08T23:05:32.202752400Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, "/bucket"+tc.path, r.URL.Path)
			}))

			defer srv.Close()

			bucketsClient := newBucketsClient(srv.URL, "")

			response, err := bucketsClient.Delete("testBucket")
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

// Test_bucketsClient_Purge tests the bucketsClient.Purge() function.
func Test_bucketsClient_Purge(t *testing.T) {
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
			name:             "Purge works",
			method:           http.MethodDelete,
			path:             "/testBucket/purge",
			statusCode:       http.StatusOK,
			response:         `{"acknowledged":true}`,
			expectedResponse: gjson.Parse(`{"acknowledged":true}`),
			err:              nil,
		},

		// Error case
		{
			name:       "Purge returns 404 Not Found",
			method:     http.MethodDelete,
			path:       "/testBucket/purge",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1002,
			"messages": [
				"The bucket 'testBucket' was not found."
			],
			"timestamp": "2025-09-08T23:05:32.202752400Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1002,
			"messages": [
				"The bucket 'testBucket' was not found."
			],
			"timestamp": "2025-09-08T23:05:32.202752400Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, "/bucket"+tc.path, r.URL.Path)
			}))

			defer srv.Close()

			bucketsClient := newBucketsClient(srv.URL, "")

			response, err := bucketsClient.Purge("testBucket")
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

// Test_bucketsClient_CreateIndex tests the bucketsClient.CreateIndex().
func Test_bucketsClient_CreateIndex(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		indexConfig      []string
		response         string
		expectedResponse gjson.Result
		err              error
	}{
		// Working case
		{
			name:             "Create Index works",
			method:           http.MethodPut,
			path:             "/testBucket/index/testIndex",
			statusCode:       http.StatusOK,
			response:         `{"acknowledged":true}`,
			expectedResponse: gjson.Parse(`{"acknowledged":true}`),
			indexConfig:      []string{`{"fieldName":"ASC"}`, `{"author":"DESC"}`},
			err:              nil,
		},

		// Error case
		{
			name:       "Create Index returns 404 Conflict",
			method:     http.MethodPut,
			path:       "/testBucket/index/testIndex",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1002,
			"messages": [
				"The bucket 'testBucket' was not found."
			],
			"timestamp": "2025-09-08T23:05:32.202752400Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1002,
			"messages": [
				"The bucket 'testBucket' was not found."
			],
			"timestamp": "2025-09-08T23:05:32.202752400Z"
			}`)},
			indexConfig: []string{`{"fieldName":"ASC"}`, `{"author":"DESC"}`},
		},
		{
			name:       "Create Index returns 400 Index already exists.",
			method:     http.MethodPut,
			path:       "/testBucket/index/testIndex",
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"An index with the same fields already exists"
			],
			"timestamp": "2025-09-08T23:51:00.869751600Z"
			}`,
			indexConfig:      []string{`{"fieldName":"ASC"}`, `{"author":"DESC"}`},
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"An index with the same fields already exists"
			],
			"timestamp": "2025-09-08T23:51:00.869751600Z"
			}`)},
		},
		{
			name:       "Create Index returns 400 Invalid JSON",
			method:     http.MethodPut,
			path:       "/testBucket/index/testIndex",
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3001,
			"messages": [
				"Invalid JSON: Cannot deserialize value of type "java.util.Map$Entry<java.lang.String,io.micronaut.data.model.Sort$Order$Direction>" from String value (token JsonToken.VALUE_STRING)\n at [Source: REDACTED (StreamReadFeature.INCLUDE_SOURCE_IN_LOCATION disabled); line: 1, column: 22] (through reference chain: java.util.ArrayList[1])"
			],
			"timestamp": "2025-09-09T00:01:57.445509100Z"
			}`,
			indexConfig:      []string{`{"fieldName":"ASC"}`, `"author":"DESC"}`},
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3001,
			"messages": [
				"Invalid JSON: Cannot deserialize value of type "java.util.Map$Entry<java.lang.String,io.micronaut.data.model.Sort$Order$Direction>" from String value (token JsonToken.VALUE_STRING)\n at [Source: REDACTED (StreamReadFeature.INCLUDE_SOURCE_IN_LOCATION disabled); line: 1, column: 22] (through reference chain: java.util.ArrayList[1])"
			],
			"timestamp": "2025-09-09T00:01:57.445509100Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, "/bucket"+tc.path, r.URL.Path)
				if tc.err == nil {
					body, _ := io.ReadAll(r.Body)
					jsonArray := gjson.ParseBytes(body).Array()
					for _, index := range jsonArray {
						assert.Contains(t, tc.indexConfig, index.Raw)
					}
				}
			}))

			defer srv.Close()

			bucketsClient := newBucketsClient(srv.URL, "")

			indices := []gjson.Result{gjson.Parse(tc.indexConfig[0]), gjson.Parse(tc.indexConfig[1])}
			response, err := bucketsClient.CreateIndex("testBucket", "testIndex", indices)
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

// Test_bucketsClient_DeleteIndex tests the bucketsClient.DeleteIndex() function.
func Test_bucketsClient_DeleteIndex(t *testing.T) {
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
			name:             "Delete index works",
			method:           http.MethodDelete,
			path:             "/testBucket/index/testIndex",
			statusCode:       http.StatusOK,
			response:         `{"acknowledged":true}`,
			expectedResponse: gjson.Parse(`{"acknowledged":true}`),
			err:              nil,
		},

		// Error case
		{
			name:       "Delete index returns 404 Bucket Not Found",
			method:     http.MethodDelete,
			path:       "/testBucket/index/testIndex",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1002,
			"messages": [
				"The bucket 'testBucket' was not found."
			],
			"timestamp": "2025-09-08T23:05:32.202752400Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1002,
			"messages": [
				"The bucket 'testBucket' was not found."
			],
			"timestamp": "2025-09-08T23:05:32.202752400Z"
			}`)},
		},
		{
			name:       "Delete index returns 404 Index Not Found",
			method:     http.MethodDelete,
			path:       "/testBucket/index/testIndex",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1003,
			"messages": [
				"The index 'testIndex' was not found"
			],
			"timestamp": "2025-09-09T00:07:39.219014900Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"The index 'testIndex' was not found"
			],
			"timestamp": "2025-09-09T00:07:39.219014900Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, "/bucket"+tc.path, r.URL.Path)
			}))

			defer srv.Close()

			bucketsClient := newBucketsClient(srv.URL, "")

			response, err := bucketsClient.DeleteIndex("testBucket", "testIndex")
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

// Test_newContentClient tests the contentClient constructor.
func Test_newContentClient(t *testing.T) {
	url := "http://localhost:12020"
	apiKey := "Api Key"
	bucketName := "testBucket"
	c := newContentClient(url, apiKey, bucketName)

	assert.Equal(t, apiKey, c.client.ApiKey)
	assert.Equal(t, url+"/content/"+bucketName, c.client.client.BaseURL)
}

// Test_contentClient_Store tests the contentClient.Store() function.
func Test_contentClient_Store(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		response         string
		expectedResponse gjson.Result
		bucketName       string
		contentId        string
		parentId         string
		err              error
	}{
		// Working case
		{
			name:       "Store works with bucketName, parentId and contentId",
			method:     http.MethodPost,
			path:       "/c28db957887e1aae75e7ab1dd0fd34e9",
			statusCode: http.StatusCreated,
			response: `{
			"id": "c28db957887e1aae75e7ab1dd0fd34e9",
			"creationTimestamp": "2025-09-09T00:44:09Z",
			"lastUpdatedTimestamp": "2025-09-09T14:51:27Z",
			"parentId": "d758c733466967ea6f13b20bcbfcebb5",
			"action": "STORE",
			"checksum": "6b65188c0a7878ad4ba2d8f3e8109b7e",
			"content": {
				"reference": "https://pureinsights.com/blog/2024/five-common-challenges-when-implementing-rag-retrieval-augmented-generation/",
				"title": "5 Challenges Implementing Retrieval Augmented Generation (RAG) - Pureinsights",
				"description": "A blog on 5 common challenges when implementing RAG (Retrieval Augmented Generation) and possible solutions for search applications.",
				"author": "Matt Willsmore"
			},
			"transaction": "68c03eef1816a507481717b7"
			}`,
			expectedResponse: gjson.Parse(`{
			"id": "c28db957887e1aae75e7ab1dd0fd34e9",
			"creationTimestamp": "2025-09-09T00:44:09Z",
			"lastUpdatedTimestamp": "2025-09-09T14:51:27Z",
			"parentId": "d758c733466967ea6f13b20bcbfcebb5",
			"action": "STORE",
			"checksum": "6b65188c0a7878ad4ba2d8f3e8109b7e",
			"content": {
				"reference": "https://pureinsights.com/blog/2024/five-common-challenges-when-implementing-rag-retrieval-augmented-generation/",
				"title": "5 Challenges Implementing Retrieval Augmented Generation (RAG) - Pureinsights",
				"description": "A blog on 5 common challenges when implementing RAG (Retrieval Augmented Generation) and possible solutions for search applications.",
				"author": "Matt Willsmore"
			},
			"transaction": "68c03eef1816a507481717b7"
			}`),
			bucketName: "testBucket",
			contentId:  "c28db957887e1aae75e7ab1dd0fd34e9",
			parentId:   "d758c733466967ea6f13b20bcbfcebb5",
			err:        nil,
		},
		{
			name:       "Store works with no parent ID",
			method:     http.MethodPost,
			path:       "/c28db957887e1aae75e7ab1dd0fd34e9",
			statusCode: http.StatusCreated,
			response: `{
			"id": "c28db957887e1aae75e7ab1dd0fd34e9",
			"creationTimestamp": "2025-09-09T00:40:26Z",
			"lastUpdatedTimestamp": "2025-09-09T00:41:40Z",
			"action": "STORE",
			"checksum": "6b65188c0a7878ad4ba2d8f3e8109b7e",
			"content": {
				"reference": "https://pureinsights.com/blog/2024/five-common-challenges-when-implementing-rag-retrieval-augmented-generation/",
				"title": "5 Challenges Implementing Retrieval Augmented Generation (RAG) - Pureinsights",
				"description": "A blog on 5 common challenges when implementing RAG (Retrieval Augmented Generation) and possible solutions for search applications.",
				"author": "Matt Willsmore"
			},
			"transaction": "68bf77c41816a507481717ac"
			}`,
			expectedResponse: gjson.Parse(`{
			"id": "c28db957887e1aae75e7ab1dd0fd34e9",
			"creationTimestamp": "2025-09-09T00:40:26Z",
			"lastUpdatedTimestamp": "2025-09-09T00:41:40Z",
			"action": "STORE",
			"checksum": "6b65188c0a7878ad4ba2d8f3e8109b7e",
			"content": {
				"reference": "https://pureinsights.com/blog/2024/five-common-challenges-when-implementing-rag-retrieval-augmented-generation/",
				"title": "5 Challenges Implementing Retrieval Augmented Generation (RAG) - Pureinsights",
				"description": "A blog on 5 common challenges when implementing RAG (Retrieval Augmented Generation) and possible solutions for search applications.",
				"author": "Matt Willsmore"
			},
			"transaction": "68bf77c41816a507481717ac"
			}`),
			bucketName: "testBucket",
			contentId:  "c28db957887e1aae75e7ab1dd0fd34e9",
			parentId:   "",
			err:        nil,
		},

		// Error case
		{
			name:       "Store returns 400 Invalid JSON",
			method:     http.MethodPost,
			path:       "/c28db957887e1aae75e7ab1dd0fd34e9",
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3001,
			"messages": [
				"Invalid JSON: Unrecognized token 'test': was expecting (JSON String, Number, Array, Object or token 'null', 'true' or 'false')\n at [Source: REDACTED ('StreamReadFeature.INCLUDE_SOURCE_IN_LOCATION' disabled); line: 1, column: 5]"
			],
			"timestamp": "2025-09-09T00:54:42.457812Z"
			}`,
			bucketName:       "testBucket",
			contentId:        "c28db957887e1aae75e7ab1dd0fd34e9",
			parentId:         "d758c733466967ea6f13b20bcbfcebb5",
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3001,
			"messages": [
				"Invalid JSON: Unrecognized token 'test': was expecting (JSON String, Number, Array, Object or token 'null', 'true' or 'false')\n at [Source: REDACTED ('StreamReadFeature.INCLUDE_SOURCE_IN_LOCATION' disabled); line: 1, column: 5]"
			],
			"timestamp": "2025-09-09T00:54:42.457812Z"
			}`)},
		},
		{
			name:       "Store returns 400 Blank input values",
			method:     http.MethodPost,
			path:       "/   ",
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"bucketName: must not be blank",
				"contentId: must not be blank"
			],
			"timestamp": "2025-09-09T14:31:13.275303600Z"
			}`,
			bucketName:       "   ",
			contentId:        "   ",
			parentId:         "   ",
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"bucketName: must not be blank",
				"contentId: must not be blank"
			],
			"timestamp": "2025-09-09T14:31:13.275303600Z"
			}`)},
		},
		{
			name:       "Store returns 413 Request Entity Too Large",
			method:     http.MethodPost,
			path:       "/c28db957887e1aae75e7ab1dd0fd34e9",
			statusCode: http.StatusRequestEntityTooLarge,
			response: `{
			"status": 413,
			"code": 3001,
			"messages": [
				"Request Entity is too large"
			],
			"timestamp": "2025-09-09T00:54:42.457812Z"
			}`,
			bucketName:       "testBucket",
			contentId:        "c28db957887e1aae75e7ab1dd0fd34e9",
			parentId:         "d758c733466967ea6f13b20bcbfcebb5",
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusRequestEntityTooLarge, Body: gjson.Parse(`{
			"status": 413,
			"code": 3001,
			"messages": [
				"Request Entity is too large"
			],
			"timestamp": "2025-09-09T00:54:42.457812Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jsonString := `{
				"reference": "https://pureinsights.com/blog/2024/five-common-challenges-when-implementing-rag-retrieval-augmented-generation/",
				"title": "5 Challenges Implementing Retrieval Augmented Generation (RAG) - Pureinsights",
				"description": "A blog on 5 common challenges when implementing RAG (Retrieval Augmented Generation) and possible solutions for search applications.",
				"author": "Matt Willsmore"
			}`
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, "/content/"+tc.bucketName+tc.path, r.URL.Path)
				body, _ := io.ReadAll(r.Body)
				assert.Equal(t, jsonString, string(body))
				if tc.parentId != "" {
					assert.Equal(t, tc.parentId, r.URL.Query().Get("parentId"))
				}
			}))

			defer srv.Close()

			json := gjson.Parse(jsonString)
			contentClient := newContentClient(srv.URL, "", tc.bucketName)

			response, err := contentClient.Store(tc.contentId, tc.parentId, json)
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

// Test_contentClient_Get tests the contentClient.Get() function.
func Test_contentClient_Get(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		response         string
		expectedResponse gjson.Result
		bucketName       string
		contentId        string
		getOptions       []stagingGetContentOption
		err              error
	}{
		// Working case
		{
			name:       "Get works",
			method:     http.MethodGet,
			path:       "/c28db957887e1aae75e7ab1dd0fd34e9",
			statusCode: http.StatusOK,
			response: `{
			"id": "c28db957887e1aae75e7ab1dd0fd34e9",
			"creationTimestamp": "2025-09-09T00:44:09Z",
			"lastUpdatedTimestamp": "2025-09-09T14:51:27Z",
			"parentId": "d758c733466967ea6f13b20bcbfcebb5",
			"action": "STORE",
			"checksum": "6b65188c0a7878ad4ba2d8f3e8109b7e",
			"content": {
				"reference": "https://pureinsights.com/blog/2024/five-common-challenges-when-implementing-rag-retrieval-augmented-generation/",
				"title": "5 Challenges Implementing Retrieval Augmented Generation (RAG) - Pureinsights",
				"description": "A blog on 5 common challenges when implementing RAG (Retrieval Augmented Generation) and possible solutions for search applications.",
				"author": "Matt Willsmore"
			},
			"transaction": "68c03eef1816a507481717b7"
			}`,
			expectedResponse: gjson.Parse(`{
			"id": "c28db957887e1aae75e7ab1dd0fd34e9",
			"creationTimestamp": "2025-09-09T00:44:09Z",
			"lastUpdatedTimestamp": "2025-09-09T14:51:27Z",
			"parentId": "d758c733466967ea6f13b20bcbfcebb5",
			"action": "STORE",
			"checksum": "6b65188c0a7878ad4ba2d8f3e8109b7e",
			"content": {
				"reference": "https://pureinsights.com/blog/2024/five-common-challenges-when-implementing-rag-retrieval-augmented-generation/",
				"title": "5 Challenges Implementing Retrieval Augmented Generation (RAG) - Pureinsights",
				"description": "A blog on 5 common challenges when implementing RAG (Retrieval Augmented Generation) and possible solutions for search applications.",
				"author": "Matt Willsmore"
			},
			"transaction": "68c03eef1816a507481717b7"
			}`),
			bucketName: "testBucket",
			contentId:  "c28db957887e1aae75e7ab1dd0fd34e9",
			err:        nil,
			getOptions: []stagingGetContentOption{WithContentAction("STORE"), WithIncludeProjections([]string{"author", "header"}), WithExcludeProjections([]string{"author", "link"})},
		},

		// Error case
		{
			name:       "Get returns 400 Blank input values",
			method:     http.MethodGet,
			path:       "/c28db957887e1aae75e7ab1dd0fd34e9",
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"bucketName: must not be blank"
			],
			"timestamp": "2025-09-09T14:31:13.275303600Z"
			}`,
			bucketName:       "   ",
			contentId:        "c28db957887e1aae75e7ab1dd0fd34e9",
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"bucketName: must not be blank"
			],
			"timestamp": "2025-09-09T14:31:13.275303600Z"
			}`)},
			getOptions: []stagingGetContentOption(nil),
		},
		{
			name:       "Get returns 404 Not found",
			method:     http.MethodGet,
			path:       "/c28db957887e1aae75e7ab1dd0fd34e9",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1003,
			"messages": [
				"Document with id 'c28db957887e1aae75e7ab1dd0fd34e9' on bucket 'testBucket' not found."
			],
			"timestamp": "2025-09-09T15:47:26.883457300Z"
			}`,
			bucketName:       "testBucket",
			contentId:        "c28db957887e1aae75e7ab1dd0fd34e9",
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Document with id 'c28db957887e1aae75e7ab1dd0fd34e9' on bucket 'testBucket' not found."
			],
			"timestamp": "2025-09-09T15:47:26.883457300Z"
			}`)},
			getOptions: []stagingGetContentOption(nil),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, "/content/"+tc.bucketName+tc.path, r.URL.Path)
				queryParams := make(map[string][]string)
				for _, opt := range tc.getOptions {
					opt(&queryParams)
				}
				for k, v := range queryParams {
					assert.Equal(t, v, r.URL.Query()[k])
				}
			}))

			defer srv.Close()

			contentClient := newContentClient(srv.URL, "", tc.bucketName)

			response, err := contentClient.Get(tc.contentId, tc.getOptions...)
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

// Test_contentClient_Delete tests the contentClient.Delete() function.
func Test_contentClient_Delete(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		response         string
		expectedResponse gjson.Result
		bucketName       string
		contentId        string
		err              error
	}{
		// Working case
		{
			name:             "Delete works",
			method:           http.MethodDelete,
			path:             "/c28db957887e1aae75e7ab1dd0fd34e9",
			statusCode:       http.StatusOK,
			response:         `{"acknowledged":true}`,
			expectedResponse: gjson.Parse(`{"acknowledged":true}`),
			bucketName:       "testBucket",
			contentId:        "c28db957887e1aae75e7ab1dd0fd34e9",
			err:              nil,
		},

		// Error case
		{
			name:       "Delete returns 400 Blank input values",
			method:     http.MethodDelete,
			path:       "/  ",
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"contentId: must not be blank"
			],
			"timestamp": "2025-09-09T14:31:13.275303600Z"
			}`,
			expectedResponse: gjson.Result{},
			bucketName:       "testBucket",
			contentId:        "  ",
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"contentId: must not be blank"
			],
			"timestamp": "2025-09-09T14:31:13.275303600Z"
			}`)},
		},
		{
			name:       "Delete returns 404 Not found",
			method:     http.MethodDelete,
			path:       "/c28db957887e1aae75e7ab1dd0fd34e9",
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1003,
			"messages": [
				"Document with id 'c28db957887e1aae75e7ab1dd0fd34e9' on bucket 'testBucket' not found."
			],
			"timestamp": "2025-09-09T15:47:26.883457300Z"
			}`,
			expectedResponse: gjson.Result{},
			bucketName:       "testBucket",
			contentId:        "c28db957887e1aae75e7ab1dd0fd34e9",
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Document with id 'c28db957887e1aae75e7ab1dd0fd34e9' on bucket 'testBucket' not found."
			],
			"timestamp": "2025-09-09T15:47:26.883457300Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, "/content/"+tc.bucketName+tc.path, r.URL.Path)
			}))

			defer srv.Close()

			contentClient := newContentClient(srv.URL, "", tc.bucketName)

			response, err := contentClient.Delete(tc.contentId)
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

// Test_contentClient_DeleteMany tests the contentClient.DeleteMany() function.
func Test_contentClient_DeleteMany(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		statusCode       int
		response         string
		expectedResponse gjson.Result
		bucketName       string
		parentId         string
		filters          string
		err              error
	}{
		// Working case
		{
			name:             "Delete Many works with parentId and no filters",
			method:           http.MethodDelete,
			statusCode:       http.StatusOK,
			response:         `{"acknowledged":true}`,
			expectedResponse: gjson.Parse(`{"acknowledged":true}`),
			bucketName:       "testBucket",
			parentId:         "d758c733466967ea6f13b20bcbfcebb5",
			filters:          ``,
			err:              nil,
		},
		{
			name:             "Delete Many works with no parentId and filters",
			method:           http.MethodDelete,
			statusCode:       http.StatusOK,
			response:         `{"acknowledged":true}`,
			expectedResponse: gjson.Parse(`{"acknowledged":true}`),
			bucketName:       "testBucket",
			parentId:         "",
			filters: `{
			"equals": {
				"field": "author",
				"value": "Martin Bayton",
				"normalize": true
			}
			}`,
			err: nil,
		},

		// Error case
		{
			name:       "Delete returns 400 Blank input values",
			method:     http.MethodDelete,
			statusCode: http.StatusBadRequest,
			response: `{
			"status": 400,
			"code": 3002,
			"messages": [
				"bucketName: must not be blank"
			],
			"timestamp": "2025-09-09T14:31:13.275303600Z"
			}`,
			bucketName:       "  ",
			parentId:         "d758c733466967ea6f13b20bcbfcebb5",
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"bucketName: must not be blank"
			],
			"timestamp": "2025-09-09T14:31:13.275303600Z"
			}`)},
		},
		{
			name:       "Delete returns 404 Not found",
			method:     http.MethodDelete,
			statusCode: http.StatusNotFound,
			response: `{
			"status": 404,
			"code": 1002,
			"messages": [
				"The bucket 'testBucket' was not found."
			],
			"timestamp": "2025-09-09T16:35:57.753512900Z"
			}`,
			bucketName:       "testBucket",
			parentId:         "c28db957887e1aae75e7ab1dd0fd34e9",
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
			"status": 404,
			"code": 1002,
			"messages": [
				"The bucket 'testBucket' was not found."
			],
			"timestamp": "2025-09-09T16:35:57.753512900Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, "/content/"+tc.bucketName, r.URL.Path)
				if tc.filters != "" {
					body, _ := io.ReadAll(r.Body)
					assert.Equal(t, tc.filters, string(body))
				}
				if tc.parentId != "" {
					assert.Equal(t, tc.parentId, r.URL.Query().Get("parentId"))
				}
			}))

			defer srv.Close()

			contentClient := newContentClient(srv.URL, "", tc.bucketName)

			response, err := contentClient.DeleteMany(tc.parentId, gjson.Parse(tc.filters))
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

// Test_scrollWithPagination_HTTPResponseCases tests how the scrollWithPagination() function behaves with various HTTP responses.
func Test_scrollWithPagination_HTTPResponseCases(t *testing.T) {
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
			name:        "scrollWithPagination returns no content",
			method:      http.MethodPost,
			path:        "/",
			statusCode:  http.StatusNoContent,
			response:    `{"content": []}`,
			expectedLen: 0,
			err:         nil,
		},
		{
			name:        "scrollWithPagination has no content field",
			method:      http.MethodPost,
			path:        "/",
			statusCode:  http.StatusNoContent,
			response:    ``,
			expectedLen: 0,
			err:         nil,
		},

		// Error cases
		{
			name:       "scrollWithPagination returns a 401 Unauthorized",
			method:     http.MethodPost,
			path:       "/",
			statusCode: http.StatusUnauthorized,
			response:   `{"error":"unauthorized"}`,
			err:        Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, tc.path, r.URL.Path)
			}))

			defer srv.Close()

			c := newClient(srv.URL, "")
			results, err := scrollWithPagination(c, tc.method, "")
			if tc.err == nil {
				require.NoError(t, err)
				assert.Len(t, results, tc.expectedLen)
			} else {
				assert.Equal(t, []gjson.Result(nil), results)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

// Test_scrollWithPagination_ErrorInSecondPage tests when scrollWithPagination fails in a request while trying to get every content from every page.
func Test_scrollWithPagination_ErrorInSecondPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/content/my-bucket/scroll", r.URL.Path)
			token := r.URL.Query().Get("token")
			w.Header().Set("Content-Type", "application/json")
			if token == "694eb7f378aedc7a163da907" {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"Internal Server Error"}`))
			} else {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{
			"token": "694eb7f378aedc7a163da907",
			"content": [
                  {
                          "id": "1",
                          "creationTimestamp": "2025-12-26T16:28:38Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:28:38Z",
                          "action": "STORE",
                          "checksum": "58b3d1b06729f1491373b97fd8287ae1",
                          "content": {
                                  "_id": "5625c64483bef0d48e9ad91aca9b2f94",
                                  "link": "https://pureinsights.com/blog/2024/pureinsights-named-mongodbs-2024-ai-partner-of-the-year/",
                                  "author": "Graham Gillen",
                                  "header": "Pureinsights Named MongoDB's 2024 AI Partner of the Year - Pureinsights: PRESS RELEASE - Pureinsights named MongoDB's Service AI Partner of the Year for 2024 and also joins the MongoDB AI Application Program (MAAP)."
                          },
                          "transaction": "694eb7b678aedc7a163da8ff"
                  },
                  {
                          "id": "2",
                          "creationTimestamp": "2025-12-26T16:28:46Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:28:46Z",
                          "action": "STORE",
                          "checksum": "b76292db9fd1c7aef145512dce131f4d",
                          "content": {
                                  "_id": "768b0a3bcee501dc624484ba8a0d7f6d",
                                  "link": "https://pureinsights.com/blog/2024/five-common-challenges-when-implementing-rag-retrieval-augmented-generation/",
                                  "author": "Matt Willsmore",
                                  "header": "5 Challenges Implementing Retrieval Augmented Generation (RAG) - Pureinsights: A blog on 5 common challenges when implementing RAG (Retrieval Augmented Generation) and possible solutions for search applications."
                          },
                          "transaction": "694eb7be78aedc7a163da900"
                  },
                  {
                          "id": "3",
                          "creationTimestamp": "2025-12-26T16:28:54Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:28:54Z",
                          "action": "STORE",
                          "checksum": "cbffeeba8f4739650ae048fb382c8870",
                          "content": {
                                  "_id": "d758c733466967ea6f13b20bcbfcebb5",
                                  "link": "https://pureinsights.com/blog/2024/modernizing-search-with-generative-ai/",
                                  "author": "Martin Bayton",
                                  "header": "Modernizing Search with Generative AI - Pureinsights: Blog: why you should implement Retrieval-Augmented Generation (RAG) and how platforms like Pureinsights Discovery streamline the process."
                          },
                          "transaction": "694eb7c678aedc7a163da901"
                  }
          ],
		  "empty": false
			}`))
			}
		}))
	t.Cleanup(srv.Close)

	c := newClient(srv.URL, "")
	response, err := scrollWithPagination(c, http.MethodPost, "/content/my-bucket/scroll")
	assert.Equal(t, []gjson.Result(nil), response)
	var errStruct Error
	require.ErrorAs(t, err, &errStruct)
	assert.EqualError(t, err, Error{Status: http.StatusInternalServerError, Body: gjson.Parse(`{"error":"Internal Server Error"}`)}.Error())
}

// Test_scrollWithPagination_RestyReturnsError tests what happens when the Resty client fails to execute the request.
func Test_scrollWithPagination_RestyReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	base := srv.URL
	srv.Close()

	c := newClient(base, "")
	response, err := scrollWithPagination(c, http.MethodPost, "/down")
	require.Error(t, err)
	assert.Equal(t, response, []gjson.Result(nil))
	assert.Contains(t, err.Error(), base+"/down")
}

// Test_scrollWithPagination_ContentInSecondPage tests that the scrollWithPagination() function
// can successfully get all content when there are two pages with content in them.
func Test_scrollWithPagination_ContentInSecondPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/content/my-bucket/scroll", r.URL.Path)
		token := r.URL.Query().Get("token")
		w.Header().Set("Content-Type", "application/json")
		switch token {
		case "694eb7f378aedc7a163da908":
			w.WriteHeader(http.StatusNoContent)
			_, _ = w.Write([]byte(`[]`))
		case "694eb7f378aedc7a163da907":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			"token": "694eb7f378aedc7a163da908",
			"content": [
                  {
                          "id": "4",
                          "creationTimestamp": "2025-12-26T16:28:59Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:28:59Z",
                          "action": "STORE",
                          "checksum": "855609b26c318a627760fd36d2d6fe8f",
                          "content": {
                                  "_id": "4e7c8a47efd829ef7f710d64da661786",
                                  "link": "https://pureinsights.com/blog/2024/kmworld-2024-key-takeaways-from-the-exhibit-hall/",
                                  "author": "Graham Gillen",
                                  "header": "KMWorld 2024: Key Takeaways from the Exhibit Hall - Pureinsights: Key insights from KMWorld 2024: AI's impact on knowledge management, standout vendors, and challenges for traditional players adapting to AI."
                          },
                          "transaction": "694eb7cb78aedc7a163da902"
                  },
                  {
                          "id": "5",
                          "creationTimestamp": "2025-12-26T16:29:05Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:29:05Z",
                          "action": "STORE",
                          "checksum": "855609b26c318a627760fd36d2d6fe8f",
                          "content": {
                                  "_id": "b1e3e4f42c0818b1580e306eb776d4a1",
                                  "link": "https://pureinsights.com/blog/2024/google-unveils-ai-enhanced-search-features-at-2024-io-conference/",
                                  "author": "Martin Bayton",
                                  "header": "Google Unveils AI-Enhanced Search Features at I/O Conference - Pureinsights: Google I/O 2024 Developer Conference key takeaways, including AI-generated summaries and other features for search."
                          },
                          "transaction": "694eb7d178aedc7a163da903"
                  },
                  {
                          "id": "6",
                          "creationTimestamp": "2025-12-26T16:29:12Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:29:12Z",
                          "action": "STORE",
                          "checksum": "228cc56c873a457041454280c448b4e3",
                          "content": {
                                  "_id": "232638a332048c4cb159f8cf6636507f",
                                  "link": "https://pureinsights.com/blog/2025/7-tech-trends-in-ai-and-search-for-2025/",
                                  "author": "Phil Lewis",
                                  "header": "7 Tech Trends in AI and Search for 2025 - Pureinsights: 7 Tech Trends is AI and Search for 2025 - presented by Pureinsights CTO, Phil Lewis. A blog about key trends to look for in the coming year."
                          },
                          "transaction": "694eb7d878aedc7a163da904"
                  }
          ],
		  "empty": false
			}`))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			"token": "694eb7f378aedc7a163da907",
			"content": [
                  {
                          "id": "1",
                          "creationTimestamp": "2025-12-26T16:28:38Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:28:38Z",
                          "action": "STORE",
                          "checksum": "58b3d1b06729f1491373b97fd8287ae1",
                          "content": {
                                  "_id": "5625c64483bef0d48e9ad91aca9b2f94",
                                  "link": "https://pureinsights.com/blog/2024/pureinsights-named-mongodbs-2024-ai-partner-of-the-year/",
                                  "author": "Graham Gillen",
                                  "header": "Pureinsights Named MongoDB's 2024 AI Partner of the Year - Pureinsights: PRESS RELEASE - Pureinsights named MongoDB's Service AI Partner of the Year for 2024 and also joins the MongoDB AI Application Program (MAAP)."
                          },
                          "transaction": "694eb7b678aedc7a163da8ff"
                  },
                  {
                          "id": "2",
                          "creationTimestamp": "2025-12-26T16:28:46Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:28:46Z",
                          "action": "STORE",
                          "checksum": "b76292db9fd1c7aef145512dce131f4d",
                          "content": {
                                  "_id": "768b0a3bcee501dc624484ba8a0d7f6d",
                                  "link": "https://pureinsights.com/blog/2024/five-common-challenges-when-implementing-rag-retrieval-augmented-generation/",
                                  "author": "Matt Willsmore",
                                  "header": "5 Challenges Implementing Retrieval Augmented Generation (RAG) - Pureinsights: A blog on 5 common challenges when implementing RAG (Retrieval Augmented Generation) and possible solutions for search applications."
                          },
                          "transaction": "694eb7be78aedc7a163da900"
                  },
                  {
                          "id": "3",
                          "creationTimestamp": "2025-12-26T16:28:54Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:28:54Z",
                          "action": "STORE",
                          "checksum": "cbffeeba8f4739650ae048fb382c8870",
                          "content": {
                                  "_id": "d758c733466967ea6f13b20bcbfcebb5",
                                  "link": "https://pureinsights.com/blog/2024/modernizing-search-with-generative-ai/",
                                  "author": "Martin Bayton",
                                  "header": "Modernizing Search with Generative AI - Pureinsights: Blog: why you should implement Retrieval-Augmented Generation (RAG) and how platforms like Pureinsights Discovery streamline the process."
                          },
                          "transaction": "694eb7c678aedc7a163da901"
                  }
          ],
		  "empty": false
			}`))
		}
	}))
	t.Cleanup(srv.Close)

	c := newClient(srv.URL, "")
	response, err := scrollWithPagination(c, http.MethodPost, "/content/my-bucket/scroll")
	require.NoError(t, err)
	assert.Len(t, response, 6)
}

// Test_contentClient_Scroll tests the contentClient.Scroll() function.
func Test_contentClient_Scroll(t *testing.T) {
	body := `{"filters":{
	"equals": {
		"field": "author",
		"value": "Martin Bayton",
		"normalize": true
	}
},"fields":{
    "includes": [
		"author",
		"header"
	]
}}`
	filters := `{
	"equals": {
		"field": "author",
		"value": "Martin Bayton",
		"normalize": true
	}
}`
	projections := `{
    "includes": [
		"author",
		"header"
	]
}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		requestBody, _ := io.ReadAll(r.Body)
		assert.Equal(t, gjson.Parse(body), gjson.Parse(string(requestBody)))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "/content/my-bucket/scroll", r.URL.Path)
		assert.Equal(t, "3", r.URL.Query().Get("size"))
		token := r.URL.Query().Get("token")
		w.Header().Set("Content-Type", "application/json")
		switch token {
		case "694eb7f378aedc7a163da908":
			w.WriteHeader(http.StatusNoContent)
			_, _ = w.Write([]byte(`[]`))
		case "694eb7f378aedc7a163da907":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			"token": "694eb7f378aedc7a163da908",
			"content": [
                  {
                          "id": "4",
                          "creationTimestamp": "2025-12-26T16:28:59Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:28:59Z",
                          "action": "STORE",
                          "checksum": "855609b26c318a627760fd36d2d6fe8f",
                          "content": {
                                  "_id": "4e7c8a47efd829ef7f710d64da661786",
                                  "link": "https://pureinsights.com/blog/2024/kmworld-2024-key-takeaways-from-the-exhibit-hall/",
                                  "author": "Graham Gillen",
                                  "header": "KMWorld 2024: Key Takeaways from the Exhibit Hall - Pureinsights: Key insights from KMWorld 2024: AI's impact on knowledge management, standout vendors, and challenges for traditional players adapting to AI."
                          },
                          "transaction": "694eb7cb78aedc7a163da902"
                  },
                  {
                          "id": "5",
                          "creationTimestamp": "2025-12-26T16:29:05Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:29:05Z",
                          "action": "STORE",
                          "checksum": "855609b26c318a627760fd36d2d6fe8f",
                          "content": {
                                  "_id": "b1e3e4f42c0818b1580e306eb776d4a1",
                                  "link": "https://pureinsights.com/blog/2024/google-unveils-ai-enhanced-search-features-at-2024-io-conference/",
                                  "author": "Martin Bayton",
                                  "header": "Google Unveils AI-Enhanced Search Features at I/O Conference - Pureinsights: Google I/O 2024 Developer Conference key takeaways, including AI-generated summaries and other features for search."
                          },
                          "transaction": "694eb7d178aedc7a163da903"
                  },
                  {
                          "id": "6",
                          "creationTimestamp": "2025-12-26T16:29:12Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:29:12Z",
                          "action": "STORE",
                          "checksum": "228cc56c873a457041454280c448b4e3",
                          "content": {
                                  "_id": "232638a332048c4cb159f8cf6636507f",
                                  "link": "https://pureinsights.com/blog/2025/7-tech-trends-in-ai-and-search-for-2025/",
                                  "author": "Phil Lewis",
                                  "header": "7 Tech Trends in AI and Search for 2025 - Pureinsights: 7 Tech Trends is AI and Search for 2025 - presented by Pureinsights CTO, Phil Lewis. A blog about key trends to look for in the coming year."
                          },
                          "transaction": "694eb7d878aedc7a163da904"
                  }
          ],
		  "empty": false
			}`))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			"token": "694eb7f378aedc7a163da907",
			"content": [
                  {
                          "id": "1",
                          "creationTimestamp": "2025-12-26T16:28:38Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:28:38Z",
                          "action": "STORE",
                          "checksum": "58b3d1b06729f1491373b97fd8287ae1",
                          "content": {
                                  "_id": "5625c64483bef0d48e9ad91aca9b2f94",
                                  "link": "https://pureinsights.com/blog/2024/pureinsights-named-mongodbs-2024-ai-partner-of-the-year/",
                                  "author": "Graham Gillen",
                                  "header": "Pureinsights Named MongoDB's 2024 AI Partner of the Year - Pureinsights: PRESS RELEASE - Pureinsights named MongoDB's Service AI Partner of the Year for 2024 and also joins the MongoDB AI Application Program (MAAP)."
                          },
                          "transaction": "694eb7b678aedc7a163da8ff"
                  },
                  {
                          "id": "2",
                          "creationTimestamp": "2025-12-26T16:28:46Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:28:46Z",
                          "action": "STORE",
                          "checksum": "b76292db9fd1c7aef145512dce131f4d",
                          "content": {
                                  "_id": "768b0a3bcee501dc624484ba8a0d7f6d",
                                  "link": "https://pureinsights.com/blog/2024/five-common-challenges-when-implementing-rag-retrieval-augmented-generation/",
                                  "author": "Matt Willsmore",
                                  "header": "5 Challenges Implementing Retrieval Augmented Generation (RAG) - Pureinsights: A blog on 5 common challenges when implementing RAG (Retrieval Augmented Generation) and possible solutions for search applications."
                          },
                          "transaction": "694eb7be78aedc7a163da900"
                  },
                  {
                          "id": "3",
                          "creationTimestamp": "2025-12-26T16:28:54Z",
                          "lastUpdatedTimestamp": "2025-12-26T16:28:54Z",
                          "action": "STORE",
                          "checksum": "cbffeeba8f4739650ae048fb382c8870",
                          "content": {
                                  "_id": "d758c733466967ea6f13b20bcbfcebb5",
                                  "link": "https://pureinsights.com/blog/2024/modernizing-search-with-generative-ai/",
                                  "author": "Martin Bayton",
                                  "header": "Modernizing Search with Generative AI - Pureinsights: Blog: why you should implement Retrieval-Augmented Generation (RAG) and how platforms like Pureinsights Discovery streamline the process."
                          },
                          "transaction": "694eb7c678aedc7a163da901"
                  }
          ],
		  "empty": false
			}`))
		}
	}))
	t.Cleanup(srv.Close)

	contentClient := newContentClient(srv.URL, "", "my-bucket")
	size := 3
	response, err := contentClient.Scroll(gjson.Parse(filters), gjson.Parse(projections), &size)
	require.NoError(t, err)
	assert.Len(t, response, 6)
}

// TestWithContentAction tests the WithContentAction functional option.
// It uses the Get function to call the option.
func TestWithContentAction(t *testing.T) {
	srv := httptest.NewServer(
		testutils.HttpHandler(t,
			http.StatusOK, "application/json", `{"ok":true}`,
			func(t *testing.T, r *http.Request) {
				assert.Equal(t, "STORE", r.URL.Query().Get("action"))
			}))
	t.Cleanup(srv.Close)

	contentClient := newContentClient(srv.URL, "", "testBucket")

	contentClient.Get("c28db957887e1aae75e7ab1dd0fd34e9", WithContentAction("STORE"))
}

// TestWithIncludeProjections tests the WithIncludeProjections functional option.
// It uses the Get function to call the option.
func TestWithIncludeProjections(t *testing.T) {
	srv := httptest.NewServer(
		testutils.HttpHandler(t,
			http.StatusOK, "application/json", `{"ok":true}`,
			func(t *testing.T, r *http.Request) {
				assert.Equal(t, []string{"author", "title", "description"}, r.URL.Query()["include"])
			}))
	t.Cleanup(srv.Close)

	contentClient := newContentClient(srv.URL, "", "testBucket")

	contentClient.Get("c28db957887e1aae75e7ab1dd0fd34e9", WithIncludeProjections([]string{"author", "title", "description"}))
}

// TestWithExcludeProjections tests the WithExcludeProjections functional option.
// It uses the Get function to call the option.
func TestWithExcludeProjections(t *testing.T) {
	srv := httptest.NewServer(
		testutils.HttpHandler(t,
			http.StatusOK, "application/json", `{"ok":true}`,
			func(t *testing.T, r *http.Request) {
				assert.Equal(t, []string{"author", "title", "description"}, r.URL.Query()["exclude"])
			}))
	t.Cleanup(srv.Close)

	contentClient := newContentClient(srv.URL, "", "testBucket")

	contentClient.Get("c28db957887e1aae75e7ab1dd0fd34e9", WithExcludeProjections([]string{"author", "title", "description"}))
}

// Test_staging_Buckets tests the staging.Buckets() function.
func Test_staging_Buckets(t *testing.T) {
	url := "http://localhost:12020"
	apiKey := "Api Key"
	staging := NewStaging(url, apiKey)
	c := staging.Buckets()

	assert.Equal(t, apiKey, c.client.ApiKey)
	assert.Equal(t, url+"/v2/bucket", c.client.client.BaseURL)
}

// Test_staging_Content tests the staging.Content() function.
func Test_staging_Content(t *testing.T) {
	url := "http://localhost:12020"
	apiKey := "Api Key"
	bucketName := "testBucket"
	staging := NewStaging(url, apiKey)
	c := staging.Content(bucketName)

	assert.Equal(t, apiKey, c.client.ApiKey)
	assert.Equal(t, url+"/v2/content/"+bucketName, c.client.client.BaseURL)
}

// Test_staging_StatusChecker tests the staging.StatusChecker() function.
func Test_staging_StatusChecker(t *testing.T) {
	s := NewStaging("http://localhost:12020", "Api Key")
	bc := s.StatusChecker()

	assert.Equal(t, s.ApiKey, bc.ApiKey)
	assert.Equal(t, "http://localhost:12020", bc.client.client.BaseURL)
}

// TestNewStaging tests the staging client constructor.
func TestNewStaging(t *testing.T) {
	url := "http://localhost:12020/v2"
	apiKey := "Api Key"
	s := NewStaging(url, apiKey)

	assert.Equal(t, apiKey, s.ApiKey, "ApiKey should be stored")
	assert.Equal(t, url+"/v2", s.Url, "BaseURL should match server URL")
}

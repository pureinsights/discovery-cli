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
	}{
		// Working case
		{
			name:       "Create works",
			method:     http.MethodPost,
			path:       "/testBucket",
			statusCode: http.StatusCreated,
			response:   `{"acknowledged": true}`,
			err:        false,
		},

		// Error case
		{
			name:       "Create returns 409 Conflict",
			method:     http.MethodPost,
			path:       "/testBucket",
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

			response, err := bucketsClient.Create("testBucket", jsonBody)
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

func Test_bucketsClient_Get(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		response   string
		err        bool
	}{
		// Working case
		{
			name:       "Get works",
			method:     http.MethodGet,
			path:       "/testBucket",
			statusCode: http.StatusOK,
			response:   `{"name":"testBucket","documentCount":{"STORE":3},"indices":[{"name":"authorIndex","fields":[{"author":"DESC"}],"unique":false}]}`,
			err:        false,
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
			err: true,
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
			if !(tc.err) {
				require.NoError(t, err)
				assert.Equal(t, "testBucket", response.Get("name").String())
				assert.Equal(t, int64(3), response.Get("documentCount.STORE").Int())
				assert.Equal(t, "authorIndex", response.Get("indices.0.name").String())
			} else {
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", tc.statusCode, tc.response))
			}
		})
	}
}

func Test_bucketsClient_GetAll(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		statusCode  int
		response    string
		expectedLen int
		err         bool
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
			expectedLen: 3,
			err:         false,
		}, {
			name:        "GetAll returns no content",
			method:      http.MethodGet,
			path:        "",
			statusCode:  http.StatusNoContent,
			response:    ``,
			expectedLen: 0,
			err:         false,
		},
		// Error case
		{
			name:        "GetAll returns an internal server error",
			method:      http.MethodGet,
			path:        "",
			statusCode:  http.StatusInternalServerError,
			response:    ``,
			expectedLen: 0,
			err:         true,
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
			if !(tc.err) {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedLen, len(response))
			} else {
				assert.Equal(t, []string(nil), response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", tc.statusCode, tc.response))
			}
		})
	}
}

func Test_bucketsClient_GetAll_InvalidJSONResponse(t *testing.T) {
	srv := httptest.NewServer(
		testutils.HttpHandler(t,
			http.StatusOK, "application/json", `{"message"} : "This cannot be marshalled."`,
			nil))
	t.Cleanup(srv.Close)

	bucketsClient := newBucketsClient(srv.URL, "")

	response, err := bucketsClient.GetAll()
	assert.Equal(t, []string(nil), response)
	assert.EqualError(t, err, "invalid character '}' after object key")
}

func Test_bucketsClient_Delete(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		response   string
		err        bool
	}{
		// Working case
		{
			name:       "Delete works",
			method:     http.MethodDelete,
			path:       "/testBucket",
			statusCode: http.StatusOK,
			response:   `{"acknowledged":true}`,
			err:        false,
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
			err: true,
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

func Test_bucketsClient_Purge(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		response   string
		err        bool
	}{
		// Working case
		{
			name:       "Purge works",
			method:     http.MethodDelete,
			path:       "/testBucket/purge",
			statusCode: http.StatusOK,
			response:   `{"acknowledged":true}`,
			err:        false,
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
			err: true,
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

func Test_bucketsClient_CreateIndex(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		statusCode  int
		response    string
		err         bool
		indexConfig []string
	}{
		// Working case
		{
			name:        "Create Index works",
			method:      http.MethodPut,
			path:        "/testBucket/index/testIndex",
			statusCode:  http.StatusOK,
			response:    `{"acknowledged":true}`,
			indexConfig: []string{`{"fieldName":"ASC"}`, `{"author":"DESC"}`},
			err:         false,
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
			indexConfig: []string{`{"fieldName":"ASC"}`, `{"author":"DESC"}`},
			err:         true,
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
			indexConfig: []string{`{"fieldName":"ASC"}`, `{"author":"DESC"}`},
			err:         true,
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
			indexConfig: []string{`{"fieldName":"ASC"}`, `"author":"DESC"}`},
			err:         true,
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

			indices := []gjson.Result{gjson.Parse(tc.indexConfig[0]), gjson.Parse(tc.indexConfig[1])}
			response, err := bucketsClient.CreateIndex("testBucket", "testIndex", indices)
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

func Test_bucketsClient_DeleteIndex(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		response   string
		err        bool
	}{
		// Working case
		{
			name:       "Delete index works",
			method:     http.MethodDelete,
			path:       "/testBucket/index/testIndex",
			statusCode: http.StatusOK,
			response:   `{"acknowledged":true}`,
			err:        false,
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
			err: true,
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
			err: true,
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

func Test_newContentClient(t *testing.T) {
	url := "http://localhost:8081/v2"
	apiKey := "Api Key"
	bucketName := "testBucket"
	c := newContentClient(url, apiKey, bucketName)

	assert.Equal(t, apiKey, c.client.ApiKey)
	assert.Equal(t, url+"/content/"+bucketName, c.client.client.BaseURL)
}

func Test_contentClient_Store(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		response   string
		bucketName string
		contentId  string
		parentId   string
		err        bool
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
			bucketName: "testBucket",
			contentId:  "c28db957887e1aae75e7ab1dd0fd34e9",
			parentId:   "d758c733466967ea6f13b20bcbfcebb5",
			err:        false,
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
			bucketName: "testBucket",
			contentId:  "c28db957887e1aae75e7ab1dd0fd34e9",
			parentId:   "",
			err:        false,
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
			bucketName: "testBucket",
			contentId:  "c28db957887e1aae75e7ab1dd0fd34e9",
			parentId:   "d758c733466967ea6f13b20bcbfcebb5",
			err:        true,
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
			bucketName: "   ",
			contentId:  "   ",
			parentId:   "   ",
			err:        true,
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
			bucketName: "testBucket",
			contentId:  "c28db957887e1aae75e7ab1dd0fd34e9",
			parentId:   "d758c733466967ea6f13b20bcbfcebb5",
			err:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, "/content/"+tc.bucketName+tc.path, r.URL.Path)
			}))

			defer srv.Close()

			json := gjson.Parse(`{
				"reference": "https://pureinsights.com/blog/2024/five-common-challenges-when-implementing-rag-retrieval-augmented-generation/",
				"title": "5 Challenges Implementing Retrieval Augmented Generation (RAG) - Pureinsights",
				"description": "A blog on 5 common challenges when implementing RAG (Retrieval Augmented Generation) and possible solutions for search applications.",
				"author": "Matt Willsmore"
			}`)
			contentClient := newContentClient(srv.URL, "", tc.bucketName)

			response, err := contentClient.Store(tc.contentId, tc.parentId, json)
			if !(tc.err) {
				require.NoError(t, err)
				assert.Equal(t, "Matt Willsmore", response.Get("content.author").String())
			} else {
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", tc.statusCode, tc.response))
			}
		})
	}
}

func Test_contentClient_Get(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		response   string
		bucketName string
		contentId  string
		err        bool
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
			bucketName: "testBucket",
			contentId:  "c28db957887e1aae75e7ab1dd0fd34e9",
			err:        false,
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
			bucketName: "   ",
			contentId:  "c28db957887e1aae75e7ab1dd0fd34e9",
			err:        true,
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
			bucketName: "testBucket",
			contentId:  "c28db957887e1aae75e7ab1dd0fd34e9",
			err:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, "/content/"+tc.bucketName+tc.path, r.URL.Path)
				assert.Equal(t, "STORE", r.URL.Query().Get("action"))
			}))

			defer srv.Close()

			contentClient := newContentClient(srv.URL, "", tc.bucketName)

			response, err := contentClient.Get(tc.contentId, WithContentAction("STORE"))
			if !(tc.err) {
				require.NoError(t, err)
				assert.Equal(t, "Matt Willsmore", response.Get("content.author").String())
			} else {
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", tc.statusCode, tc.response))
			}
		})
	}
}

func Test_contentClient_Delete(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		response   string
		bucketName string
		contentId  string
		err        bool
	}{
		// Working case
		{
			name:       "Delete works",
			method:     http.MethodDelete,
			path:       "/c28db957887e1aae75e7ab1dd0fd34e9",
			statusCode: http.StatusOK,
			response:   `{"acknowledged":true}`,
			bucketName: "testBucket",
			contentId:  "c28db957887e1aae75e7ab1dd0fd34e9",
			err:        false,
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
			bucketName: "testBucket",
			contentId:  "  ",
			err:        true,
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
			bucketName: "testBucket",
			contentId:  "c28db957887e1aae75e7ab1dd0fd34e9",
			err:        true,
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

func Test_contentClient_DeleteMany(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		statusCode int
		response   string
		bucketName string
		parentId   string
		filters    string
		err        bool
	}{
		// Working case
		{
			name:       "Delete Many works with parentId and no filters",
			method:     http.MethodDelete,
			statusCode: http.StatusOK,
			response:   `{"acknowledged":true}`,
			bucketName: "testBucket",
			parentId:   "d758c733466967ea6f13b20bcbfcebb5",
			filters:    ``,
			err:        false,
		},
		{
			name:       "Delete Many works with no parentId and filters",
			method:     http.MethodDelete,
			statusCode: http.StatusOK,
			response:   `{"acknowledged":true}`,
			bucketName: "testBucket",
			parentId:   "",
			filters: `{
			"equals": {
				"field": "author",
				"value": "Martin Bayton",
				"normalize": true
			}
			}`,
			err: false,
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
			bucketName: "  ",
			parentId:   "d758c733466967ea6f13b20bcbfcebb5",
			err:        true,
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
			bucketName: "testBucket",
			parentId:   "c28db957887e1aae75e7ab1dd0fd34e9",
			err:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, "/content/"+tc.bucketName, r.URL.Path)
				if tc.parentId != "" {
					assert.Equal(t, tc.parentId, r.URL.Query().Get("parentId"))
				}
			}))

			defer srv.Close()

			contentClient := newContentClient(srv.URL, "", tc.bucketName)

			response, err := contentClient.DeleteMany(tc.parentId, gjson.Parse(tc.filters))
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
	url := "http://localhost:8081/v2"
	apiKey := "Api Key"
	staging := NewStaging(url, apiKey)
	c := staging.Buckets()

	assert.Equal(t, apiKey, c.client.ApiKey)
	assert.Equal(t, url+"/bucket", c.client.client.BaseURL)
}

// Test_staging_Content tests the staging.Content() function.
func Test_staging_Content(t *testing.T) {
	url := "http://localhost:8081/v2"
	apiKey := "Api Key"
	bucketName := "testBucket"
	staging := NewStaging(url, apiKey)
	c := staging.Content(bucketName)

	assert.Equal(t, apiKey, c.client.ApiKey)
	assert.Equal(t, url+"/content/"+bucketName, c.client.client.BaseURL)
}

// TestNewStaging tests the staging client constructor.
func TestNewStaging(t *testing.T) {
	url := "http://localhost:8081/v2"
	apiKey := "Api Key"
	s := NewStaging(url, apiKey)

	assert.Equal(t, apiKey, s.ApiKey, "ApiKey should be stored")
	assert.Equal(t, url, s.Url, "BaseURL should match server URL")
}

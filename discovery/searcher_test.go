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

// Test_searcher_Search_HTTPResponseCases tests how the executeWithPagination() function behaves with various HTTP responses.
func Test_searcher_Search_HTTPResponseCases(t *testing.T) {
	body := gjson.Parse(`{
	"equals": {
		"field": "type",
		"value": "mongo",
		"normalize": true
	}
	}`)
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
			name:       "executeWithPagination returns array",
			method:     http.MethodPost,
			path:       "/search",
			statusCode: http.StatusOK,
			response: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone",
					"labels": [],
					"active": true,
					"id": "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
					"creationTimestamp": "2025-09-29T15:50:17Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:17Z"
				},
				"highlight": {},
				"score": 0.14617437
				},
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone 1",
					"labels": [],
					"active": true,
					"id": "8f14c11c-bb66-49d3-aa2a-dedff4608c17",
					"creationTimestamp": "2025-09-29T15:50:19Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:19Z"
				},
				"highlight": {},
				"score": 0.14617437
				},
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone 3",
					"labels": [],
					"active": true,
					"id": "3a0214a4-72cc-4eee-ad0c-9e3af9b08a6c",
					"creationTimestamp": "2025-09-29T15:50:20Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:20Z"
				},
				"highlight": {},
				"score": 0.14617437
				},
			],
			"pageable": {
				"page": 0,
				"size": 25,
				"sort": []
			},
			"totalSize": 3,
			"totalPages": 1,
			"empty": false,
			"size": 25,
			"offset": 0,
			"pageNumber": 0,
			"numberOfElements": 3
			}`,
			expectedLen: 3,
			err:         nil,
		},
		{
			name:        "executeWithPagination returns no content",
			method:      http.MethodPost,
			path:        "/search",
			statusCode:  http.StatusNoContent,
			response:    `{"content": []}`,
			expectedLen: 0,
			err:         nil,
		},
		{
			name:        "executeWithPagination has no content field",
			method:      http.MethodPost,
			path:        "/search",
			statusCode:  http.StatusNoContent,
			response:    ``,
			expectedLen: 0,
			err:         nil,
		},

		// Error cases
		{
			name:       "executeWithPagination returns a 401 Unauthorized",
			method:     http.MethodPost,
			path:       "/search",
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
				requestBody, _ := io.ReadAll(r.Body)
				assert.Equal(t, body.Raw, string(requestBody))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			}))

			defer srv.Close()

			s := searcher{client: newClient(srv.URL, "")}
			results, err := s.Search(body)
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

// Test_searcher_Search_HTTPResponseCases tests how the executeWithPagination() function behaves with various HTTP responses.
func Test_searcher_SearchByName(t *testing.T) {
	filterString := `{
		"equals": {
			"field": "name",
			"value": "%s"
		}
	}`
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		nameFilter string
		response   string
		result     gjson.Result
		err        error
	}{
		// Working cases
		{
			name:       "executeWithPagination returns array",
			method:     http.MethodPost,
			path:       "/search",
			statusCode: http.StatusOK,
			nameFilter: "MongoDB Atlas Server",
			result: gjson.Parse(`{
					"type": "mongo",
					"name": "MongoDB Atlas Server",
					"labels": [],
					"active": true,
					"id": "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
					"creationTimestamp": "2025-09-29T15:50:17Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:17Z"
				}`),
			response: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas Server",
					"labels": [],
					"active": true,
					"id": "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
					"creationTimestamp": "2025-09-29T15:50:17Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:17Z"
				},
				"highlight": {
					"name": [
						"<em>MongoDB</em> <em>Atlas</em> <em>Server</em>"
					]
				},
				"score": 0.321809,
				},
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone 1",
					"labels": [],
					"active": true,
					"id": "8f14c11c-bb66-49d3-aa2a-dedff4608c17",
					"creationTimestamp": "2025-09-29T15:50:19Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:19Z"
				},
				"highlight": {
					"name": [
						"<em>MongoDB</em> <em>Atlas</em> <em>server</em> clone 1"
					]
				},
				"score": 0.29428056,
				},
				{
				"source": {
					"type": "openai",
					"name": "OpenAI server",
					"labels": [],
					"active": true,
					"id": "3a0214a4-72cc-4eee-ad0c-9e3af9b08a6c",
					"creationTimestamp": "2025-09-29T15:50:20Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:20Z"
				},
				"highlight": {
					"name": [
                        "OpenAI <em>Server</em>"
                ]},
				"score": 0.013445788
				},
			],
			"pageable": {
				"page": 0,
				"size": 25,
				"sort": []
			},
			"totalSize": 3,
			"totalPages": 1,
			"empty": false,
			"size": 25,
			"offset": 0,
			"pageNumber": 0,
			"numberOfElements": 3
			}`,
			err: nil,
		},

		// Error cases
		{
			name:       "executeWithPagination returns no content",
			method:     http.MethodPost,
			path:       "/search",
			statusCode: http.StatusNoContent,
			response:   `{"content": []}`,
			nameFilter: "MongoDB Atlas Server",
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name "MongoDB Atlas Server" does not exist"
	],
	"timestamp": "2025-09-30T15:38:42.885125200Z"
}`)},
		},
		{
			name:       "executeWithPagination has no content field",
			method:     http.MethodPost,
			path:       "/search",
			statusCode: http.StatusNoContent,
			response:   ``,
			nameFilter: "MongoDB Atlas Server",
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name "MongoDB Atlas Server" does not exist"
	],
	"timestamp": "2025-09-30T15:38:42.885125200Z"
}`)},
		},
		{
			name:       "The first result does not have the same name",
			method:     http.MethodPost,
			path:       "/search",
			statusCode: http.StatusOK,
			nameFilter: "MongoDB Atlas Server",
			response: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone",
					"labels": [],
					"active": true,
					"id": "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
					"creationTimestamp": "2025-09-29T15:50:17Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:17Z"
				},
				"highlight": {
					"name": [
						"<em>MongoDB</em> <em>Atlas</em> <em>server</em> clone"
					]
				},
				"score": 0.321809,
				},
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone 1",
					"labels": [],
					"active": true,
					"id": "8f14c11c-bb66-49d3-aa2a-dedff4608c17",
					"creationTimestamp": "2025-09-29T15:50:19Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:19Z"
				},
				"highlight": {
					"name": [
						"<em>MongoDB</em> <em>Atlas</em> <em>server</em> clone 1"
					]
				},
				"score": 0.29428056,
				},
				{
				"source": {
					"type": "openai",
					"name": "OpenAI server",
					"labels": [],
					"active": true,
					"id": "3a0214a4-72cc-4eee-ad0c-9e3af9b08a6c",
					"creationTimestamp": "2025-09-29T15:50:20Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:20Z"
				},
				"highlight": {
					"name": [
                        "OpenAI <em>Server</em>"
                ]},
				"score": 0.013445788
				},
			],
			"pageable": {
				"page": 0,
				"size": 25,
				"sort": []
			},
			"totalSize": 3,
			"totalPages": 1,
			"empty": false,
			"size": 25,
			"offset": 0,
			"pageNumber": 0,
			"numberOfElements": 3
			}`,
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name "MongoDB Atlas Server" does not exist"
	],
	"timestamp": "2025-09-30T15:38:42.885125200Z"
}`)},
		},
		{
			name:       "executeWithPagination returns a 401 Unauthorized",
			method:     http.MethodPost,
			path:       "/search",
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
				requestBody, _ := io.ReadAll(r.Body)
				assert.Equal(t, fmt.Sprintf(filterString, tc.nameFilter), string(requestBody))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			}))

			defer srv.Close()

			s := searcher{client: newClient(srv.URL, "")}
			result, err := s.SearchByName(tc.nameFilter)
			if tc.err == nil {
				require.NoError(t, err)
				assert.Equal(t, tc.result.Raw, result.Raw)
			} else {
				assert.Equal(t, gjson.Result{}, result)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

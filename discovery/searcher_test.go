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

// Test_searcher_Search_HTTPResponseCases tests the searcher.Search() function.
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

// Test_searcher_SearchByName tests searcher.SearchByName().
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
		responses  map[string]testutils.MockResponse
		result     gjson.Result
		err        error
	}{
		// Working cases
		{
			name:       "Search By Name returns a result with matching name",
			method:     http.MethodPost,
			path:       "/search",
			statusCode: http.StatusOK,
			nameFilter: "my-credential",
			result: gjson.Parse(`{
					"type": "mongo",
					"name": "my-credential",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "4957145b-6192-4862-a5da-e97853974e9f",
					"creationTimestamp": "2025-10-17T22:37:53Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
				}`),
			responses: map[string]testutils.MockResponse{
				"POST:/search": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "my-credential",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "3b32e410-2f33-412d-9fb8-17970131921c",
					"creationTimestamp": "2025-10-17T22:37:57Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:57Z"
				},
				"highlight": {}
				"score": 1.4854797
				},
				{
				"source": {
					"type": "mongo",
					"name": "my-credential",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "4957145b-6192-4862-a5da-e97853974e9f",
					"creationTimestamp": "2025-10-17T22:37:53Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
				},
				"highlight": {
					"name": [
					"<em>label</em> <em>test</em> 1 <em>clone</em>"
					]
				},
				"score": 0.3980717
				}
			],
			"pageable": {
				"page": 0,
				"size": 25,
				"sort": []
			},
			"totalSize": 18,
			"totalPages": 1,
			"empty": false,
			"size": 25,
			"offset": 0,
			"numberOfElements": 18,
			"pageNumber": 0
			}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/search", r.URL.Path)
						assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
						requestBody, _ := io.ReadAll(r.Body)
						assert.Equal(t, fmt.Sprintf(filterString, "my-credential"), string(requestBody))
					},
				},
				"GET:/3b32e410-2f33-412d-9fb8-17970131921c": {
					StatusCode: http.StatusOK,
					Body: `{
					"type": "mongo",
					"name": "my-credential",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "4957145b-6192-4862-a5da-e97853974e9f",
					"creationTimestamp": "2025-10-17T22:37:53Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
				}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/3b32e410-2f33-412d-9fb8-17970131921c", r.URL.Path)
					},
				},
			},
			err: nil,
		},

		// Error cases
		{
			name:       "SearchByName returns no content",
			method:     http.MethodPost,
			path:       "/search",
			statusCode: http.StatusNoContent,
			responses: map[string]testutils.MockResponse{
				"POST:/search": {
					StatusCode:  http.StatusNoContent,
					Body:        `{"content": []}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/search", r.URL.Path)
						assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
						requestBody, _ := io.ReadAll(r.Body)
						assert.Equal(t, fmt.Sprintf(filterString, "MongoDB Atlas Server"), string(requestBody))
					},
				},
			},
			nameFilter: "MongoDB Atlas Server",
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name "MongoDB Atlas Server" does not exist"
	]
}`)},
		},
		{
			name:       "SearchByName results have no content field",
			method:     http.MethodPost,
			path:       "/search",
			statusCode: http.StatusNoContent,
			responses: map[string]testutils.MockResponse{
				"POST:/search": {
					StatusCode:  http.StatusNoContent,
					Body:        ``,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/search", r.URL.Path)
						assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
						requestBody, _ := io.ReadAll(r.Body)
						assert.Equal(t, fmt.Sprintf(filterString, "MongoDB Atlas Server"), string(requestBody))
					},
				},
			},
			nameFilter: "MongoDB Atlas Server",
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name "MongoDB Atlas Server" does not exist"
	]
}`)},
		},
		{
			name:       "The first result does not have the same name",
			method:     http.MethodPost,
			path:       "/search",
			statusCode: http.StatusOK,
			nameFilter: "MongoDB Atlas Server",
			responses: map[string]testutils.MockResponse{
				"POST:/search": {
					StatusCode: http.StatusOK,
					Body: `{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "my-credential",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "3b32e410-2f33-412d-9fb8-17970131921c",
					"creationTimestamp": "2025-10-17T22:37:57Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:57Z"
				},
				"highlight": {}
				"score": 1.4854797
				},
				{
				"source": {
					"type": "mongo",
					"name": "my-credential",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "4957145b-6192-4862-a5da-e97853974e9f",
					"creationTimestamp": "2025-10-17T22:37:53Z",
					"lastUpdatedTimestamp": "2025-10-17T22:37:53Z"
				},
				"highlight": {
					"name": [
					"<em>label</em> <em>test</em> 1 <em>clone</em>"
					]
				},
				"score": 0.3980717
				}
			],
			"pageable": {
				"page": 0,
				"size": 25,
				"sort": []
			},
			"totalSize": 18,
			"totalPages": 1,
			"empty": false,
			"size": 25,
			"offset": 0,
			"numberOfElements": 18,
			"pageNumber": 0
			}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/search", r.URL.Path)
						assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
						requestBody, _ := io.ReadAll(r.Body)
						assert.Equal(t, fmt.Sprintf(filterString, "MongoDB Atlas Server"), string(requestBody))
					},
				},
			},
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name "MongoDB Atlas Server" does not exist"
	]
}`)},
		},
		{
			name:       "SearchByName returns a 401 Unauthorized",
			method:     http.MethodPost,
			path:       "/search",
			statusCode: http.StatusUnauthorized,
			nameFilter: "MongoDB Atlas Server",
			responses: map[string]testutils.MockResponse{
				"POST:/search": {
					StatusCode:  http.StatusUnauthorized,
					Body:        `{"error":"unauthorized"}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/search", r.URL.Path)
						assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
						requestBody, _ := io.ReadAll(r.Body)
						assert.Equal(t, fmt.Sprintf(filterString, "MongoDB Atlas Server"), string(requestBody))
					},
				},
			},
			err: Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpMultiResponseHandler(t, tc.responses))

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

package discovery

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
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

// Test_searcher_Search_ErrorInSecondPage tests when executeWithPagination fails in a request while trying to get every content from every page.
func Test_searcher_Search_ErrorInSecondPage(t *testing.T) {
	body := gjson.Parse(`{
	"equals": {
		"field": "type",
		"value": "mongo",
		"normalize": true
	}
	}`)
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			requestBody, _ := io.ReadAll(r.Body)
			assert.Equal(t, body.Raw, string(requestBody))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "/search", r.URL.Path)
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
					}
				],
				"pageable": {
					"page": 0,
					"size": 3,
					"sort": []
				},
				"totalSize": 12,
				"totalPages": 4,
				"empty": false,
				"size": 3,
				"offset": 0,
				"pageNumber": 0,
				"numberOfElements": 3
				}`))
			}
		}))
	t.Cleanup(srv.Close)

	s := searcher{client: newClient(srv.URL, "")}
	response, err := s.Search(body)
	assert.Equal(t, []gjson.Result(nil), response)
	var errStruct Error
	require.ErrorAs(t, err, &errStruct)
	assert.EqualError(t, err, Error{Status: http.StatusInternalServerError, Body: gjson.Parse(`{"error":"Internal Server Error"}`)}.Error())
}

// Test_searcher_Search_RestyReturnsError tests what happens when the Resty client fails to execute the request.
func Test_searcher_Search_RestyReturnsError(t *testing.T) {
	body := gjson.Parse(`{
	"equals": {
		"field": "type",
		"value": "mongo",
		"normalize": true
	}
	}`)

	srv := httptest.NewServer(http.NotFoundHandler())
	base := srv.URL
	srv.Close()

	s := searcher{client: newClient(srv.URL, "")}
	response, err := s.Search(body)
	require.Error(t, err)
	assert.Equal(t, response, []gjson.Result(nil))
	assert.Contains(t, err.Error(), base+"/search")
}

// Test_searcher_Search_ContentInSecondPage tests that the executeWithPagination() function
// can successfully get all content when there are two pages with content in them
func Test_searcher_Search_ContentInSecondPage(t *testing.T) {
	body := gjson.Parse(`{
	"equals": {
		"field": "type",
		"value": "mongo",
		"normalize": true
	}
	}`)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		requestBody, _ := io.ReadAll(r.Body)
		assert.Equal(t, body.Raw, string(requestBody))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "/search", r.URL.Path)
		pageNumber, _ := strconv.Atoi(r.URL.Query().Get("page"))
		w.Header().Set("Content-Type", "application/json")
		if pageNumber > 0 {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			"content": [
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone 8",
					"labels": [],
					"active": true,
					"id": "3edc9c72-a875-49d7-8929-af09f3e9c01c",
					"creationTimestamp": "2025-09-29T15:50:24Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:24Z"
				},
				"highlight": {},
				"score": 0.14617437
				},
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone 9",
					"labels": [],
					"active": true,
					"id": "2b839453-ddad-4ced-8e13-2c7860af60a7",
					"creationTimestamp": "2025-09-29T15:50:26Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:26Z"
				},
				"highlight": {},
				"score": 0.14617437
				},
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone 10",
					"labels": [],
					"active": true,
					"id": "612e160b-5d9a-4ce8-a3d7-7e8bb4f91756",
					"creationTimestamp": "2025-09-29T15:50:36Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:36Z"
				},
				"highlight": {},
				"score": 0.14617437
				},
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone 11",
					"labels": [],
					"active": true,
					"id": "025347a7-e2bd-4ba1-880f-db3e51319abb",
					"creationTimestamp": "2025-09-29T15:50:37Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:37Z"
				},
				"highlight": {},
				"score": 0.14617437
				},
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone 2",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "21029da3-041c-43b5-a67e-870251f2f6a6",
					"creationTimestamp": "2025-09-29T15:50:19Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:19Z"
				},
				"highlight": {},
				"score": 0.14617437
				},
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone 4",
					"labels": [
					{
						"key": "A",
						"value": "A"
					}
					],
					"active": true,
					"id": "a798cd5b-aa7a-4fc5-9292-1de6fe8e8b7f",
					"creationTimestamp": "2025-09-29T15:50:21Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:21Z"
				},
				"highlight": {},
				"score": 0.14617437
				}
			],
			"pageable": {
				"page": 1,
				"size": 6,
				"sort": []
			},
			"totalSize": 12,
			"totalPages": 2,
			"empty": false,
			"size": 6,
			"offset": 6,
			"pageNumber": 1,
			"numberOfElements": 6
			}`))
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
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
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone 5",
					"labels": [],
					"active": true,
					"id": "6ffc7784-481e-4da8-8ee3-6817d15a757c",
					"creationTimestamp": "2025-09-29T15:50:22Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:22Z"
				},
				"highlight": {},
				"score": 0.14617437
				},
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone 6",
					"labels": [],
					"active": true,
					"id": "226e8a0b-5016-4ebe-9963-1461edd39d0a",
					"creationTimestamp": "2025-09-29T15:50:22Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:22Z"
				},
				"highlight": {},
				"score": 0.14617437
				},
				{
				"source": {
					"type": "mongo",
					"name": "MongoDB Atlas server clone 7",
					"labels": [],
					"active": true,
					"id": "cb187c0e-9423-4137-a602-82bb941ebd71",
					"creationTimestamp": "2025-09-29T15:50:23Z",
					"lastUpdatedTimestamp": "2025-09-29T15:50:23Z"
				},
				"highlight": {},
				"score": 0.14617437
				}
			],
			"pageable": {
				"page": 0,
				"size": 6,
				"sort": []
			},
			"totalSize": 12,
			"totalPages": 2,
			"empty": false,
			"size": 6,
			"offset": 0,
			"pageNumber": 0,
			"numberOfElements": 6
			}`))
		}
	}))
	t.Cleanup(srv.Close)

	s := searcher{client: newClient(srv.URL, "")}
	response, err := s.Search(body)
	require.NoError(t, err)
	assert.Len(t, response, 12)
}

// Test_searcher_Search_NoContentInSecondPage tests what happens if one of the later pages returns No Content
func Test_searcher_Search_NoContentInSecondPage(t *testing.T) {
	body := gjson.Parse(`{
	"equals": {
		"field": "type",
		"value": "mongo",
		"normalize": true
	}
	}`)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		requestBody, _ := io.ReadAll(r.Body)
		assert.Equal(t, body.Raw, string(requestBody))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "/search", r.URL.Path)
		pageNumber, _ := strconv.Atoi(r.URL.Query().Get("page"))
		w.Header().Set("Content-Type", "application/json")
		if pageNumber > 0 {
			w.WriteHeader(http.StatusNoContent)
			_, _ = w.Write([]byte(``))
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
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
				}
			],
			"pageable": {
				"page": 0,
				"size": 3,
				"sort": []
			},
			"totalSize": 6,
			"totalPages": 2,
			"empty": false,
			"size": 3,
			"offset": 0,
			"pageNumber": 0,
			"numberOfElements": 3
			}`))
		}
	}))
	t.Cleanup(srv.Close)

	s := searcher{client: newClient(srv.URL, "")}
	response, err := s.Search(body)
	require.NoError(t, err)
	assert.Len(t, response, 3)
}

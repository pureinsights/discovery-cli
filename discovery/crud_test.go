package discovery

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestCRUD(t *testing.T) {
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
			name:       "Get by ID returns object",
			method:     http.MethodGet,
			path:       "/5f125024-1e5e-4591-9fee-365dc20eeeed",
			statusCode: http.StatusOK,
			response:   `{"id":"5f125024-1e5e-4591-9fee-365dc20eeeed","name":"test-secret"}`,
			testFunc: func(t *testing.T, c crud) {
				id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
				response, err := c.Get(id)
				require.NoError(t, err)
				assert.Equal(t, "test-secret", response.Get("name").String())
			},
		},
		{
			name:       "GetAll returns array",
			method:     http.MethodGet,
			path:       "/",
			statusCode: http.StatusOK,
			response: `{
			"content": [
				{
				"type": "mongo",
				"name": "MongoDB text processor 4",
				"labels": [],
				"active": true,
				"id": "3393f6d9-94c1-4b70-ba02-5f582727d998",
				"creationTimestamp": "2025-08-21T17:57:16Z",
				"lastUpdatedTimestamp": "2025-08-21T17:57:16Z"
				},
				{
				"type": "mongo",
				"name": "MongoDB text processor",
				"labels": [],
				"active": true,
				"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
				},
				{
				"type": "script",
				"name": "Script processor",
				"labels": [],
				"active": true,
				"id": "86e7f920-a4e4-4b64-be84-5437a7673db8",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-14T18:02:38Z"
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
			"numberOfElements": 3,
			"pageNumber": 0
			}`,
			testFunc: func(t *testing.T, c crud) {
				results, err := c.GetAll()
				require.NoError(t, err)
				assert.Len(t, results, 12)
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
				assert.Len(t, results, 0)
			},
		},
		{
			name:       "Create works",
			method:     http.MethodPost,
			path:       "/",
			statusCode: http.StatusCreated,
			response:   `{"id":"5f125024-1e5e-4591-9fee-365dc20eeeed","name":"new-secret"}`,
			testFunc: func(t *testing.T, c crud) {
				config := gjson.Parse(`{"name":"new-secret"}`)
				response, err := c.Create(config)
				require.NoError(t, err)
				assert.Equal(t, "5f125024-1e5e-4591-9fee-365dc20eeeed", response.Get("id").String())
			},
		},
		{
			name:       "Update works",
			method:     http.MethodPut,
			path:       "/5f125024-1e5e-4591-9fee-365dc20eeeed",
			statusCode: http.StatusOK,
			response:   `{"id":"5f125024-1e5e-4591-9fee-365dc20eeeed","name":"new-secret"}`,
			testFunc: func(t *testing.T, c crud) {
				id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
				config := gjson.Parse(`{"name":"updated-secret"}`)
				response, err := c.Update(id, config)
				require.NoError(t, err)
				assert.Equal(t, "5f125024-1e5e-4591-9fee-365dc20eeeed", response.Get("id").String())
			},
		},
		{
			name:       "Delete works",
			method:     http.MethodDelete,
			path:       "/5f125024-1e5e-4591-9fee-365dc20eeeed",
			statusCode: http.StatusOK,
			response:   `{"acknowledged": true}`,
			testFunc: func(t *testing.T, c crud) {
				id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
				response, err := c.Delete(id)
				require.NoError(t, err)
				assert.True(t, response.Get("acknowledged").Bool())
			},
		},

		// Error cases
		{
			name:       "Get by ID returns 404 Not Found",
			method:     http.MethodGet,
			path:       "/5f125024-1e5e-4591-9fee-365dc20eeeed",
			statusCode: http.StatusNotFound,
			response:   `{"messages": ["Secret not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`,
			testFunc: func(t *testing.T, c crud) {
				id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
				response, err := c.Get(id)
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusNotFound, []byte(`{"messages": ["Secret not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`)))
			},
		},
		{
			name:       "GetAll returns a 401 Unauthorized",
			method:     http.MethodGet,
			path:       "/",
			statusCode: http.StatusUnauthorized,
			response:   `{"error":"unauthorized"}`,
			testFunc: func(t *testing.T, c crud) {
				response, err := c.GetAll()
				assert.Equal(t, []gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusUnauthorized, []byte(`{"error":"unauthorized"}`)))
			},
		},
		{
			name:       "GetAll has no content field",
			method:     http.MethodGet,
			path:       "/",
			statusCode: http.StatusNoContent,
			response:   `[]`,
			testFunc: func(t *testing.T, c crud) {
				results, err := c.GetAll()
				require.NoError(t, err)
				assert.Len(t, results, 0)
			},
		},
		{
			name:       "Create returns 403 Forbidden",
			method:     http.MethodPost,
			path:       "/",
			statusCode: http.StatusForbidden,
			response:   `{"error":"forbidden"}`,
			testFunc: func(t *testing.T, c crud) {
				config := gjson.Parse(`{"name":"new-secret"}`)
				response, err := c.Create(config)
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusForbidden, []byte(`{"error":"forbidden"}`)))
			},
		},
		{
			name:       "Update returns 500 Internal Server Error",
			method:     http.MethodPut,
			path:       "/5f125024-1e5e-4591-9fee-365dc20eeeed",
			statusCode: http.StatusInternalServerError,
			response:   `{"error":"internal server error"}`,
			testFunc: func(t *testing.T, c crud) {
				id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
				config := gjson.Parse(`{"name":"updated-secret"}`)
				response, err := c.Update(id, config)
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusInternalServerError, []byte(`{"error":"internal server error"}`)))
			},
		},
		{
			name:       "Delete returns 404 Not Found",
			method:     http.MethodDelete,
			path:       "/5f125024-1e5e-4591-9fee-365dc20eeeed",
			statusCode: http.StatusNotFound,
			response:   `{"messages": ["Secret not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`,
			testFunc: func(t *testing.T, c crud) {
				id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
				response, err := c.Delete(id)
				assert.Equal(t, gjson.Result{}, response)
				assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusNotFound, []byte(`{"messages": ["Secret not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`)))
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

// Test_getter_GetAll_ErrorInSecondPage tests when GetAll fails in a request while trying to get every content from every page.
func Test_getter_GetAll_ErrorInSecondPage(t *testing.T) {
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
				"type": "mongo",
				"name": "MongoDB text processor 4",
				"labels": [],
				"active": true,
				"id": "3393f6d9-94c1-4b70-ba02-5f582727d998",
				"creationTimestamp": "2025-08-21T17:57:16Z",
				"lastUpdatedTimestamp": "2025-08-21T17:57:16Z"
				},
				{
				"type": "mongo",
				"name": "MongoDB text processor",
				"labels": [],
				"active": true,
				"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
				},
				{
				"type": "script",
				"name": "Script processor",
				"labels": [],
				"active": true,
				"id": "86e7f920-a4e4-4b64-be84-5437a7673db8",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-14T18:02:38Z"
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
			"numberOfElements": 3,
			"pageNumber": 0
			}`))
			}
		}))
	t.Cleanup(srv.Close)

	c := crud{getter{newClient(srv.URL, "")}}

	response, err := c.GetAll()
	assert.Equal(t, []gjson.Result{}, response)
	assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", http.StatusInternalServerError, []byte(`{"error":"Internal Server Error"}`)))
}

// Test_getter_GetAll_NoContentInSecondPage tests what happens if one of the later pages returns No Content
func Test_getter_GetAll_NoContentInSecondPage(t *testing.T) {
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
				"type": "mongo",
				"name": "MongoDB text processor 4",
				"labels": [],
				"active": true,
				"id": "3393f6d9-94c1-4b70-ba02-5f582727d998",
				"creationTimestamp": "2025-08-21T17:57:16Z",
				"lastUpdatedTimestamp": "2025-08-21T17:57:16Z"
				},
				{
				"type": "mongo",
				"name": "MongoDB text processor",
				"labels": [],
				"active": true,
				"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
				},
				{
				"type": "script",
				"name": "Script processor",
				"labels": [],
				"active": true,
				"id": "86e7f920-a4e4-4b64-be84-5437a7673db8",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-14T18:02:38Z"
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

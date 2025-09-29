package discovery

import (
	"io"
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

// Test_getter_Get tests the getter.Get() function
func Test_getter_Get(t *testing.T) {
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
			name:             "Get by ID returns object",
			method:           http.MethodGet,
			path:             "/5f125024-1e5e-4591-9fee-365dc20eeeed",
			statusCode:       http.StatusOK,
			response:         `{"id":"5f125024-1e5e-4591-9fee-365dc20eeeed","name":"test-secret"}`,
			expectedResponse: gjson.Parse(`{"id":"5f125024-1e5e-4591-9fee-365dc20eeeed","name":"test-secret"}`),
			err:              nil,
		},

		// Error case
		{
			name:             "Get by ID returns 404 Not Found",
			method:           http.MethodGet,
			path:             "/5f125024-1e5e-4591-9fee-365dc20eeeed",
			statusCode:       http.StatusNotFound,
			response:         `{"messages": ["Secret not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`,
			expectedResponse: gjson.Result{},
			err:              Error{Status: http.StatusNotFound, Body: gjson.Parse(`{"messages": ["Secret not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`)},
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
			id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
			response, err := c.Get(id)
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

// Test_getter_GetAll_HTTPResponseCases tests how the getter.GetAll() function behaves when receiving different HTTP responses and errors.
// It does not test if reading all the pages works.
func Test_getter_GetAll_HTTPResponseCases(t *testing.T) {
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
			expectedLen: 12,
			err:         nil,
		},
		{
			name:        "GetAll returns no content",
			method:      http.MethodGet,
			path:        "/",
			statusCode:  http.StatusNoContent,
			response:    `{"content": []}`,
			expectedLen: 0,
			err:         nil,
		},
		{
			name:        "GetAll has no content field",
			method:      http.MethodGet,
			path:        "/",
			statusCode:  http.StatusNoContent,
			response:    ``,
			expectedLen: 0,
			err:         nil,
		},

		// Error cases
		{
			name:       "GetAll returns a 401 Unauthorized",
			method:     http.MethodGet,
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

			c := crud{getter{newClient(srv.URL, "")}}
			results, err := c.GetAll()
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
	assert.Equal(t, []gjson.Result(nil), response)
	var errStruct Error
	require.ErrorAs(t, err, &errStruct)
	assert.EqualError(t, err, Error{Status: http.StatusInternalServerError, Body: gjson.Parse(`{"error":"Internal Server Error"}`)}.Error())
}

// Test_getter_GetAll_ContentInSecondPage tests when there are two pages with content in them
func Test_getter_GetAll_ContentInSecondPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		pageNumber, _ := strconv.Atoi(r.URL.Query().Get("page"))
		w.Header().Set("Content-Type", "application/json")
		if pageNumber > 0 {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			"content": [
				{
					"type": "openai",
					"name": "OpenAI Chat Processor",
					"labels": [],
					"active": true,
					"id": "8a399b1c-95fc-406c-a220-7d321aaa7b0e",
					"creationTimestamp": "2025-08-14T18:02:38Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:38Z"
				},
				{
					"type": "mongo",
					"name": "MongoDB vector processor",
					"labels": [],
					"active": true,
					"id": "a5ee116b-bd95-474e-9d50-db7be988b196",
					"creationTimestamp": "2025-08-14T18:02:38Z",
					"lastUpdatedTimestamp": "2025-08-14T18:02:38Z"
				},
				{
					"type": "openai",
					"name": "OpenAI embeddings processor",
					"labels": [],
					"active": true,
					"id": "b5c25cd3-e7c9-4fd2-b7e6-2bcf6e2caf89",
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
			"pageNumber": 1
			}`))
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
	assert.Len(t, response, 6)
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

// Test_crud_Create tests the crud.Create() function.
func Test_crud_Create(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		body             string
		response         string
		expectedResponse gjson.Result
		err              error
	}{
		// Working case
		{
			name:             "Create works",
			method:           http.MethodPost,
			path:             "/",
			statusCode:       http.StatusCreated,
			body:             `{"name":"new-secret"}`,
			response:         `{"id":"5f125024-1e5e-4591-9fee-365dc20eeeed","name":"new-secret"}`,
			expectedResponse: gjson.Parse(`{"id":"5f125024-1e5e-4591-9fee-365dc20eeeed","name":"new-secret"}`),
			err:              nil,
		},

		// Error case
		{
			name:             "Create returns 403 Forbidden",
			method:           http.MethodPost,
			path:             "/",
			statusCode:       http.StatusForbidden,
			body:             `{"name":"new-secret"}`,
			response:         `{"error":"forbidden"}`,
			expectedResponse: gjson.Result{},
			err:              Error{Status: http.StatusForbidden, Body: gjson.Parse(`{"error":"forbidden"}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, tc.path, r.URL.Path)
				body, _ := io.ReadAll(r.Body)
				assert.Equal(t, tc.body, string(body))
			}))

			defer srv.Close()

			c := crud{getter{newClient(srv.URL, "")}}
			config := gjson.Parse(tc.body)
			response, err := c.Create(config)
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

// Test_crud_Update tests the crud.Update() function.
func Test_crud_Update(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		body             string
		response         string
		expectedResponse gjson.Result
		err              error
	}{
		// Working case
		{
			name:             "Update works",
			method:           http.MethodPut,
			path:             "/5f125024-1e5e-4591-9fee-365dc20eeeed",
			statusCode:       http.StatusOK,
			body:             `{"name":"updated-secret"}`,
			response:         `{"id":"5f125024-1e5e-4591-9fee-365dc20eeeed","name":"new-secret"}`,
			expectedResponse: gjson.Parse(`{"id":"5f125024-1e5e-4591-9fee-365dc20eeeed","name":"new-secret"}`),
			err:              nil,
		},

		// Error case
		{
			name:             "Update returns 500 Internal Server Error",
			method:           http.MethodPut,
			path:             "/5f125024-1e5e-4591-9fee-365dc20eeeed",
			statusCode:       http.StatusInternalServerError,
			body:             `{"name":"updated-secret"}`,
			response:         `{"error":"internal server error"}`,
			expectedResponse: gjson.Result{},
			err:              Error{Status: http.StatusInternalServerError, Body: gjson.Parse(`{"error":"internal server error"}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)
				assert.Equal(t, tc.path, r.URL.Path)
				body, _ := io.ReadAll(r.Body)
				assert.Equal(t, tc.body, string(body))
			}))

			defer srv.Close()

			c := crud{getter{newClient(srv.URL, "")}}
			id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
			config := gjson.Parse(tc.body)
			response, err := c.Update(id, config)
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

// Test_crud_Delete tests the crud.Delete() function
func Test_crud_Delete(t *testing.T) {
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
			path:             "/5f125024-1e5e-4591-9fee-365dc20eeeed",
			statusCode:       http.StatusOK,
			response:         `{"acknowledged": true}`,
			expectedResponse: gjson.Parse(`{"acknowledged": true}`),
			err:              nil,
		},

		// Error case
		{
			name:             "Delete returns 404 Not Found",
			method:           http.MethodDelete,
			path:             "/5f125024-1e5e-4591-9fee-365dc20eeeed",
			statusCode:       http.StatusNotFound,
			response:         `{"messages": ["Secret not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`,
			expectedResponse: gjson.Result{},
			err:              Error{Status: http.StatusNotFound, Body: gjson.Parse(`{"messages": ["Secret not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"]}`)},
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
			id := uuid.MustParse("5f125024-1e5e-4591-9fee-365dc20eeeed")
			response, err := c.Delete(id)
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

package discovery

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Test_backupRestore_Import has table-driven tests to test the Import method.
func Test_backupRestore_Import(t *testing.T) {
	fileName := "testdata/test-export.zip"
	bytes, err := os.ReadFile(fileName)
	require.NoError(t, err)

	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		response         string
		expectedResponse gjson.Result
		onConflict       OnConflict
		err              error
	}{
		// Working cases
		{
			name:             "Restore restores the entities with on conflict update.",
			method:           http.MethodPost,
			path:             "/import",
			statusCode:       http.StatusMultiStatus,
			response:         `{"Processor":[{"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","status":200},{"id":"72a57085-5da9-4c96-9d6c-26b019b80a7c","status":201},{"id":"7b192ea1-ac43-439b-9396-5e022f81f2cb","status":200},{"id":"87d85a14-fb17-4899-8bdc-9fc2b1b28857","status":200},{"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8","status":200}],"Pipeline":[{"id":"8f13eae4-73a5-45c9-9239-5c1996c0378f","status":200},{"id":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","status":200}],"Seed":[{"id":"2acd0a61-852c-4f38-af2b-9c84e152873e","status":200},{"id":"30fc6d99-ceb7-45f4-98f9-27e08c5f2d4c","status":200}],"SeedSchedule":[{"id":"e71b122c-ddff-4306-84ba-e230527d445c","status":200}]}`,
			expectedResponse: gjson.Parse(`{"Processor":[{"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","status":200},{"id":"72a57085-5da9-4c96-9d6c-26b019b80a7c","status":201},{"id":"7b192ea1-ac43-439b-9396-5e022f81f2cb","status":200},{"id":"87d85a14-fb17-4899-8bdc-9fc2b1b28857","status":200},{"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8","status":200}],"Pipeline":[{"id":"8f13eae4-73a5-45c9-9239-5c1996c0378f","status":200},{"id":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","status":200}],"Seed":[{"id":"2acd0a61-852c-4f38-af2b-9c84e152873e","status":200},{"id":"30fc6d99-ceb7-45f4-98f9-27e08c5f2d4c","status":200}],"SeedSchedule":[{"id":"e71b122c-ddff-4306-84ba-e230527d445c","status":200}]}`),
			onConflict:       OnConflictUpdate,
			err:              nil,
		},
		{
			name:             "Restore restores the entities with on conflict ignore.",
			method:           http.MethodPost,
			path:             "/import",
			statusCode:       http.StatusMultiStatus,
			response:         `{"Processor":[{"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","status":204},{"id":"70f700c5-c70f-4036-b6ed-5e931140b715","status":201},{"id":"7b192ea1-ac43-439b-9396-5e022f81f2cb","status":204},{"id":"87d85a14-fb17-4899-8bdc-9fc2b1b28857","status":204},{"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8","status":204}],"Pipeline":[{"id":"8f13eae4-73a5-45c9-9239-5c1996c0378f","status":204},{"id":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","status":204}],"Seed":[{"id":"2acd0a61-852c-4f38-af2b-9c84e152873e","status":204},{"id":"30fc6d99-ceb7-45f4-98f9-27e08c5f2d4c","status":204}],"SeedSchedule":[{"id":"e71b122c-ddff-4306-84ba-e230527d445c","status":204}]}`,
			expectedResponse: gjson.Parse(`{"Processor":[{"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","status":204},{"id":"70f700c5-c70f-4036-b6ed-5e931140b715","status":201},{"id":"7b192ea1-ac43-439b-9396-5e022f81f2cb","status":204},{"id":"87d85a14-fb17-4899-8bdc-9fc2b1b28857","status":204},{"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8","status":204}],"Pipeline":[{"id":"8f13eae4-73a5-45c9-9239-5c1996c0378f","status":204},{"id":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","status":204}],"Seed":[{"id":"2acd0a61-852c-4f38-af2b-9c84e152873e","status":204},{"id":"30fc6d99-ceb7-45f4-98f9-27e08c5f2d4c","status":204}],"SeedSchedule":[{"id":"e71b122c-ddff-4306-84ba-e230527d445c","status":204}]}`),
			onConflict:       OnConflictIgnore,
			err:              nil,
		},
		{
			name:             "Restore restores the entities with on conflict fail.",
			method:           http.MethodPost,
			path:             "/import",
			statusCode:       http.StatusMultiStatus,
			response:         `{"Processor":[{"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","status":409,"errorCode":2001,"errors":["Duplicated entity: 516d4a8a-e8ae-488c-9e37-d5746a907454"]},{"id":"5536e705-0c5d-4711-9a99-e272b32948cb","status":201},{"id":"7b192ea1-ac43-439b-9396-5e022f81f2cb","status":409,"errorCode":2001,"errors":["Duplicated entity: 7b192ea1-ac43-439b-9396-5e022f81f2cb"]},{"id":"87d85a14-fb17-4899-8bdc-9fc2b1b28857","status":409,"errorCode":2001,"errors":["Duplicated entity: 87d85a14-fb17-4899-8bdc-9fc2b1b28857"]},{"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8","status":409,"errorCode":2001,"errors":["Duplicated entity: aa0186f1-746f-4b20-b1b0-313bd79e78b8"]}],"Pipeline":[{"id":"8f13eae4-73a5-45c9-9239-5c1996c0378f","status":409,"errorCode":2001,"errors":["Duplicated entity: 8f13eae4-73a5-45c9-9239-5c1996c0378f"]},{"id":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","status":409,"errorCode":2001,"errors":["Duplicated entity: 9a74bf3a-eb2a-4334-b803-c92bf1bc45fe"]}],"Seed":[{"id":"2acd0a61-852c-4f38-af2b-9c84e152873e","status":409,"errorCode":2001,"errors":["Duplicated entity: 2acd0a61-852c-4f38-af2b-9c84e152873e"]},{"id":"30fc6d99-ceb7-45f4-98f9-27e08c5f2d4c","status":409,"errorCode":2001,"errors":["Duplicated entity: 30fc6d99-ceb7-45f4-98f9-27e08c5f2d4c"]}],"SeedSchedule":[{"id":"e71b122c-ddff-4306-84ba-e230527d445c","status":409,"errorCode":2001,"errors":["Duplicated entity: e71b122c-ddff-4306-84ba-e230527d445c"]}]}`,
			expectedResponse: gjson.Parse(`{"Processor":[{"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","status":409,"errorCode":2001,"errors":["Duplicated entity: 516d4a8a-e8ae-488c-9e37-d5746a907454"]},{"id":"5536e705-0c5d-4711-9a99-e272b32948cb","status":201},{"id":"7b192ea1-ac43-439b-9396-5e022f81f2cb","status":409,"errorCode":2001,"errors":["Duplicated entity: 7b192ea1-ac43-439b-9396-5e022f81f2cb"]},{"id":"87d85a14-fb17-4899-8bdc-9fc2b1b28857","status":409,"errorCode":2001,"errors":["Duplicated entity: 87d85a14-fb17-4899-8bdc-9fc2b1b28857"]},{"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8","status":409,"errorCode":2001,"errors":["Duplicated entity: aa0186f1-746f-4b20-b1b0-313bd79e78b8"]}],"Pipeline":[{"id":"8f13eae4-73a5-45c9-9239-5c1996c0378f","status":409,"errorCode":2001,"errors":["Duplicated entity: 8f13eae4-73a5-45c9-9239-5c1996c0378f"]},{"id":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","status":409,"errorCode":2001,"errors":["Duplicated entity: 9a74bf3a-eb2a-4334-b803-c92bf1bc45fe"]}],"Seed":[{"id":"2acd0a61-852c-4f38-af2b-9c84e152873e","status":409,"errorCode":2001,"errors":["Duplicated entity: 2acd0a61-852c-4f38-af2b-9c84e152873e"]},{"id":"30fc6d99-ceb7-45f4-98f9-27e08c5f2d4c","status":409,"errorCode":2001,"errors":["Duplicated entity: 30fc6d99-ceb7-45f4-98f9-27e08c5f2d4c"]}],"SeedSchedule":[{"id":"e71b122c-ddff-4306-84ba-e230527d445c","status":409,"errorCode":2001,"errors":["Duplicated entity: e71b122c-ddff-4306-84ba-e230527d445c"]}]}`),
			onConflict:       OnConflictFail,
			err:              nil,
		},

		// Error cases
		{
			name:       "Restore fails with method not allowed",
			method:     http.MethodPost,
			path:       "/import",
			statusCode: http.StatusMethodNotAllowed,
			onConflict: OnConflictUpdate,
			response: `{
				"status": 405,
				"code": 1001,
				"messages": [
					"Method [GET] not allowed for URI [/v2/import]. Allowed methods: [POST]"
				],
				"timestamp": "2025-08-25T21:09:12.607204700Z"
			}`,
			expectedResponse: gjson.Result{},
			err: Error{Status: http.StatusMethodNotAllowed, Body: gjson.Parse(`{
				"status": 405,
				"code": 1001,
				"messages": [
					"Method [GET] not allowed for URI [/v2/import]. Allowed methods: [POST]"
				],
				"timestamp": "2025-08-25T21:09:12.607204700Z"
			}`)},
		},
		{
			name:             "Restore fails with invalid on conflict.",
			method:           http.MethodPost,
			path:             "/import",
			statusCode:       http.StatusBadRequest,
			response:         `{"status":400,"code":3002,"messages":["Failed to convert argument [onConflict] for value "test" due to: No enum constant com.pureinsights.pdp.core.data.backup.ImportOptions.ConflictResolutionStrategy.test"],"timestamp":"2025-08-25T19:48:41.068729500Z"}`,
			expectedResponse: gjson.Result{},
			onConflict:       OnConflictUpdate,
			err:              Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{"status":400,"code":3002,"messages":["Failed to convert argument [onConflict] for value "test" due to: No enum constant com.pureinsights.pdp.core.data.backup.ImportOptions.ConflictResolutionStrategy.test"],"timestamp":"2025-08-25T19:48:41.068729500Z"}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/json", string(tc.response), func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, tc.path, r.URL.Path)
					body, _ := io.ReadAll(r.Body)
					assert.Contains(t, string(body), string(bytes))
				}))
			defer srv.Close()

			b := backupRestore{client: newClient(srv.URL, "")}
			response, err := b.Import(tc.onConflict, fileName)
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

// Test_backupRestore_Export has table-driven tests to test the Export method.
func Test_backupRestore_Export(t *testing.T) {
	fileName := "testdata/test-export.zip"
	bytes, err := os.ReadFile(fileName)
	require.NoError(t, err)

	tests := []struct {
		name          string
		method        string
		path          string
		statusCode    int
		response      []byte
		expectedTexts []string
		err           error
	}{
		// Working case
		{
			name:          "Backup exports the entities.",
			method:        http.MethodGet,
			path:          "/export",
			statusCode:    http.StatusOK,
			response:      bytes,
			expectedTexts: []string{"Processor.ndjson", "Pipeline.ndjson", "Seed.ndjson"},
			err:           nil,
		},

		// Error case
		{
			name:       "Backup fails",
			method:     http.MethodGet,
			path:       "/export",
			statusCode: http.StatusNotFound,
			response: []byte(`{
				"status": 404,
				"code": 1003,
				"messages": [
					"Page Not Found"
				],
				"timestamp": "2025-08-25T20:25:49.609245Z"
			}`),
			err: Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
				"status": 404,
				"code": 1003,
				"messages": [
					"Page Not Found"
				],
				"timestamp": "2025-08-25T20:25:49.609245Z"
			}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/octet-stream", string(tc.response), func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, tc.path, r.URL.Path)
				}))
			defer srv.Close()

			b := backupRestore{client: newClient(srv.URL, "")}
			response, err := b.Export()
			if tc.err == nil {
				require.NoError(t, err)
				assert.Equal(t, bytes, response)
				for _, text := range tc.expectedTexts {
					assert.Contains(t, string(response), text)
				}
			} else {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

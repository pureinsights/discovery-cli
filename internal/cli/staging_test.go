package cli

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// WorkingStagingBucketControllerNoConflict simulates when the StagingBucketController works.
type WorkingStagingBucketControllerNoConflict struct {
	mock.Mock
}

// Create returns a working result.
func (s *WorkingStagingBucketControllerNoConflict) Create(string, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// Delete implements the interface.
func (s *WorkingStagingBucketControllerNoConflict) Delete(string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// Get returns a bucket.
func (s *WorkingStagingBucketControllerNoConflict) Get(string) (gjson.Result, error) {
	return gjson.Parse(`{
  "name": "my-bucket",
  "documentCount": {},
  "indices": [
    {
      "name": "myIndexA",
      "fields": [
        {
          "fieldName": "DESC"
        }
      ],
      "unique": false
    },
    {
      "name": "myIndexC",
      "fields": [
        {
          "my-field": "DESC"
        }
      ],
      "unique": false
    }
  ]
}`), nil
}

// CreateIndex implements the interface.
func (s *WorkingStagingBucketControllerNoConflict) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// DeleteIndex implements the interface.
func (s *WorkingStagingBucketControllerNoConflict) DeleteIndex(bucket, index string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// WorkingStagingBucketControllerNameConflict simulates when the bucket already exists, but the updates succeed.
type WorkingStagingBucketControllerNameConflict struct {
	mock.Mock
	call int
}

// Create returns a conflict error.
func (s *WorkingStagingBucketControllerNameConflict) Create(string, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

// Delete implements the interface.
func (s *WorkingStagingBucketControllerNameConflict) Delete(string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// Get returns different results based on the number of calls to simulate that the update of the indices worked.
func (s *WorkingStagingBucketControllerNameConflict) Get(string) (gjson.Result, error) {
	s.call++

	if s.call%2 == 1 {
		return gjson.Parse(`{
  "name": "my-bucket",
  "documentCount": {},
  "indices": [
    {
      "name": "myIndexA",
      "fields": [
        {
          "fieldName": "DESC"
        }
      ],
      "unique": false
    },
    {
      "name": "myIndexC",
      "fields": [
        {
          "my-field": "DESC"
        }
      ],
      "unique": false
    }
  ]
}`), nil
	}

	return gjson.Parse(`{
  "name": "my-bucket",
  "documentCount": {},
  "indices": [
    {
      "name": "myIndexA",
      "fields": [
        { "fieldA": "ASC" },
        { "fieldB": "DESC" }
      ],
      "unique": true
    },
    {
      "name": "myIndexB",
      "fields": [
        { "fieldB": "ASC" },
        { "fieldA": "DESC" }
      ],
      "unique": true
    }
  ]
}`), nil
}

// CreateIndex simulates a working Index update.
func (s *WorkingStagingBucketControllerNameConflict) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// DeleteIndex simulates a working Index deletion.
func (s *WorkingStagingBucketControllerNameConflict) DeleteIndex(bucket, index string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// FailingStagingBucketControllerNotDiscoveryError mocks when the create request does not return a Discovery error.
type FailingStagingBucketControllerNotDiscoveryError struct {
	mock.Mock
}

// Create returns a different error.
func (s *FailingStagingBucketControllerNotDiscoveryError) Create(string, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), errors.New("different error")
}

// Delete implements the interface.
func (s *FailingStagingBucketControllerNotDiscoveryError) Delete(string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// Get returns a bucket.
func (s *FailingStagingBucketControllerNotDiscoveryError) Get(string) (gjson.Result, error) {
	return gjson.Parse(`{
  "name": "test",
  "documentCount": {},
  "indices": [
    {
      "name": "myIndexA",
      "fields": [
        {
          "fieldName": "DESC"
        }
      ],
      "unique": false
    },
    {
      "name": "myIndexC",
      "fields": [
        {
          "my-field": "DESC"
        }
      ],
      "unique": false
    }
  ]
}`), nil
}

// CreateIndex implements the interface.
func (s *FailingStagingBucketControllerNotDiscoveryError) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// DeleteIndex implements the interface.
func (s *FailingStagingBucketControllerNotDiscoveryError) DeleteIndex(bucket, index string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// FailingStagingBucketControllerNotFoundError mocks when the function receives a Discovery error that is not a conflict.
type FailingStagingBucketControllerNotFoundError struct {
	mock.Mock
}

// Create returns not found error.
func (s *FailingStagingBucketControllerNotFoundError) Create(string, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

// Delete implements the interface.
func (s *FailingStagingBucketControllerNotFoundError) Delete(string) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "status": 404,
  "code": 1002,
  "messages": [
    "The bucket 'my-bucket' was not found."
  ],
  "timestamp": "2025-12-23T14:53:32.321524600Z"
}`)}
}

// Get returns a bucket.
func (s *FailingStagingBucketControllerNotFoundError) Get(string) (gjson.Result, error) {
	return gjson.Parse(`{
  "name": "test",
  "documentCount": {},
  "indices": [
    {
      "name": "myIndexA",
      "fields": [
        {
          "fieldName": "DESC"
        }
      ],
      "unique": false
    },
    {
      "name": "myIndexC",
      "fields": [
        {
          "my-field": "DESC"
        }
      ],
      "unique": false
    }
  ]
}`), nil
}

// CreateIndex implements the interface.
func (s *FailingStagingBucketControllerNotFoundError) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// DeleteIndex implements the interface.
func (s *FailingStagingBucketControllerNotFoundError) DeleteIndex(bucket, index string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// FailingStagingBucketControllerIndexCreationFails mocks a failing index creation.
type FailingStagingBucketControllerIndexCreationFails struct {
	mock.Mock
}

// Create returns a conflict to make the function go through that path.
func (s *FailingStagingBucketControllerIndexCreationFails) Create(string, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

// Delete implements the interface.
func (s *FailingStagingBucketControllerIndexCreationFails) Delete(string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// Get returns a bucket.
func (s *FailingStagingBucketControllerIndexCreationFails) Get(string) (gjson.Result, error) {
	return gjson.Parse(`{
  "name": "test",
  "documentCount": {},
  "indices": [
    {
      "name": "myIndexA",
      "fields": [
        {
          "fieldName": "DESC"
        }
      ],
      "unique": false
    },
    {
      "name": "myIndexC",
      "fields": [
        {
          "my-field": "DESC"
        }
      ],
      "unique": false
    }
  ]
}`), nil
}

// CreateIndex returns an error.
func (s *FailingStagingBucketControllerIndexCreationFails) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

// DeleteIndex implements the interface.
func (s *FailingStagingBucketControllerIndexCreationFails) DeleteIndex(bucket, index string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// FailingStagingBucketControllerIndexDeletionFails simulates when deleting an index fails.
type FailingStagingBucketControllerIndexDeletionFails struct {
	mock.Mock
}

// Create returns a conflict error.
func (s *FailingStagingBucketControllerIndexDeletionFails) Create(string, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

// Delete implements the interface.
func (s *FailingStagingBucketControllerIndexDeletionFails) Delete(string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// Get returns a bucket.
func (s *FailingStagingBucketControllerIndexDeletionFails) Get(string) (gjson.Result, error) {
	return gjson.Parse(`{
  "name": "test",
  "documentCount": {},
  "indices": [
    {
      "name": "myIndexA",
      "fields": [
        {
          "fieldName": "DESC"
        }
      ],
      "unique": false
    },
    {
      "name": "myIndexC",
      "fields": [
        {
          "my-field": "DESC"
        }
      ],
      "unique": false
    }
  ]
}`), nil
}

// CreateIndex implements the interface.
func (s *FailingStagingBucketControllerIndexDeletionFails) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// DeleteIndex returns an error.
func (s *FailingStagingBucketControllerIndexDeletionFails) DeleteIndex(bucket, index string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

// FailingStagingBucketControllerLastGetFails simulates when the last get of the bucket fails.
type FailingStagingBucketControllerLastGetFails struct {
	mock.Mock
}

// Create implements the interface.
func (s *FailingStagingBucketControllerLastGetFails) Create(string, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// Delete implements the interface.
func (s *FailingStagingBucketControllerLastGetFails) Delete(string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// Get implements the interface.
func (s *FailingStagingBucketControllerLastGetFails) Get(string) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "status": 404,
  "code": 1002,
  "messages": [
    "The bucket 'my-bucket' was not found."
  ],
  "timestamp": "2025-12-22T21:29:24.255774300Z"
}`)}
}

// CreateIndex implements the interface.
func (s *FailingStagingBucketControllerLastGetFails) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// DeleteIndex implements the interface.
func (s *FailingStagingBucketControllerLastGetFails) DeleteIndex(bucket, index string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// FailingStagingBucketControllerFirstGetFails mocks when the first get fails.
type FailingStagingBucketControllerFirstGetFails struct {
	mock.Mock
}

// Create returns a conflict.
func (s *FailingStagingBucketControllerFirstGetFails) Create(string, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

// Delete implements the interface.
func (s *FailingStagingBucketControllerFirstGetFails) Delete(string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// Get returns an error.
func (s *FailingStagingBucketControllerFirstGetFails) Get(string) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "status": 404,
  "code": 1002,
  "messages": [
    "The bucket 'my-bucket' was not found."
  ],
  "timestamp": "2025-12-22T21:29:24.255774300Z"
}`)}
}

// CreateIndex implements the interface.
func (s *FailingStagingBucketControllerFirstGetFails) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// DeleteIndex implements the interface.
func (s *FailingStagingBucketControllerFirstGetFails) DeleteIndex(bucket, index string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// Test_updateIndices tests the updateIndices() function.
func Test_updateIndices(t *testing.T) {
	tests := []struct {
		name       string
		client     StagingBucketController
		newIndices gjson.Result
		oldIndices []gjson.Result
		bucketName string
		err        error
	}{
		// Working case
		{
			name:   "updateIndices works correctly",
			client: new(WorkingStagingBucketControllerNameConflict),
			newIndices: gjson.Parse(`[
    {
      "name": "myIndexA",
      "fields": [
        { "fieldA": "ASC" },
        { "fieldB": "DESC" }
      ],
      "unique": true
    },
    {
      "name": "myIndexB",
      "fields": [
        { "fieldB": "ASC" },
        { "fieldA": "DESC" }
      ],
      "unique": true
    }
  ]`),
			oldIndices: gjson.Parse(`[
    {
      "name": "myIndexA",
      "fields": [
        {
          "fieldName": "DESC"
        }
      ],
      "unique": false
    },
    {
      "name": "myIndexC",
      "fields": [
        {
          "my-field": "DESC"
        }
      ],
      "unique": false
    }
  ]`).Array(),
			bucketName: "my-bucket",
			err:        nil,
		},
		{
			name:   "CreateIndex fails",
			client: new(FailingStagingBucketControllerIndexCreationFails),
			newIndices: gjson.Parse(`[
    {
      "name": "myIndexA",
      "fields": [
        { "fieldA": "ASC" },
        { "fieldB": "DESC" }
      ],
      "unique": true
    },
    {
      "name": "myIndexB",
      "fields": [
        { "fieldB": "ASC" },
        { "fieldA": "DESC" }
      ],
      "unique": true
    }
  ]`),
			oldIndices: gjson.Parse(`[
    {
      "name": "myIndexA",
      "fields": [
        {
          "fieldName": "DESC"
        }
      ],
      "unique": false
    },
    {
      "name": "myIndexC",
      "fields": [
        {
          "my-field": "DESC"
        }
      ],
      "unique": false
    }
  ]`).Array(),
			bucketName: "my-bucket",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
  "acknowledged": false
}`)}, `Could not update index with name "myIndexA" of bucket "my-bucket".`),
		},
		{
			name:   "DeleteIndex fails",
			client: new(FailingStagingBucketControllerIndexDeletionFails),
			newIndices: gjson.Parse(`[
    {
      "name": "myIndexA",
      "fields": [
        { "fieldA": "ASC" },
        { "fieldB": "DESC" }
      ],
      "unique": true
    },
    {
      "name": "myIndexB",
      "fields": [
        { "fieldB": "ASC" },
        { "fieldA": "DESC" }
      ],
      "unique": true
    }
  ]`),
			oldIndices: gjson.Parse(`[
    {
      "name": "myIndexA",
      "fields": [
        {
          "fieldName": "DESC"
        }
      ],
      "unique": false
    },
    {
      "name": "myIndexC",
      "fields": [
        {
          "my-field": "DESC"
        }
      ],
      "unique": false
    }
  ]`).Array(),
			bucketName: "my-bucket",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "acknowledged": false
}`)}, `Could not delete index with name "myIndexC" of bucket "my-bucket".`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := updateIndices(tc.client, tc.bucketName, tc.oldIndices, tc.newIndices)

			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test_discovery_StoreBucket tests the discovery.StoreBucket() function.
func Test_discovery_StoreBucket(t *testing.T) {
	tests := []struct {
		name           string
		client         StagingBucketController
		config         gjson.Result
		bucketName     string
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:       "StoreBucket returns the created bucket",
			client:     new(WorkingStagingBucketControllerNoConflict),
			bucketName: "my-bucket",
			config: gjson.Parse(`{
  "indices": [
    {
      "name": "myIndexA",
      "fields": [
        { "fieldA": "ASC" },
        { "fieldB": "DESC" }
      ],
      "unique": true
    },
    {
      "name": "myIndexB",
      "fields": [
        { "fieldB": "ASC" },
        { "fieldA": "DESC" }
      ],
      "unique": true
    }
  ],
  "config":{}
}`),
			printer:        nil,
			expectedOutput: "{\n  \"documentCount\": {},\n  \"indices\": [\n    {\n      \"fields\": [\n        {\n          \"fieldName\": \"DESC\"\n        }\n      ],\n      \"name\": \"myIndexA\",\n      \"unique\": false\n    },\n    {\n      \"fields\": [\n        {\n          \"my-field\": \"DESC\"\n        }\n      ],\n      \"name\": \"myIndexC\",\n      \"unique\": false\n    }\n  ],\n  \"name\": \"my-bucket\"\n}\n",
			err:            nil,
		},
		{
			name:       "StoreBucket correctly updates indices",
			client:     new(WorkingStagingBucketControllerNameConflict),
			bucketName: "my-bucket",
			config: gjson.Parse(`{
  "indices": [
    {
      "name": "myIndexA",
      "fields": [
        { "fieldA": "ASC" },
        { "fieldB": "DESC" }
      ],
      "unique": true
    },
    {
      "name": "myIndexB",
      "fields": [
        { "fieldB": "ASC" },
        { "fieldA": "DESC" }
      ],
      "unique": true
    }
  ],
  "config":{}
}`),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"documentCount\":{},\"indices\":[{\"fields\":[{\"fieldA\":\"ASC\"},{\"fieldB\":\"DESC\"}],\"name\":\"myIndexA\",\"unique\":true},{\"fields\":[{\"fieldB\":\"ASC\"},{\"fieldA\":\"DESC\"}],\"name\":\"myIndexB\",\"unique\":true}],\"name\":\"my-bucket\"}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "StoreBucket fails with not a discovery error",
			client:         new(FailingStagingBucketControllerNotDiscoveryError),
			bucketName:     "my-bucket",
			config:         gjson.Parse(``),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"creationTimestamp\":\"2025-09-04T19:29:41.119013Z\",\"id\":\"a056c7fb-0ca1-45f6-97ea-ec849a0701fd\",\"lastUpdatedTimestamp\":\"2025-09-04T19:29:41.119013Z\",\"properties\":{\"stagingBucket\":\"testBucket\"},\"scanType\":\"INCREMENTAL\",\"status\":\"CREATED\",\"triggerType\":\"MANUAL\"}\n",
			err:            NewErrorWithCause(ErrorExitCode, errors.New("different error"), "Could not create bucket with name \"my-bucket\"."),
		},
		{
			name:           "StoreBucket fails with a discovery not found error",
			client:         new(FailingStagingBucketControllerNotFoundError),
			bucketName:     "my-bucket",
			config:         gjson.Parse(``),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"creationTimestamp\":\"2025-09-04T19:29:41.119013Z\",\"id\":\"a056c7fb-0ca1-45f6-97ea-ec849a0701fd\",\"lastUpdatedTimestamp\":\"2025-09-04T19:29:41.119013Z\",\"properties\":{\"stagingBucket\":\"testBucket\"},\"scanType\":\"INCREMENTAL\",\"status\":\"CREATED\",\"triggerType\":\"MANUAL\"}\n",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "acknowledged": false
}`)}, "Could not create bucket with name \"my-bucket\"."),
		},
		{
			name:       "StoreBucket fails with discovery conflict error, but has no indices in config",
			client:     new(WorkingStagingBucketControllerNameConflict),
			bucketName: "my-bucket",
			config: gjson.Parse(`{
	"config":{}
}`),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"creationTimestamp\":\"2025-09-04T19:29:41.119013Z\",\"id\":\"a056c7fb-0ca1-45f6-97ea-ec849a0701fd\",\"lastUpdatedTimestamp\":\"2025-09-04T19:29:41.119013Z\",\"properties\":{\"stagingBucket\":\"testBucket\"},\"scanType\":\"INCREMENTAL\",\"status\":\"CREATED\",\"triggerType\":\"MANUAL\"}\n",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
  "acknowledged": false
}`)}, "Could not create bucket with name \"my-bucket\"."),
		},
		{
			name:       "Printing fails",
			bucketName: "my-bucket",
			client:     new(WorkingStagingBucketControllerNoConflict),
			printer:    nil,
			outWriter:  testutils.ErrWriter{Err: errors.New("write failed")},
			err:        NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
		},
		{
			name:   "Last Get fails",
			client: new(FailingStagingBucketControllerLastGetFails),
			config: gjson.Parse(`{
  "indices": [
    {
      "name": "myIndexA",
      "fields": [
        { "fieldA": "ASC" },
        { "fieldB": "DESC" }
      ],
      "unique": true
    },
    {
      "name": "myIndexB",
      "fields": [
        { "fieldB": "ASC" },
        { "fieldA": "DESC" }
      ],
      "unique": true
    }
  ],
  "config":{}
}`),
			bucketName: "my-bucket",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "status": 404,
  "code": 1002,
  "messages": [
    "The bucket 'my-bucket' was not found."
  ],
  "timestamp": "2025-12-22T21:29:24.255774300Z"
}`)}, `Could not get the information of bucket with name "my-bucket".`),
		},
		{
			name:   "First Get fails",
			client: new(FailingStagingBucketControllerFirstGetFails),
			config: gjson.Parse(`{
  "indices": [
    {
      "name": "myIndexA",
      "fields": [
        { "fieldA": "ASC" },
        { "fieldB": "DESC" }
      ],
      "unique": true
    },
    {
      "name": "myIndexB",
      "fields": [
        { "fieldB": "ASC" },
        { "fieldA": "DESC" }
      ],
      "unique": true
    }
  ],
  "config":{}
}`),
			bucketName: "my-bucket",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "status": 404,
  "code": 1002,
  "messages": [
    "The bucket 'my-bucket' was not found."
  ],
  "timestamp": "2025-12-22T21:29:24.255774300Z"
}`)}, `Could not get bucket with name "my-bucket" to update it.`),
		},
		{
			name:   "CreateIndex fails",
			client: new(FailingStagingBucketControllerIndexCreationFails),
			config: gjson.Parse(`{
  "indices": [
    {
      "name": "myIndexA",
      "fields": [
        { "fieldA": "ASC" },
        { "fieldB": "DESC" }
      ],
      "unique": true
    },
    {
      "name": "myIndexB",
      "fields": [
        { "fieldB": "ASC" },
        { "fieldA": "DESC" }
      ],
      "unique": true
    }
  ],
  "config":{}
}`),
			bucketName: "my-bucket",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
  "acknowledged": false
}`)}, `Could not update index with name "myIndexA" of bucket "my-bucket".`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			var out io.Writer
			if tc.outWriter != nil {
				out = tc.outWriter
			} else {
				out = buf
			}

			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: out,
				Err: os.Stderr,
			}

			d := NewDiscovery(&ios, viper.New(), "")
			err := d.StoreBucket(tc.client, tc.bucketName, tc.config, tc.printer)

			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}

// Test_discovery_DeleteBucket tests the discovery.DeleteBucket() function.
func Test_discovery_DeleteBucket(t *testing.T) {
	tests := []struct {
		name           string
		client         StagingBucketController
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "DeleteBucket correctly prints the deletion confirmation with the pretty printer",
			client:         new(WorkingStagingBucketControllerNoConflict),
			printer:        nil,
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		{
			name:           "DeleteBucket correctly prints an object with JSON ugly printer",
			client:         new(WorkingStagingBucketControllerNoConflict),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"acknowledged\":true}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "Delete returns 404 Bad Request",
			client:         new(FailingStagingBucketControllerNotFoundError),
			printer:        nil,
			expectedOutput: "",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
  "status": 404,
  "code": 1002,
  "messages": [
    "The bucket 'my-bucket' was not found."
  ],
  "timestamp": "2025-12-23T14:53:32.321524600Z"
}`),
			}, "Could not delete the bucket with name \"my-bucket\"."),
		},
		{
			name:      "Printing fails",
			client:    new(WorkingStagingBucketControllerNoConflict),
			printer:   nil,
			outWriter: testutils.ErrWriter{Err: errors.New("write failed")},
			err:       NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			var out io.Writer
			if tc.outWriter != nil {
				out = tc.outWriter
			} else {
				out = buf
			}

			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: out,
				Err: os.Stderr,
			}

			d := NewDiscovery(&ios, viper.New(), "")
			err := d.DeleteBucket(tc.client, "my-bucket", tc.printer)

			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}

// WorkingStagingContentController mocks a working content controller.
type WorkingStagingContentController struct {
	mock.Mock
}

// Scroll implements the interface.
func (s *WorkingStagingContentController) Scroll(filters, projections gjson.Result, size *int) ([]gjson.Result, error) {
	return gjson.Parse(`[
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
    }
]`).Array(), nil
}

// WorkingStagingContentControllerNoContent mocks when the scroll returns no content.
type WorkingStagingContentControllerNoContent struct {
	mock.Mock
}

// Scroll returns an empty array.
func (s *WorkingStagingContentControllerNoContent) Scroll(filters, projections gjson.Result, size *int) ([]gjson.Result, error) {
	return []gjson.Result{}, nil
}

// FailingStagingContentController mocks a failing content controller.
type FailingStagingContentController struct {
	mock.Mock
}

// Scroll returns an error.
func (s *FailingStagingContentController) Scroll(filters, projections gjson.Result, size *int) ([]gjson.Result, error) {
	return []gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
  "status": 404,
  "code": 1002,
  "messages": [
    "The bucket 'my-bucket' was not found."
  ],
  "timestamp": "2025-12-23T14:53:32.321524600Z"
}`),
	}
}

// Test_discovery_DumpBucket tests the discovery.DumpBucket() function.
func Test_discovery_DumpBucket(t *testing.T) {
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
	tests := []struct {
		name           string
		client         StagingContentController
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "DumpBucket correctly prints the received records with the ugly printer",
			client:         new(WorkingStagingContentController),
			printer:        nil,
			expectedOutput: "{\"action\":\"STORE\",\"checksum\":\"58b3d1b06729f1491373b97fd8287ae1\",\"content\":{\"_id\":\"5625c64483bef0d48e9ad91aca9b2f94\",\"author\":\"Graham Gillen\",\"header\":\"Pureinsights Named MongoDB's 2024 AI Partner of the Year - Pureinsights: PRESS RELEASE - Pureinsights named MongoDB's Service AI Partner of the Year for 2024 and also joins the MongoDB AI Application Program (MAAP).\",\"link\":\"https://pureinsights.com/blog/2024/pureinsights-named-mongodbs-2024-ai-partner-of-the-year/\"},\"creationTimestamp\":\"2025-12-26T16:28:38Z\",\"id\":\"1\",\"lastUpdatedTimestamp\":\"2025-12-26T16:28:38Z\",\"transaction\":\"694eb7b678aedc7a163da8ff\"}\n{\"action\":\"STORE\",\"checksum\":\"b76292db9fd1c7aef145512dce131f4d\",\"content\":{\"_id\":\"768b0a3bcee501dc624484ba8a0d7f6d\",\"author\":\"Matt Willsmore\",\"header\":\"5 Challenges Implementing Retrieval Augmented Generation (RAG) - Pureinsights: A blog on 5 common challenges when implementing RAG (Retrieval Augmented Generation) and possible solutions for search applications.\",\"link\":\"https://pureinsights.com/blog/2024/five-common-challenges-when-implementing-rag-retrieval-augmented-generation/\"},\"creationTimestamp\":\"2025-12-26T16:28:46Z\",\"id\":\"2\",\"lastUpdatedTimestamp\":\"2025-12-26T16:28:46Z\",\"transaction\":\"694eb7be78aedc7a163da900\"}\n",
			err:            nil,
		},
		{
			name:           "DumpBucket correctly prints the array with JSON pretty printer",
			client:         new(WorkingStagingContentController),
			printer:        JsonArrayPrinter(true),
			expectedOutput: "[\n{\n  \"action\": \"STORE\",\n  \"checksum\": \"58b3d1b06729f1491373b97fd8287ae1\",\n  \"content\": {\n    \"_id\": \"5625c64483bef0d48e9ad91aca9b2f94\",\n    \"author\": \"Graham Gillen\",\n    \"header\": \"Pureinsights Named MongoDB's 2024 AI Partner of the Year - Pureinsights: PRESS RELEASE - Pureinsights named MongoDB's Service AI Partner of the Year for 2024 and also joins the MongoDB AI Application Program (MAAP).\",\n    \"link\": \"https://pureinsights.com/blog/2024/pureinsights-named-mongodbs-2024-ai-partner-of-the-year/\"\n  },\n  \"creationTimestamp\": \"2025-12-26T16:28:38Z\",\n  \"id\": \"1\",\n  \"lastUpdatedTimestamp\": \"2025-12-26T16:28:38Z\",\n  \"transaction\": \"694eb7b678aedc7a163da8ff\"\n},\n{\n  \"action\": \"STORE\",\n  \"checksum\": \"b76292db9fd1c7aef145512dce131f4d\",\n  \"content\": {\n    \"_id\": \"768b0a3bcee501dc624484ba8a0d7f6d\",\n    \"author\": \"Matt Willsmore\",\n    \"header\": \"5 Challenges Implementing Retrieval Augmented Generation (RAG) - Pureinsights: A blog on 5 common challenges when implementing RAG (Retrieval Augmented Generation) and possible solutions for search applications.\",\n    \"link\": \"https://pureinsights.com/blog/2024/five-common-challenges-when-implementing-rag-retrieval-augmented-generation/\"\n  },\n  \"creationTimestamp\": \"2025-12-26T16:28:46Z\",\n  \"id\": \"2\",\n  \"lastUpdatedTimestamp\": \"2025-12-26T16:28:46Z\",\n  \"transaction\": \"694eb7be78aedc7a163da900\"\n}\n]\n",
			err:            nil,
		},
		{
			name:           "DumpBucket correctly prints nothing with no content returned.",
			client:         new(WorkingStagingContentControllerNoContent),
			printer:        nil,
			expectedOutput: "",
			err:            nil,
		},

		// Error case
		{
			name:           "Dump returns an error",
			client:         new(FailingStagingContentController),
			printer:        nil,
			expectedOutput: "",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
  "status": 404,
  "code": 1002,
  "messages": [
    "The bucket 'my-bucket' was not found."
  ],
  "timestamp": "2025-12-23T14:53:32.321524600Z"
}`),
			}, "Could not scroll the bucket with name \"my-bucket\"."),
		},
		{
			name:      "Printing fails",
			client:    new(WorkingStagingContentController),
			printer:   nil,
			outWriter: testutils.ErrWriter{Err: errors.New("write failed")},
			err:       NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON Array"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			var out io.Writer
			if tc.outWriter != nil {
				out = tc.outWriter
			} else {
				out = buf
			}

			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: out,
				Err: os.Stderr,
			}

			d := NewDiscovery(&ios, viper.New(), "")
			max := 3
			err := d.DumpBucket(tc.client, "my-bucket", "", gjson.Parse(filters), gjson.Parse(projections), &max, tc.printer)

			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}

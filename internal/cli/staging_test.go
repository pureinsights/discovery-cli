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

type WorkingStagingBucketControllerNoConflict struct {
	mock.Mock
}

func (s *WorkingStagingBucketControllerNoConflict) Create(string, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

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

func (s *WorkingStagingBucketControllerNoConflict) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Result{}, nil
}

func (s *WorkingStagingBucketControllerNoConflict) DeleteIndex(bucket, index string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

type WorkingStagingBucketControllerNameConflict struct {
	mock.Mock
	call int
}

func (s *WorkingStagingBucketControllerNameConflict) Create(string, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

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

func (s *WorkingStagingBucketControllerNameConflict) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

func (s *WorkingStagingBucketControllerNameConflict) DeleteIndex(bucket, index string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

type FailingStagingBucketControllerNotDiscoveryError struct {
	mock.Mock
}

func (s *FailingStagingBucketControllerNotDiscoveryError) Create(string, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), errors.New("different error")
}

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

func (s *FailingStagingBucketControllerNotDiscoveryError) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

func (s *FailingStagingBucketControllerNotDiscoveryError) DeleteIndex(bucket, index string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

type FailingStagingBucketControllerNotFoundError struct {
	mock.Mock
}

func (s *FailingStagingBucketControllerNotFoundError) Create(string, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

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

func (s *FailingStagingBucketControllerNotFoundError) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

func (s *FailingStagingBucketControllerNotFoundError) DeleteIndex(bucket, index string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

type FailingStagingBucketControllerIndexCreationFails struct {
	mock.Mock
}

func (s *FailingStagingBucketControllerIndexCreationFails) Create(string, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

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

func (s *FailingStagingBucketControllerIndexCreationFails) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

func (s *FailingStagingBucketControllerIndexCreationFails) DeleteIndex(bucket, index string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

type FailingStagingBucketControllerIndexDeletionFails struct {
	mock.Mock
}

func (s *FailingStagingBucketControllerIndexDeletionFails) Create(string, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

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

func (s *FailingStagingBucketControllerIndexDeletionFails) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

func (s *FailingStagingBucketControllerIndexDeletionFails) DeleteIndex(bucket, index string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

type FailingStagingBucketControllerLastGetFails struct {
	mock.Mock
}

func (s *FailingStagingBucketControllerLastGetFails) Create(string, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

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

func (s *FailingStagingBucketControllerLastGetFails) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Result{}, nil
}

func (s *FailingStagingBucketControllerLastGetFails) DeleteIndex(bucket, index string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

type FailingStagingBucketControllerFirstGetFails struct {
	mock.Mock
}

func (s *FailingStagingBucketControllerFirstGetFails) Create(string, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

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

func (s *FailingStagingBucketControllerFirstGetFails) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

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

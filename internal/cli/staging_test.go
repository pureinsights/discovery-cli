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

func (s *WorkingStagingBucketControllerNoConflict) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Result{}, nil
}

type WorkingStagingBucketControllerNameConflict struct {
	mock.Mock
}

func (s *WorkingStagingBucketControllerNameConflict) Create(string, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

func (s *WorkingStagingBucketControllerNameConflict) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
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

func (s *FailingStagingBucketControllerNotDiscoveryError) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
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

func (s *FailingStagingBucketControllerNotFoundError) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
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

func (s *FailingStagingBucketControllerIndexCreationFails) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

// Test_updateIndices tests the updateIndices() function.
func Test_updateIndices(t *testing.T) {
	tests := []struct {
		name           string
		client         StagingBucketController
		indices        []gjson.Result
		bucketName     string
		expectedResult gjson.Result
		err            error
	}{
		// Working case
		{
			name:   "updateIndices works correctly",
			client: new(WorkingStagingBucketControllerNameConflict),
			indices: gjson.Parse(`[
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
  ]`).Array(),
			bucketName:     "my-bucket",
			expectedResult: gjson.Parse("{\"indices\":{\"myIndexA\":{\n  \"acknowledged\": true\n},\"myIndexB\":{\n  \"acknowledged\": true\n}}}"),
			err:            nil,
		},
		{
			name:   "CreateIndex fails",
			client: new(FailingStagingBucketControllerIndexCreationFails),
			indices: gjson.Parse(`[
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
  ]`).Array(),
			bucketName:     "my-bucket",
			expectedResult: gjson.Parse("{\"indices\":{\"myIndexA\":\"status: 409, body: {\\n  \\\"acknowledged\\\": false\\n}\\n\",\"myIndexB\":\"status: 409, body: {\\n  \\\"acknowledged\\\": false\\n}\\n\"}}"),
			err:            nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := updateIndices(tc.client, tc.bucketName, tc.indices)

			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
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
			name:       "StoreBucket correctly prints acknowledged true",
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
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		{
			name:       "StoreBucket correctly updates indices and prints out their results",
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
			expectedOutput: "{\"indices\":{\"myIndexA\":{\"acknowledged\":true},\"myIndexB\":{\"acknowledged\":true}}}\n",
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

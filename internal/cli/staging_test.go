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
	"github.com/pureinsights/discovery-cli/internal/testutils/mocks"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

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
			client: new(mocks.WorkingStagingBucketControllerNameConflict),
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
			client: new(mocks.FailingStagingBucketControllerIndexCreationFails),
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
			client: new(mocks.FailingStagingBucketControllerIndexDeletionFails),
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
			client:     new(mocks.WorkingStagingBucketControllerNoConflict),
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
			client:     new(mocks.WorkingStagingBucketControllerNameConflict),
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
			client:         new(mocks.FailingStagingBucketControllerNotDiscoveryError),
			bucketName:     "my-bucket",
			config:         gjson.Parse(``),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"creationTimestamp\":\"2025-09-04T19:29:41.119013Z\",\"id\":\"a056c7fb-0ca1-45f6-97ea-ec849a0701fd\",\"lastUpdatedTimestamp\":\"2025-09-04T19:29:41.119013Z\",\"properties\":{\"stagingBucket\":\"testBucket\"},\"scanType\":\"INCREMENTAL\",\"status\":\"CREATED\",\"triggerType\":\"MANUAL\"}\n",
			err:            NewErrorWithCause(ErrorExitCode, errors.New("different error"), "Could not create bucket with name \"my-bucket\"."),
		},
		{
			name:           "StoreBucket fails with a discovery not found error",
			client:         new(mocks.FailingStagingBucketControllerNotFoundError),
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
			client:     new(mocks.WorkingStagingBucketControllerNameConflict),
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
			client:     new(mocks.WorkingStagingBucketControllerNoConflict),
			printer:    nil,
			outWriter:  testutils.ErrWriter{Err: errors.New("write failed")},
			err:        NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
		},
		{
			name:   "Last Get fails",
			client: new(mocks.FailingStagingBucketControllerLastGetFails),
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
			client: new(mocks.FailingStagingBucketControllerFirstGetFails),
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
			client: new(mocks.FailingStagingBucketControllerIndexCreationFails),
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
			client:         new(mocks.WorkingStagingBucketControllerNoConflict),
			printer:        nil,
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		{
			name:           "DeleteBucket correctly prints an object with JSON ugly printer",
			client:         new(mocks.WorkingStagingBucketControllerNoConflict),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"acknowledged\":true}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "Delete returns 404 Bad Request",
			client:         new(mocks.FailingStagingBucketControllerNotFoundError),
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
			client:    new(mocks.WorkingStagingBucketControllerNoConflict),
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
			client:         new(mocks.WorkingStagingContentController),
			printer:        nil,
			expectedOutput: "{\"action\":\"STORE\",\"checksum\":\"58b3d1b06729f1491373b97fd8287ae1\",\"content\":{\"_id\":\"5625c64483bef0d48e9ad91aca9b2f94\",\"author\":\"Graham Gillen\",\"header\":\"Pureinsights Named MongoDB's 2024 AI Partner of the Year - Pureinsights: PRESS RELEASE - Pureinsights named MongoDB's Service AI Partner of the Year for 2024 and also joins the MongoDB AI Application Program (MAAP).\",\"link\":\"https://pureinsights.com/blog/2024/pureinsights-named-mongodbs-2024-ai-partner-of-the-year/\"},\"creationTimestamp\":\"2025-12-26T16:28:38Z\",\"id\":\"1\",\"lastUpdatedTimestamp\":\"2025-12-26T16:28:38Z\",\"transaction\":\"694eb7b678aedc7a163da8ff\"}\n{\"action\":\"STORE\",\"checksum\":\"b76292db9fd1c7aef145512dce131f4d\",\"content\":{\"_id\":\"768b0a3bcee501dc624484ba8a0d7f6d\",\"author\":\"Matt Willsmore\",\"header\":\"5 Challenges Implementing Retrieval Augmented Generation (RAG) - Pureinsights: A blog on 5 common challenges when implementing RAG (Retrieval Augmented Generation) and possible solutions for search applications.\",\"link\":\"https://pureinsights.com/blog/2024/five-common-challenges-when-implementing-rag-retrieval-augmented-generation/\"},\"creationTimestamp\":\"2025-12-26T16:28:46Z\",\"id\":\"2\",\"lastUpdatedTimestamp\":\"2025-12-26T16:28:46Z\",\"transaction\":\"694eb7be78aedc7a163da900\"}\n",
			err:            nil,
		},
		{
			name:           "DumpBucket correctly prints the array with JSON pretty printer",
			client:         new(mocks.WorkingStagingContentController),
			printer:        JsonArrayPrinter(true),
			expectedOutput: "[\n{\n  \"action\": \"STORE\",\n  \"checksum\": \"58b3d1b06729f1491373b97fd8287ae1\",\n  \"content\": {\n    \"_id\": \"5625c64483bef0d48e9ad91aca9b2f94\",\n    \"author\": \"Graham Gillen\",\n    \"header\": \"Pureinsights Named MongoDB's 2024 AI Partner of the Year - Pureinsights: PRESS RELEASE - Pureinsights named MongoDB's Service AI Partner of the Year for 2024 and also joins the MongoDB AI Application Program (MAAP).\",\n    \"link\": \"https://pureinsights.com/blog/2024/pureinsights-named-mongodbs-2024-ai-partner-of-the-year/\"\n  },\n  \"creationTimestamp\": \"2025-12-26T16:28:38Z\",\n  \"id\": \"1\",\n  \"lastUpdatedTimestamp\": \"2025-12-26T16:28:38Z\",\n  \"transaction\": \"694eb7b678aedc7a163da8ff\"\n},\n{\n  \"action\": \"STORE\",\n  \"checksum\": \"b76292db9fd1c7aef145512dce131f4d\",\n  \"content\": {\n    \"_id\": \"768b0a3bcee501dc624484ba8a0d7f6d\",\n    \"author\": \"Matt Willsmore\",\n    \"header\": \"5 Challenges Implementing Retrieval Augmented Generation (RAG) - Pureinsights: A blog on 5 common challenges when implementing RAG (Retrieval Augmented Generation) and possible solutions for search applications.\",\n    \"link\": \"https://pureinsights.com/blog/2024/five-common-challenges-when-implementing-rag-retrieval-augmented-generation/\"\n  },\n  \"creationTimestamp\": \"2025-12-26T16:28:46Z\",\n  \"id\": \"2\",\n  \"lastUpdatedTimestamp\": \"2025-12-26T16:28:46Z\",\n  \"transaction\": \"694eb7be78aedc7a163da900\"\n}\n]\n",
			err:            nil,
		},
		{
			name:           "DumpBucket correctly prints nothing with no content returned.",
			client:         new(mocks.WorkingStagingContentControllerNoContent),
			printer:        nil,
			expectedOutput: "",
			err:            nil,
		},

		// Error case
		{
			name:           "Dump returns an error",
			client:         new(mocks.FailingStagingContentController),
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
			client:    new(mocks.WorkingStagingContentController),
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
			err := d.DumpBucket(tc.client, "my-bucket", gjson.Parse(filters), gjson.Parse(projections), &max, tc.printer)

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

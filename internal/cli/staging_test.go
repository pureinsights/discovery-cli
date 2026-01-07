package cli

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

// Test_writeRecordsToFile_AllFilesWritten tests the writeRecordsToFile() function to verify that all the files are written correctly.
func Test_writeRecordsToFile_AllFilesWritten(t *testing.T) {
	records := gjson.Parse(`[
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
]`).Array()

	dir, err := writeRecordsToFile(records, "my-bucket")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	assert.Contains(t, dir, "dump-my-bucket-")

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Equal(t, 2, len(entries))
	assert.Equal(t, "694eb7b678aedc7a163da8ff.json", entries[0].Name())
	assert.Equal(t, "694eb7be78aedc7a163da900.json", entries[1].Name())
}

// Test_writeRecordsToFile_WriteFails tests the writeRecordsToFile() function when writing a file fails.
func Test_writeRecordsToFile_WriteFails(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("TMPDIR", tempDir)
	t.Setenv("TEMP", tempDir)
	t.Setenv("TMP", tempDir)

	records := gjson.Parse(`[
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
            "transaction": "doesnotexist/694eb7be78aedc7a163da900"
    }
]`).Array()

	dir, err := writeRecordsToFile(records, "my-bucket")
	assert.EqualError(t, err, errors.New("the given path does not exist: "+filepath.Join(dir, "doesnotexist", "694eb7be78aedc7a163da900.json")).Error())
}

// Test_writeRecordsToFile_MkDirTempFails tests the writeRecordsToFile() function when it fails because the temporary directory could not be created.
func Test_writeRecordsToFile_MkDirTempFails(t *testing.T) {
	t.Setenv("TMPDIR", "/does/not/exist")
	t.Setenv("TEMP", "/does/not/exist")
	t.Setenv("TMP", "/does/not/exist")

	records := gjson.Parse(`[
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
]`).Array()

	dir, err := writeRecordsToFile(records, "my-bucket")
	assert.Equal(t, "", dir)
	assert.Contains(t, err.Error(), "the given path does not exist:")
	assert.Contains(t, err.Error(), strings.ToLower(filepath.FromSlash("/does/not/exist")))
}

// Test_zipRecords_ZipIsCreated tests the zipRecords() function to verify that the zip contains all of the record files.
func Test_zipRecords_ZipIsCreated(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("TMPDIR", tempDir)
	t.Setenv("TEMP", tempDir)
	t.Setenv("TMP", tempDir)

	records := gjson.Parse(`[
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
]`).Array()

	dir, err := writeRecordsToFile(records, "my-bucket")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	savePath := filepath.Join(t.TempDir(), "my-bucket.zip")
	err = zipRecords(savePath, dir)

	zipFile, err := os.ReadFile(savePath)
	require.NoError(t, err)
	zipReader, err := zip.NewReader(bytes.NewReader(zipFile), int64(len(zipFile)))
	require.NoError(t, err)
	files := make(map[string]*zip.File, len(zipReader.File))
	for _, f := range zipReader.File {
		files[f.Name] = f
	}

	recordFile1, ok := files["694eb7b678aedc7a163da8ff.json"]
	require.True(t, ok)
	fileContent1, err := recordFile1.Open()
	require.NoError(t, err)
	gotBytes, err := io.ReadAll(fileContent1)
	require.NoError(t, err)
	fileContent1.Close()
	assert.Equal(t, `{
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
    }`, string(gotBytes))

	recordFile2, ok := files["694eb7be78aedc7a163da900.json"]
	require.True(t, ok)
	fileContent2, err := recordFile2.Open()
	require.NoError(t, err)
	gotBytes2, err := io.ReadAll(fileContent2)
	require.NoError(t, err)
	fileContent2.Close()
	assert.Equal(t, `{
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
    }`, string(gotBytes2))
}

// Test_zipRecords_EmptyZip tests the zipRecords() function to verify that an empty zip is created if the scroll returns an empty array.
func Test_zipRecords_EmptyZip(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("TMPDIR", tempDir)
	t.Setenv("TEMP", tempDir)
	t.Setenv("TMP", tempDir)

	records := []gjson.Result{}

	dir, err := writeRecordsToFile(records, "my-bucket")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	savePath := filepath.Join(t.TempDir(), "my-bucket.zip")
	err = zipRecords(savePath, dir)

	zipFile, err := os.ReadFile(savePath)
	require.NoError(t, err)
	zipReader, err := zip.NewReader(bytes.NewReader(zipFile), int64(len(zipFile)))
	require.NoError(t, err)
	assert.Equal(t, 0, len(zipReader.File))
}

// Test_zipRecords_TempDirDoesNotExist tests the zipRecords() function when reading the temporary directory fails.
func Test_zipRecords_TempDirDoesNotExist(t *testing.T) {
	savePath := filepath.Join(t.TempDir(), "my-bucket.zip")
	err := zipRecords(savePath, "doesnotexist")

	require.Error(t, err)
	assert.EqualError(t, err, "the given path does not exist: doesnotexist")
}

// Test_zipRecords_FileDirDoesNotExist tests the zipRecords() function when creating the given file fails.
func Test_zipRecords_FileDirDoesNotExist(t *testing.T) {
	savePath := filepath.Join("doesnotexist", "my-bucket.zip")
	err := zipRecords(savePath, "directory")

	require.Error(t, err)
	assert.EqualError(t, err, "the given path does not exist: "+savePath)
}

// Test_zipRecords_DirContainsDirectory tests the zipRecords() function when the given temporary directory contains another directory.
func Test_zipRecords_DirContainsDirectory(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("TMPDIR", tempDir)
	t.Setenv("TEMP", tempDir)
	t.Setenv("TMP", tempDir)

	records := gjson.Parse(`[
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
]`).Array()

	dir, err := writeRecordsToFile(records, "my-bucket")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	os.Mkdir(filepath.Join(dir, "directory"), 0o755)

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Equal(t, 3, len(entries))

	savePath := filepath.Join(t.TempDir(), "my-bucket.zip")
	err = zipRecords(savePath, dir)

	zipFile, err := os.ReadFile(savePath)
	require.NoError(t, err)
	zipReader, err := zip.NewReader(bytes.NewReader(zipFile), int64(len(zipFile)))
	require.NoError(t, err)
	files := make(map[string]*zip.File, len(zipReader.File))
	require.Equal(t, 2, len(zipReader.File))
	for _, f := range zipReader.File {
		files[f.Name] = f
	}

	assert.NotContains(t, files, "directory")
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
		file           string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "DumpBucket correctly prints the received records with the ugly printer",
			client:         new(mocks.WorkingStagingContentController),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"acknowledged\":true}\n",
			file:           filepath.Join(t.TempDir(), "my-bucket.zip"),
			err:            nil,
		},
		{
			name:           "DumpBucket correctly prints the array with JSON pretty printer",
			client:         new(mocks.WorkingStagingContentController),
			printer:        nil,
			file:           filepath.Join(t.TempDir(), "my-bucket.zip"),
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		{
			name:           "DumpBucket correctly prints nothing with no content returned.",
			client:         new(mocks.WorkingStagingContentControllerNoContent),
			printer:        nil,
			file:           filepath.Join(t.TempDir(), "my-bucket.zip"),
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
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
			name:           "zipRecords fails",
			client:         new(mocks.WorkingStagingContentController),
			printer:        nil,
			expectedOutput: "",
			file:           filepath.Join("doesnotexist", "my-bucket.zip"),
			err:            NewErrorWithCause(ErrorExitCode, errors.New("the given path does not exist: "+filepath.Join("doesnotexist", "my-bucket.zip")), "Could not write dump to file."),
		},
		{
			name:      "Printing fails",
			client:    new(mocks.WorkingStagingContentController),
			printer:   nil,
			file:      filepath.Join(t.TempDir(), "my-bucket.zip"),
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
			max := 3
			err := d.DumpBucket(tc.client, "my-bucket", tc.file, gjson.Parse(filters), gjson.Parse(projections), &max, tc.printer)

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

// Test_discovery_DumpBucket_writeRecordsToFileFails tests the discovery.DumpBucket() function when it fails because the temporary directory could not be created.
func Test_discovery_DumpBucket_writeRecordsToFileFails(t *testing.T) {
	t.Setenv("TMPDIR", "/does/not/exist")
	t.Setenv("TEMP", "/does/not/exist")
	t.Setenv("TMP", "/does/not/exist")

	buf := &bytes.Buffer{}
	var out io.Writer
	out = buf

	ios := iostreams.IOStreams{
		In:  os.Stdin,
		Out: out,
		Err: os.Stderr,
	}

	d := NewDiscovery(&ios, viper.New(), "")
	max := 3
	err := d.DumpBucket(new(mocks.WorkingStagingContentController), "my-bucket", "", gjson.Result{}, gjson.Result{}, &max, nil)
	assert.Contains(t, err.Error(), "the given path does not exist:")
	assert.Contains(t, err.Error(), strings.ToLower(filepath.FromSlash("/does/not/exist")))
	assert.Contains(t, err.Error(), "Could not write records to temporary folder.")
}

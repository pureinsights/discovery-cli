package mocks

import (
	"errors"
	"net/http"

	"github.com/tidwall/gjson"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
)

// WorkingStagingBucketControllerNoConflict simulates when the StagingBucketController works.
type WorkingStagingBucketControllerNoConflict struct{}

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
func (s *WorkingStagingBucketControllerNoConflict) CreateIndex(string, string, []gjson.Result) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// DeleteIndex implements the interface.
func (s *WorkingStagingBucketControllerNoConflict) DeleteIndex(string, string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// WorkingStagingBucketControllerNameConflict simulates when the bucket already exists, but the updates succeed.
type WorkingStagingBucketControllerNameConflict struct {
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
func (s *WorkingStagingBucketControllerNameConflict) CreateIndex(string, string, []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// DeleteIndex simulates a working Index deletion.
func (s *WorkingStagingBucketControllerNameConflict) DeleteIndex(string, string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// FailingStagingBucketControllerNotDiscoveryError mocks when the create request does not return a Discovery error.
type FailingStagingBucketControllerNotDiscoveryError struct{}

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
func (s *FailingStagingBucketControllerNotDiscoveryError) CreateIndex(string, string, []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// DeleteIndex implements the interface.
func (s *FailingStagingBucketControllerNotDiscoveryError) DeleteIndex(string, string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// FailingStagingBucketControllerNotFoundError mocks when the function receives a Discovery error that is not a conflict.
type FailingStagingBucketControllerNotFoundError struct{}

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
func (s *FailingStagingBucketControllerNotFoundError) CreateIndex(string, string, []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// DeleteIndex implements the interface.
func (s *FailingStagingBucketControllerNotFoundError) DeleteIndex(string, string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// FailingStagingBucketControllerIndexCreationFails mocks a failing index creation.
type FailingStagingBucketControllerIndexCreationFails struct{}

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
func (s *FailingStagingBucketControllerIndexCreationFails) CreateIndex(string, string, []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

// DeleteIndex implements the interface.
func (s *FailingStagingBucketControllerIndexCreationFails) DeleteIndex(string, string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// FailingStagingBucketControllerIndexDeletionFails simulates when deleting an index fails.
type FailingStagingBucketControllerIndexDeletionFails struct{}

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
func (s *FailingStagingBucketControllerIndexDeletionFails) CreateIndex(string, string, []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// DeleteIndex returns an error.
func (s *FailingStagingBucketControllerIndexDeletionFails) DeleteIndex(string, string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": false
}`), discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "acknowledged": false
}`)}
}

// FailingStagingBucketControllerLastGetFails simulates when the last get of the bucket fails.
type FailingStagingBucketControllerLastGetFails struct{}

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
func (s *FailingStagingBucketControllerLastGetFails) CreateIndex(string, string, []gjson.Result) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// DeleteIndex implements the interface.
func (s *FailingStagingBucketControllerLastGetFails) DeleteIndex(string, string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// FailingStagingBucketControllerFirstGetFails mocks when the first get fails.
type FailingStagingBucketControllerFirstGetFails struct{}

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
func (s *FailingStagingBucketControllerFirstGetFails) CreateIndex(string, string, []gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// DeleteIndex implements the interface.
func (s *FailingStagingBucketControllerFirstGetFails) DeleteIndex(string, string) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// WorkingStagingContentController mocks a working content controller.
type WorkingStagingContentController struct{}

// Scroll implements the interface.
func (s *WorkingStagingContentController) Scroll(gjson.Result, gjson.Result, *int) ([]gjson.Result, error) {
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
type WorkingStagingContentControllerNoContent struct{}

// Scroll returns an empty array.
func (s *WorkingStagingContentControllerNoContent) Scroll(gjson.Result, gjson.Result, *int) ([]gjson.Result, error) {
	return []gjson.Result{}, nil
}

// FailingStagingContentController mocks a failing content controller.
type FailingStagingContentController struct{}

// Scroll returns an error.
func (s *FailingStagingContentController) Scroll(gjson.Result, gjson.Result, *int) ([]gjson.Result, error) {
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

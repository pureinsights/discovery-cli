package mocks

import (
	"net/http"

	"github.com/tidwall/gjson"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery/v2"
)

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

// Count returns a working result.
func (s *WorkingStagingContentController) Count(gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{"total": 2}`), nil
}

// WorkingStagingContentControllerNoContent mocks when the scroll returns no content.
type WorkingStagingContentControllerNoContent struct{}

// Scroll returns an empty array.
func (s *WorkingStagingContentControllerNoContent) Scroll(gjson.Result, gjson.Result, *int) ([]gjson.Result, error) {
	return []gjson.Result{}, nil
}

// Count returns a working result with no records.
func (s *WorkingStagingContentControllerNoContent) Count(gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{"total": 0}`), nil
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

// Count returns an error.
func (s *FailingStagingContentController) Count(gjson.Result) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
  "status": 404,
  "code": 1003,
  "messages": [
    "Entity not found: entity with name 'my-bucket' does not exist"
  ],
  "timestamp": "2025-12-23T14:53:32.321524600Z"
}`),
	}
}

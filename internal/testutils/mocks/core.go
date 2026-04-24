package mocks

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
)
// WorkingFileClient simulates a working file client.
type WorkingFileClient struct{}

func (w *WorkingFileClient) List() ([]gjson.Result, error) {
	return gjson.Parse(`[
		"Credential.ndjson",
		"Server.ndjson",
		"buildContextPrompt.js",
		"buildSimplePrompt.js",
		"constructPrompt.js",
		"constructSuggestedPrompt.js",
		"elastic-extraction.py",
		"extractReference.groovy",
		"extractReferenceAtlas.groovy",
		"formatAnalysisResponse.js",
		"formatAutocompleteResponse.js",
		"formatChunksResponse.js",
		"formatKeywordResponse.js",
		"formatKeywordResponseAtlas.js",
		"formatKeywordSearch.js",
		"formatQuestionsResponse.js",
		"formatSearchResponse.js",
		"formatSearchResponseAtlas.js",
		"formatSemanticResponse.js",
		"formatSuggestionsResponse.js",
		"keywordSearchTemplateAtlas.json",
		"searchTemplate.json",
		"searchTemplateAtlas.json"
	]`).Array(), nil
}

func (w *WorkingFileClient) Upload(key, file string) (gjson.Result, error) {
	return gjson.Parse(`{"acknowledged": true}`),nil
}

func (w *WorkingFileClient) Retrieve(key string) ([]byte, error) {
	return []byte(`
	def main():
		print("Hello, World!")

	if __name__ == "__main__":
		main()

	`),nil
}

func (w *WorkingFileClient) Delete(key string) (gjson.Result, error) {
	return gjson.Result{},nil
}

// WorkingFileClient simulates a working file client.
type FailingFileClient struct{}

func (w *FailingFileClient) List() ([]gjson.Result, error) {
	return []gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusInternalServerError,
		Body:   gjson.Parse(`{
	"status": 500,
	"code": 1003,
	"messages": [
		"Internal server error"
	],
	"timestamp": "2025-10-16T17:46:45.386963700Z"
}`),
	}
}

func (w *FailingFileClient) Upload(key, file string) (gjson.Result, error) {
	return gjson.Result{},discoveryPackage.Error{
		Status: http.StatusBadRequest,
		Body:   gjson.Parse(`{
	"status": 400,
	"code": 3002,
	"messages": [
		"key: Invalid format for file path, use only alphanumeric symbols with a limit of 255 characters and a max of 10 path levels."
	],
	"timestamp": "2025-10-16T17:46:45.386963700Z"
}`),
	}
}

func (w *FailingFileClient) Retrieve(key string) ([]byte, error) {
	return nil,discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body:   gjson.Result{},
	}
}

func (w *FailingFileClient) Delete(key string) (gjson.Result, error) {
	return gjson.Result{},nil
}


// WorkingServerPinger simulates when a ping to a server worked.
type WorkingServerPinger struct{}

// Ping returns the response of a working ping.
func (s *WorkingServerPinger) Ping(uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// SearchByName returns a server result.
func (s *WorkingServerPinger) SearchByName(string) (gjson.Result, error) {
	return gjson.Parse(`{
  "type": "mongo",
  "name": "MongoDB Atlas server",
  "labels": [],
  "active": true,
  "id": "21029da3-041c-43b5-a67e-870251f2f6a6",
  "creationTimestamp": "2025-11-20T00:06:05Z",
  "lastUpdatedTimestamp": "2025-11-20T00:06:05Z",
  "config": {
    "servers": [
      "mongodb+srv://cluster0.dleud.mongodb.net/"
    ],
    "connection": {
      "readTimeout": "30s",
      "connectTimeout": "1m"
    },
    "credentialId": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c"
  }
}`), nil
}

// Search implements the searcher interface.
func (s *WorkingServerPinger) Search(gjson.Result) ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body:   gjson.Result{},
	}
}

// Get implements the Searcher interface.
func (s *WorkingServerPinger) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: 21029da3-041c-43b5-a67e-870251f2f6a6"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
	}
}

// GetAll implements the searcher interface.
func (s *WorkingServerPinger) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// FailingServerPinger simulates when a server that does not exist was pinged.
type FailingServerPingerServerNotFound struct{}

// Ping returns the response of a failing ping.
func (s *FailingServerPingerServerNotFound) Ping(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusUnprocessableEntity, Body: gjson.Parse(`{
	"status": 422,
	"code": 4001,
	"messages": [
		"Client of type openai cannot be validated."
	],
	"timestamp": "2025-10-23T22:35:38.345647200Z"
	}`)}
}

// SearchByName returns a not found error.
func (s *FailingServerPingerServerNotFound) SearchByName(string) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: 21029da3-041c-43b5-a67e-870251f2f6a6"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
	}
}

// Search implements the searcher interface.
func (s *FailingServerPingerServerNotFound) Search(gjson.Result) ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body:   gjson.Result{},
	}
}

// Get implements the Searcher interface.
func (s *FailingServerPingerServerNotFound) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: 21029da3-041c-43b5-a67e-870251f2f6a6"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
	}
}

// GetAll implements the searcher interface.
func (s *FailingServerPingerServerNotFound) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// FailingServerPinger simulates when a ping to a server fails.
type FailingServerPingerPingFailed struct{}

// Ping returns the response of a failing ping.
func (s *FailingServerPingerPingFailed) Ping(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusUnprocessableEntity, Body: gjson.Parse(`{
	"status": 422,
	"code": 4001,
	"messages": [
		"Client of type openai cannot be validated."
	],
	"timestamp": "2025-10-23T22:35:38.345647200Z"
	}`)}
}

// SearchByName returns a result of a server.
func (s *FailingServerPingerPingFailed) SearchByName(string) (gjson.Result, error) {
	return gjson.Parse(`{
  "type": "mongo",
  "name": "MongoDB Atlas server",
  "labels": [],
  "active": true,
  "id": "21029da3-041c-43b5-a67e-870251f2f6a6",
  "creationTimestamp": "2025-11-20T00:06:05Z",
  "lastUpdatedTimestamp": "2025-11-20T00:06:05Z",
  "config": {
    "servers": [
      "mongodb+srv://cluster0.dleud.mongodb.net/"
    ],
    "connection": {
      "readTimeout": "30s",
      "connectTimeout": "1m"
    },
    "credentialId": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c"
  }
}`), nil
}

// Search implements the searcher interface.
func (s *FailingServerPingerPingFailed) Search(gjson.Result) ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body:   gjson.Result{},
	}
}

// Get implements the Searcher interface.
func (s *FailingServerPingerPingFailed) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: 21029da3-041c-43b5-a67e-870251f2f6a6"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
	}
}

// GetAll implements the searcher interface.
func (s *FailingServerPingerPingFailed) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

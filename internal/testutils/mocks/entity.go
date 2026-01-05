package mocks

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/tidwall/gjson"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
)

// WorkingGetter mocks the discovery.Getter struct to always answer a working result.
type WorkingGetter struct {
	mock.Mock
}

// Get returns a working processor as if the request worked successfully.
func (g *WorkingGetter) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB text processor",
		"labels": [],
		"active": true,
		"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
		"creationTimestamp": "2025-08-14T18:02:38Z",
		"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
	}`), nil
}

// GetAll returns a list of processors.
func (g *WorkingGetter) GetAll() ([]gjson.Result, error) {
	return gjson.Parse(`[
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
	]`).Array(), nil
}

// FailingGetter mocks the discovery.Getter struct to always return an HTTP error.
type FailingGetter struct {
	mock.Mock
}

// Get returns a 404 Not Found.
func (g *FailingGetter) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Secret not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
	}
}

// GetAll returns 401 unauthorized.
func (g *FailingGetter) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// WorkingSearcher mocks the discovery.Searcher struct.
// This struct is used to test when the search functions work.
type WorkingSearcher struct {
	mock.Mock
}

// Search returns an array of results as it if correctly found matches.
func (s *WorkingSearcher) Search(gjson.Result) ([]gjson.Result, error) {
	return gjson.Parse(`[
                  {
					"source": {
						"type": "mongo",
						"name": "MongoDB Atlas server clone",
						"labels": [],
						"active": true,
						"id": "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
						"creationTimestamp": "2025-09-29T15:50:17Z",
						"lastUpdatedTimestamp": "2025-09-29T15:50:17Z"
					},
					"highlight": {},
					"score": 0.20970252
                  },
                  {
					"source": {
						"type": "mongo",
						"name": "MongoDB Atlas server clone 1",
						"labels": [],
						"active": true,
						"id": "8f14c11c-bb66-49d3-aa2a-dedff4608c17",
						"creationTimestamp": "2025-09-29T15:50:19Z",
						"lastUpdatedTimestamp": "2025-09-29T15:50:19Z"
					},
					"highlight": {},
					"score": 0.20970252
                  },
                  {
					"source": {
						"type": "mongo",
						"name": "MongoDB Atlas server clone 3",
						"labels": [],
						"active": true,
						"id": "3a0214a4-72cc-4eee-ad0c-9e3af9b08a6c",
						"creationTimestamp": "2025-09-29T15:50:20Z",
						"lastUpdatedTimestamp": "2025-09-29T15:50:20Z"
					},
					"highlight": {},
					"score": 0.20970252
                  }
          ]`).Array(), nil
}

// SearchByName returns an object as if it found correctly the entity.
func (s *WorkingSearcher) SearchByName(string) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB Atlas server",
		"labels": [],
		"active": true,
		"id": "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
		"creationTimestamp": "2025-09-29T15:50:17Z",
		"lastUpdatedTimestamp": "2025-09-29T15:50:17Z",
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

// Get returns a JSON object as if the searcher found the entity by its ID.
func (s *WorkingSearcher) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB Atlas server",
		"labels": [],
		"active": true,
		"id": "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
		"creationTimestamp": "2025-09-29T15:50:17Z",
		"lastUpdatedTimestamp": "2025-09-29T15:50:17Z",
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

// GetAll returns the result of a search.
func (s *WorkingSearcher) GetAll() ([]gjson.Result, error) {
	return gjson.Parse(`[
		{
		"type": "mongo",
		"name": "my-credential",
		"labels": [
			{
			"key": "A",
			"value": "A"
			}
		],
		"active": true,
		"id": "3b32e410-2f33-412d-9fb8-17970131921c",
		"creationTimestamp": "2025-10-17T22:37:57Z",
		"lastUpdatedTimestamp": "2025-10-17T22:37:57Z"
		},
		{
		"type": "openai",
		"name": "OpenAI credential clone clone",
		"labels": [],
		"active": true,
		"id": "5c09589e-b643-41aa-a766-3b7fc3660473",
		"creationTimestamp": "2025-10-17T22:38:12Z",
		"lastUpdatedTimestamp": "2025-10-17T22:38:12Z"
		},
	]`).Array(), nil
}

// FailingSearcher mocks the discovery.Searcher struct when its functions return errors.
type FailingSearcher struct {
	mock.Mock
}

// Search implements the searcher interface.
func (s *FailingSearcher) Search(gjson.Result) ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body:   gjson.Result{},
	}
}

// SearchByName returns 404 so that the searchEntity function enters the err != nil code branch.
func (s *FailingSearcher) SearchByName(string) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusBadRequest,
		Body: gjson.Parse(`{
	"status": 400,
	"code": 3002,
	"messages": [
		"Invalid JSON: Unexpected end-of-input:"
	],
	"timestamp": "2025-10-17T17:43:52.817308100Z"
	}`),
	}
}

// Get returns an error so mock when the user searches for an entity that does not exist.
func (s *FailingSearcher) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: 986ce864-af76-4fcb-8b4f-f4e4c6ab0951"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
	}
}

// GetAll implements the searcher interface.
func (s *FailingSearcher) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// FailingSearcherWorkingGetter mocks the discovery.searcher struct when the search by name fails, but the get does succeed.
type FailingSearcherWorkingGetter struct {
	mock.Mock
}

// Search implements the searcher interface.
func (s *FailingSearcherWorkingGetter) Search(gjson.Result) ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body:   gjson.Result{},
	}
}

// SearchByName returns 404 not found to make the test go through the err != nil code branch.
func (s *FailingSearcherWorkingGetter) SearchByName(name string) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(fmt.Sprintf(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name %q does not exist"
	]
}`, name)),
	}
}

// Get returns a JSON object to mock that the SearchByName failed, but the Get succeeded.
func (s *FailingSearcherWorkingGetter) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB Atlas server clone",
		"labels": [],
		"active": true,
		"id": "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
		"creationTimestamp": "2025-09-29T15:50:17Z",
		"lastUpdatedTimestamp": "2025-09-29T15:50:17Z",
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

// GetAll implements the searcher interface.
func (s *FailingSearcherWorkingGetter) GetAll() ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// FailingSearcherFailingGetter mocks the discovery.searcher struct when both the searchByName and Get function fails.
type FailingSearcherFailingGetter struct {
	mock.Mock
}

// Search returns an error.
func (s *FailingSearcherFailingGetter) Search(gjson.Result) ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// SearchByName returns a 404 Not Found error to make the test go through the err != nil branch.
func (s *FailingSearcherFailingGetter) SearchByName(name string) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(fmt.Sprintf(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name %q does not exist"
	]
}`, name)),
	}
}

// Get Returns error to mock that it also failed to find the entity.
func (s *FailingSearcherFailingGetter) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: 986ce864-af76-4fcb-8b4f-f4e4c6ab0951"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
	}
}

// GetAll implements the searcher interface.
func (s *FailingSearcherFailingGetter) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// SearcherReturnsOtherError is a struct that mocks discovery.searcher when the search functions do not return a discovery.Error.
type SearcherReturnsOtherError struct {
	mock.Mock
}

// Search implements the searcher interface.
func (s *SearcherReturnsOtherError) Search(gjson.Result) ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body:   gjson.Parse(``),
	}
}

// SearchByName does not return a discovery.Error.
func (s *SearcherReturnsOtherError) SearchByName(string) (gjson.Result, error) {
	return gjson.Result{}, errors.New("not discovery error")
}

// Get implements the searcher interface.
func (s *SearcherReturnsOtherError) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body:   gjson.Parse(``),
	}
}

// GetAll implements the searcher interface.
func (s *SearcherReturnsOtherError) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(``)}
}

// SearcherIDNotUUID simulates when the searcher returns a result with an ID that is not a UUID.
type SearcherIDNotUUID struct {
	mock.Mock
}

// Search implements the searcher interface.
func (s *SearcherIDNotUUID) Search(gjson.Result) ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body:   gjson.Result{},
	}
}

// SearchByName returns a result with an ID that is not a UUID so that the conversion can fail.
func (s *SearcherIDNotUUID) SearchByName(string) (gjson.Result, error) {
	return gjson.Parse(`{
			"type": "mongo",
			"name": "MongoDB Atlas seed clone",
			"labels": [],
			"active": true,
			"id": "test",
			"creationTimestamp": "2025-09-29T15:50:17Z",
			"lastUpdatedTimestamp": "2025-09-29T15:50:17Z"
		}`), nil
}

// Get implements the Searcher interface.
func (s *SearcherIDNotUUID) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Seed not found: 986ce864-af76-4fcb-8b4f-f4e4c6ab0951"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
	}
}

// GetAll implements the searcher interface.
func (s *SearcherIDNotUUID) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// WorkingCreator mocks when creating and updating entities works.
type WorkingCreator struct {
	mock.Mock
}

// Create returns a JSON as if it worked successfully.
func (g *WorkingCreator) Create(gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB credential",
		"labels": [],
		"active": true,
		"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	}`), nil
}

// Update returns a JSON as if it worked successfully.
func (g *WorkingCreator) Update(uuid.UUID, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB credential",
		"labels": [],
		"active": true,
		"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	}`), nil
}

// FailingCreator mocks when creating and updating entities fails.
type FailingCreator struct {
	mock.Mock
}

// Create returns a JSON as if it worked successfully.
func (g *FailingCreator) Create(gjson.Result) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
  "status": 400,
  "code": 3002,
  "messages": [
    "Invalid JSON: Illegal unquoted character ((CTRL-CHAR, code 10)): has to be escaped using backslash to be included in name\n at [Source: REDACTED (StreamReadFeature.INCLUDE_SOURCE_IN_LOCATION disabled); line: 5, column: 17]"
  ],
  "timestamp": "2025-10-29T14:46:48.055840300Z"
}`)}
}

// Update returns a JSON as if it worked successfully.
func (g *FailingCreator) Update(uuid.UUID, gjson.Result) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
  "status": 404,
  "code": 1003,
  "messages": [
    "Entity not found: 9ababe08-0b74-4672-bb7c-e7a8227d6d4d"
  ],
  "timestamp": "2025-10-29T14:47:36.290329Z"
}`)}
}

// FailingCreator mocks when creating and updating entities fails.
type FailingCreatorCreateWorksUpdateFails struct {
	mock.Mock
}

// Create returns a JSON as if it worked successfully.
func (g *FailingCreatorCreateWorksUpdateFails) Create(gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB credential",
		"labels": [],
		"active": true,
		"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	}`), nil
}

// Update returns an error that is not a Discovery.Error
func (g *FailingCreatorCreateWorksUpdateFails) Update(uuid.UUID, gjson.Result) (gjson.Result, error) {
	return gjson.Result{}, errors.New(`invalid UUID length: 4`)
}

// WorkingDeleter mocks when the deleter interface works correctly.
type WorkingDeleter struct {
	mock.Mock
}

// Get returns a working processor as if the request worked successfully.
func (g *WorkingDeleter) Delete(uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
		"acknowledged": true
	}`), nil
}

// FailingDeleter mocks the deleter interface when the Delete() method fails.
type FailingDeleter struct {
	mock.Mock
}

// Get returns a working processor as if the request worked successfully.
func (g *FailingDeleter) Delete(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusBadRequest,
		Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [id] for value [test] due to: Invalid UUID string: test"
			],
			"timestamp": "2025-10-23T22:35:38.345647200Z"
			}`),
	}
}

// WorkingSearchDeleter correctly finds the entity by its name and deletes it with its ID.
type WorkingSearchDeleter struct {
	mock.Mock
}

// Get returns a working processor as if the request worked successfully.
func (g *WorkingSearchDeleter) Delete(uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
		"acknowledged": true
	}`), nil
}

// SearchByName returns a valid JSON.
func (g *WorkingSearchDeleter) SearchByName(string) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB text processor",
		"labels": [],
		"active": true,
		"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
		"creationTimestamp": "2025-08-14T18:02:38Z",
		"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
	}`), nil
}

// Search implements the searchDeleter interface.
func (g *WorkingSearchDeleter) Search(gjson.Result) ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// Get implements the searchDeleter interface.
func (g *WorkingSearchDeleter) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB text processor",
		"labels": [],
		"active": true,
		"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
		"creationTimestamp": "2025-08-14T18:02:38Z",
		"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
	}`), nil
}

// GetAll implements the searchDeleter interface.
func (g *WorkingSearchDeleter) GetAll() ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// FailingSearchDeleterSearchFails fails in the SearchByName() function.
type FailingSearchDeleterSearchFails struct {
	mock.Mock
}

// SearchByName returns a not found error, so the entity does not exist.
func (g *FailingSearchDeleterSearchFails) SearchByName(name string) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(fmt.Sprintf(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name %q does not exist"
	],
	"timestamp": "2025-09-30T15:38:42.885125200Z"
}`, name)),
	}
}

// Search implements the searchDeleter interface.
func (g *FailingSearchDeleterSearchFails) Search(gjson.Result) ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// Get implements the searchDeleter interface.
func (g *FailingSearchDeleterSearchFails) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// GetAll implements the searchDeleter interface.
func (g *FailingSearchDeleterSearchFails) GetAll() ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// Get returns a working processor as if the request worked successfully.
func (g *FailingSearchDeleterSearchFails) Delete(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusBadRequest,
		Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [id] for value [test] due to: Invalid UUID string: test"
			],
			"timestamp": "2025-10-23T22:35:38.345647200Z"
			}`),
	}
}

// FailingSearchDeleterSearchFails fails in the DeleteEntity() function.
type FailingSearchDeleterDeleteFails struct {
	mock.Mock
}

// SearchByName returns a valid JSON.
func (g *FailingSearchDeleterDeleteFails) SearchByName(string) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB text processor",
		"labels": [],
		"active": true,
		"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
		"creationTimestamp": "2025-08-14T18:02:38Z",
		"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
	}`), nil
}

// Search implements the searchDeleter interface.
func (g *FailingSearchDeleterDeleteFails) Search(gjson.Result) ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// Get implements the searchDeleter interface.
func (g *FailingSearchDeleterDeleteFails) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB text processor",
		"labels": [],
		"active": true,
		"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
		"creationTimestamp": "2025-08-14T18:02:38Z",
		"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
	}`), nil
}

// GetAll implements the searchDeleter interface.
func (g *FailingSearchDeleterDeleteFails) GetAll() ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// Delete fails due to bad request.
func (g *FailingSearchDeleterDeleteFails) Delete(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusBadRequest,
		Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [id] for value [test] due to: Invalid UUID string: test"
			],
			"timestamp": "2025-10-23T22:35:38.345647200Z"
			}`),
	}
}

// FailingSearchDeleterParsingUUIDFails fails when trying to parse the id in the received search result.
type FailingSearchDeleterParsingUUIDFails struct {
	mock.Mock
}

// SearchByName returns a valid JSON.
func (g *FailingSearchDeleterParsingUUIDFails) SearchByName(string) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB text processor",
		"labels": [],
		"active": true,
		"id": "notuuid",
		"creationTimestamp": "2025-08-14T18:02:38Z",
		"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
	}`), nil
}

// Search implements the searchDeleter interface.
func (g *FailingSearchDeleterParsingUUIDFails) Search(gjson.Result) ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// Get implements the searchDeleter interface.
func (g *FailingSearchDeleterParsingUUIDFails) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB text processor",
		"labels": [],
		"active": true,
		"id": "notuuid",
		"creationTimestamp": "2025-08-14T18:02:38Z",
		"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
	}`), nil
}

// GetAll implements the searchDeleter interface.
func (g *FailingSearchDeleterParsingUUIDFails) GetAll() ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// Delete implements the searchDeleter interface.
func (g *FailingSearchDeleterParsingUUIDFails) Delete(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, nil
}

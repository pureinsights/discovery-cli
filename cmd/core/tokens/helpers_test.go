package tokens

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"strings"
	"testing"

	"github.com/pureinsights/discovery-cli/v2/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// Variables
var accountId = "8f840aef-4fb2-4ea3-8aac-d743aaf13a90"

const tokenId = "b4bcd270-e824-4112-913b-65fb89e151d3"

const apiKey = "apiKey123"

var serviceAccountPath = "/v2/account/service/"

var tokenPath = serviceAccountPath + accountId + "/token"

// Bodies (search + entity)
var serviceAccountBody = `{
	"name": "ci",
	"description": "Continuous Integration account",
	"labels": [],
	"active": true,
	"id": "` + accountId + `",
	"creationTimestamp": "2026-10-29T15:37:47Z",
	"lastUpdatedTimestamp": "2026-11-03T11:50:39Z",
	"role": "OPERATE"
}`

var serviceAccountSearchBody = `{
	"content": [
		{
			"source": {
				"name": "ci",
				"description": "Continuous Integration account",
				"labels": [],
				"active": true,
				"id": "` + accountId + `",
				"creationTimestamp": "2026-10-29T15:37:47Z",
				"lastUpdatedTimestamp": "2026-11-03T11:50:39Z"
			},
			"highlight": {},
			"score": 1
		}
	],
	"pageable": {"page": 0, "size": 25, "sort": []},
	"totalSize": 1,
	"totalPages": 1,
	"empty": false,
	"size": 25,
	"offset": 0,
	"numberOfElements": 1,
	"pageNumber": 0
}`

var tokenBody = `{
	"id": "` + tokenId + `",
	"creationTimestamp": "2026-11-18T16:48:42Z",
	"lastUpdatedTimestamp": "2026-11-18T16:48:42Z",
	"name": "My Token",
	"active": true,
	"account": "` + accountId + `",
	"expirationTimestamp": "2026-12-18T16:48:42Z",
	"scopes": ["core:*"]
}`

var tokenSearchBody = `{
	"content": [{"source": ` + tokenBody + `, "highlight": {}, "score": 1}],
	"pageable": {"page": 0, "size": 25, "sort": []},
	"totalSize": 1,
	"totalPages": 1,
	"empty": false,
	"size": 25,
	"offset": 0,
	"numberOfElements": 1,
	"pageNumber": 0
}`

var tokenGetAllBody = `{
    "content": [
        ` + tokenBody + `,
        {
            "id": "5c688285-3a3a-4f93-913e-d828db95a79c",
            "creationTimestamp": "2026-11-18T16:45:19Z",
            "lastUpdatedTimestamp": "2026-11-18T16:45:19Z",
            "name": "token_core_infinite",
            "active": true,
            "account": "8f840aef-4fb2-4ea3-8aac-d743aaf13a90",
            "scopes": [
                "core:*"
            ]
        }
    ],
    "pageable": {
        "page": 0,
        "size": 25,
        "sort": []
    },
    "totalSize": 2,
    "totalPages": 1,
    "empty": false,
    "size": 25,
    "offset": 0,
    "pageNumber": 0,
    "numberOfElements": 2
}`

var renewedTokenBody = `{
	"id": "` + tokenId + `",
	"creationTimestamp": "2026-11-18T16:48:42Z",
	"lastUpdatedTimestamp": "2026-11-19T10:00:00Z",
	"name": "My Token",
	"active": true,
	"account": "` + accountId + `",
	"expirationTimestamp": "2027-01-18T10:00:00Z",
	"scopes": ["core:*"],
	"token": "pdp_q3K9xV1mZ8bT4nR7wY2cL6hJ0sD5fG1aE9uI3oP8kM4"
}`

// createdTokenBody returns the token Core answers with when it is created.
func createdTokenBody(tokenId, name, scopes, expiration string, active bool) string {
	scopesJson, _ := json.Marshal(splitScopes(scopes))

	var response = `{
	"account": "` + accountId + `",
	"active": ` + fmt.Sprint(active) + `,
	"id": "` + tokenId + `",
	"expirationTimestamp": "` + expiration + `",
	"name": "` + name + `",
	"scopes": ` + string(scopesJson) + `,
	"token": "pdp_OysmlP790ufVjNQZabxZ2hp2cJsV1nL6b29lXo-ChaI"
}`
	if expiration == "" {
		response = strings.Replace(response, `"expirationTimestamp": "`+expiration+`",`, "", 1)
	}
	return response
}

// mock returns a JSON response that asserts the method, path and API key of the request
func mock(status int, body, method, path string) testutils.MockResponse {
	return testutils.MockResponse{
		StatusCode:  status,
		Body:        body,
		ContentType: "application/json",
		Assertions: func(t *testing.T, r *http.Request) {
			assert.Equal(t, method, r.Method)
			assert.Equal(t, path, r.URL.Path)
			assert.Equal(t, apiKey, r.Header.Get("X-API-Key"))
		},
	}
}

// serviceAccountFoundPhase returns the mocked responses for the service account search and get phases.
func serviceAccountFoundPhase() map[string]testutils.MockResponse {
	return map[string]testutils.MockResponse{
		"POST:" + serviceAccountPath + "search": mock(http.StatusOK, serviceAccountSearchBody, http.MethodPost, serviceAccountPath+"search"),
		"GET:" + serviceAccountPath + accountId: mock(http.StatusOK, serviceAccountBody, http.MethodGet, serviceAccountPath+accountId),
	}
}

// serviceAccountNotFoundPhase mocks searching for a service account name that doesn't exist
func serviceAccountNotFoundPhase() map[string]testutils.MockResponse {
	return map[string]testutils.MockResponse{
		"POST:" + serviceAccountPath + "search": mock(http.StatusNoContent, ``, http.MethodPost, serviceAccountPath+"search"),
	}
}

// findTokenByNamePhase mocks finding "My Token" by name
func findTokenByNamePhase() map[string]testutils.MockResponse {
	return map[string]testutils.MockResponse{
		"POST:" + tokenPath + "/search":    mock(http.StatusOK, tokenSearchBody, http.MethodPost, tokenPath+"/search"),
		"GET:" + tokenPath + "/" + tokenId: mock(http.StatusOK, tokenBody, http.MethodGet, tokenPath+"/"+tokenId),
	}
}

// findTokenByIdPhase mocks finding the token by its UUID
func findTokenByIdPhase() map[string]testutils.MockResponse {
	return map[string]testutils.MockResponse{
		"POST:" + tokenPath + "/search":    mock(http.StatusNoContent, ``, http.MethodPost, tokenPath+"/search"),
		"GET:" + tokenPath + "/" + tokenId: mock(http.StatusOK, tokenBody, http.MethodGet, tokenPath+"/"+tokenId),
	}
}

// tokenNotFoundPhase mocks searching for a token name that doesn't exist
func tokenNotFoundPhase() map[string]testutils.MockResponse {
	return map[string]testutils.MockResponse{
		"POST:" + tokenPath + "/search": mock(http.StatusNoContent, ``, http.MethodPost, tokenPath+"/search"),
	}
}

// mockResponses merges the mocked responses of each phase
func mockResponses(phases ...map[string]testutils.MockResponse) map[string]testutils.MockResponse {
	all := map[string]testutils.MockResponse{}
	for _, phase := range phases {
		maps.Copy(all, phase)
	}
	return all
}

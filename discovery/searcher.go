package discovery

import (
	"net/http"

	"github.com/tidwall/gjson"
)

// Searcher is a struct that adds the path to a request to search for entities with filters.
type searcher struct {
	client
}

// Search iterates through every page of the results and returns an array with only the JSON objects.
// Each object has a score that grades how well it matches the given filters.
func (s searcher) Search(filter gjson.Result) ([]gjson.Result, error) {
	return executeWithPagination(s.client, http.MethodPost, "/search", WithJSONBody(filter.Raw))
}

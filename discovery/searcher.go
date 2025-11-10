package discovery

import (
	"fmt"
	"net/http"

	"github.com/tidwall/gjson"
)

// Searcher is a struct that adds the path to a request to search for entities with filters.
type searcher struct {
	client
}

const (
	// NotFoundError contains the template for the error that is returned when the entity could not be found.
	NotFoundError string = `{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name %q does not exist"
	]
}`
)

// Search iterates through every page of the results and returns an array with only the JSON objects.
// Each object has a score that grades how well it matches the given filters.
func (s searcher) Search(filter gjson.Result) ([]gjson.Result, error) {
	results, err := executeWithPagination(s.client, http.MethodPost, "/search", WithJSONBody(filter.Raw))
	if err != nil {
		return []gjson.Result(nil), err
	}
	for index, entity := range results {
		results[index] = entity.Get("source")
	}
	return results, nil
}

// SearchByName creates the filter to search an entity by the given name and calls the searcher.Search() function.
// It returns the first result if any or an error if it was not found or the search failed.
func (s searcher) SearchByName(name string) (gjson.Result, error) {
	byNameFilter := gjson.Parse(fmt.Sprintf(`{
		"equals": {
			"field": "name",
			"value": "%s"
		}
	}`, name))

	results, err := s.Search(byNameFilter)
	if err != nil {
		return gjson.Result{}, err
	}

	if len(results) == 0 || results[0].Get("name").String() != name {
		return gjson.Result{}, Error{
			Status: http.StatusNotFound, Body: gjson.Parse(fmt.Sprintf(NotFoundError, name)),
		}
	}

	return execute(s.client, http.MethodGet, "/"+results[0].Get("id").String())
}

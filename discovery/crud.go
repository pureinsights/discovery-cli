package discovery

import (
	"net/http"
	"strconv"

	"github.com/tidwall/gjson"

	"github.com/google/uuid"
)

// Getter is a struct that does the request to get entities.
type getter struct {
	client
}

// Get executes a GET request to the client's endpoint to get a single entity.
// It returns the JSON object it receives or an error if the request failed.
func (getter getter) Get(id uuid.UUID) (gjson.Result, error) {
	response, err := execute(getter.client, http.MethodGet, "/"+id.String())
	if err != nil {
		return gjson.Result{}, err
	}

	return response, nil
}

// GetAll retrieves every entity. It iterates through every page to get all of the results.
// It returns an array of JSON objects or an error if the request failed.
func (getter getter) GetAll() ([]gjson.Result, error) {
	response, err := execute(getter.client, http.MethodGet, "")
	if err != nil {
		return []gjson.Result{}, err
	}

	elementNumber := response.Get("numberOfElements").Int()
	pageNumber := response.Get("pageNumber").Int()
	totalPages := response.Get("totalPages").Int()
	totalSize := response.Get("totalSize").Int()
	elements := response.Get("content").Array()
	pageNumber++
	for pageNumber < totalPages && elementNumber < totalSize {
		response, err = execute(getter.client, http.MethodGet, "", WithQueryParameters(map[string][]string{"page": {strconv.FormatInt(pageNumber, 10)}}))
		if err != nil {
			return []gjson.Result{}, err
		}

		pageElements := response.Get("content").Array()
		elements = append(elements, pageElements...)

		pageNumber++
		pageElementNumber := response.Get("numberOfElements").Int()
		elementNumber += pageElementNumber
	}
	return elements, nil
}

// Crud is a struct that has creates, reads, updates, and deletes entities.
type crud struct {
	getter
}

// Create creates an entity.
// It returns the body of the entity if it was created or an error if the request failed.
func (crud crud) Create(config gjson.Result) (gjson.Result, error) {
	response, err := execute(crud.client, http.MethodPost, "", WithJSONBody(config.Raw))
	if err != nil {
		return gjson.Result{}, err
	}

	return response, nil
}

// Update updates an entity.
// It returns the body of the entity if it was updated or an error if the request failed.
func (crud crud) Update(id uuid.UUID, config gjson.Result) (gjson.Result, error) {
	response, err := execute(crud.client, http.MethodPut, "/"+id.String(), WithJSONBody(config.Raw))
	if err != nil {
		return gjson.Result{}, err
	}

	return response, nil
}

// Delete deletes an entity.
// It returns the the acknowledged message if it was deleted or an error if the request failed.
func (crud crud) Delete(id uuid.UUID) (gjson.Result, error) {
	response, err := execute(crud.client, http.MethodDelete, "/"+id.String())
	if err != nil {
		return gjson.Result{}, err
	}

	return response, nil
}

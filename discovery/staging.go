package discovery

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

// stagingGetContentOption is a type definition used for the functional options pattern.
// It adds query parameters to the contentClient.Get().
type stagingGetContentOption func(*map[string][]string)

// WithContentAction adds the given action as query parameter to the Get function.
func WithContentAction(action string) stagingGetContentOption {
	return func(m *map[string][]string) {
		(*m)["action"] = append((*m)["action"], action)
	}
}

// WithIncludeProjections adds the query parameters to set the given fields as the ones the results will include.
func WithIncludeProjections(include []string) stagingGetContentOption {
	return func(m *map[string][]string) {
		(*m)["include"] = append((*m)["include"], include...)
	}
}

// WithExcludeProjections adds the query parameters to set the given fields as the ones the results will exclude.
func WithExcludeProjections(exclude []string) stagingGetContentOption {
	return func(m *map[string][]string) {
		(*m)["exclude"] = append((*m)["exclude"], exclude...)
	}
}

// contentClient is the struct that manages the content inside the Staging Repository's buckets.
type contentClient struct {
	client
}

// newContentClient is the constructor of the contentClient struct.
func newContentClient(url, apiKey, bucketName string) contentClient {
	return contentClient{
		client: newClient(url+"/content/"+bucketName, apiKey),
	}
}

// Store adds the given JSON content with the contentId parameter. The parentId parameter can be used to set hierarchical relationships between documents.
func (c contentClient) Store(contentId, parentId string, content gjson.Result) (gjson.Result, error) {
	if parentId == "" {
		return execute(c.client, http.MethodPost, "/"+contentId, WithJSONBody(content.Raw))
	} else {
		return execute(c.client, http.MethodPost, "/"+contentId, WithQueryParameters(map[string][]string{
			"parentId": {parentId},
		}), WithJSONBody(content.Raw))
	}
}

// Get obtains the information of the record in the bucket with the given contentId.
// It can receive functional options to add the action, include, and exclude query parameters.
func (c contentClient) Get(contentId string, options ...stagingGetContentOption) (gjson.Result, error) {
	queryParams := make(map[string][]string)
	for _, opt := range options {
		opt(&queryParams)
	}
	return execute(c.client, http.MethodGet, "/"+contentId, WithQueryParameters(queryParams))
}

// Delete deletes the document with the given contentId in the bucket.
func (c contentClient) Delete(contentId string) (gjson.Result, error) {
	return execute(c.client, http.MethodDelete, "/"+contentId)
}

// DeleteMany deletes the documents that match the given parentId or filters.
func (c contentClient) DeleteMany(parentId string, filter gjson.Result) (gjson.Result, error) {
	options := []RequestOption{}
	if parentId != "" {
		options = append(options, WithQueryParameters(map[string][]string{
			"parentId": {parentId},
		}))
	}

	if filter.Exists() {
		options = append(options, WithJSONBody(filter.Raw))
	}

	return execute(c.client, http.MethodDelete, "", options...)
}

// bucketsClient is the struct that manages buckets in the Staging Repository.
type bucketsClient struct {
	client
}

// newBuckets is the constructor of the bucketsClient struct.
func newBucketsClient(url, apiKey string) bucketsClient {
	return bucketsClient{
		client: newClient(url+"/bucket", apiKey),
	}
}

// Create adds a new bucket with the given name and options, which can be used to create indices and set configurations.
func (b bucketsClient) Create(bucket string, options gjson.Result) (gjson.Result, error) {
	return execute(b.client, http.MethodPost, "/"+bucket, WithJSONBody(options.Raw))
}

// GetAll obtains a list with the names of every bucket.
func (b bucketsClient) GetAll() ([]string, error) {
	bucketsBytes, err := b.execute(http.MethodGet, "")
	if err != nil {
		return []string(nil), err
	}
	if len(bucketsBytes) > 0 {
		var bucketNames []string
		if err := json.Unmarshal(bucketsBytes, &bucketNames); err != nil {
			return []string(nil), err
		}
		return bucketNames, nil
	} else {
		return []string{}, nil
	}
}

// Get obtains the information of a bucket with the given name.
func (b bucketsClient) Get(bucket string) (gjson.Result, error) {
	return execute(b.client, http.MethodGet, "/"+bucket)
}

// Delete deletes the bucket with the given name.
func (b bucketsClient) Delete(bucket string) (gjson.Result, error) {
	return execute(b.client, http.MethodDelete, "/"+bucket)
}

// Purge deletes all of the records in the given bucket.
func (b bucketsClient) Purge(bucket string) (gjson.Result, error) {
	return execute(b.client, http.MethodDelete, "/"+bucket+"/purge")
}

// CreateIndex adds an index with the given name and configuration to a bucket.
func (b bucketsClient) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	var parts []string
	for _, r := range config {
		parts = append(parts, r.Raw)
	}
	jsonArray := "[" + strings.Join(parts, ",") + "]"

	return execute(b.client, http.MethodPut, "/"+bucket+"/index/"+index, WithJSONBody(jsonArray))
}

// DeleteIndex removes the index of a bucket.
func (b bucketsClient) DeleteIndex(bucket, index string) (gjson.Result, error) {
	return execute(b.client, http.MethodDelete, "/"+bucket+"/index/"+index)
}

// staging is the struct for the client that can carry out every Staging operation.
type staging struct {
	Url, ApiKey string
}

// Buckets creates a new bucketsClient.
func (s staging) Buckets() bucketsClient {
	return newBucketsClient(s.Url, s.ApiKey)
}

// Content creates a new contentClient.
func (s staging) Content(bucket string) contentClient {
	return newContentClient(s.Url, s.ApiKey, bucket)
}

// StatusChecker creates a statusChecker with Staging's URL and API Key.
func (s staging) StatusChecker() statusChecker {
	return statusChecker{
		client: newClient(s.Url[:len(s.Url)-3], s.ApiKey),
	}
}

// NewStaging is the constructor for the staging struct.
// It adds a /v2 path to the URL in order to properly connect to Discovery.
func NewStaging(url, apiKey string) staging {
	return staging{Url: url + "/v2", ApiKey: apiKey}
}

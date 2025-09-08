package discovery

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

type stagingGetContentOption func(*map[string][]string)

func WithContentAction(action string) stagingGetContentOption {
	return func(m *map[string][]string) {
		(*m)["action"] = append((*m)["action"], action)
	}
}

func WithIncludeProjections(include []string) stagingGetContentOption {
	return func(m *map[string][]string) {
		(*m)["include"] = append((*m)["include"], include...)
	}
}

func WithExcludeProjections(exclude []string) stagingGetContentOption {
	return func(m *map[string][]string) {
		(*m)["exclude"] = append((*m)["exclude"], exclude...)
	}
}

type contentClient struct {
	client
}

func newContentClient(url, apiKey, bucketName string) bucketsClient {
	return bucketsClient{
		client: newClient(url+"/content/"+bucketName, apiKey),
	}
}

func (c contentClient) Store(contentId, parentId string, content gjson.Result) (gjson.Result, error) {
	return execute(c.client, http.MethodPost, "/"+contentId, WithQueryParameters(map[string][]string{
		"parentId": {parentId},
	}), WithJSONBody(content.Raw))
}

func (c contentClient) Get(contentId string, options ...stagingGetContentOption) (gjson.Result, error) {
	queryParams := make(map[string][]string)
	for _, opt := range options {
		opt(&queryParams)
	}
	return execute(c.client, http.MethodGet, "/"+contentId, WithQueryParameters(queryParams))
}

func (c contentClient) Delete(contentId string) (gjson.Result, error) {
	return execute(c.client, http.MethodDelete, "/"+contentId)
}

func (c contentClient) DeleteMany(parentId string, filter gjson.Result) (gjson.Result, error) {
	return execute(c.client, http.MethodDelete, "", WithQueryParameters(map[string][]string{
		"parentId": {parentId},
	}), WithJSONBody(filter.Raw))
}

type bucketsClient struct {
	client
}

func newBucketsClient(url, apiKey string) bucketsClient {
	return bucketsClient{
		client: newClient(url+"/bucket", apiKey),
	}
}

func (b bucketsClient) Create(bucket string, options gjson.Result) (gjson.Result, error) {
	return execute(b.client, http.MethodPost, "/"+bucket, WithJSONBody(options.Raw))
}

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

func (b bucketsClient) Get(bucket string) (gjson.Result, error) {
	return execute(b.client, http.MethodGet, "/"+bucket)
}

func (b bucketsClient) Delete(bucket string) (gjson.Result, error) {
	return execute(b.client, http.MethodDelete, "/"+bucket)
}

func (b bucketsClient) Purge(bucket string) (gjson.Result, error) {
	return execute(b.client, http.MethodDelete, "/"+bucket+"/purge")
}

func (b bucketsClient) CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error) {
	var parts []string
	for _, r := range config {
		parts = append(parts, r.Raw)
	}
	jsonArray := "[" + strings.Join(parts, ",") + "]"

	return execute(b.client, http.MethodPut, "/"+bucket+"/index/"+index, WithJSONBody(jsonArray))
}

func (b bucketsClient) DeleteIndex(bucket, index string) (gjson.Result, error) {
	return execute(b.client, http.MethodDelete, "/"+bucket+"/index/"+index)
}

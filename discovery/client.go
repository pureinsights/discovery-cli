package discovery

import (
	"strings"

	"github.com/tidwall/gjson"

	"github.com/go-resty/resty/v2"
)

// RequestOption is a type definition used for the functional options pattern.
// It modifies the request, like setting the type of the expected result, adding query parameters, and setting the body.
type RequestOption func(*resty.Request) error

// WithQueryParameters sets the query parameters to the request.
// It works with single value parameters and arrays.
// For example: ?q=query&items=item1&items=item2&items=item3
func WithQueryParameters(params map[string][]string) RequestOption {
	return func(r *resty.Request) error {
		r.SetQueryParamsFromValues(params)
		return nil
	}
}

// WithFile reads a file and adds its contents to the request.
func WithFile(path string) RequestOption {
	return func(r *resty.Request) error {
		r.SetFile("file", path)
		return nil
	}
}

// WithJSONBody sets the JSON string as the body and the application/json content type.
func WithJSONBody(body string) RequestOption {
	return func(r *resty.Request) error {
		r.SetBody(body)
		r.SetHeader("Content-Type", "application/json")
		return nil
	}
}

// Client is a struct that contains the API Key to connect to Discovery and the Resty Client to execute the requests.
type client struct {
	ApiKey string
	client *resty.Client
}

// NewClient returns an instance of a [client] struct.
// The url parameter is the url to which the request is sent.
// For example, http://localhost:8080
func newClient(url, apiKey string) client {
	restyClient := resty.New()
	restyClient.SetBaseURL(url)
	return client{apiKey, restyClient}
}

// NewSubClient returns an instance of a [client] struct whose base URL is the parent client’s base URL with an added path.
// For example, http://localhost:8080/seed
func newSubClient(c client, path string) client {
	newUrl := strings.TrimRight(c.client.BaseURL, "/") + "/" + strings.TrimLeft(path, "/")
	subClient := resty.New()
	subClient.SetBaseURL(newUrl)
	return client{c.ApiKey, subClient}
}

// Execute runs an HTTP request with the client.
// The method parameter is the HTTP verb to be executed.
// The path is added to the client's base URL.
// The request is modified with the specified request options.
// If set, the client's API key is set as the X-API-Key header.
// This function returns the response as a byte array or an error it failed.
func (c client) execute(method, path string, options ...RequestOption) ([]byte, error) {
	request := c.client.R()

	if c.ApiKey != "" {
		request.SetHeader("X-API-Key", c.ApiKey)
	}

	for _, opt := range options {
		if err := opt(request); err != nil {
			return nil, err
		}
	}

	response, err := request.Execute(method, c.client.BaseURL+path)
	if err != nil {
		return nil, err
	}

	if response.IsError() {
		return nil, Error{
			Status: response.StatusCode(),
			Body:   gjson.ParseBytes(response.Body()),
		}
	}

	return response.Body(), nil
}

// Execute runs the client.execute(function), but returns a parsed gjson.Result object instead of a byte array.
// This function is only recommended if the response is known to return a JSON object or array.
func execute(client client, method, path string, options ...RequestOption) (gjson.Result, error) {
	response, err := client.execute(method, path, options...)
	if err != nil {
		return gjson.Result{}, err
	}
	resultJson := gjson.ParseBytes(response)
	return resultJson, nil
}

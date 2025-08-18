package discovery

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/go-resty/resty/v2"
)

// RequestOption is a type definition used for the functional options pattern.
// It modifies the request, like setting the type of the expected result, adding query parameters, and setting the body.
type RequestOption func(*resty.Request) error

func WithResult[T any]() RequestOption {
	return func(r *resty.Request) error {
		var zeroValue T
		r.SetResult(zeroValue)
		return nil
	}
}

func WithQueryParameters(params map[string][]string) RequestOption {
	return func(r *resty.Request) error {
		r.SetQueryParamsFromValues(params)
		return nil
	}
}

func WithBody(body any) RequestOption {
	return func(r *resty.Request) error {
		r.SetBody(body)
		return nil
	}
}

func WithFile(name, path string) RequestOption {
	return func(r *resty.Request) error {
		r.SetFile(name, path)
		return nil
	}
}

func WithContentType(contentType string) RequestOption {
	fmt.Println("hola")
	return func(r *resty.Request) error {
		r.SetHeader("Content-Type", contentType)
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
// This function returns the response with its correct type if it was set, the raw response if not, and an error if any occured.
func (c client) execute(method, path string, options ...RequestOption) (any, error) {
	request := c.client.R()

	if c.ApiKey != "" {
		request.SetHeader("X-API-Key", c.ApiKey)
	}

	for _, opt := range options {
		if err := opt(request); err != nil {
			return nil, err
		}
	}

	fmt.Println(request)
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

	if r := response.Result(); r != nil {
		return r, nil
	}

	return response.Body(), nil
}

func execute[T any](client client, method, path string, options ...RequestOption) (T, error) {
	options = append(options, WithResult[T]())
	response, err := client.execute(method, path, options...)
	if err != nil {
		var zeroValue T
		return zeroValue, err
	}
	tResponse, ok := response.(T)
	if !ok {
		var zeroValue T
		return zeroValue, fmt.Errorf("expected type %T, but got type %T", zeroValue, response)
	}
	return tResponse, nil
}

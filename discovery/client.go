package discovery

import (
	"fmt"
	"net/http"

	"github.com/tidwall/gjson"

	"github.com/go-resty/resty/v2"
)

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
	subClient := resty.New()
	subClient.SetBaseURL(c.client.BaseURL + path)
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

	response, err := request.Execute(method, c.client.BaseURL+path)

	if response.IsError() {
		return nil, Error{
			Status: response.StatusCode(),
			Body:   gjson.ParseBytes(response.Body()),
		}
	}

	if err != nil {
		return nil, Error{
			Status: http.StatusInternalServerError,
			Body:   gjson.Parse(`{"error":"` + err.Error() + `"}`),
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

package discovery

import (
	"net/http"

	"github.com/tidwall/gjson"

	"github.com/go-resty/resty/v2"
)

type RequestOption func(*resty.Request) error

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

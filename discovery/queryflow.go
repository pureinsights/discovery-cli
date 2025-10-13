package discovery

import (
	"strings"

	"github.com/tidwall/gjson"
)

// QueryFlowProcessorsClient is a struct that performs the CRUD of processors.
type queryFlowProcessorsClient struct {
	crud
	cloner
}

// NewQueryFlowProcessorsClient is the constructor of a queryFlowProcessorsClient
func newQueryFlowProcessorsClient(url, apiKey string) queryFlowProcessorsClient {
	client := newClient(url+"/processor", apiKey)
	return queryFlowProcessorsClient{
		crud: crud{
			getter{
				client: client,
			},
		},
		cloner: cloner{
			client: client,
		},
	}
}

// EndpointsClient is a struct that performs the CRUD of endpoints.
type endpointsClient struct {
	crud
	cloner
	enabler
}

// NewEndpointsClient is the constructor of a newEndpointsClient
func newEndpointsClient(url, apiKey string) endpointsClient {
	client := newClient(url+"/endpoint", apiKey)
	return endpointsClient{
		crud: crud{
			getter{
				client: client,
			},
		},
		cloner: cloner{
			client: client,
		},
		enabler: enabler{
			client: client,
		},
	}
}

// QueryFlow is the struct for the client that can carry out every QueryFlow operation.
type queryFlow struct {
	Url, ApiKey string
}

// Processors creates a queryFlowProcessorsClient with QueryFlow's URL and API Key
func (q queryFlow) Processors() queryFlowProcessorsClient {
	return newQueryFlowProcessorsClient(q.Url, q.ApiKey)
}

// Endpoints creates a endpointsClient with QueryFlow's URL and API Key
func (q queryFlow) Endpoints() endpointsClient {
	return newEndpointsClient(q.Url, q.ApiKey)
}

// BackupRestore creates a backupRestore with QueryFlow's URL and API Key
func (q queryFlow) BackupRestore() backupRestore {
	return backupRestore{
		client: newClient(q.Url, q.ApiKey),
	}
}

// Invoke is a function that calls the API version of an endpoint.
// It returns the endpoint's response as a gjson.Result or an error if any occurred.
func (q queryFlow) Invoke(method, uri string, options ...RequestOption) (gjson.Result, error) {
	newUri := "/api/" + strings.TrimPrefix(uri, "/")
	client := newClient(q.Url, q.ApiKey)
	return execute(client, method, newUri, options...)
}

// Debug is a function that calls the Debug version of an endpoint.
// It returns the endpoint's response as a gjson.Result or an error if one occurred.
func (q queryFlow) Debug(method, uri string, options ...RequestOption) (gjson.Result, error) {
	newUri := "/debug/" + strings.TrimPrefix(uri, "/")
	client := newClient(q.Url, q.ApiKey)
	return execute(client, method, newUri, options...)
}

// NewQueryFlow is the constructor for the QueryFlow struct.
func NewQueryFlow(url, apiKey string) queryFlow {
	return queryFlow{Url: url, ApiKey: apiKey}
}

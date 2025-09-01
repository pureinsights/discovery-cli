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

type queryFlow struct {
	Url, ApiKey string
}

func (q queryFlow) Processors() queryFlowProcessorsClient {
	return newQueryFlowProcessorsClient(q.Url, q.ApiKey)
}

func (q queryFlow) Endpoints() endpointsClient {
	return newEndpointsClient(q.Url, q.ApiKey)
}

func (q queryFlow) Invoke(method, uri string, options ...RequestOption) (gjson.Result, error) {
	newUri := "/api/" + strings.TrimLeft(uri, "/")
	client := newClient(q.Url, q.ApiKey)
	response, err := execute(client, method, newUri, options...)
	if err != nil {
		return gjson.Result{}, nil
	}

	return response, nil
}

func (q queryFlow) Debug(method, uri string, options ...RequestOption) (gjson.Result, error) {
	newUri := "/debug/" + strings.TrimLeft(uri, "/")
	client := newClient(q.Url, q.ApiKey)
	response, err := execute(client, method, newUri, options...)
	if err != nil {
		return gjson.Result{}, nil
	}

	return response, nil
}

func (q queryFlow) BackupRestore() backupRestore {
	return backupRestore{
		client: newClient(q.Url, q.ApiKey),
	}
}

package discovery

import (
	"strings"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

// queryFlowProcessorsClient is a struct that performs the CRUD of processors.
type queryFlowProcessorsClient struct {
	crud
	cloner
	searcher
}

// newQueryFlowProcessorsClient is the constructor of a queryFlowProcessorsClient.
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
		searcher: searcher{
			client: client,
		},
	}
}

// queryFlowPipelinesClient is the struct that performs the CRUD and cloning of pipelines.
type queryFlowPipelinesClient struct {
	crud
	cloner
	searcher
}

// NewQueryFlowPipelinesClient is the constructor of a queryFlowPipelinesClient.
func newQueryFlowPipelinesClient(url, apiKey string) queryFlowPipelinesClient {
	client := newClient(url+"/pipeline", apiKey)
	return queryFlowPipelinesClient{
		crud: crud{
			getter{
				client: client,
			},
		},
		cloner: cloner{
			client: client,
		},
		searcher: searcher{
			client: client,
		},
	}
}

// EndpointsClient is a struct that performs the CRUD of endpoints.
type endpointsClient struct {
	crud
	cloner
	enabler
	searcher
}

// newEndpointsClient is the constructor of a new endpointsClient.
func newEndpointsClient(url, apiKey string) endpointsClient {
	client := newClient(url+"/entrypoint/endpoint", apiKey)
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
		searcher: searcher{
			client: client,
		},
	}
}

// mcpServersClient is a struct that performs the CRUD of MCP servers.
type mcpServersClient struct {
	crud
	cloner
	enabler
	searcher
}

// newMCPServersClient is the constructor of a new mcpServersClient.
func newMCPServersClient(url, apiKey string) mcpServersClient {
	client := newClient(url+"/entrypoint/mcp-server", apiKey)
	return mcpServersClient{
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
		searcher: searcher{
			client: client,
		},
	}
}

// toolsClient is a struct that performs the CRUD of tools in MCP servers.
type toolsClient struct {
	crud
	cloner
	searcher
}

// newToolsClient is the constructor of a toolsClient.
func newToolsClient(sc mcpServersClient, serverId uuid.UUID) toolsClient {
	client := newSubClient(sc.crud.client, "/"+serverId.String()+"/tool")
	return toolsClient{
		crud: crud{
			getter{
				client: client,
			},
		},
		cloner: cloner{
			client: client,
		},
		searcher: searcher{
			client: client,
		},
	}
}

// Tools creates a new toolsClient.
func (sc mcpServersClient) Tools(serverId uuid.UUID) toolsClient {
	return newToolsClient(sc, serverId)
}

// queryFlow is the struct for the client that can carry out every QueryFlow operation.
type queryFlow struct {
	Url, ApiKey string
}

// Processors creates a queryFlowProcessorsClient with QueryFlow's URL and API Key.
func (q queryFlow) Processors() queryFlowProcessorsClient {
	return newQueryFlowProcessorsClient(q.Url, q.ApiKey)
}

// Pipelines is used to create a queryFlowPipelinesClient.
func (q queryFlow) Pipelines() queryFlowPipelinesClient {
	return newQueryFlowPipelinesClient(q.Url, q.ApiKey)
}

// Endpoints creates an endpointsClient with QueryFlow's URL and API Key.
func (q queryFlow) Endpoints() endpointsClient {
	return newEndpointsClient(q.Url, q.ApiKey)
}

// MCPServers creates a new mcpServersClient with QueryFlow's URL and API Key.
func (q queryFlow) MCPServers() mcpServersClient {
	return newMCPServersClient(q.Url, q.ApiKey)
}

// BackupRestore creates a backupRestore with QueryFlow's URL and API Key.
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

// StatusChecker creates a statusChecker with QueryFlow's URL and API Key.
func (q queryFlow) StatusChecker() statusChecker {
	return statusChecker{
		client: newClient(q.Url[:len(q.Url)-3], q.ApiKey),
	}
}

// NewQueryFlow is the constructor for the QueryFlow struct.
// It adds a /v2 path to the URL in order to properly connect to Discovery.
func NewQueryFlow(url, apiKey string) queryFlow {
	return queryFlow{Url: strings.TrimRight(url, "/") + "/v2", ApiKey: apiKey}
}

package discovery

// QueryFlowProcessorsClient is a struct that performs the CRUD of processors.
type queryFlowProcessorsClient struct {
	crud
	cloner
}

// NewQueryFlowProcessorsClient is the constructor of a queryFlowProcessorsClient
func newQueryFlowProcessorsClient(queryflow client) queryFlowProcessorsClient {
	client := newSubClient(queryflow, "/processor")
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
func newEndpointsClient(queryflow client) endpointsClient {
	client := newSubClient(queryflow, "/endpoint")
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

package discovery

type queryFlowProcessorsClient struct {
	crud
	cloner
}

// newQueryFlowProcessorsClient
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

type endpointsClient struct {
	crud
	cloner
	enabler
}

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

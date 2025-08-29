package discovery

type queryFlowProcessorsClient struct {
	crud
	cloner
}

// NewLabelsClient is the constructor of a labelsClient
func newLabelsClient(core client) labelsClient {
	return labelsClient{
		crud{
			getter{
				client: newSubClient(core, "/label"),
			},
		},
	}
}

type endpointsClient struct {
	crud
	cloner
	enabler
}

package discovery

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

type labelsClient struct {
	crud
}

func newLabelsClient(core client) labelsClient {
	return labelsClient{
		crud{
			getter{
				client: newSubClient(core, "/label"),
			},
		},
	}
}

type secretsClient struct {
	crud
}

func newSecretsClient(core client) secretsClient {
	return secretsClient{
		crud{
			getter{
				client: newSubClient(core, "/secret"),
			},
		},
	}
}

type credentialsClient struct {
	crud
	searcher
	cloner
}

func newCredentialsClient(core client) credentialsClient {
	client := newSubClient(core, "/credential")
	return credentialsClient{
		crud: crud{
			getter{
				client: client,
			},
		},
		searcher: searcher{
			client: client,
		},
		cloner: cloner{
			client: client,
		},
	}
}

type serversClient struct {
	crud
	searcher
	cloner
}

func newServersClient(core client) serversClient {
	client := newSubClient(core, "/server")
	return serversClient{
		crud: crud{
			getter{
				client: client,
			},
		},
		searcher: searcher{
			client: client,
		},
		cloner: cloner{
			client: client,
		},
	}
}

func Ping(id uuid.UUID) serversClient

type filesClient struct {
	client
}

func newFilesClient(core client) filesClient {
	client := newSubClient(core, "/file")
	return filesClient{
		client: client,
	}
}

func (fc filesClient) Upload(key, file string) (gjson.Result, error) {
	response, err := execute(fc.client, http.MethodPut, key, WithFile(file))
	if err != nil {
		return gjson.Result{}, err
	}

	return response, nil
}

func Retrieve(key string) ([]byte, error)
func List() ([]gjson.Result, error)
func Delete() (gjson.Result, error)

type maintenanceClient struct {
	client
}

func newMaintenanceClient(core client) maintenanceClient {
	return maintenanceClient{
		client: newSubClient(core, "/maintenance"),
	}
}

func Log(componentName, level, loggerName string) error

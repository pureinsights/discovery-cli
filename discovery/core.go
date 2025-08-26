package discovery

import (
	"encoding/json"
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
	// searcher
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
		// searcher: searcher{
		// 	client: client,
		// },
		cloner: cloner{
			client: client,
		},
	}
}

type serversClient struct {
	crud
	//	searcher
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
		// searcher: searcher{
		// 	client: client,
		// },
		cloner: cloner{
			client: client,
		},
	}
}

func (sc serversClient) Ping(id uuid.UUID) (gjson.Result, error) {
	pingServer, err := execute(sc.client, http.MethodGet, "/"+id.String()+"/ping")
	if err != nil {
		return gjson.Result{}, err
	}

	return pingServer, nil
}

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
	response, err := execute(fc.client, http.MethodPut, "/"+key, WithFile(file))
	if err != nil {
		return gjson.Result{}, err
	}

	return response, nil
}

func (fc filesClient) Retrieve(key string) ([]byte, error) {
	file, err := fc.execute(http.MethodGet, "/"+key)
	if err != nil {
		return []byte{}, err
	}

	return file, nil
}

func (fc filesClient) List() ([]string, error) {
	filesBytes, err := fc.execute(http.MethodGet, "")
	if err != nil {
		return []string{}, err
	}

	var files []string
	if err := json.Unmarshal(filesBytes, &files); err != nil {
		return []string{}, err
	}
	return files, nil
}

func (fc filesClient) Delete(key string) (gjson.Result, error) {
	acknowledged, err := execute(fc.client, http.MethodDelete, "/"+key)
	if err != nil {
		return gjson.Result{}, err
	}

	return acknowledged, nil
}

type LogLevel string

// The constants represent the options to ignore the new duplicated entities, fail if there are duplicated entities, and update the duplicated entities with the new values.
const (
	LevelError LogLevel = "ERROR"
	LevelWarn  LogLevel = "WARN"
	LevelInfo  LogLevel = "INFO"
	LevelDebug LogLevel = "DEBUG"
	LevelTrace LogLevel = "TRACE"
)

type maintenanceClient struct {
	client
}

func newMaintenanceClient(core client) maintenanceClient {
	return maintenanceClient{
		client: newSubClient(core, "/maintenance"),
	}
}

func (mc maintenanceClient) Log(componentName string, level LogLevel, loggerName string) (gjson.Result, error) {
	acknowledged, err := execute(mc.client, http.MethodPost, "/log", WithQueryParameters(map[string][]string{"componentName": {componentName}, "level": {string(level)}, "loggerName": {loggerName}}))
	if err != nil {
		return gjson.Result{}, err
	}

	return acknowledged, nil
}

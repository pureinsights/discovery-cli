package discovery

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

// LabelsClient is the struct that performs the CRUD of labels
type labelsClient struct {
	crud
}

// NewLabelsClient is the constructor of a labelsClient
func newLabelsClient(url, apiKey string) labelsClient {
	return labelsClient{
		crud{
			getter{
				client: newClient(url+"/label", apiKey),
			},
		},
	}
}

// SecretsClient is the struct that performs the CRUD of secrets
type secretsClient struct {
	crud
}

// NewSecretsClient creates a new secretsClient
func newSecretsClient(url, apiKey string) secretsClient {
	return secretsClient{
		crud{
			getter{
				client: newClient(url+"/secret", apiKey),
			},
		},
	}
}

// CredentialsClient is the struct that performs the CRUD of credentials
type credentialsClient struct {
	crud
	cloner
}

// NewCredentialsClient creates a new credentialsClient.
func newCredentialsClient(url, apiKey string) credentialsClient {
	client := newClient(url+"/credential", apiKey)
	return credentialsClient{
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

// ServersClient is the struct that performs the CRUD of servers
type serversClient struct {
	crud
	cloner
}

// NewServersClient creates a new serversClient
func newServersClient(url, apiKey string) serversClient {
	client := newClient(url+"/server", apiKey)
	return serversClient{
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

// Ping calls the endpoint to verify the connection to a server.
// It returns acknowledged: true if the connection was successful.
func (sc serversClient) Ping(id uuid.UUID) (gjson.Result, error) {
	return execute(sc.client, http.MethodGet, "/"+id.String()+"/ping")
}

// FilesClient is the struct that performs the CRUD of files
type filesClient struct {
	client
}

// NewFilesClient is the constructor of the filesClient struct
func newFilesClient(url, apiKey string) filesClient {
	client := newClient(url+"/file", apiKey)
	return filesClient{
		client: client,
	}
}

// Upload receives a key and file and sends it to Discovery.
// It returns acknowledged: true if the upload was successful.
func (fc filesClient) Upload(key, file string) (gjson.Result, error) {
	return execute(fc.client, http.MethodPut, "/"+key, WithFile(file))
}

// Retrieve obtains a file's data and returns it as an array of bytes.
// It receives the key that corresponds to the file.
func (fc filesClient) Retrieve(key string) ([]byte, error) {
	return fc.execute(http.MethodGet, "/"+key)
}

// List displays an array of strings that contains every file key that is stored in Discovery.
// If there are no keys, the endpoint returns a No Content response and the function returns an empty array.
func (fc filesClient) List() ([]string, error) {
	filesBytes, err := fc.execute(http.MethodGet, "")
	if err != nil {
		return []string(nil), err
	}
	if len(filesBytes) > 0 {
		var files []string
		if err := json.Unmarshal(filesBytes, &files); err != nil {
			return []string(nil), err
		}
		return files, nil
	} else {
		return []string{}, nil
	}
}

// Delete removes a file from Discovery based on the sent key.
// It returns a JSON with the deletion acknowledgement or an error if any occured.
func (fc filesClient) Delete(key string) (gjson.Result, error) {
	return execute(fc.client, http.MethodDelete, "/"+key)
}

// LogLevel is used as an enum to easily represent the logging levels.
type LogLevel string

// The constants represent the respective log level.
const (
	LevelError LogLevel = "ERROR"
	LevelWarn  LogLevel = "WARN"
	LevelInfo  LogLevel = "INFO"
	LevelDebug LogLevel = "DEBUG"
	LevelTrace LogLevel = "TRACE"
)

// MaintenanceClient is the struct that the Core's maintenance operations.
type maintenanceClient struct {
	client
}

// newMaintenanceClient creates a maintenanceClient.
func newMaintenanceClient(url, apiKey string) maintenanceClient {
	return maintenanceClient{
		client: newClient(url+"/maintenance", apiKey),
	}
}

// Log receives the component's name, log level, and an optional logger name to change that component's log level.
// If the logger name is empty, all of the loggers in the component receive the new log level.
// If the logger name is specified, only that logger has its log level changed.
// The log endpoint often returns an acknowledged: true, even if the component does not exist.
// If the request to change the log level failed, a specific log with details of what happens appear in the Discovery component's logs, not on the response to the request.
func (mc maintenanceClient) Log(componentName string, level LogLevel, loggerName string) (gjson.Result, error) {
	return execute(mc.client, http.MethodPost, "/log", WithQueryParameters(map[string][]string{"componentName": {componentName}, "level": {string(level)}, "loggerName": {loggerName}}))
}

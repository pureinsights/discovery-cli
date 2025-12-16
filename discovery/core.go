package discovery

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

// labelsClient is the struct that performs the CRUD of labels.
type labelsClient struct {
	crud
}

// newLabelsClient is the constructor of a labelsClient.
func newLabelsClient(url, apiKey string) labelsClient {
	return labelsClient{
		crud{
			getter{
				client: newClient(url+"/label", apiKey),
			},
		},
	}
}

// secretsClient is the struct that performs the CRUD of secrets.
type secretsClient struct {
	crud
}

// newSecretsClient creates a new secretsClient.
func newSecretsClient(url, apiKey string) secretsClient {
	return secretsClient{
		crud{
			getter{
				client: newClient(url+"/secret", apiKey),
			},
		},
	}
}

// credentialsClient is the struct that performs the CRUD of credentials.
type credentialsClient struct {
	crud
	cloner
	searcher
}

// newCredentialsClient creates a new credentialsClient.
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
		searcher: searcher{
			client: client,
		},
	}
}

// serversClient is the struct that performs the CRUD of servers.
type serversClient struct {
	crud
	cloner
	searcher
}

// newServersClient creates a new serversClient.
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
		searcher: searcher{
			client: client,
		},
	}
}

// Ping calls the endpoint to verify the connection to a server.
// It returns acknowledged: true if the connection was successful.
func (sc serversClient) Ping(id uuid.UUID) (gjson.Result, error) {
	return execute(sc.crud.client, http.MethodGet, "/"+id.String()+"/ping")
}

// filesClient is the struct that performs the CRUD of files.
type filesClient struct {
	client
}

// newFilesClient is the constructor of the filesClient struct.
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
	// LevelError is used to change the log level to "ERROR"
	LevelError LogLevel = "ERROR"
	// LevelWarn is used to change the log level to "WARN"
	LevelWarn LogLevel = "WARN"
	// LevelInfo is used to change the log level to "INFO"
	LevelInfo LogLevel = "INFO"
	// LevelDebug is used to change the log level to "DEBUG"
	LevelDebug LogLevel = "DEBUG"
	// LevelTrace is used to change the log level to "TRACE"
	LevelTrace LogLevel = "TRACE"
)

// maintenanceClient is the struct that the Core's maintenance operations.
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

// core is the struct for the client that can execute every Core operation.
type core struct {
	Url, ApiKey string
}

// Servers creates a serversClient with the core's URL and API Key.
func (c core) Servers() serversClient {
	return newServersClient(c.Url, c.ApiKey)
}

// Credentials creates a credentialsClient with the core's URL and API Key.
func (c core) Credentials() credentialsClient {
	return newCredentialsClient(c.Url, c.ApiKey)
}

// Secrets creates a secretsClient with the core's URL and API Key.
func (c core) Secrets() secretsClient {
	return newSecretsClient(c.Url, c.ApiKey)
}

// Labels creates a labelsClient with the core's URL and API Key.
func (c core) Labels() labelsClient {
	return newLabelsClient(c.Url, c.ApiKey)
}

// Files creates a filesClient with the core's URL and API Key.
func (c core) Files() filesClient {
	return newFilesClient(c.Url, c.ApiKey)
}

// Maintenance creates a maintenanceClient with the core's URL and API Key.
func (c core) Maintenance() maintenanceClient {
	return newMaintenanceClient(c.Url, c.ApiKey)
}

// BackupRestore creates a backupRestore with the core's URL and API Key.
func (c core) BackupRestore() backupRestore {
	return backupRestore{
		client: newClient(c.Url, c.ApiKey),
	}
}

func (c core) StatusChecker() statusChecker {
	return statusChecker{
		client: newClient(c.Url[:len(c.Url)-3], c.ApiKey),
	}
}

// NewCore is the constructor for the core struct.
// It adds a /v2 path to the URL in order to properly connect to Discovery.
func NewCore(url, apiKey string) core {
	return core{Url: url + "/v2", ApiKey: apiKey}
}

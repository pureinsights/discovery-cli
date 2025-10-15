# Pureinsights Discovery Platform: Command Line Interface

## Client
This struct contains a Resty client to do requests to Discovery's APIs and also stores an API key for authentication. It has two methods to execute an HTTP request to Discovery's APIs, based on the URL configured in the client and the path parameter. One method returns a byte array and the other returns a [gjson.Result](https://github.com/tidwall/gjson#result-type) object, which stores a JSON. 

The execute functions can receive functional options to modify their request. The following are available:

| Option | Description |
| --- | --- |
| WithQueryParameters | This option adds query parameters to the request. They are received with a map of strings to an array of strings. This means that a single parameter can have multiple values. An example of these query parameters is `map[string][]string{"query": {"What is Pureinsights Discovery"}, "items": {"item1", "item2", "item3"}}`.|
| WithFile | This option adds a file to the request. It needs to be able to find the file with the received path.|
| WithJSONBody | This option sets the body as the received JSON body, which must be a valid JSON string. It also sets the Content Type to `application/json`.|

## Common structs
Thanks to the highly standardized API of the Discovery products, most of the endpoints for different entities are almost identical. In some cases, the difference is basically the base URL and some parameters. For this reason, common structs were created to expedite the development of the CLI. They implement methods that many of Discovery's components and entities need.

### Getter
This struct performs all the GET operations.

It inherits from:
* [Client](#client)

It has the following methods:
| Name | Method | Path | Response | Description |
| --- | --- | --- | --- | --- | 
| Get | GET | `{URL}/{UUID}` | `application/json` | Receives a UUID to get the entity referenced by it. It returns the result as a gjson.Result object. |
| GetAll | GET | `{URL}/`  |`application/json` | Obtains every entity from the endpoint. |

### CRUD
This struct creates, reads, updates, and deletes entities.

It inherits from:
* [Client](#client)
* [Getter](#getter)

It has the following methods:
| Name | Method | Path | Request Body | Response | Description |
| --- | --- | --- | --- | --- | --- | 
| Create | POST | `{URL}/{UUID}` | `application/json` | `application/json` | Receives a UUID to get the entity referenced by it. It returns the result as a gjson.Result object. |
| Update | PUT | `{URL}/{UUID}`  | `application/json` | `application/json` | Receives a UUID and the configuration JSON to update an entity in Discovery. It returns the result as a gjson.Result object |
| Delete | DELETE | `{URL}/{UUID}`  | | `application/json` | Deletes the entity identified by the received UUID. |

### Cloner
This struct is used to clone entities. It adds the `/clone` path to its client's URL to duplicate entities.

It inherits from:
* [Client](#client)

It has the following method:
| Name | Method | Path | Response | Description |
| --- | --- | --- | --- | --- | 
| Clone | POST | `{URL}/{UUID}/clone` | `application/json` | Clones the entity with the received ID, which must be a valid UUID. It adds the parameters it receives to the request. These can be the name of the cloned entity and the depth of the cloning operation (like shallow or deep copying). |

### Summarizer
This struct is used to get the summary of an entity, like the summary of jobs from a seed execution, summary of records from a seed execution, or the summary of records from a seed.

It inherits from:
* [Client](#client)

It has the following method:
| Name | Method | Path | Response | Description |
| --- | --- | --- | --- | --- | 
| Summary | POST | `{URL}/{UUID}/summary` | `application/json` | This function adds the `/summary` path to the client's URL and executes the request. It then returns the result of the operation. |

### Enabler
This struct can be used to enable and disable entities, like QueryFlow endpoints and Ingestion seed schedules. 

It inherits from:
* [Client](#client)

It has the following method:
| Name | Method | Path | Response | Description |
| --- | --- | --- | --- | --- | 
| Enable | PATCH | `{URL}/{UUID}/enable` | `application/json` | Receives a valid UUID and enables the entity referenced by it.|
| Disable | PATCH | `{URL}/{UUID}/disable` | `application/json` | Receives a valid UUID and disables the entity referenced by it.|

### BackupRestore
This struct allows for exporting and importing entities from Discovery's components.

It inherits from:
* [Client](#client)

It has the following method:
| Name | Method | Path | Request Body | Query Parameters | Response | Description |
| --- | --- | --- | --- | --- | --- | --- | 
| Export | GET | `{URL}/export` |  |  | `application/octet-stream` | Calls the `/export` endpoint. It returns the result of the endpoint in bytes, which should be written to a ZIP file so that it can be restored later. |
| Import | POST | `{URL}/import` | `multipart/form-data` | `onConflict`: `UPDATE`, `IGNORE`, `FAIL` | `application/json` | Calls the `/import` endpoint. It receives the given file to restore the entities contained within. |

### Searcher
This struct has methods to search for entities in Discovery's components.

It inherits from:
* [Client](#client)

It has the following method:

| Name | Method | Path | Request Body | Response | Description |
| --- | --- | --- | --- | --- | --- |
| Search | POST | `{URL}/search` | `application/json` | `application/json` | Returns an array with the entities that match the given filters. |

## Discovery Clients
### Core Client
Discovery has a Core client struct. 

Its fields are:
| Field | Description |
| --- | --- |
| Url   | The URL of Discovery's Core component. The URL should contain the URL up to the version. For example, `http://localhost:12010/v2`. |
| ApiKey | The API key to authenticate each request to Discovery's Core. |

#### Sub-Clients

##### SecretsClient
This struct manages secrets. 

It inherits from:
* [CRUD](#crud)

Creating a `secretsClient` can be done with `core.Secrets()` or `newSecretsClient(URL, API Key)`.

##### CredentialsClient
This struct manages credentials. 

It inherits from:
* [CRUD](#crud)
* [Cloner](#cloner)

Creating a `credentialsClient` can be done with `core.Credentials()` or `newCredentialsClient(URL, API Key)`.

##### ServersClient
This struct manages servers. 

It inherits from:
* [CRUD](#crud)
* [Cloner](#cloner)

It has the following additional method:
| Name | Method | Path | Response | Description |
| --- | --- | --- | --- | --- | 
| Ping | GET | `{URL}/{UUID}/ping` | `application/json` | This function verifies if the connection to the server was successful and the server is reachable. |

Creating a `serversClient` can be done with `core.Servers()` or `newServersClient(URL, API Key)`.



##### LabelsClient
This struct manages labels. 

It inherits from:
* [CRUD](#crud)

Creating a `labelsClient` can be done with `core.Labels()` or `newLabelsClient(URL, API Key)`.

##### FilesClient
This struct manages Discovery's files. 

It inherits from:
* [Client](#client)

It has the following additional methods:
| Name | Method | Path | Request Body | Response | Description |
| --- | --- | --- | --- | --- | --- | 
| Upload | PUT | `{URL}/file/{KEY}` | `multipart/form-data` | `application/json` | Sends a file to Discovery. |
| Retrieve | GET | `{URL}/file/{KEY}`  |  | `application/octet-stream` | Returns file's data inside a byte array. |
| List | GET | `{URL}/file`  |  | `application/json` | Returns an array with all of the file keys in Discovery. |
| Delete | DELETE | `{URL}/file/{KEY}`  | | `application/json` | Removes the file. |

Creating a `filesClient` can be done with `core.Files()` or `newServersClient(URL, API Key)`.

##### MaintenanceClient
The `maintenanceClient` struct carries out the Core's maintenance operations.

It inherits from:
* [Client](#client)

It has the following method:
| Name | Method | Path | Query Parameters | Response | Description |
| --- | --- | --- | --- | --- | --- 
| Log | POST | `{URL}/maintenance/log` | • `componentName`<br>• `level`: `ERROR`, `WARN`, `INFO`, `DEBUG`, `TRACE`<br>• loggerName | `application/json` | This function changes the log level of a component. It can also change the level of a specific logger inside a component. |
 
 Creating a `maintenanceClient` can be done with `core.Maintenance()` or `newMaintenanceClient(URL, API Key)`.

##### BackupRestore
This struct imports and exports the Core's entities. It is the same struct as the [BackupRestore](#backuprestore) struct

Creating a `backupRestore` can be done with `core.BackupRestore()`.

### QueryFlow Client
Discovery has a QueryFlow client struct. 

Its fields are:
| Field | Description |
| --- | --- |
| Url   | The URL of Discovery's QueryFlow component. The URL should contain the URL up to the version. For example, `http://localhost:12040/v2`. |
| ApiKey | The API key to authenticate each request to Discovery's QueryFlow. |

It has the following methods:
| Name | Method | Path | Request Body | Response | Description |
| --- | --- | --- | --- | --- | --- | 
| Invoke | `{method}` | `{URL}/api/{URI}` | `{Functional Options}` | `application/json` | Calls the endpoint with a `/api` root path added to the URI, which makes QueryFlow return a normal response of the endpoint. It can receive the [Client's](#client) functional options to modify the request. |
| Debug | `{method}` | `{URL}/{UUID}` | `{Functional Options}` | `application/json` | Calls the endpoint with a `/debug` root path, which makes QueryFlow respond with the entire trace of execution the state machine took. Each one of the states, their output, their errors and the overall flow followed by the state machine will be displayed. It can receive the [Client's](#client) functional options to modify the request. |

These are very similar to `client.execute()`, but are used to call QueryFlow's endpoints. The response can vary depending on the URI used on the request.

#### Sub-Clients

##### QueryFlowProcessorsClient
This struct manages QueryFlow's processors. 

It inherits from:
* [CRUD](#crud)
* [Cloner](#cloner)

Creating a `queryFlowProcessorsClient` can be done with `queryFlow.Processors()` or `newQueryFlowProcessorsClient(URL, API Key)`.

##### EndpointsClient
This struct manages QueryFlow's endpoints. 

It inherits from:
* [CRUD](#crud)
* [Cloner](#cloner)
* [Enabler](#enabler)

Creating a `endpointsClient` can be done with `queryFlow.endpointsClient()` or `newEndpointsClient(URL, API Key)`.

##### BackupRestore
This struct imports and exports QueryFlow's entities. It is the same struct as the [BackupRestore](#backuprestore) struct

Creating a `backupRestore` can be done with `queryflow.BackupRestore()`.
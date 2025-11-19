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
| Export | GET | `{URL}/export` |  |  | `application/octet-stream` | Calls the `/export` endpoint. It returns two results. The first one is the bytes of the export, which should be written to a ZIP file so that it can be restored later. The second result is the name of the file of the export that is sent by Discovery in the response's headers. This name is usually `export-{TIMESTAMP}.zip`. |
| Import | POST | `{URL}/import` | `multipart/form-data` | `onConflict`: `UPDATE`, `IGNORE`, `FAIL` | `application/json` | Calls the `/import` endpoint. It receives the given file to restore the entities contained within. |

### Searcher
This struct has methods to search for entities in Discovery's components.

It inherits from:
* [Client](#client)

It has the following method:
| Name | Method | Path | Request Body | Response | Description |
| --- | --- | --- | --- | --- | --- |
| Search | POST | `{URL}/search` | `application/json` | `application/json` | Returns an array with the entities that match the given filters. |
| SearchByName | POST | `{URL}/search` | `application/json` | `application/json` | Returns the JSON object with the best match to the given name or an error if any occured or the entity was not found. |

## Discovery Clients
### Core Client
Discovery has a Core client struct. 

Its fields are:
| Field | Description |
| --- | --- |
| Url   | The URL of Discovery's Core component. The URL must not contain the version, as it is added automatically. For example, `http://localhost:12010`. |
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
* [Searcher](#searcher)

Creating a `credentialsClient` can be done with `core.Credentials()` or `newCredentialsClient(URL, API Key)`.

##### ServersClient
This struct manages servers. 

It inherits from:
* [CRUD](#crud)
* [Cloner](#cloner)
* [Searcher](#searcher)

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
| Url   | The URL of Discovery's QueryFlow component. The URL must not contain the version, as it is added automatically. For example, `http://localhost:12040`. |
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

### Ingestion Client
Discovery has a Ingestion client struct. 

Its fields are:
| Field | Description |
| --- | --- |
| Url | The URL of Discovery's Ingestion component. The URL must not contain the version, as it is added automatically. For example, `http://localhost:12030`. |
| ApiKey | The API key to authenticate each request to Discovery's Ingestion. |

#### Sub-Clients

##### IngestionProcessorsClient
This struct manages Ingestion's processors. 

It inherits from:
* [CRUD](#crud)
* [Cloner](#cloner)

Creating a `ingestionProcessorsClient` can be done with `ingestion.Processors()` or `newIngestionProcessorsClient(URL, API Key)`.

##### PipelinesClient
This struct manages Ingestion's pipelines. 

It inherits from:
* [CRUD](#crud)
* [Cloner](#cloner)

Creating a `pipelinesClient` can be done with `ingestion.Pipelines()` or `newPipelinesClient(URL, API Key)`.

##### SeedsClient
This struct manages Ingestion's seeds. 

It inherits from:
* [CRUD](#crud)
* [Cloner](#cloner)

It has the following methods:
| Name | Method | Path | Query Parameters | Response | Description |
| --- | --- | --- | --- | --- | --- | 
| Start | POST | `{URL}/seed/{UUID}` | `scanType`: `FULL`, `INCREMENTAL` | `application/json` | Starts the execution of a seed. |
| Halt | POST | `{URL}/seed/{UUID}/halt` |  | `application/json` | Halts all of the executions of a seed. |
| Reset | POST | `{URL}/seed/{UUID}/reset` |  | `application/json` | Resets the metadata of the seed and deletes its records. It only works if the seed does not have any active executions. |

Creating a `seedsClient` can be done with `ingestion.Seeds()` or `newSeedsClient(URL, API Key)`.

##### SeedExecutionsClient
This struct manages an execution of a seed. 

It inherits from:
* [Getter](#getter)

It has the following methods:
| Name | Method | Path | Response | Description |
| --- | --- | --- | --- | --- |
| Halt | POST | `{URL}/seed/{UUID}/execution/{UUID}/halt` | `application/json` | Stops the seed's execution based on the given IDs. |
| Audit | GET | `{URL}/seed/{UUID}/execution/{UUID}/audit` | `application/json` | Returns an array with the audited changes of the seed execution, or the stages it has completed up to the method's call. |
| Seed | GET | `{URL}/seed/{UUID}/execution/{UUID}/config/seed` | `application/json` | Returns the configuration of the seed of the execution. |
| Pipeline | GET | `{URL}/seed/{UUID}/execution/{UUID}/config/pipeline/{UUID}` | `application/json` | Returns the configuration of the pipeline the seed execution uses. |
| Processor | GET | `{URL}/seed/{UUID}/execution/{UUID}/config/processor/{UUID}` | `application/json` | Returns the configuration of a processor the seed's pipeline uses, based on the processor's ID. |
| Server | GET | `{URL}/seed/{UUID}/execution/{UUID}/config/server/{UUID}` | `application/json` | Returns the configuration of a server one of the processors in the pipeline uses. |
| Credential | GET | `{URL}/seed/{UUID}/execution/{UUID}/config/credential/{UUID}` | `application/json` | Returns the configuration of the credential that is used by a server in the pipeline. |

Creating a `seedExecutionsClient` can be done with `newSeedExecutionsClient(seedsClient, Seed ID)` or with `seedsClient.Executions()`.

##### SeedExecutionRecordsClient
The `seedExecutionRecordsClient` is the struct that can get the summary of records from a seed execution.

It inherits from:
* [Summarizer](#summarizer)

It can be created with `seedExecutionsClient.Records(Execution ID)` or `newSeedExecutionRecordsClient(seedExecutionsClient, Execution ID)`.

##### SeedExecutionJobsClient
The `seedExecutionJobsClient` is the struct that can get the summary of jobs from a seed execution.

It inherits from:
* [Summarizer](#summarizer)

It can be created with `seedExecutionsClient.Jobs(Execution ID)` or `newSeedExecutionJobsClient(seedExecutionsClient, Execution ID)`.

##### SeedRecordsClient
The `seedRecordsClient` is the struct that can get the records and their summary from a seed.

It inherits from:
* [Summarizer](#summarizer)

It has the following methods:
| Name | Method | Path | Response | Description |
| --- | --- | --- | --- | --- |
| Get | GET | `{URL}/seed/{UUID}/record/{RECORDID}` | `application/json` | Returns the seed record with the given id. |
| GetAll | GET | `{URL}/seed/{UUID}/record` | `application/json` | Returns an array with all of the seed's records. |

It can be created with `seedsClient.Records()` or `newSeedRecordsClient(seedsClient, Seed ID)`.

##### BackupRestore
This struct imports and exports Ingestion's entities. It is the same struct as the [BackupRestore](#backuprestore) struct

Creating a `backupRestore` can be done with `ingestion.BackupRestore()`.

### Staging Client
Discovery has a Staging client struct. 

Its fields are:
| Field | Description |
| --- | --- |
| Url   | The URL of Discovery's Staging component. The URL must not contain the version, as it is added automatically. For example, `http://localhost:12020`. |
| ApiKey | The API key to authenticate each request to Discovery's Staging. |

#### Sub-Clients

##### BucketsClient
This struct manages Staging's buckets. 

It inherits from:
* [Client](#client)

It has the following methods:
| Name | Method | Path | Request Body | Response | Description |
| --- | --- | --- | --- | --- | --- | 
| Create | POST | `{URL}/bucket/{bucketName}` | `application/json` | `application/json` | Creates a new bucket in the Staging Repository. It can receive an options JSON to add configurations and create indices. |
| GetAll | GET | `{URL}/bucket` |  | `application/json` | Obtains the names of every bucket in the Staging Repository. |
| Get | GET | `{URL}/bucket/{bucketName}` |  | `application/json` | Gets the information of the bucket. |
| Delete | DELETE | `{URL}/bucket/{bucketName}` | | `application/json` | Deletes the bucket. |
| Purge | DELETE | `{URL}/bucket/{bucketName}/purge` | | `application/json` | Deletes all of the records in the bucket. |
| CreateIndex | PUT | `{URL}/bucket/{bucketName}/index/{indexName}` | `application/json` | `application/json` | Creates an index on the given bucket. The configuration it receives is an array of JSONs, each with the information of the fields that will be indexed. |
| DeleteIndex | DELETE | `{URL}/bucket/{bucketName}/index/{indexName}` |  | `application/json` | Removes the index on the bucket. |

Creating a `bucketsClient` can be done with `staging.Buckets()` or `newBucketsClient(URL, API Key)`.

##### ContentClient
This struct manages a bucket's content. 

It inherits from:
* [Client](#client)

It has the following methods:
| Name | Method | Path | Request Body | Query Parameters | Response | Description |
| --- | --- | --- | --- | --- | --- | --- | 
| Store | POST | `{URL}/content/{bucketName}/{contentId}` | `application/json` | • `parentId` |`application/json` | Adds the received content JSON to a document with the given Content ID. The Parent ID can be used to establish hierarchical relationships between documents. |
| Get | GET | `{URL}/content/{bucketName}/{contentId}` || • `action`: `STORE`, `DELETE`<br>• `include`<br>• `exclude` | `application/json` | Obtains the information of the record with the given Content ID in the bucket. It can receive functional options described later in the docuumentation. |
| Delete | DELETE | `{URL}/content/{bucketName}/{contentId}` |  |  |`application/json` | Deletes the document with the given content ID in the bucket. |

The functional options for the `Get` method are the following:
| Option | Description |
| --- | --- |
| WithContentAction | Adds the `action` query parameter to the request. Some examples of the values are `STORE` and `DELETE`. This will make the `Get()` function obtain the record with that action in its configuration. |
| WithIncludeProjections | This function receives an array of fields that need to be included in the result of the `GET` request. |
| WithExcludeProjections | This function receives an array of fields that need to be excluded from the result of the `GET` request. |

Creating a `contentClient` can be done with `staging.Content(Bucket Name)` or `newContentClient(URL, API Key, Bucket Name)`.
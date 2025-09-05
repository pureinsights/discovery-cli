# Pureinsights Discovery Platform: Command Line Interface

## Common structs
Thanks to the highly standardized API of the Discovery products, most of the endpoints for different entities are almost identical. In some cases, the difference is basically the base URL and some parameters. For some reason, common structs were created to expedite the development of the CLI. They implement methods that many of Discovery's components and entities need.

### Client
This struct contains a Resty client to do requests to Discovery's APIs. It also stores an API key for authentication. It has the following related methods:
- `newClient(URL, API key)`: Creates a new client.
- `newSubClient(client, path)`: Creates a new subclient, which has the same URL as the client, but with the path at the end.
- `client.execute(method, path, Request Options)`: This method executes a request to Discovery's APIs, based on the URL configured in the client and the path parameter. This function returns the response's body as a byte array. The request options follow the [functional options pattern](https://www.sohamkamani.com/golang/options-pattern/). They are added to the request. The following options are available:
  - `WithQueryParameters(params)`: This option adds query parameters to the request. They are received with a map of strings to an array of strings. This means that a single parameter can have multiple values. An example of these query parameters is `map[string][]string{"query": {"What is Pureinsights Discovery"}, "items": {"item1", "item2", "item3"}}`.
  - `WithFile(path)`: This option adds a file to the request. It needs to be able to find the path with the received path.
  - `WithJSONBody(body)`: This option sets the body as the received JSON body, which must be a valid JSON string. It also sets the Content Type to `application/json`.
- `execute(client, method, path, Request Options)`: This function is essentially the same as `client.execute()`, but returns a parsed response as a JSON Object, which in this case is a [gjson.Result](https://github.com/tidwall/gjson#result-type) type.

### Getter
This struct performs all the `GET` operations. It receives a client field by composition, so it can access the API Key and Resty Client fields. The Resty client must have its URL pointing to the entity's endpoint. It has the following methods:
- `Get(id)`: This function receives a UUID to get the entity referenced by it. It returns the result as `gjson.Result` object.
- `GetAll()`: This function retrieves every entity from the endpoint. It returns the results as an array of `gjson.Result` objects. It automatically iterates through every page of the results to get all entities.

### CRUD
This struct creates, reads, updates, and deletes entities. It receives a getter field by composition, which means it can access the `Get()` and `GetAll()`, as well as the fields and methods from the `Client` struct. The Resty client needs to have its URL set to the URL of the entity's endpoint. It has the following methods:
- `Create(Configuration)`: This function receives the configuration of the entity as a JSON object and executes the request to create the entity in the Discovery component.
- `Update(id, configuration)`: This function receives a UUID and the configuration JSON to update an entity in Discovery.
- `Delete(id)`: This function deletes the entity identified by the received UUID.

### Cloner
This struct is used to clone entities. It has a client field by composition. It adds the `/clone` path to its client's URL to duplicate entities. It has the following method:
- `Clone(id, params)`: This function clones the entity with the received ID, which must be a valid UUID. It adds the parameters it receives to the request. These can be the name of the cloned entity and the depth of the cloning operation (like shallow or deep copying).

### Summarizer
This struct is used to get the summary of ingestion seeds, like the summary of jobs from a seed execution, summary of records from a seed execution, or the summary of records from a seed. It has a client field by composition. To obtain the three mentioned summaries, three different summarizers need to be made, each with its client pointing to the correct endpoint in its URL. The struct has one method:
- `Summarize()`: This function adds the `/summary` path to the client's URL and executes the request. It then returns the result of the operation.

### Enabler
This struct can be used to enable and disable entities, like Queryflow endpoints and Ingestion seed schedules. It has a client field by composition. It has two methods:
- `Enable(id)`: This function receives a valid UUID and enables the entity referenced by it.
- `Disable(id)`: This function receives a valid UUID and disables the entity referenced by it.  

### BackupRestore
This struct allows for exporting and importing entities from Discovery's components. It has a client field by composition. Its Resty Client needs to be pointing to the Discovery component's base URL. It has two methods:
- `Export()`: This function calls the `/export` endpoint. It returns the result of the endpoint in bytes, which should be written to a ZIP file so that it can be restored later.
- `Import(On Conflict, file)`: This function calls the `/import` endpoint. The `onConflict` parameter should be one of the constants that represent the conflict resolution strategies. `OnConflictIgnore` sets the strategy to `IGNORE`, which ignores the entities that already exist. `OnConflictFail` sets the strategy to `FAIL`, which makes the operation fail if entities already exist. `OnConflictUpdate` sets the strategy to `UPDATE`, which updates the entities that already exist. The file it receives must be the path to a file that exists.

### Error
This struct is used as the Errors that the CLI returns. It has a Status, an integer, and a Body, a JSON object (`gjson.Result`). It has one method:
- `Error()`: This method serves to fulfill Go's error interface. It returns a string with the error's information.

## Core Client
Discovery has a core client struct. Its fields are:
- Url: The URL of Discovery's Core component. The URL should contain the URL up to the version. For example, `http://localhost:8080/v2`. 
- ApiKey: The API key needed to authenticate to Discovery's Core.  

To create a Core client, the `NewCore(URL, API Key)` is used.
The core client can create subclients that handle the Core's functions. These are the following:

### SecretsClient
The `secretsClient` manages secrets. It is a struct with an embedded `CRUD` struct. It has access to the `execute()`, `Create()`, `Get()`, `GetAll()`, `Update()`, and `Delete()` methods. Creating a `secretsClient` can be done with `core.Secrets()` or `newSecretsClient(URL, API Key)`.

### CredentialsClient
The `credentialsClient` manages credentials. Its struct has embedded `CRUD` and `Cloner` structs. It has access to the `Create()`, `Get()`, `GetAll()`, `Update()`, `Delete()`, and `Clone()` methods. Creating a `credentialsClient` can be done with `core.Credentials()` or `newCredentialsClient(URL, API Key)`.

### ServersClient
The `serversClient` manages servers. It is a struct with embedded `CRUD` and `Cloner` structs. It has access to the `execute()`, `Create()`, `Get()`, `GetAll()`, `Update()`, `Delete()`, and `Clone()`. Aditionally, it has a `Ping()` method to verify if the connection to the server was successful and the server is reachable. Creating a `serversClient` can be done with `core.Servers()` or `newServersClient(URL, API Key)`.

### FilesClient
The `filesClient` manages Discovery's files. It has an embedded `Client` struct, so it can access the `Client`'s `execute()` method. It has an `Upload()` method to send files to Discovery, `Retrieve()` to get a file's data in a byte array, `List()` to get the keys of all of the files stored in Discovery, and `Delete()` to remove a file. Creating a `filesClient` can be done with `core.Files()` or `newServersClient(URL, API Key)`.

### BackupRestore
The `backupRestore` struct imports and exports entities. Its `Export()` method obtains the data of all of the entities, which can later be saved to a ZIP file. The `Import()` method restores the entities described in the sent file. If there are conflicts, Discovery can be set to ignore them, fail, or update them. Creating a `backupRestore` can be done with `core.BackupRestore()`.

### LabelsClient
The `labelsClient` is a struct that manages labels. It has an embedded `CRUD` struct. It has access to the `execute()`, `Create()`, `Get()`, `GetAll()`, `Update()`, and `Delete()` methods. Creating a `labelsClient` can be done with `core.Labels()` or `newLabelsClient(URL, API Key)`.

### MaintenanceClient
The `maintenanceClient` struct has a `Client` struct. It has access to the `Client`'s `execute()` method. It has a `Log()` method that changes the log level of a component. It can also change the level of a specific logger inside a component. Creating a `maintenanceClient` can be done with `core.Maintenance()` or `newMaintenanceClient(URL, API Key)`.

## QueryFlow Client
Discovery has a QueryFlow client struct. Its fields are:
- Url: The URL of Discovery's QueryFlow component. The URL should contain the URL up to the version. For example, `http://localhost:8088/v2`. 
- ApiKey: The API key needed to authenticate to QueryFlow.  

To create a QueryFlow client, the `NewQueryFlow(URL, API Key)` is used.
The QueryFlow client can create subclients with useful functions. These are the following:

### QueryFlowProcessorsClient
The `queryFlowProcessorsClient` manages QueryFlow's processors. It is a struct with embedded `CRUD` and `Cloner` structs. It has access to the `execute()`, `Create()`, `Get()`, `GetAll()`, `Update()`, `Delete()`, and `Clone()` methods. Creating a `queryFlowProcessorsClient` can be done with `queryFlow.Processors()` or `newQueryFlowProcessorsClient(URL, API Key)`.

### EndpointsClient
The `endpointsClient` manages credentials. Its struct has embedded `CRUD`, `Cloner`, and `Enabler` structs. It has access to the `Create()`, `Get()`, `GetAll()`, `Update()`, `Delete()`, `Clone()`, `Enable`, and `Disable` methods. Creating a `endpointsClient` can be done with `queryFlow.endpointsClient()` or `newEndpointsClient(URL, API Key)`.

### BackupRestore
The `backupRestore` struct imports and exports entities. Its `Export()` method obtains the data of all of the entities, which can later be saved to a ZIP file. The `Import()` method restores the entities described in the sent file. If there are conflicts, Discovery can be set to ignore them, fail, or update them. Creating a `backupRestore` can be done with `queryFlow.BackupRestore()`.

### Invoke and Debug
The QueryFlow client also has two important methods: `Invoke()` and `Debug()`. These are very similar to `client.execute()`, but are used to call QueryFlow's endpoints. The response can vary depending on the URI used on the request. `Invoke()` calls the endpoint with a `/api` root path to the URI, which adds makes QueryFlow return a normal response of the endpoint. On the other hand, `Debug()` calls the endpoint with a `/debug` root path, which makes QueryFlow respond with the entire trace of execution the state machine took. Each one of the states, their output, their errors and the overall flow followed by the state machine will be displayed.

## Ingestion Client
Discovery has an Ingestion client struct. Its fields are:
- Url: The URL of Discovery's Ingestion component. The URL should contain the URL up to the version. For example, `http://localhost:8083/v2`. 
- ApiKey: The API key needed to authenticate to Ingestion.  

To create a Ingestion client, the `NewIngestion(URL, API Key)` is used.
The Ingestion client can create subclients with useful functions. These are the following:

### IngestionProcessorsClient
The `ingestionProcessorsClient` manages Ingestion's processors. It is a struct with embedded `CRUD` and `Cloner` structs. It has access to the `execute()`, `Create()`, `Get()`, `GetAll()`, `Update()`, `Delete()`, and `Clone()` methods. Creating a `ingestionProcessorsClient` can be done with `ingestion.Processors()` or `newIngestionProcessorsClient(URL, API Key)`.

### PipelinesClient
The `pipelinesClient` manages Ingestion's pipelines. It is a struct with embedded `CRUD` and `Cloner` structs. It has access to the `execute()`, `Create()`, `Get()`, `GetAll()`, `Update()`, `Delete()`, and `Clone()` methods. Creating a `pipelinesClient` can be done with `ingestion.Pipelines()` or `newPipelinesClient(URL, API Key)`.

### SeedsClient
The `seedsClient` manages Ingestion's seeds. It is a struct with embedded `CRUD` and `Cloner` structs. It has access to the `execute()`, `Create()`, `Get()`, `GetAll()`, `Update()`, `Delete()`, and `Clone()` methods. Creating a `seedsClient` can be done with `ingestion.Seeds()` or `newSeedsClient(URL, API Key)`. 

This struct has additional methods:
- `Start(Seed ID)`: Starts the execution of a seed.
- `Halt(Seed ID)`: Halts all of the executions of a seed.
- `Reset(Seed ID)`: Resets the metadata of the seed and deletes its records. It only works if the seed does not have any active executions.
- `Records(Seed ID)`: Creates a `seedRecordsClient`.
- `Executions(Seed ID)`: Creates a `seedExecutionsClient`.

### SeedExecutionsClient
The `seedExecutionsClient` manages an execution of a seed. It has an embedded `Getter` struct, so it has access to the `Get()` and `GetAll()` functions. Creating a `seedExecutionsClient` can be done with `newSeedExecutionsClient(seedsClient, Seed ID)` or with `seedsClient.Executions()`.

This struct has additional methods:
- `Halt(Execution ID)`: Stops the execution of a seed's execution based on the execution ID.
-  `Audit(Execution ID)`: Returns an array with the audited changes of the seed execution, or the stages it has completed up to the method's call.
-  `Seed(Execution ID)`: Returns the configuration of the seed of the execution.
-  `Pipeline(Execution ID, Pipeline ID)`: Returns the configuration of the pipeline the seed execution uses.
-  `Processor(Execution ID, Processor ID)`: Returns the configuration of a processor the seed's pipeline uses, based on the processor's ID.
-  `Server(Execution ID, Server ID)`: Returns the configuration of a server one of the processors in the pipeline uses.
-  `Credential(Execution ID, Credential ID)`: Returns the configuration of the credential that is used by a server in the pipeline.
-  `Records(Execution ID)`: Creates a `seedExecutionRecordsClient` with the execution ID.
-  `Jobs(Execution ID)`: Creates a `seedExecutionJobsClient` with the execution ID.

#### SeedExecutionRecordsClient
The `seedExecutionRecordsClient` is the struct that can get the summary of records from a seed execution. It has an embedded `Summarizer` struct, so it has access to the `Summarize()` method. It can be created with `seedExecutionsClient.Records(Execution ID)` or `newSeedExecutionRecordsClient(seedExecutionsClient, Execution ID)`.

#### SeedExecutionJobsClient
The `seedExecutionJobsClient` is the struct that can get the summary of jobs from a seed execution. It has an embedded `Summarizer` struct, so it has access to the `Summarize()` method. It can be created with `seedExecutionsClient.Jobs(Execution ID)` or `newSeedExecutionJobsClient(seedExecutionsClient, Execution ID)`.

### SeedRecordsClient
The `seedRecordsClient` is the struct that can get the records and their summary from a seed. It has embedded `Getter` and `Summarizer` structs, so it has access to the `Get(Record ID)`, `GetAll()`, `Summarize()` method. The `Get()` method had to be overridden because records do not use a UUID as their ID, so this iteration receives a string. It can be created with `seedsClient.Records()` or `newSeedRecordsClient(seedsClient, Seed ID)`.

### BackupRestore
The `backupRestore` struct imports and exports entities. Its `Export()` method obtains the data of all of the entities, which can later be saved to a ZIP file. The `Import()` method restores the entities described in the sent file. If there are conflicts, Discovery can be set to ignore them, fail, or update them. Creating a `backupRestore` can be done with `ingestion.BackupRestore()`.
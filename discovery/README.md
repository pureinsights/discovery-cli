# Pureinsights Discovery Platform: Command Line Interface

## Common structs
Thanks to the highly standardized API of the Discovery products, most of the endpoints for different entities are almost identical. In some cases, the difference is basically the base URL and some parameters. For some reason, common structs were created to expedite the development of the CLI. They implement methods that many of Discovery's components and entities need.

### Client
This struct contains a Resty client to do requests to Discovery's APIs. It also stores an API key for authentication. It has the following related methods:
- newClient(URL, API key): Creates a new client.
- newSubClient(client, path): Creates a new subclient, which has the same URL as the client, but with the path at the end.
- client.execute(method, path, Request Options): This method executes a request to Discovery's APIs, based on the URL configured in the client and the path parameter. This function returns the response's body as a byte array. The request options follow the [functional options pattern](https://www.sohamkamani.com/golang/options-pattern/). They are added to the request. The following options are available:
  - WithQueryParameters(params): This option adds query parameters to the request. They are received with a map of strings to an array of strings. This means that a single parameter can have multiple values. An example of these query parameters is `map[string][]string{"query": {"What is Pureinsights Discovery"}, "items": {"item1", "item2", "item3"}}`.
  - WithFile(path): This option adds a file to the request. It needs to be able to find the path with the received path.
  - WithJSONBody(body): This option sets the body as the received JSON body, which must be a valid JSON string. It also sets the Content Type to `application/json`.
- execute(client, method, path, Request Options): This function is essentially the same as client.execute(), but returns a parsed response as a JSON Object, which in this case is a [gjson.Result](https://github.com/tidwall/gjson#result-type) type.

### Getter
This struct performs all the GET operations. It receives a client field by composition, so it can access the API Key and Resty Client fields. The Resty client must have its URL pointing to the entity's endpoint. It has the following methods:
- Get(id): This function receives a UUID to get the entity referenced by it. It returns the result as gjson.Result object.
- GetAll(): This function retrieves every entity from the endpoint. It returns the results as an array of gjson.Result objects. It automatically iterates through every page of the results to get all entities.

### CRUD
This struct creates, reads, updates, and deletes entities. It receives a getter field by composition, which means it can access the Get() and GetAll(), as well as the fields and methods from the Client struct. The Resty client needs to have its URL set to the URL of the entity's endpoint. It has the following methods:
- Create(Configuration): This function receives the configuration of the entity as a JSON object and executes the request to create the entity in the Discovery component.
- Update(id, configuration): This function receives a UUID and the configuration JSON to update an entity in Discovery.
- Delete(id): This function deletes the entity identified by the received UUID.

### Cloner
This struct is used to clone entities. It has a client field by composition. It adds the `/clone` path to its client's URL to duplicate entities. It has the following method:
- Clone(id, params): This function clones the entity with the received ID, which must be a valid UUID. It adds the parameters it receives to the request. These can be the name of the cloned entity and the depth of the cloning operation (like shallow or deep copying).

### Summarizer
This struct is used to get the summary of ingestion seeds, like the summary of jobs from a seed execution, summary of records from a seed execution, or the summary of records from a seed. It has a client field by composition. To obtain the three mentioned summaries, three different summarizers need to be made, each with its client pointing to the correct endpoint in its URL. The struct has one method:
- Summarize(): This function adds the `/summary` path to the client's URL and executes the request. It then returns the result of the operation.

### Enabler
This struct can be used to enable and disable entities, like Queryflow endpoints and Ingestion seed schedules. It has a client field by composition. It has two methods:
- Enable(id): This function receives a valid UUID and enables the entity referenced by it.
- Disable(id): This function receives a valid UUID and disables the entity referenced by it.  

### BackupRestore
This struct allows for exporting and importing entities from Discovery's components. It has a client field by composition. Its Resty Client needs to be pointing to the Discovery component's base URL. It has two methods:
- Export(): This function calls the `/export` endpoint. It returns the result of the endpoint in bytes, which should be written to a ZIP file so that it can be restored later.
- Import(On Conflict, file): This function calls the `/import` endpoint. The `onConflict` parameter should be one of the constants that represent the conflict resolution strategies. `OnConflictIgnore` sets the strategy to `IGNORE`, which ignores the entities that already exist. `OnConflictFail` sets the strategy to `FAIL`, which fails if entities already exist. `OnConflictUpdate` sets the strategy to `UPDATE`, which updates the entities that already exist. The file it receives must be the path to a file that exists.

### Error
This struct is used as the Errors that the CLI returns. It has a Status, an integer, and a Body, a JSON object (gjson.Result). It has one method:
- Error(): This method serves to fulfill Go's error interface. It returns a string with the error's information.

## Core Client
Discovery has a core client struct. Its fields are:
- Url: The URL of Discovery's Core component. The URL should contain the URL up to the version. For example, http://localhost:8080/v2. 
- ApiKey: The API key needed to authenticate to Discovery's Core.  

The core client can create subclients that handle the Core's functions. These are the following:

### SecretsClient
The secretsClient manages secrets. It is a struct with an embedded CRUD struct. It has access to the execute(), Create(), Get(), GetAll(), Update(), and Delete() methods. Creating a secretsClient can be done with core.Secrets() or newSecretsClient(coreClient).

### CredentialsClient
The credentialsClient manages credentials. Its struct has embedded CRUD and Cloner structs. It has access to the Create(), Get(), GetAll(), Update(), Delete(), and Clone() methods. Creating a credentialsClient can be done with core.Credentials() or newCredentialsClient(coreClient).

### ServersClient
The serversClient manages servers. It is a struct with embedded CRUD and Cloner structs. It has access to the execute(), Create(), Get(), GetAll(), Update(), Delete(), and Clone(). Aditionally, it has a Ping() method to verify if the connection to the server was successful and the server is reachable. Creating a serversClient can be done with core.Servers() or newServersClient(coreClient).

### FilesClient
The filesClient manages Discovery's files. It has an embedded Client struct, so it can access the Client's execute() method. It has an Upload() method to send files to Discovery, Retrieve() to get a file's data in a byte array, List() to get the keys of all of the files stored in Discovery, and Delete() to remove a file. Creating a filesClient can be done with core.Files() or newServersClient(coreClient).

## BackupRestore
The backupRestore struct imports and exports entities. Its Export() method obtains the data of all of the entities, which can later be saved to a ZIP file. The Import() method restores the entities described in the sent file. If there are conflicts, Discovery can be set to ignore them, fail, or update them. Creating a backupRestore can be done with core.BackupRestore().

### LabelsClient
The labelsClient is a struct that manages labels. It has an embedded CRUD struct. It has access to the execute(), Create(), Get(), GetAll(), Update(), and Delete() methods. Creating a labelsClient can be done with core.Labels() or newLabelsClient(coreClient).

### MaintenanceClient
The maintenanceClient struct has an Client struct. It has access to the Client's execute() method. It has a Log() method that changes the log level of a component. It can also change the level of a specific logger inside a component. Creating a maintenanceClient can be done with core.Maintenance() or newMaintenanceClient(coreClient).
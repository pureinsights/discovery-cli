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
| Delete | DELETE | `{URL}/{UUID}`  | | `application/json` | Deletes the entity identified by the received UUID.

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
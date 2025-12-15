# Pureinsights Discovery Platform: Command Line Interface Documentation

## Installation

TODO

## Getting started

TODO

## Documentation

### Discovery

`discovery` is the Discovery CLI's root command. This is the command used to run the CLI. It contains all of the other subcommands.

Usage: `discovery [command]`

Flags:
 
`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Print Discovery's help
discovery -h
```

#### Config
`config` is the main command used to interact with Discovery's configuration for a profile. This command by itself asks the user to save Discovery's configuration for the given profile. The command prints the property to be modified along with its current value. If the property currently being shown is sensitive, its value is obfuscated. To keep the current value, the user must press "Enter" without any text, and to set the value as empty, a sole whitespace must be inputted. 

Usage: `discovery config [subcommand] [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Ask the user for the configuration of profile "cn".
discovery -p cn config
Editing profile "cn". Press Enter to keep the value shown, type a single space to set empty.

Core URL [http://discovery.core.cn]: https://discovery.core.cn
Core API Key [*************.core.cn]: 
Ingestion URL [http://discovery.ingestion.cn]:      
Ingestion API Key [****************gestion.cn]: ingestion123
QueryFlow URL [http://discovery.queryflow.cn]: http://localhost:12040 
QueryFlow API Key [****************eryflow.cn]: queryflow213
Staging URL [http://discovery.staging.cn]: 
Staging API Key [***************taging.cn]: 
```

```bash
# Config works without setting the profile. The rest of the command's output is omitted.
discovery config
Editing profile "default". Press Enter to keep the value shown, type a single space to set empty.
```

```bash
# The profile flag can be set after the command. The rest of the command's output is omitted.
discovery config --profile cn
Editing profile "cn". Press Enter to keep the value shown, type a single space to set empty.
```

##### Get
`get` is the command used to obtain Discovery's configuration for a given profile. If the API keys are sensitive, the `sensitive` flag can be set to true in order to obfuscate them before printing them out. If a configuration property was not set, it is not displayed.

Usage: `discovery config get [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-s, --sensitive`::
(Optional, bool) Obfuscates the API Keys if true. Defaults to `true`.

Examples:

```bash
# Print the configuration of the "cn" profile with obfuscated API keys.
discovery config get -p cn
Showing the configuration of profile "cn":

Core URL: "https://discovery.core.cn"
Core API Key: "*************.core.cn"
Ingestion URL: "http://discovery.ingestion.cn"
Ingestion API Key: "********n123"
QueryFlow URL: "http://localhost:12040"
QueryFlow API Key: "********w213"
Staging URL: "http://discovery.staging.cn"
Staging API Key: "***************taging.cn"
```

```bash
# Print the configuration of the "default" profile.
discovery config get -s
Showing the configuration of profile "default":

Core URL: "http://localhost:12010"
Core API Key: ""
Ingestion URL: "http://localhost:12030"
Ingestion API Key: ""
QueryFlow URL: "http://localhost:12040"
QueryFlow API Key: ""
Staging URL: "http://localhost:12020"
Staging API Key: ""
```

```bash
# Print the configuration of the "cn" profile with unobfuscated API keys.
discovery config get -p cn --sensitive=false
Showing the configuration of profile "cn":

Core URL: "https://discovery.core.cn"
Core API Key: "discovery.key.core.cn"
Ingestion URL: "http://discovery.ingestion.cn"
Ingestion API Key: "ingestion123"
QueryFlow URL: "http://localhost:12040"
QueryFlow API Key: "queryflow213"
Staging URL: "http://discovery.staging.cn"
Staging API Key: "discovery.key.staging.cn"
```

#### Export
`export` is the command used to backup all of Discovery's entities at once. With the `file` flag, the user can send the specific file in which to save the configurations. If not, they will be saved in a zip file in the current directory. The resulting zip file contains three zip files containing the entities of Discovery Core, Ingestion, and QueryFlow. If an export fails, the error is reported in the returned JSON.

Usage: `discovery export [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-f, --file`::
(Optional, string) The file that will contain the exported entities.

Examples:

```bash
# Export the entities using profile "cn".
discovery export -p cn
{"core":{"acknowledged":true},"ingestion":{"acknowledged":true},"queryflow":{"acknowledged":true}}
```

```bash
# Export the entities to a specific file.
# In this example, the Ingestion export failed.
discovery export --file "entities/discovery.zip".
{"core":{"acknowledged":true},"ingestion":{"acknowledged":false,"error":"Get \"http://localhost:12030/v2/export\": dial tcp [::1]:12030: connectex: No connection could be made because the target machine actively refused it."},"queryflow":{"acknowledged":true}}
```
#### Import
`import` is the command used to restore entities to all of Discovery's products at once. With the `file` flag, the user must send the specific file that has the entities' configuration. This file is a compressed zip file that contains the zip files product by the `/export` endpoint in a Discovery product. It should have at most three zip files: one for Core, one for Ingestion, and a final one for QueryFlow. The export file for a Discovery product has the format `productName-*`. For example, the Core can be called `core-export-20251112T1629.zip` and the one for Ingestion can be called `ingestion-export-20251110T1607.zip`. The sent file does not need to contain the export files for all of Discovery's products. This command can restore entities to one, two, or all products. With the `on-conflict` flag, the user can send the conflict resolution strategy in case there are duplicate entities.

Usage: `discovery import [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-f, --file`::
(Required, string) The file that contains the files with the exported entities of the Discovery products.

`--on-conflict`::
(Optional, string) Sets the conflict resolution strategy when importing entities with the same id. The default value is "FAIL".

Examples:

```bash
# Import the entities to Discovery Core and Ingestion using profile "cn" and ignore conflict resolution strategy.
# The rest of the command's output is omitted.
discovery import -p cn --file "entities/discovery.zip" --on-conflict IGNORE
{
  "core": {
    "Credential": [
      {
        "id": "6e2f1c2a-9885-4263-8945-38b0cda4b6d3",
        "status": 204
      },
      {
        "id": "721997cd-b16f-4acb-93cf-b44a959dbcf2",
        "status": 204
      }
    ],
    "Server": [
      {
        "id": "6817ccf5-b4bc-4f97-82f5-c8016d26f2fb",
        "status": 204
      },
      {
        "id": "f7a65744-a3b1-4655-b472-c612bb490ff9",
        "status": 204
      }
    ]
  },
  "ingestion": {
    "Pipeline": [
      {
        "id": "128b1127-0ea0-4aa5-9a4e-9160285d2f61",
        "status": 204
      }
    ],
    "Processor": [
      {
        "id": "11de1d9b-d037-4d27-8304-37b62e79d044",
        "status": 204
      }
    ],
    "Seed": [
      {
        "id": "bb8d13c6-73b5-47a1-b0fb-06a141e32309",
        "status": 204
      }
    ],
    "SeedSchedule": []
  }
}
```

#### Core
`core` is the main command used to interact with Discovery's Core. 

Usage: `discovery core [subcommand] [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

##### Config
`config` is the command used to interact with Discovery Core's configuration for a profile. This command by itself asks the user to save Discovery Core's configuration for the given profile. The command prints the property to be modified along with its current value. If the property currently being shown is sensitive, its value is obfuscated. To keep the current value, the user must press \"Enter\" without any text, and to set the value as empty, a sole whitespace must be inputted.

Usage: `discovery core config [subcommand] [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Ask the user for the configuration of profile "cn"
discovery core config -p cn
Editing profile "cn". Press Enter to keep the value shown, type a single space to set empty.

Core URL [http://discovery.core.cn]: https://discovery.core.cn
Core API Key [*************.core.cn]: 
```

```bash
# Config works without the profile. The rest of the command's output is omitted.
discovery core config
Editing profile "default". Press Enter to keep the value shown, type a single space to set empty.
```

###### Get
`get` is the command used to obtain Discovery Core's configuration for a given profile. If the API keys are sensitive, the `sensitive` flag can be set to true in order to obfuscate them before printing them out. If a configuration property was not set, it is not displayed.

Usage: `discovery core config get [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-s, --sensitive`::
(Optional, bool) Obfuscates the API Keys if true. Defaults to `true`.

Examples: 

```bash
# Print the configuration of the "cn" profile with obfuscated API keys.
discovery core config get -p cn
Showing the configuration of profile "cn":

Core URL: "https://discovery.core.cn"
Core API Key: "*************.core.cn"
```

```bash
# Print the configuration of the "default" profile.
discovery core config get -s
Showing the configuration of profile "default":

Core URL: "http://localhost:12010"
Core API Key: ""
```

```bash
# Print the configuration of the "cn" profile with unobfuscated API keys.
discovery core config get -p cn --sensitive=false
Showing the configuration of profile "cn":

Core URL: "https://discovery.core.cn"
Core API Key: "discovery.key.core.cn"
```

##### Export
`export` is the command used to backup Discovery Core's entities. With the `file` flag, the user can send the specific file in which to save the configurations. If not, they will be saved in a zip file in the current directory.

Usage: `discovery core export [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-f, --file`::
(Optional, string) The file that will contain the exported entities.

Examples:

```bash
# Export the entities using profile "cn".
discovery core export -p cn
{"acknowledged":true}
```

```bash
# Export the entities to a specific file.
discovery core export -p cn --file "entities/core.zip"
{"acknowledged":true}
```

##### Import
`import` is the command used to restore Discovery Core's entities. With the `file` flag, the user must send the specific file that has the entities' configuration. With the `on-conflict` flag, the user can send the conflict resolution strategy in case there are duplicate entities.

Usage: `discovery core import [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-f, --file`::
(Required, string) The file that contains the configurations of the entities.

`--on-conflict`::
(Optional, string) Sets the conflict resolution strategy when importing entities with the same id. The default value is "FAIL".

Examples:

```bash
# Import the entities using profile "cn" and update conflict resolution strategy.
# The rest of the command's output is omitted.
discovery core import -p cn --file "entities/core.zip" --on-conflict UPDATE
{
  "Credential": [
    {
      "id": "3b32e410-2f33-412d-9fb8-17970131921c",
      "status": 200
    },
    {
      "id": "458d245a-6ed2-4c2b-a73f-5540d550a479",
      "status": 200
    },
    {
      "id": "46cb4fff-28be-4901-b059-1dd618e74ee4",
      "status": 200
    },
    ...
```

##### Label
`label` is the command used to manage labels in Discovery Core. This command contains various subcommands used to create, read, update, and delete.

Usage: `discovery core label [subcommand] [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

###### Get
`get` is the command used to obtain Discovery Core's labels. The user can send a UUID to get a specific label. If no UUID is given, then the command retrieves every label. The optional argument must be a UUID. This command does not support filters or referencing an entity by name.

Usage: `discovery core label get [flags] [<uuid>]`

Arguments:
`uuid`::
(Optional, string) The UUID of the label that will be retrieved.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Get a label by id
discovery core label get 3d51beef-8b90-40aa-84b5-033241dc6239
{"creationTimestamp":"2025-08-27T19:22:06Z","id":"3d51beef-8b90-40aa-84b5-033241dc6239","key":"A","lastUpdatedTimestamp":"2025-08-27T19:22:47Z","value":"B"}
```

```bash
# Get all labels using the configuration in profile "cn"
discovery core label get -p cn
{"creationTimestamp":"2025-10-15T20:28:39Z","id":"5467ab23-7827-4fae-aa78-dfd4800549ee","key":"D","lastUpdatedTimestamp":"2025-10-15T20:28:39Z","value":"F"}
{"creationTimestamp":"2025-10-15T20:25:29Z","id":"7d0cb8c9-6555-4592-9b6c-1f4ed7fca9f4","key":"D","lastUpdatedTimestamp":"2025-10-15T20:25:29Z","value":"D"}
{"creationTimestamp":"2025-09-29T17:00:47Z","id":"a77fed6a-021e-440b-bb32-91e22ea31598","key":"A","lastUpdatedTimestamp":"2025-09-29T17:00:47Z","value":"A"}
```

###### Store
`store` is the command used to create and update Discovery Core's labels. With the `data` flag, the user can send a single JSON configuration or an array to upsert multiple labels. With the `file` flag, the user can also send the path of a file that contains the JSON configurations. The `data` and `file` flags are required, but mutually exclusive.

Usage: `discovery core label store [flags]`

Flags:

`-d, --data`::
(Required, string) Set the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `file` flag.

`-f, --file`::
(Required, string) Set the path of the file that contains the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `data` flag.

`--abort-on-error`::
(Optional, bool) Aborts the operation when an error occurs. The default value is `false`.

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Store a label with the JSON configuration in a file
discovery core label store --file "labeljsonfile.json"
{"creationTimestamp":"2025-08-27T19:22:06Z","id":"3d51beef-8b90-40aa-84b5-033241dc6239","key":"label1","lastUpdatedTimestamp":"2025-10-29T22:41:37Z","value":"value1"}
{"code":1003,"messages":["Entity not found: 3d51beef-8b90-40aa-84b5-033241dc6230"],"status":404,"timestamp":"2025-10-30T00:05:35.995533500Z"}
{"creationTimestamp":"2025-10-30T00:05:36.004363Z","id":"4967bc7b-ed89-4843-ab0f-1fd73daad30d","key":"label3","lastUpdatedTimestamp":"2025-10-30T00:05:36.004363Z","value":"value3"}
```

```bash
# Store a label with the JSON configuration in the data flag
discovery core label store --data  '[{"key":"label","value":"labelvalue"}]'
{"creationTimestamp":"2025-10-30T00:07:07.244729Z","id":"e7870373-da6d-41af-b5ec-91cfd087ee91","key":"label","lastUpdatedTimestamp":"2025-10-30T00:07:07.244729Z","value":"labelvalue"}
```

###### Delete
`delete` is the command used to delete Discovery Core's labels. The user must send a UUID to delete a specific label. If no UUID is given, then an error is returned. This command does not support referencing an entity by name.

Usage: `discovery core label delete [flags] <uuid>`

Arguments:
`uuid`::
(Required, string) The UUID of the label that will be deleted.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Delete a label by id
discovery core label delete 3d51beef-8b90-40aa-84b5-033241dc6239
{"acknowledged":true}
```

##### Secret
`secret` is the command used to manage secrets in Discovery Core. This command contains various subcommands used to create, read, update, and delete.

Usage: `discovery core secret [subcommand] [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

###### Get
`get` is the command used to obtain Discovery Core's secrets. The user can send a UUID to get a specific secret. If no UUID is given, then the command retrieves every secret. The optional argument must be a UUID. This command does not support filters or referencing an entity by name.

Usage: `discovery core secret get [flags] [<uuid>]`

Arguments:
`uuid`::
(Optional, string) The UUID of the secret that will be retrieved.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Get a secret by id
discovery core secret get 81ca1ac6-3058-4ecd-a292-e439827a675a
{"active":true,"creationTimestamp":"2025-08-26T21:56:50Z","id":"81ca1ac6-3058-4ecd-a292-e439827a675a","labels":[],"lastUpdatedTimestamp":"2025-08-26T21:56:50Z","name":"openai-secret"}
```

```bash
# Get all secrets using the configuration in profile "cn"
discovery core secret get -p cn
{"active":true,"creationTimestamp":"2025-08-26T21:56:50Z","id":"81ca1ac6-3058-4ecd-a292-e439827a675a","labels":[],"lastUpdatedTimestamp":"2025-08-26T21:56:50Z","name":"openai-secret"}
{"active":true,"creationTimestamp":"2025-08-14T18:01:59Z","id":"cfa0ef51-1fd9-47e2-8fdb-262ac9712781","labels":[],"lastUpdatedTimestamp":"2025-08-14T18:01:59Z","name":"mongo-secret"}
```

###### Store
`store` is the command used to create and update Discovery Core's secrets. With the `data` flag, the user can send a single JSON configuration or an array to upsert multiple secrets. With the `file` flag, the user can also send the path of a file that contains the JSON configurations. The `data` and `file` flags are required, but mutually exclusive.

Usage: `discovery core secret store [flags]`

Flags:

`-d, --data`::
(Required, string) Set the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `file` flag.

`-f, --file`::
(Required, string) Set the path of the file that contains the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `data` flag.

`--abort-on-error`::
(Optional, bool) Aborts the operation when an error occurs. The default value is `false`.

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Store a secret with the JSON configuration in a file
discovery core secret store --file "secretjsonfile.json"
{"active":true,"creationTimestamp":"2025-10-30T15:09:16Z","id":"b8bd5ec3-8f60-4502-b25e-8f6d36c98410","lastUpdatedTimestamp":"2025-10-30T15:15:22.738365Z","name":"openai-secret"}
{"code":1003,"messages":["Entity not found: b8bd5ec3-8f60-4502-b25e-8f6d36c98415"],"status":404,"timestamp":"2025-10-30T15:15:22.778371Z"}
{"active":true,"creationTimestamp":"2025-10-30T15:15:22.801771Z","id":"c9731417-38c9-4a65-8bbc-78c5f59b9cbb","lastUpdatedTimestamp":"2025-10-30T15:15:22.801771Z","name":"mongo-user"}
```

```bash
# Store a secret with the JSON configuration in the data flag
discovery core secret store --data  '{"name":"my-secret","active":true,"id":"b8bd5ec3-8f60-4502-b25e-8f6d36c98410","content":{"apiKey":"apiKey"}}'
{"active":true,"creationTimestamp":"2025-10-30T15:09:16Z","id":"b8bd5ec3-8f60-4502-b25e-8f6d36c98410","lastUpdatedTimestamp":"2025-10-30T15:43:52.496829Z","name":"my-secret"}
```

###### Delete
`delete` is the command used to delete Discovery Core's secrets. The user must send a UUID to delete a specific secret. If no UUID is given, then an error is returned. This command does not support referencing an entity by name.

Usage: `discovery core secret delete [flags] <uuid>`

Arguments:
`uuid`::
(Required, string) The UUID of the secret that will be deleted.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Delete a secret by id
discovery core secret delete 3d51beef-8b90-40aa-84b5-033241dc6239
{"acknowledged":true}
```

##### Credential
`credential` is the command used to manage credentials in Discovery Core. This command contains various subcommands used to create, read, update, and delete.

Usage: `discovery core credential [subcommand] [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

###### Get
`get` is the command used to obtain Discovery Core's credentials. The user can send a name or UUID to get a specific credential. If no argument is given, then the command retrieves every credential. The command also supports filters with the flag `--filter` followed by the filter in the format `filter=key:value`.

Usage: `discovery core credential get [flags] [<arg>]`

Arguments:
`arg`::
(Optional, string) The name or UUID of the credential that will be retrieved.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-f, --filter`::
(Optional, Array of strings) Add a filter to the search. The available filters are the following:
- Label: The format is `label={key}[:{value}]`, where the value is optional.
- Type: The format is `type={type}`.

Examples:

```bash
# Get a credential by id
discovery core credential get 3b32e410-2f33-412d-9fb8-17970131921c
{"active":true,"creationTimestamp":"2025-10-17T22:37:57Z","id":"3b32e410-2f33-412d-9fb8-17970131921c","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T22:37:57Z","name":"my-credential","secret":"mongo-secret","type":"mongo"}
```

```bash
# Get credential by name
discovery core credential get "my-credential"
{"active":true,"creationTimestamp":"2025-11-20T00:08:14Z","id":"9be0e625-a510-46c5-8130-438823f849c2","labels":[],"lastUpdatedTimestamp":"2025-11-20T00:08:14Z","name":"my-credential","secret":"my-secret","type":"openai"}
```

```bash
# Get credentials using filters
discovery core credential get --filter label=A:A --filter type=mongo
{"active":true,"creationTimestamp":"2025-10-17T15:33:58Z","id":"8c243a1d-9384-421d-8f99-4ef28d4e0ab0","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T15:33:58Z","name":"my-credential","type":"mongo"}
{"active":true,"creationTimestamp":"2025-10-17T22:37:53Z","id":"4957145b-6192-4862-a5da-e97853974e9f","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T22:37:53Z","name":"my-credential-2","type":"mongo"}
```

```bash
# Get all credentials using the configuration in profile "cn"
discovery core credential get -p cn
{"active":true,"creationTimestamp":"2025-10-17T22:37:57Z","id":"3b32e410-2f33-412d-9fb8-17970131921c","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T22:37:57Z","name":"my-credential","type":"mongo"}
{"active":true,"creationTimestamp":"2025-10-17T22:40:15Z","id":"458d245a-6ed2-4c2b-a73f-5540d550a479","labels":[{"key":"A","value":"B"}],"lastUpdatedTimestamp":"2025-10-17T22:40:15Z","name":"openai-credential","type":"openai"}
```

###### Store
`store` is the command used to create and update Discovery Core's credentials. With the `data` flag, the user can send a single JSON configuration or an array to upsert multiple credentials. With the `file` flag, the user can also send the path of a file that contains the JSON configurations. The `data` and `file` flags are required, but mutually exclusive.

Usage: `discovery core credential store [flags]`

Flags:

`-d, --data`::
(Required, string) Set the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `file` flag.

`-f, --file`::
(Required, string) Set the path of the file that contains the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `data` flag.

`--abort-on-error`::
(Optional, bool) Aborts the operation when an error occurs. The default value is `false`.

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Store a credential with the JSON configuration in a file
discovery core credential store --file "credentialjsonfile.json"
{"active":true,"creationTimestamp":"2025-10-17T22:37:57Z","id":"3b32e410-2f33-412d-9fb8-17970131921c","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T22:37:57Z","name":"my-credential-1","secret":"my-secret","type":"mongo"}
{"code":1003,"messages":["Entity not found: 3b32e410-2f33-412d-9fb8-17970131921d"],"status":404,"timestamp":"2025-10-30T16:50:38.250661200Z"}
{"active":true,"creationTimestamp":"2025-10-30T16:50:38.262086Z","id":"5b76ae0d-f383-47e5-be6f-90e9046092cd","labels":[{"key":"A","value":"B"}],"lastUpdatedTimestamp":"2025-10-30T16:50:38.262086Z","name":"my-credential-2","secret":"my-secret","type":"mongo"}
```

```bash
# Store a credential with the JSON configuration in the data flag
discovery core credential store --data '{"type":"mongo","name":"my-credential","labels":[{"key":"A","value":"A"}],"active":true,"id":"3b32e410-2f33-412d-9fb8-17970131921c","creationTimestamp":"2025-10-17T22:37:57Z","lastUpdatedTimestamp":"2025-10-17T22:37:57Z","secret":"my-secret"}'
{"active":true,"creationTimestamp":"2025-10-17T22:37:57Z","id":"3b32e410-2f33-412d-9fb8-17970131921c","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T22:37:57Z","name":"my-credential","secret":"my-secret","type":"mongo"}
```

###### Delete
`delete` is the command used to delete Discovery Core's credentials. The user must send a name or UUID to delete a specific credential.

Usage: `discovery core credential delete [flags] <arg>`

Arguments:
`arg`::
(Required, string) The name or UUID of the credential that will be deleted.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

```bash
# Delete a credential by id
discovery core credential delete 3d51beef-8b90-40aa-84b5-033241dc6239
{"acknowledged":true}
```

```bash
# Delete a credential by name
discovery core credential delete my-credential
{"acknowledged":true}
```

##### Server
`server` is the command used to manage servers in Discovery Core. This command contains various subcommands used to create, read, update, and delete.

Usage: `discovery core server [subcommand] [flags]

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

###### Get
`get` is the command used to obtain Discovery Core's servers. The user can send a name or UUID to get a specific server. If no argument is given, then the command retrieves every server. The command also supports filters with the flag `--filter` followed by the filter in the format `filter=key:value`.

Usage: `discovery core server get [flags] [<arg>]`

Arguments:
`arg`::
(Optional, string) The name or UUID of the server that will be retrieved.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-f, --filter`::
(Optional, Array of strings) Add a filter to the search. The available filters are the following:
- Label: The format is `label={key}[:{value}]`, where the value is optional.
- Type: The format is `type={type}`.

Examples:

```bash
# Get a server by id
discovery core server get 21029da3-041c-43b5-a67e-870251f2f6a6
{"active":true,"config":{"connection":{"connectTimeout":"1m","readTimeout":"30s"},"credentialId":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","servers":["mongodb+srv://cluster0.dleud.mongodb.net/"]},"creationTimestamp":"2025-09-29T15:50:19Z","id":"21029da3-041c-43b5-a67e-870251f2f6a6","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-09-29T15:50:19Z","name":"my-server","type":"mongo"}
```

```bash
# Get server by name
discovery core server get "my-server"
{"active":true,"config":{"connection":{"connectTimeout":"1m","readTimeout":"30s"},"credentialId":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","servers":["mongodb+srv://cluster0.dleud.mongodb.net/"]},"creationTimestamp":"2025-09-29T15:50:19Z","id":"21029da3-041c-43b5-a67e-870251f2f6a6","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-09-29T15:50:19Z","name":"my-server","type":"mongo"}
```

```bash
# Get servers using filters
discovery core server get --filter label=A:A -f type=mongo
{"active":true,"creationTimestamp":"2025-09-29T15:50:19Z","id":"21029da3-041c-43b5-a67e-870251f2f6a6","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-09-29T15:50:19Z","name":"my-server","type":"mongo"}
{"active":true,"creationTimestamp":"2025-09-29T15:50:21Z","id":"a798cd5b-aa7a-4fc5-9292-1de6fe8e8b7f","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-09-29T15:50:21Z","name":"my-server-2","type":"mongo"}
```

```bash
# Get all servers using the configuration in profile "cn"
discovery core server get -p cn
{"active":true,"creationTimestamp":"2025-09-29T15:50:37Z","id":"025347a7-e2bd-4ba1-880f-db3e51319abb","labels":[],"lastUpdatedTimestamp":"2025-09-29T15:50:37Z","name":"MongoDB Atlas server","type":"mongo"}
{"active":true,"creationTimestamp":"2025-10-15T20:26:27Z","id":"192c3793-600a-4366-9778-7d80a0df07ce","labels":[{"key":"E","value":"G"},{"key":"H","value":"F"},{"key":"D","value":"D"}],"lastUpdatedTimestamp":"2025-10-15T20:26:27Z","name":"OpenAI Server","type":"openai"}
```

###### Store
`store` is the command used to create and update Discovery Core's servers. With the `data` flag, the user can send a single JSON configuration or an array to upsert multiple servers. With the `file` flag, the user can also send the path of a file that contains the JSON configurations. The `data` and `file` flags are required, but mutually exclusive.

Usage: `discovery core server store [flags]`

Flags:

`-d, --data`::
(Required, string) Set the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `file` flag.

`-f, --file`::
(Required, string) Set the path of the file that contains the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `data` flag.

`--abort-on-error`::
(Optional, bool) Aborts the operation when an error occurs. The default value is `false`.

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Store a server with the JSON configuration in a file
discovery core server store --file "serverjsonfile.json"
{"active":true,"config":{"connection":{"connectTimeout":"1m","readTimeout":"30s"},"credentialId":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","servers":["mongodb+srv://cluster0.dleud.mongodb.net/"]},"creationTimestamp":"2025-09-29T15:50:26Z","id":"2b839453-ddad-4ced-8e13-2c7860af60a7","labels":[],"lastUpdatedTimestamp":"2025-09-29T15:50:26Z","name":"my-server","type":"mongo"}       
{"code":1003,"messages":["Entity not found: 2b839453-ddad-4ced-8e13-2c7860af60a8"],"status":404,"timestamp":"2025-10-30T17:45:48.176913700Z"}
{"active":true,"config":{"connection":{"connectTimeout":"1m","readTimeout":"30s"},"credentialId":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","servers":["mongodb+srv://cluster0.dleud.mongodb.net/"]},"creationTimestamp":"2025-10-30T17:45:48.184774Z","id":"152e1175-e54d-4de6-90b9-388d45f8256e","labels":[],"lastUpdatedTimestamp":"2025-10-30T17:45:48.184774Z","name":"my-server-2","type":"mongo"}
```

```bash
# Store a server with the JSON configuration in the data flag
discovery core server store --data '{"type":"mongo","name":"my-server","labels":[],"active":true,"id":"2b839453-ddad-4ced-8e13-2c7860af60a7","creationTimestamp":"2025-09-29T15:50:26Z","lastUpdatedTimestamp":"2025-09-29T15:50:26Z","config":{"servers":["mongodb+srv://cluster0.dleud.mongodb.net/"],"connection":{"readTimeout":"30s","connectTimeout":"1m"},"credentialId":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c"}}'
{"active":true,"config":{"connection":{"connectTimeout":"1m","readTimeout":"30s"},"credentialId":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","servers":["mongodb+srv://cluster0.dleud.mongodb.net/"]},"creationTimestamp":"2025-09-29T15:50:26Z","id":"2b839453-ddad-4ced-8e13-2c7860af60a7","labels":[],"lastUpdatedTimestamp":"2025-09-29T15:50:26Z","name":"my-server","type":"mongo"}
```

###### Delete
`delete` is the command used to delete Discovery Core's servers. The user must send a name or UUID to delete a specific server.

Usage: `discovery core server delete [flags] <arg>`

Arguments:
`arg`::
(Required, string) The name or UUID of the server that will be deleted.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

```bash
# Delete a server by id
discovery core server delete 3d51beef-8b90-40aa-84b5-033241dc6239
{"acknowledged":true}
```

```bash
# Delete a server by name
discovery core server delete my-server
{"acknowledged":true}
```

#### Ingestion
`ingestion` is the main command used to interact with Discovery's Ingestion. 

Usage: `discovery ingestion [subcommand] [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

##### Config
`config` is the command used to interact with Discovery Ingestion's configuration for a profile. This command by itself asks the user to save Discovery Ingestion's configuration for the given profile. The command prints the property to be modified along with its current value. If the property currently being shown is sensitive, its value is obfuscated. To keep the current value, the user must press \"Enter\" without any text, and to set the value as empty, a sole whitespace must be inputted.

Usage: `discovery ingestion config [subcommand] [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Ask the user for the configuration of profile "cn"
discovery ingestion config -p cn
Editing profile "cn". Press Enter to keep the value shown, type a single space to set empty.

Ingestion URL [http://discovery.ingestion.cn]: https://discovery.ingestion.cn
Ingestion API Key [*************.ingestion.cn]: 
```

```bash
# Config works without the profile. The rest of the command's output is omitted.
discovery ingestion config
Editing profile "default". Press Enter to keep the value shown, type a single space to set empty.
```

###### Get
`get` is the command used to obtain Discovery Ingestion's configuration for a given profile. If the API keys are sensitive, the `sensitive` flag can be set to true in order to obfuscate them before printing them out. If a configuration property was not set, it is not displayed.

Usage: `discovery ingestion config get [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-s, --sensitive`::
(Optional, bool) Obfuscates the API Keys if true. Defaults to `true`.

Examples: 

```bash
# Print the configuration of the "cn" profile with obfuscated API keys.
discovery ingestion config get -p cn
Showing the configuration of profile "cn":

Ingestion URL: "https://discovery.ingestion.cn"
Ingestion API Key: "*************.ingestion.cn"
```

```bash
# Print the configuration of the "default" profile.
discovery ingestion config get -s
Showing the configuration of profile "default":

Ingestion URL: "http://localhost:12010"
Ingestion API Key: ""
```

```bash
# Print the configuration of the "cn" profile with unobfuscated API keys.
discovery ingestion config get -p cn --sensitive=false
Showing the configuration of profile "cn":

Ingestion URL: "https://discovery.ingestion.cn"
Ingestion API Key: "discovery.key.ingestion.cn"
```

##### Export
`export` is the command used to backup Discovery Ingestion's entities. With the `file` flag, the user can send the specific file in which to save the configurations. If not, they will be saved in a zip file in the current directory.

Usage: `discovery ingestion export [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-f, --file`::
(Optional, string) The file that will contain the exported entities.

Examples:

```bash
# Export the entities using profile "cn".
discovery ingestion export -p cn
{"acknowledged":true}
```

```bash
# Export the entities to a specific file
discovery ingestion export --file "entities/ingestion.zip"
{"acknowledged":true}
```

##### Import
`import` is the command used to restore Discovery Ingestion's entities. With the `file` flag, the user must send the specific file that has the entities' configuration. With the `on-conflict` flag, the user can send the conflict resolution strategy in case there are duplicate entities.

Usage: `discovery Ingestion import [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-f, --file`::
(Required, string) The file that contains the configurations of the entities.

`--on-conflict`::
(Optional, string) Sets the conflict resolution strategy when importing entities with the same id. The default value is "FAIL".

Examples:

```bash
# Import the entities using profile "cn" and ignore conflict resolution strategy.
# The rest of the command's output is omitted.
discovery ingestion import -p cn --file "entities/ingestion.zip" --on-conflict IGNORE
{
  "Pipeline": [
    {
      "id": "0d3f476d-9003-4fc8-b9a9-8ba6ebf9445b",
      "status": 204
    },
    {
      "id": "25012a20-fe60-4ad6-a05c-9abcbfc1dfb1",
      "status": 204
    },
    {
      "id": "36f8ce72-f23d-4768-91e8-58693ff1b272",
      "status": 204
    },
    ...
```

##### Processor
`processor` is the command used to manage processors in Discovery Ingestion. This command contains various subcommands used to create, read, update, and delete.

Usage: `discovery ingestion processor [subcommand] [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

###### Get
`get` is the command used to obtain Discovery Ingestion's processors. The user can send a name or UUID to get a specific processor. If no argument is given, then the command retrieves every processor. The command also supports filters with the flag `--filter` followed by the filter in the format `filter=key:value`.

Usage: `discovery ingestion processor get [flags] [<arg>]`

Arguments:
`arg`::
(Optional, string) The name or UUID of the processor that will be retrieved.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-f, --filter`::
(Optional, Array of strings) Add a filter to the search. The available filters are the following:
- Label: The format is `label={key}[:{value}]`, where the value is optional.
- Type: The format is `type={type}`.
- 
Examples:

```bash
# Get a processor by id
discovery ingestion processor get 90675678-fc9f-47ec-8bab-89969dc204f0
{"active":true,"config":{"action":"hydrate","collection":"blogs","data":{"author":"#{ data('/author') }","header":"#{ data('/header') }","link":"#{ data('/reference') }"},"database":"pureinsights"},"creationTimestamp":"2025-10-30T20:07:43Z","id":"90675678-fc9f-47ec-8bab-89969dc204f0","labels":[],"lastUpdatedTimestamp":"2025-10-30T20:07:43Z","name":"my-processor","server":{"credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","id":"f6950327-3175-4a98-a570-658df852424a"},"type":"mongo"}
```

```bash
# Get processor by name
discovery ingestion processor get "my-processor"
{"active":true,"config":{"action":"select","charset":"UTF-8","file":"#{data('/file')}","selectors":{"section0":{"mode":"HTML","selector":".sect0"},"section1":{"mode":"HTML","selector":".sect1"}}},"creationTimestamp":"2025-11-17T22:38:26Z","id":"56ace252-4731-4428-84b8-7cd13bf059d3","labels":[],"lastUpdatedTimestamp":"2025-11-17T22:38:26Z","name":"my-processor","type":"html"}
```

```bash
# Get processors using filters
discovery ingestion processor get --filter label=A:A -f type=mongo
{"active":true,"creationTimestamp":"2025-10-31T21:50:54Z","id":"89103d32-6007-489a-8e25-dc9a6001f8e8","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T21:50:54Z","name":"my-processor","type":"mongo"}
```

```bash
# Get all processors using the configuration in profile "cn"
discovery ingestion processor get -p cn
{"active":true,"creationTimestamp":"2025-08-21T21:52:02Z","id":"516d4a8a-e8ae-488c-9e37-d5746a907454","labels":[],"lastUpdatedTimestamp":"2025-08-21T21:52:02Z","name":"my-processor","type":"template"}
{"active":true,"creationTimestamp":"2025-10-30T20:07:43Z","id":"7569f1a5-521e-4d8c-94d1-9f53ad065320","labels":[],"lastUpdatedTimestamp":"2025-10-30T20:07:43Z","name":"my-processor-2","type":"mongo"}
{"active":true,"creationTimestamp":"2025-08-21T21:52:02Z","id":"7b192ea1-ac43-439b-9396-5e022f81f2cb","labels":[],"lastUpdatedTimestamp":"2025-08-21T21:52:02Z","name":"my-processor-3","type":"openai"}
```

###### Store
`store` is the command used to create and update Discovery Ingestion's processors. With the `data` flag, the user can send a single JSON configuration or an array to upsert multiple processors. With the `file` flag, the user can also send the path of a file that contains the JSON configurations. The `data` and `file` flags are required, but mutually exclusive.

Usage: `discovery ingestion processor store [flags]`

Flags:

`-d, --data`::
(Required, string) Set the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `file` flag.

`-f, --file`::
(Required, string) Set the path of the file that contains the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `data` flag.

`--abort-on-error`::
(Optional, bool) Aborts the operation when an error occurs. The default value is `false`.

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Store a processor with the JSON configuration in a file
discovery ingestion processor store --file "ingestionprocessorjsonfile.json"
{"active":true,"config":{"action":"hydrate","collection":"blogs","data":{"author":"#{ data('/author') }","header":"#{ data('/header') }","link":"#{ data('/reference') }"},"database":"pureinsights"},"creationTimestamp":"2025-10-30T20:07:44Z","id":"e9c4173f-6906-43a8-b3ca-7319d3d24754","labels":[],"lastUpdatedTimestamp":"2025-10-30T20:07:44Z","name":"my-processor","server":{"credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","id":"f6950327-3175-4a98-a570-658df852424a"},"type":"mongo"}
{"code":1003,"messages":["Entity not found: e9c4173f-6906-43a8-b3ca-7319d3d24755"],"status":404,"timestamp":"2025-10-30T20:09:29.314467Z"}
{"active":true,"config":{"action":"hydrate","collection":"blogs","data":{"author":"#{ data('/author') }","header":"#{ data('/header') }","link":"#{ data('/reference') }"},"database":"pureinsights"},"creationTimestamp":"2025-10-30T20:09:29.346792Z","id":"aef648d8-171d-479a-a6fd-14ec9b235dc7","labels":[],"lastUpdatedTimestamp":"2025-10-30T20:09:29.346792Z","name":"my-processor-2","server":{"credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","id":"f6950327-3175-4a98-a570-658df852424a"},"type":"mongo"}
```

```bash
# Store a processor with the JSON configuration in the data flag
discovery ingestion processor store --data '{"type":"mongo","name":"my-processor","labels":[],"active":true,"id":"e9c4173f-6906-43a8-b3ca-7319d3d24754","creationTimestamp":"2025-10-30T20:07:43.825231Z","lastUpdatedTimestamp":"2025-10-30T20:07:43.825231Z","config":{"data":{"link":"#{ data('/reference') }","author":"#{ data('/author') }","header":"#{ data('/header') }"},"action":"hydrate","database":"pureinsights","collection":"blogs"},"server":{"id":"f6950327-3175-4a98-a570-658df852424a","credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c"}}'
{"active":true,"config":{"action":"hydrate","collection":"blogs","data":{"author":"#{ data(/author) }","header":"#{ data(/header) }","link":"#{ data(/reference) }"},"database":"pureinsights"},"creationTimestamp":"2025-10-30T20:07:44Z","id":"e9c4173f-6906-43a8-b3ca-7319d3d24754","labels":[],"lastUpdatedTimestamp":"2025-10-30T20:10:23.698799Z","name":"my-processor","server":{"credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","id":"f6950327-3175-4a98-a570-658df852424a"},"type":"mongo"}
```

###### Delete
`delete` is the command used to delete Discovery Ingestion's processors. The user must send a name or UUID to delete a specific processor.

Usage: `discovery ingestion processor delete [flags] <arg>`

Arguments:
`arg`::
(Required, string) The name or UUID of the processor that will be deleted.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

```bash
# Delete a processor by id
discovery ingestion processor delete 83a009d5-5d2f-481c-b8bf-f96d3a35c240
{"acknowledged":true}
```

```bash
# Delete a processor by name
discovery ingestion processor delete "my-processor"
{"acknowledged":true}
```

##### Pipeline
`pipeline` is the command used to manage pipelines in Discovery Ingestion. This command contains various subcommands used to create, read, update, and delete.

Usage: `discovery ingestion pipeline [subcommand] [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

###### Get
`get` is the command used to obtain Discovery Ingestion's pipelines. The user can send a name or UUID to get a specific pipeline. If no argument is given, then the command retrieves every pipeline. The command also supports filters with the flag `--filter` followed by the filter in the format `filter=key:value`.

Usage: `discovery ingestion pipeline get [flags] [<arg>]`

Arguments:
`arg`::
(Optional, string) The name or UUID of the pipeline that will be retrieved.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-f, --filter`::
(Optional, Array of strings) Add a filter to the search. The available filter is the following:
- Label: The format is `label={key}[:{value}]`, where the value is optional.

Examples:

```bash
# Get a pipeline by id
discovery ingestion pipeline get 04536687-f083-4353-8ecc-b7348e14b748
{"active":true,"creationTimestamp":"2025-10-31T22:07:02Z","id":"04536687-f083-4353-8ecc-b7348e14b748","initialState":"ingestionState","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:07:02Z","name":"my-pipeline","recordPolicy":{"errorPolicy":"FAIL","idPolicy":{},"outboundPolicy":{"batchPolicy":{"flushAfter":"PT1M","maxCount":25},"mode":"INLINE","splitPolicy":{"children":{"idPolicy":{},"snapshotPolicy":{}},"source":{"snapshotPolicy":{}}}},"retryPolicy":{"active":true,"maxRetries":3},"timeoutPolicy":{"record":"PT1M"}},"states":{"ingestionState":{"processors":[{"active":true,"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","outputField":"header"},{"active":true,"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8"}],"type":"processor"}}}
```

```bash
# Get pipeline by name
discovery ingestion pipeline get "my-pipeline"
{"active":true,"creationTimestamp":"2025-10-31T22:07:02Z","id":"04536687-f083-4353-8ecc-b7348e14b748","initialState":"ingestionState","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:07:02Z","name":"my-pipeline","recordPolicy":{"errorPolicy":"FAIL","idPolicy":{},"outboundPolicy":{"batchPolicy":{"flushAfter":"PT1M","maxCount":25},"mode":"INLINE","splitPolicy":{"children":{"idPolicy":{},"snapshotPolicy":{}},"source":{"snapshotPolicy":{}}}},"retryPolicy":{"active":true,"maxRetries":3},"timeoutPolicy":{"record":"PT1M"}},"states":{"ingestionState":{"processors":[{"active":true,"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","outputField":"header"},{"active":true,"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8"}],"type":"processor"}}}
```

```bash
# Get pipelines using filters
discovery ingestion pipeline get --filter label=A:A
{"active":true,"creationTimestamp":"2025-10-31T22:06:27Z","id":"8d9560b3-631b-490b-994f-4708d9880e3b","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:06:27Z","name":"Smy-pipeline"}
{"active":true,"creationTimestamp":"2025-10-31T22:07:00Z","id":"e15ca96b-3d42-4ab9-be16-3c6d6713b04e","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:07:00Z","name":"my-pipeline-2"}
```

```bash
# Get all pipelines using the configuration in profile "cn"
discovery ingestion pipeline get -p cn
{"active":true,"creationTimestamp":"2025-10-31T22:07:02Z","id":"04536687-f083-4353-8ecc-b7348e14b748","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:07:02Z","name":"my-pipeline"}
{"active":true,"creationTimestamp":"2025-10-31T19:41:16Z","id":"0d3f476d-9003-4fc8-b9a9-8ba6ebf9445b","labels":[],"lastUpdatedTimestamp":"2025-10-31T19:41:16Z","name":"my-pipeline-2"}
{"active":true,"creationTimestamp":"2025-10-31T22:07:00Z","id":"22b1f0fe-d7c1-476f-a609-1a12ee97655f","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:07:00Z","name":"my-pipeline-3"}
```

###### Store
`store` is the command used to create and update Discovery Ingestion's pipelines. With the `data` flag, the user can send a single JSON configuration or an array to upsert multiple pipelines. With the `file` flag, the user can also send the path of a file that contains the JSON configurations. The `data` and `file` flags are required, but mutually exclusive.

Usage: `discovery ingestion pipeline store [flags]`

Flags:

`-d, --data`::
(Required, string) Set the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `file` flag.

`-f, --file`::
(Required, string) Set the path of the file that contains the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `data` flag.

`--abort-on-error`::
(Optional, bool) Aborts the operation when an error occurs. The default value is `false`.

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Store a pipeline with the JSON configuration in a file
discovery ingestion pipeline store --file pipelines.json
{"active":true,"creationTimestamp":"2025-10-31T19:41:13Z","id":"36f8ce72-f23d-4768-91e8-58693ff1b272","initialState":"ingestionState","labels":[],"lastUpdatedTimestamp":"2025-10-31T19:54:23Z","name":"my-pipeline","recordPolicy":{"errorPolicy":"FAIL","idPolicy":{},"outboundPolicy":{"batchPolicy":{"flushAfter":"PT1M","maxCount":25},"mode":"INLINE","splitPolicy":{"children":{"idPolicy":{},"snapshotPolicy":{}},"source":{"snapshotPolicy":{}}}},"retryPolicy":{"active":true,"maxRetries":3},"timeoutPolicy":{"record":"PT1M"}},"states":{"ingestionState":{"processors":[{"active":true,"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","outputField":"header"},{"active":true,"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8"}],"type":"processor"}}}
{"code":1003,"messages":["Entity not found: 5888b852-d7d4-4761-9058-738b2ad1b5c9"],"status":404,"timestamp":"2025-10-31T19:55:34.723693100Z"}
{"active":true,"creationTimestamp":"2025-10-31T19:55:34.758757Z","id":"bfb882a7-59e6-4cd6-afe4-7732163637f1","initialState":"ingestionState","labels":[],"lastUpdatedTimestamp":"2025-10-31T19:55:34.758757Z","name":"my-pipeline-3","recordPolicy":{"errorPolicy":"FAIL","idPolicy":{},"outboundPolicy":{"batchPolicy":{"flushAfter":"PT1M","maxCount":25},"mode":"INLINE","splitPolicy":{"children":{"idPolicy":{},"snapshotPolicy":{}},"source":{"snapshotPolicy":{}}}},"retryPolicy":{"active":true,"maxRetries":3},"timeoutPolicy":{"record":"PT1M"}},"states":{"ingestionState":{"processors":[{"active":true,"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","outputField":"header"},{"active":true,"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8"}],"type":"processor"}}}
```

```bash
# Store a pipeline with the JSON configuration in the data flag
discovery ingestion pipeline store --data '{"name":"my-pipeline","labels":[],"active":true,"id":"36f8ce72-f23d-4768-91e8-58693ff1b272","creationTimestamp":"2025-10-31T19:41:13Z","lastUpdatedTimestamp":"2025-10-31T19:41:13Z","initialState":"ingestionState","states":{"ingestionState":{"type":"processor","processors":[{"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","outputField":"header","active":true},{"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8","active":true}]}},"recordPolicy":{"idPolicy":{},"retryPolicy":{"active":true,"maxRetries":3},"timeoutPolicy":{"record":"PT1M"},"errorPolicy":"FAIL","outboundPolicy":{"batchPolicy":{"maxCount":25,"flushAfter":"PT1M"}}}}'
{"active":true,"creationTimestamp":"2025-10-31T19:41:13Z","id":"36f8ce72-f23d-4768-91e8-58693ff1b272","initialState":"ingestionState","labels":[],"lastUpdatedTimestamp":"2025-10-31T19:54:23Z","name":"my-pipeline","recordPolicy":{"errorPolicy":"FAIL","idPolicy":{},"outboundPolicy":{"batchPolicy":{"flushAfter":"PT1M","maxCount":25},"mode":"INLINE","splitPolicy":{"children":{"idPolicy":{},"snapshotPolicy":{}},"source":{"snapshotPolicy":{}}}},"retryPolicy":{"active":true,"maxRetries":3},"timeoutPolicy":{"record":"PT1M"}},"states":{"ingestionState":{"processors":[{"active":true,"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","outputField":"header"},{"active":true,"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8"}],"type":"processor"}}}
```

###### Delete
`delete` is the command used to delete Discovery Ingestion's pipelines. The user must send a name or UUID to delete a specific pipeline.

Usage: `discovery ingestion pipeline delete [flags] <arg>`

Arguments:
`arg`::
(Required, string) The name or UUID of the pipeline that will be deleted.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

```bash
# Delete a pipeline by id
discovery ingestion pipeline delete 04536687-f083-4353-8ecc-b7348e14b748
{"acknowledged":true}
```

```bash
# Delete a pipeline by name
discovery ingestion pipeline delete "my-pipeline"
{"acknowledged":true}
```

##### Seed
`seed` is the command used to manage seeds in Discovery Ingestion. This command contains various subcommands used to create, read, update, and delete.

Usage: `discovery ingestion seed [subcommand] [flags]`

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

###### Get
`get` is the command used to obtain Discovery Ingestion's seeds. The user can send a name or UUID to get a specific seed. If no argument is given, then the command retrieves every seed. The command also supports filters with the flag `filter` followed by the filter in the format `filter=key:value`. The `get` command can also get records from the seed with the `record` flag. Finally, the get command can retrieve a seed execution using the `execution` flag. When combined with the `details` flag, it provides more detailed information about the execution.

Usage: `discovery ingestion seed get [flags] [<arg>]`

Arguments:
`arg`::
(Optional, string) The name or UUID of the seed that will be retrieved.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-f, --filter`::
(Optional, Array of strings) Add a filter to the search. The available filters are the following:
- Label: The format is `label={key}[:{value}]`, where the value is optional.
- Type: The format is `type={type}`.

`--record`::
(Optional, string) The id of the record that will be retrieved. The result is appended to the seed in a `record` field.

`--execution`::
(Optional, string) The UUID of the seed execution that will be retrieved.

`--details`::
(Optional, string) Makes the get operation retrieve more information when getting a seed execution, like the audited changes and record and job summaries.

The `filter`, `execution`, and `record` flags are mutually exclusive. The `details` flag can only be used with the `execution` flag.

Examples:

```bash
# Get a seed by id
discovery ingestion seed get 7251d693-7382-452f-91dc-859add803a43
{"active":true,"config":{"action":"scroll","bucket":"blogs"},"creationTimestamp":"2025-10-31T22:54:08Z","id":"7251d693-7382-452f-91dc-859add803a43","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:54:08Z","name":"my-seed","pipeline":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","recordPolicy":{"errorPolicy":"FATAL","outboundPolicy":{"batchPolicy":{"flushAfter":"PT1M","maxCount":25},"idPolicy":{}},"timeoutPolicy":{"slice":"PT1H"}},"type":"staging"}
```

```bash
# Get seed by name
discovery ingestion seed get "my-seed"
{"active":true,"config":{"action":"scroll","bucket":"blogs"},"creationTimestamp":"2025-10-31T22:54:08Z","id":"7251d693-7382-452f-91dc-859add803a43","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:54:08Z","name":"my-seed","pipeline":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","recordPolicy":{"errorPolicy":"FATAL","outboundPolicy":{"batchPolicy":{"flushAfter":"PT1M","maxCount":25},"idPolicy":{}},"timeoutPolicy":{"slice":"PT1H"}},"type":"staging"}
```

```bash
# Get seeds using filters
discovery ingestion seed get --filter label=A:A -f type=staging
{"active":true,"creationTimestamp":"2025-10-31T22:53:54Z","id":"326a6b94-8931-4d18-b3a6-77adad14f2c0","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:53:54Z","name":"my-seed","type":"staging"}
{"active":true,"creationTimestamp":"2025-10-31T22:54:04Z","id":"8fbd57cd-9f82-409d-8d85-98e4b9225f3a","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:54:04Z","name":"my-seed-2","type":"staging"}
```

```bash
# Get all seeds using the configuration in profile "cn"
discovery ingestion seed get -p cn
{"active":true,"creationTimestamp":"2025-09-05T19:19:30Z","id":"026c6cf3-cba4-4d68-9806-1e534eebb99d","labels":[],"lastUpdatedTimestamp":"2025-09-05T19:19:30Z","name":"my-seed","type":"staging"}
{"active":true,"creationTimestamp":"2025-10-31T22:54:05Z","id":"028120a6-1859-47c7-b69a-f417e54b4a4a","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:54:05Z","name":"my-seed-2","type":"staging"}
{"active":true,"creationTimestamp":"2025-09-05T19:48:00Z","id":"0517a87a-86f7-4a71-bb3f-adfa0c87a269","labels":[],"lastUpdatedTimestamp":"2025-09-05T19:48:00Z","name":"my-seed-3","type":"staging"}
```

```bash
# Get a seed record by id
discovery ingestion seed get 2acd0a61-852c-4f38-af2b-9c84e152873e --record A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=
{"active":true,"config":{"action":"scroll","bucket":"blogs"},"creationTimestamp":"2025-08-21T21:52:03Z","id":"2acd0a61-852c-4f38-af2b-9c84e152873e","labels":[],"lastUpdatedTimestamp":"2025-08-21T21:52:03Z","name":"my-seed","pipeline":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","record":{"creationTimestamp":"2025-09-04T21:05:25Z","id":{"hash":"A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=","plain":"4e7c8a47efd829ef7f710d64da661786"},"lastUpdatedTimestamp":"2025-09-04T21:05:25Z","status":"SUCCESS"},"recordPolicy":{"errorPolicy":"FATAL","outboundPolicy":{"batchPolicy":{"flushAfter":"PT1M","maxCount":25},"idPolicy":{}},"timeoutPolicy":{"slice":"PT1H"}},"type":"staging"}
```

```bash
# Get a seed execution by id and with details
discovery ingestion seed get 2acd0a61-852c-4f38-af2b-9c84e152873e --execution 0f20f984-1854-4741-81ea-30f8b965b007 --details
{
  "audit": [
    {
      "stages": [],
      "status": "CREATED",
      "timestamp": "2025-11-18T16:22:23.865Z"
    },
    {
      "stages": [],
      "status": "RUNNING",
      "timestamp": "2025-11-18T16:22:34.655Z"
    },
    {
      "stages": [
        "BEFORE_HOOKS"
      ],
      "status": "RUNNING",
      "timestamp": "2025-11-18T16:23:13.120Z"
    }
  ],
  "creationTimestamp": "2025-11-18T16:22:24Z",
  "id": "0f20f984-1854-4741-81ea-30f8b965b007",
  "jobs": {
    "DONE": 3,
    "RUNNING": 1
  },
  "lastUpdatedTimestamp": "2025-11-18T16:23:13Z",
  "records": {
    "CREATE": {
      "PROCESSING": 2
    }
  },
  "scanType": "FULL",
  "stages": [
    "BEFORE_HOOKS"
  ],
  "status": "RUNNING",
  "triggerType": "MANUAL"
}
```

###### Store
`store` is the command used to create and update Discovery Ingestion's seeds. With the `data` flag, the user can send a single JSON configuration or an array to upsert multiple seeds. With the `file` flag, the user can also send the path of a file that contains the JSON configurations. The `data` and `file` flags are required, but mutually exclusive.

Usage: `discovery ingestion seed store [flags]`

Flags:

`-d, --data`::
(Required, string) Set the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `file` flag.

`-f, --file`::
(Required, string) Set the path of the file that contains the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `data` flag.

`--abort-on-error`::
(Optional, bool) Aborts the operation when an error occurs. The default value is `false`.

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Store a seed with the JSON configuration in a file
discovery ingestion seed store --file seeds.json
{"active":true,"config":{"action":"scroll","bucket":"blogs"},"creationTimestamp":"2025-09-04T15:50:08Z","id":"1d81d3d5-58a2-44a5-9acf-3fc8358afe09","labels":[],"lastUpdatedTimestamp":"2025-09-04T15:50:08Z","name":"my-seed","pipeline":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","recordPolicy":{"errorPolicy":"FATAL","outboundPolicy":{"batchPolicy":{"flushAfter":"PT1M","maxCount":25},"idPolicy":{}},"timeoutPolicy":{"slice":"PT1H"}},"type":"staging"}
{"code":1003,"messages":["Entity not found: 1d81d3d5-58a2-44a5-9acf-3fc8358afe00"],"status":404,"timestamp":"2025-10-31T20:32:39.832877700Z"}
{"active":true,"config":{"action":"scroll","bucket":"blogs"},"creationTimestamp":"2025-10-31T20:32:39.855952Z","id":"d818d852-18ac-4059-8f17-37a1b649bbfd","labels":[],"lastUpdatedTimestamp":"2025-10-31T20:32:39.855952Z","name":"my-seed-3","pipeline":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","recordPolicy":{"errorPolicy":"FATAL","outboundPolicy":{"batchPolicy":{"flushAfter":"PT1M","maxCount":25},"idPolicy":{}},"timeoutPolicy":{"slice":"PT1H"}},"type":"staging"}
```

```bash
# Store a seed with the JSON configuration in the data flag
discovery ingestion seed store --data '{"type":"staging","name":"my-seed","labels":[],"active":true,"id":"1d81d3d5-58a2-44a5-9acf-3fc8358afe09","creationTimestamp":"2025-09-04T15:50:08Z","lastUpdatedTimestamp":"2025-09-04T15:50:08Z","config":{"action":"scroll","bucket":"blogs"},"pipeline":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","recordPolicy":{"errorPolicy":"FATAL","timeoutPolicy":{"slice":"PT1H"},"outboundPolicy":{"idPolicy":{},"batchPolicy":{"maxCount":25,"flushAfter":"PT1M"}}}}'
{"active":true,"config":{"action":"scroll","bucket":"blogs"},"creationTimestamp":"2025-09-04T15:50:08Z","id":"1d81d3d5-58a2-44a5-9acf-3fc8358afe09","labels":[],"lastUpdatedTimestamp":"2025-09-04T15:50:08Z","name":"my-seed","pipeline":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","recordPolicy":{"errorPolicy":"FATAL","outboundPolicy":{"batchPolicy":{"flushAfter":"PT1M","maxCount":25},"idPolicy":{}},"timeoutPolicy":{"slice":"PT1H"}},"type":"staging"}
```

###### Delete
`delete` is the command used to delete Discovery Ingestion's seeds. The user must send a name or UUID to delete a specific seed.

Usage: `discovery ingestion seed delete [flags] <arg>`

Arguments:
`arg`::
(Required, string) The name or UUID of the seed that will be deleted.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

```bash
# Delete a seed by id
discovery ingestion seed delete 04536687-f083-4353-8ecc-b7348e14b748
{"acknowledged":true}
```

```bash
# Delete a seed by name
discovery ingestion seed delete "my-seed"
{"acknowledged":true}
```

###### Start
`start` is the command used to start a seed execution in Discovery Ingestion. With the `properties` flag, the user can set the execution properties with which to run the seed. With the `scan-type` flag, the user can set the scan type of the execution: `FULL` or `INCREMENTAL`.

Usage: `discovery ingestion seed start <arg> [flags]`

Arguments:
`arg`::
(Required, string) The name or UUID of the seed that will be executed.

Flags:

`--properties`::
(Optional, string) Set the properties of the seed execution.

`--scan-type`::
(Optional, string) Sets the scan type of the seed execution. It can be `FULL` or `INCREMENTAL`.

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Start a seed execution with no flags
discovery ingestion seed start 1d81d3d5-58a2-44a5-9acf-3fc8358afe09
{"creationTimestamp":"2025-11-03T23:56:18.513923Z","id":"f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36","lastUpdatedTimestamp":"2025-11-03T23:56:18.513923Z","scanType":"FULL","status":"CREATED","triggerType":"MANUAL"}
```

```bash
# Start a seed execution with no flags using the seed's name
discovery ingestion seed start "my-seed"
{"creationTimestamp":"2025-11-03T23:56:18.513923Z","id":"f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36","lastUpdatedTimestamp":"2025-11-03T23:56:18.513923Z","scanType":"FULL","status":"CREATED","triggerType":"MANUAL"}
```

```bash
# Start a seed execution with the properties and scan-type flags
discovery ingestion seed start --scan-type FULL --properties '{"stagingBucket":"my-bucket"}' 0ce1bece-5a01-4d4a-bf92-5ca3cd5327f3
{"creationTimestamp":"2025-11-03T23:58:23.972883Z","id":"cb48ab6b-577a-4354-8edf-981e1b0c9acb","lastUpdatedTimestamp":"2025-11-03T23:58:23.972883Z","properties":{"stagingBucket":"my-bucket"},"scanType":"FULL","status":"CREATED","triggerType":"MANUAL"}
```

###### Halt
`halt` is the command used to halt a seed execution in Discovery Ingestion. With the `execution` flag, the user can specify the specific execution that will be halted. If there is no `execution` flag, all of the active executions are halted.

Usage: `discovery ingestion seed halt <seed> [flags] `

Arguments:
`seed`::
(Required, string) The name or UUID of the seed that will have its executions halted.

Flags:

`--execution`::
(Optional, string) The UUID of the execution that will be halted.

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Halt all active seed executions
discovery ingestion seed halt 0ce1bece-5a01-4d4a-bf92-5ca3cd5327f3
{"id":"cb48ab6b-577a-4354-8edf-981e1b0c9acb","status":202}
```

```bash
# Halt a single seed execution
discovery ingestion seed halt 1d81d3d5-58a2-44a5-9acf-3fc8358afe09 --execution f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36
{"acknowledged":true}
```

#### QueryFlow
`queryflow` is the main command used to interact with Discovery's QueryFlow.

Usage: `discovery queryflow [subcommand] [flags]`

Flags:
`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

##### Config
`config` is the command used to interact with Discovery QueryFlow's configuration for a profile. This command by itself asks the user to save Discovery QueryFlow's configuration for the given profile. The command prints the property to be modified along with its current value. If the property currently being shown is sensitive, its value is obfuscated. To keep the current value, the user must press \"Enter\" without any text, and to set the value as empty, a sole whitespace must be inputted.

Usage: `discovery queryflow config [subcommand] [flags]`

Flags:
`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Ask the user for the configuration of profile "cn"
discovery queryflow config -p cn
Editing profile "cn". Press Enter to keep the value shown, type a single space to set empty.

QueryFlow URL [http://discovery.queryflow.cn]: https://discovery.queryflow.cn
QueryFlow API Key [*************.queryflow.cn]: 
```

```bash
# Config works without the profile. The rest of the command's output is omitted.
discovery queryflow config
Editing profile "default". Press Enter to keep the value shown, type a single space to set empty.
```

###### Get
`get` is the command used to obtain Discovery QueryFlow's configuration for a given profile. If the API keys are sensitive, the `sensitive` flag can be set to true in order to obfuscate them before printing them out. If a configuration property was not set, it is not displayed.

Usage: `discovery queryflow config get [flags]`

Flags:
`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-s, --sensitive`::
(Optional, bool) Obfuscates the API Keys if true. Defaults to `true`.

Examples: 

```bash
# Print the configuration of the "cn" profile with obfuscated API keys.
discovery queryflow config get -p cn
Showing the configuration of profile "cn":

QueryFlow URL: "https://discovery.queryflow.cn"
QueryFlow API Key: "*************.queryflow.cn"
```

```bash
# Print the configuration of the "default" profile.
discovery queryflow config get -s
Showing the configuration of profile "default":

QueryFlow URL: "http://localhost:12010"
QueryFlow API Key: ""
```

```bash
# Print the configuration of the "cn" profile with unobfuscated API keys.
discovery queryflow config get -p cn --sensitive=false
Showing the configuration of profile "cn":

QueryFlow URL: "https://discovery.queryflow.cn"
QueryFlow API Key: "discovery.key.queryflow.cn"
```

##### Export
`export` is the command used to backup Discovery QueryFlow's entities. With the `file` flag, the user can send the specific file in which to save the configurations. If not, they will be saved in a zip file in the current directory.

Usage: `discovery queryflow export [flags]`

Flags:
`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-f, --file`::
(Optional, string) The file that will contain the exported entities.

Examples:

```bash
# Export the entities using profile "cn".
discovery queryflow export -p cn
{"acknowledged":true}
```

```bash
# Export the entities to a specific file
discovery queryflow export --file "entities/queryflow.zip".
{"acknowledged":true}
```

##### Import
`import` is the command used to restore Discovery QueryFlow's entities. With the `file` flag, the user must send the specific file that has the entities' configuration. With the `on-conflict` flag, the user can send the conflict resolution strategy in case there are duplicate entities.

Usage: `discovery queryflow import [flags]`

Flags:
`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-f, --file`::
(Required, string) The file that contains the configurations of the entities.

`--on-conflict`::
(Optional, string) Sets the conflict resolution strategy when importing entities with the same id. The default value is "FAIL".

Examples:

```bash
# Import the entities using profile "cn" and fail conflict resolution strategy.
# The rest of the command's output is omitted.
discovery queryflow import -p cn --file "entities/queryflow.zip"
{
  "Endpoint": [
    {
      "errorCode": 2001,
      "errors": [
        "Duplicated entity: 2fee5e27-4147-48de-ba1e-d7f32476a4a2"
      ],
      "id": "2fee5e27-4147-48de-ba1e-d7f32476a4a2",
      "status": 409
    },
    {
      "errorCode": 2001,
      "errors": [
        "Duplicated entity: 4ef9da31-2ba6-442c-86bb-1c9566dac4c2"
      ],
      "id": "4ef9da31-2ba6-442c-86bb-1c9566dac4c2",
      "status": 409
    }
    ...
```

##### Processor
`processor` is the command used to manage processors in Discovery QueryFlow. This command contains various subcommands used to create, read, update, and delete.

Usage: `discovery queryflow processor [subcommand] [flags]`

Flags:
`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

###### Get
`get` is the command used to obtain Discovery QueryFlow's processors. The user can send a name or UUID to get a specific processor. If no argument is given, then the command retrieves every processor. The command also supports filters with the flag `filter` followed by the filter in the format `filter=key:value`.

Usage: `discovery queryflow processor get [flags] [<arg>]`

Arguments:
`arg`::
(Optional, string) The name or UUID of the processor that will be retrieved.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-f, --filter`::
(Optional, string) Add a filter to the search. The available filters are the following:
- Label: The format is `label={key}[:{value}]`, where the value is optional.
- Type: The format is `type={type}`.

Examples:

```bash
# Get a processor by id
discovery queryflow processor get 8e9ce4af-0f0b-44c7-bff7-c3c4f546e577
{"active":true,"config":{"action":"aggregate","collection":"blogs","database":"pureinsights","stages":[{"$match":{"$text":{"$search":"#{ data(\"/httpRequest/queryParams/q\") }"}}}]},"creationTimestamp":"2025-11-06T14:52:14Z","id":"8e9ce4af-0f0b-44c7-bff7-c3c4f546e577","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-11-06T14:52:14Z","name":"MongoDB text processor","server":{"credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","id":"f6950327-3175-4a98-a570-658df852424a"},"type":"mongo"}
```

```bash
# Get processor by name
discovery queryflow processor get "OpenAI Chat Processor"
{"active":true,"config":{"action":"chat-completion","messages":[{"content":"#{ data(\"/script\") }","role":"user"}],"model":"gpt-4.1"},"creationTimestamp":"2025-11-20T00:10:50Z","id":"8a399b1c-95fc-406c-a220-7d321aaa7b0e","labels":[],"lastUpdatedTimestamp":"2025-11-20T00:10:50Z","name":"OpenAI Chat Processor","server":{"credential":"9be0e625-a510-46c5-8130-438823f849c2","id":"741df47e-208f-47c1-812f-53cc62c726af"},"type":"openai"}
```

```bash
# Get processors using filters
discovery queryflow processor get --filter label=A:A -f type=mongo
{"active":true,"creationTimestamp":"2025-11-06T14:52:01Z","id":"628d4b24-84cc-4070-8eed-c3155cf40fe9","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-11-06T14:52:01Z","name":"MongoDB text processor","type":"mongo"}
{"active":true,"creationTimestamp":"2025-11-06T14:52:14Z","id":"8e9ce4af-0f0b-44c7-bff7-c3c4f546e577","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-11-06T14:52:14Z","name":"MongoDB store processor","type":"mongo"}
```

```bash
# Get all processors using the configuration in profile "cn"
discovery queryflow processor get -p cn
{"active":true,"creationTimestamp":"2025-11-06T14:52:16Z","id":"019ecd8e-76c9-41ee-b047-299b8aa14aba","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-11-06T14:52:16Z","name":"MongoDB text processor","type":"mongo"}
{"active":true,"creationTimestamp":"2025-11-06T14:52:17Z","id":"0a7caa9b-99aa-4a63-aa6d-a1e40941984d","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-11-06T14:52:17Z","name":"MongoDB store processor","type":"mongo"}
```

###### Store
`store` is the command used to create and update Discovery QueryFlow's processors. With the `data` flag, the user can send a single JSON configuration or an array to upsert multiple processors. With the `file` flag, the user can also send the path of a file that contains the JSON configurations. The `data` and `file` flags are required, but mutually exclusive.

Usage: `discovery queryflow processor store [flags]`

Flags:

`-d, --data`::
(Required, string) Set the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `file` flag.

`-f, --file`::
(Required, string) Set the path of the file that contains the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `data` flag.

`--abort-on-error`::
(Optional, bool) Aborts the operation when an error occurs. The default value is `false`.

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Store a processor with the JSON configuration in a file
discovery queryflow processor store --file "queryflowprocessorjsonfile.json"
{"active":true,"config":{"action":"aggregate","collection":"blogs","database":"pureinsights","stages":[{"$match":{"$text":{"$search":"#{ data(\"/httpRequest/queryParams/q\") }"}}}]},"creationTimestamp":"2025-11-20T00:08:23Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","labels":[],"lastUpdatedTimestamp":"2025-11-20T00:08:23Z","name":"MongoDB text processor","server":{"credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","id":"f6950327-3175-4a98-a570-658df852424a"},"type":"mongo"}
{"code":1003,"messages":["Entity not found: 0a7caa9b-99aa-4a63-aa6d-a1e40941984e"],"status":404,"timestamp":"2025-11-20T00:16:10.216366100Z"}
{"active":true,"config":{"action":"aggregate","collection":"blogs","database":"pureinsights","stages":[{"$match":{"$text":{"$search":"#{ data(\"/httpRequest/queryParams/q\") }"}}}]},"creationTimestamp":"2025-11-20T00:16:10.253229Z","id":"74f4cffb-1a4f-4470-8485-f759cdc203bd","labels":[],"lastUpdatedTimestamp":"2025-11-20T00:16:10.253229Z","name":"MongoDB store processor","server":{"credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","id":"f6950327-3175-4a98-a570-658df852424a"},"type":"mongo"}
```

```bash
# Store a processor with the JSON configuration in the data flag
discovery queryflow processor store --data '{"type":"mongo","name":"MongoDB text processor","labels":[],"active":true,"id":"3393f6d9-94c1-4b70-ba02-5f582727d998","creationTimestamp":"2025-11-20T00:08:23Z","lastUpdatedTimestamp":"2025-11-20T00:08:23Z","config":{"action":"aggregate","stages":[{"$match":{"$text":{"$search":"#{ data(\"/httpRequest/queryParams/q\") }"}}}],"database":"pureinsights","collection":"blogs"},"server":{"id":"f6950327-3175-4a98-a570-658df852424a","credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c"}}'
{"active":true,"config":{"action":"aggregate","collection":"blogs","database":"pureinsights","stages":[{"$match":{"$text":{"$search":"#{ data(\"/httpRequest/queryParams/q\") }"}}}]},"creationTimestamp":"2025-11-20T00:08:23Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","labels":[],"lastUpdatedTimestamp":"2025-11-20T00:08:23Z","name":"MongoDB text processor","server":{"credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","id":"f6950327-3175-4a98-a570-658df852424a"},"type":"mongo"}
```

###### Delete
`delete` is the command used to delete Discovery QueryFlow's processors. The user must send a name or UUID to delete a specific processor.

Usage: `discovery queryflow processor delete [flags] <arg>`

Arguments:
`arg`::
(Required, string) The name or UUID of the processor that will be deleted.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

```bash
# Delete a processor by id
discovery queryflow processor delete 189b3fa5-e011-43aa-ae57-f6e4a6f4b552
{"acknowledged":true}
```

```bash
# Delete a processor by name
discovery queryflow processor delete processor1
{"acknowledged":true}
```

##### Endpoint
`endpoint` is the command used to manage endpoints in Discovery QueryFlow. This command contains various subcommands used to create, read, update, and delete.

Usage: `discovery queryflow endpoint [subcommand] [flags]`

Flags:
`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

###### Get
`get` is the command used to obtain Discovery QueryFlow's endpoints. The user can send a name or UUID to get a specific endpoint. If no argument is given, then the command retrieves every endpoint. The command also supports filters with the flag `filter` followed by the filter in the format `filter=key:value`.

Usage: `discovery queryflow endpoint get [flags] [<arg>]`

Arguments:
`arg`::
(Optional, string) The name or UUID of the endpoint that will be retrieved.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-f, --filter`::
(Optional, string) Add a filter to the search. The available filters are the following:
- Label: The format is `label={key}[:{value}]`, where the value is optional.
- Type: The format is `type={type}`.

Examples:

```bash
# Get an endpoint by id
discovery queryflow endpoint get cf56470f-0ab4-4754-b05c-f760669315af
{"active":true,"creationTimestamp":"2025-11-06T16:24:40Z","httpMethod":"GET","id":"cf56470f-0ab4-4754-b05c-f760669315af","initialState":"searchState","labels":[{"key":"A","value":"B"}],"lastUpdatedTimestamp":"2025-11-06T16:24:40Z","name":"Wikis endpoint","states":{"responseState":{"body":{"answer":"#{ data('/answer/choices/0/message/content') }"},"statusCode":200,"type":"response"},"searchState":{"mode":{"type":"group"},"next":"responseState","processors":[{"active":true,"continueOnError":false,"id":"b5c25cd3-e7c9-4fd2-b7e6-2bcf6e2caf89"},{"active":true,"continueOnError":false,"id":"a5ee116b-bd95-474e-9d50-db7be988b196"},{"active":true,"continueOnError":false,"id":"86e7f920-a4e4-4b64-be84-5437a7673db8"},{"active":true,"continueOnError":false,"id":"8a399b1c-95fc-406c-a220-7d321aaa7b0e","outputField":"answer"}],"type":"processor"}},"timeout":"PT1H","type":"default","uri":"/wikis-search"}
```

```bash
# Get an endpoint by name
discovery queryflow endpoint get "Blogs endpoint"
{"active":true,"creationTimestamp":"2025-11-20T00:08:26Z","httpMethod":"GET","id":"4ef9da31-2ba6-442c-86bb-1c9566dac4c2","initialState":"searchState","labels":[],"lastUpdatedTimestamp":"2025-11-20T00:08:26Z","name":"Blogs endpoint","states":{"searchState":{"mode":{"type":"group"},"processors":[{"active":true,"continueOnError":false,"id":"5f125024-1e5e-4591-9fee-365dc20eeeed"}],"type":"processor"}},"timeout":"PT1H","type":"default","uri":"/blogs-search"}
```

```bash
# Get endpoints using filters
discovery queryflow endpoint get --filter label=A:B
{"active":true,"creationTimestamp":"2025-11-06T16:24:40Z","httpMethod":"GET","id":"cf56470f-0ab4-4754-b05c-f760669315af","labels":[{"key":"A","value":"B"}],"lastUpdatedTimestamp":"2025-11-06T16:24:40Z","name":"Wikis endpoint","timeout":"PT1H","type":"default","uri":"/wikis-search"}
{"active":true,"creationTimestamp":"2025-11-06T16:24:54Z","httpMethod":"GET","id":"2fee5e27-4147-48de-ba1e-d7f32476a4a2","labels":[{"key":"A","value":"B"}],"lastUpdatedTimestamp":"2025-11-06T16:24:54Z","name":"Blogs search endpoint","timeout":"PT1H","type":"default","uri":"/blogs-search"}
```

```bash
# Get all endpoints using the configuration in profile "cn"
discovery queryflow endpoint get -p cn
{"active":true,"creationTimestamp":"2025-11-06T16:24:54Z","httpMethod":"GET","id":"2fee5e27-4147-48de-ba1e-d7f32476a4a2","labels":[{"key":"A","value":"B"}],"lastUpdatedTimestamp":"2025-11-06T16:24:54Z","name":"Wikis endpoint","timeout":"PT1H","type":"default","uri":"/wikis-search"}
{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","httpMethod":"GET","id":"4ef9da31-2ba6-442c-86bb-1c9566dac4c2","labels":[],"lastUpdatedTimestamp":"2025-08-25T16:47:24Z","name":"Blogs endpoint","timeout":"PT1H","type":"default","uri":"/blogs-search"}
```

###### Store
`store` is the command used to create and update Discovery QueryFlow's endpoints. With the `data` flag, the user can send a single JSON configuration or an array to upsert multiple endpoints. With the `file` flag, the user can also send the path of a file that contains the JSON configurations. The `data` and `file` flags are required, but mutually exclusive.

Usage: `discovery queryflow endpoint store [flags]`

Flags:

`-d, --data`::
(Required, string) Set the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `file` flag.

`-f, --file`::
(Required, string) Set the path of the file that contains the JSON configurations of the entities that will be stored. This flag is mutually exclusive to the `data` flag.

`--abort-on-error`::
(Optional, bool) Aborts the operation when an error occurs. The default value is `false`.

`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Store an endpoint with the JSON configuration in a file
discovery queryflow endpoint store --file endpointjsonfile.json
{"active":true,"creationTimestamp":"2025-11-20T00:10:53Z","httpMethod":"GET","id":"cf56470f-0ab4-4754-b05c-f760669315af","initialState":"searchState","labels":[{"key":"A","value":"B"}],"lastUpdatedTimestamp":"2025-11-20T00:10:53Z","name":"Wikis endpoint","states":{"responseState":{"body":{"answer":"#{ data('/answer/choices/0/message/content') }"},"statusCode":200,"type":"response"},"searchState":{"mode":{"type":"group"},"next":"responseState","processors":[{"active":true,"continueOnError":false,"id":"b5c25cd3-e7c9-4fd2-b7e6-2bcf6e2caf89"},{"active":true,"continueOnError":false,"id":"a5ee116b-bd95-474e-9d50-db7be988b196"},{"active":true,"continueOnError":false,"id":"86e7f920-a4e4-4b64-be84-5437a7673db8"},{"active":true,"continueOnError":false,"id":"8a399b1c-95fc-406c-a220-7d321aaa7b0e","outputField":"answer"}],"type":"processor"}},"timeout":"PT1H","type":"default","uri":"/wikis-search"}
{"code":1003,"messages":["Entity not found: 2fee5e27-4147-48de-ba1e-d7f32476a4a3"],"status":404,"timestamp":"2025-11-20T00:37:02.827065700Z"}
{"active":true,"creationTimestamp":"2025-11-20T00:37:02.857266Z","httpMethod":"GET","id":"7324b140-3240-4e67-90cb-9ffe5e7f574b","initialState":"searchState","labels":[{"key":"A","value":"B"}],"lastUpdatedTimestamp":"2025-11-20T00:37:02.857266Z","name":"Blogs search endpoint","states":{"responseState":{"body":{"answer":"#{ data('/answer/choices/0/message/content') }"},"statusCode":200,"type":"response"},"searchState":{"mode":{"type":"group"},"next":"responseState","processors":[{"active":true,"continueOnError":false,"id":"b5c25cd3-e7c9-4fd2-b7e6-2bcf6e2caf89"},{"active":true,"continueOnError":false,"id":"a5ee116b-bd95-474e-9d50-db7be988b196"},{"active":true,"continueOnError":false,"id":"86e7f920-a4e4-4b64-be84-5437a7673db8"},{"active":true,"continueOnError":false,"id":"8a399b1c-95fc-406c-a220-7d321aaa7b0e","outputField":"answer"}],"type":"processor"}},"timeout":"PT1H","type":"default","uri":"/blog-search"}
```

```bash
# Store an endpoint with the JSON configuration in the data flag
discovery queryflow endpoint store --data '{"type":"default","name":"Wikis endpoint","labels":[{"key":"A","value":"B"}],"active":true,"id":"cf56470f-0ab4-4754-b05c-f760669315af","creationTimestamp":"2025-11-20T00:10:53Z","lastUpdatedTimestamp":"2025-11-20T00:10:53Z","httpMethod":"GET","uri":"/wikis-search","timeout":"PT1H","initialState":"searchState","states":{"searchState":{"type":"processor","processors":[{"id":"b5c25cd3-e7c9-4fd2-b7e6-2bcf6e2caf89","continueOnError":false,"active":true},{"id":"a5ee116b-bd95-474e-9d50-db7be988b196","continueOnError":false,"active":true},{"id":"86e7f920-a4e4-4b64-be84-5437a7673db8","continueOnError":false,"active":true},{"id":"8a399b1c-95fc-406c-a220-7d321aaa7b0e","outputField":"answer","continueOnError":false,"active":true}],"mode":{"type":"group"},"next":"responseState"},"responseState":{"type":"response","statusCode":200,"body":{"answer":"#{ data('/answer/choices/0/message/content') }"}}}}'
{"active":true,"creationTimestamp":"2025-11-20T00:10:53Z","httpMethod":"GET","id":"cf56470f-0ab4-4754-b05c-f760669315af","initialState":"searchState","labels":[{"key":"A","value":"B"}],"lastUpdatedTimestamp":"2025-11-20T00:10:53Z","name":"Wikis endpoint","states":{"responseState":{"body":{"answer":"#{ data('/answer/choices/0/message/content') }"},"statusCode":200,"type":"response"},"searchState":{"mode":{"type":"group"},"next":"responseState","processors":[{"active":true,"continueOnError":false,"id":"b5c25cd3-e7c9-4fd2-b7e6-2bcf6e2caf89"},{"active":true,"continueOnError":false,"id":"a5ee116b-bd95-474e-9d50-db7be988b196"},{"active":true,"continueOnError":false,"id":"86e7f920-a4e4-4b64-be84-5437a7673db8"},{"active":true,"continueOnError":false,"id":"8a399b1c-95fc-406c-a220-7d321aaa7b0e","outputField":"answer"}],"type":"processor"}},"timeout":"PT1H","type":"default","uri":"/wikis-search"}
```

###### Delete
`delete` is the command used to delete Discovery QueryFlow's endpoints. The user must send a name or UUID to delete a specific endpoint.

Usage: `discovery queryflow endpoint delete [flags] <arg>`

Arguments:
`arg`::
(Required, string) The name or UUID of the endpoint that will be deleted.

Flags:

`-h, --help`::
(Optional, bool) Prints the usage of the command.
`-p, --profile`::

(Optional, string) Set the configuration profile that will execute the command.

```bash
# Delete a endpoint by id
discovery queryflow endpoint delete ea02fc14-f07b-49f2-b185-e9ceaedcb367
{"acknowledged":true}

```
```bash
# Delete a endpoint by name
discovery queryflow endpoint delete endpoint1
{"acknowledged":true}
```

#### Staging
`staging` is the main command used to interact with Discovery's Staging. 

Usage: `discovery staging [subcommand] [flags]`

Flags:
`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

##### Config
`config` is the command used to interact with Discovery Staging's configuration for a profile. This command by itself asks the user to save Discovery Staging's configuration for the given profile. The command prints the property to be modified along with its current value. If the property currently being shown is sensitive, its value is obfuscated. To keep the current value, the user must press \"Enter\" without any text, and to set the value as empty, a sole whitespace must be inputted.

Usage: `discovery staging config [subcommand] [flags]`

Flags:
`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

Examples:

```bash
# Ask the user for the configuration of profile "cn"
discovery staging config -p cn
Editing profile "cn". Press Enter to keep the value shown, type a single space to set empty.

Staging URL [http://discovery.staging.cn]: https://discovery.staging.cn
Staging API Key [*************.staging.cn]: 
```

```bash
# Config works without the profile. The rest of the command's output is omitted.
discovery staging config
Editing profile "default". Press Enter to keep the value shown, type a single space to set empty.
```

###### Get
`get` is the command used to obtain Discovery Staging's configuration for a given profile. If the API keys are sensitive, the `sensitive` flag can be set to true in order to obfuscate them before printing them out. If a configuration property was not set, it is not displayed.

Usage: `discovery staging config get [flags]`

Flags:
`-h, --help`::
(Optional, bool) Prints the usage of the command.

`-p, --profile`::
(Optional, string) Set the configuration profile that will execute the command.

`-s, --sensitive`::
(Optional, bool) Obfuscates the API Keys if true. Defaults to `true`.

Examples: 

```bash
# Print the configuration of the "cn" profile with obfuscated API keys.
discovery staging config get -p cn
Showing the configuration of profile "cn":

Staging URL: "https://discovery.staging.cn"
Staging API Key: "*************.staging.cn"
```

```bash
# Print the configuration of the "default" profile.
discovery staging config get -s
Showing the configuration of profile "default":

Staging URL: "http://localhost:12010"
Staging API Key: ""
```

```bash
# Print the configuration of the "cn" profile with unobfuscated API keys.
discovery staging config get -p cn --sensitive=false
Showing the configuration of profile "cn":

Staging URL: "https://discovery.staging.cn"
Staging API Key: "discovery.key.staging.cn"
```

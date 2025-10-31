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
QueryFlow URL [http://discovery.queryflow.cn]: http://localhost:12040/v2 
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
QueryFlow URL: "http://localhost:12040/v2"
QueryFlow API Key: "********w213"
Staging URL: "http://discovery.staging.cn"
Staging API Key: "***************taging.cn"
```

```bash
# Print the configuration of the "default" profile.
discovery config get -s
Showing the configuration of profile "default":

Core URL: "http://localhost:12010/v2"
Core API Key: ""
Ingestion URL: "http://localhost:12030/v2"
Ingestion API Key: ""
QueryFlow URL: "http://localhost:12040/v2"
QueryFlow API Key: ""
Staging URL: "http://localhost:12020/v2"
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
QueryFlow URL: "http://localhost:12040/v2"
QueryFlow API Key: "queryflow213"
Staging URL: "http://discovery.staging.cn"
Staging API Key: "discovery.key.staging.cn"
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

Core URL: "http://localhost:12010/v2"
Core API Key: ""
```

```bash
# Print the configuration of the "cn" profile with unobfuscated API keys.
discovery core config get -p cn --sensitive=false
Showing the configuration of profile "cn":

Core URL: "https://discovery.core.cn"
Core API Key: "discovery.key.core.cn"
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

Usage: `discovery core label get [flags] <uuid>`

Arguments:
`uuid`::
(Optional, String) The UUID of the label that will be retrieved.

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
{"creationTimestamp":"2025-09-29T19:45:51Z","id":"b667b650-9ddf-490a-bc89-276987c4076f","key":"B","lastUpdatedTimestamp":"2025-09-29T19:45:51Z","value":"B"}
{"creationTimestamp":"2025-10-15T20:25:29Z","id":"bf0e20b7-24de-448d-b7c8-e4721f51e3dc","key":"E","lastUpdatedTimestamp":"2025-10-15T20:25:29Z","value":"F"}
{"creationTimestamp":"2025-10-15T20:26:27Z","id":"e113751a-b063-40f5-9a8b-f7cd82094cc7","key":"E","lastUpdatedTimestamp":"2025-10-15T20:26:27Z","value":"G"}
{"creationTimestamp":"2025-10-15T20:26:27Z","id":"f37907e3-1f7a-481c-88f4-bc263eff2613","key":"H","lastUpdatedTimestamp":"2025-10-15T20:26:27Z","value":"F"}
{"creationTimestamp":"2025-09-29T19:45:52Z","id":"f5e01fb8-1503-4401-ba56-039548259739","key":"C","lastUpdatedTimestamp":"2025-09-29T19:45:52Z","value":"C"}
```

```bash
# Try to get label by name
discovery core label get label1
Error: Could not convert given id "label1" to UUID. This command does not support filters or referencing an entity by name
invalid UUID length: 6
Usage:
  discovery core label get [flags]

Flags:
  -h, --help        help for get

Global Flags:
  -p, --profile string   configuration profile to use (default "default")
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

Usage: `discovery core secret get [flags] <uuid>`

Arguments:
`uuid`::
(Optional, String) The UUID of the secret that will be retrieved.

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

```bash
# Try to get secret by name
discovery core secret get secret1
Error: Could not convert given id "secret1" to UUID. This command does not support filters or referencing an entity by name
invalid UUID length: 7
Usage:
  discovery core secret get [flags]

Flags:
  -h, --help        help for get

Global Flags:
  -p, --profile string   configuration profile to use (default "default")
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

Usage: `discovery core credential get [flags] <arg>`

Arguments:
`arg`::
(Optional, String) The name or UUID of the credential that will be retrieved.

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
# Get a credential by id
discovery core credential get 3b32e410-2f33-412d-9fb8-17970131921c
{"active":true,"creationTimestamp":"2025-10-17T22:37:57Z","id":"3b32e410-2f33-412d-9fb8-17970131921c","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T22:37:57Z","name":"label test 1 clone 10","secret":"mongo-secret","type":"mongo"}

```

```bash
# Get credential by name
discovery core credential get "label test 1 clone 10"
{"highlight":{"name":["\u003cem\u003elabel\u003c/em\u003e \u003cem\u003etest\u003c/em\u003e \u003cem\u003e1\u003c/em\u003e \u003cem\u003eclone\u003c/em\u003e \u003cem\u003e10\u003c/em\u003e"]},"score":1.63641,"source":{"active":true,"creationTimestamp":"2025-10-17T22:37:57Z","id":"3b32e410-2f33-412d-9fb8-17970131921c","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T22:37:57Z","name":"label test 1 clone 10","type":"mongo"}}
```

```bash
# Get credentials using filters
discovery core credential get --filter label=A:A --filter type=mongo
{"highlight":{},"score":0.22338435,"source":{"active":true,"creationTimestamp":"2025-10-17T15:33:58Z","id":"8c243a1d-9384-421d-8f99-4ef28d4e0ab0","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T15:33:58Z","name":"label test 1","type":"mongo"}}
```

```bash
# Get all credentials using the configuration in profile "cn"
discovery core credential get -p cn
{"active":true,"creationTimestamp":"2025-10-17T22:37:57Z","id":"3b32e410-2f33-412d-9fb8-17970131921c","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T22:37:57Z","name":"label test 1 clone 10","type":"mongo"}
{"active":true,"creationTimestamp":"2025-10-17T22:40:15Z","id":"458d245a-6ed2-4c2b-a73f-5540d550a479","labels":[{"key":"A","value":"B"}],"lastUpdatedTimestamp":"2025-10-17T22:40:15Z","name":"label test 2","type":"mongo"}
{"active":true,"creationTimestamp":"2025-10-17T22:37:56Z","id":"46cb4fff-28be-4901-b059-1dd618e74ee4","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T22:37:56Z","name":"label test 1 clone 9","type":"mongo"}
{"active":true,"creationTimestamp":"2025-10-17T22:37:53Z","id":"4957145b-6192-4862-a5da-e97853974e9f","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T22:37:53Z","name":"label test 1 clone","type":"mongo"}
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
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
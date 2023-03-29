# PDP-CLI - **User Documentation**

## Installing

## Getting started

Get started and create a PDP project using the CLI-PDP. It will make easier the interaction with each
PDP product and even migrate a project to a different environment.

### Create a PDP project

To create the initial structure of a PDP project run the following command.

#### Make sure everything is working fine.

```bash
pdp health 
```

#### Create a PDP project named HelloWorld with some initial entites already configured.

```bash
pdp config init -n HelloWorld --template random_generator 
```

## Documentation

### PDP Command

This is the command responsible to pass the configuration or namespace to all the other commands.
If no argument is passed to it will use the default values.

#### Flags:

- **--namespace**: Namespace in which the PDP components are running. **Default is "pdp"**.
- **--profile**: Configuration profile to load specific configurations from pdp.ini. **Default is "DEFAULT"**.

```bash
pdp --namespace <namespace> --profile <configuration_profile>
```

### Health Command

This command is used to assure that the CLI works fine. It doesn't have any flags.

```bash
pdp health
```

### Init Command

Creates a new project from existing sources or from scratch.
It will create the folder structure for a PDP project.

#### Flags:

- **-n,--project-name**: The name of the resulting directory, will try to fetch existing configurations from the APIs
  referenced in ~/.pdp. Notice that imported configs have id fields, don't change those.
  **Default is ```my-pdp-project```**.
- **--empty/--no-empty**: If ```True``` it will create a project from a template (the default template)
  If ```False``` it will try to import the entities for the products urls.
  **Default is ```False```**.
- **-u,--product-url**: The base URL for the given product API. The product URL must be provided with the following
  format **PRODUCT_NAME:URL**. The command allows multiple flags to define multiples products.
  Default are ```ingestion http://localhost:8080```, ```staging http://localhost:8081```,
  ```core http://localhost:8082```, ```discovery http://localhost:8088```.
- **--force/--no-force**: If there is a project with the same name it will to override it. **Default is False**.
- **--template**: If the project will be created from a template, it will use the name of the templated provided
  by the user. Default is ```random_generator```. Available are ```empty, random_generator```.

#### An example to a project from scratch

```bash
pdp config init -n <project_name> --empty --template empty
```

#### An example to create a project from existing entities

```bash
pdp config init -n <project_name> --no-empty -u ingestion <ingestion_url> -u core <core_url> -u staging <staging_url> -u discovery <discovery_url>
```

### Deploy Command

Deploys project configurations to the target products.
Must be run within the directory from a project created with the 'init' command.
Will replace any name reference with IDs. Names are case-sensitive. If the "id"
field is missing from an entity, assumes this is a new instance.

#### Flags:

- **-d,--dir**: The path to a directory with the structure and the pdp.ini that init command creates. **Default
  is ```./```**.
- **--target**: The name of the product where you want to deploy the entities. The command allows multiple flags to
  define multiple targets. **Default are ```[ingestion, core, discovery]```**.
- **-v,--verbose**: It will show more information about the deployment results. **Default is ```False```**.
- **-g,--ignore-ids/--no-ignore-ids**: Will cause existing ids to be ignored, hence everything will be created as a new
  instance. This is useful when moving configs from one instance to another. **Default is ```False```**.
- **-q,--quiet**: Display only the seed ids. Warnings and Errors will not be shown neither. **Default is ```False```**.

#### Deploy entities to Discovery and Ingestion

```bash
pdp config deploy -d ./my-pdp-project --target discovery --target ingestion
```

#### Deploy entities ignoring ids with verbose mode

```bash
pdp config deploy -d ./my-pdp-project -g -v --target discovery --target core
```

#### Deploy entities to all targets with quiet mode

```bash
pdp config deploy -d ./my-pdp-project -q
```
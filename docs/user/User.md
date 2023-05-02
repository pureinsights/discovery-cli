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
- **-d,--dir**: The path to a directory with the structure and the pdp.ini that init command creates. **Default
  is ```./```**.

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
- **--empty/--no-empty**: If ```True``` will create a project from a template (the default template)
  If ```False``` will try to import the entities for the products urls.
  **Default is ```False```**.
- **-u,--product-url**: The base URL for the given product API. The product URL must be provided with the following
  format **PRODUCT_NAME:URL**. The command allows multiple flags to define multiples products.
  Default are ```ingestion http://localhost:8080```, ```staging http://localhost:8081```,
  ```core http://localhost:8082```, ```discovery http://localhost:8088```.
- **--force/--no-force**: If there is a project with the same name will to override it. **Default is False**.
- **--template**: If the project will be created from a template, will use the name of the templated provided
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

- **--target**: The name of the product where you want to deploy the entities. The command allows multiple flags to
  define multiple targets. **Default are ```[ingestion, core, discovery]```**.
- **-v,--verbose**: Will show more information about the deployment results. **Default is ```False```**.
- **-g,--ignore-ids/--no-ignore-ids**: Will cause existing ids to be ignored, hence everything will be created as a new
  instance. This is useful when moving configs from one instance to another. **Default is ```False```**.
- **-q,--quiet**: Display only the seed ids. Warnings and Errors will not be shown neither. **Default is ```False```**.

#### Deploy entities to Discovery and Ingestion

```bash
pdp -d ./my-pdp-project config deploy --target discovery --target ingestion
```

#### Deploy entities ignoring ids with verbose mode

```bash
pdp -d ./my-pdp-project config deploy -g -v --target discovery --target core
```

#### Deploy entities to all targets with quiet mode

```bash
pdp -d ./my-pdp-project config deploy -q
```

### Create Command

Add a new entity configuration to the entities on the current project. The configuration for each entity it will have
default values depending on the template name provided, or you can specify your own entity configuration with the --file
and/or --interactive flags. You can also deploy the entities to their respective product.

#### Flags

- `REQUIRED`**-t, --entity-type**: This is the type of the entity that will be created. The entity types supported at
  the moment are: ```[seed, ingestionProcessor, pipeline, Scheduler, Endpoint, discoveryProcessor]```.
- **--entity-template**: This is the template's name of the entity to use. Default is ```None```.
- **--deploy**: It will deploy the entity configuration to the corresponding product. Default is ```False```.
- **--file**: The path to the file that contains the configuration for the entity or entities. If the configuration
  contains an id property it will be updated instead. Default is ```None```.
- **--interactive**: This is a Boolean flag. Will launch your default text editor to allow you to modify the entity
  configuration. Default is ```False```.
- **-j, --json**: This is a Boolean flag. Will print the results in JSON format. Default is ```False```.
- **-g, --ignore-ids**: Will cause existing ids to be ignored, hence everything will be created as a new instance. This
  is useful when moving configs from one instance to another. Default is ```False```.

#### Add entities to a project and deploy them

```bash
pdp -d ./my-pdp-project config create --entity-type pipeline --file ./my-new-entities.json --deploy -g -j
```

#### Add entity to a project from a template and edit it.

```bash
pdp -d ./my-pdp-project config create --entity-type pipeline --entity-template empty_pipeline --interactive
```

### Get Command

Retrieves information about all the entities deployed on PDP Products. You can search by products, entity types or
even by id. And you can filter the results by giving a property and the expected value to match the entities.

#### Flags

- `OPTIONAL`**--product**: Will filter the entities based on the name entered (Ingestion, Core or Discovery).
  Default is All.
- `OPTIONAL`**-t,--entity-type**: Will filter and only show the entities of the type entered. Default is All.
- `OPTIONAL`**-i,--entity-id**: Will only retrieve information for the component specified by the ID. Default is None.
  The command allows multiple flags of -i.
- **-j,--json**: This is a boolean flag. Will print the results in JSON format. Default is False.
- **-v,--verbose**: Will show more information. Default is False.
- **-f,--filter**: Will filter the results by the key-value given by the user with the following format property value.
  The command allows multiple flags of filter. Default is [].
- **-p,--page**: The number of the page to show. Min 0. Default is 0.
- **-s,--size**: The size of the page to show. Range 1 - 100. Default is 25.
- **--asc**: The name of the property to sort in ascending order. Multiple flags are supported. Default is [].
- **--desc**: The name of the property to sort in descending order. Multiple flags are supported. Default is [].

#### Get all entities from all products from each entity type

```bash
pdp config get -v
```

#### Get all entities from all products from each entity type and printed as JSON

```bash
pdp config get -j
```

#### Get all entities a product

```bash
pdp config get --product ingestion
```

#### Get all entities from an entity type

```bash
pdp config get --entity-type ingestionProcessor
```

#### Get entities from ids

```bash
pdp config get -i 29a9b5e600704853983b0dd855a11cc6 -i 6376af03-1af2-41a2-aef6-62aefc73a870
```

### Delete Command

Will attempt to delete the entity or entities from the product and the configuration files.
If an entity is referenced by another can’t be deleted.

#### Flags

- `DEPENDENT`**--product**: Will filter the entities based on the name entered (Ingestion, core or discovery). Default
  is All.
- `DEPENDENT`**-t,--entity-type**: Will filter and only show the entities of the type entered. Default is All.
- `DEPENDENT`**-i,--entity-id**: Will delete the entity with the specified id. Default is None. The command allows
  multiple flags of -i.
- **-a, --all**: Will try to delete entities based on `DEPENDENT` flags, that is, if the id is not provided by the user,
  it will attempt to delete all entities of the type provided by the user, and if the type of entity is not entered by
  the user, then it will attempt to delete all types of entities from a product, and so on.
- **--local**: Will delete the configuration of the entities from the PDP Project files. Default is False.

### Delete all entities

```bash
pdp config delete --all
```

### Delete all entities from Ingestion

```bash
pdp config delete --product ingestion --all
```

### Delete all entities of type endpoint

```bash
pdp config delete --entity-type endpoint --all
```

### Delete entities by ids

```bash
pdp config delete -i 504821c342f4e717cb1297739a83e7e3 -i 3b5086e9-e5bf-4c82-b23e-358036bc4a1b
```

### Delete entity in cascade

```bash
pdp config delete -i 8e6e1674-fb03-45b3-9ffd-5bf4b780af49 --cascade
```

#### Export Command

Will create a .zip file with the configuration files for the entities given by the user.

### Flags

- `OPTIONAL`**--product**: Will filter the entities based on the name entered (Ingestion, core or discovery). Default
  is All.
- `OPTIONAL`**-t,--entity-type**: Will filter and only export the entities of the type entered. Default is All.
- `OPTIONAL`**-i,--entity-id**: Will only export the component specified by the ID. Default is None.
- **--include-dependencies/--no-include-dependencies**: Will include those entities which are dependencies for the
  entity identified with the given id. Default is False.

### Export all entities

```bash
pdp config export
```

### Export all entities from a product

```bash
pdp config export --product discovery
```

### Export an entity

```bash
pdp config export --entity-type seed -i 3b5086e9-e5bf-4c82-b23e-358036bc4a1b
```

### Export an entity with his dependencies

```bash
pdp config export --entity-type seed -i 3b5086e9-e5bf-4c82-b23e-358036bc4a1b --include-dependencies
```

#### Import Command

Will import a .zip to a given product. The commands assume that the zip contains the files and structure necessary for
each product.

### Flags

- `REQUIRED`**--target**: Will import the given file to the specified product. (Ingestion, Core or Discovery).
- `REQUIRED`**--zip**: The path to the zip that will be imported.

### Import entities

```bash
pdp config import --product discovery --zip endpoints.zip
```

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

#### Create a PDP project named HelloWorld with some initial entities already configured.

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

#### Delete all entities

```bash
pdp config delete --all
```

#### Delete all entities from Ingestion

```bash
pdp config delete --product ingestion --all
```

#### Delete all entities of type endpoint

```bash
pdp config delete --entity-type endpoint --all
```

#### Delete entities by ids

```bash
pdp config delete -i 504821c342f4e717cb1297739a83e7e3 -i 3b5086e9-e5bf-4c82-b23e-358036bc4a1b
```

#### Delete entity in cascade

```bash
pdp config delete -i 8e6e1674-fb03-45b3-9ffd-5bf4b780af49 --cascade
```

### Export Command

Will create a .zip file with the configuration files for the entities given by the user.

#### Flags

- `OPTIONAL`**--product**: Will filter the entities based on the name entered (Ingestion, core or discovery). Default
  is All.
- `OPTIONAL`**-t,--entity-type**: Will filter and only export the entities of the type entered. Default is All.
- `OPTIONAL`**-i,--entity-id**: Will only export the component specified by the ID. Default is None.
- **--include-dependencies/--no-include-dependencies**: Will include those entities which are dependencies for the
  entity identified with the given id. Default is False.

#### Export all entities

```bash
pdp config export
```

#### Export all entities from a product

```bash
pdp config export --product discovery
```

#### Export an entity

```bash
pdp config export --entity-type seed -i 3b5086e9-e5bf-4c82-b23e-358036bc4a1b
```

#### Export an entity with his dependencies

```bash
pdp config export --entity-type seed -i 3b5086e9-e5bf-4c82-b23e-358036bc4a1b --include-dependencies
```

### Import Command

Will import a .zip to a given product. The commands assume that the zip contains the files and structure necessary for
each product.

#### Flags

- `REQUIRED`**--target**: Will import the given file to the specified product. (Ingestion, Core or Discovery).
- `REQUIRED`**--zip**: The path to the zip that will be imported.

#### Import entities

```bash
pdp config import --product discovery --zip endpoints.zip
```

### Core Command

#### Search Command

Search for entities of all products Ingestion, Core and Discovery. And also, is a group command to chain the 'replace'
command and replace the entities on the search results.

##### Flags

- **-l,--label**: Label key or label key and value of the entity. Format: <key>:<value> | <key>. Multiple label flags
  are supported. Default is None.
- **-t,--entity-type**: Type of the entity. Format <product>:<entityType>. Product values supported Ingestion and
  Discovery. Default is None.
- **-q**: The name or description of the entity. Default is None.
- **-p,--page**: The number of the page to show. Min 0. Default is 0.
- **-s,--size**: The size of the page to show. Range 1 - 100. Default is 25.
- **--asc**: The name of the property to sort in ascending order. Multiple flags are supported. Default is [].
- **--desc**: The name of the property to sort in descending order. Multiple flags are supported. Default is [].

##### Search for all entities

```bash
pdp core search
```

##### Search for all entities of types

```bash
pdp core search -t ingestion:processor -t discovery:processor
```

##### Search for entities with filters

```bash
pdp core search -q entity_name --page 2 --size 5 --asc id --desc name
```

#### Log-Level Command

Change the logging level of a component.

##### Flags

- `REQURIED`**--component**: The name of the component that you want to change the log level.
- `REQUIRED`**--level**: The level log you want to change to. Values supported ERROR,WARN, INFO,DEBUG and TRACE.
- **--logger**: The of the logger. Default is None.

##### Change the log leve of a component

```bash
pdp core log-level --component component_name --level ERROR
```

##### Change the log leve of a logger from component

```bash
pdp core log-level --component component_name --level ERROR --logger logger_name
```

#### File Command

##### Upload Command

Try to upload the provided file to the Core API.

###### Flags

- **--name**: The name of the file, if no name is provided, then the name will be the name found in the path.
- `REQUIRED` **--path**: The path where the file is located. If just a name is passed instead of a path to the file the
  cli will try to find the file in the ./Core/files/.

###### Upload a file with a specific name

```bash
pdp core file upload --name seed --path ./seed_entities.json
```

###### Upload a file without specified the name

```bash
# The name of the file will be entities
pdp core file upload --path ./entities.zip
```

##### Download Command

Will try to download a file previously uploaded to the Core API.

###### Flags

- `REQUIRED`**--name**: The name of the file that you want to download.
- **--path**: The path where the file will be written. Default is ./Core/files/ if you are in a PDP project, if not,
  default is ./.

###### Download a file

```bash
pdp core file download --name entities.json --path ./my-entities/
```

###### Download a file and rename it

```bash
pdp core file download --name entities.json --path ./my-entities/pdp_entities.json
```

###### Download a file on the current folder

```bash
pdp core file download --name entities.json
```

##### Delete Command

Will delete a file from the Core API.

###### Flags

- `REQUIRED`**-n, --name**: The name of the file you want to delete. You can provide a full path too to use it with the
  --local flag. The command allows multiple flags of -n.
- `DEPENDENT`**--local**: This is a boolean flag, it will try to delete the file from your pc too. It will use the path
  provided by the flag name, if just a name was passed and not a path it will search for the file on the ./core/files.
  Default is False.

###### Delete files

```bash
pdp core file delete --name file_name --name another_file
```

###### Delete file from local too

```bash
pdp core file delete --name file_name --local
```

##### Ls Command

Show the list of files uploaded to the Core API.

###### Flags

- **--json**: This is a boolean flag. Will print the results in JSON format. Default is False.

###### Show files from Core API

```bash
pdp core file ls
```

###### Show files from Core API as JSON format

```bash
pdp core file ls --json
```

### Seed-Exec Command

#### Start Command

Try to start the scanning process of a seed. Note that a seed can only have one active execution at a time.

##### Flags

- `REQUIRED`**--seed**: The id of the seed to start the scanning process.
- **--scan-type**: The strategy to apply during the scan phase. Values supported INCREMENTAL and FULL. Default is FULL.

##### Start execution of a seed

```bash
pdp seed-exec start --seed 3b5086e9-e5bf-4c82-b23e-358036bc4a1b
```

##### Start execution of a seed with a specific scan type

```bash
pdp seed-exec start --seed 3b5086e9-e5bf-4c82-b23e-358036bc4a1b --scan-type incremental
```

#### Reset Command

Reset all the associated data of the given seed.

##### Flags

- `REQUIRED`**--seed**: The id of the seed to reset the associated data.

##### Reset associated data of a seed

```bash
pdp seed-exec reset --seed 3b5086e9-e5bf-4c82-b23e-358036bc4a1b
```

#### Control Command

Triggers and action on all active executions for the given seed.

##### Flags

- `REQUIRED`**--seed**: The id of the seed to trigger the action.
- **--action**: The action you want to trigger. Values supported HALT, PAUSE and RESUME. Default is HALT.

##### Control the execution of a seed

```bash
pdp seed-exec control --seed 3b5086e9-e5bf-4c82-b23e-358036bc4a1b --action halt
```

#### Get Command

Retrieves the executions of a given seed.

##### Flags

- `REQUIRED`**--seed**: The id of the seed you want to get the active executions.
- **--execution**: The id of the execution you want to get information. Default is None. The command allows multiple
  flags of --execution.
- **-j, --json**: This is a boolean flag. It will print the results in JSON format. Default is False.
- **-v, --verbose**: It will show more information about the deploy results. Default is False.
- **-p, --page**: The number of the page to show. Min 0. Default is 0.
- **-s, --size**: The size of the page to show. Range 1 - 100. Default is 25.
- **--asc**: The name of the property to sort in ascending order. Multiple flags are supported. Default is [].
- **--desc**: The name of the property to sort in descending order. Multiple flags are supported. Default is [].

##### Get executions from a seed

```bash
pdp seed-exec get --seed 3b5086e9-e5bf-4c82-b23e-358036bc4a1b
```

##### Get executions from a seed with parameters

```bash
pdp seed-exec get --seed 3b5086e9-e5bf-4c82-b23e-358036bc4a1b --page 2 --size 1 --asc pipelineId --desc status -v
```

##### Get executions from a seed by id

```bash
pdp seed-exec get --seed 3b5086e9-e5bf-4c82-b23e-358036bc4a1b --execution ac5086e9-e5bf-4c82-b23e-358036bc4a23 --execution ac5086e9-e5bf-4c82-b23e-358036bc4b2a -v
```

##### Get executions from a seed in JSON format

```bash
pdp seed-exec get --seed 3b5086e9-e5bf-4c82-b23e-358036bc4a1b -j
```

### Staging Command

#### Bucket Command

This command encloses all commands that let you perform actions on a bucket on the staging API.

##### Get Command

Retrieves all the items for a given bucket. You can filter or use pagination, but just once at time. If you provide
--token or --content-type will prioritize the filter over pagination.

###### Flags

- `REQUIRED` **--bucket**:The name for the bucket to get the items.
- **--token**: The token of the contents you want to filter.
- **--content-type**: The content-type of the query. Default is CONTENT.
- **--page**: The number of the page to query.
- **--size**: The size of the page to query.
- **--asc**: The name of the property to sort in ascending order. Multiple flags are supported. Default is [].
- **--desc**: The name of the property to sort in descending order. Multiple flags are supported. Default is [].
- **-j, --json**: This is a boolean flag. It will print the results in JSON format. Default is False.

###### Get all the items of the bucket

```bash
pdp staging bucket get --bucket bucket_test
```

###### Get all the items of the bucket in JSON format

```bash
pdp staging bucket get --bucket bucket_test -j
```

###### Filter the bucket items

```bash
pdp staging bucket get --bucket bucket_test --token token_id --content-type BOTH --size 25
```

###### Use pagination on the bucket items

```bash
pdp staging bucket get --bucket bucket_test --page 2 --size 25 --asc property
```

##### Delete Command

Will delete the given bucket from the Staging API.

###### Flags

- **--bucket**: The name of the bucket to delete. If this flag is not provided an error will be raised. The command
  allows multiple flags of --bucket. Default is [].

###### Delete a bucket

```bash
pdp staging bucket delete --bucket bucket_name
```

###### Delete more than one bucket at once

```bash
pdp staging bucket delete --bucket bucket_name1 --bucket bucket_name2 --bucket bucket_name3
```

##### Status Command

Retrieves the status for all the buckets or a given bucket by the user.

###### Flags

- **--bucket**: The name of the bucket to get the status. Default is None.
- **-p, --page**: The number of the page to show. Min 0. Default is 0.
- **-s, --size**: The size of the page to show. Range 1 - 100. Default is 25.
- **--asc**: The name of the property to sort in ascending order. Multiple flags are supported. Default is [].
- **--desc**: The name of the property to sort in descending order. Multiple flags are supported. Default is [].
- **-j, --json**: This is a boolean flag. It will print the results in JSON format. Default is False.

###### Get the status for all the buckets

```bash 
pdp staging bucket status --page 0 --size 1 --asc property --desc property2
```

###### Get the status for a specific bucket

```bash 
pdp staging bucket status --bucket bucket_name
```

###### Get the status for a specific bucket in JSON format

```bash 
pdp staging bucket status --bucket bucket_name --json
``` 

##### Batch Command

Performs a list of actions such as ADD and DELETE to a given bucket within the Staging API.

###### Flags

- `REQUIRED`**--bucket**: The name of the bucket to perform the actions.
- **--file**: The path to the file that contains the body for the query on a JSON format.
- **--interactive**: Will open a text editor to let you write the body for the request.
- **-j, --json**: This is a boolean flag. It will print the results in JSON format. Default is False.

###### Process a batch of actions from a file

```bash
pdp staging bucket batch --bucket bucket_name --file file/path
```

###### Process a batch of actions interactively

```bash
pdp staging bucket batch --bucket bucket_name --interactive
```

###### Process a batch of actions from a file and edit its contents

```bash
pdp staging bucket batch --bucket bucket_name --file file/path --interactive
```

###### Process a batch of actions from a file and print the results in JSON format

```bash
pdp staging bucket batch --bucket bucket_name --file file/path -j
``` 

#### Item Command

This command encloses all commands that let you perform actions on an item on the staging API.

##### Add Command

Adds a new item to a given bucket within the staging API. If the bucket doesn't exist, will be created.

###### Flags

- `REQURIED`**--bucket**: The name of the bucket where the item will be added.
- **--item-id**: The id of the new item. If no id is provided, then an auto-generated hash will be set. Default is None.
- **--parent**: This allows you to add an item within an existing item. Default is None.
- **--interactive**: This is a Boolean flag. Will launch your default text editor to allow you to modify the entity
  configuration. Default is False.
- **--file**: The path to the file that contains the content of the item. Default is None.
- **-j, --json**: This is a boolean flag. It will print the results in JSON format. Default is False.
- **-v, --verbose**: It will show more information about the item upload. Default is False.

###### Add an item

```bash
pdp staging item add --bucket bucket --item-id item_id --file item.json
```

###### Add an item interactively and with an autogenerated id

```bash
 pdp staging item add --bucket bucket --file item.json --interactive
```

##### Get Command

Retrieves the information of the given item.

###### Flags

- `REQURIED`**--bucket**: The name of the bucket where the item will be added.
- `REQURIED`**-i, --item-id**: The id of the item to show. Default is []. The command allows multiple flags of -i.
- **--content-type**: The content-type of the query. Default is CONTENT. Allowed are CONTENT, METADATA, BOTH.
- **-j, --json**: This is a boolean flag. It will print the results in JSON format. Default is False.

###### Get the content of the item within a bucket

```bash
pdp staging item get --bucket bucket --item item_id
```

###### Get the metadata of the item within a bucket

```bash
pdp staging item get --bucket bucket --item item_id --content-type metadata
```

###### Get the content and the metadata of the item within a bucket

```bash
pdp staging item get --bucket bucket --item item_id --content-type both
```

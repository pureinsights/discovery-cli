# PDP-CLI - **User Documentation**

## Installing

## Getting started

Get started and create a PDP project using the CLI-PDP. It will make easier the interaction with each
PDP product and even migrate a project to a different environment.

### Create a PDP project

To create the initial structure of a PDP project run the following command.

```bash
# Make sure everything is working fine.
pdp health 
# Create a PDP project named HelloWorld with some initial entites already configured.
pdp config init -n HelloWorld --template random_generator 
```

## Documentation

### Command PDP

This is the command responsible to pass the configuration or namespace to all the other commands.
If no argument is passed to it will use the default values.

#### Flags:

- **--namespace**: Namespace in which the PDP components are running. **Default is "pdp"**.
- **--profile**: Configuration profile to load specific configurations from pdp.ini. **Default is "DEFAULT"**.

```bash
pdp --namespace <namespace> --profile <configuration_profile>
```

### Command Health

This command is used to assure that the CLI works fine. It doesn't have any flags.

```bash
pdp health
```

### Command init

Creates a new project from existing sources or from scratch.
It will create the folder structure for a PDP project.

#### Flags:

- **-n,--project-name**: The name of the resulting directory, will try to fetch existing configurations from the APIs
  referenced in ~/.pdp. Notice that imported configs have id fields, don`t change those.
  **Default is my-pdp-project**.
- **--empty/--no-empty**: If is ```True``` it will create a project from a template (the default template)
  If is ```False``` it will try to import the entities for the products urls.
  **Default is False**.
- **-u,--product-url**: The base URL for the given product API. The product URL must be provided with the following
  format **PRODUCT_NAME:URL**. The command allows multiple flags to define multiples products.
  Default are ```ingestion http://localhost:8080```, ```staging http://localhost:8081```,
  ```core http://localhost:8082```, ```discovery http://localhost:8088```.
- **--force/--no-force**: If there is a project with the same name it will to override it. **Default is False**.
- **--template**: If the project will be created from a template, it will use the name of the templated provided
  by the user. Default is ```random_generator```. Available are ```empty, random_generator```.

```bash
# An example to a project from scratch
pdp config init -n <project_name> --empty --template empty
# An example to create a project from existing entities
pdp config init -n <project_name> --no-empty -u ingestion <ingestion_url> -u core <core_url> -u staging <staging_url> -u discovery <discovery_url>
```

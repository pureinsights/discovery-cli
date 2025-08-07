# Pureinsights Discovery Platform: Command Line Interface

## Development guidelines

This project uses the [cobra](https://github.com/spf13/cobra) library to build the Discovery CLI.
This section contains guidelines and best practices to contribute to the project and fix or extend its functionality.

### Usage and help messages

Cobra provides a feature to generate automatically help and usage messages for commands.
Which allows users to use the flag `--help` to show the generated message and get the command documentation.

However, we as developers have the responsibility to provide the information and documentation for each to command to cobra.
We can leverage the fields provided by the `cobra.Command` struct as follows:

```go
cobra.Command{
		Use:   "command [options] [flags]", // Follow the recommended syntax by cobra: https://pkg.go.dev/github.com/spf13/cobra#Command
		Short: "A short description for the command, shown in the 'help' output.",
    Long: `A more detailed and long description, shown in the 'help' of the specific command.
    Can be multiple lines.`,
    Example: `
      # Execute the command
      > discovery command option -flag
    `
}
```


### Flags 

Names should use `kebab` case, so those names with two or more words must be separated by `-`, e.g `--my-flag`. 

Whenever is possible flags should have a short flag form, e.g `--flag` `-f`.

Always use the same short flag for the same flag, across all the commands. That means, if you used `-f` as short flag of `--flag`, you 
shouldn't use `-f` as short flag of any other flag that is not `--flag` in any command.

Try to use meaningful short flags. A good approach could be use the first letter of the flag name, as long as the short flag doesn't conflict with other short flags for other flags.


### User documentation 

Documentation for each command should be added into the `USER.md` file.

Each command must have its own section or sub-section, depending if its a sub-command.

````markdown
<!-- The nesting of the title depends on the parent command -->
# Command Name

A description of the command, what it does, which positional arguments it expects.
Add any detail that the user using the command should now, or be aware of.

Usage: `discovery command <arg1> <arg2> [<optionalArg1> [<optionalArg2>]] (-r | --required-flag) [flags]`

Arguments:

`arg1`:: 
(Required, String) The description of the argument.

`arg2`:: 
(Required, Int) The description of the argument.

`optionalArg1`:: 
(Optional, String) The description of the argument. 

`optionalArg2`:: 
(Optional, String) The description of the argument.


Flags:

`-r, --required-flag`:: 
(Required, bool) The description of the required flag.

`-f, --flag`:: 
(Optional, float) The description of the flag.

Examples:

```bash
# Example description
discovery command "example string" 3 -r
```

```bash
# Example description with optionals
discovery command "example string" 3 example strings --required-flag -f 3.4
```
````



## Getting Started

### Build the CLI

To build the project and generate a binary file run the following command:

```bash
go build -o build/discovery
```

### Run the CLI

To run the project you can run the binary file directly.

```bash
cd ./build
./discovery
```

or you can run the project without building it.

```bash
go run main.go
```

### Install dependencies

To install dependencies declared in the `go.mod` file run:

```bash
go mod tidy
```

To install a package and add it to the `go.mod` file run:

```bash
# go get <package>
go get github.com/spf13/cobra
```


## Testing

### Unit Tests

To create unit tests you create a `*_test.go` file in the same package with functions with the form `func Test*(t *testing.T)` so Go can now that those functions are tests.

Example:

```go
// ./example/add.go
package example
func Add(n1, n2 int) int { return n1 + n2 } 
```

```go
// ./example/add_test.go
package example
func TestAdd(t *testing.T) {
  // The body of the test
}
```

#### Run tests

To run the tests of a specific package you can run the following command:

```bash
# go test <path/to/package> 
go test ./
```

To run the tests of a package and its subdirectories run:

```bash
# go test <path/to/package>/...
go test ./...
```

#### Coverage

To run the tests with coverage reporting run the following commands:

```bash
# Run the tests and generate the coverage profile
go test ./... -coverprofile=coverage.out
```

```bash
# Generates a HTML report based on the coverage profile
go tool cover -html=coverage.out
```


#### Naming convention

As mentioned before, the name of the file and the test function must follow a specific form in order to Go detected as tests, but to make tests names more readable, were
going to follow the next name conventions:

##### Testing functions

The name convention for normal functions is `func Test<name>[_ScenarioDescription](t *testing.T)`, for example:

```go
package example
func TestFunctionExample(t *testing.T) {} 
func TestFunctionExample_WithADescription(t *testing.T) {} 
```

##### Testing struct methods

The name convention for struct methods is `func Test<struct>_<method>[_ScenarioDescription](t *testing.T)`, for example:

```go
package example
func TestStructA_ExampleMethod(t *testing.T) {} 
func TestStructA_ExampleMethod_WithADescription(t *testing.T) {} 
```

##### Avoid naming collisions

To avoid name collisions, prepend those unexported `structs`, `methods` or `functions` with `_`.

```go
// Testing a unexported function
func Test_exampleFunction(t *testing.T) {}
func Test_exampleFunction_WithDescription(t *testing.T) {}
```

```go
// Testing a unexported struct with an unexported method
func Test_structA_exampleMethod(t *testing.T) {}
func Test_structA_exampleMethod_WithDescription(t *testing.T) {}
```
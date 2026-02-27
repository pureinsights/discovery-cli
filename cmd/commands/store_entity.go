package commands

import (
	"os"

	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/tidwall/gjson"
)

const (
	// LongStore is the message used in the Long field of the Store commands.
	LongStore string = "store is the command used to create and update Discovery %[2]s's %[1]ss. With the --data flag, the user can send a single JSON configuration or an array to upsert multiple %[1]s. On the other hand, the user can also send multiple arguments with the paths of files that contain JSON configurations. Each of these files will be processed individually, but all entities will be upserted. The --data flag and file arguments are required, but mutually exclusive. The user can only send the data flag or file arguments, not both at the same time."
	// DataEmptyError is the error message when the given data is empty.
	DataEmptyError string = "Data cannot be empty"
)

// StoreCommandConfig contains the parameters sent to the StoreCommand function.
type storeCommandConfig struct {
	commandConfig
	abortOnError bool
	data         string
	files        []string
}

// StoreCommandConfig creates a storeCommandConfig with the required parameters.
func StoreCommandConfig(baseConfig commandConfig, abortOnError bool, data string, files []string) storeCommandConfig {
	return storeCommandConfig{
		commandConfig: baseConfig,
		abortOnError:  abortOnError,
		data:          data,
		files:         files,
	}
}

// readDataFromFile is an auxiliary function to process the files sent as arguments.
func readDataFromFile(file string) (gjson.Result, error) {
	jsonBytes, err := os.ReadFile(file)
	if err != nil {
		err = cli.NormalizeReadFileError(file, err)
		return gjson.Result{}, cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not read file %q", file)
	}

	if len(jsonBytes) == 0 || string(jsonBytes) == "" {
		return gjson.Result{}, cli.NewError(cli.ErrorExitCode, DataEmptyError)
	}

	return gjson.ParseBytes(jsonBytes), nil
}

// prepareStoreCommand checks the credentials and returns the configured printer
func prepareStoreCommand(d cli.Discovery, config storeCommandConfig) (cli.Printer, error) {
	err := CheckCredentials(d, config.profile, config.componentName, config.url)
	if err != nil {
		return nil, err
	}

	output := config.output
	if output == "pretty-json" {
		output = "json"
	}
	return cli.GetArrayPrinter(output), nil
}

// StoreCommand has the command logic to upsert an entity into Discovery.
func StoreCommand(d cli.Discovery, client cli.Creator, config storeCommandConfig) error {
	printer, err := prepareStoreCommand(d, config)
	if err != nil {
		return err
	}

	if len(config.files) != 0 {
		if config.data != "" {
			return cli.NewError(cli.ErrorExitCode, "There cannot be both a file argument and the data flag")
		}

		for _, file := range config.files {
			data, err := readDataFromFile(file)
			if err != nil {
				return err
			}

			err = d.UpsertEntities(client, data, config.abortOnError, printer)
			if err != nil {
				return err
			}
		}
		return nil
	} else {
		if config.data == "" {
			return cli.NewError(cli.ErrorExitCode, DataEmptyError)
		}

		return d.UpsertEntities(client, gjson.Parse(config.data), config.abortOnError, printer)
	}
}

// SearchStoreCommand has the command logic to upsert an entity into Discovery and update an entity using its name.
func SearchStoreCommand(d cli.Discovery, client cli.SearchCreator, config storeCommandConfig) error {
	printer, err := prepareStoreCommand(d, config)
	if err != nil {
		return err
	}

	if len(config.files) != 0 {
		if config.data != "" {
			return cli.NewError(cli.ErrorExitCode, "There cannot be both a file argument and the data flag")
		}

		for _, file := range config.files {
			data, err := readDataFromFile(file)
			if err != nil {
				return err
			}

			err = d.SearchUpsertEntities(client, data, config.abortOnError, printer)
			if err != nil {
				return err
			}
		}
		return nil
	} else {
		if config.data == "" {
			return cli.NewError(cli.ErrorExitCode, DataEmptyError)
		}

		return d.SearchUpsertEntities(client, gjson.Parse(config.data), config.abortOnError, printer)
	}
}

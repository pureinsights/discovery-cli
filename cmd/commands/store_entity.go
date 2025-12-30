package commands

import (
	"os"

	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/tidwall/gjson"
)

const (
	// LongStore is the message used in the Long field of the Store commands.
	LongStore string = "store is the command used to create and update Discovery %[2]s's %[1]ss. With the --data flag, the user can send a single JSON configuration or an array to upsert multiple %[1]s. On the other hand, the user can also send multiple arguments with the paths of files that contain JSON configurations. Each of these files will be processed individually, but all entities will be upserted. The --data flag and file arguments are required, but mutually exclusive. The user can only send the data flag or file arguments, not both at the same time."
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

// upsertFromFiles is an auxiliary function to process the files sent as arguments and upload their entities.
func upsertFromFiles(d cli.Discovery, client cli.Creator, config storeCommandConfig, printer cli.Printer) error {
	for _, file := range config.files {
		jsonBytes, err := os.ReadFile(file)
		if err != nil {
			err = cli.NormalizeReadFileError(file, err)
			return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not read file %q", file)
		}

		if len(jsonBytes) == 0 || string(jsonBytes) == "" {
			return cli.NewError(cli.ErrorExitCode, "Data cannot be empty")
		}

		err = d.UpsertEntities(client, gjson.ParseBytes(jsonBytes), config.abortOnError, printer)
		if err != nil {
			return err
		}
	}

	return nil
}

// StoreCommand has the command logic to upsert an entity into Discovery.
func StoreCommand(d cli.Discovery, client cli.Creator, config storeCommandConfig) error {
	err := CheckCredentials(d, config.profile, config.componentName, config.url)
	if err != nil {
		return err
	}

	output := config.output
	if output == "pretty-json" {
		output = "json"
	}
	printer := cli.GetArrayPrinter(output)

	if len(config.files) != 0 {
		if config.data != "" {
			return cli.NewError(cli.ErrorExitCode, "There cannot be both a file argument and the data flag")
		}
		return upsertFromFiles(d, client, config, printer)
	} else {
		if config.data == "" {
			return cli.NewError(cli.ErrorExitCode, "Data cannot be empty")
		}

		return d.UpsertEntities(client, gjson.Parse(config.data), config.abortOnError, printer)
	}
}

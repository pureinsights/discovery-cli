package commands

import (
	"os"

	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/tidwall/gjson"
)

const (
	// LongStore is the message used in the Long field of the Store commands.
	LongStore string = "store is the command used to create and update Discovery %[2]s's %[1]ss. With the --data flag, the user can send a single JSON configuration or an array to upsert multiple %[1]s. With the --file flag, the user can also send the path of a file that contains the JSON configurations. The --data and --file flags are required, but mutually exclusive."
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
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not upsert entities in file %q", file)
			}
		}

		return nil
	} else {
		if config.data == "" {
			return cli.NewError(cli.ErrorExitCode, "Data cannot be empty")
		}

		return d.UpsertEntities(client, gjson.Parse(config.data), config.abortOnError, printer)
	}
}

package commands

import (
	"os"

	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/tidwall/gjson"
)

// StoreCommandConfig contains the parameters sent to the StoreCommand function.
type storeCommandConfig struct {
	commandConfig
	abortOnError bool
	data         string
	file         string
}

// StoreCommandConfig creates a storeCommandConfig with the required parameters.
func StoreCommandConfig(baseConfig commandConfig, abortOnError bool, data, file string) storeCommandConfig {
	return storeCommandConfig{
		commandConfig: baseConfig,
		abortOnError:  abortOnError,
		data:          data,
		file:          file,
	}
}

// StoreCommand has the command logic to upsert an entity into Discovery.
func StoreCommand(d cli.Discovery, client cli.Creator, config storeCommandConfig) error {
	err := checkCredentials(d, config.profile, config.componentName, config.url, config.apiKey)
	if err != nil {
		return err
	}

	if config.file != "" {
		jsonBytes, err := os.ReadFile(config.file)
		if err != nil {
			return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not read file %q", config.file)
		}

		config.data = string(jsonBytes)
	}

	if config.data == "" {
		return cli.NewError(cli.ErrorExitCode, "Data cannot be empty")

	}

	printer := cli.GetArrayPrinter(config.output)
	return d.UpsertEntities(client, gjson.Parse(config.data), config.abortOnError, printer)
}

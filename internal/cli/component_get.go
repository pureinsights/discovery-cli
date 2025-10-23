package cli

import (
	"github.com/google/uuid"
)

// GetCommand is the function that executes the get operation for the get commands that do not work with names or filters.
func GetCommand(args []string, d Discovery, client getter, config commandConfig) error {
	err := checkCredentials(d, config.profile, config.componentName, config.url, config.apiKey)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		id, err := uuid.Parse(args[0])
		if err != nil {
			return NewErrorWithCause(ErrorExitCode, err, "Could not convert given id %q to UUID. This command does not support filters or referencing an entity by name.", args[0])
		}
		printer := GetObjectPrinter(config.output)
		err = d.GetEntity(client, id, printer)
		return err
	} else {
		printer := GetArrayPrinter(config.output)
		err = d.GetEntities(client, printer)
		return err
	}
}

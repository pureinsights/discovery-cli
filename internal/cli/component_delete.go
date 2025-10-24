package cli

import (
	"github.com/google/uuid"
)

const (
	// LongDeleteNoNames is the message used in the Long field of the Delete commands that do not support getting by name.
	LongDeleteNoNames string = "get is the command used to obtain Discovery %[2]s's %[1]ss. The user can send a UUID to get a specific %[1]s. If no UUID is given, then the command retrieves every %[1]s. The optional argument must be a UUID. This command does not support filters or referencing an entity by name."
)

// DeleteCommand is the function that executes the delete operation for the delete commands that do not work with names.
func DeleteCommand(id string, d Discovery, client deleter, config commandConfig) error {
	err := checkCredentials(d, config.profile, config.componentName, config.url, config.apiKey)
	if err != nil {
		return err
	}

	deleteId, err := uuid.Parse(id)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not convert given id %q to UUID. This command does not support referencing an entity by name.", id)
	}
	printer := GetObjectPrinter(config.output)
	err = d.DeleteEntity(client, deleteId, printer)
	return err
}

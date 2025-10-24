package cli

import (
	"github.com/google/uuid"
)

const (
	// LongDeleteNoNames is the message used in the Long field of the Delete commands that do not support deleting by name.
	LongDeleteNoNames string = "delete is the command used to delete Discovery %[2]s's %[1]ss. The user must send a UUID to delete a specific %[1]s. If no UUID is given, then an error is returned. This command does not support referencing an entity by name."
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

// SearchDeleteCommand is the function that executes the delete operation for the delete commands that can also work with names.
func SearchDeleteCommand(id string, d Discovery, client searchDeleter, config commandConfig) error {
	err := checkCredentials(d, config.profile, config.componentName, config.url, config.apiKey)
	if err != nil {
		return err
	}

	printer := GetObjectPrinter(config.output)
	err = d.SearchDeleteEntity(client, id, printer)
	return err
}

package commands

import (
	"github.com/google/uuid"
	"github.com/pureinsights/pdp-cli/internal/cli"
)

const (
	// LongDeleteNoNames is the message used in the Long field of the Delete commands that do not support deleting by name.
	LongDeleteNoNames string = "delete is the command used to delete Discovery %[2]s's %[1]ss. The user must send a UUID to delete a specific %[1]s. If no UUID is given, then an error is returned. This command does not support referencing an entity by name."
)

// DeleteCommand is the function that executes the delete operation for the delete commands that do not work with names.
func DeleteCommand(id string, d cli.Discovery, client cli.Deleter, config commandConfig) error {
	err := checkCredentials(d, config.profile, config.componentName, config.url)
	if err != nil {
		return err
	}

	deleteId, err := uuid.Parse(id)
	if err != nil {
		return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not convert given id %q to UUID. This command does not support referencing an entity by name.", id)
	}
	printer := cli.GetObjectPrinter(config.output)
	return d.DeleteEntity(client, deleteId, printer)
}

// SearchDeleteCommand is the function that executes the delete operation for the delete commands that can also work with names.
func SearchDeleteCommand(id string, d cli.Discovery, client cli.SearchDeleter, config commandConfig) error {
	err := checkCredentials(d, config.profile, config.componentName, config.url)
	if err != nil {
		return err
	}

	printer := cli.GetObjectPrinter(config.output)
	return d.SearchDeleteEntity(client, id, printer)
}

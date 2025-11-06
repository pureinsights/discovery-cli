package commands

import (
	"github.com/google/uuid"
	"github.com/pureinsights/pdp-cli/internal/cli"
)

const (
	// LongGetNoNames is the message used in the Long field of the Get commands that do not support getting by name or using filters.
	LongGetNoNames string = "get is the command used to obtain Discovery %[2]s's %[1]ss. The user can send a UUID to get a specific %[1]s. If no UUID is given, then the command retrieves every %[1]s. The optional argument must be a UUID. This command does not support filters or referencing an entity by name."
)

// GetCommand is the function that executes the get operation for the get commands that do not work with names or filters.
func GetCommand(args []string, d cli.Discovery, client cli.Getter, config commandConfig) error {
	err := checkCredentials(d, config.profile, config.componentName, config.url)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		id, err := uuid.Parse(args[0])
		if err != nil {
			return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not convert given id %q to UUID. This command does not support filters or referencing an entity by name.", args[0])
		}
		printer := cli.GetObjectPrinter(config.output)
		err = d.GetEntity(client, id, printer)
		return err
	} else {
		printer := cli.GetArrayPrinter(config.output)
		err = d.GetEntities(client, printer)
		return err
	}
}

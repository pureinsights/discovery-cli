package cli

import (
	"github.com/google/uuid"
)

func GetCommand(args []string, d Discovery, client getter, profile, output, componentName, urlProperty, apiProperty string) error {
	err := checkCredentials(d, profile, componentName, urlProperty, apiProperty)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		id, err := uuid.Parse(args[0])
		if err != nil {
			return NewErrorWithCause(ErrorExitCode, err, "Could not convert given id %q to UUID. This command does not support filters or referencing an entity by name.", args[0])
		}
		printer := GetObjectPrinter(output)
		err = d.GetEntity(client, id, printer)
		return err
	} else {
		printer := GetArrayPrinter(output)
		err = d.GetEntities(client, printer)
		return err
	}
}

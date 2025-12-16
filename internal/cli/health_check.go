package cli

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type StatusChecker interface {
	StatusCheck() (gjson.Result, error)
}

// ExportEntitiesFromClient exports the entities from a single Discovery product and prints the acknowledgement message.
func (d discovery) StatusCheck(client StatusChecker, product string, printer Printer) error {
	result, err := client.StatusCheck()
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Discovery %s is not online.", product)
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.iostreams, result)
}

// BackupRestoreClientEntry is used to easily store the different backup and restore structs of the Discovery products.
type StatusCheckClientEntry struct {
	Name   string
	Client StatusChecker
}

// ExportEntitiesFromClients exports the entities from Discovery Core, Ingestion, and QueryFlow, writes the export files into the given path, and prints out the results.
func (d discovery) CheckStatusOfClients(clients []StatusCheckClientEntry, printer Printer) error {
	results := "{}"
	for _, entry := range clients {
		client := entry.Client
		status, err := client.StatusCheck()
		if err == nil {
			results, err = sjson.SetRaw(results, entry.Name, status.Raw)
			if err != nil {
				return NewErrorWithCause(ErrorExitCode, err, "Could not get the status of the Discovery products")
			}
		} else {
			results, err = sjson.Set(results, entry.Name, err.Error())
			if err != nil {
				return NewErrorWithCause(ErrorExitCode, err, "Could not get the status of the Discovery products")
			}
		}
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.iostreams, gjson.Parse(results))
}

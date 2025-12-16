package cli

import "github.com/tidwall/gjson"

type HealthChecker interface {
	HealthCheck() (gjson.Result, error)
}

// ExportEntitiesFromClient exports the entities from a single Discovery product and prints the acknowledgement message.
func (d discovery) HealthCheck(client HealthChecker, product string, printer Printer) error {
	result, err := client.HealthCheck()
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Discovery %s is not online.", product)
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.iostreams, result)
}

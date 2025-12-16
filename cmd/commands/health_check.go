package commands

import (
	"github.com/pureinsights/discovery-cli/internal/cli"
)

const (
	// LongExport is the message used in the Long field of the Export commands.
	LongHealthCheck string = "status is the command used to check of Discovery %s. If it is healthy, it should return a JSON with an \"UP\" status field"
)

// ExportCommand is the function that executes the export operation.
func HealthCheckCommand(d cli.Discovery, client cli.HealthChecker, product string, config commandConfig) error {
	err := CheckCredentials(d, config.profile, config.componentName, config.url)
	if err != nil {
		return err
	}

	printer := cli.GetObjectPrinter(config.output)

	return d.HealthCheck(client, product, printer)
}

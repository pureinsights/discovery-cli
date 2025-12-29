package commands

import (
	"github.com/pureinsights/discovery-cli/internal/cli"
)

const (
	// LongStatusCheck is the message used in the Long field of the Status commands.
	LongStatusCheck string = "status is the command used to check the status of Discovery %s. If it is healthy, it should return a JSON with an \"UP\" status field"
)

// StatusCheckCommand is the function that executes the status check operation.
func StatusCheckCommand(d cli.Discovery, client cli.StatusChecker, product string, config commandConfig) error {
	err := CheckCredentials(d, config.profile, config.componentName, config.url)
	if err != nil {
		return err
	}

	printer := cli.GetObjectPrinter(config.output)

	return d.StatusCheck(client, product, printer)
}

package commands

import "github.com/pureinsights/pdp-cli/internal/cli"

func ExportCommand(d cli.Discovery, client cli.BackupRestore, file string, config commandConfig) error {
	err := CheckCredentials(d, config.profile, config.componentName, config.url)
	if err != nil {
		return err
	}

	acknowledged, err := d.WriteExport(client, file)
	if err != nil {
		return err
	}

	printer := cli.GetObjectPrinter(d.Config().GetString("output"))
	if printer == nil {
		jsonPrinter := cli.JsonObjectPrinter(false)
		return jsonPrinter(*d.IOStreams(), acknowledged)
	} else {
		return printer(*d.IOStreams(), acknowledged)
	}
}

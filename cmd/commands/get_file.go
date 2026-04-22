package commands

import (
	"github.com/pureinsights/discovery-cli/internal/cli"
)

const (
	// LongGetFiles is the message used in the Long field of the Get command to list files from Core API.
	LongGetFiles string = "get is the command used to obtain the list of all Discovery %[2]s's %[1]ss."
	// LongGetFiles is the message used in the Long field of the Get command to list files from Core API.
	LongDownloadFiles string = "download is the command used to download Discovery %[2]s's %[1]ss. The user can send a key, representing a path, to get a specific %[1]s or multiple keys can be specify to download multiple %[1]ss."
)

// RetrieveCommand is the function that executes the retrive operation for the get commands that needs to download files.
func GetFilesCommand(args []string, d cli.Discovery, client cli.CoreFileController, config commandConfig) error {
	err := CheckCredentials(d, config.profile, config.componentName, config.url)
	if err != nil {
		return err
	}

	output := config.output
	if output == prettyJson {
		output = "json"
	}
	printer := cli.GetArrayPrinter(output)
	return d.GetFileList(client, printer)
}

// RetrieveCommand is the function that executes the retrive operation for the get commands that needs to download files.
func DownloadCommand(args []string, d cli.Discovery, client cli.CoreFileController, config commandConfig, outputPath string) error {
	err := CheckCredentials(d, config.profile, config.componentName, config.url)
	if err != nil {
		return err
	}

	printer := cli.GetObjectPrinter(config.output)
	if outputPath == "" {
		outputPath = "."
	}

	return d.GetFiles(client, args, outputPath, printer)
}
	
	
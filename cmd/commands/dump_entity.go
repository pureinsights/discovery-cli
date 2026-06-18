package commands

import (
	"github.com/pureinsights/discovery-cli/internal/cli"
)

const (
	LongDumpSearch string = "dump is the command used to scroll a bucket's content in Discovery %[2]s. The user must send a name or UUID to dump a specific bucket."
)

// SearchDumpCommand is the function that executes the dump operation for the dump commands that can also work with names.
func SearchDumpCommand(id string, d cli.Discovery, client cli.Searcher, contentProvider func(string) cli.StagingContentController, config cli.DumpConfig, printer cli.Printer) error {
	return d.SearchDumpBucket(client, contentProvider, id, config, printer)
}

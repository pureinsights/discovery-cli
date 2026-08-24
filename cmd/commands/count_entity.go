package commands

import (
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/tidwall/gjson"
)

const (
	LongCountSearch string = "count is the command used to count the number of records in a bucket in Discovery %[2]s. The user must send a name or UUID to count a specific bucket."
)

// SearchCountCommand is the function that executes the count operation for the count commands that can also work with names.
func SearchCountCommand(id string, d cli.Discovery, client cli.Searcher, contentProvider func(string) cli.StagingContentController, filters gjson.Result, printer cli.Printer) error {
	return d.SearchCountBucket(client, contentProvider, id, filters, printer)
}

package commands

import (
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/tidwall/gjson"
)

const (
	LongDumpSearch string = "dump is the command used to scroll a bucket's content in Discovery %[2]s. The user must send a name or UUID to dump a specific bucket."
)

func DumpCommand(id string, d cli.Discovery, client cli.SearchDumper, contentProvider func(string) cli.StagingContentController, file string, filters, projections gjson.Result, size *int, printer cli.Printer) error {
	return d.SearchDumpBucket(client, contentProvider, id, file, filters, projections, size, printer)
}

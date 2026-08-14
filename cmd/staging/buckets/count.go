package buckets

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

// NewCountCommand creates the bucket count command.
func NewCountCommand(d cli.Discovery) *cobra.Command {
	var filters string
	count := &cobra.Command{
		Use:   "count <bucketName>",
		Short: "The command that counts the number of records in a bucket in Discovery Staging.",
		Long:  fmt.Sprintf(commands.LongCountSearch, "bucket", "Staging"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, printer, err := prepareStagingCommand(d, cmd)
			if err != nil {
				return err
			}

			vpr := d.Config()
			stagingClient := discoveryPackage.NewStaging(vpr.GetString(profile+".staging_url"), vpr.GetString(profile+".staging_key"))

			return commands.SearchCountCommand(args[0], d, stagingClient.Buckets(), func(name string) cli.StagingContentController {
				return stagingClient.Content(name)
			}, gjson.Parse(filters), printer)
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Count the number of records in a bucket by name with filters
	discovery staging bucket count "my-bucket" -f '{"equals":{"field":"my-field","value":"my-value"}}'
	
	# Count the number of records in a bucket by UUID
	discovery staging bucket count ab0b4548-909d-4b23-aa62-69d8a6f8ed50'
	`,
	}

	count.Flags().StringVarP(&filters, "filter", "f", "", "the DSL containing the filters that will be applied to the scroll")

	return count
}

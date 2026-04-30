package buckets

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the bucket get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	var filters []string
	get := &cobra.Command{
		Use:   "get [<bucket>]",
		Short: "The command that obtains buckets from Discovery QueryFlow.",
		Long:  fmt.Sprintf(commands.LongGetSearch, "bucket", "QueryFlow"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key"))
			err := CheckCredentials(d, config.profile, config.componentName, config.url)
			if err != nil {
				return err
			}

			if len(args) > 0 {
				printer := cli.GetObjectPrinter(config.output)
				return d.SearchEntity(client, args[0], printer)
			} else {
				output := config.output
				if output == prettyJson {
					output = "json"
				}
				printer := cli.GetArrayPrinter(output)
				return d.GetEntities(client, printer)
			}
		},
		Args: cobra.MaximumNArgs(1),
		Example: `	# Get pipeline by name
	discovery queryflow pipeline get "my-pipeline"

	# Get pipelines using filters
	discovery queryflow pipeline get --filter label=A:A

	# Get all pipelines using the configuration in profile "cn"
	discovery queryflow pipeline get -p cn`,
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `apply filters in the format "filter=key:value". The available filters are:
- Label: The format is label={key}[:{value}], where the value is optional
- Type: The format is type={type}`)
	return get
}

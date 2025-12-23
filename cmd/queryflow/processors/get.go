package processors

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the processor get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	var filters []string
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains processors from Discovery QueryFlow.",
		Long:  fmt.Sprintf(commands.LongGetSearch, "processor", "QueryFlow"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key"))
			return commands.SearchCommand(args, d, queryflowClient.Processors(), commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"), &filters)
		},
		Args: cobra.MaximumNArgs(1),
		Example: `	# Get processor by name
	discovery queryflow processor get "my-processor"

	# Get processors using filters
	discovery queryflow processor get --filter label=A:A -f type=mongo

	# Get all processors using the configuration in profile "cn"
	discovery queryflow processor get -p cn`,
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `apply filters in the format "filter=key:value". The available filters are:
- Label: The format is label={key}[:{value}], where the value is optional
- Type: The format is type={type}`)
	return get
}

package processors

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the processor delete command
func NewDeleteCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "delete",
		Short: "The command that deletes processors from Discovery QueryFlow.",
		Long:  fmt.Sprintf(commands.LongDeleteNoNames, "processor", "QueryFlow"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key"))
			return commands.SearchDeleteCommand(args[0], d, queryflowClient.Processors(), commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"))
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Delete a processor by id
	discovery queryflow processor delete 189b3fa5-e011-43aa-ae57-f6e4a6f4b552
	{"acknowledged":true}

	# Delete a processor by name
	discovery queryflow processor delete processor1`,
	}
	return get
}

package pipelines

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the pipeline delete command
func NewDeleteCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "delete",
		Short: "The command that deletes pipelines from Discovery QueryFlow.",
		Long:  fmt.Sprintf(commands.LongDeleteNoNames, "pipeline", "QueryFlow"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key"))
			return commands.SearchDeleteCommand(args[0], d, queryflowClient.Pipelines(), commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"))
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Delete a pipeline by id
	discovery queryflow pipeline delete 04536687-f083-4353-8ecc-b7348e14b748

	# Delete a pipeline by name
	discovery queryflow pipeline delete "Search pipeline"`,
	}
	return get
}

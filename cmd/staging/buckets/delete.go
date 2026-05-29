package buckets

import (
	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the bucket delete command.
func NewDeleteCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "delete",
		Short: "The command that deletes buckets from Discovery Staging.",
		Long:  "delete is the command used to delete Discovery Staging's buckets. The user must send the bucket's name as a required argument.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			stagingClient := discoveryPackage.NewStaging(vpr.GetString(profile+".staging_url"), vpr.GetString(profile+".staging_key"))
			return commands.SearchDeleteCommand(args[0], d, stagingClient.Buckets(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Staging", "staging_url"))
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Delete a bucket by id
	discovery staging bucket delete ea02fc14-f07b-49f2-b185-e9ceaedcb367

	# Delete a bucket by name
	discovery staging bucket delete my-bucket`,
	}
	return get
}

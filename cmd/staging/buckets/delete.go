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
		Long:  "delete is the command used to delete Discovery Staging's buckets. The user must send the bucket's name.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			err = commands.CheckCredentials(d, profile, "Staging", "staging_url")
			if err != nil {
				return err
			}

			vpr := d.Config()

			stagingClient := discoveryPackage.NewStaging(vpr.GetString(profile+".staging_url"), vpr.GetString(profile+".staging_key"))
			printer := cli.GetObjectPrinter(vpr.GetString("output"))

			return d.DeleteBucket(stagingClient.Buckets(), args[0], printer)
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Delete a bucket by name
	discovery staging bucket delete my-bucket`,
	}
	return get
}

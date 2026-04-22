package file

import (
	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the file delete command.
func NewDeleteCommand(d cli.Discovery) *cobra.Command {
	delete := &cobra.Command{
		Use:   "delete <file>",
		Short: "The command that deletes files from Discovery Core's object storage.",
		Long:  "delete is the command used to delete files from Discovery Core's object storage. The user sends the file's key and Discovery returns an acknowledgement message.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			err = commands.CheckCredentials(d, profile, "Core", "core_url")
			if err != nil {
				return err
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			printer := cli.GetObjectPrinter(vpr.GetString("output"))
			return d.DeleteFile(coreClient.Files(), args[0], printer)
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Delete a file by its key
	discovery core file delete "my-file"`,
	}

	return delete
}

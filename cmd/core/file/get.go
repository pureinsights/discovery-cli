package file

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the credential get command.
func NewGetCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains the list of all files from Discovery Core.",
		Long:  fmt.Sprintf(commands.LongGetFiles, "file", "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.GetFilesCommand(args, d, coreClient.Files(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"))
		},
		Args: cobra.NoArgs,
		Example: `	# Get the list of all files
	discovery core file get

	# Get the list of all files using the configuration in profile "cn"
	discovery core file get -p "cn"`,

	}
	return get
}

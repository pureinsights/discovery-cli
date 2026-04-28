package file

import (
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
		Long:  "get is the command used to obtain the list of all Discovery Core's files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))

			err = commands.CheckCredentials(d, profile, "Core", "core_url")
			if err != nil {
				return err
			}

			printer := cli.GetArrayPrinter("json")
			return d.GetFileList(coreClient.Files(), printer)
		},
		Args: cobra.NoArgs,
		Example: `	# Get the list of all files
	discovery core file get

	# Get the list of all files using the configuration in profile "cn"
	discovery core file get -p "cn"`,
	}
	return get
}

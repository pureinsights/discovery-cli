package file

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewDownloadCommand creates the file download command.
func NewDownloadCommand(d cli.Discovery) *cobra.Command {
	var output string
	download := &cobra.Command{
		Use:   "download [<file>]...",
		Short: "The command that obtains files from Discovery Core.",
		Long:  fmt.Sprintf(commands.LongDownloadFiles, "file", "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.DownloadCommand(args, d, coreClient.Files(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"),output)
		},
		Args: cobra.MinimumNArgs(1),
		Example: `	# Download file by name
	discovery core file download "my_file.json"

	# Output file to different directory
	discovery core file download "my_file.json" -o "./my_directory"

	# Download file by nested path
	discovery core file download "my_directory/my_file.json"

	# Download multiple files by specifying nested paths or names
	discovery core file download "my_directory/my_file.json" "my_other_file.json"`,
	}
	download.Flags().StringVarP(&output, "output", "o", "", "the path/directory to download the file to")
	return download
}

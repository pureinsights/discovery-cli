package file

import (
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
		Long:  "download is the command used to download Discovery Core's files. The user can send a key, representing a path, to get a specific file or multiple keys can be specified to download multiple files. When specifying multiple keys, downloads are attempted sequentially. If you specify three keys and the second one fails, only the first file will be downloaded and the remaining downloads (second and third) will fail. You can specify an output directory using the --output flag. Both absolute and relative paths are supported. If the specified directory does not exist, it will be created. Any required nested directories will also be created.",
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

			printer := cli.GetObjectPrinter(vpr.GetString("output"))
			if output == "" {
				output = "."
			}

			return d.GetFiles(coreClient.Files(), args, output, printer)
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

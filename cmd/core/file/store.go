package file

import (
	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStoreCommand creates the file store command.
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var recursive bool
	store := &cobra.Command{
		Use:   "store",
		Short: "The command to store files inside Discovery Core.",
		Long:  "store is the command used to store files inside Discovery Core",
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

			return d.StoreFiles(coreClient.Files(),args[0], recursive, printer)
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Store the file using the name as key
	discovery core file store "my_file.json"

	# If the input is a path, store each file
	discovery core file store .

	# With the recursive flag, go recursively and store each file using the relative path as key
	discovery core file store "my_path/" --recursive`,
	}
	store.Flags().BoolVarP(&recursive, "recursive", "r", false, "whether to recursively store every file in the specified key/path")
	return store
}

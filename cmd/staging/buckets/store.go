package buckets

import (
	"os"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

// NewStoreCommand creates the bucket store command.
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var data string
	store := &cobra.Command{
		Use:   "store <bucketName> [configFile]",
		Short: "The command that stores buckets to Discovery Staging.",
		Long:  "store is the command used to create and update buckets in the Discovery Staging Repository. The bucket's name is sent as the mandatory first argument. The creation options, like the indices and bucket configuration, can be sent either through the optional second argument, which contains the name of the file with the information, or through the --data flag as a JSON string. The --data flag and the file name argument are mutually exclusive. When the bucket already exists, the command will try to modify its indices by updating them and deleting the ones no longer needed.",
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
			options := gjson.Result{}
			if cmd.Flags().Changed("data") {
				if len(args) > 1 {
					return cli.NewError(cli.ErrorExitCode, "The data flag can only have the bucket name argument.")
				}
				options = gjson.Parse(data)
			}
			if len(args) > 1 {
				jsonBytes, err := os.ReadFile(args[1])
				if err != nil {
					err = cli.NormalizeReadFileError(args[1], err)
					return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not read file %q", args[1])
				}
				options = gjson.ParseBytes(jsonBytes)
			}

			return d.StoreBucket(stagingClient.Buckets(), args[0], options, printer)
		},
		Args: cobra.RangeArgs(1, 2),
		Example: `	# Store a bucket with the JSON configuration in a file
	discovery staging bucket store my-bucket configFile.json

	# Store a bucket with the JSON configuration in the data flag
	discovery staging bucket store my-bucket --data '{"indices":[{"name":"myIndexA","fields":[{"fieldName":"ASC"}],"unique":false},{"name":"myIndexB","fields":[{"fieldName2":"DESC"}],"unique":false}],"config":{}}`,
	}

	store.Flags().StringVarP(&data, "data", "d", "", "the JSON with the configuration of the bucket")

	return store
}

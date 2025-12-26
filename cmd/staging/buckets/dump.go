package buckets

import (
	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

// NewDumpCommand creates the bucket dump command.
func NewDumpCommand(d cli.Discovery) *cobra.Command {
	var filters string
	var projections string
	var max int
	dump := &cobra.Command{
		Use:   "dump <bucketName> [configFile]",
		Short: "The command that dumps buckets to Discovery Staging.",
		Long:  "dump is the command used to create and update buckets in the Discovery Staging Repository. The bucket's name is sent as the mandatory first argument. The creation options, like the indices and bucket configuration, can be sent either through the optional second argument, which contains the name of the file with the information, or through the --data flag as a JSON string. The --data flag and the file name argument are mutually exclusive. When the bucket already exists, the command will try to modify its indices by updating them and deleting the ones no longer needed.",
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

			output := vpr.GetString("output")
			if output == "pretty-json" {
				output = "json"
			}
			printer := cli.GetArrayPrinter(output)

			var size *int
			if cmd.Flags().Changed("max") {
				if max < 1 {
					return cli.NewError(cli.ErrorExitCode, "The size flag can only be greater than or equal to 1.")
				}
				size = &max
			}

			return d.DumpBucket(stagingClient.Content(args[0]), args[0], gjson.Parse(filters), gjson.Parse(projections), size, printer)
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Dump a bucket with the JSON configuration in a file
	discovery staging bucket dump my-bucket configFile.json

	# Dump a bucket with the JSON configuration in the data flag
	discovery staging bucket dump my-bucket --data '{"indices":[{"name":"myIndexA","fields":[{"fieldName":"ASC"}],"unique":false},{"name":"myIndexB","fields":[{"fieldName2":"DESC"}],"unique":false}],"config":{}}`,
	}

	dump.Flags().StringVarP(&filters, "filter", "f", "", "the DSL containing the filters that will be applied to the scroll")
	dump.Flags().StringVar(&projections, "projection", "", "the DSL containing the fields that will be included and excluded in the records that will be retrieved from the bucket")
	dump.Flags().IntVar(&max, "max", -1, "the size of the pages that will be used when retrieving the records")

	return dump
}

package buckets

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the pipeline get command.
func NewGetCommand(d cli.Discovery) *cobra.Command {
	var filters []string
	get := &cobra.Command{
		Use:   "get <bucketName>",
		Short: "The command that obtains buckets from Discovery Staging.",
		Long:  fmt.Sprintf(commands.LongGetSearch, "bucket", "Staging"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			stagingClient := discoveryPackage.NewStaging(vpr.GetString(profile+".staging_url"), vpr.GetString(profile+".staging_key"))
			return commands.SearchCommand(args, d, stagingClient.Buckets(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Staging", "staging_url"), &filters)
		},
		Args: cobra.MaximumNArgs(1),
		Example: `	# Get bucket by name
	discovery staging bucket get "my-bucket"

	# Get buckets using filters
	discovery staging bucket get --filter label=A:A

	# Get all buckets using the configuration in profile "cn"
	discovery staging buckets get -p cn`,
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `apply filters in the format "filter=key:value". The available filter is:
- Label: The format is label={key}[:{value}], where the value is optional
- Type: The format is type={type}`)
	return get
}

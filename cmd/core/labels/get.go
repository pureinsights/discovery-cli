package labels

import (
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the label get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains labels from Discovery Core.",
		Long:  "get is the command used to obtain Discovery Core's labels. The user can send a UUID to get a specific label. If no UUID is given, then the command retrieves every label. The optional argument must be a UUID. This command does not support filters or referencing an entity by name.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return err
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return cli.GetCommand(args, d, coreClient.Labels(), cli.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url", "core_key"))
		},
	}
	return get
}

package secrets

import (
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the secret get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains secrets from Discovery Core.",
		Long:  "get is the command used to obtain Discovery Core's secrets. The user can send a UUID to get a specific secret. If no UUID is given, then the command retrieves every secret. The optional argument must be a UUID. This command does not support filters or referencing an entity by name.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return cli.GetCommand(args, d, coreClient.Secrets(), cli.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url", "core_key"))
		},
		Args: cobra.MaximumNArgs(1),
	}
	return get
}

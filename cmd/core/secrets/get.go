package secrets

import (
	"github.com/google/uuid"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the secret get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	missingConfig := "The Discovery Core %s is missing for profile %q.\nTo set the %s for the Discovery Core API, run any of the following commands:\n      discovery config  --profile {profile}\n      discovery core config --profile {profile}"
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains secrets from Discovery Core.",
		Long:  "get is the command used to obtain Discovery Core's secrets. The user can send a UUID to get a specific secret. If no UUID is given, then the command retrieves every secret. The optional argument must be a UUID. This command does not support filters or referencing an entity by name.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return err
			}

			vpr := d.Config()
			switch {
			case !vpr.IsSet(profile + ".core_url"):
				return cli.NewError(cli.ErrorExitCode, missingConfig, "URL", profile, "URL")
			case !vpr.IsSet(profile + ".core_key"):
				return cli.NewError(cli.ErrorExitCode, missingConfig, "API key", profile, "API key")
			}

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			if len(args) > 0 {
				id, err := uuid.Parse(args[0])
				if err != nil {
					return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not convert given id %q to UUID. This command does not support filters or referencing an entity by name.", args[0])
				}
				printer := cli.GetObjectPrinter(vpr.GetString("output"))
				err = d.GetEntity(coreClient.Secrets(), id, printer)
				return err
			} else {
				printer := cli.GetArrayPrinter(vpr.GetString("output"))
				err = d.GetEntities(coreClient.Secrets(), printer)
				return err
			}
		},
		Args: cobra.MaximumNArgs(1),
	}
	return get
}

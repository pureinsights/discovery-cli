package labels

import (
	"github.com/google/uuid"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the label get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	missingConfig := "The Discovery Core %s is missing for profile %q.\nTo set the %s for the Discovery Core API, run any of the following commands:\n      discovery config  --profile {profile}\n      discovery core config --profile {profile}"
	get := &cobra.Command{
		Use:   "get",
		Short: "Print Discovery Core's configuration",
		Long:  "get is the command used to obtain Discovery Core's configuration for a given profile. If the API keys are sensitive, the `sensitive` flag can be set to true in order to obfuscate them before printing them out. If a configuration property was not set, it is not displayed.",
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
				err = d.GetEntity(coreClient.Labels(), id, printer)
				return err
			} else {
				printer := cli.GetArrayPrinter(vpr.GetString("output"))
				err = d.GetEntities(coreClient.Labels(), printer)
				return err
			}
		},
		Args: cobra.MaximumNArgs(1),
	}
	get.Flags().BoolP("sensitive", "s", true, "--sensitive=true")
	return get
}

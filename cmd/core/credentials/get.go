package credentials

import (
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the credential get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	missingConfig := "The Discovery Core %s is missing for profile %q.\nTo set the %s for the Discovery Core API, run any of the following commands:\n      discovery config  --profile {profile}\n      discovery core config --profile {profile}"
	var filters []string
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains credentials from Discovery Core.",
		Long:  "",
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

				printer := cli.GetObjectPrinter(vpr.GetString("output"))
				err = d.SearchEntity(coreClient.Credentials(), args[0], printer)
				return err
			} else if len(filters) > 0 {
				printer := cli.GetArrayPrinter(vpr.GetString("output"))
				filter, err := cli.BuildEntitiesFilter(filters)
				if err != nil {
					return err
				}

				err = d.SearchEntities(coreClient.Credentials(), filter, printer)
				return err
			} else {
				printer := cli.GetArrayPrinter(vpr.GetString("output"))
				err = d.GetEntities(coreClient.Credentials(), printer)
				return err
			}
		},
		Args: cobra.MaximumNArgs(1),
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, "Apply filters in the format \"filter=key:value\"")
	return get
}

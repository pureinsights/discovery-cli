package servers

import (
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the server get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	missingConfig := "The Discovery Core %s is missing for profile %q.\nTo set the %s for the Discovery Core API, run any of the following commands:\n      discovery config  --profile {profile}\n      discovery core config --profile {profile}"
	var filters []string
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains servers from Discovery Core.",
		Long:  "get is the command used to obtain Discovery Core's servers. The user can send a name or UUID to get a specific server. If no argument is given, then the command retrieves every server. The command also supports filters with the flag --filter followed by the filter in the format filter=key:value.",
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
			return cli.SearchCommand(args, d, coreClient.Servers(), cli.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url", "core_key"), &filters)
		},
		Args: cobra.MaximumNArgs(1),
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `Apply filters in the format "filter=key:value". The available filters are:
- Label: The format is label={key}[:{value}], where the value is optional.
- Type: The format is type={type}.`)
	return get
}

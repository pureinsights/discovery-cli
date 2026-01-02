package servers

import (
	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewPingCommand creates the server ping command.
func NewPingCommand(d cli.Discovery) *cobra.Command {
	ping := &cobra.Command{
		Use:   "ping <server>",
		Short: "The command that pings servers from Discovery Core.",
		Long:  "ping is the command used to check if a server in Discovery Core is reachable. If it is, it should return an acknowledgement message. Some type of servers cannot be pinged, like OpenAI servers. Consult the Discovery documentation for more information.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			err = commands.CheckCredentials(d, profile, "Core", "core_url")
			if err != nil {
				return err
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			printer := cli.GetObjectPrinter(vpr.GetString("output"))
			return d.PingServer(coreClient.Servers(), args[0], printer)
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Ping server by name
	discovery core server ping "my-server"

	# Ping server by id
	discovery core server ping 21029da3-041c-43b5-a67e-870251f2f6a6`,
	}

	return ping
}

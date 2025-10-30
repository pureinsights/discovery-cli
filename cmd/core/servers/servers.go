package servers

import (
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewServerCommand creates the server command.
func NewServerCommand(d cli.Discovery) *cobra.Command {
	server := &cobra.Command{
		Use:   "server [subcommand] [flags]",
		Short: "The command to interact with Discovery Core's servers.",
	}

	server.AddCommand(NewGetCommand(d))
	server.AddCommand(NewStoreCommand(d))

	return server
}

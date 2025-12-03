package secrets

import (
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewSecretCommand creates the secret command.
func NewSecretCommand(d cli.Discovery) *cobra.Command {
	secret := &cobra.Command{
		Use:   "secret [subcommand] [flags]",
		Short: "The command to interact with Discovery Core's secrets.",
	}

	secret.AddCommand(NewGetCommand(d))
	secret.AddCommand(NewStoreCommand(d))
	secret.AddCommand(NewDeleteCommand(d))

	return secret
}

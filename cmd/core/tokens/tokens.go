package tokens

import (
	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/spf13/cobra"
)

// NewAccessTokenCommand creates the access token command.
func NewTokenCommand(d cli.Discovery) *cobra.Command {
	token := &cobra.Command{
		Use:   "token [subcommand] [flags]",
		Short: "The command to interact with Discovery Core's access tokens.",
	}

	token.AddCommand(NewGetCommand(d))
	token.AddCommand(NewCreateCommand(d))
	token.AddCommand(NewUpdateCommand(d))
	token.AddCommand(NewEnableCommand(d))
	token.AddCommand(NewDisableCommand(d))
	token.AddCommand(NewRevokeCommand(d))
	token.AddCommand(NewRenewCommand(d))

	return token
}

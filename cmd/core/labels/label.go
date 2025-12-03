package labels

import (
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewLabelCommand creates the label command.
func NewLabelCommand(d cli.Discovery) *cobra.Command {
	label := &cobra.Command{
		Use:   "label [subcommand] [flags]",
		Short: "The command to interact with Discovery Core's labels.",
	}

	label.AddCommand(NewGetCommand(d))
	label.AddCommand(NewStoreCommand(d))

	return label
}

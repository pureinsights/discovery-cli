package buckets

import (
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewBucketCommand creates the bucket command.
func NewBucketCommand(d cli.Discovery) *cobra.Command {
	bucket := &cobra.Command{
		Use:   "bucket [subcommand] [flags]",
		Short: "The command to interact with Discovery Staging's buckets.",
	}

	bucket.AddCommand(NewStoreCommand(d))
	bucket.AddCommand(NewDeleteCommand(d))
	bucket.AddCommand(NewDumpCommand(d))

	return bucket
}

package seed_schedules

import (
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewSeedScheduleCommand creates the seed-schedule command.
func NewSeedScheduleCommand(d cli.Discovery) *cobra.Command {
	seedSchedule := &cobra.Command{
		Use:   "seed-schedule [subcommand] [flags]",
		Short: "The command to interact with Discovery Ingestion's seed schedules.",
	}

	seedSchedule.AddCommand(NewGetCommand(d))
	seedSchedule.AddCommand(NewStoreCommand(d))
	seedSchedule.AddCommand(NewDeleteCommand(d))

	return seedSchedule
}

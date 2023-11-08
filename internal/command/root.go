package command

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/internal/gitlab"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

var rootCmd = &cobra.Command{
	Use:           "anka-gle",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.AddCommand(cleanupCommand, prepareCommand, runCommand, configCommand)
}

func Execute(ctx context.Context) error {

	if debug, ok, err := gitlab.GetBoolVar(gitlab.VarDebug); ok {
		if err != nil {
			return fmt.Errorf("failed to read debug variable: %w", err)
		}
		log.SetDebug(debug)
	}

	return rootCmd.ExecuteContext(ctx)
}

package commands

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/internal/commands/cleanup"
	"veertu.com/anka-cloud-gitlab-executor/internal/commands/config"
	"veertu.com/anka-cloud-gitlab-executor/internal/commands/prepare"
	"veertu.com/anka-cloud-gitlab-executor/internal/commands/run"
	"veertu.com/anka-cloud-gitlab-executor/internal/env"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

var rootCmd = &cobra.Command{
	Use:           "anka-gle",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.AddCommand(cleanup.Command)
	rootCmd.AddCommand(prepare.Command)
	rootCmd.AddCommand(run.Command)
	rootCmd.AddCommand(config.Command)
}

func Execute(ctx context.Context) {
	_, ok := env.Get(env.AnkaVar("DEBUG"))
	if ok {
		log.SetDebug(true)
	}

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}
}

package commands

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/internal/commands/cleanup"
	"veertu.com/anka-cloud-gitlab-executor/internal/commands/config"
	"veertu.com/anka-cloud-gitlab-executor/internal/commands/prepare"
	"veertu.com/anka-cloud-gitlab-executor/internal/commands/run"
	"veertu.com/anka-cloud-gitlab-executor/internal/gitlab"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

var rootCmd = &cobra.Command{
	Use: "anka-gle",
}

func init() {
	rootCmd.AddCommand(cleanup.Command)
	rootCmd.AddCommand(prepare.Command)
	rootCmd.AddCommand(run.Command)
	rootCmd.AddCommand(config.Command)
}

func Execute(ctx context.Context) {
	if _, err := gitlab.GetAnkaCloudEnvVar("DEBUG"); err == nil {
		log.SetDebug(true)
	}

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

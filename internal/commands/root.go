package commands

import (
	"os"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/internal/commands/cleanup"
	"veertu.com/anka-cloud-gitlab-executor/internal/commands/config"
	"veertu.com/anka-cloud-gitlab-executor/internal/commands/prepare"
	"veertu.com/anka-cloud-gitlab-executor/internal/commands/run"
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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

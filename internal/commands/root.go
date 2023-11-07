package commands

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/internal/env"
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

func Execute(ctx context.Context) {
	_, ok := os.LookupEnv(env.VarDebug)
	if ok {
		log.SetDebug(true)
	}

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}
}

package command

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/internal/gitlab"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

type contextKey string

var rootCmd = &cobra.Command{
	Use:           "anka-gle",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.AddCommand(cleanupCommand, prepareCommand, runCommand, configCommand)
}

func Execute(ctx context.Context) error {

	env, err := gitlab.InitEnv()
	if err != nil {
		return fmt.Errorf("failed to initialize environment: %s", err)
	}

	log.SetDebug(env.Debug)

	return rootCmd.ExecuteContext(context.WithValue(ctx, contextKey("env"), env))
}

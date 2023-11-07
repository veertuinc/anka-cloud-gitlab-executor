package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"veertu.com/anka-cloud-gitlab-executor/internal/commands"
	"veertu.com/anka-cloud-gitlab-executor/internal/gitlab"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

var (
	buildFailureExitCode  = 1
	systemFailureExitCode = 2
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	if buildFailureExitCodeEnvVar, ok, err := gitlab.GetIntVar(gitlab.VarBuildFailureExitCode); ok {
		if err != nil {
			log.Printf("failed reading build failure exit code: %s", err)
			os.Exit(buildFailureExitCode)
		}
		buildFailureExitCode = buildFailureExitCodeEnvVar
	}

	if systemFailureExitCodeEnvVar, ok, err := gitlab.GetIntVar(gitlab.VarSystemFailureExitCode); ok {
		if err != nil {
			log.Printf("failed reading system failure exit code: %s", err)
			os.Exit(buildFailureExitCode)
		}
		systemFailureExitCode = systemFailureExitCodeEnvVar
	}

	if err := commands.Execute(ctx); err != nil {
		log.Printf("error: %s", err)
		if errors.Is(err, gitlab.ErrTransient) {
			os.Exit(systemFailureExitCode)
		}
		os.Exit(buildFailureExitCode)
	}
}

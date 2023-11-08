package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"veertu.com/anka-cloud-gitlab-executor/internal/command"
	"veertu.com/anka-cloud-gitlab-executor/internal/gitlab"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

var (
	buildFailureExitCode  = 1
	systemFailureExitCode = 2
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	if buildFailureExitCodeEnvVar, ok, err := gitlab.GetIntVar(gitlab.VarBuildFailureExitCode); ok {
		if err != nil {
			log.Printf("failed reading build failure exit code: %s", err)
			return buildFailureExitCode
		}
		buildFailureExitCode = buildFailureExitCodeEnvVar
	}

	if systemFailureExitCodeEnvVar, ok, err := gitlab.GetIntVar(gitlab.VarSystemFailureExitCode); ok {
		if err != nil {
			log.Printf("failed reading system failure exit code: %s", err)
			return buildFailureExitCode
		}
		systemFailureExitCode = systemFailureExitCodeEnvVar
	}

	if err := command.Execute(ctx); err != nil {
		log.Printf("error: %s", err)
		if errors.Is(err, gitlab.ErrTransient) {
			return systemFailureExitCode
		}
		return buildFailureExitCode
	}
	return 0
}

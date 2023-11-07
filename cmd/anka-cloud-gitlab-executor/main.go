package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"veertu.com/anka-cloud-gitlab-executor/internal/commands"
	"veertu.com/anka-cloud-gitlab-executor/internal/gitlab"
)

var (
	buildFailureExitCode  = 1
	systemFailureExitCode = 2
)

func main() {
	// TODO: handle Exit Codes

	ctx, cancel := context.WithCancel(context.Background())

	if buildFailureExitCodeEnvVar, ok, err := gitlab.GetIntVar(gitlab.VarBuildFailureExitCode); ok && err == nil {
		buildFailureExitCode = buildFailureExitCodeEnvVar
	}

	if systemFailureExitCodeEnvVar, ok, err := gitlab.GetIntVar(gitlab.VarBuildFailureExitCode); ok && err == nil {
		systemFailureExitCode = systemFailureExitCodeEnvVar
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
		os.Exit(buildFailureExitCode)
	}()

	if err := commands.Execute(ctx); err != nil {
		log.Printf("error: %s", err)
		if errors.Is(err, gitlab.ErrTransient) {
			os.Exit(systemFailureExitCode)
		}
		os.Exit(buildFailureExitCode)
	}

	os.Exit(0)
}

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/command"
	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/gitlab"
	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/log"
	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/version"
)

var (
	buildFailureExitCode  = 1
	systemFailureExitCode = 2
)

const (
	varBuildFailureExitCode  = "BUILD_FAILURE_EXIT_CODE"
	varSystemFailureExitCode = "SYSTEM_FAILURE_EXIT_CODE"
)

func main() {
	os.Exit(run())
}

func run() int {

	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("%s", version.Get())
		return 0
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	if buildFailureExitCodeEnvVar, ok, err := gitlab.GetIntEnvVar(varBuildFailureExitCode); ok {
		if err != nil {
			log.Printf("failed reading build failure exit code: %s", err)
			return buildFailureExitCode
		}
		buildFailureExitCode = buildFailureExitCodeEnvVar
	}

	if systemFailureExitCodeEnvVar, ok, err := gitlab.GetIntEnvVar(varSystemFailureExitCode); ok {
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

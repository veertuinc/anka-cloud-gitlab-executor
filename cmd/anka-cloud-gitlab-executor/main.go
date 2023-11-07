package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"veertu.com/anka-cloud-gitlab-executor/internal/commands"
)

func main() {
	// TODO: handle Exit Codes

	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
		os.Exit(1)
	}()

	commands.Execute(ctx)
}

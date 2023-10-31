package main

import (
	"context"

	"veertu.com/anka-cloud-gitlab-executor/internal/commands"
)

func main() {
	// TODO: Add signal handling
	// TODO: handle Exit Codes
	commands.Execute(context.Background())
}

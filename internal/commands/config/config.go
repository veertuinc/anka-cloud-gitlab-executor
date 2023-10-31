package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:  "config",
	RunE: execute,
}

func execute(cmd *cobra.Command, args []string) error {
	// TODO: Handle properly returning configuration

	fmt.Fprintf(os.Stdout, "{}")
	return nil
}

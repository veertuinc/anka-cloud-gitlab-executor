package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/gitlab"
	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/log"
	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/version"
)

var configCommand = &cobra.Command{
	Use: "config",
	RunE: func(cmd *cobra.Command, args []string) error {
		env, ok := cmd.Context().Value(contextKey("env")).(gitlab.Environment)
		if !ok {
			return fmt.Errorf("failed to get environment from context")
		}

		return executeConfig(cmd.Context(), env)
	},
}

type output struct {
	BuildsDir       string            `json:"builds_dir,omitempty"`
	CacheDir        string            `json:"cache_dir,omitempty"`
	BuildsDirShared bool              `json:"builds_dir_is_shared"`
	Driver          driver            `json:"driver"`
	JobEnv          map[string]string `json:"job_env,omitempty"`
}

type driver struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func executeConfig(ctx context.Context, env gitlab.Environment) error {
	log.SetOutput(os.Stderr)
	log.Println("running config stage")

	output := output{
		BuildsDir:       fmt.Sprintf("/tmp/build/%s", env.GitlabJobId),
		CacheDir:        fmt.Sprintf("/tmp/cache/%s", env.GitlabJobId),
		BuildsDirShared: false,
		Driver: driver{
			Name:    "Anka Cloud Gitlab Executor",
			Version: version.Get(),
		},
	}
	jsonBytes, err := json.MarshalIndent(&output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to JSON marshal output %+v: %w", output, err)
	}

	fmt.Fprintln(os.Stdout, string(jsonBytes))
	return nil
}

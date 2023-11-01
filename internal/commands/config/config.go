package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"veertu.com/anka-cloud-gitlab-executor/internal/gitlab"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

var Command = &cobra.Command{
	Use:  "config",
	RunE: execute,
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

func execute(cmd *cobra.Command, args []string) error {
	log.SetOutput(os.Stderr)
	log.Println("Running config stage")

	jobId, err := gitlab.GetGitlabEnvVar("CI_JOB_ID")
	if err != nil {
		return err
	}
	output := output{
		BuildsDir:       fmt.Sprintf("/tmp/build/%s", jobId),
		CacheDir:        fmt.Sprintf("/tmp/cache/%s", jobId),
		BuildsDirShared: false,
		Driver: driver{
			Name:    "Anka Cloud Gitlab Executor",
			Version: "1.0.0", // TODO incorporate versioning into build scripts and here also
		},
	}
	jsonBytes, err := json.MarshalIndent(&output, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, string(jsonBytes))
	return nil
}

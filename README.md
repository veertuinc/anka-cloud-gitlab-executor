# Gitlab Custom Executor

A [Gitlab Runner Custom Executor](https://docs.gitlab.com/runner/executors/custom.html) utilizing [Anka Build Cloud](https://veertu.com/anka-build/) to run your Gitlab jobs on macOS VMs.

### Pre-requirements:
1. Install and Register (at least one) self-managed Gitlab Runner. See [here](https://docs.gitlab.com/runner/install/index.html) for more info.
2. Have a live and running [Anka Build Cloud](https://veertu.com/anka-build/)

### Configuration:
1. Download the binary to the same machine your Gitlab Runner is running on.
2. Add the following `[runners.custom]` block to Gitlab Runner configuration:
    > By default, runner config is at `~/.gitlab-runner/config.toml`
    ```
    [runners.custom]
        config_exec = "PATH_TO_BINARY_HERE"
        config_args = ["config"]
        prepare_exec = "PATH_TO_BINARY_HERE"
        prepare_args = ["prepare"]
        run_exec = "PATH_TO_BINARY_HERE"
        run_args = ["run"]
        cleanup_exec = "PATH_TO_BINARY_HERE"
        cleanup_args = ["cleanup"]
    ```
3. Add `executor = "custom"` to `[[runners]]` block

Check out the [full configuration spec](https://docs.gitlab.com/runner/executors/custom.html#configuration) for more info

### Usage:
The following variables are required:
| Variable name | Type | Description |
| ------------- | ---- | ----------- |
| ANKA_CLOUD_CONTROLLER_URL | string | Anka Build Cloud's Controller URL. Inlcuding "http[s]" prefix. Port optional |
| ANKA_CLOUD_TEMPLATE_ID | string | VM Template ID to use. Must exist on the Registry |

Example pipeline:
```
variables:
  ANKA_CLOUD_CONTROLLER_URL: "http://anka.contoller:8090"
  ANKA_CLOUD_TEMPLATE_ID: "8c592f53-65a4-444e-9342-79d3ff07837c"
  
build-job:
  stage: build
  tags:
    - anka_runner
  script:
    - uname -a
```
# Anka Cloud Gitlab Executor

A [Gitlab Runner Custom Executor](https://docs.gitlab.com/runner/executors/custom.html) utilizing [Anka Build Cloud](https://veertu.com/anka-build/) to run your Gitlab jobs on macOS VMs.

## Pre-requirements:
1. Install and Register (at least one) self-managed Gitlab Runner. See [here](https://docs.gitlab.com/runner/install/index.html) for more info.
2. An active [Anka Build Cloud](https://veertu.com/anka-build/)

## Configuration:
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
4. (Optional) [Tag](https://docs.gitlab.com/ee/ci/runners/configure_runners.html#use-tags-to-control-which-jobs-a-runner-can-run) your Anka runners

Check out the [full configuration spec](https://docs.gitlab.com/runner/executors/custom.html#configuration) for more info

### Variables:
> All of our variables are parsed as strings, since Gitlab is stating: "To ensure consistent behavior, you should always put variable values in single or double quotes." [Reference](https://docs.gitlab.com/ee/ci/variables/)

| Variable name | Required | Description |
| ------------- |:--------:| ----------- |
| ANKA_CLOUD_CONTROLLER_URL |     ✅ | Anka Build Cloud's Controller URL. Inlcuding `http[s]` prefix. Port optional |
| ANKA_CLOUD_TEMPLATE_ID | ✅ | VM Template ID to use. Must exist on the Registry and have SSH port forwarding |
| ANKA_CLOUD_TEMPLATE_TAG | ❌ | Template tag to use |
| ANKA_CLOUD_NODE_ID | ❌ | Run VM on this specific node |
| ANKA_CLOUD_PRIORITY | ❌ | Priority in range 1-10000 (lower is more urgent) |
| ANKA_CLOUD_NODE_GROUP_ID | ❌ | Run the VM on a specific Node Group, by Group ID |
| ANKA_CLOUD_SKIP_TLS_VERIFY | ❌ | If Controller is using a self-signed cert, this allows the Runner to skip validation of the certificate |
| ANKA_CLOUD_CA_CERT_PATH | ❌ | If Controller is using a self-signed cert, CA file can be passed in for the runner to use when communicating with Controller. **_The path is accessed locally by the Runner_** |
| ANKA_CLOUD_CLIENT_CERT_PATH | ❌ | If Client Cert Authentication is enabled, this is the path for the Certificate. **_The path is accessed locally by the Runner_** |
| ANKA_CLOUD_CLIENT_CERT_KEY_PATH | ❌ | If Client Cert Authentication is enabled, this is the path for the Key. **_The path is accessed locally by the Runner_** |

Example basic pipeline:
```
variables:
  ANKA_CLOUD_CONTROLLER_URL: "http://anka.contoller:8090"
  ANKA_CLOUD_TEMPLATE_ID: "8c592f53-65a4-444e-9342-79d3ff07837c"

build-job:
  stage: build
  tags:
    - anka_runner
  script:
    - echo "building"

test-job:
  stage: test
  tags:
    - anka_runner
  script:
    - echo "running tests"
```

Example pipeline with all available configurations:
```
variables:
  ANKA_CLOUD_CONTROLLER_URL: "http://anka.contoller:8090"
  ANKA_CLOUD_TEMPLATE_ID: "8c592f53-65a4-444e-9342-79d3ff07837c"
  ANKA_CLOUD_TEMPLATE_TAG: "builder"
  ANKA_CLOUD_NODE_ID: "038d5b6f-2e54-44fd-a8bc-53629b086887"
  
build-job-1:
  stage: build
  tags:
    - anka_runner
  script:
    - echo "building part 1"

build-job-2:
  stage: build
  tags:
    - anka_runner
  variables:
    ANKA_CLOUD_NODE_GROUP_ID: "ff7e840e-9510-42e3-44ae-7a65c00c5979"
  script:
    - echo "building part 2"


unit-test-job:
  stage: test
  tags:
    - anka_runner
  variables:
    ANKA_CLOUD_TEMPLATE_TAG: "tester"
    ANKA_CLOUD_PRIORITY: "1"
  script:
    - echo "running unit tests"

integration-test-job:
  stage: test
  tags:
    - anka_runner
  variables:
    ANKA_CLOUD_TEMPLATE_TAG: "tester"
    ANKA_CLOUD_PRIORITY: "10"
  script:
    - echo "running integration tests"
```
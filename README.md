# Anka Cloud Gitlab Executor

A [Gitlab Runner Custom Executor](https://docs.gitlab.com/runner/executors/custom.html) utilizing [Anka Build Cloud](https://veertu.com/anka-build/) to run your Gitlab jobs on macOS VMs.

## Pre-requirements
1. Install and Register a self-managed Gitlab Runner. See [here](https://docs.gitlab.com/runner/install/index.html) for more info.
2. An active [Anka Build Cloud](https://veertu.com/anka-build/), with a template with ssh port forwarding configured.

## Configuration
1. Download the binary to the same machine your Gitlab Runner is running on.
2. Add the following `[runners.custom]` block to Gitlab Runner configuration:
    > By default, runner config is at `~/.gitlab-runner/config.toml`
    ```
    [runners.custom]
        config_exec = "/path/to/anka-cloud-gitlab-executor"
        config_args = ["config"]
        prepare_exec = "/path/to/anka-cloud-gitlab-executor"
        prepare_args = ["prepare"]
        run_exec = "/path/to/anka-cloud-gitlab-executor"
        run_args = ["run"]
        cleanup_exec = "/path/to/anka-cloud-gitlab-executor"
        cleanup_args = ["cleanup"]
    ```
3. Add `executor = "custom"` to `[[runners]]` block
4. (Optional) [Tag](https://docs.gitlab.com/ee/ci/runners/configure_runners.html#use-tags-to-control-which-jobs-a-runner-can-run) your Anka runners

Check out the [full configuration spec](https://docs.gitlab.com/runner/executors/custom.html#configuration) for more info

### Variables
All variables are parsed as strings, since Gitlab is stating: "To ensure consistent behavior, you should always put variable values in single or double quotes." [Reference](https://docs.gitlab.com/ee/ci/variables/)

Accepted values for booleans are: "1", "t", "T", "true", "TRUE", "True", "0", "f", "F", "false", "FALSE", "False"

| Variable name | Required | Type | Description |
| ------------- |:--------:|:----:| ----------- |
| ANKA_CLOUD_CONTROLLER_URL |     ✅ | String | Anka Build Cloud's Controller URL. Inlcuding `http[s]` prefix. Port optional |
| ANKA_CLOUD_TEMPLATE_ID | ✅ | String | VM Template ID to use. Must exist on the Registry and have SSH port forwarding |
| ANKA_CLOUD_DEBUG |     ❌ | Boolean | Output Anka Cloud debug info |
| ANKA_CLOUD_TEMPLATE_TAG | ❌ | String | Template tag to use |
| ANKA_CLOUD_NODE_ID | ❌ | String | Run VM on this specific node |
| ANKA_CLOUD_PRIORITY | ❌ | Number | Priority in range 1-10000 (lower is more urgent) |
| ANKA_CLOUD_NODE_GROUP_ID | ❌ | String | Run the VM on a specific Node Group, by Group ID |
| ANKA_CLOUD_SKIP_TLS_VERIFY | ❌ | Boolean | If Controller is using a self-signed cert, this allows the Runner to skip validation of the certificate |
| ANKA_CLOUD_CA_CERT_PATH | ❌ | String | If Controller is using a self-signed cert, CA file can be passed in for the runner to use when communicating with Controller. **_The path is accessed locally by the Runner_** |
| ANKA_CLOUD_CLIENT_CERT_PATH | ❌ | String | If Client Cert Authentication is enabled, this is the path for the Certificate. **_The path is accessed locally by the Runner_** |
| ANKA_CLOUD_CLIENT_CERT_KEY_PATH | ❌ | String | If Client Cert Authentication is enabled, this is the path for the Key. **_The path is accessed locally by the Runner_** |
| ANKA_CLOUD_SSH_USER_NAME | ❌ | String | SSH user name to use inside VM. Defaults to "anka" |
| ANKA_CLOUD_SSH_PASSWORD | ❌ | String | SSH password to use inside VM. Defaults to "admin" |
| ANKA_CLOUD_CUSTOM_HTTP_HEADER | ❌ | Object | key-value JSON object for custom headers to set when communicatin with Controller. Both keys and values must be strings  |
| ANKA_CLOUD_KEEP_ALIVE_ON_ERROR | ❌ | Boolean | Do not terminate Instance if job failed. This will leave the VM running until manually cleaned. Usually, this is used to inspect the VM post failing. If job was canceled, VM will be cleaned regardless of this variable. **There will be no indication for this behavior on the Job's output unless Gitlab Debug is enabled** |

### Examples
Example basic pipeline:
```
variables:
  ANKA_CLOUD_CONTROLLER_URL: "http://anka.contoller:8090"
  ANKA_CLOUD_TEMPLATE_ID: "8c592f53-65a4-444e-9342-79d3ff07837c"

build-job:
  stage: build
  tags:
    - anka_cloud_executor
  script:
    - echo "building"

test-job:
  stage: test
  tags:
    - anka_cloud_executor
  script:
    - echo "running tests"
```

Example pipeline with all available configurations:
```
variables:
  ANKA_CLOUD_CONTROLLER_URL: "https://anka.contoller:8090"
  ANKA_CLOUD_TEMPLATE_ID: "8c592f53-65a4-444e-9342-79d3ff07837c"
  ANKA_CLOUD_TEMPLATE_TAG: "builder"
  ANKA_CLOUD_NODE_ID: "038d5b6f-2e54-44fd-a8bc-53629b086887"
  ANKA_CLOUD_DEBUG: "true"
  ANKA_CLOUD_SKIP_TLS_VERIFY: "TRUE"
  ANKA_CLOUD_CA_CERT_PATH: "/mnt/certs/ca.pem"
  ANKA_CLOUD_CLIENT_CERT_PATH: "/mnt/certs/devops-crt.pem"
  ANKA_CLOUD_CLIENT_CERT_KEY_PATH: "/mnt/certs/devops-key.pem"
  ANKA_CLOUD_SSH_USER_NAME: "veertu"
  ANKA_CLOUD_SSH_PASSWORD: "P@$$w0rd"
  ANKA_CLOUD_CUSTOM_HTTP_HEADERS: "{\"X-CUSTOM-HEADER\": \"custom-value\"}"
  
build-job-1:
  stage: build
  tags:
    - anka_cloud_executor
  script:
    - echo "building part 1"

build-job-2:
  stage: build
  tags:
    - anka_cloud_executor
  variables:
    ANKA_CLOUD_NODE_GROUP_ID: "ff7e840e-9510-42e3-44ae-7a65c00c5979"
  script:
    - echo "building part 2"


unit-test-job:
  stage: test
  tags:
    - anka_cloud_executor
  variables:
    ANKA_CLOUD_TEMPLATE_TAG: "tester"
    ANKA_CLOUD_PRIORITY: "1"
  script:
    - echo "running unit tests"

integration-test-job:
  stage: test
  tags:
    - anka_cloud_executor
  variables:
    ANKA_CLOUD_TEMPLATE_TAG: "tester"
    ANKA_CLOUD_PRIORITY: "10"
  script:
    - echo "running integration tests"
```


## Development

### Overview
Please read Gitlab Custom Executor docs first [here](https://docs.gitlab.com/runner/executors/custom.html)

This project produces a single binary, that accepts the current Gitlab stage as its first argument:
1. Config
2. Prepare (Creates the Instance on the Anka Cloud, waiting for it to get scheduled)
3. Run (Creates a remote shell on the VM, and pushes Gitlab provided script with stdin)
4. Cleanup (Performs Termination request to the Anka Cloud Controller)
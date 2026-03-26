# Releasing

1. Run tests: https://bitbucket.org/ORGNAME HERE/anka-cloud-gitlab-executor-integration-test/src/master/README.md
2. Merge into main branch.
3. Create a new release on Github (push the semver tag, e.g. `v1.5.3`).
4. Build and watch https://jenkins/job/anka-cloud-gitlab-executor-release/ to see the release build.
5. Update the release description and title if needed. The workflow will attach artifacts to the release.

## How `--version` is produced

Release binaries are built with GoReleaser (see `.goreleaser.yaml`). The string printed by `anka-cloud-gitlab-executor --version` is:

`{{.Version}}-{{.ShortCommit}}` → e.g. `1.5.3-7fae15f`

Those values are injected via `-ldflags` at link time from **the git metadata of the Jenkins workspace** at build time. They are normally consistent with each other: **the compiled source is whatever commit `HEAD` was when `goreleaser release` ran**, and `.ShortCommit` is derived from that same `HEAD`.

You do **not** get “tag `1.5.3` with commit `7fae15f` printed but older source” from a correct release run, unless something overrides git/GoReleaser behavior (e.g. building from the wrong ref while forcing tag env vars—avoid that).

## Release job requirements

- The **release** Jenkins job (`ci/Jenkinsfile.release`) must check out the **exact tag ref** (or a detached `HEAD` at that tag), not an arbitrary `main` tip. The pipeline includes a **Verify release checkout** stage that fails if `HEAD` is not an exact tag match.
- Do not re-use **snapshot** artifacts from `ci/Jenkinsfile.build` (`goreleaser build --snapshot`) as official release binaries; that job is for CI artifacts only.
- If the job used a **shallow** clone, ensure tags are fetched (`git fetch --tags`); the verify stage runs `git fetch --tags --force`.
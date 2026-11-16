# Releasing

Default branch: `edge`.

The version is the git tag. This repository has no version file. Do not add one for a security release.

Tag form: `vMAJOR.MINOR.PATCH`.

The next security tag is the latest GitHub Release tag plus one patch. Ignore drafts and prereleases. Read that release when you choose the tag. Example: latest tag `v1.5.3` gives next tag `v1.5.4`.

Jenkins server: `https://jenkins.veertu.com`

View: [Gitlab Executor](https://jenkins.veertu.com/view/Gitlab%20Executor/)

The view has three jobs. Run them in this order: build, test, release. The build job makes the snapshot that the test uses. The release job builds the release binary again. Publish only the release binary.

Run every step below on one commit of `edge`. Stop at the first failure. Report the failed step and the URL.

## 1. Checks on the commit

Wait until these GitHub checks are success on that commit:

- `Tests / Test and Build` (workflow `Tests`, job `Test and Build`)
- `Analyze (go)` (workflow `CodeQL`)
- `Analyze (actions)` (workflow `CodeQL`)

Stop when a check is missing or the result is not success.

## 2. Snapshot build

Trigger Jenkins job `anka-cloud-gitlab-executor-build`.

Job URL: `https://jenkins.veertu.com/job/anka-cloud-gitlab-executor-build/`

Script path: `ci/Jenkinsfile.build`

Parameters: none. The Jenkinsfile has no `parameters` block. Send an empty parameter list.

The job runs `goreleaser build --clean --snapshot` on node label `cloud_container_host` and archives `dist/`. That archive is the test snapshot.

Wait for result `SUCCESS`. Stop on `FAILURE`, `UNSTABLE`, or `ABORTED`.

The Jenkinsfile has no git ref parameter. The job SCM must check out this same `edge` commit. Stop when the build used a different commit.

## 3. Integration test

The integration-test repository is private GitHub repository `veertuinc/anka-cloud-gitlab-executor-integration-test`.

URL: `https://github.com/veertuinc/anka-cloud-gitlab-executor-integration-test`

This product repository has no test Jenkinsfile. The test job is the other job on the Gitlab Executor view. It is not `anka-cloud-gitlab-executor-build`. It is not `anka-cloud-gitlab-executor-release`.

Read the view. When the view lists exactly one other job, that job is the test. Trigger it after the snapshot build is `SUCCESS`. The test uses the snapshot from step 2.

Pass no parameters when the job defines none. When the job requires a parameter and that parameter has no default, stop and report the parameter name.

Stop when the view is unreadable. Stop when the view does not contain exactly one other job. Report the job names that the view returned.

`ORGNAMEHERE` is not an organization. Use the GitHub repository above.

Wait for result `SUCCESS`. Stop on `FAILURE`, `UNSTABLE`, or `ABORTED`.

## 4. Tag

Do this step for a version release. Skip it to publish the edge release `v0.0.0`.

1. Read the latest GitHub Release that is not a draft and not a prerelease.
2. Add 1 to the patch number. That value is `vX.Y.Z`.
3. Create an annotated tag on the same commit and push the tag:

```bash
git tag -a vX.Y.Z <commit> -m "vX.Y.Z"
git push origin vX.Y.Z
```

Push the tag only. Leave the GitHub Release for step 5. Do not run `gh release create` in this step.

## 5. Release build

Trigger Jenkins job `anka-cloud-gitlab-executor-release`.

Job URL: `https://jenkins.veertu.com/job/anka-cloud-gitlab-executor-release/`

Script path: `ci/Jenkinsfile.release`

Parameters: none. Send an empty parameter list.

The job SCM checks out branch `edge`. The Jenkinsfile has no git ref parameter.

The job fetches tags and reads the latest GitHub Release that is not a draft and not a prerelease.

When a final tag `vX.Y.Z` is greater than that release, the job checks out that tag. `.goreleaser.yaml` sets `release.mode: keep-existing`. GoReleaser creates the GitHub Release when the tag has none. When a release for that tag already exists, GoReleaser keeps the current body and uploads the binaries. It does not open a second release.

When no such tag exists, the job checks out `edge`. It moves tag `v0.0.0` to that commit. It deletes the GitHub Release `v0.0.0` when one exists, then creates it again. The release name is `v0.0.0 (edge)`. The release is a prerelease. It is not the latest release. The next edge run replaces that release.

The job uses Jenkins credential `AnkaCloudGitlabExecutorGithubToken` as `GITHUB_TOKEN`. It runs `goreleaser release --clean` on node label `cloud_container_host`.

That command creates the one GitHub Release and uploads the binaries. This is a new build of the release binary. It does not publish the snapshot from step 2.

The Jenkins release job is the only creator of the GitHub Release. Do not create one before this job.

Wait for result `SUCCESS`. On `FAILURE`, `UNSTABLE`, or `ABORTED`: stop and report the build URL. Keep a version tag. A failed edge publish can leave tag `v0.0.0` on this commit with no GitHub Release. Run the job again.

## 6. Release notes

After `SUCCESS`, set the title and the body:

```bash
gh release edit vX.Y.Z --repo veertuinc/anka-cloud-gitlab-executor --title "vX.Y.Z" --notes "<notes>"
```

The body lists each security fix. Each item has the GHSA id or CVE id, the package name, the old version, and the new version.

Put those notes on the GitHub Release. Do not add a CHANGELOG file. Skip this step for the edge release. Its name and body are set by the release job.

## Binary version string

`anka-cloud-gitlab-executor --version` prints `VERSION-SHORTCOMMIT`. Example: `1.5.4-7fae15f`. An edge release prints `0.0.0-<shortcommit>`.

GoReleaser injects `.Version` and `.ShortCommit` with `-ldflags` from the git metadata of the Jenkins workspace. The compiled source is that same `HEAD`. A version release checks out that tag. An edge release checks out `edge`. A shallow clone must still have tags. The job fetches tags before it chooses the tag.

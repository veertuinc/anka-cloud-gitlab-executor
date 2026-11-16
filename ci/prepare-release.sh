#!/usr/bin/env bash
# Checkout the commit to release and print its tag on stdout.
# A version tag is a final vX.Y.Z greater than the latest GitHub Release.
# Otherwise this checks out origin/edge, moves tag v0.0.0, and deletes the
# v0.0.0 GitHub Release so the release job can create it again.
set -euo pipefail

if [ -z "${GITHUB_TOKEN:-}" ]; then
    echo "ERROR: GITHUB_TOKEN is required" >&2
    exit 1
fi

# Run from a copy outside the repo. git checkout replaces files in the repo,
# and bash would then read the wrong script.
cd "${REPO_ROOT:-$PWD}"
selector="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)/select-release-tag.sh"

repo=veertuinc/anka-cloud-gitlab-executor
api="https://api.github.com/repos/${repo}"

github_curl() {
    curl -sS \
        -H "Authorization: Bearer ${GITHUB_TOKEN}" \
        -H "Accept: application/vnd.github+json" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        "$@"
}

release_id_from_file() {
    awk -F'[:,]' '/"id":/ { gsub(/[^0-9]/, "", $2); print $2; exit }' "$1"
}

delete_github_release() {
    rel_tag=$1
    body=$(mktemp)
    code=$(github_curl -o "$body" -w '%{http_code}' "${api}/releases/tags/${rel_tag}")
    if [ "$code" = 404 ]; then
        rm -f "$body"
        return 0
    fi
    if [ "$code" != 200 ]; then
        echo "ERROR: cannot read GitHub Release ${rel_tag}: HTTP ${code}" >&2
        cat "$body" >&2
        rm -f "$body"
        exit 1
    fi
    rel_id=$(release_id_from_file "$body")
    rm -f "$body"
    if [ -z "$rel_id" ]; then
        echo "ERROR: GitHub Release ${rel_tag} has no id" >&2
        exit 1
    fi
    code=$(github_curl -o /dev/null -w '%{http_code}' -X DELETE "${api}/releases/${rel_id}")
    if [ "$code" != 204 ]; then
        echo "ERROR: cannot delete GitHub Release ${rel_tag}: HTTP ${code}" >&2
        exit 1
    fi
}

git fetch --tags --force --quiet
git fetch --quiet origin edge

latest_body=$(mktemp)
latest_code=$(github_curl -o "$latest_body" -w '%{http_code}' "${api}/releases/latest")
if [ "$latest_code" = 404 ]; then
    latest=v0.0.0
elif [ "$latest_code" = 200 ]; then
    latest=$(awk -F'"' '/"tag_name"/ { print $4; exit }' "$latest_body")
else
    echo "ERROR: cannot read the latest GitHub Release: HTTP ${latest_code}" >&2
    cat "$latest_body" >&2
    rm -f "$latest_body"
    exit 1
fi
rm -f "$latest_body"
if [ -z "$latest" ]; then
    echo "ERROR: latest GitHub Release has no tag" >&2
    exit 1
fi

release_tag=$(git tag --list | bash "$selector" "$latest")
if [ "$release_tag" = v0.0.0 ]; then
    git checkout -f -q origin/edge
    delete_github_release v0.0.0
    git tag -f -a v0.0.0 HEAD -m "v0.0.0 (edge)" >/dev/null
    git push --quiet --force \
        "https://x-access-token:${GITHUB_TOKEN}@github.com/${repo}.git" \
        "refs/tags/v0.0.0:refs/tags/v0.0.0"
else
    git checkout -f -q "$release_tag"
fi

printf '%s\n' "$release_tag"

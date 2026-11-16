#!/usr/bin/env bash
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
selector="${root}/select-release-tag.sh"
failed=0

expect_tag() {
    name=$1
    expected=$2
    latest=$3
    shift 3
    got=$(printf '%s\n' "$@" | bash "$selector" "$latest")
    if [ "$got" != "$expected" ]; then
        printf 'FAIL %s: got %s, expected %s\n' "$name" "$got" "$expected" >&2
        failed=1
    else
        printf 'ok %s\n' "$name"
    fi
}

expect_tag no-newer-tag v0.0.0 v1.5.3 v0.0.1 test v1.5.3
expect_tag next-patch v1.5.4 v1.5.3 v0.0.1 v1.5.4
expect_tag highest-newer v1.10.0 v1.5.3 v1.9.0 v1.10.0 v0.0.1
expect_tag prerelease-ignored v0.0.0 v1.5.3 v1.5.0-beta2
expect_tag already-released v0.0.0 v1.5.3 v1.5.3
expect_tag edge-tag-ignored v0.0.0 v1.5.3 v0.0.0
expect_tag only-test-tag v0.0.0 v1.5.3 test

got=$(printf '' | bash "$selector" v1.5.3)
if [ "$got" != v0.0.0 ]; then
    printf 'FAIL empty: got %s\n' "$got" >&2
    failed=1
else
    printf 'ok empty\n'
fi

got=$(printf '%s\n' v1.2.3 | bash "$selector" "")
if [ "$got" != v1.2.3 ]; then
    printf 'FAIL blank-latest: got %s\n' "$got" >&2
    failed=1
else
    printf 'ok blank-latest\n'
fi

exit "$failed"

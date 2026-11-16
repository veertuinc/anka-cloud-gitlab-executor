#!/usr/bin/env bash
# Print the tag to publish.
# Usage: select-release-tag.sh LATEST_RELEASE_TAG < tag-names
# A final tag vX.Y.Z that is greater than LATEST_RELEASE_TAG is a version release.
# v0.0.0 is never that tag. When no version tag exists, print v0.0.0 (edge).
set -eu

latest=${1:-v0.0.0}
if [ -z "$latest" ]; then
    latest=v0.0.0
fi

is_final_tag() {
    printf '%s\n' "$1" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'
}

version_is_greater() {
    [ "$1" != "$2" ] && [ "$(printf '%s\n%s\n' "$1" "$2" | sort -V | tail -n 1)" = "$1" ]
}

best=
while IFS= read -r tag || [ -n "$tag" ]; do
    tag=${tag%$'\r'}
    [ -n "$tag" ] || continue
    is_final_tag "$tag" || continue
    [ "$tag" = v0.0.0 ] && continue
    if version_is_greater "$tag" "$latest"; then
        if [ -z "$best" ] || version_is_greater "$tag" "$best"; then
            best=$tag
        fi
    fi
done

if [ -n "$best" ]; then
    printf '%s\n' "$best"
else
    printf '%s\n' v0.0.0
fi

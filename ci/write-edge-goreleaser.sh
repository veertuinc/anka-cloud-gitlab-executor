#!/usr/bin/env bash
# Write a GoReleaser config that replaces the v0.0.0 (edge) GitHub Release.
# Usage: write-edge-goreleaser.sh SOURCE DEST
set -eu

source_file=${1:?source config}
dest_file=${2:?dest config}

awk '
    /^release:/ {
        print "release:"
        print "  mode: replace"
        print "  replace_existing_artifacts: true"
        print "  make_latest: false"
        print "  prerelease: true"
        print "  name_template: \"v0.0.0 (edge)\""
        skip = 1
        next
    }
    skip && /^[^[:space:]#]/ { skip = 0 }
    skip { next }
    { print }
' "$source_file" > "$dest_file"

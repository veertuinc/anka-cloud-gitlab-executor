#!/usr/bin/env bash

set -e
go test ./...
go build -o anka-gle cmd/anka-cloud-gitlab-executor/main.go

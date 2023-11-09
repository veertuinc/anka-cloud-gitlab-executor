#!/usr/bin/env bash

set -e
go test ./...
go build -ldflags "-X github.com/veertuinc/anka-cloud-gitlab-executor/internal/version.version=dev -X github.com/veertuinc/anka-cloud-gitlab-executor/internal/version.commit=$(git rev-parse --short HEAD)" -o anka-gle cmd/anka-cloud-gitlab-executor/main.go

#!/usr/bin/env bash

go test ./...
go build -o anka-gle cmd/anka-cloud-gitlab-executor/main.go
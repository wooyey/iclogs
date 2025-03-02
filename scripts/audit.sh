#!/usr/bin/env bash
# Run audit apps for the project

SCRIPT_PATH=$(dirname "$(realpath "$0")")
. ${SCRIPT_PATH}/common.sh

run go mod tidy -diff
run go mod verify
run gofmt -l ${PROJECT_ROOT}
run go vet ${PROJECT_ROOT}/...
run go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all ${PROJECT_ROOT}/...
run go run golang.org/x/vuln/cmd/govulncheck@latest ${PROJECT_ROOT}/...

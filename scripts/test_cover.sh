#!/usr/bin/env bash
# Run test with coverage report

SCRIPT_PATH=$(dirname "$(realpath "$0")")
. ${SCRIPT_PATH}/common.sh

COVERAGE_FILE=${PROJECT_ROOT}/coverage.out

run go test -v -race -buildvcs -coverprofile=${COVERAGE_FILE} ${PROJECT_ROOT}/...
run go tool cover -html=${COVERAGE_FILE}
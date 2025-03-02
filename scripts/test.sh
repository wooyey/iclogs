#!/usr/bin/env bash
# Run unit tests

SCRIPT_PATH=$(dirname "$(realpath "$0")")
. ${SCRIPT_PATH}/common.sh

run go test -v -race -buildvcs ${PROJECT_ROOT}/...
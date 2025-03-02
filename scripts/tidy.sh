#!/usr/bin/env bash
# Tidy mod files

SCRIPT_PATH=$(dirname "$(realpath "$0")")
. ${SCRIPT_PATH}/common.sh

run go mod tidy -v
run go fmt ${PROJECT_ROOT}/...
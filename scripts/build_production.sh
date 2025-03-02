#!/usr/bin/env bash
# Tidy mod files

SCRIPT_PATH=$(dirname "$(realpath "$0")")
EXPECTED_PARAMS=3
USAGE_TEXT="Usage: $0 binary_name version /main/pkg/path"
. ${SCRIPT_PATH}/common.sh

NAME=${1}
VERSION=${2}
MAIN_PATH=${3}

# List of binaries to build
ARCHITECTURES=( "darwin:arm64" "darwin:amd64" "linux:amd64" )

for A in ${ARCHITECTURES[@]}; do
    IFS=: read -r GOOS GOARCH <<< ${A}
	run CGO_ENABLED=1 GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${NAME}.darwin.arm64 -ldflags \"-X main.version=${VERSION} -w -s\" ${MAIN_PATH}
done

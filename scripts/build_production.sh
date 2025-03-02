#!/usr/bin/env bash
# Build production binaries

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
    IFS=: read -r OS ARCH <<< ${A}
	run CGO_ENABLED=1 GOOS=${OS} GOARCH=${ARCH} go build -o ${NAME}.${OS}.${ARCH} -ldflags \"-X main.version=${VERSION} -w -s\" ${MAIN_PATH}
done

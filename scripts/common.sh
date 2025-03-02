#!/usr/bin/env bash
# Common functionality for shell scripts here

: "${EXPECTED_PARAMS:=1}"
: "${USAGE_TEXT:=Usage: $0 /path/to/project/root}"

if [ $# -lt ${EXPECTED_PARAMS} ]; then
    echo ${USAGE_TEXT}
    exit 1
fi

PROJECT_ROOT=${1}

function run {
    CMD=${*}

    echo -n "Running '${CMD}' "
    RESULT=$(bash -c "${CMD}" 2>&1)
    if [ $? -eq 0 ]; then
        echo "✅"
    else
        echo "❌"
        echo "Error details:"
        echo ${RESULT}
        exit 2
    fi
}
#!/usr/bin/env bash

export LOGS_API_KEY=someapikey
export LOGS_ENDPOINT=https://<instance-id>.api.<region-id>.logs.cloud.ibm.com

DIR=$(dirname "$0") # Script directory
${DIR}/iclogs $*

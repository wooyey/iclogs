# iclogs

IBM Cloud Logs CLI

It's a small project to learn Go Lang in hard but useful way.
So trying to avoid any non-standard libraries ...

## How to build

For build you can use attached `Makefile` and `make`:

```shell
make build
```

## How to use

To use it you need to know/have:

- API key of user with granted access to IBM Cloud Logs
- URL to IBM Cloud Logs Endpoint
- URL to IAM (Authorization) IBM Endpoint (default it is `https://iam.cloud.ibm.com`).

I recommend to use environmental variables (`LOGS_API_KEY`, `LOGS_ENDPOINT`) to store above information.
Of course you can override this values with CLI options.

### Usage message

```
Usage of iclogs: [options] <lucene query>

  -a, --auth-url string
        Authorization Endpoint URL. (default https://iam.cloud.ibm.com)
  -f, --from 2006-01-02T15:04
        Start time for log search in format 2006-01-02T15:04.
  -j, --show-json
        Show record as JSON.
  -k, --key LOG_API_KEY
        API Key to use. Overrides LOG_API_KEY environment variable.
  -l, --logs-url LOGS_ENDPOINT
        URL of IBM Cloud Log Endpoint. Overrides LOGS_ENDPOINT environment variable.
  -r, --range duration
        Relative time for log search, from now (or from end time if specified). (default 1h0m0s)
  --show-labels
        Show record labels.
  --show-severity
        Show record severity.
  --show-timestamp
        Show record timestamp.
  -t, --to 2006-01-02T15:04
        End time for log search in range format 2006-01-02T15:04.
  --version
        Show binary version.
```

### Example queries

#### Logs search from last 3 hours

```shell
./iclogs -r 3h --logs_url https://<instance-id>.api.<region-id>.logs.cloud.ibm.com --key someapikey 'kubernetes.pod_name:name-of-the-pod-with-some-random-uuid*'
```

`-r` specifies time duration from now in past or if specified end time

Last element `'kubernetes.pod_name:name-of-the-pod-with-some-random-uuid*'` was [Lucene](https://lucene.apache.org/core/2_9_4/queryparsersyntax.html) query looking for particular Pod logs.

#### Logs search using .env file

Example `.env` file:

```shell
LOGS_API_KEY=someapikey
LOGS_ENDPOINT=https://<instance-id>.api.<region-id>.logs.cloud.ibm.com
```

For such simple file you can use below trick to create environmental variables per Bash session:

```shell
export $(cat .env | xargs)
```

After above command - logs query as above is much simpler:

```shell
./iclogs -r 3h 'kubernetes.pod_name:name-of-the-pod-with-some-random-uuid*'
```

#### Logs search using shell script

Another approach to not repeat flags is to use shell script like below within `iclogs` binary directory:

```bash
#!/usr/bin/env bash

export LOGS_API_KEY=someapikey
export LOGS_ENDPOINT=https://<instance-id>.api.<region-id>.logs.cloud.ibm.com

DIR=$(dirname "$0") # Script directory
$DIR/iclogs $*
```

#### Logs using 1password CLI

Create a `.env` file with references to your [1Password](https://1password.com) vault, ie.:

```env
LOGS_API_KEY=op://${NAME}/logs/credential
LOGS_ENDPOINT=op://${NAME}/logs/url
```

Where `${NAME}` is environmental variable with vault name.
Then run `iclogs` with `op run` command, like this:

```bash
NAME="my_site" op run --env-file="./.env" -- ./iclogs 'kubernetes.pod_name:name-of-the-pod-with-some-random-uuid*'
```

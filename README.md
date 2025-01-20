# iclogs

IBM Cloud Logs CLI

It's a small project to learn Go Lang in hard but useful way.

## How to build

For now only manual recipe to build it:

```shell
go build cmd/iclogs/iclogs.go
```

## How to use

To use it you need to know/have:

- API key of user with granted access to IBM Cloud Logs
- URL to IBM Cloud Logs Endpoint
- URL to IAM (Authorization) IBM Endpoint.

I recommend to use environmental variables (`LOGS_API_KEY`, `LOGS_ENDPOINT`, `IAM_ENDPOINT`) to store above information.
Of course you can override this values with CLI options.

### Example queries

#### Logs search from last 3 hours

```shell
./iclogs -r 3h -auth_url https://iam.cloud.ibm.com -logs_url https://8e22d8ef-1xx9-49fb-be9a-zz287944864f.api.us-south.logs.cloud.ibm.com -key someapikey 'kubernetes.pod_name:name-of-the-pod-with-some-random-uuid*'
```

`-r` specifies time duration from now in past or if specified end time

Last element `'kubernetes.pod_name:name-of-the-pod-with-some-random-uuid*'` was [Lucene](https://lucene.apache.org/core/2_9_4/queryparsersyntax.html) query looking for particular Pod logs.

#### Logs search using .env file

Example `.env` file:

```shell
LOGS_API_KEY=someapikey
LOGS_ENDPOINT=https://8e22d8ef-1xx9-49fb-be9a-zz287944864f.api.us-south.logs.cloud.ibm.com
IAM_ENDPOINT=https://iam.cloud.ibm.com
```

For such simple file you can use below trick to create environmental variables per Bash session:

```shell
export $(cat .env | xargs)
```

After above command - logs query as above is much simpler:

```shell
./iclogs -r 3h 'kubernetes.pod_name:name-of-the-pod-with-some-random-uuid*'
```

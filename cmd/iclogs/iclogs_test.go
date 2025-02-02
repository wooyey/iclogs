package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

func TestParseArgs(t *testing.T) {

	testCases := []struct {
		name  string
		input string
		envs  map[string]string
		want  CmdArgs
	}{
		{
			name:  "LongOptions",
			input: "./iclogs --key ApiKey --from 2024-03-12T12:00 --to 2024-03-12T13:00 --range 30m --logs-url https://logs.endpoint.cloud.ibm.com --auth-url https://iam.different.cloud.ibm.com lucene query",
			envs:  map[string]string{},
			want: CmdArgs{
				ApiKey:    "ApiKey",
				TimeRange: time.Minute * 30,
				LogsURL:   "https://logs.endpoint.cloud.ibm.com",
				AuthURL:   "https://iam.different.cloud.ibm.com",
				StartTime: timestamp(time.Date(2024, 3, 12, 12, 0, 0, 0, time.Local)),
				EndTime:   timestamp(time.Date(2024, 3, 12, 13, 0, 0, 0, time.Local)),
				Query:     "lucene query",
			},
		},
		{
			name:  "ShortOptions",
			input: "./iclogs -k ApiKey -f 2024-03-12T12:00 -t 2024-03-12T13:00 -r 30m -l https://logs.endpoint.cloud.ibm.com -a https://iam.different.cloud.ibm.com lucene query",
			envs:  map[string]string{},
			want: CmdArgs{
				ApiKey:    "ApiKey",
				TimeRange: time.Minute * 30,
				LogsURL:   "https://logs.endpoint.cloud.ibm.com",
				AuthURL:   "https://iam.different.cloud.ibm.com",
				StartTime: timestamp(time.Date(2024, 3, 12, 12, 0, 0, 0, time.Local)),
				EndTime:   timestamp(time.Date(2024, 3, 12, 13, 0, 0, 0, time.Local)),
				Query:     "lucene query",
			},
		},
		{
			name:  "DefaultValues",
			input: "./iclogs lucene query",
			envs:  map[string]string{},
			want: CmdArgs{
				TimeRange: defaultTimeRange,
				AuthURL:   iamURL,
				Query:     "lucene query",
			},
		},
		{
			name:  "UpdateValuesWithEnvs",
			input: "./iclogs lucene query",
			envs:  map[string]string{"LOGS_API_KEY": "api_key", "LOGS_ENDPOINT": "https://logs.cloud.ibm.com"},
			want: CmdArgs{
				TimeRange: defaultTimeRange,
				AuthURL:   iamURL,
				Query:     "lucene query",
				LogsURL:   "https://logs.cloud.ibm.com",
				ApiKey:    "api_key",
			},
		},
		{
			name:  "DontUpdateExistingValuesWithEnvs",
			input: "./iclogs -k some_key lucene query",
			envs:  map[string]string{"LOGS_API_KEY": "api_key", "LOGS_ENDPOINT": "https://logs.cloud.ibm.com"},
			want: CmdArgs{
				TimeRange: defaultTimeRange,
				AuthURL:   iamURL,
				Query:     "lucene query",
				LogsURL:   "https://logs.cloud.ibm.com",
				ApiKey:    "some_key",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = strings.Split(tt.input, " ")

			for k, v := range tt.envs {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envs {
					os.Unsetenv(k)
				}
			}()

			got := parseArgs()
			if got != tt.want {
				t.Errorf("\nGot:\t%+v\nWant:\t%+v", got, tt.want)
			}
		})
	}

}

func TestPrintUsage(t *testing.T) {

	b := bytes.Buffer{}
	os.Args = []string{"./iclogs"}

	initParser(&CmdArgs{})
	printUsage(&b)
	got := b.String()

	want := `Usage of ./iclogs: [options] <lucene query>

  -a, --auth-url string
        Authorization Endpoint URL. (default https://iam.cloud.ibm.com)
  -f, --from 2006-01-02T15:04
        Start time for log search in format 2006-01-02T15:04.
  -k, --key LOG_API_KEY
        API Key to use. Overrides LOG_API_KEY environment variable.
  -l, --logs-url LOGS_ENDPOINT
        URL of IBM Cloud Log Endpoint. Overrides LOGS_ENDPOINT environment variable.
  -r, --range duration
        Relative time for log search, from now (or from end time if specified). (default 1h0m0s)
  -t, --to 2006-01-02T15:04
        End time for log search in range format 2006-01-02T15:04.
  --version
        Show binary version.
`

	if got != want {
		t.Errorf("\nGot:\t%q\nWant:\t%q", got, want)
	}

}

func TestGetVersion(t *testing.T) {

	version = "v1.0.0"
	got := getVersion()
	want := "iclogs version v1.0.0"

	if got != want {
		t.Errorf("\nGot:\t'%s'\nWant:\t'%s'", got, want)
	}
}

package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wooyey/iclogs/internal/platform/logs"
)

func assert[T comparable](t testing.TB, got T, want T) {
	t.Helper()
	if got != want {
		t.Errorf("\nGot:\t%+v\nWant:\t%+v", got, want)
	}
}

func assertError(t testing.TB, got, want error) {
	t.Helper()
	if want == nil && got != nil {
		t.Fatalf("got an error but didn't want one: '%+v'", got)
	}

	if want != got {
		t.Errorf("\nGot:\t%+v\nWant:\t%s", got, want)
	}
}

func TestParseArgs(t *testing.T) {

	testCases := []struct {
		name  string
		input string
		envs  map[string]string
		want  CmdArgs
	}{
		{
			name:  "LongOptions",
			input: "./iclogs --key ApiKey --from 2024-03-12T12:00 --to 2024-03-12T13:00 --range 30m --logs-url https://logs.endpoint.cloud.ibm.com --auth-url https://iam.different.cloud.ibm.com --message-fields another,keys lucene query",
			envs:  map[string]string{},
			want: CmdArgs{
				APIKey:    "ApiKey",
				TimeRange: time.Minute * 30,
				LogsURL:   "https://logs.endpoint.cloud.ibm.com",
				AuthURL:   "https://iam.different.cloud.ibm.com",
				StartTime: timestamp(time.Date(2024, 3, 12, 12, 0, 0, 0, time.Local)),
				EndTime:   timestamp(time.Date(2024, 3, 12, 13, 0, 0, 0, time.Local)),
				Query:     "lucene query",
				KeyNames:  "another,keys",
			},
		},
		{
			name:  "ShortOptions",
			input: "./iclogs -k ApiKey -f 2024-03-12T12:00 -t 2024-03-12T13:00 -r 30m -l https://logs.endpoint.cloud.ibm.com -a https://iam.different.cloud.ibm.com -m some,keys lucene query",
			envs:  map[string]string{},
			want: CmdArgs{
				APIKey:    "ApiKey",
				TimeRange: time.Minute * 30,
				LogsURL:   "https://logs.endpoint.cloud.ibm.com",
				AuthURL:   "https://iam.different.cloud.ibm.com",
				StartTime: timestamp(time.Date(2024, 3, 12, 12, 0, 0, 0, time.Local)),
				EndTime:   timestamp(time.Date(2024, 3, 12, 13, 0, 0, 0, time.Local)),
				Query:     "lucene query",
				KeyNames:  "some,keys",
			},
		},
		{
			name:  "DefaultValues",
			input: "./iclogs lucene query",
			envs:  map[string]string{},
			want: CmdArgs{
				TimeRange: defaultTimeRange,
				AuthURL:   defaultIAMURL,
				Query:     "lucene query",
				KeyNames:  defaultKeyNames,
			},
		},
		{
			name:  "UpdateValuesWithEnvs",
			input: "./iclogs lucene query",
			envs:  map[string]string{"LOGS_API_KEY": "api_key", "LOGS_ENDPOINT": "https://logs.cloud.ibm.com"},
			want: CmdArgs{
				TimeRange: defaultTimeRange,
				AuthURL:   defaultIAMURL,
				Query:     "lucene query",
				LogsURL:   "https://logs.cloud.ibm.com",
				APIKey:    "api_key",
				KeyNames:  defaultKeyNames,
			},
		},
		{
			name:  "DontUpdateExistingValuesWithEnvs",
			input: "./iclogs -k some_key lucene query",
			envs:  map[string]string{"LOGS_API_KEY": "api_key", "LOGS_ENDPOINT": "https://logs.cloud.ibm.com"},
			want: CmdArgs{
				TimeRange: defaultTimeRange,
				AuthURL:   defaultIAMURL,
				Query:     "lucene query",
				LogsURL:   "https://logs.cloud.ibm.com",
				APIKey:    "some_key",
				KeyNames:  defaultKeyNames,
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
			assert(t, got, tt.want)
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
  -j, --show-json
        Show record as JSON.
  -k, --key LOG_API_KEY
        API Key to use. Overrides LOG_API_KEY environment variable.
  -l, --logs-url LOGS_ENDPOINT
        URL of IBM Cloud Log Endpoint. Overrides LOGS_ENDPOINT environment variable.
  -m, --message-fields string
        Comma separated message field names. (default message,message_obj.msg,log)
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
`

	assert(t, got, want)
}

func TestGetVersion(t *testing.T) {

	version = "v1.0.0"
	got := getVersion()
	want := "iclogs version v1.0.0"

	assert(t, got, want)
}

func TestValidateArgs(t *testing.T) {
	testCases := []struct {
		name  string
		input CmdArgs
		want  error
	}{
		{
			name:  "AllOk",
			input: CmdArgs{APIKey: "api_key", LogsURL: "url", Query: "some query"},
			want:  nil,
		},
		{
			name:  "MissingAPIKey",
			input: CmdArgs{LogsURL: "url", Query: "some query"},
			want:  errMissingAPIKey,
		},
		{
			name:  "MissingURL",
			input: CmdArgs{APIKey: "api_key", Query: "some query"},
			want:  errMissingURL,
		},
		{
			name:  "MissingQuery",
			input: CmdArgs{APIKey: "api_key", LogsURL: "url"},
			want:  errMissingQuery,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got := validateArgs(&tt.input)

			assertError(t, got, tt.want)
		})
	}

}

func TestPrintLogs(t *testing.T) {
	logs := []logs.Log{
		{
			Time:     time.Date(2025, 1, 11, 18, 52, 21, 26304000, time.Local),
			Severity: "Debug",
			UserData: `{"message":"some_message"}`,
			Labels:   []string{"label:\"value-of-label\""},
		},
	}

	testCases := []struct {
		name string
		args CmdArgs
		want string
	}{
		{
			name: "Default",
			args: CmdArgs{KeyNames: defaultKeyNames},
			want: "some_message\n",
		},
		{
			name: "ShowTimestamp",
			args: CmdArgs{KeyNames: defaultKeyNames, Timestamp: true},
			want: "2025-01-11 18:52:21: some_message\n",
		},
		{
			name: "ShowSeverity",
			args: CmdArgs{KeyNames: defaultKeyNames, Severity: true},
			want: "[Debug] some_message\n",
		},
		{
			name: "ShowLabels",
			args: CmdArgs{KeyNames: defaultKeyNames, Labels: true},
			want: "<label:\"value-of-label\"> some_message\n",
		},
		{
			name: "ShowJSON",
			args: CmdArgs{KeyNames: defaultKeyNames, JSON: true},
			want: "{\"message\":\"some_message\"}\n",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.Buffer{}
			printLogs(&buffer, &logs, &tt.args)
			got := buffer.String()
			assert(t, got, tt.want)
		})
	}

}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wooyey/iclogs/internal/platform/auth"
	"github.com/wooyey/iclogs/internal/platform/logs"
	"github.com/wooyey/iclogs/internal/platform/logs/syntax"
	"github.com/wooyey/iclogs/internal/platform/logs/tier"
)

const (
	apiKeyVar       = "LOGS_API_KEY"
	logsEndpointVar = "LOGS_ENDPOINT"
	iamEndpointVar  = "IAM_ENDPOINT"
)

const timeFormat = "2006-01-02T15:04"

var (
	flagKey       string
	flagTimeRange time.Duration
	flagLogsUrl   string
	flagAuthUrl   string
	flagFromTime  string
	flagToTime    string
	argQuery      string
)

func parseTime(t string) (time.Time, error) {
	return time.ParseInLocation(timeFormat, t, time.Local)
}

func getEnvOption(value, key, name string) string {
	if value != "" {
		return value
	}
	v := os.Getenv(key)

	if v == "" {
		log.Fatalf("%s is not set. Use proper option or `%s` env variable.", name, key)
	}

	return v
}

func parseArgs() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage of %s: [options] <lucene query>\n", os.Args[0])
		flag.PrintDefaults()
	}

	const (
		defaultTimeRange = time.Hour
		usageTimeRange   = "Relative time for log search, from now (or from end time if specified)."
		usageFromTime    = "Start time for log search in format `" + timeFormat + "`."
		usageToTime      = "End time for log search in range format `" + timeFormat + "`."
	)

	flag.StringVar(&flagKey, "key", "", "API Key to use. Overrides `"+apiKeyVar+"` environment variable.")
	flag.StringVar(&flagAuthUrl, "auth_url", "", "Authorization Endpoint URL. Overrides `"+iamEndpointVar+"` environment variable.")
	flag.StringVar(&flagLogsUrl, "logs_url", "", "URL of IBM Cloud Log Endpoint. Overrides `"+logsEndpointVar+"` environment variable.")
	flag.DurationVar(&flagTimeRange, "time_range", defaultTimeRange, usageTimeRange)
	flag.DurationVar(&flagTimeRange, "r", defaultTimeRange, usageTimeRange+" (shorthand)")
	flag.StringVar(&flagFromTime, "from", "", usageFromTime)
	flag.StringVar(&flagFromTime, "f", "", usageFromTime+" (shorthand)")
	flag.StringVar(&flagToTime, "to", "", usageFromTime)
	flag.StringVar(&flagToTime, "t", "", usageToTime+" (shorthand)")

	flag.Parse()
	argQuery = flag.Arg(0)
}

func main() {

	var (
		endDate   time.Time
		startDate time.Time
	)

	parseArgs()

	if argQuery == "" {
		log.Fatal("Logs query cannot be empty. Use `Lucene` syntax.")
	}

	flagKey = getEnvOption(flagKey, apiKeyVar, "API key")
	flagLogsUrl = getEnvOption(flagLogsUrl, logsEndpointVar, "Logs Endpoint")
	flagAuthUrl = getEnvOption(flagAuthUrl, iamEndpointVar, "IAM Endpoint")

	token, err := auth.GetToken(flagAuthUrl, flagKey)

	if err != nil {
		log.Fatalf("Cannot get token from '%s': %v", flagAuthUrl, err)
	}

	if flagFromTime != "" {
		startDate, err = parseTime(flagFromTime)
		if err != nil {
			log.Fatalf("Cannot parse from time: '%s'", flagFromTime)
		}
	}

	if flagToTime != "" {
		endDate, err = parseTime(flagToTime)
		if err != nil {
			log.Fatalf("Cannot parse to time: '%s'", flagToTime)
		}
	}

	if endDate.IsZero() {
		endDate = time.Now()
	}

	if startDate.IsZero() {
		startDate = endDate.Add(-flagTimeRange)
	}

	spec := logs.QuerySpec{
		Syntax:    syntax.Lucene,
		Tier:      tier.Archive,
		Limit:     tier.LimitArchive,
		StartDate: startDate,
		EndDate:   endDate,
	}

	l, err := logs.QueryLogs(flagLogsUrl, token.Value, argQuery, spec)
	if err != nil {
		log.Fatalf("Cannot get logs from '%s': %v", flagLogsUrl, err)
	}

	for _, line := range l {
		fmt.Println(line.Message)
	}

}

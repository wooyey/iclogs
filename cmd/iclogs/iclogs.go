package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/wooyey/iclogs/internal/platform/auth"
	"github.com/wooyey/iclogs/internal/platform/logs"
	"github.com/wooyey/iclogs/internal/platform/logs/syntax"
	"github.com/wooyey/iclogs/internal/platform/logs/tier"
)

const (
	timeFormat       = "2006-01-02T15:04"
	defaultTimeRange = time.Hour
)

const iamURL = "https://iam.cloud.ibm.com"

// Possible errors list for easier testing later on
const (
	errorMissingURL    = "you need to provide IBM Cloud Logs endpoint URL"
	errorMissingAPIKey = "you need to provide API key"
	errorMissingQuery  = "you need to provide logs query string"
)

// Will be set in compile time
var version string

func parseTime(t string) (time.Time, error) {
	return time.ParseInLocation(timeFormat, t, time.Local)
}

type timestamp time.Time

func (t *timestamp) String() string {
	return fmt.Sprint(*t)
}

func (t *timestamp) Set(value string) error {
	pt, err := parseTime(value)
	if err != nil {
		return err
	}
	*t = timestamp(pt)
	return nil
}

// CmdArgs includes all options
// need to have exportable fields for reflect ...
type CmdArgs struct {
	APIKey    string `env:"LOGS_API_KEY"`
	TimeRange time.Duration
	LogsURL   string `env:"LOGS_ENDPOINT"`
	AuthURL   string
	StartTime timestamp
	EndTime   timestamp
	Query     string
	Version   bool
	JSON      bool
	Labels    bool
	Severity  bool
	Timestamp bool
}

func getEnvArgs(args *CmdArgs) {

	t := reflect.TypeOf(*args)

	for i, f := range reflect.VisibleFields(t) {
		if k := f.Tag.Get("env"); k != "" {
			if fv := reflect.ValueOf(args).Elem().Field(i); fv.String() == "" {
				v := os.Getenv(k)
				fv.SetString(v)
			}
		}
	}
}

func addFlagsVar(value interface{}, names []string, usage string, defaultValue interface{}) error {
	for _, name := range names {
		switch v := value.(type) {
		case *string:
			flag.StringVar(v, name, defaultValue.(string), usage)
		case *time.Duration:
			flag.DurationVar(v, name, defaultValue.(time.Duration), usage)
		case flag.Value:
			flag.Var(v, name, usage)
		case *bool:
			flag.BoolVar(v, name, defaultValue.(bool), usage)
		default:
			return errors.New("unknown type of flag value")
		}
	}
	return nil
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, "Usage of %s: [options] <lucene query>\n\n", os.Args[0])

	args := map[string]struct {
		names    []string
		name     string
		defValue string
	}{}

	flag.VisitAll(func(f *flag.Flag) {
		name, usage := flag.UnquoteUsage(f)

		// Use option `usage` as a unique key
		option := args[usage]

		option.names = append(option.names, f.Name)
		// Sort options names by their length
		sort.SliceStable(option.names, func(i, j int) bool {
			return len(option.names[i]) < len(option.names[j])
		})

		option.name = name

		// almost copy pasta from flag to check zero value of default value
		typ := reflect.TypeOf(f.Value)
		var z reflect.Value
		if typ.Kind() == reflect.Pointer {
			z = reflect.New(typ.Elem())
		} else {
			z = reflect.Zero(typ)
		}

		// Add default value if it is not zero
		if f.DefValue != z.Interface().(flag.Value).String() {
			option.defValue = f.DefValue
		}

		args[usage] = option
	})

	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}

	// Sort printout in alphabetical order of flag names
	sort.SliceStable(keys, func(i, j int) bool {
		return args[keys[i]].names[0] < args[keys[j]].names[0]
	})

	for _, k := range keys {

		// Add proper number of dashes to options
		names := make([]string, len(args[k].names))
		for i, n := range args[k].names {
			if len(n) > 1 {
				names[i] = "--" + n
			} else {
				names[i] = "-" + n
			}
		}

		// flags
		fmt.Fprintf(w, "  %s", strings.Join(names, ", "))

		// type names
		if args[k].name != "" {
			fmt.Fprintf(w, " %s", args[k].name)
		}

		// usage
		fmt.Fprintf(w, "\n        %s", k)
		if args[k].defValue != "" {
			fmt.Fprintf(w, " (default %s)", args[k].defValue)
		}
		fmt.Fprint(w, "\n")
	}

}

// Configure command line arguments parsing
func initParser(args *CmdArgs) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	addFlagsVar(&args.APIKey, []string{"key", "k"}, "API Key to use. Overrides `LOG_API_KEY` environment variable.", "")
	addFlagsVar(&args.AuthURL, []string{"auth-url", "a"}, "Authorization Endpoint URL.", iamURL)
	addFlagsVar(&args.LogsURL, []string{"logs-url", "l"}, "URL of IBM Cloud Log Endpoint. Overrides `LOGS_ENDPOINT` environment variable.", "")
	addFlagsVar(&args.TimeRange, []string{"range", "r"}, "Relative time for log search, from now (or from end time if specified).", defaultTimeRange)
	addFlagsVar(&args.StartTime, []string{"from", "f"}, "Start time for log search in format `"+timeFormat+"`.", nil)
	addFlagsVar(&args.EndTime, []string{"to", "t"}, "End time for log search in range format `"+timeFormat+"`.", nil)
	addFlagsVar(&args.Version, []string{"version"}, "Show binary version.", false)
	addFlagsVar(&args.JSON, []string{"j", "show-json"}, "Show record as JSON.", false)
	addFlagsVar(&args.Labels, []string{"show-labels"}, "Show record labels.", false)
	addFlagsVar(&args.Severity, []string{"show-severity"}, "Show record severity.", false)
	addFlagsVar(&args.Timestamp, []string{"show-timestamp"}, "Show record timestamp.", false)
}

// Parse command line args
func parseArgs() CmdArgs {

	// Re-init FlagSet to avoid unit tests dependency
	// flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	args := CmdArgs{}
	initParser(&args)

	flag.CommandLine.Usage = func() {
		w := flag.CommandLine.Output()
		printUsage(w)
	}

	flag.Parse()
	args.Query = strings.Join(flag.Args(), " ")

	getEnvArgs(&args)

	return args
}

// Simple produce version string
func getVersion() string {
	return fmt.Sprintf("iclogs version %s", version)
}

// Validate if CmdArgs has proper values
func validateArgs(args *CmdArgs) error {

	if args.APIKey == "" {
		return errors.New(errorMissingAPIKey)
	}

	if args.LogsURL == "" {
		return errors.New(errorMissingURL)
	}

	if args.Query == "" {
		return errors.New(errorMissingQuery)
	}

	return nil
}

func main() {

	args := parseArgs()

	if args.Version {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n", getVersion())
		os.Exit(0)
	}

	if err := validateArgs(&args); err != nil {
		log.Fatalf("Error in parsing arguments: %v", err)
	}

	token, err := auth.GetToken(args.AuthURL, args.APIKey)

	if err != nil {
		log.Fatalf("Cannot get token from '%s': %v", args.AuthURL, err)
	}

	endDate := time.Time(args.EndTime)
	startDate := time.Time(args.StartTime)

	if endDate.IsZero() {
		endDate = time.Now()
	}

	if startDate.IsZero() {
		startDate = endDate.Add(-args.TimeRange)
	}

	spec := logs.QuerySpec{
		Syntax:    syntax.Lucene,
		Tier:      tier.Archive,
		Limit:     tier.LimitArchive,
		StartDate: startDate,
		EndDate:   endDate,
	}

	l, err := logs.QueryLogs(args.LogsURL, token.Value, args.Query, spec)
	if err != nil {
		log.Fatalf("Cannot get logs from '%s': %v", args.LogsURL, err)
	}

	for _, line := range l {
		if args.Timestamp {
			fmt.Printf("%s: ", line.Time)
		}

		if args.Severity {
			fmt.Printf("[%s] ", line.Severity)
		}

		if args.Labels {
			fmt.Printf("<%s> ", strings.Join(line.Labels, ", "))
		}

		if args.JSON {
			fmt.Println(line.UserData)
		} else {
			fmt.Println(line.Message)
		}
	}

}

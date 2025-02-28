// Package logs to communicate with IBM Cloud Logs API
package logs

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/wooyey/iclogs/internal/platform/logs/syntax"
	"github.com/wooyey/iclogs/internal/platform/logs/tier"
)

const (
	dataPrefix     = "data: "
	timeFormat     = "2006-01-02T15:04:05.999999"
	timestampField = "timestamp"
	severityField  = "severity"
)

const queryPath = "/v1/query"

const maxLineSize = 2048 * 1024 // Max line size - 2MB should be enough.

type QuerySpec struct {
	Syntax    syntax.Syntax `json:"syntax"`
	Limit     int           `json:"limit"`
	Tier      tier.Tier     `json:"tier"`
	StartDate time.Time     `json:"start_date"`
	EndDate   time.Time     `json:"end_date"`
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type Log struct {
	Time     time.Time
	Severity string
	Message  string
	UserData string // RAW User Data JSON string
	Labels   []string
}

type UserData struct {
	Message    any `json:"message"` // could be float/int or string
	MessageObj struct {
		Message any `json:"msg"` // just in case pretend that it can be anything, see above ...
	} `json:"message_obj"`
}

type Record struct {
	Data     string     `json:"user_data"`
	Metadata []KeyValue `json:"metadata"`
	Labels   []KeyValue `json:"labels"`
}

type MessageResult struct {
	Result struct {
		Results []Record `json:"results"`
	} `json:"result"`
}

type Query struct {
	Query    string          `json:"query"`
	Metadata *map[string]any `json:"metadata"`
}

var GetQueryURL = func(endpoint string) (string, error) {
	return url.JoinPath(endpoint, queryPath)
}

var QueryTimeout = time.Duration(3) * time.Minute // HTTP query timeout - default 3 minutes

func structToMap(data any, m *map[string]any) {
	fields := reflect.VisibleFields(reflect.TypeOf(data))
	values := reflect.ValueOf(data)
	for _, field := range fields {
		v := values.FieldByName(field.Name)
		if v.IsZero() {
			continue
		}

		// Using `json` tag for map keys
		(*m)[field.Tag.Get("json")] = v.Interface()
	}
}

func getValue(kv []KeyValue, key string) (string, error) {

	for _, v := range kv {
		if v.Key == key {
			return v.Value, nil
		}
	}

	return "", fmt.Errorf("cannot find value for key: '%s'", key)
}

func parseRecord(record *Record) (Log, error) {

	timestamp, err := getValue(record.Metadata, timestampField)
	if err != nil {
		return Log{}, fmt.Errorf("cannot parse timestamp: %w", err)
	}

	severity, err := getValue(record.Metadata, severityField)
	if err != nil {
		return Log{}, fmt.Errorf("cannot parse severity: %w", err)
	}

	t, err := time.ParseInLocation(timeFormat, timestamp, time.Local)
	if err != nil {
		return Log{}, fmt.Errorf("cannot parse timestamp: %w", err)
	}

	ud := UserData{}
	if err := json.Unmarshal([]byte(record.Data), &ud); err != nil {
		return Log{}, fmt.Errorf("cannot unmarshal user data: %w", err)
	}

	m := ud.Message

	// If Message is nil maybe we have message_obj
	if m == nil {
		m = ud.MessageObj.Message
	}

	// If no luck lets have it as an empty string
	if m == nil {
		m = ""
	}

	labels := make([]string, len(record.Labels))
	for i, label := range record.Labels {
		labels[i] = fmt.Sprintf("%s:\"%s\"", label.Key, label.Value)
	}

	log := Log{
		Time:     t,
		Severity: severity,
		Message:  fmt.Sprintf("%v", m),
		UserData: record.Data,
		Labels:   labels,
	}

	return log, nil
}

func parseResponse(response io.Reader) ([]Log, error) {

	logs := []Log{}

	scanner := bufio.NewScanner(response)

	buf := make([]byte, maxLineSize)
	scanner.Buffer(buf, maxLineSize)

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, dataPrefix) {
			continue
		}

		d := line[len(dataPrefix):]
		// data := make(map[string]any)
		data := MessageResult{}

		if err := json.Unmarshal([]byte(d), &data); err != nil {
			return nil, fmt.Errorf("cannot unmarshal data line payload: %w", err)
		}

		for _, r := range data.Result.Results {

			l, err := parseRecord(&r)
			if err != nil {
				return nil, fmt.Errorf("cannot parse record from results: %w", err)
			}

			logs = append(logs, l)

		}

	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Sort logs
	sort.Slice(logs, func(i, j int) bool { return logs[i].Time.Compare(logs[j].Time) < 0 })

	return logs, nil
}

func QueryLogs(endpoint, token, query string, spec QuerySpec) ([]Log, error) {

	q := Query{Query: query}

	if spec != (QuerySpec{}) {
		meta := make(map[string]any)
		structToMap(spec, &meta)

		q.Metadata = &meta
	}

	j, err := json.Marshal(q)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal payload: %w", err)
	}

	payload := bytes.NewBuffer(j)

	addr, err := GetQueryURL(endpoint)
	if err != nil {
		return nil, fmt.Errorf("cannot create query URL: %w", err)
	}

	c := http.Client{Timeout: QueryTimeout}
	req, err := http.NewRequest("POST", addr, payload)
	if err != nil {
		return nil, fmt.Errorf("cannot create POST request: %w", err)
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("authorization", "Bearer "+token)

	resp, err := c.Do(req)

	if err != nil {
		return nil, fmt.Errorf("cannot POST data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return nil, fmt.Errorf("cannot read body: %w", err)
		}

		return nil, fmt.Errorf("got HTTP error code: %d, message: '%s'", resp.StatusCode, body)
	}

	logs, err := parseResponse(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("error when parsing results: %w", err)
	}

	return logs, nil

}

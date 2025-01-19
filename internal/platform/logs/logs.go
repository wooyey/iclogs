package logs

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/wooyey/iclogs/internal/platform/logs/syntax"
	"github.com/wooyey/iclogs/internal/platform/logs/tier"
)

const dataPrefix = "data: "
const timeFormat = "2006-01-02T15:04:05.999999"
const timestampField = "timestamp"
const severityField = "severity"

type QuerySpec struct {
	Syntax    syntax.Syntax `json:"syntax"`
	Limit     int           `json:"limit"`
	Tier      tier.Tier     `json:"tier"`
	StartDate time.Time     `json:"start_date"`
	EndDate   time.Time     `json:"end_date"`
}

type Log struct {
	Time     time.Time
	Severity string
	Message  string
}

type Query struct {
	Query    string
	Metadata *map[string]any
}

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

func getKeyValue(s []any, key string) (string, error) {

	idx := slices.IndexFunc(s, func(m any) bool { return m.(map[string]any)["key"].(string) == key })

	if idx < 0 {
		return "", fmt.Errorf("Cannot find value for key: '%s'", key)
	}

	return s[idx].(map[string]any)["value"].(string), nil
}

func parseResult(m map[string]any) (Log, error) {

	metadata := m["metadata"].([]any)
	userData := m["user_data"].(string)

	timestamp, err := getKeyValue(metadata, timestampField)
	if err != nil {
		return Log{}, fmt.Errorf("Cannot parse timestamp: %w", err)
	}
	severity, err := getKeyValue(metadata, severityField)
	if err != nil {
		return Log{}, fmt.Errorf("Cannot parse severity: %w", err)
	}

	t, err := time.ParseInLocation(timeFormat, timestamp, time.Local)
	if err != nil {
		return Log{}, fmt.Errorf("Cannot parse timestamp: %w", err)
	}

	ud := make(map[string]any)
	if err := json.Unmarshal([]byte(userData), &ud); err != nil {
		return Log{}, fmt.Errorf("Cannot Unmarshal User Data: %w", err)
	}

	log := Log{
		Time:     t,
		Severity: severity,
	}

	if message, ok := ud["message"]; ok {
		log.Message = message.(string)
	}

	return log, nil

}

func parseResponse(response io.Reader) ([]Log, error) {

	logs := []Log{}

	// response shouldn't have lines larger than 4K
	scanner := bufio.NewScanner(response)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, dataPrefix) {
			continue
		}

		d := line[len(dataPrefix):]
		data := make(map[string]any)

		if err := json.Unmarshal([]byte(d), &data); err != nil {
			return nil, fmt.Errorf("Cannot Unmarshal Data: %w", err)
		}

		if val, ok := data["result"]; ok {

			results := val.(map[string]any)["results"]

			for _, result := range results.([]any) {
				r := result.(map[string]any)

				l, err := parseResult(r)

				if err != nil {
					return nil, fmt.Errorf("Cannot parse result: %w", err)
				}

				logs = append(logs, l)

			}

		}

	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Sort logs
	sort.Slice(logs, func(i, j int) bool { return logs[i].Time.Compare(logs[j].Time) < 0 })

	return logs, nil
}

func QueryLogs(addr, token, query string, spec QuerySpec) ([]Log, error) {

	q := Query{Query: query}

	if spec != (QuerySpec{}) {
		meta := make(map[string]any)
		structToMap(spec, &meta)

		q.Metadata = &meta
	}

	j, err := json.Marshal(q)

	payload := bytes.NewBuffer(j)

	c := http.Client{Timeout: time.Duration(1) * time.Second}
	req, err := http.NewRequest("POST", addr, payload)
	if err != nil {
		return nil, fmt.Errorf("Cannot create POST request: %w", err)
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("authorization", "Bearer "+token)

	resp, err := c.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Cannot POST data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return nil, fmt.Errorf("Cannot read body: %w", err)
		}

		return nil, fmt.Errorf("Got HTTP error code: %d, message: '%s'", resp.StatusCode, body)
	}

	logs, err := parseResponse(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("Error when parsing results: %w", err)
	}

	return logs, nil

}

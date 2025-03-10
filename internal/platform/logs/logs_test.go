package logs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wooyey/iclogs/internal/platform/logs/syntax"
	"github.com/wooyey/iclogs/internal/platform/logs/tier"
)

type Metadata struct {
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	DefaultSource    string    `json:"default_source"`
	Tier             tier.Tier
	Syntax           syntax.Syntax
	Limit            int32
	StrictValidation bool `json:"strict_fields_validation"`
}

type LogsQuery struct {
	Query    string
	Metadata Metadata
}

var respNoLogs = `: success
data: {"query_id":{"query_id":"d7e258c9-2a3b-442b-947e-1c62f149321f"}}

:

:

:

:

:

:

:

:

:

:
`

var respWarnings = `: success
data: {"query_id":{"query_id":"ee49715e-4337-4c2a-a058-2bc38bcc33ea"}}

: success
data: {"warning":{"compile_warning":{"warning_message":"keypath does not exist\n'w.e' in line 0 at column 0"}}}

: success
data: {"warning":{"compile_warning":{"warning_message":"tokens less than 4 bytes or more than 64 bytes in UTF-8 are not indexed and will likely be excluded from the query\n'12' in line 0 at column 22"}}}

: success
data: {"warning":{"compile_warning":{"warning_message":"tokens less than 4 bytes or more than 64 bytes in UTF-8 are not indexed and will likely be excluded from the query\n'2' in line 0 at column 25"}}}
`

var respResults = `: success
data: {"query_id":{"query_id":"3b131b87-9b14-43e3-94fb-611967d9d62b"}}

: success
data: {"result":{"results":[{"metadata":[{"key":"timestamp","value":"2025-01-11T18:52:23.026304"},{"key":"severity","value":"Info"},{"key":"logid","value":"2875ffa6-d102-4043-b9dd-a8daf3f7d3c7"},{"key":"priorityclass","value":"high"},{"key":"processingOutputTimestampNanos","value":"1736621543823000000"},{"key":"processingOutputTimestampMicros","value":"1736621543823000"},{"key":"timestampMicros","value":"1736621543026304"},{"key":"ingressTimestamp","value":"2025-01-11T18:52:23.403000"},{"key":"templateid","value":"aca6bdbb-12ed-907e-0585-782649c126c8"},{"key":"branchid","value":"6d2a7580-6d52-a552-a9fa-35bbeee8545b"}],"labels":[{"key":"applicationname","value":"some-observe"},{"key":"subsystemname","value":"some-agent"},{"key":"computername","value":""},{"key":"threadid","value":""},{"key":"ipaddress","value":""}],"user_data":"{\"node_name\":\"10.10.10.10\",\"kubernetes\":{\"annotations\":{\"kubectl.kubernetes.io/restartedAt\":\"2024-03-15T11:44:11+05:30\",\"kubernetes.io/config.seen\":\"2025-01-06T08:44:29.371412369Z\",\"kubernetes.io/config.source\":\"api\"},\"container_hash\":\"url.com/ext/some/agent@sha256:7594347727a76fab1b6759575d84389ac1788bff6782046b330c730d67db790c\",\"container_image\":\"url.com/ext/some/agent:latest\",\"container_name\":\"some-agent\",\"docker_id\":\"7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7\",\"host\":\"10.10.10.10\",\"labels\":{\"app\":\"some-agent\",\"controller-revision-hash\":\"f69c8df74\",\"pod-template-generation\":\"12\"},\"namespace_name\":\"some-observe\",\"pod_id\":\"3ba098ee-cc88-4cb7-b986-f61e182b6936\",\"pod_name\":\"some-agent-c7gz7\"},\"tag\":\"kube.var.log.containers.some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log\",\"meta\":{\"cluster_name\":\"wml-core-dallas-yp-qa\"},\"stream\":\"stdout\",\"logtag\":\"F\",\"message\":\"2025-01-11 18:52:23.025, 347267.347747, Information, Example message\",\"file\":\"/var/log/containers/some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log\"}"},{"metadata":[{"key":"timestamp","value":"2025-01-11T18:52:23.026360"},{"key":"severity","value":"Info"},{"key":"logid","value":"dc1a1257-a13a-4e9a-beca-f4ed5bc8cc2a"},{"key":"priorityclass","value":"high"},{"key":"processingOutputTimestampNanos","value":"1736621543823000000"},{"key":"processingOutputTimestampMicros","value":"1736621543823000"},{"key":"timestampMicros","value":"1736621543026360"},{"key":"ingressTimestamp","value":"2025-01-11T18:52:23.403000"},{"key":"templateid","value":"aca6bdbb-12ed-907e-0585-782649c126c8"},{"key":"branchid","value":"6d2a7580-6d52-a552-a9fa-35bbeee8545b"}],"labels":[{"key":"applicationname","value":"some-observe"},{"key":"subsystemname","value":"some-agent"},{"key":"computername","value":""},{"key":"threadid","value":""},{"key":"ipaddress","value":""}],"user_data":"{\"node_name\":\"10.10.10.10\",\"kubernetes\":{\"annotations\":{\"kubectl.kubernetes.io/restartedAt\":\"2024-03-15T11:44:11+05:30\",\"kubernetes.io/config.seen\":\"2025-01-06T08:44:29.371412369Z\",\"kubernetes.io/config.source\":\"api\"},\"container_hash\":\"url.com/ext/some/agent@sha256:7594347727a76fab1b6759575d84389ac1788bff6782046b330c730d67db790c\",\"container_image\":\"url.com/ext/some/agent:latest\",\"container_name\":\"some-agent\",\"docker_id\":\"7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7\",\"host\":\"10.10.10.10\",\"labels\":{\"app\":\"some-agent\",\"controller-revision-hash\":\"f69c8df74\",\"pod-template-generation\":\"12\"},\"namespace_name\":\"some-observe\",\"pod_id\":\"3ba098ee-cc88-4cb7-b986-f61e182b6936\",\"pod_name\":\"some-agent-c7gz7\"},\"tag\":\"kube.var.log.containers.some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log\",\"meta\":{\"cluster_name\":\"wml-core-dallas-yp-qa\"},\"stream\":\"stdout\",\"logtag\":\"F\",\"message\":\"2025-01-11 18:52:23.026, 347267.347747, Information, Next message\",\"file\":\"/var/log/containers/some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log\"}"}]}}

: success
data: {"result":{"results":[{"metadata":[{"key":"timestamp","value":"2025-01-11T18:52:21.026304"},{"key":"severity","value":"Debug"},{"key":"logid","value":"2875ffa6-d102-4043-b9dd-a8daf3f7d3c7"},{"key":"priorityclass","value":"high"},{"key":"processingOutputTimestampNanos","value":"1736621543823000000"},{"key":"processingOutputTimestampMicros","value":"1736621543823000"},{"key":"timestampMicros","value":"1736621543026304"},{"key":"ingressTimestamp","value":"2025-01-11T18:52:21.403000"},{"key":"templateid","value":"aca6bdbb-12ed-907e-0585-782649c126c8"},{"key":"branchid","value":"6d2a7580-6d52-a552-a9fa-35bbeee8545b"}],"labels":[{"key":"applicationname","value":"some-observe"},{"key":"subsystemname","value":"some-agent"},{"key":"computername","value":""},{"key":"threadid","value":""},{"key":"ipaddress","value":""}],"user_data":"{\"node_name\":\"10.10.10.10\",\"kubernetes\":{\"annotations\":{\"kubectl.kubernetes.io/restartedAt\":\"2024-03-15T11:44:11+05:30\",\"kubernetes.io/config.seen\":\"2025-01-06T08:44:29.371412369Z\",\"kubernetes.io/config.source\":\"api\"},\"container_hash\":\"url.com/ext/some/agent@sha256:7594347727a76fab1b6759575d84389ac1788bff6782046b330c730d67db790c\",\"container_image\":\"url.com/ext/some/agent:latest\",\"container_name\":\"some-agent\",\"docker_id\":\"7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7\",\"host\":\"10.10.10.10\",\"labels\":{\"app\":\"some-agent\",\"controller-revision-hash\":\"f69c8df74\",\"pod-template-generation\":\"12\"},\"namespace_name\":\"some-observe\",\"pod_id\":\"3ba098ee-cc88-4cb7-b986-f61e182b6936\",\"pod_name\":\"some-agent-c7gz7\"},\"tag\":\"kube.var.log.containers.some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log\",\"meta\":{\"cluster_name\":\"wml-core-dallas-yp-qa\"},\"stream\":\"stdout\",\"logtag\":\"F\",\"message\":\"2025-01-11 18:52:23.025, 347267.347747, Debug, Example message first\",\"file\":\"/var/log/containers/some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log\"}"},{"metadata":[{"key":"timestamp","value":"2025-01-11T18:52:21.026360"},{"key":"severity","value":"Info"},{"key":"logid","value":"dc1a1257-a13a-4e9a-beca-f4ed5bc8cc2a"},{"key":"priorityclass","value":"high"},{"key":"processingOutputTimestampNanos","value":"1736621543823000000"},{"key":"processingOutputTimestampMicros","value":"1736621543823000"},{"key":"timestampMicros","value":"1736621543026360"},{"key":"ingressTimestamp","value":"2025-01-11T18:52:23.403000"},{"key":"templateid","value":"aca6bdbb-12ed-907e-0585-782649c126c8"},{"key":"branchid","value":"6d2a7580-6d52-a552-a9fa-35bbeee8545b"}],"labels":[{"key":"applicationname","value":"some-observe"},{"key":"subsystemname","value":"some-agent"},{"key":"computername","value":""},{"key":"threadid","value":""},{"key":"ipaddress","value":""}],"user_data":"{\"node_name\":\"10.10.10.10\",\"kubernetes\":{\"annotations\":{\"kubectl.kubernetes.io/restartedAt\":\"2024-03-15T11:44:11+05:30\",\"kubernetes.io/config.seen\":\"2025-01-06T08:44:29.371412369Z\",\"kubernetes.io/config.source\":\"api\"},\"container_hash\":\"url.com/ext/some/agent@sha256:7594347727a76fab1b6759575d84389ac1788bff6782046b330c730d67db790c\",\"container_image\":\"url.com/ext/some/agent:latest\",\"container_name\":\"some-agent\",\"docker_id\":\"7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7\",\"host\":\"10.10.10.10\",\"labels\":{\"app\":\"some-agent\",\"controller-revision-hash\":\"f69c8df74\",\"pod-template-generation\":\"12\"},\"namespace_name\":\"some-observe\",\"pod_id\":\"3ba098ee-cc88-4cb7-b986-f61e182b6936\",\"pod_name\":\"some-agent-c7gz7\"},\"tag\":\"kube.var.log.containers.some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log\",\"meta\":{\"cluster_name\":\"wml-core-dallas-yp-qa\"},\"stream\":\"stdout\",\"logtag\":\"F\",\"message\":\"2025-01-11 18:52:23.026, 347267.347747, Information, second message\",\"file\":\"/var/log/containers/some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log\"}"}]}}
`

// 1MB line simulation
var respLongLine = respResults + `
: success
` + strings.Repeat(" ", 1024*1024)

var respFailParse = `{
	"errors": [
		{
			"code": "bad_request_or_unspecified",
			"message": "Failed to deserialize JSON"
		}
	],
	"trace": "c37f5a58-8ee9-4c58-ae13-db03409853fa",
	"status_code": 400
}`

func mockServer(response string) *httptest.Server {
	f := func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if !slices.Contains(r.Header["Authorization"], "Bearer Good_Token") {
			w.WriteHeader(403)
			fmt.Fprint(w, "Access denied!")
			return
		}

		if !slices.Contains(r.Header["Content-Type"], "application/json") {
			w.WriteHeader(400)
			fmt.Fprint(w, respFailParse)
			return
		}

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		var q LogsQuery
		err := dec.Decode(&q)

		if err != nil {
			w.WriteHeader(400)
			fmt.Fprint(w, respFailParse)
			return
		}

		if q.Query == "" || q.Query != "Good Query" {
			w.WriteHeader(400)
			fmt.Fprint(w, respFailParse)
			return
		}

		w.WriteHeader(200)
		fmt.Fprint(w, response)
	}

	return httptest.NewServer(http.HandlerFunc(f))
}

var expectedLabels = []string{
	"applicationname:\"some-observe\"",
	"subsystemname:\"some-agent\"",
	"computername:\"\"",
	"threadid:\"\"",
	"ipaddress:\"\"",
}

var expectedLogs = []Log{
	{
		Time:     time.Date(2025, 1, 11, 18, 52, 21, 26304000, time.Local),
		Severity: "Debug",
		UserData: `{"node_name":"10.10.10.10","kubernetes":{"annotations":{"kubectl.kubernetes.io/restartedAt":"2024-03-15T11:44:11+05:30","kubernetes.io/config.seen":"2025-01-06T08:44:29.371412369Z","kubernetes.io/config.source":"api"},"container_hash":"url.com/ext/some/agent@sha256:7594347727a76fab1b6759575d84389ac1788bff6782046b330c730d67db790c","container_image":"url.com/ext/some/agent:latest","container_name":"some-agent","docker_id":"7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7","host":"10.10.10.10","labels":{"app":"some-agent","controller-revision-hash":"f69c8df74","pod-template-generation":"12"},"namespace_name":"some-observe","pod_id":"3ba098ee-cc88-4cb7-b986-f61e182b6936","pod_name":"some-agent-c7gz7"},"tag":"kube.var.log.containers.some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log","meta":{"cluster_name":"wml-core-dallas-yp-qa"},"stream":"stdout","logtag":"F","message":"2025-01-11 18:52:23.025, 347267.347747, Debug, Example message first","file":"/var/log/containers/some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log"}`,
		Labels:   expectedLabels,
	},
	{
		Time:     time.Date(2025, 1, 11, 18, 52, 21, 26360000, time.Local),
		Severity: "Info",
		UserData: `{"node_name":"10.10.10.10","kubernetes":{"annotations":{"kubectl.kubernetes.io/restartedAt":"2024-03-15T11:44:11+05:30","kubernetes.io/config.seen":"2025-01-06T08:44:29.371412369Z","kubernetes.io/config.source":"api"},"container_hash":"url.com/ext/some/agent@sha256:7594347727a76fab1b6759575d84389ac1788bff6782046b330c730d67db790c","container_image":"url.com/ext/some/agent:latest","container_name":"some-agent","docker_id":"7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7","host":"10.10.10.10","labels":{"app":"some-agent","controller-revision-hash":"f69c8df74","pod-template-generation":"12"},"namespace_name":"some-observe","pod_id":"3ba098ee-cc88-4cb7-b986-f61e182b6936","pod_name":"some-agent-c7gz7"},"tag":"kube.var.log.containers.some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log","meta":{"cluster_name":"wml-core-dallas-yp-qa"},"stream":"stdout","logtag":"F","message":"2025-01-11 18:52:23.026, 347267.347747, Information, second message","file":"/var/log/containers/some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log"}`,
		Labels:   expectedLabels,
	},
	{
		Time:     time.Date(2025, 1, 11, 18, 52, 23, 26304000, time.Local),
		Severity: "Info",
		UserData: `{"node_name":"10.10.10.10","kubernetes":{"annotations":{"kubectl.kubernetes.io/restartedAt":"2024-03-15T11:44:11+05:30","kubernetes.io/config.seen":"2025-01-06T08:44:29.371412369Z","kubernetes.io/config.source":"api"},"container_hash":"url.com/ext/some/agent@sha256:7594347727a76fab1b6759575d84389ac1788bff6782046b330c730d67db790c","container_image":"url.com/ext/some/agent:latest","container_name":"some-agent","docker_id":"7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7","host":"10.10.10.10","labels":{"app":"some-agent","controller-revision-hash":"f69c8df74","pod-template-generation":"12"},"namespace_name":"some-observe","pod_id":"3ba098ee-cc88-4cb7-b986-f61e182b6936","pod_name":"some-agent-c7gz7"},"tag":"kube.var.log.containers.some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log","meta":{"cluster_name":"wml-core-dallas-yp-qa"},"stream":"stdout","logtag":"F","message":"2025-01-11 18:52:23.025, 347267.347747, Information, Example message","file":"/var/log/containers/some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log"}`,
		Labels:   expectedLabels,
	},
	{
		Time:     time.Date(2025, 1, 11, 18, 52, 23, 26360000, time.Local),
		Severity: "Info",
		UserData: `{"node_name":"10.10.10.10","kubernetes":{"annotations":{"kubectl.kubernetes.io/restartedAt":"2024-03-15T11:44:11+05:30","kubernetes.io/config.seen":"2025-01-06T08:44:29.371412369Z","kubernetes.io/config.source":"api"},"container_hash":"url.com/ext/some/agent@sha256:7594347727a76fab1b6759575d84389ac1788bff6782046b330c730d67db790c","container_image":"url.com/ext/some/agent:latest","container_name":"some-agent","docker_id":"7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7","host":"10.10.10.10","labels":{"app":"some-agent","controller-revision-hash":"f69c8df74","pod-template-generation":"12"},"namespace_name":"some-observe","pod_id":"3ba098ee-cc88-4cb7-b986-f61e182b6936","pod_name":"some-agent-c7gz7"},"tag":"kube.var.log.containers.some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log","meta":{"cluster_name":"wml-core-dallas-yp-qa"},"stream":"stdout","logtag":"F","message":"2025-01-11 18:52:23.026, 347267.347747, Information, Next message","file":"/var/log/containers/some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log"}`,
		Labels:   expectedLabels,
	},
}

var userData = map[string]string{
	"message":     `{"node_name":"10.10.10.10","kubernetes":{"annotations":{"kubectl.kubernetes.io/restartedAt":"2024-03-15T11:44:11+05:30","kubernetes.io/config.seen":"2025-01-06T08:44:29.371412369Z","kubernetes.io/config.source":"api"},"container_hash":"url.com/ext/some/agent@sha256:7594347727a76fab1b6759575d84389ac1788bff6782046b330c730d67db790c","container_image":"url.com/ext/some/agent:latest","container_name":"some-agent","docker_id":"7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7","host":"10.10.10.10","labels":{"app":"some-agent","controller-revision-hash":"f69c8df74","pod-template-generation":"12"},"namespace_name":"some-observe","pod_id":"3ba098ee-cc88-4cb7-b986-f61e182b6936","pod_name":"some-agent-c7gz7"},"tag":"kube.var.log.containers.some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log","meta":{"cluster_name":"wml-core-dallas-yp-qa"},"stream":"stdout","logtag":"F","message":"2025-01-11 18:52:23.025, 347267.347747, Debug, Example message first","file":"/var/log/containers/some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log"}`,
	"message_obj": `{"node_name":"10.10.10.10","kubernetes":{"annotations":{"kubectl.kubernetes.io/restartedAt":"2024-03-15T11:44:11+05:30","kubernetes.io/config.seen":"2025-01-06T08:44:29.371412369Z","kubernetes.io/config.source":"api"},"container_hash":"url.com/ext/some/agent@sha256:7594347727a76fab1b6759575d84389ac1788bff6782046b330c730d67db790c","container_image":"url.com/ext/some/agent:latest","container_name":"some-agent","docker_id":"7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7","host":"10.10.10.10","labels":{"app":"some-agent","controller-revision-hash":"f69c8df74","pod-template-generation":"12"},"namespace_name":"some-observe","pod_id":"3ba098ee-cc88-4cb7-b986-f61e182b6936","pod_name":"some-agent-c7gz7"},"tag":"kube.var.log.containers.some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log","meta":{"cluster_name":"wml-core-dallas-yp-qa"},"stream":"stdout","logtag":"F","message_obj":{"msg":"2025-01-11 18:52:23.025, 347267.347747, Information, Example message","level":"debug","ts":"2025-01-01T10:44:00.082Z","caller":"runtime/runtime.go:83"},"file":"/var/log/containers/some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log"}`,
	"log":         `{"node_name":"10.10.10.10","kubernetes":{"annotations":{"kubectl.kubernetes.io/restartedAt":"2024-03-15T11:44:11+05:30","kubernetes.io/config.seen":"2025-01-06T08:44:29.371412369Z","kubernetes.io/config.source":"api"},"container_hash":"url.com/ext/some/agent@sha256:7594347727a76fab1b6759575d84389ac1788bff6782046b330c730d67db790c","container_image":"url.com/ext/some/agent:latest","container_name":"some-agent","docker_id":"7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7","host":"10.10.10.10","labels":{"app":"some-agent","controller-revision-hash":"f69c8df74","pod-template-generation":"12"},"namespace_name":"some-observe","pod_id":"3ba098ee-cc88-4cb7-b986-f61e182b6936","pod_name":"some-agent-c7gz7"},"tag":"kube.var.log.containers.some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log","meta":{"cluster_name":"wml-core-dallas-yp-qa"},"stream":"stdout","logtag":"F","log":"2025-01-11 18:52:23.025, 347267.347747, Debug, Example message first","file":"/var/log/containers/some-agent-c7gz7_some-observe_some-agent-7ca9add76b8a725f0da735a948cb133965de0eb36ac31d6252060eaaaabb0fb7.log"}`,
}

func TestQueryLogs(t *testing.T) {

	testCases := []struct {
		name     string
		token    string
		query    string
		response string
		spec     QuerySpec
		want     []Log
		err      any
	}{
		{name: "GoodToken", token: "Good_Token", query: "Good Query", spec: QuerySpec{Syntax: syntax.Lucene}, response: respResults, want: expectedLogs, err: nil},
		{name: "NoLogs", token: "Good_Token", query: "Good Query", spec: QuerySpec{Syntax: syntax.Lucene}, response: respNoLogs, want: []Log{}, err: nil},
		{name: "OnlyWarnings", token: "Good_Token", query: "Good Query", spec: QuerySpec{Syntax: syntax.Lucene}, response: respWarnings, want: []Log{}, err: nil},
		{name: "LongLine", token: "Good_Token", query: "Good Query", spec: QuerySpec{Syntax: syntax.Lucene}, response: respLongLine, want: expectedLogs, err: nil},
	}

	for _, tt := range testCases {

		t.Run(tt.name, func(t *testing.T) {
			server := mockServer(tt.response)
			defer server.Close()

			got, err := QueryLogs(server.URL, tt.token, tt.query, tt.spec)

			if tt.err == nil && err != nil {
				t.Errorf("Got error: '%v'", err)
				return
			}

			if tt.err != nil && err != tt.err {
				t.Errorf("Didn't get error. Got: '%v', want: '%v'", err, tt.err)
				return
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("\nGot:\t'%+v',\n Want:\t'%+v'", got, tt.want)
			}

		})
	}

}

func TestGetMessage(t *testing.T) {

	testCases := []struct {
		name     string
		userData string
		keyNames []string
		want     string
		err      bool
	}{
		{name: "Message", userData: userData["message"], keyNames: []string{"message"}, want: "2025-01-11 18:52:23.025, 347267.347747, Debug, Example message first", err: false},
		{name: "MessageObj", userData: userData["message_obj"], keyNames: []string{"message_obj.msg"}, want: "2025-01-11 18:52:23.025, 347267.347747, Information, Example message", err: false},
		{name: "Error", userData: userData["message"], keyNames: []string{"message_obj.msg"}, want: "", err: true},
		{name: "Log", userData: userData["log"], keyNames: []string{"message_obj.msg", "message", "log"}, want: "2025-01-11 18:52:23.025, 347267.347747, Debug, Example message first", err: false},
	}

	for _, tt := range testCases {

		t.Run(tt.name, func(t *testing.T) {

			got, err := GetMessage(&tt.userData, &tt.keyNames)

			if !tt.err && err != nil {
				t.Errorf("\nGot an error:\t'%v'", err)
				return
			}

			if tt.err && err == nil {
				t.Error("\nShould get an error!")
				return
			}

			if got != tt.want {
				t.Errorf("\nGot:\t'%s'\nWant:\t'%s'", got, tt.want)
			}
		})
	}
}

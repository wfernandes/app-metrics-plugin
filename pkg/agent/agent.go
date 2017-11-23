package agent

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/wfernandes/apps-metrics-plugin/pkg/parser"
)

type MetricOuput struct {
	Instance int
	Output   string
	Error    string
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Parser interface {
	Parse([]byte) ([]byte, error)
}

type Agent struct {
	app    *plugin_models.GetAppModel
	path   string
	client HTTPClient
	parser Parser
}

type AgentOpt func(*Agent)

func WithClient(c HTTPClient) AgentOpt {
	return func(a *Agent) {
		a.client = c
	}
}

func WithParser(p Parser) AgentOpt {
	return func(a *Agent) {
		a.parser = p
	}
}

func WithMetricsPath(p string) AgentOpt {
	return func(a *Agent) {
		a.path = p
	}
}

func New(model *plugin_models.GetAppModel, opts ...AgentOpt) *Agent {
	a := &Agent{
		app:    model,
		path:   "/debug/metrics",
		client: &http.Client{Timeout: 5 * time.Second},
		parser: parser.NewNoOp(),
	}

	for _, o := range opts {
		o(a)
	}

	return a
}

func (a *Agent) GetMetrics() ([]MetricOuput, error) {
	url, err := a.buildURL()
	if err != nil {
		return nil, err
	}
	outputs := make([]MetricOuput, 0, a.app.RunningInstances)
	results := make(chan MetricOuput, a.app.RunningInstances)
	defer close(results)

	for i, instance := range a.app.Instances {
		if instance.State != "running" {
			continue
		}
		go func(idx int) {
			mo := a.makeRequest(url, idx)
			results <- mo
		}(i)
	}

	for r := range results {
		outputs = append(outputs, r)
		if len(outputs) == a.app.RunningInstances {
			return outputs, nil
		}
	}

	return outputs, nil
}

func (a *Agent) makeRequest(url string, i int) MetricOuput {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return buildMetricOutput(i, "", err)
	}
	request.Header.Add("X-CF-APP-INSTANCE", fmt.Sprintf("%s:%d", a.app.Guid, i))

	resp, err := a.client.Do(request)
	if err != nil {
		return buildMetricOutput(i, "", err)
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return buildMetricOutput(i, "", err)
	}

	parsed, err := a.parser.Parse(bytes)
	if err != nil {
		parsed = bytes
	}

	return buildMetricOutput(i, string(parsed), nil)
}

func (a *Agent) buildURL() (string, error) {
	// TODO: be able to parse multiple routes
	if len(a.app.Routes) == 0 {
		return "", errors.New("app does not have any routes to hit")
	}
	route := a.app.Routes[0]
	var url string
	if route.Host == "" {
		url = "http://" + route.Domain.Name + a.path
	} else {
		url = "http://" + route.Host + "." + route.Domain.Name + a.path
	}
	return url, nil
}

func buildMetricOutput(instance int, output string, err error) MetricOuput {
	mo := MetricOuput{
		Instance: instance,
		Output:   output,
	}
	if err != nil {
		mo.Error = err.Error()
	}
	return mo
}

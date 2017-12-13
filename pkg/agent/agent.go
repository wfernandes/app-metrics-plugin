package agent

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"code.cloudfoundry.org/cli/plugin/models"
)

type InstanceMetric struct {
	Instance int
	Error    string
	Metrics  map[string]interface{}
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Parser interface {
	Parse([]byte) (map[string]interface{}, error)
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

func WithMetricsPath(p string) AgentOpt {
	return func(a *Agent) {
		a.path = p
	}
}

func New(m *plugin_models.GetAppModel, p Parser, opts ...AgentOpt) *Agent {
	a := &Agent{
		app:    m,
		parser: p,
		path:   "/debug/metrics",
		client: &http.Client{Timeout: 5 * time.Second},
	}

	for _, o := range opts {
		o(a)
	}

	return a
}

func (a *Agent) GetMetrics(ctx context.Context) (outputs []InstanceMetric, err error) {
	url, err := a.buildURL()
	if err != nil {
		return nil, err
	}
	outputs = make([]InstanceMetric, 0, a.app.RunningInstances)
	defer func() {
		// make sure the output is sorted. we used named return values here because of this.
		sort.Sort(byInstance(outputs))
	}()

	results := make(chan *InstanceMetric, a.app.RunningInstances)

	for i, instance := range a.app.Instances {
		if instance.State != "running" {
			continue
		}
		// Need a better way to clean up this go routine
		go func(idx int) {
			// TODO makeRequest should return a channel so once its done, it can be responsible for closing it.
			mo := a.makeRequest(url, idx, ctx)
			results <- mo
		}(i)
	}

	for {
		select {
		case r := <-results:
			outputs = append(outputs, *r)
			if len(outputs) == a.app.RunningInstances {
				return outputs, nil
			}
		case <-ctx.Done():
			return outputs, ctx.Err()
		}
	}

	return outputs, nil
}

func (a *Agent) makeRequest(url string, i int, ctx context.Context) *InstanceMetric {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return &InstanceMetric{Instance: i, Error: err.Error()}
	}
	request.Header.Add("X-CF-APP-INSTANCE", fmt.Sprintf("%s:%d", a.app.Guid, i))
	request = request.WithContext(ctx)

	resp, err := a.client.Do(request)
	if err != nil {
		return &InstanceMetric{Instance: i, Error: err.Error()}
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &InstanceMetric{Instance: i, Error: err.Error()}
	}

	metrics, err := a.parser.Parse(bytes)
	if err != nil {
		return &InstanceMetric{Instance: i, Error: fmt.Sprintf("unable to parse response: %s", err)}
	}

	return &InstanceMetric{Instance: i, Metrics: metrics}
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

type byInstance []InstanceMetric

func (s byInstance) Len() int {
	return len(s)
}
func (s byInstance) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byInstance) Less(i, j int) bool {
	return s[i].Instance < s[j].Instance
}

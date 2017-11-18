package agent

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/plugin/models"
)

type MetricOuput struct {
	Instance int
	Output   string
	Error    string
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Agent struct {
	app    *plugin_models.GetAppModel
	path   string
	token  string
	client HTTPClient
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

func New(model *plugin_models.GetAppModel, token string, opts ...AgentOpt) *Agent {
	a := &Agent{
		app:    model,
		path:   "/debug/metrics",
		token:  token,
		client: &http.Client{Timeout: 10 * time.Second},
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
	outputs := make([]MetricOuput, 0, a.app.InstanceCount)

	for i, instance := range a.app.Instances {
		if instance.State != "running" {
			continue
		}
		request, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			outputs = append(outputs, MetricOuput{
				Instance: i,
				Error:    err.Error(),
			})
			continue
		}
		request.Header.Add("X-CF-APP-INSTANCE", fmt.Sprintf("%s:%d", a.app.Guid, i))
		request.Header.Add("Authorization", a.token)
		outputs = append(outputs, a.doRequest(request, i))

	}
	return outputs, nil
}

func (a *Agent) doRequest(request *http.Request, i int) MetricOuput {

	resp, err := a.client.Do(request)
	if err != nil {
		return MetricOuput{
			Instance: i,
			Error:    err.Error(),
		}
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return MetricOuput{
			Instance: i,
			Error:    err.Error(),
		}
	}
	return MetricOuput{
		Instance: i,
		Output:   string(bytes),
	}
}

func (a *Agent) buildURL() (string, error) {
	// TODO: be able to parse multiple routes
	if len(a.app.Routes) == 0 {
		return "", errors.New("app does not have any routes to hit")
	}
	route := a.app.Routes[0]
	var url string
	if route.Host == "" {
		url = route.Domain.Name + a.path
	} else {
		url = route.Host + "." + route.Domain.Name + a.path
	}
	return url, nil
}

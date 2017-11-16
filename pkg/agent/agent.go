package agent

import (
	"io/ioutil"
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/plugin/models"
)

type Agent struct {
	app    *plugin_models.GetAppModel
	path   string
	client *http.Client
}

func New(model *plugin_models.GetAppModel, metricsPath string) *Agent {

	c := &http.Client{
		Timeout: 10 * time.Second,
	}
	return &Agent{
		app:    model,
		path:   metricsPath,
		client: c,
	}
}

func (a *Agent) GetMetrics() (string, error) {
	url := a.app.Routes[0].Domain.Name + a.path
	resp, err := a.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	return string(bytes), err
}

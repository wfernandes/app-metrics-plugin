package main

import (
	"encoding/json"
	"os"

	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/plugin"
	"github.com/wfernandes/apps-metrics-plugin/pkg/agent"
)

type AppsMetricsPlugin struct {
	ui terminal.UI
}

func (c *AppsMetricsPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	traceLogger := trace.NewLogger(os.Stdout, true, os.Getenv("CF_TRACE"), "")
	c.ui = terminal.NewUI(os.Stdin, os.Stdout, terminal.NewTeePrinter(os.Stdout), traceLogger)

	switch args[0] {
	case "apps-metrics":

		if len(args) < 2 {
			c.ui.Say(c.GetMetadata().Commands[0].UsageDetails.Usage)
			return
		}

		c.getMetrics(cliConnection, args)
	case "CLI-MESSAGE-UNINSTALL":
		c.ui.Say("Thank you for using apps-metrics")
	}

}

func (c *AppsMetricsPlugin) getMetrics(cliConnection plugin.CliConnection, args []string) {
	app, err := cliConnection.GetApp(args[1])
	if err != nil {
		c.ui.Failed(err.Error())
		return
	}

	token, err := cliConnection.AccessToken()
	if err != nil {
		c.ui.Failed(err.Error())
		return
	}

	fc, err := parseArguments(args)
	if err != nil {
		c.ui.Failed(err.Error())
		return
	}

	var client *agent.Agent
	if fc.IsSet("endpoint") {
		client = agent.New(&app, token, agent.WithMetricsPath(fc.String("endpoint")))
	} else {
		client = agent.New(&app, token)
	}

	metrics, err := client.GetMetrics()
	if err != nil {
		c.ui.Failed("unable to get metrics: %s\n", err)
	}

	bytes, err := json.Marshal(metrics)
	if err != nil {
		c.ui.Warn("unable to marshal metrics: %s\n", err)
	}
	c.ui.Say("%s\n", string(bytes))
}

func parseArguments(args []string) (flags.FlagContext, error) {
	fc := flags.New()
	fc.NewStringFlag("endpoint", "e", "Path of the metrics endpoint")

	err := fc.Parse(args...)
	if err != nil {
		return nil, err
	}
	return fc, nil
}

func (c *AppsMetricsPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "AppsMetricsPlugin",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 0,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 7,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "apps-metrics",
				HelpText: "Hits the metrics endpoint across all your app instances",

				UsageDetails: plugin.Usage{
					Usage: "cf apps-metrics APP_NAME",
					Options: map[string]string{
						"endpoint": "metrics endpoint",
					},
				},
			},
		},
	}
}

func main() {
	plugin.Start(new(AppsMetricsPlugin))
}

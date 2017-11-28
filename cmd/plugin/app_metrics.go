package main

import (
	"context"
	"encoding/json"
	"os"
	"text/template"

	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/plugin"
	"github.com/wfernandes/app-metrics-plugin/pkg/agent"
	"github.com/wfernandes/app-metrics-plugin/pkg/parser"
	"github.com/wfernandes/app-metrics-plugin/pkg/views"
)

type AppsMetricsPlugin struct {
	ui terminal.UI
}

func (c *AppsMetricsPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "AppMetricsPlugin",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 1,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 7,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "app-metrics",
				HelpText: "Hits the metrics endpoint across all your app instances",

				UsageDetails: plugin.Usage{
					Usage: "cf app-metrics APP_NAME",
					Options: map[string]string{
						"endpoint": "path of the metrics endpoint",
						"template": "path of the template files to render metrics",
						"raw":      "prints raw json output",
					},
				},
			},
		},
	}
}

func (c *AppsMetricsPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	traceLogger := trace.NewLogger(os.Stdout, true, os.Getenv("CF_TRACE"), "")
	c.ui = terminal.NewUI(os.Stdin, os.Stdout, terminal.NewTeePrinter(os.Stdout), traceLogger)

	switch args[0] {
	case "app-metrics":

		if len(args) < 2 {
			c.ui.Say(c.GetMetadata().Commands[0].UsageDetails.Usage)
			return
		}

		c.getMetrics(cliConnection, args)
	case "CLI-MESSAGE-UNINSTALL":
		c.ui.Say("Thank you for using app-metrics")
	}

}

func (c *AppsMetricsPlugin) getMetrics(cliConnection plugin.CliConnection, args []string) {
	// Verify we have access to the app
	app, err := cliConnection.GetApp(args[1])
	if err != nil {
		c.ui.Failed(err.Error())
		return
	}

	// Parse any flags that were provided
	fc, err := parseArguments(args)
	if err != nil {
		c.ui.Failed(err.Error())
		return
	}

	// Create the client that will GET the metrics. Currently, it is fixed
	// to Expvar style but other parsers can be written. For example, Prometheus.
	// We are forcing to ignore the `cmdline` and `memstats` properties as they
	// clutter the output.
	var client *agent.Agent
	if fc.IsSet("endpoint") {
		client = agent.New(
			&app,
			agent.WithParser(parser.NewExpvar([]string{"cmdline", "memstats"})),
			agent.WithMetricsPath(fc.String("endpoint")))
	} else {
		client = agent.New(
			&app,
			agent.WithParser(parser.NewExpvar([]string{"cmdline", "memstats"})))
	}

	// Make the request(s) and get the data
	metrics, err := client.GetMetrics(context.Background())
	if err != nil {
		c.ui.Failed("unable to get metrics: %s\n", err)
	}

	// Print json output when raw flag is specified
	if fc.IsSet("raw") {
		c.printDefault(metrics)
		return
	}

	// Present the data
	var view *views.View
	if fc.IsSet("template") {
		tmpl, err := template.ParseFiles(fc.String("template"))
		if err != nil {
			c.ui.Failed("unable to parse template files: %s\n", err)
			return
		}
		view = views.New(views.WithTemplate(tmpl))
	} else {
		view = views.New()
	}
	err = view.Present(metrics)
	if err != nil {
		c.ui.Warn(err.Error())
		c.printDefault(metrics)
	}
}

func (c *AppsMetricsPlugin) printDefault(metrics []agent.MetricOuput) {
	bytes, err := json.Marshal(metrics)
	if err != nil {
		c.ui.Warn("unable to marshal metrics: %s\n", err)
		return
	}
	c.ui.Say("%s\n", string(bytes))
}

func parseArguments(args []string) (flags.FlagContext, error) {
	fc := flags.New()
	fc.NewStringFlag("endpoint", "e", "Path of the metrics endpoint")
	fc.NewStringFlag("template", "t", "Path of the template files to render metrics")
	fc.NewBoolFlag("raw", "r", "Prints raw json output")

	err := fc.Parse(args...)
	if err != nil {
		return nil, err
	}
	return fc, nil
}

func main() {
	plugin.Start(new(AppsMetricsPlugin))
}

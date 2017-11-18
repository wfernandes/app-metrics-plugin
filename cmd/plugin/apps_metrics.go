package main

import (
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/wfernandes/apps-metrics-plugin/pkg/agent"
)

type AppsMetricsPlugin struct {
}

func (c *AppsMetricsPlugin) Run(cliConnection plugin.CliConnection, args []string) {

	if len(args) < 2 {
		fmt.Println("APP_NAME is required")
		return
	}

	appName := args[1]
	app, err := cliConnection.GetApp(appName)
	if err != nil {
		fmt.Println(err.Error())
	}

	// TODO: Add a test for this erroring out
	token, err := cliConnection.AccessToken()
	if err != nil {
		fmt.Println(err.Error())
	}
	client := agent.New(&app, token)
	metrics, err := client.GetMetrics()
	if err != nil {
		fmt.Printf("unable to get metrics: %s\n", err)
	}
	for _, m := range metrics {
		// TODO: Add a test to handle this err
		bytes, err := json.Marshal(m)
		if err != nil {
			fmt.Printf("unable to marshal metrics: %s\n", err)
		}
		fmt.Printf("%s\n", string(bytes))
	}

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
					Usage: "apps-metrics\n   cf apps-metrics APP_NAME",
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

package main

import (
	"fmt"

	"code.cloudfoundry.org/cli/plugin"
)

type AppsMetricsPlugin struct{}

func (c *AppsMetricsPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if len(args) < 2 {
		fmt.Println("APP_NAME is required")
		return
	}

	if args[0] == "apps-metrics" {
		fmt.Println("Running the apps-metrics plugin")
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

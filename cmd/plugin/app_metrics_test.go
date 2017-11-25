package main_test

import (
	"fmt"

	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "code.cloudfoundry.org/cli/util/testhelpers/io"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/wfernandes/app-metrics-plugin/cmd/plugin"
)

var _ = Describe("AppsMetrics", func() {

	It("requires the application name", func() {
		fakeCliConnection := &pluginfakes.FakeCliConnection{}
		appsMetricsPlugin := &AppsMetricsPlugin{}

		output := CaptureOutput(func() {
			appsMetricsPlugin.Run(fakeCliConnection, []string{"app-metrics"})
		})

		Expect(output).To(ContainElement("cf app-metrics APP_NAME"))
	})

	It("prints error for unknown app", func() {
		fakeCliConnection := &pluginfakes.FakeCliConnection{}
		fakeCliConnection.GetAppReturns(plugin_models.GetAppModel{}, fmt.Errorf("iDoNotExist does not exist"))
		appsMetricsPlugin := &AppsMetricsPlugin{}

		output := CaptureOutput(func() {
			appsMetricsPlugin.Run(fakeCliConnection, []string{"app-metrics", "iDoNotExist"})
		})

		Expect(output).To(ContainElement("iDoNotExist does not exist"))
	})

	It("prints error when unable to get metrics", func() {
		fakeCliConnection := &pluginfakes.FakeCliConnection{}
		// an app with no routes will trigger an error
		model := plugin_models.GetAppModel{
			Routes: []plugin_models.GetApp_RouteSummary{},
		}
		fakeCliConnection.GetAppReturns(model, nil)
		plugin := &AppsMetricsPlugin{}

		output := CaptureOutput(func() {
			plugin.Run(fakeCliConnection, []string{"app-metrics", "some-app"})
		})

		Expect(output).To(ContainElement("unable to get metrics: app does not have any routes to hit"))
	})

	It("prints error when unrecognized flag is set", func() {
		fakeCliConnection := &pluginfakes.FakeCliConnection{}
		model := plugin_models.GetAppModel{}
		fakeCliConnection.GetAppReturns(model, nil)
		plugin := &AppsMetricsPlugin{}

		output := CaptureOutput(func() {
			plugin.Run(fakeCliConnection, []string{"app-metrics", "some-app", "-unknownFlag"})
		})

		Expect(output).To(ContainElement("Invalid flag: -unknownFlag"))
	})

	It("prints uninstall message when uninstalling", func() {
		fakeCliConnection := &pluginfakes.FakeCliConnection{}
		plugin := &AppsMetricsPlugin{}

		output := CaptureOutput(func() {
			plugin.Run(fakeCliConnection, []string{"CLI-MESSAGE-UNINSTALL"})
		})

		Expect(output).To(ContainElement("Thank you for using app-metrics"))
	})

})

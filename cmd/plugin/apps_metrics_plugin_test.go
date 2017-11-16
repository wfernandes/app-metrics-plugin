package main_test

import (
	"fmt"

	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/cloudfoundry/cli/plugin/pluginfakes"
	. "github.com/cloudfoundry/cli/util/testhelpers/io"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/wfernandes/apps-metrics-plugin/cmd/plugin"
)

var _ = Describe("AppsMetricsPlugin", func() {

	It("requires the application name", func() {

		fakeCliConnection := &pluginfakes.FakeCliConnection{}
		appsMetricsPlugin := &AppsMetricsPlugin{}

		output := CaptureOutput(func() {
			appsMetricsPlugin.Run(fakeCliConnection, []string{"apps-metrics"})
		})
		Expect(output).To(ContainElement("APP_NAME is required"))
	})

	It("prints error for unknown app", func() {
		fakeCliConnection := &pluginfakes.FakeCliConnection{}
		fakeCliConnection.GetAppReturns(plugin_models.GetAppModel{}, fmt.Errorf("iDoNotExist does not exist"))
		appsMetricsPlugin := &AppsMetricsPlugin{}

		output := CaptureOutput(func() {
			appsMetricsPlugin.Run(fakeCliConnection, []string{"apps-metrics", "iDoNotExist"})
		})
		Expect(output).To(ContainElement("iDoNotExist does not exist"))
	})

})

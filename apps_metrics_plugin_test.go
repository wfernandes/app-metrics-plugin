package main_test

import (
	fakes "github.com/cloudfoundry/cli/plugin/pluginfakes"
	. "github.com/cloudfoundry/cli/util/testhelpers/io"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/wfernandes/apps-metrics-plugin"
)

var _ = Describe("AppsMetricsPlugin", func() {

	Context("Command Syntax", func() {
		It("requires the application name", func() {

			fakeCliConnection := &fakes.FakeCliConnection{}
			appsMetricsPlugin := &AppsMetricsPlugin{}

			output := CaptureOutput(func() {
				appsMetricsPlugin.Run(fakeCliConnection, []string{"apps-metrics"})
			})
			Expect(output).To(ContainElement("APP_NAME is required"))
		})
	})
})

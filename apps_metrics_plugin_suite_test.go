package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAppsMetricsPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AppsMetricsPlugin Suite")
}

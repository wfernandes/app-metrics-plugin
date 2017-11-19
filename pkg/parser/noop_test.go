package parser_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/wfernandes/apps-metrics-plugin/pkg/parser"
)

var _ = Describe("NoOp", func() {
	It("does not modify the bytes", func() {
		p := parser.NewNoOp()
		input := `{"abc": 123, "def": 456}`
		output, err := p.Parse([]byte(input))
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(MatchJSON(input))
	})
})

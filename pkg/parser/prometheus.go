package parser

import (
	"bytes"

	"github.com/prometheus/common/expfmt"
)

type Prometheus struct {
	parser expfmt.TextParser
}

func NewPrometheus() *Prometheus {
	return &Prometheus{
		parser: expfmt.TextParser{},
	}
}

func (p *Prometheus) Parse(b []byte) (map[string]interface{}, error) {

	m := make(map[string]interface{})
	buf := bytes.NewBuffer(b)
	metricFamilies, err := p.parser.TextToMetricFamilies(buf)
	if err != nil {
		return m, err
	}

	for k, v := range metricFamilies {
		m[k] = v
	}

	return m, nil
}

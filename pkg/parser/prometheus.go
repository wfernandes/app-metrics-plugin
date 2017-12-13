package parser

import (
	"bytes"
	"fmt"

	dto "github.com/prometheus/client_model/go"
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
		m[k] = newFamily(v)
	}

	return m, nil
}

// Copied below from https://github.com/prometheus/prom2json/blob/49f15d03cf3744b17ef104f648663e95e0486341/prom2json.go
// By default the dto.MetricFamily object cannot be marshalled into json because Prometheus allows sample values
// like NaN or +Inf, which cannot be encoded as JSON numbers
// I had to copy this code because prom2json.NewFamily() requires a dto.MetricFamily which is typed to its vendored directory.
// See error below
//
// cannot use v (type *"github.com/prometheus/client_model/go".MetricFamily) as type *"github.com/prometheus/prom2json/vendor/github.com/prometheus/client_model/go".MetricFamily in argument to prom2json.NewFamily

// Family mirrors the MetricFamily proto message.
type Family struct {
	//Time    time.Time
	Name    string        `json:"name"`
	Help    string        `json:"help"`
	Type    string        `json:"type"`
	Metrics []interface{} `json:"metrics,omitempty"` // Either metric or summary.
}

// Metric is for all "single value" metrics, i.e. Counter, Gauge, and Untyped.
type Metric struct {
	Labels map[string]string `json:"labels,omitempty"`
	Value  string            `json:"value"`
}

// Summary mirrors the Summary proto message.
type Summary struct {
	Labels    map[string]string `json:"labels,omitempty"`
	Quantiles map[string]string `json:"quantiles,omitempty"`
	Count     string            `json:"count"`
	Sum       string            `json:"sum"`
}

// Histogram mirrors the Histogram proto message.
type Histogram struct {
	Labels  map[string]string `json:"labels,omitempty"`
	Buckets map[string]string `json:"buckets,omitempty"`
	Count   string            `json:"count"`
	Sum     string            `json:"sum"`
}

// NewFamily consumes a MetricFamily and transforms it to the local Family type.
func newFamily(dtoMF *dto.MetricFamily) *Family {
	mf := &Family{
		//Time:    time.Now(),
		Name:    dtoMF.GetName(),
		Help:    dtoMF.GetHelp(),
		Type:    dtoMF.GetType().String(),
		Metrics: make([]interface{}, len(dtoMF.Metric)),
	}
	for i, m := range dtoMF.Metric {
		if dtoMF.GetType() == dto.MetricType_SUMMARY {
			mf.Metrics[i] = Summary{
				Labels:    makeLabels(m),
				Quantiles: makeQuantiles(m),
				Count:     fmt.Sprint(m.GetSummary().GetSampleCount()),
				Sum:       fmt.Sprint(m.GetSummary().GetSampleSum()),
			}
		} else if dtoMF.GetType() == dto.MetricType_HISTOGRAM {
			mf.Metrics[i] = Histogram{
				Labels:  makeLabels(m),
				Buckets: makeBuckets(m),
				Count:   fmt.Sprint(m.GetHistogram().GetSampleCount()),
				Sum:     fmt.Sprint(m.GetSummary().GetSampleSum()),
			}
		} else {
			mf.Metrics[i] = Metric{
				Labels: makeLabels(m),
				Value:  fmt.Sprint(getValue(m)),
			}
		}
	}
	return mf
}

func makeLabels(m *dto.Metric) map[string]string {
	result := map[string]string{}
	for _, lp := range m.Label {
		result[lp.GetName()] = lp.GetValue()
	}
	return result
}

func makeQuantiles(m *dto.Metric) map[string]string {
	result := map[string]string{}
	for _, q := range m.GetSummary().Quantile {
		result[fmt.Sprint(q.GetQuantile())] = fmt.Sprint(q.GetValue())
	}
	return result
}

func makeBuckets(m *dto.Metric) map[string]string {
	result := map[string]string{}
	for _, b := range m.GetHistogram().Bucket {
		result[fmt.Sprint(b.GetUpperBound())] = fmt.Sprint(b.GetCumulativeCount())
	}
	return result
}

func getValue(m *dto.Metric) float64 {
	if m.Gauge != nil {
		return m.GetGauge().GetValue()
	}
	if m.Counter != nil {
		return m.GetCounter().GetValue()
	}
	if m.Untyped != nil {
		return m.GetUntyped().GetValue()
	}
	return 0.
}

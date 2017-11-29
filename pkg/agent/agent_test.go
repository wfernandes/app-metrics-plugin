package agent_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/wfernandes/app-metrics-plugin/pkg/agent"
	"github.com/wfernandes/app-metrics-plugin/pkg/parser"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Agent", func() {

	It("makes request using domain name only if there is no host", func() {
		fakeApp := &plugin_models.GetAppModel{
			RunningInstances: 1,
			Instances: []plugin_models.GetApp_AppInstanceFields{
				{
					State: "running",
				},
			},
			Routes: []plugin_models.GetApp_RouteSummary{
				{
					Domain: plugin_models.GetApp_DomainFields{
						Name: "domain.cf-app.com",
					},
				},
			},
		}

		fakeClient := NewFakeClient()
		a := agent.New(fakeApp, NewFakeParser(), agent.WithClient(fakeClient))
		_, err := a.GetMetrics(context.Background())
		Expect(err).ToNot(HaveOccurred())
		Expect(fakeClient.LastRequest().URL.String()).To(Equal("http://domain.cf-app.com/debug/metrics"))
	})

	It("makes request using host and domain name", func() {
		fakeApp := &plugin_models.GetAppModel{
			RunningInstances: 1,
			Instances: []plugin_models.GetApp_AppInstanceFields{
				{
					State: "running",
				},
			},
			Routes: []plugin_models.GetApp_RouteSummary{
				{
					Host: "my-app-host",
					Domain: plugin_models.GetApp_DomainFields{
						Name: "domain.cf-app.com",
					},
				},
			},
		}
		fakeClient := NewFakeClient()

		a := agent.New(fakeApp, NewFakeParser(), agent.WithClient(fakeClient))
		_, err := a.GetMetrics(context.Background())

		Expect(err).ToNot(HaveOccurred())
		Expect(fakeClient.LastRequest().URL.String()).To(Equal("http://my-app-host.domain.cf-app.com/debug/metrics"))
	})

	It("returns metric output upon successful request", func() {
		fakeClient := NewFakeClient()
		fakeClient.SetResponse(expvarJSON)

		fakeApp := &plugin_models.GetAppModel{
			RunningInstances: 1,
			Instances: []plugin_models.GetApp_AppInstanceFields{
				{
					State: "running",
				},
			},
			Routes: []plugin_models.GetApp_RouteSummary{
				{
					Domain: plugin_models.GetApp_DomainFields{
						Name: "domain.cf-app.com",
					},
				},
			},
		}

		a := agent.New(fakeApp, parser.NewExpvar(), agent.WithClient(fakeClient))
		metrics, err := a.GetMetrics(context.Background())

		request := fakeClient.LastRequest()
		Expect(err).ToNot(HaveOccurred())
		Expect(metrics).To(HaveLen(1))
		metric := metrics[0]
		Expect(metric.Instance).To(Equal(0))
		Expect(metric.Error).To(BeEmpty())
		Expect(metric.Metrics).To(HaveLen(5))
		Expect(metric.Metrics).To(HaveKeyWithValue("metric.float", 123.345))
		Expect(metric.Metrics).To(HaveKeyWithValue("metric.string", "expvarApp"))
		Expect(metric.Metrics).To(HaveKeyWithValue("metric.int", float64(10)))
		Expect(metric.Metrics["cmdline"]).To(ContainElement("bin/event-alerts"))
		Expect(metric.Metrics["metric.map"]).To(HaveKeyWithValue("metric1", float64(10)))
		Expect(metric.Metrics["metric.map"]).To(HaveKeyWithValue("metric2", float64(11)))
		Expect(request.Header.Get("X-CF-APP-INSTANCE")).ToNot(BeEmpty())
	})

	It("returns error upon unsuccessful request", func() {
		fakeApp := &plugin_models.GetAppModel{
			RunningInstances: 1,
			Instances: []plugin_models.GetApp_AppInstanceFields{
				{
					State: "running",
				},
			},
			Routes: []plugin_models.GetApp_RouteSummary{
				{
					Domain: plugin_models.GetApp_DomainFields{
						Name: "http://192.168.0.%31/", // To force an error in http.NewRequest
					},
				},
			},
		}

		a := agent.New(fakeApp, NewFakeParser())
		instanceMetrics, err := a.GetMetrics(context.Background())

		Expect(err).ToNot(HaveOccurred())
		Expect(instanceMetrics).To(HaveLen(1))
		Expect(instanceMetrics[0].Error).ToNot(BeEmpty())
	})

	It("returns error if app has no routes", func() {
		fakeApp := &plugin_models.GetAppModel{
			RunningInstances: 1,
			Instances: []plugin_models.GetApp_AppInstanceFields{
				{
					State: "running",
				},
			},
			Routes: []plugin_models.GetApp_RouteSummary{},
		}

		a := agent.New(fakeApp, NewFakeParser())
		_, err := a.GetMetrics(context.Background())

		Expect(err).To(HaveOccurred())
	})

	It("returns output error upon failing request", func() {
		fakeClient := NewFakeClient()
		fakeClient.SetError(errors.New("some request error"))

		fakeApp := &plugin_models.GetAppModel{
			RunningInstances: 1,
			Instances: []plugin_models.GetApp_AppInstanceFields{
				{
					State: "running",
				},
			},
			Routes: []plugin_models.GetApp_RouteSummary{
				{
					Domain: plugin_models.GetApp_DomainFields{
						Name: "domain.cf-app.com",
					},
				},
			},
		}

		a := agent.New(fakeApp, NewFakeParser(), agent.WithClient(fakeClient))
		output, err := a.GetMetrics(context.Background())

		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(HaveLen(1))
		Expect(output[0].Instance).To(Equal(0))
		Expect(output[0].Metrics).To(BeEmpty())
		Expect(output[0].Error).To(Equal("some request error"))
	})

	It("returns output error upon failing to read response body", func() {
		fakeClient := NewFakeClient()
		fakeClient.SetBadResponse()
		fakeApp := &plugin_models.GetAppModel{
			RunningInstances: 1,
			Instances: []plugin_models.GetApp_AppInstanceFields{
				{
					State: "running",
				},
			},
			Routes: []plugin_models.GetApp_RouteSummary{
				{
					Domain: plugin_models.GetApp_DomainFields{
						Name: "domain.cf-app.com",
					},
				},
			},
		}

		a := agent.New(fakeApp, NewFakeParser(), agent.WithClient(fakeClient))
		output, err := a.GetMetrics(context.Background())

		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(HaveLen(1))
		Expect(output[0].Instance).To(Equal(0))
		Expect(output[0].Metrics).To(BeEmpty())
		Expect(output[0].Error).To(Equal(io.ErrUnexpectedEOF.Error()))
	})

	It("returns output error when failing to parse metric json", func() {
		fakeClient := NewFakeClient()
		fakeClient.SetResponse("404 page not found\n")

		fakeApp := &plugin_models.GetAppModel{
			RunningInstances: 1,
			Instances: []plugin_models.GetApp_AppInstanceFields{
				{
					State: "running",
				},
			},
			Routes: []plugin_models.GetApp_RouteSummary{
				{
					Domain: plugin_models.GetApp_DomainFields{
						Name: "domain.cf-app.com",
					},
				},
			},
		}

		a := agent.New(fakeApp, parser.NewExpvar(), agent.WithClient(fakeClient))
		instanceMetrics, err := a.GetMetrics(context.Background())

		Expect(err).ToNot(HaveOccurred())
		Expect(instanceMetrics).To(HaveLen(1))
		Expect(instanceMetrics[0].Error).ToNot(BeEmpty())
	})

	It("sends GET request with X-CF-APP-INSTANCE header for app with multiple instances", func() {
		fakeClient := NewFakeClient()
		fakeApp := &plugin_models.GetAppModel{
			Guid:             "some-app-guid",
			RunningInstances: 2,
			Instances: []plugin_models.GetApp_AppInstanceFields{
				{
					State: "running",
				},
				{
					State: "stopped",
				},
				{
					State: "running",
				},
			},
			Routes: []plugin_models.GetApp_RouteSummary{
				{
					Domain: plugin_models.GetApp_DomainFields{
						Name: "domain.cf-app.com",
					},
				},
			},
		}

		a := agent.New(fakeApp, NewFakeParser(), agent.WithClient(fakeClient))
		_, err := a.GetMetrics(context.Background())

		Expect(err).ToNot(HaveOccurred())
		Eventually(fakeClient.Requests).Should(HaveLen(2))
		var headers []string
		for _, r := range fakeClient.Requests() {
			headers = append(headers, r.Header.Get("X-CF-APP-INSTANCE"))
		}
		Expect(headers).To(ConsistOf("some-app-guid:0", "some-app-guid:2"))
	})
})

var expvarJSON = `{
"cmdline": ["bin/event-alerts"],
"metric.float": 123.345,
"metric.int": 10,
"metric.map": {"metric1": 10, "metric2": 11},
"metric.string": "expvarApp"
}`

type FakeParser struct {
	mu              sync.Mutex
	parseCalledWith []byte
	err             error
}

func NewFakeParser() *FakeParser {
	return &FakeParser{
		// making this size big because `copy` copies the min of len(dst) and len(src)
		parseCalledWith: make([]byte, 1024),
	}
}

func (p *FakeParser) SetError(e error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.err = e
}

func (p *FakeParser) Parse(b []byte) (map[string]interface{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.parseCalledWith = b
	return nil, p.err
}

func (p *FakeParser) ParseCalledWith() []byte {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.parseCalledWith
}

type FakeClient struct {
	mu         sync.Mutex
	requests   []*http.Request
	body       string
	err        error
	readerFail bool
}

func NewFakeClient() *FakeClient {
	return &FakeClient{
		requests: make([]*http.Request, 0),
		body:     "some default response",
	}
}

func (f *FakeClient) Do(r *http.Request) (*http.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.requests = append(f.requests, r)
	var resp *http.Response
	if f.readerFail {
		resp = &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(&FakeReader{}),
		}
	} else {
		resp = &http.Response{
			Body: ioutil.NopCloser(bytes.NewBufferString(f.body)),
		}
	}

	return resp, f.err
}

func (f *FakeClient) Requests() []*http.Request {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.requests
}

func (f *FakeClient) LastRequest() *http.Request {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.requests[len(f.requests)-1]
}

func (f *FakeClient) SetError(e error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.err = e
}

func (f *FakeClient) SetResponse(body string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.body = body
}

func (f *FakeClient) SetBadResponse() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.readerFail = true

}

type FakeReader struct{}

func (frc *FakeReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

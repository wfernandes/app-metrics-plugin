package agent_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"

	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/wfernandes/apps-metrics-plugin/pkg/agent"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Agent", func() {

	var (
		mux *http.ServeMux
		ts  *httptest.Server
	)

	It("returns error upon unsuccessful request", func() {
		fakeApp := &plugin_models.GetAppModel{
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
		a := agent.New(fakeApp, "some-token")
		output, err := a.GetMetrics()
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(HaveLen(1))
		Expect(output[0].Error).ToNot(BeEmpty())
	})

	It("returns error if app has no routes", func() {
		fakeApp := &plugin_models.GetAppModel{
			Instances: []plugin_models.GetApp_AppInstanceFields{
				{
					State: "running",
				},
			},
			Routes: []plugin_models.GetApp_RouteSummary{},
		}
		a := agent.New(fakeApp, "some-token")
		_, err := a.GetMetrics()
		Expect(err).To(HaveOccurred())
	})

	It("makes request using domain name only if there is no host", func() {
		fakeApp := &plugin_models.GetAppModel{
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
		a := agent.New(fakeApp, "some-token", agent.WithClient(fakeClient))
		_, err := a.GetMetrics()
		Expect(err).ToNot(HaveOccurred())
		Expect(fakeClient.LastRequest().URL.String()).To(Equal("domain.cf-app.com/debug/metrics"))
	})

	It("makes request using host and domain name", func() {
		fakeApp := &plugin_models.GetAppModel{
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
		a := agent.New(fakeApp, "some-token", agent.WithClient(fakeClient))
		_, err := a.GetMetrics()
		Expect(err).ToNot(HaveOccurred())
		Expect(fakeClient.LastRequest().URL.String()).To(Equal("my-app-host.domain.cf-app.com/debug/metrics"))
	})

	It("returns response body upon successful request", func() {
		fakeClient := NewFakeClient()
		fakeClient.SetResponse("some response body")

		fakeApp := &plugin_models.GetAppModel{
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

		a := agent.New(fakeApp, "some-token", agent.WithClient(fakeClient))
		output, err := a.GetMetrics()

		request := fakeClient.LastRequest()
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(HaveLen(1))
		Expect(output[0].Instance).To(Equal(0))
		Expect(output[0].Output).To(Equal("some response body"))
		Expect(output[0].Error).To(BeEmpty())
		Expect(request.Header.Get("X-CF-APP-INSTANCE")).ToNot(BeEmpty())
		Expect(request.Header.Get("Authorization")).To(Equal("some-token"))
	})

	It("returns output error upon failing request", func() {
		fakeClient := NewFakeClient()
		fakeClient.SetError(errors.New("some request error"))

		fakeApp := &plugin_models.GetAppModel{
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

		a := agent.New(fakeApp, "some-token", agent.WithClient(fakeClient))
		output, err := a.GetMetrics()

		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(HaveLen(1))
		Expect(output[0].Instance).To(Equal(0))
		Expect(output[0].Output).To(BeEmpty())
		Expect(output[0].Error).To(Equal("some request error"))
	})

	It("returns output error upon failing to read response body", func() {
		fakeClient := NewFakeClient()
		fakeClient.SetBadResponse()
		fakeApp := &plugin_models.GetAppModel{
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

		a := agent.New(fakeApp, "some-token", agent.WithClient(fakeClient))
		output, err := a.GetMetrics()

		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(HaveLen(1))
		Expect(output[0].Instance).To(Equal(0))
		Expect(output[0].Output).To(BeEmpty())
		Expect(output[0].Error).To(Equal(io.ErrUnexpectedEOF.Error()))
	})

	It("sends GET request with X-CF-APP-INSTANCE header for app with multiple instances", func() {
		mux = http.NewServeMux()
		ts = httptest.NewServer(mux)
		defer ts.Close()
		requests := make([]*http.Request, 0, 3)
		mux.HandleFunc("/debug/metrics", func(w http.ResponseWriter, r *http.Request) {
			requests = append(requests, r)
			fmt.Fprintf(w, "request %d", len(requests))
		})
		fakeApp := &plugin_models.GetAppModel{
			Guid: "some-app-guid",
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
						Name: ts.URL,
					},
				},
			},
		}
		a := agent.New(fakeApp, "some-token")

		_, err := a.GetMetrics()

		Expect(err).ToNot(HaveOccurred())
		Eventually(requests).Should(HaveLen(2))
		var headers []string
		for _, r := range requests {
			headers = append(headers, r.Header.Get("X-CF-APP-INSTANCE"))
		}
		Expect(headers).To(ConsistOf("some-app-guid:0", "some-app-guid:2"))
	})

})

type FakeClient struct {
	mu      sync.Mutex
	request *http.Request
	body    io.ReadCloser
	err     error
}

func NewFakeClient() *FakeClient {
	return &FakeClient{
		body: ioutil.NopCloser(bytes.NewBufferString("some default response")),
	}
}

func (f *FakeClient) Do(r *http.Request) (*http.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.request = r
	return &http.Response{
		Body: f.body,
	}, f.err
}

func (f *FakeClient) LastRequest() *http.Request {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.request
}

func (f *FakeClient) SetError(e error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.err = e
}

func (f *FakeClient) SetResponse(body string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.body = ioutil.NopCloser(bytes.NewBufferString(body))
}

func (f *FakeClient) SetBadResponse() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.body = ioutil.NopCloser(&FakeReader{})

}

type FakeReader struct{}

func (frc *FakeReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

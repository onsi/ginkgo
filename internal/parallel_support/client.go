package parallel_support

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/onsi/ginkgo/types"
)

var ErrorGone = fmt.Errorf("gone!")
var ErrorFailed = fmt.Errorf("failed!")

var POLLING_INTERVAL = 50 * time.Millisecond

type Client struct {
	serverHost string
}

func NewClient(serverHost string) Client {
	return Client{
		serverHost: serverHost,
	}
}

func (client Client) post(path string, data interface{}) error {
	var body io.Reader
	if data != nil {
		encoded, err := json.Marshal(data)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(encoded)
	}
	resp, err := http.Post(client.serverHost+path, "application/json", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received unexpected status code %d", resp.StatusCode)
	}
	return nil
}

func (client Client) poll(path string, data interface{}) error {
	for {
		resp, err := http.Get(client.serverHost + path)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusTooEarly {
			resp.Body.Close()
			time.Sleep(POLLING_INTERVAL)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusGone {
			return ErrorGone
		}
		if resp.StatusCode == http.StatusFailedDependency {
			return ErrorFailed
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("received unexpected status code %d", resp.StatusCode)
		}
		if data != nil {
			return json.NewDecoder(resp.Body).Decode(data)
		}
		return nil
	}
}

func (client Client) IsZero() bool {
	return client.serverHost == ""
}

func (client Client) PostSuiteWillBegin(report types.Report) error {
	return client.post("/suite-will-begin", report)
}

func (client Client) PostDidRun(report types.SpecReport) error {
	return client.post("/did-run", report)
}

func (client Client) PostSuiteDidEnd(report types.Report) error {
	return client.post("/suite-did-end", report)
}

func (client Client) PostSynchronizedBeforeSuiteSucceeded(data []byte) error {
	return client.post("/before-suite-succeeded", data)
}

func (client Client) PostSynchronizedBeforeSuiteFailed() error {
	return client.post("/before-suite-failed", nil)
}

func (client Client) BlockUntilSynchronizedBeforeSuiteData() ([]byte, error) {
	var data []byte
	err := client.poll("/before-suite-state", &data)
	if err == ErrorGone {
		return nil, types.GinkgoErrors.SynchronizedBeforeSuiteDisappearedOnNode1()
	} else if err == ErrorFailed {
		return nil, types.GinkgoErrors.SynchronizedBeforeSuiteFailedOnNode1()
	}
	return data, err
}

func (client Client) BlockUntilNonprimaryNodesHaveFinished() error {
	return client.poll("/have-nonprimary-nodes-finished", nil)
}

func (client Client) BlockUntilAggregatedNonprimaryNodesReport() (types.Report, error) {
	var report types.Report
	err := client.poll("/aggregated-nonprimary-nodes-report", &report)
	if err == ErrorGone {
		return types.Report{}, types.GinkgoErrors.AggregatedReportUnavailableDueToNodeDisappearing()
	}
	return report, err
}

func (client Client) FetchNextCounter() (int, error) {
	var counter ParallelIndexCounter
	err := client.poll("/counter", &counter)
	return counter.Index, err
}

func (client Client) CheckServerUp() bool {
	resp, err := http.Get(client.serverHost + "/up")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (client Client) PostAbort() error {
	return client.post("/abort", nil)
}

func (client Client) ShouldAbort() bool {
	err := client.poll("/abort", nil)
	if err == ErrorGone {
		return true
	}
	return false
}

func (client Client) Write(p []byte) (int, error) {
	resp, err := http.Post(client.serverHost+"/stream-output", "text/plain;charset=UTF-8 ", bytes.NewReader(p))
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to stream output")
	}
	return len(p), err
}

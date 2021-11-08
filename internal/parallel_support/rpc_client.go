package parallel_support

import (
	"net/rpc"
	"time"

	"github.com/onsi/ginkgo/types"
)

// TODO:
// - get RPC working
// - performance test
// - add DeferCleanup to test helper

type rpcClient struct {
	serverHost string
	client     *rpc.Client
}

func newRPCClient(serverHost string) *rpcClient {
	return &rpcClient{
		serverHost: serverHost,
	}
}

func (client *rpcClient) Connect() bool {
	var err error
	if client.client != nil {
		return true
	}
	client.client, err = rpc.DialHTTPPath("tcp", client.serverHost, "/")
	if err != nil {
		client.client = nil
		return false
	}
	return true
}

func (client *rpcClient) Close() error {
	return client.client.Close()
}

func (client *rpcClient) poll(method string, data interface{}) error {
	for {
		err := client.client.Call(method, voidSender, data)
		if err == nil {
			return nil
		}
		switch err.Error() {
		case ErrorEarly.Error():
			time.Sleep(POLLING_INTERVAL)
		case ErrorGone.Error():
			return ErrorGone
		case ErrorFailed.Error():
			return ErrorFailed
		default:
			return err
		}
	}
}

func (client *rpcClient) PostSuiteWillBegin(report types.Report) error {
	return client.client.Call("Server.SpecSuiteWillBegin", report, voidReceiver)
}

func (client *rpcClient) PostDidRun(report types.SpecReport) error {
	return client.client.Call("Server.DidRun", report, voidReceiver)
}

func (client *rpcClient) PostSuiteDidEnd(report types.Report) error {
	return client.client.Call("Server.SpecSuiteDidEnd", report, voidReceiver)
}

func (client *rpcClient) Write(p []byte) (int, error) {
	var n int
	err := client.client.Call("Server.EmitOutput", p, &n)
	return n, err
}

func (client *rpcClient) PostSynchronizedBeforeSuiteSucceeded(data []byte) error {
	return client.client.Call("Server.BeforeSuiteSucceeded", data, voidReceiver)
}

func (client *rpcClient) PostSynchronizedBeforeSuiteFailed() error {
	return client.client.Call("Server.BeforeSuiteFailed", voidSender, voidReceiver)
}

func (client *rpcClient) BlockUntilSynchronizedBeforeSuiteData() ([]byte, error) {
	var data []byte
	err := client.poll("Server.BeforeSuiteState", &data)
	if err == ErrorGone {
		return nil, types.GinkgoErrors.SynchronizedBeforeSuiteDisappearedOnProc1()
	} else if err == ErrorFailed {
		return nil, types.GinkgoErrors.SynchronizedBeforeSuiteFailedOnProc1()
	}
	return data, err
}

func (client *rpcClient) BlockUntilNonprimaryNodesHaveFinished() error {
	return client.poll("Server.HaveNonprimaryNodesFinished", voidReceiver)
}

func (client *rpcClient) BlockUntilAggregatedNonprimaryNodesReport() (types.Report, error) {
	var report types.Report
	err := client.poll("Server.AggregatedNonprimaryNodesReport", &report)
	if err == ErrorGone {
		return types.Report{}, types.GinkgoErrors.AggregatedReportUnavailableDueToNodeDisappearing()
	}
	return report, err
}

func (client *rpcClient) FetchNextCounter() (int, error) {
	var counter int
	err := client.client.Call("Server.Counter", voidSender, &counter)
	return counter, err
}

func (client *rpcClient) PostAbort() error {
	return client.client.Call("Server.Abort", voidSender, voidReceiver)
}

func (client *rpcClient) ShouldAbort() bool {
	var shouldAbort bool
	client.client.Call("Server.ShouldAbort", voidSender, &shouldAbort)
	return shouldAbort
}

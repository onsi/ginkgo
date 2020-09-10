package parallel_support_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestParallelSupport(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Parallel Support Suite")
}

type post struct {
	url         string
	bodyType    string
	bodyContent []byte
}

type fakePoster struct {
	posts []post
}

func newFakePoster() *fakePoster {
	return &fakePoster{
		posts: make([]post, 0),
	}
}

func (poster *fakePoster) Post(url string, bodyType string, body io.Reader) (resp *http.Response, err error) {
	bodyContent, _ := ioutil.ReadAll(body)
	poster.posts = append(poster.posts, post{
		url:         url,
		bodyType:    bodyType,
		bodyContent: bodyContent,
	})
	return nil, nil
}

type fakeOutputInterceptor struct {
	DidStartInterceptingOutput bool
	DidStopInterceptingOutput  bool
	InterceptedOutput          string
}

func (interceptor *fakeOutputInterceptor) StartInterceptingOutput() error {
	interceptor.DidStartInterceptingOutput = true
	return nil
}

func (interceptor *fakeOutputInterceptor) StopInterceptingAndReturnOutput() (string, error) {
	interceptor.DidStopInterceptingOutput = true
	return interceptor.InterceptedOutput, nil
}

func (interceptor *fakeOutputInterceptor) StreamTo(*os.File) {
}

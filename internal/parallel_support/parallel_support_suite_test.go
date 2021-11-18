package parallel_support_test

import (
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
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
	bodyContent, _ := io.ReadAll(body)
	poster.posts = append(poster.posts, post{
		url:         url,
		bodyType:    bodyType,
		bodyContent: bodyContent,
	})
	return nil, nil
}

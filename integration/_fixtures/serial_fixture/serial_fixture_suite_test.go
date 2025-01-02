package serial_fixture_test

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var noSerial *bool

func init() {
	noSerial = flag.CommandLine.Bool("no-serial", false, "set to turn off serial decoration")
}

var SerialDecoration = []any{Serial}

func TestSerialFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	if *noSerial {
		SerialDecoration = []any{}
	}

	RunSpecs(t, "SerialFixture Suite")
}

var addr string

var _ = SynchronizedBeforeSuite(func() []byte {
	server := ghttp.NewServer()
	lock := &sync.Mutex{}
	count := 0
	server.RouteToHandler("POST", "/counter", func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		count += 1
		lock.Unlock()
		w.WriteHeader(http.StatusOK)
	})
	server.RouteToHandler("GET", "/counter", func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		data := []byte(fmt.Sprintf("%d", count))
		lock.Unlock()
		w.Write(data)
	})
	return []byte(server.HTTPTestServer.URL)
}, func(data []byte) {
	addr = string(data) + "/counter"
})

var postToCounter = func(g Gomega) {
	req, err := http.Post(addr, "", nil)
	g.Ω(err).ShouldNot(HaveOccurred())
	g.Ω(req.StatusCode).Should(Equal(http.StatusOK))
	g.Ω(req.Body.Close()).Should(Succeed())
}

var getCounter = func(g Gomega) string {
	req, err := http.Get(addr)
	g.Ω(err).ShouldNot(HaveOccurred())
	content, err := io.ReadAll(req.Body)
	g.Ω(err).ShouldNot(HaveOccurred())
	g.Ω(req.Body.Close()).Should(Succeed())
	return string(content)
}

var _ = SynchronizedAfterSuite(func() {
	Consistently(postToCounter, 200*time.Millisecond, 10*time.Millisecond).Should(Succeed())
}, func() {
	initialValue := getCounter(Default)
	Consistently(func(g Gomega) {
		currentValue := getCounter(g)
		g.Ω(currentValue).Should(Equal(initialValue))
	}, 100*time.Millisecond, 10*time.Millisecond).Should(Succeed())
})

var _ = Describe("tests", func() {
	for i := 0; i < 10; i += 1 {
		It("runs in parallel", func() {
			Consistently(postToCounter, 200*time.Millisecond, 10*time.Millisecond).Should(Succeed())
		})
	}

	for i := 0; i < 5; i += 1 {
		It("runs in series", SerialDecoration, func() {
			Ω(GinkgoParallelProcess()).Should(Equal(1))
			initialValue := getCounter(Default)
			Consistently(func(g Gomega) {
				currentValue := getCounter(g)
				g.Ω(currentValue).Should(Equal(initialValue))
			}, 100*time.Millisecond, 10*time.Millisecond).Should(Succeed())
		})
	}
})

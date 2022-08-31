package progress_reporter_fixture_test

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("ProgressReporter", func() {
	It("can track on demand", func() {
		By("Step A")
		By("Step B")
		fmt.Printf("READY %d\n", os.Getpid())
		time.Sleep(time.Second)
	})

	It("--poll-progress-after tracks things that take too long", Label("parallel"), func() {
		time.Sleep(2 * time.Second)
	})

	It("decorator tracks things that take too long", Label("parallel"), func() {
		time.Sleep(1 * time.Second)
	}, PollProgressAfter(500*time.Millisecond))
})

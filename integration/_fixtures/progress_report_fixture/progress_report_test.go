package progress_report_fixture_test

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("ProgressReport", func() {
	BeforeEach(func() {
		DeferCleanup(AttachProgressReporter(func() string {
			return fmt.Sprintf("Some global information: %d", GinkgoParallelProcess())
		}))
	})
	It("can track on demand", func() {
		By("Step A")
		By("Step B")
		for i := 1; i <= 12; i++ {
			GinkgoWriter.Printf("ginkgo-writer-output-%d\n", i)
		}
		fmt.Printf("READY %d\n", os.Getpid())
		time.Sleep(time.Second)
	})

	It("--poll-progress-after tracks things that take too long", Label("parallel"), func() {
		time.Sleep(2 * time.Second)
	})

	It("decorator tracks things that take too long", Label("parallel", "one-second"), func() {
		time.Sleep(1 * time.Second)
	}, PollProgressAfter(500*time.Millisecond))
})

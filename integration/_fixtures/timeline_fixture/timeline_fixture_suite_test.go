package timeline_fixture_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

func TestTimelineFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TimelineFixture Suite")
}

var _ = Describe("a full timeline", Serial, func() {
	Describe("a flaky test", func() {
		BeforeEach(func() {
			By("logging some events")
			GinkgoWriter.Println("hello!")
			AddReportEntry("a report!", "Of {{bold}}great{{/}} value")
			DeferCleanup(func() {
				By("cleaning up a bit", func() {
					time.Sleep(time.Millisecond * 50)
					GinkgoWriter.Println("all done!")
				})
			})
		})

		i := 0

		It("retries a few times", func() {
			i += 1
			GinkgoWriter.Println("let's try...")
			if i < 3 {
				Fail("bam!")
			}
			GinkgoWriter.Println("hooray!")
		}, FlakeAttempts(3))

		AfterEach(func() {
			if i == 3 {
				GinkgoWriter.Println("feeling sleepy...")
				time.Sleep(time.Millisecond * 200)
			}
		}, PollProgressAfter(time.Millisecond*100))
	})

	Describe("a test with multiple failures", func() {
		It("times out", func(ctx SpecContext) {
			By("waiting...")
			<-ctx.Done()
			GinkgoWriter.Println("then failing!")
			Fail("welp")
		}, NodeTimeout(time.Millisecond*100))

		AfterEach(func() {
			panic("aaah!")
		})
	})

	It("passes happily", func() {
		AddReportEntry("a verbose-only report", types.ReportEntryVisibilityFailureOrVerbose)
		AddReportEntry("a hidden report", types.ReportEntryVisibilityNever)
	})
})

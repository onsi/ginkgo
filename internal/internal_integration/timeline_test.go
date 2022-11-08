package internal_integration_test

import (
	"sort"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generating Timelines", func() {
	BeforeEach(func() {
		cl = types.NewCodeLocation(0)
		success, _ := RunFixture("current test description", func() {
			Context("Ordering", func() {
				Describe("a timeout", func() {
					BeforeEach(func() {
						writer.Print("bef\n")
						time.Sleep(time.Millisecond * 200)
					}, PollProgressAfter(time.Millisecond*100))
					It("spec", func(ctx SpecContext) {
						writer.Print("it\n")
						By("waiting", func() {
							<-ctx.Done()
						})
						AddReportEntry("report-entry")
						F("welp", types.NewCodeLocation(0))
					}, NodeTimeout(time.Millisecond*100))
					AfterEach(func() {
						writer.Print("aft\n")
						By("trying to clean up")
						panic("bam")
					})
					AfterEach(func() { F("boom", types.NewCodeLocation(0)) })
				})
				Describe("a flake", func() {
					i := 0
					It("flakes repeatedly", func() {
						writer.Print("running\n")
						if i < 2 {
							i += 1
							F("flake", types.NewCodeLocation(0))
						}
					}, FlakeAttempts(3))
				})
			})
		})

		Ω(success).Should(BeFalse())
	})

	It("runs the specs", func() {
		Ω(reporter.Did.Find("spec")).Should(HaveTimedOut("A node timeout occurred", clLine(8), CapturedGinkgoWriterOutput("bef\nit\naft\n")))
		Ω(reporter.Did.Find("flakes repeatedly")).Should(HavePassed(clLine(8), 3, CapturedGinkgoWriterOutput("running\nrunning\nrunning\n")))
	})

	It("generates a correctly sorted timeline for the timedout spec", func() {
		timeline := reporter.Did.Find("spec").Timeline()
		sort.Sort(timeline)
		Ω(timeline).Should(BeTimelineExactlyMatching(
			BeSpecEvent(types.SpecEventNodeStart, types.NodeTypeBeforeEach, TLWithOffset(0), clLine(4), "a timeout"),
			BeProgressReport("Automatically polling progress:", TLWithOffset("bef\n"), clLine(4), types.NodeTypeBeforeEach),
			BeSpecEvent(types.SpecEventNodeEnd, types.NodeTypeBeforeEach, TLWithOffset("bef\n"), clLine(4), time.Millisecond*200),

			BeSpecEvent(types.SpecEventNodeStart, types.NodeTypeIt, TLWithOffset("bef\n"), clLine(8), "spec"),
			BeSpecEvent(types.SpecEventByStart, TLWithOffset("bef\nit\n"), clLine(10), "waiting"),
			And(
				HaveField("Message", "A node timeout occurred"),
				HaveField("TimelineLocation.Offset", TLWithOffset("bef\nit\n").Offset),
			),
			BeSpecEvent(types.SpecEventByEnd, TLWithOffset("bef\nit\n"), clLine(10), "waiting", time.Millisecond*100),
			BeReportEntry("report-entry", ReportEntryVisibilityAlways, TLWithOffset("bef\nit\n"), clLine(13)),
			HaveFailed("A node timeout occurred and then the following failure was recorded in the timedout node before it exited:\nwelp", clLine(14), TLWithOffset("bef\nit\n")),
			BeSpecEvent(types.SpecEventNodeEnd, types.NodeTypeIt, TLWithOffset("bef\nit\n"), clLine(8), "spec", time.Millisecond*100),

			BeSpecEvent(types.SpecEventNodeStart, types.NodeTypeAfterEach, TLWithOffset("bef\nit\n"), clLine(16), "a timeout"),
			BeSpecEvent(types.SpecEventByStart, TLWithOffset("bef\nit\naft\n"), clLine(18), "trying to clean up"),
			HavePanicked("bam", TLWithOffset("bef\nit\naft\n"), FailureNodeType(types.NodeTypeAfterEach)),
			BeSpecEvent(types.SpecEventNodeEnd, types.NodeTypeAfterEach, TLWithOffset("bef\nit\naft\n"), clLine(16), "a timeout"),

			BeSpecEvent(types.SpecEventNodeStart, types.NodeTypeAfterEach, TLWithOffset("bef\nit\naft\n"), clLine(21), "a timeout"),
			HaveFailed("boom", TLWithOffset("bef\nit\naft\n"), FailureNodeType(types.NodeTypeAfterEach), clLine(21)),
			BeSpecEvent(types.SpecEventNodeEnd, types.NodeTypeAfterEach, TLWithOffset("bef\nit\naft\n"), clLine(21), "a timeout"),
		))
	})

	It("generates a correctly sorted timeline for the flakey spec", func() {
		timeline := reporter.Did.Find("flakes repeatedly").Timeline()
		sort.Sort(timeline)
		Ω(timeline).Should(BeTimelineExactlyMatching(
			BeSpecEvent(types.SpecEventNodeStart, types.NodeTypeIt, TLWithOffset(0), clLine(25), "flakes repeatedly"),
			HaveFailed("Failure recorded during attempt 1:\nflake", TLWithOffset("running\n"), clLine(29)),
			BeSpecEvent(types.SpecEventNodeEnd, types.NodeTypeIt, TLWithOffset("running\n"), clLine(25), "flakes repeatedly"),

			BeSpecEvent(types.SpecEventSpecRetry, TLWithOffset("running\n"), 1),

			BeSpecEvent(types.SpecEventNodeStart, types.NodeTypeIt, TLWithOffset("running\n"), clLine(25), "flakes repeatedly"),
			HaveFailed("Failure recorded during attempt 2:\nflake", TLWithOffset("running\nrunning\n"), clLine(29)),
			BeSpecEvent(types.SpecEventNodeEnd, types.NodeTypeIt, TLWithOffset("running\nrunning\n"), clLine(25), "flakes repeatedly"),

			BeSpecEvent(types.SpecEventSpecRetry, TLWithOffset("running\nrunning\n"), 2),

			BeSpecEvent(types.SpecEventNodeStart, types.NodeTypeIt, TLWithOffset("running\nrunning\n"), clLine(25), "flakes repeatedly"),
			BeSpecEvent(types.SpecEventNodeEnd, types.NodeTypeIt, TLWithOffset("running\nrunning\nrunning\n"), clLine(25), "flakes repeatedly"),
		))
	})
	It("emits all these timeline events along the way", func() {
		t := types.Timeline{}
		for _, failure := range reporter.Failures {
			t = append(t, failure)
		}
		Ω(t).Should(BeTimelineExactlyMatching(
			HaveTimedOut("A node timeout occurred", TLWithOffset("bef\nit\n")),
			HaveFailed("A node timeout occurred and then the following failure was recorded in the timedout node before it exited:\nwelp", clLine(14), TLWithOffset("bef\nit\n")),
			HavePanicked("bam", TLWithOffset("bef\nit\naft\n"), FailureNodeType(types.NodeTypeAfterEach)),
			HaveFailed("boom", TLWithOffset("bef\nit\naft\n"), FailureNodeType(types.NodeTypeAfterEach), clLine(21)),
			HaveFailed("flake", TLWithOffset("running\n"), clLine(29)),
			HaveFailed("flake", TLWithOffset("running\nrunning\n"), clLine(29)),
		))

		t = types.Timeline{}
		for _, event := range reporter.SpecEvents {
			t = append(t, event)
		}
		Ω(t).Should(BeTimelineExactlyMatching(
			BeSpecEvent(types.SpecEventNodeStart, types.NodeTypeBeforeEach, TLWithOffset(0), clLine(4), "a timeout"),
			BeSpecEvent(types.SpecEventNodeEnd, types.NodeTypeBeforeEach, TLWithOffset("bef\n"), clLine(4), time.Millisecond*200),
			BeSpecEvent(types.SpecEventNodeStart, types.NodeTypeIt, TLWithOffset("bef\n"), clLine(8), "spec"),
			BeSpecEvent(types.SpecEventByStart, TLWithOffset("bef\nit\n"), clLine(10), "waiting"),
			BeSpecEvent(types.SpecEventByEnd, TLWithOffset("bef\nit\n"), clLine(10), "waiting", time.Millisecond*100),
			BeSpecEvent(types.SpecEventNodeEnd, types.NodeTypeIt, TLWithOffset("bef\nit\n"), clLine(8), "spec", time.Millisecond*100),
			BeSpecEvent(types.SpecEventNodeStart, types.NodeTypeAfterEach, TLWithOffset("bef\nit\n"), clLine(16), "a timeout"),
			BeSpecEvent(types.SpecEventByStart, TLWithOffset("bef\nit\naft\n"), clLine(18), "trying to clean up"),
			BeSpecEvent(types.SpecEventNodeEnd, types.NodeTypeAfterEach, TLWithOffset("bef\nit\naft\n"), clLine(16), "a timeout"),
			BeSpecEvent(types.SpecEventNodeStart, types.NodeTypeAfterEach, TLWithOffset("bef\nit\naft\n"), clLine(21), "a timeout"),
			BeSpecEvent(types.SpecEventNodeEnd, types.NodeTypeAfterEach, TLWithOffset("bef\nit\naft\n"), clLine(21), "a timeout"),
			BeSpecEvent(types.SpecEventNodeStart, types.NodeTypeIt, TLWithOffset(0), clLine(25), "flakes repeatedly"),
			BeSpecEvent(types.SpecEventNodeEnd, types.NodeTypeIt, TLWithOffset("running\n"), clLine(25), "flakes repeatedly"),
			BeSpecEvent(types.SpecEventSpecRetry, TLWithOffset("running\n"), 1),
			BeSpecEvent(types.SpecEventNodeStart, types.NodeTypeIt, TLWithOffset("running\n"), clLine(25), "flakes repeatedly"),
			BeSpecEvent(types.SpecEventNodeEnd, types.NodeTypeIt, TLWithOffset("running\nrunning\n"), clLine(25), "flakes repeatedly"),
			BeSpecEvent(types.SpecEventSpecRetry, TLWithOffset("running\nrunning\n"), 2),
			BeSpecEvent(types.SpecEventNodeStart, types.NodeTypeIt, TLWithOffset("running\nrunning\n"), clLine(25), "flakes repeatedly"),
			BeSpecEvent(types.SpecEventNodeEnd, types.NodeTypeIt, TLWithOffset("running\nrunning\nrunning\n"), clLine(25), "flakes repeatedly"),
		))

		t = types.Timeline{}
		for _, report := range reporter.ProgressReports {
			t = append(t, report)
		}
		Ω(t).Should(BeTimelineExactlyMatching(
			BeProgressReport("Automatically polling progress:", TLWithOffset("bef\n"), clLine(4), types.NodeTypeBeforeEach),
		))

		t = types.Timeline{}
		for _, report := range reporter.ReportEntries {
			t = append(t, report)
		}
		Ω(t).Should(BeTimelineExactlyMatching(
			BeReportEntry("report-entry", ReportEntryVisibilityAlways, TLWithOffset("bef\nit\n"), clLine(13)),
		))
	})
})

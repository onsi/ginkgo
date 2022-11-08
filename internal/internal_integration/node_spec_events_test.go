package internal_integration_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("Emitting Node SpecEvents", func() {
	BeforeEach(func() {
		cl = types.NewCodeLocation(0)
		success, _ := RunFixture("emitting spec progress", func() {
			BeforeSuite(func() {})
			Describe("a container", func() {
				BeforeEach(func() {})
				It("A", func() {
					time.Sleep(time.Millisecond * 20)
					writer.Print("hello\n")
				})
				It("B", func() {})
				AfterEach(func() {})
				ReportAfterEach(func(_ SpecReport) {})
			})
			AfterSuite(func() {})
			ReportAfterEach(func(_ SpecReport) {})
		})
		Ω(success).Should(BeTrue())
	})

	It("attaches appropriate spec events to each report", func() {
		Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeBeforeSuite).Timeline()).Should(BeTimelineContaining(
			BeSpecEvent(types.SpecEventNodeStart, "TOP-LEVEL", types.NodeTypeBeforeSuite, clLine(2), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "TOP-LEVEL", types.NodeTypeBeforeSuite, clLine(2), TLWithOffset(0)),
		))

		Ω(reporter.Did.Find("A").Timeline()).Should(BeTimelineContaining(
			BeSpecEvent(types.SpecEventNodeStart, "a container", types.NodeTypeBeforeEach, clLine(4), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "a container", types.NodeTypeBeforeEach, clLine(4), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeStart, "A", types.NodeTypeIt, clLine(5), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "A", types.NodeTypeIt, clLine(5), TLWithOffset("hello\n"), time.Millisecond*20),
			BeSpecEvent(types.SpecEventNodeStart, "a container", types.NodeTypeAfterEach, clLine(10), TLWithOffset("hello\n")),
			BeSpecEvent(types.SpecEventNodeEnd, "a container", types.NodeTypeAfterEach, clLine(10), TLWithOffset("hello\n")),
			BeSpecEvent(types.SpecEventNodeStart, "a container", types.NodeTypeReportAfterEach, clLine(11), TLWithOffset("hello\n")),
			BeSpecEvent(types.SpecEventNodeEnd, "a container", types.NodeTypeReportAfterEach, clLine(11), TLWithOffset("hello\n")),
			BeSpecEvent(types.SpecEventNodeStart, "TOP-LEVEL", types.NodeTypeReportAfterEach, clLine(14), TLWithOffset("hello\n")),
			BeSpecEvent(types.SpecEventNodeEnd, "TOP-LEVEL", types.NodeTypeReportAfterEach, clLine(14), TLWithOffset("hello\n")),
		))
		Ω(reporter.Did.Find("A").Timeline()).ShouldNot(BeTimelineContaining(BeSpecEvent(types.NodeTypeReportAfterEach, clLine(15))))

		Ω(reporter.Did.Find("B").Timeline()).Should(BeTimelineContaining(
			BeSpecEvent(types.SpecEventNodeStart, "a container", types.NodeTypeBeforeEach, clLine(4), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "a container", types.NodeTypeBeforeEach, clLine(4), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeStart, "B", types.NodeTypeIt, clLine(9), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "B", types.NodeTypeIt, clLine(9), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeStart, "a container", types.NodeTypeAfterEach, clLine(10), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "a container", types.NodeTypeAfterEach, clLine(10), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeStart, "a container", types.NodeTypeReportAfterEach, clLine(11), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "a container", types.NodeTypeReportAfterEach, clLine(11), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeStart, "TOP-LEVEL", types.NodeTypeReportAfterEach, clLine(14), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "TOP-LEVEL", types.NodeTypeReportAfterEach, clLine(14), TLWithOffset(0)),
		))

		Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeAfterSuite).Timeline()).Should(BeTimelineContaining(
			BeSpecEvent(types.SpecEventNodeStart, "TOP-LEVEL", types.NodeTypeAfterSuite, clLine(13), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "TOP-LEVEL", types.NodeTypeAfterSuite, clLine(13), TLWithOffset(0)),
		))
	})

	It("emits each spec event as it goes", func() {
		timeline := types.Timeline{}
		for _, specEvent := range reporter.SpecEvents {
			timeline = append(timeline, specEvent)
		}
		Ω(timeline).Should(BeTimelineExactlyMatching(
			BeSpecEvent(types.SpecEventNodeStart, "TOP-LEVEL", types.NodeTypeBeforeSuite, clLine(2), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "TOP-LEVEL", types.NodeTypeBeforeSuite, clLine(2), TLWithOffset(0)),

			BeSpecEvent(types.SpecEventNodeStart, "a container", types.NodeTypeBeforeEach, clLine(4), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "a container", types.NodeTypeBeforeEach, clLine(4), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeStart, "A", types.NodeTypeIt, clLine(5), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "A", types.NodeTypeIt, clLine(5), TLWithOffset("hello\n"), time.Millisecond*20),
			BeSpecEvent(types.SpecEventNodeStart, "a container", types.NodeTypeAfterEach, clLine(10), TLWithOffset("hello\n")),
			BeSpecEvent(types.SpecEventNodeEnd, "a container", types.NodeTypeAfterEach, clLine(10), TLWithOffset("hello\n")),
			BeSpecEvent(types.SpecEventNodeStart, "a container", types.NodeTypeReportAfterEach, clLine(11), TLWithOffset("hello\n")),
			BeSpecEvent(types.SpecEventNodeEnd, "a container", types.NodeTypeReportAfterEach, clLine(11), TLWithOffset("hello\n")),
			BeSpecEvent(types.SpecEventNodeStart, "TOP-LEVEL", types.NodeTypeReportAfterEach, clLine(14), TLWithOffset("hello\n")),
			BeSpecEvent(types.SpecEventNodeEnd, "TOP-LEVEL", types.NodeTypeReportAfterEach, clLine(14), TLWithOffset("hello\n")),

			BeSpecEvent(types.SpecEventNodeStart, "a container", types.NodeTypeBeforeEach, clLine(4), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "a container", types.NodeTypeBeforeEach, clLine(4), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeStart, "B", types.NodeTypeIt, clLine(9), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "B", types.NodeTypeIt, clLine(9), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeStart, "a container", types.NodeTypeAfterEach, clLine(10), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "a container", types.NodeTypeAfterEach, clLine(10), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeStart, "a container", types.NodeTypeReportAfterEach, clLine(11), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "a container", types.NodeTypeReportAfterEach, clLine(11), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeStart, "TOP-LEVEL", types.NodeTypeReportAfterEach, clLine(14), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "TOP-LEVEL", types.NodeTypeReportAfterEach, clLine(14), TLWithOffset(0)),

			BeSpecEvent(types.SpecEventNodeStart, "TOP-LEVEL", types.NodeTypeAfterSuite, clLine(13), TLWithOffset(0)),
			BeSpecEvent(types.SpecEventNodeEnd, "TOP-LEVEL", types.NodeTypeAfterSuite, clLine(13), TLWithOffset(0)),
		))
	})
})

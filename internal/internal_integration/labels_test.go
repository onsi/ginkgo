package internal_integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
)

var _ = Describe("Labels", func() {
	Describe("when a suite has labelled tests", func() {
		fixture := func() {
			Describe("outer container", func() {
				It("A", rt.T("A"), Focus, Label("cat"))
				It("B", rt.T("B"), Label("dog"))
				Describe("container", Label("cow", "cat"), func() {
					It("C", rt.T("C"))
					It("D", rt.T("D"), Label("fish", "cat"))
				})
				Describe("other container", Label("     giraffe     "), func() {
					It("E", rt.T("E"))
					It("F", rt.T("F"), Label("dog"))

					Describe("inner container", Label("cow"), func() {
						It("G", rt.T("G"), Pending, Label("fish", "chicken"))
						It("H", rt.T("H"), Label("fish", "chicken"))
					})
				})
			})
		}
		BeforeEach(func() {
			conf.LabelFilter = "TopLevelLabel && (dog || cow)"
			success, hPF := RunFixture("labelled tests", fixture)
			Ω(success).Should(BeTrue())
			Ω(hPF).Should(BeFalse())
		})

		It("includes the labels in the spec report", func() {
			Ω(reporter.Did.Find("A").ContainerHierarchyLabels).Should(Equal([][]string{{}}))
			Ω(reporter.Did.Find("A").LeafNodeLabels).Should(Equal([]string{"cat"}))
			Ω(reporter.Did.Find("A").Labels()).Should(Equal([]string{"cat"}))

			Ω(reporter.Did.Find("B").ContainerHierarchyLabels).Should(Equal([][]string{{}}))
			Ω(reporter.Did.Find("B").LeafNodeLabels).Should(Equal([]string{"dog"}))
			Ω(reporter.Did.Find("B").Labels()).Should(Equal([]string{"dog"}))

			Ω(reporter.Did.Find("C").ContainerHierarchyLabels).Should(Equal([][]string{{}, {"cow", "cat"}}))
			Ω(reporter.Did.Find("C").LeafNodeLabels).Should(Equal([]string{}))
			Ω(reporter.Did.Find("C").Labels()).Should(Equal([]string{"cow", "cat"}))

			Ω(reporter.Did.Find("D").ContainerHierarchyLabels).Should(Equal([][]string{{}, {"cow", "cat"}}))
			Ω(reporter.Did.Find("D").LeafNodeLabels).Should(Equal([]string{"fish", "cat"}))
			Ω(reporter.Did.Find("D").Labels()).Should(Equal([]string{"cow", "cat", "fish"}))

			Ω(reporter.Did.Find("E").ContainerHierarchyLabels).Should(Equal([][]string{{}, {"giraffe"}}))
			Ω(reporter.Did.Find("E").LeafNodeLabels).Should(Equal([]string{}))
			Ω(reporter.Did.Find("E").Labels()).Should(Equal([]string{"giraffe"}))

			Ω(reporter.Did.Find("F").ContainerHierarchyLabels).Should(Equal([][]string{{}, {"giraffe"}}))
			Ω(reporter.Did.Find("F").LeafNodeLabels).Should(Equal([]string{"dog"}))
			Ω(reporter.Did.Find("F").Labels()).Should(Equal([]string{"giraffe", "dog"}))

			Ω(reporter.Did.Find("G").ContainerHierarchyLabels).Should(Equal([][]string{{}, {"giraffe"}, {"cow"}}))
			Ω(reporter.Did.Find("G").LeafNodeLabels).Should(Equal([]string{"fish", "chicken"}))
			Ω(reporter.Did.Find("G").Labels()).Should(Equal([]string{"giraffe", "cow", "fish", "chicken"}))

			Ω(reporter.Did.Find("H").ContainerHierarchyLabels).Should(Equal([][]string{{}, {"giraffe"}, {"cow"}}))
			Ω(reporter.Did.Find("H").LeafNodeLabels).Should(Equal([]string{"fish", "chicken"}))
			Ω(reporter.Did.Find("H").Labels()).Should(Equal([]string{"giraffe", "cow", "fish", "chicken"}))
		})

		It("includes suite labels in the suite report", func() {
			Ω(reporter.Begin.SuiteLabels).Should(Equal([]string{"TopLevelLabel"}))
			Ω(reporter.End.SuiteLabels).Should(Equal([]string{"TopLevelLabel"}))
		})

		It("honors the LabelFilter config and skips tests appropriately", func() {
			Ω(rt).Should(HaveTracked("B", "C", "D", "F", "H"))
			Ω(reporter.Did.WithState(types.SpecStatePassed).Names()).Should(ConsistOf("B", "C", "D", "F", "H"))
			Ω(reporter.Did.WithState(types.SpecStateSkipped).Names()).Should(ConsistOf("A", "E"))
			Ω(reporter.Did.WithState(types.SpecStatePending).Names()).Should(ConsistOf("G"))
			Ω(reporter.End).Should(BeASuiteSummary(true, NPassed(5), NSkipped(2), NPending(1), NSpecs(8), NWillRun(5)))
		})
	})

	Context("when a suite-level label is filtered out by the label-filter", func() {
		BeforeEach(func() {
			conf.LabelFilter = "!TopLevelLabel"
			success, hPF := RunFixture("labelled tests", func() {
				BeforeSuite(rt.T("before-suite"))
				Describe("outer container", func() {
					It("A", rt.T("A"))
					It("B", rt.T("B"))
				})
				ReportAfterEach(func(r SpecReport) {
					rt.Run("RAE-" + r.LeafNodeText)
				})
				AfterSuite(rt.T("after-suite"))
				ReportAfterSuite("AfterSuite", func(r Report) {
					rt.Run("RAS")
				})
			})
			Ω(success).Should(BeTrue())
			Ω(hPF).Should(BeFalse())
		})

		It("doesn't run anything except for reporters", func() {
			Ω(rt).Should(HaveTracked("RAE-A", "RAE-B", "RAS"))
		})

		It("skip severything", func() {
			Ω(reporter.Did.Find("A")).Should(HaveBeenSkipped())
			Ω(reporter.Did.Find("B")).Should(HaveBeenSkipped())
		})
	})
})

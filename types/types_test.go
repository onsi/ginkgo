package types_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/onsi/ginkgo/types"
)

var _ = Describe("Types", func() {
	Describe("NodeType", func() {
		Describe("Is", func() {
			It("returns true when the NodeType is in the passed-in list", func() {
				Ω(types.NodeTypeContainer.Is(types.NodeTypeIt, types.NodeTypeContainer)).Should(BeTrue())
			})

			It("returns false when the NodeType is not in the passed-in list", func() {
				Ω(types.NodeTypeContainer.Is(types.NodeTypeIt, types.NodeTypeBeforeEach)).Should(BeFalse())
			})
		})

		DescribeTable("String and AsComponentType", func(nodeType types.NodeType, expectedString string) {
			Ω(nodeType.String()).Should(Equal(expectedString))
		},
			Entry("Container", types.NodeTypeContainer, "Container"),
			Entry("It", types.NodeTypeIt, "It"),
			Entry("BeforeEach", types.NodeTypeBeforeEach, "BeforeEach"),
			Entry("JustBeforeEach", types.NodeTypeJustBeforeEach, "JustBeforeEach"),
			Entry("AfterEach", types.NodeTypeAfterEach, "AfterEach"),
			Entry("JustAfterEach", types.NodeTypeJustAfterEach, "JustAfterEach"),
			Entry("BeforeSuite", types.NodeTypeBeforeSuite, "BeforeSuite"),
			Entry("SynchronizedBeforeSuite", types.NodeTypeSynchronizedBeforeSuite, "SynchronizedBeforeSuite"),
			Entry("AfterSuite", types.NodeTypeAfterSuite, "AfterSuite"),
			Entry("SynchronizedAfterSuite", types.NodeTypeSynchronizedAfterSuite, "SynchronizedAfterSuite"),
			Entry("Invalid", types.NodeTypeInvalid, "INVALID NODE TYPE"),
		)
	})

	Describe("SpecReport Helper Functions", func() {
		Describe("CombinedOutput", func() {
			Context("with no GinkgoWriter or StdOutErr output", func() {
				It("comes back empty", func() {
					Ω(types.SpecReport{}.CombinedOutput()).Should(Equal(""))
				})
			})

			Context("wtih only StdOutErr output", func() {
				It("returns that output", func() {
					Ω(types.SpecReport{
						CapturedStdOutErr: "hello",
					}.CombinedOutput()).Should(Equal("hello"))
				})
			})

			Context("wtih only GinkgoWriter output", func() {
				It("returns that output", func() {
					Ω(types.SpecReport{
						CapturedGinkgoWriterOutput: "hello",
					}.CombinedOutput()).Should(Equal("hello"))
				})
			})

			Context("wtih both", func() {
				It("returns both concatenated", func() {
					Ω(types.SpecReport{
						CapturedGinkgoWriterOutput: "gw",
						CapturedStdOutErr:          "std",
					}.CombinedOutput()).Should(Equal("std\ngw"))
				})
			})
		})

		It("can report on whether state is a failed state", func() {
			Ω(types.SpecReport{State: types.SpecStatePending}.Failed()).Should(BeFalse())
			Ω(types.SpecReport{State: types.SpecStateSkipped}.Failed()).Should(BeFalse())
			Ω(types.SpecReport{State: types.SpecStatePassed}.Failed()).Should(BeFalse())
			Ω(types.SpecReport{State: types.SpecStateFailed}.Failed()).Should(BeTrue())
			Ω(types.SpecReport{State: types.SpecStatePanicked}.Failed()).Should(BeTrue())
			Ω(types.SpecReport{State: types.SpecStateInterrupted}.Failed()).Should(BeTrue())
		})

		It("can return a concatenated set of texts", func() {
			Ω(CurrentSpecReport().FullText()).Should(Equal("Types SpecReport Helper Functions can return a concatenated set of texts"))
		})

		It("can return the text of the Spec itself", func() {
			Ω(CurrentSpecReport().SpecText()).Should(Equal("can return the text of the Spec itself"))
		})

		It("can return the name of the file it's spec is in", func() {
			cl := types.NewCodeLocation(0)
			Ω(CurrentSpecReport().FileName()).Should(Equal(cl.FileName))
		})

		It("can return the linenumber of the file it's spec is in", func() {
			cl := types.NewCodeLocation(0)
			Ω(CurrentSpecReport().LineNumber()).Should(Equal(cl.LineNumber - 1))
		})

		It("can return it's failure's message", func() {
			report := types.SpecReport{
				Failure: types.Failure{Message: "why this failed"},
			}
			Ω(report.FailureMessage()).Should(Equal("why this failed"))
		})

		It("can return it's failure's code location", func() {
			cl := types.NewCodeLocation(0)
			report := types.SpecReport{
				Failure: types.Failure{Location: cl},
			}
			Ω(report.FailureLocation()).Should(Equal(cl))
		})
	})
})

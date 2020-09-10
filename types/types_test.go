package types_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/onsi/ginkgo/types"
)

var _ = Describe("Types", func() {
	var _ = Describe("NodeType", func() {
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
})

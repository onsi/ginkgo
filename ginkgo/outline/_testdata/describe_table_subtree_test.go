package example_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Describe", func() {
	DescribeTableSubtree("Subtree", func() {
		It("It", func() {
			Expect(true).To(BeTrue())
		})
	},
		Entry("EntrySubtree"),
		Entry("EntrySubtree2"),
	)
	DescribeTable("Table", func() {
		Expect(true).To(BeTrue())
	},
		Entry("EntryTable"),
		Entry("EntryTable2"),
	)
	DescribeTableSubtree("MultiSpec", func() {
		It("First", func() {})
		It("Second", func() {})
		Context("Nested", func() {
			It("Inner", func() {})
		})
	},
		Entry("EntryMulti"),
	)
	FDescribeTableSubtree("FocusedSubtree", func() {
		It("FocusedIt", func() {})
	},
		Entry("FocusedEntry"),
	)
	PDescribeTableSubtree("PendingSubtree", func() {
		It("PendingIt", func() {})
	},
		Entry("PendingEntry"),
	)
	XDescribeTableSubtree("XPendingSubtree", func() {
		It("XPendingIt", func() {})
	},
		Entry("XPendingEntry"),
	)
})

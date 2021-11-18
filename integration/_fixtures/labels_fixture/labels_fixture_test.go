package labels_fixture_test

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("LabelsFixture", set1, Label("chicken"), func() {
	It("works", Label("monkey", "bird"), func() {

	})

	DescribeTable("More Labels", Label("koala"), func(_ int) {},
		Entry("J", Label("beluga"), 9),
		Entry("A", Label("panda", "owl"), 7),
		Entry("C", Label("otter", "giraffe"), 5),
	)
})

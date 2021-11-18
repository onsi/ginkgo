package reporting_sub_package_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("ReportingSubPackage", func() {
	It("passes here too", func() {
	})

	It("fails here too", func() {
		fmt.Print("some std output")
		Fail("fail!")
	})

	It("panics here too", func() {
		panic("bam!")
	})
})

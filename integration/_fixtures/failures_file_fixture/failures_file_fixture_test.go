package failures_file_fixture_test

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("FailuresFileFixture", func() {
	It("should pass", func() {

	})

	It("should pass", func() {

	})

	It("should pass", func() {

	})

	It("should fail (first)", func() {
		Fail("failed")
	})

	It("should fail (second)", func() {
		Fail("failed")
	})

	It("should fail (third)", func() {
		Fail("failed")
	})

	It("should fail (fourth)", func() {
		Fail("failed")
	})
})

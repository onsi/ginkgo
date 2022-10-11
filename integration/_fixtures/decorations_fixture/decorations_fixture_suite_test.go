package decorations_fixture_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDecorationsFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DecorationsFixture Suite")
}

var countFlake = 0
// var countRepeat = 0

var _ = Describe("some decorated tests", func() {
	Describe("focused", Focus, func() {
		OffsetIt()
	})

	It("pending it", Pending, func() {

	})

	FIt("passes eventually", func() {
		countFlake += 1
		if countFlake < 3 {
			Fail("fail")
		}
	}, FlakeAttempts(3))

	// how to/should we test negative test cases?
	// FIt("fails eventually", func() {
	// 	countRepeat += 1
	// 	if countRepeat >=3 {
	// 		Fail("fail")
	// 	}
	// }, RepeatAttempts(3))

	It("focused it", Focus, func() {
		Ω(true).Should(BeTrue())
	})
})

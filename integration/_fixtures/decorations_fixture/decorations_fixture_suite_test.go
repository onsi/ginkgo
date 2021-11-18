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

var count = 0

var _ = Describe("some decorated tests", func() {
	Describe("focused", Focus, func() {
		OffsetIt()
	})

	It("pending it", Pending, func() {

	})

	It("passes eventually", func() {
		count += 1
		if count < 3 {
			Fail("fail")
		}
	}, FlakeAttempts(3))

	It("focused it", Focus, func() {
		Ω(true).Should(BeTrue())
	})
})

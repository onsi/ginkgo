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
var countRepeat = 0

var _ = Describe("some decorated specs", func() {
	Describe("focused", Focus, func() {
		OffsetIt()
	})

	It("pending it", Pending, func() {

	})

	It("focused it", Focus, func() {
		Ω(true).Should(BeTrue())
	})
})

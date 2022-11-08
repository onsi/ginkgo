package decorations_fixture_test

import (
	"fmt"
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
	It("passes eventually", func() {
		countFlake += 1
		if countFlake < 3 {
			Fail("fail")
		}
	}, FlakeAttempts(3))

	It("fails eventually", func() {
		countRepeat += 1
		if countRepeat >= 3 {
			Fail(fmt.Sprintf("failed on attempt #%d", countRepeat))
		}
	}, MustPassRepeatedly(3))
})

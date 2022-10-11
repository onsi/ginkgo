package invalid_decorations_flakeattempts_repeatattempts_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInvalidDecorations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "InvalidDecorations Suite - RepeatAttempts and FlakeAttempts")
}

var _ = Describe("invalid decorators: repeatattempts and flakeattempts", FlakeAttempts(3), RepeatAttempts(3), func() {
	It("never runs", func() {

	})
})

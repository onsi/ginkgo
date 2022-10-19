package invalid_decorations_flakeattempts_mustpassrepeatedly_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInvalidDecorations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "InvalidDecorations Suite - MustPassRepeatedly and FlakeAttempts")
}

var _ = Describe("invalid decorators: mustpassrepeatedly and flakeattempts", FlakeAttempts(3), MustPassRepeatedly(3), func() {
	It("never runs", func() {

	})
})

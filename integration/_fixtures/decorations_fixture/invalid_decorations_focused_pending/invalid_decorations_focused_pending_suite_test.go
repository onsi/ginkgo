package invalid_decorations_focused_pending_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInvalidDecorations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "InvalidDecorations Suite - Focused and Pending")
}

var _ = Describe("invalid decorators: focused and pending", Focus, Pending, func() {
	It("never runs", func() {

	})
})

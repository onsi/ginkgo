package invalid_decorations_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInvalidDecorations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "InvalidDecorations Suite")
}

var _ = Describe("focused and pending", Focus, Pending, func() {
	It("never runs", func() {

	})
})

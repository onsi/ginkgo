package large_fixture_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLargeFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LargeFixture Suite")
}

var _ = Describe("All the Tests", func() {
	for i := 0; i < 2048; i++ {
		It(fmt.Sprintf("%d", i), func() {})
	}
})

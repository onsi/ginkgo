package performance_test

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPerformanceFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Performance Fixture Suite")
}

var _ = Describe("Performance Fixture", func() {
	for i := 0; i < 10; i++ {
		It(fmt.Sprintf("sleeps %d", i), func() {
			time.Sleep(time.Millisecond * 10)
		})
	}
})

package slow_memory_hog_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/integration/_fixtures/profile_fixture/slow_memory_hog"
)

func TestSlowMemoryHog(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SlowMemoryHog Suite")
}

var _ = It("grows", func() {
	slow_memory_hog.SomethingExpensive(4)
})

var _ = It("grows a lot", func() {
	slow_memory_hog.SomethingExpensive(17)
})

var _ = It("grows way too much", func() {
	slow_memory_hog.SomethingExpensive(23)
})

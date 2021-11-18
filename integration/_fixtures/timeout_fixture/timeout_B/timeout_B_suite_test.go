package timeout_B_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTimeoutB(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TimeoutB Suite")
}

var _ = It("sleeps", func() {
	time.Sleep(5 * time.Second)
})

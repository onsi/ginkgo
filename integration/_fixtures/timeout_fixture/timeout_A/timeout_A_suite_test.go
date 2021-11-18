package timeout_A_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTimeoutA(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TimeoutA Suite")
}

var _ = It("sleeps", func() {
	time.Sleep(5 * time.Second)
})

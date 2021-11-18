package timeout_C_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTimeoutC(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TimeoutC Suite")
}

var _ = It("sleeps", func() {
	time.Sleep(5 * time.Second)
})

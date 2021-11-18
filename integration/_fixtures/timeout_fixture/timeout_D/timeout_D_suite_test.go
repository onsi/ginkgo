package timeout_D_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTimeoutD(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TimeoutD Suite")
}

var _ = It("sleeps", func() {
	time.Sleep(5 * time.Second)
})

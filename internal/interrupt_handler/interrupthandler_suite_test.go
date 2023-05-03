package interrupt_handler_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal/interrupt_handler"
	. "github.com/onsi/gomega"
)

func TestInterrupthandler(t *testing.T) {
	interrupt_handler.ABORT_POLLING_INTERVAL = 50 * time.Millisecond
	RegisterFailHandler(Fail)
	RunSpecs(t, "Interrupthandler Suite")
}

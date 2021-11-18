package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCLIInternalSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Internal Suite")
}

var _ = It(`These unit tests do not cover all the functionality in ginkgo/intenral.
The run, compile, and profiles functions are all integration tests in ginkgo/integration.`, func() {})

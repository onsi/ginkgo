package testsuite_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTestsuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Testsuite Suite")
}

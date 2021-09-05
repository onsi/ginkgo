package reporters_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestReporters(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Reporters Suite")
}

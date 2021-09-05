package failer_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFailer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Failer Suite")
}

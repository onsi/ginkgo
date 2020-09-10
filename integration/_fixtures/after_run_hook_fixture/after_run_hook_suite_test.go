package suite_command_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAfterRunHook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "After Run Hook Suite")
}

package suite_command_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAfterRunHook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "After Run Hook Suite")
}

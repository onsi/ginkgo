package main_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInterceptorSleepFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TestInterceptorSleepFixture Suite")
}

var _ = Describe("Ensuring that ginkgo -p does not hang when output is intercepted", func() {
	It("ginkgo -p should not hang on this spec", func() {
		fmt.Fprintln(os.Stdout, "Some STDOUT output")
		fmt.Fprintln(os.Stderr, "Some STDERR output")
		cmd := exec.Command("sleep", "60")
		Ω(cmd.Start()).Should(Succeed())
	})
})

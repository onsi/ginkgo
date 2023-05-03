package main_test

import (
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInterceptorHangFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TestInterceptorHangFixture Suite")
}

var _ = Describe("Ensuring that ginkgo -p does not hang when output is intercepted", func() {
	It("ginkgo -p should not hang on this spec", func() {
		// see https://github.com/onsi/ginkgo/issues/1191
		cmd := exec.Command("sleep", "60")
		Ω(cmd.Start()).Should(Succeed())
	})
})

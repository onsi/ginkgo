package integration_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os/exec"

	"testing"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)

	installGinkgoCommand := exec.Command("go", "install", "github.com/onsi/ginkgo/ginkgo")
	err := installGinkgoCommand.Run()
	if err != nil {
		fmt.Printf("Failed to compile Ginkgo\n\t%s", err.Error())
	}

	RunSpecs(t, "Integration Suite")
}

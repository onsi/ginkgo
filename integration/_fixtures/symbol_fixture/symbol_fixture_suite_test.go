package symbol_fixture_test

import (
	"fmt"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSymbolFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SymbolFixture Suite")
}

var _ = It("prints out its symbols", func() {
	cmd := exec.Command("go", "tool", "nm", "symbol_fixture.test")
	output, _ := cmd.CombinedOutput()
	fmt.Println(string(output))
})

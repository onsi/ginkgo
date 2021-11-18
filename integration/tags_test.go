package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Tags", func() {
	BeforeEach(func() {
		fm.MountFixture("tags")
	})

	It("should honor the passed in -tags flag", func() {
		session := startGinkgo(fm.PathTo("tags"), "--no-color")
		Eventually(session).Should(gexec.Exit(0))
		output := string(session.Out.Contents())
		Ω(output).Should(ContainSubstring("Ran 1 of 1 Specs"))

		session = startGinkgo(fm.PathTo("tags"), "--no-color", "-tags=complex_tests")
		Eventually(session).Should(gexec.Exit(0))
		output = string(session.Out.Contents())
		Ω(output).Should(ContainSubstring("Ran 3 of 3 Specs"))
	})
})

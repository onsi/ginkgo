package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Deprecations", func() {
	BeforeEach(func() {
		fm.MountFixture("deprecated_features")
	})

	It("runs, succeeds, and emits deprecation warnings", func() {
		session := startGinkgo(fm.PathTo("deprecated_features"), "--randomizeAllSpecs", "--stream")
		Eventually(session).Should(gexec.Exit(0))
		contents := string(session.Out.Contents()) + string(session.Err.Contents())

		Ω(contents).Should(ContainSubstring("You are passing a Done channel to a test node to test asynchronous behavior."))
		Ω(contents).Should(ContainSubstring("Measure is deprecated and has been removed from Ginkgo V2."))
		Ω(contents).Should(ContainSubstring("--stream is deprecated"))
		Ω(contents).Should(ContainSubstring("--randomizeAllSpecs is deprecated"))
	})
})

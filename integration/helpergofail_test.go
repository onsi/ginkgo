package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Failing helpers with new go routines", func() {

	BeforeEach(func() {
		fm.MountFixture("helpergo")
	})

	It("should fail with correct details", func() {
		session := startGinkgo(fm.PathTo("helpergo"), "--no-color")
		Eventually(session).Should(gexec.Exit(1))
		output := string(session.Out.Contents())

		By(output)

		Expect(output).NotTo(ContainSubstring("SHOULD NOT SEE THIS"))

		Expect(output).To(MatchRegexp(
			`a user assertion failure\n.*\n.*\n.* at: .*/helpergo_fixture_test.go:14 `))
		Expect(output).To(MatchRegexp(
			`a helper assertion 1 failure\n.*\n.*\n.* at: .*/helpergo_fixture_test.go:21 `))
		Expect(output).To(MatchRegexp(
			`a helper assertion 2 failure\n.*\n.*\n.* at: .*/helpergo_fixture_test.go:32 `))
		Expect(output).To(MatchRegexp(
			`Test Panicked\n.*\n\n.*JKL305\n\n.*Full Stack Trace\n.*\n.*\n.*\n.*helpergo_fixture_test.go:44`))
	})

})

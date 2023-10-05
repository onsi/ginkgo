package integration_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Preview", func() {
	BeforeEach(func() {
		fm.MountFixture("preview")
	})

	It("previews the specs, honoring the passed in flags", func() {
		os.Setenv("PREVIEW", "true")
		DeferCleanup(os.Unsetenv, "PREVIEW")
		session := startGinkgo(fm.PathTo("preview"), "--label-filter=elephant")
		Eventually(session).Should(gexec.Exit(0))
		Ω(session).Should(gbytes.Say("passed specs A"))
		Ω(session).Should(gbytes.Say("passed specs B"))
		Ω(session).Should(gbytes.Say("skipped specs C"))
		Ω(session).Should(gbytes.Say("skipped specs D"))
	})

	It("fails if running in parallel", func() {
		os.Setenv("PREVIEW", "true")
		DeferCleanup(os.Unsetenv, "PREVIEW")
		session := startGinkgo(fm.PathTo("preview"), "--procs=2")
		Eventually(session).Should(gexec.Exit(1))
		Ω(session.Err).Should(gbytes.Say(`Ginkgo only supports PreviewSpecs\(\) in serial mode\.`))
	})

	It("fails if you attempt to both run and preview specs", func() {
		os.Setenv("PREVIEW", "true")
		DeferCleanup(os.Unsetenv, "PREVIEW")
		os.Setenv("RUN", "true")
		DeferCleanup(os.Unsetenv, "RUN")
		session := startGinkgo(fm.PathTo("preview"))
		Eventually(session).Should(gexec.Exit(1))
		Ω(session).Should(gbytes.Say(`It looks like you are calling RunSpecs and PreviewSpecs in the same invocation`))
	})
})

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

	It("succeeds if you attempt to both run and preview specs", func() {
		os.Setenv("PREVIEW", "true")
		DeferCleanup(os.Unsetenv, "PREVIEW")
		os.Setenv("RUN", "true")
		DeferCleanup(os.Unsetenv, "RUN")
		session := startGinkgo(fm.PathTo("preview"))
		Eventually(session).Should(gexec.Exit(0))
		Ω(session).Should(gbytes.Say(`passed specs A`))
		Ω(session).Should(gbytes.Say(`passed specs B`))
		Ω(session).Should(gbytes.Say(`passed specs C`))
		Ω(session).Should(gbytes.Say(`passed specs D`))
		Ω(session).Should(gbytes.Say(`Ran 4 of 4 Specs`))
	})

	It("works if you run in parallel", func() {
		os.Setenv("PREVIEW", "true")
		DeferCleanup(os.Unsetenv, "PREVIEW")
		os.Setenv("RUN", "true")
		DeferCleanup(os.Unsetenv, "RUN")
		session := startGinkgo(fm.PathTo("preview"), "-p")
		Eventually(session).Should(gexec.Exit(0))
		Ω(session).Should(gbytes.Say(`Ran 4 of 4 Specs`))
	})
})

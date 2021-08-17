package integration_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Watch", func() {
	var session *gexec.Session

	BeforeEach(func() {
		fm.MountFixture("watch", "A")
		fm.MountFixture("watch", "B")
		fm.MountFixture("watch", "C")
	})

	modifyFile := func(path string) {
		time.Sleep(time.Second)
		content, err := os.ReadFile(path)
		Ω(err).ShouldNot(HaveOccurred())
		content = append(content, []byte("//")...)
		err = os.WriteFile(path, content, 0666)
		Ω(err).ShouldNot(HaveOccurred())
	}

	modifyCode := func(pkgToModify string) {
		path := filepath.Join(pkgToModify, pkgToModify+".go")
		modifyFile(fm.PathTo("watch", path))
	}

	modifyJSON := func(pkgToModify string) {
		path := filepath.Join(pkgToModify, pkgToModify+".json")
		modifyFile(fm.PathTo("watch", path))
	}

	modifyTest := func(pkgToModify string) {
		path := filepath.Join(pkgToModify, pkgToModify+"_test.go")
		modifyFile(fm.PathTo("watch", path))
	}

	AfterEach(func() {
		if session != nil {
			session.Kill().Wait()
		}
	})

	It("should be set up correctly", func() {
		session = startGinkgo(fm.PathTo("watch"), "-r")
		Eventually(session).Should(gexec.Exit(0))
		Ω(session.Out.Contents()).Should(ContainSubstring("A Suite"))
		Ω(session.Out.Contents()).Should(ContainSubstring("B Suite"))
		Ω(session.Out.Contents()).Should(ContainSubstring("C Suite"))
		Ω(session.Out.Contents()).Should(ContainSubstring("Ginkgo ran 3 suites"))
	})

	Context("when watching just one test suite", func() {
		It("should immediately run, and should rerun when the test suite changes", func() {
			session = startGinkgo(fm.PathTo("watch"), "watch", "-succinct", "A")
			Eventually(session).Should(gbytes.Say("A Suite"))
			modifyCode("A")
			Eventually(session).Should(gbytes.Say("Detected changes in"))
			Eventually(session).Should(gbytes.Say("A Suite"))
			session.Kill().Wait()
		})
	})

	Context("when watching several test suites", func() {
		It("should not immediately run, but should rerun a test when its code changes", func() {
			session = startGinkgo(fm.PathTo("watch"), "watch", "-succinct", "-r")
			Eventually(session).Should(gbytes.Say("Identified 3 test suites"))
			Consistently(session).ShouldNot(gbytes.Say("A Suite|B Suite|C Suite"))
			modifyCode("A")
			Eventually(session).Should(gbytes.Say("Detected changes in"))
			Eventually(session).Should(gbytes.Say("A Suite"))
			Consistently(session).ShouldNot(gbytes.Say("B Suite|C Suite"))
			session.Kill().Wait()
		})
	})

	Describe("watching dependencies", func() {
		Context("with a depth of 2", func() {
			It("should watch down to that depth", func() {
				session = startGinkgo(fm.PathTo("watch"), "watch", "-succinct", "-r", "-depth=2")
				Eventually(session).Should(gbytes.Say("Identified 3 test suites"))
				Eventually(session).Should(gbytes.Say(`A \[`))
				Eventually(session).Should(gbytes.Say(`B \[`))
				Eventually(session).Should(gbytes.Say(`C \[`))

				modifyCode("A")
				Eventually(session).Should(gbytes.Say("Detected changes in"))
				Eventually(session).Should(gbytes.Say("A Suite"))
				Consistently(session).ShouldNot(gbytes.Say("B Suite|C Suite"))

				modifyCode("B")
				Eventually(session).Should(gbytes.Say("Detected changes in"))
				Eventually(session).Should(gbytes.Say("B Suite"))
				Eventually(session).Should(gbytes.Say("A Suite"))
				Consistently(session).ShouldNot(gbytes.Say("C Suite"))

				modifyCode("C")
				Eventually(session).Should(gbytes.Say("Detected changes in"))
				Eventually(session).Should(gbytes.Say("C Suite"))
				Eventually(session).Should(gbytes.Say("B Suite"))
				Eventually(session).Should(gbytes.Say("A Suite"))
			})
		})

		Context("with a depth of 1", func() {
			It("should watch down to that depth", func() {
				session = startGinkgo(fm.PathTo("watch"), "watch", "-succinct", "-r", "-depth=1")
				Eventually(session).Should(gbytes.Say("Identified 3 test suites"))
				Eventually(session).Should(gbytes.Say(`A \[`))
				Eventually(session).Should(gbytes.Say(`B \[`))
				Eventually(session).Should(gbytes.Say(`C \[`))

				modifyCode("A")
				Eventually(session).Should(gbytes.Say("Detected changes in"))
				Eventually(session).Should(gbytes.Say("A Suite"))
				Consistently(session).ShouldNot(gbytes.Say("B Suite|C Suite"))

				modifyCode("B")
				Eventually(session).Should(gbytes.Say("Detected changes in"))
				Eventually(session).Should(gbytes.Say("B Suite"))
				Eventually(session).Should(gbytes.Say("A Suite"))
				Consistently(session).ShouldNot(gbytes.Say("C Suite"))

				modifyCode("C")
				Eventually(session).Should(gbytes.Say("Detected changes in"))
				Eventually(session).Should(gbytes.Say("C Suite"))
				Eventually(session).Should(gbytes.Say("B Suite"))
				Consistently(session).ShouldNot(gbytes.Say("A Suite"))
			})
		})

		Context("with a depth of 0", func() {
			It("should not watch any dependencies", func() {
				session = startGinkgo(fm.PathTo("watch"), "watch", "-succinct", "-r", "-depth=0")
				Eventually(session).Should(gbytes.Say("Identified 3 test suites"))
				Eventually(session).Should(gbytes.Say(`A \[`))
				Eventually(session).Should(gbytes.Say(`B \[`))
				Eventually(session).Should(gbytes.Say(`C \[`))

				modifyCode("A")
				Eventually(session).Should(gbytes.Say("Detected changes in"))
				Eventually(session).Should(gbytes.Say("A Suite"))
				Consistently(session).ShouldNot(gbytes.Say("B Suite|C Suite"))

				modifyCode("B")
				Eventually(session).Should(gbytes.Say("Detected changes in"))
				Eventually(session).Should(gbytes.Say("B Suite"))
				Consistently(session).ShouldNot(gbytes.Say("A Suite|C Suite"))

				modifyCode("C")
				Eventually(session).Should(gbytes.Say("Detected changes in"))
				Eventually(session).Should(gbytes.Say("C Suite"))
				Consistently(session).ShouldNot(gbytes.Say("A Suite|B Suite"))
			})
		})

		It("should not trigger dependents when tests are changed", func() {
			session = startGinkgo(fm.PathTo("watch"), "watch", "-succinct", "-r", "-depth=2")
			Eventually(session).Should(gbytes.Say("Identified 3 test suites"))
			Eventually(session).Should(gbytes.Say(`A \[`))
			Eventually(session).Should(gbytes.Say(`B \[`))
			Eventually(session).Should(gbytes.Say(`C \[`))

			modifyTest("A")
			Eventually(session).Should(gbytes.Say("Detected changes in"))
			Eventually(session).Should(gbytes.Say("A Suite"))
			Consistently(session).ShouldNot(gbytes.Say("B Suite|C Suite"))

			modifyTest("B")
			Eventually(session).Should(gbytes.Say("Detected changes in"))
			Eventually(session).Should(gbytes.Say("B Suite"))
			Consistently(session).ShouldNot(gbytes.Say("A Suite|C Suite"))

			modifyTest("C")
			Eventually(session).Should(gbytes.Say("Detected changes in"))
			Eventually(session).Should(gbytes.Say("C Suite"))
			Consistently(session).ShouldNot(gbytes.Say("A Suite|B Suite"))
		})
	})

	Describe("adjusting the watch regular expression", func() {
		Describe("the default regular expression", func() {
			It("should only trigger when go files are changed", func() {
				session = startGinkgo(fm.PathTo("watch"), "watch", "-succinct", "-r", "-depth=2")
				Eventually(session).Should(gbytes.Say("Identified 3 test suites"))
				Eventually(session).Should(gbytes.Say(`A \[`))
				Eventually(session).Should(gbytes.Say(`B \[`))
				Eventually(session).Should(gbytes.Say(`C \[`))

				modifyJSON("C")
				Consistently(session).ShouldNot(gbytes.Say("Detected changes in"))
				Consistently(session).ShouldNot(gbytes.Say("A Suite|B Suite|C Suite"))
			})
		})

		Describe("modifying the regular expression", func() {
			It("should trigger if the regexp matches", func() {
				session = startGinkgo(fm.PathTo("watch"), "watch", "-succinct", "-r", "-depth=2", `-watch-regexp=\.json$`)
				Eventually(session).Should(gbytes.Say("Identified 3 test suites"))
				Eventually(session).Should(gbytes.Say(`A \[`))
				Eventually(session).Should(gbytes.Say(`B \[`))
				Eventually(session).Should(gbytes.Say(`C \[`))

				modifyJSON("C")
				Eventually(session).Should(gbytes.Say("Detected changes in"))
				Eventually(session).Should(gbytes.Say("C Suite"))
				Eventually(session).Should(gbytes.Say("B Suite"))
				Eventually(session).Should(gbytes.Say("A Suite"))
			})
		})
	})

	Describe("when new test suite is added", func() {
		It("should start monitoring that test suite", func() {
			session = startGinkgo(fm.PathTo("watch"), "watch", "-succinct", "-r", "-depth=1")
			Eventually(session).Should(gbytes.Say("Watching 3 suites"))

			fm.MountFixture("watch", "D")

			Eventually(session).Should(gbytes.Say("Detected 1 new suite"))
			Eventually(session).Should(gbytes.Say(`D \[`))
			Eventually(session).Should(gbytes.Say("D Suite"))

			modifyCode("D")

			Eventually(session).Should(gbytes.Say("Detected changes in"))
			Eventually(session).Should(gbytes.Say("D Suite"))

			modifyCode("C")

			Eventually(session).Should(gbytes.Say("Detected changes in"))
			Eventually(session).Should(gbytes.Say("C Suite"))
			Eventually(session).Should(gbytes.Say("D Suite"))
		})
	})
})

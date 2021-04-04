package integration_test

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Running Specs", func() {
	denoter := "•"
	if runtime.GOOS == "windows" {
		denoter = "+"
	}

	Context("when pointed at the current directory", func() {
		BeforeEach(func() {
			fm.MountFixture("passing_ginkgo_tests")
		})

		It("should run the tests in the working directory", func() {
			session := startGinkgo(fm.PathTo("passing_ginkgo_tests"), "--no-color")
			Eventually(session).Should(gexec.Exit(0))
			output := string(session.Out.Contents())

			Ω(output).Should(ContainSubstring("Running Suite: Passing_ginkgo_tests Suite"))
			Ω(output).Should(ContainSubstring(strings.Repeat(denoter, 4)))
			Ω(output).Should(ContainSubstring("SUCCESS! -- 4 Passed"))
			Ω(output).Should(ContainSubstring("Test Suite Passed"))
		})
	})

	Context("when passed an explicit package to run", func() {
		BeforeEach(func() {
			fm.MountFixture("passing_ginkgo_tests")
		})

		It("should run the ginkgo style tests", func() {
			session := startGinkgo(fm.TmpDir, "--no-color", "passing_ginkgo_tests")
			Eventually(session).Should(gexec.Exit(0))
			output := string(session.Out.Contents())

			Ω(output).Should(ContainSubstring("Running Suite: Passing_ginkgo_tests Suite"))
			Ω(output).Should(ContainSubstring(strings.Repeat(denoter, 4)))
			Ω(output).Should(ContainSubstring("SUCCESS! -- 4 Passed"))
			Ω(output).Should(ContainSubstring("Test Suite Passed"))
		})
	})

	Context("when passed a number of packages to run", func() {
		BeforeEach(func() {
			fm.MountFixture("passing_ginkgo_tests")
			fm.MountFixture("more_ginkgo_tests")
		})

		It("should run the ginkgo style tests", func() {
			session := startGinkgo(fm.TmpDir, "--no-color", "--succinct=false", "passing_ginkgo_tests", "more_ginkgo_tests")
			Eventually(session).Should(gexec.Exit(0))
			output := string(session.Out.Contents())

			Ω(output).Should(ContainSubstring("Running Suite: Passing_ginkgo_tests Suite"))
			Ω(output).Should(ContainSubstring("Running Suite: More_ginkgo_tests Suite"))
			Ω(output).Should(ContainSubstring("Test Suite Passed"))
		})
	})

	Context("when passed a number of packages to run, some of which have focused tests", func() {
		BeforeEach(func() {
			fm.MountFixture("passing_ginkgo_tests")
			fm.MountFixture("more_ginkgo_tests")
			fm.MountFixture("focused")
		})

		It("should exit with a status code of 2 and explain why", func() {
			session := startGinkgo(fm.TmpDir, "--no-color", "--succinct=false", "-r")
			Eventually(session).Should(gexec.Exit(types.GINKGO_FOCUS_EXIT_CODE))
			output := string(session.Out.Contents())

			Ω(output).Should(ContainSubstring("Running Suite: Passing_ginkgo_tests Suite"))
			Ω(output).Should(ContainSubstring("Running Suite: More_ginkgo_tests Suite"))
			Ω(output).Should(ContainSubstring("Test Suite Passed"))
			Ω(output).Should(ContainSubstring("Detected Programmatic Focus - setting exit status to %d", types.GINKGO_FOCUS_EXIT_CODE))
		})

		Context("when the GINKGO_EDITOR_INTEGRATION environment variable is set", func() {
			BeforeEach(func() {
				os.Setenv("GINKGO_EDITOR_INTEGRATION", "true")
			})
			AfterEach(func() {
				os.Setenv("GINKGO_EDITOR_INTEGRATION", "")
			})
			It("should exit with a status code of 0 to allow a coverage file to be generated", func() {
				session := startGinkgo(fm.TmpDir, "--no-color", "--succinct=false", "-r")
				Eventually(session).Should(gexec.Exit(0))
				output := string(session.Out.Contents())

				Ω(output).Should(ContainSubstring("Running Suite: Passing_ginkgo_tests Suite"))
				Ω(output).Should(ContainSubstring("Running Suite: More_ginkgo_tests Suite"))
				Ω(output).Should(ContainSubstring("Test Suite Passed"))
			})
		})
	})

	Context("when told to skipPackages", func() {
		BeforeEach(func() {
			fm.MountFixture("passing_ginkgo_tests")
			fm.MountFixture("more_ginkgo_tests")
			fm.MountFixture("focused")
		})

		It("should skip packages that match the list", func() {
			session := startGinkgo(fm.TmpDir, "--no-color", "--skip-package=more_ginkgo_tests,focused", "-r")
			Eventually(session).Should(gexec.Exit(0))
			output := string(session.Out.Contents())

			Ω(output).Should(ContainSubstring("Passing_ginkgo_tests Suite"))
			Ω(output).ShouldNot(ContainSubstring("More_ginkgo_tests Suite"))
			Ω(output).ShouldNot(ContainSubstring("Focused_fixture Suite"))
			Ω(output).Should(ContainSubstring("Test Suite Passed"))
		})

		Context("when all packages are skipped", func() {
			It("should not run anything, but still exit 0", func() {
				session := startGinkgo(fm.TmpDir, "--no-color", "--skip-package=passing_ginkgo_tests,more_ginkgo_tests,focused", "-r")
				Eventually(session).Should(gexec.Exit(0))
				output := string(session.Out.Contents())
				errOutput := string(session.Err.Contents())

				Ω(errOutput).Should(ContainSubstring("All tests skipped!"))
				Ω(output).ShouldNot(ContainSubstring("Passing_ginkgo_tests Suite"))
				Ω(output).ShouldNot(ContainSubstring("More_ginkgo_tests Suite"))
				Ω(output).ShouldNot(ContainSubstring("Focused_fixture Suite"))
				Ω(output).ShouldNot(ContainSubstring("Test Suite Passed"))
			})
		})
	})

	Context("when there are no tests to run", func() {
		It("should exit 1", func() {
			session := startGinkgo(fm.TmpDir, "--no-color", "-r")
			Eventually(session).Should(gexec.Exit(1))
			output := string(session.Err.Contents())

			Ω(output).Should(ContainSubstring("Found no test suites"))
		})
	})

	Context("when there are test files but `go test` reports there are no tests to run", func() {
		BeforeEach(func() {
			fm.MountFixture("no_test_fn")
		})

		It("suggests running ginkgo bootstrap", func() {
			session := startGinkgo(fm.TmpDir, "--no-color", "-r")
			Eventually(session).Should(gexec.Exit(0))
			output := string(session.Err.Contents())

			Ω(output).Should(ContainSubstring(`Found no test suites, did you forget to run "ginkgo bootstrap"?`))
		})

		It("fails if told to requireSuite", func() {
			session := startGinkgo(fm.TmpDir, "--no-color", "-r", "-require-suite")
			Eventually(session).Should(gexec.Exit(1))
			output := string(session.Err.Contents())

			Ω(output).Should(ContainSubstring(`Found no test suites, did you forget to run "ginkgo bootstrap"?`))
		})
	})

	Context("when told to randomize-suites", func() {
		BeforeEach(func() {
			fm.MountFixture("passing_ginkgo_tests")
			fm.MountFixture("more_ginkgo_tests")
		})

		It("should mix up the order of the test suites", func() {
			session := startGinkgo(fm.TmpDir, "--no-color", "--randomize-suites", "-r", "--seed=1")
			Eventually(session).Should(gexec.Exit(0))

			Ω(session).Should(gbytes.Say("More_ginkgo_tests Suite"))
			Ω(session).Should(gbytes.Say("Passing_ginkgo_tests Suite"))

			session = startGinkgo(fm.TmpDir, "--no-color", "--randomize-suites", "-r", "--seed=4")
			Eventually(session).Should(gexec.Exit(0))

			Ω(session).Should(gbytes.Say("Passing_ginkgo_tests Suite"))
			Ω(session).Should(gbytes.Say("More_ginkgo_tests Suite"))
		})
	})

	Context("when pointed at a package with xunit style tests", func() {
		BeforeEach(func() {
			fm.MountFixture("xunit")
		})

		It("should run the xunit style tests", func() {
			session := startGinkgo(fm.PathTo("xunit"))
			Eventually(session).Should(gexec.Exit(0))
			output := string(session.Out.Contents())

			Ω(output).Should(ContainSubstring("--- PASS: TestAlwaysTrue"))
			Ω(output).Should(ContainSubstring("Test Suite Passed"))
		})
	})

	Context("when pointed at a package with no tests", func() {
		BeforeEach(func() {
			fm.MountFixture("no_tests")
		})

		It("should fail", func() {
			session := startGinkgo(fm.PathTo("no_tests"), "--no-color")
			Eventually(session).Should(gexec.Exit(1))

			Ω(session.Err.Contents()).Should(ContainSubstring("Found no test suites"))
		})
	})

	Context("when pointed at a package that fails to compile", func() {
		BeforeEach(func() {
			fm.MountFixture("does_not_compile")
		})

		It("should fail", func() {
			session := startGinkgo(fm.PathTo("does_not_compile"), "--no-color")
			Eventually(session).Should(gexec.Exit(1))
			output := string(session.Out.Contents())

			Ω(output).Should(ContainSubstring("Failed to compile"))
		})
	})

	Context("when running in parallel", func() {
		BeforeEach(func() {
			fm.MountFixture("passing_ginkgo_tests")
		})

		Context("with a specific number of -nodes", func() {
			It("should use the specified number of nodes", func() {
				session := startGinkgo(fm.PathTo("passing_ginkgo_tests"), "--no-color", "-succinct", "-nodes=2")
				Eventually(session).Should(gexec.Exit(0))
				output := string(session.Out.Contents())

				Ω(output).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 4/4 specs - 2 nodes [%s]{4} SUCCESS! \d+(\.\d+)?[muµ]s`, regexp.QuoteMeta(denoter)))
				Ω(output).Should(ContainSubstring("Test Suite Passed"))
			})
		})

		Context("with -p", func() {
			It("it should autocompute the number of nodes", func() {
				session := startGinkgo(fm.PathTo("passing_ginkgo_tests"), "--no-color", "-succinct", "-p")
				Eventually(session).Should(gexec.Exit(0))
				output := string(session.Out.Contents())

				nodes := runtime.NumCPU()
				if nodes == 1 {
					Skip("Can't test parallel testings with 1 CPU")
				}
				if nodes > 4 {
					nodes = nodes - 1
				}
				Ω(output).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 4/4 specs - %d nodes [%s]{4} SUCCESS! \d+(\.\d+)?[muµ]?s`, nodes, regexp.QuoteMeta(denoter)))
				Ω(output).Should(ContainSubstring("Test Suite Passed"))
			})
		})
	})

	Context("when running in parallel with -debug-parallel", func() {
		BeforeEach(func() {
			fm.MountFixture("debug_parallel")
		})

		Context("without -v", func() {
			It("should emit node output to files on disk", func() {
				session := startGinkgo(fm.PathTo("debug_parallel"), "--nodes=2", "--debug-parallel")
				Eventually(session).Should(gexec.Exit(0))

				f0 := fm.ContentOf("debug_parallel", "ginkgo-node-1.log")
				f1 := fm.ContentOf("debug_parallel", "ginkgo-node-2.log")
				content := f0 + f1

				for i := 0; i < 10; i += 1 {
					Ω(content).Should(ContainSubstring("StdOut %d\n", i))
					Ω(content).Should(ContainSubstring("GinkgoWriter %d\n", i))
				}
			})
		})

		Context("with -v", func() {
			It("should emit node output to files on disk, without duplicating the GinkgoWriter output", func() {
				session := startGinkgo(fm.PathTo("debug_parallel"), "--nodes=2", "--debug-parallel", "-v")
				Eventually(session).Should(gexec.Exit(0))

				f0 := fm.ContentOf("debug_parallel", "ginkgo-node-1.log")
				f1 := fm.ContentOf("debug_parallel", "ginkgo-node-2.log")
				content := f0 + f1

				out := strings.Split(content, "GinkgoWriter 2")
				Ω(out).Should(HaveLen(2))
			})
		})
	})

	Context("when running multiple tests", func() {
		BeforeEach(func() {
			fm.MountFixture("passing_ginkgo_tests")
			fm.MountFixture("more_ginkgo_tests")
		})

		Context("when all the tests pass", func() {
			Context("with the -r flag", func() {
				It("should run all the tests (in succinct mode) and succeed", func() {
					session := startGinkgo(fm.TmpDir, "--no-color", "-r", ".")
					Eventually(session).Should(gexec.Exit(0))
					output := string(session.Out.Contents())

					outputLines := strings.Split(output, "\n")
					Ω(outputLines[0]).Should(MatchRegexp(`\[\d+\] More_ginkgo_tests Suite - 2/2 specs [%s]{2} SUCCESS! \d+(\.\d+)?[muµ]s PASS`, regexp.QuoteMeta(denoter)))
					Ω(outputLines[1]).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 4/4 specs [%s]{4} SUCCESS! \d+(\.\d+)?[muµ]s PASS`, regexp.QuoteMeta(denoter)))
					Ω(output).Should(ContainSubstring("Test Suite Passed"))
				})
			})
			Context("with a trailing /...", func() {
				It("should run all the tests (in succinct mode) and succeed", func() {
					session := startGinkgo(fm.TmpDir, "--no-color", "./...")
					Eventually(session).Should(gexec.Exit(0))
					output := string(session.Out.Contents())

					outputLines := strings.Split(output, "\n")
					Ω(outputLines[0]).Should(MatchRegexp(`\[\d+\] More_ginkgo_tests Suite - 2/2 specs [%s]{2} SUCCESS! \d+(\.\d+)?[muµ]s PASS`, regexp.QuoteMeta(denoter)))
					Ω(outputLines[1]).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 4/4 specs [%s]{4} SUCCESS! \d+(\.\d+)?[muµ]s PASS`, regexp.QuoteMeta(denoter)))
					Ω(output).Should(ContainSubstring("Test Suite Passed"))
				})
			})
		})

		Context("when one of the packages has a failing tests", func() {
			BeforeEach(func() {
				fm.MountFixture("failing_ginkgo_tests")
			})

			It("should fail and stop running tests", func() {
				session := startGinkgo(fm.TmpDir, "--no-color", "passing_ginkgo_tests", "failing_ginkgo_tests", "more_ginkgo_tests")
				Eventually(session).Should(gexec.Exit(1))
				output := string(session.Out.Contents())

				outputLines := strings.Split(output, "\n")
				Ω(outputLines[0]).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 4/4 specs [%s]{4} SUCCESS! \d+(\.\d+)?[muµ]s PASS`, regexp.QuoteMeta(denoter)))
				Ω(outputLines[1]).Should(MatchRegexp(`\[\d+\] Failing_ginkgo_tests Suite - 2/2 specs`))
				Ω(output).Should(ContainSubstring(fmt.Sprintf("%s [FAILED]", denoter)))
				Ω(output).ShouldNot(ContainSubstring("More_ginkgo_tests Suite"))
				Ω(output).Should(ContainSubstring("Test Suite Failed"))
			})
		})

		Context("when one of the packages fails to compile", func() {
			BeforeEach(func() {
				fm.MountFixture("does_not_compile")
			})

			It("should fail and stop running tests", func() {
				session := startGinkgo(fm.TmpDir, "--no-color", "passing_ginkgo_tests", "does_not_compile", "more_ginkgo_tests")
				Eventually(session).Should(gexec.Exit(1))
				output := string(session.Out.Contents())

				outputLines := strings.Split(output, "\n")
				Ω(outputLines[0]).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 4/4 specs [%s]{4} SUCCESS! \d+(\.\d+)?[muµ]s PASS`, regexp.QuoteMeta(denoter)))
				Ω(outputLines[1]).Should(ContainSubstring("Failed to compile does_not_compile:"))
				Ω(output).ShouldNot(ContainSubstring("More_ginkgo_tests Suite"))
				Ω(output).Should(ContainSubstring("Test Suite Failed"))
			})
		})

		Context("when either is the case, but the keep-going flag is set", func() {
			BeforeEach(func() {
				fm.MountFixture("does_not_compile")
				fm.MountFixture("failing_ginkgo_tests")
			})

			It("should soldier on", func() {
				session := startGinkgo(fm.TmpDir, "--no-color", "-keep-going", "passing_ginkgo_tests", "does_not_compile", "failing_ginkgo_tests", "more_ginkgo_tests")
				Eventually(session).Should(gexec.Exit(1))
				output := string(session.Out.Contents())

				outputLines := strings.Split(output, "\n")
				Ω(outputLines[0]).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 4/4 specs [%s]{4} SUCCESS! \d+(\.\d+)?[muµ]s PASS`, regexp.QuoteMeta(denoter)))
				Ω(outputLines[1]).Should(ContainSubstring("Failed to compile does_not_compile:"))
				Ω(output).Should(MatchRegexp(`\[\d+\] Failing_ginkgo_tests Suite - 2/2 specs`))
				Ω(output).Should(ContainSubstring(fmt.Sprintf("%s [FAILED]", denoter)))
				Ω(output).Should(MatchRegexp(`\[\d+\] More_ginkgo_tests Suite - 2/2 specs [%s]{2} SUCCESS! \d+(\.\d+)?[muµ]s PASS`, regexp.QuoteMeta(denoter)))
				Ω(output).Should(ContainSubstring("Test Suite Failed"))
			})
		})
	})
})

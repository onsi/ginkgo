package integration_test

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
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

		It("should mix up the order of the test suites", Label("slow"), func() {
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

		It("should run the xunit style tests, always setting -test.v and passing in supported go test flags", func() {
			session := startGinkgo(fm.PathTo("xunit"), "-blockprofile=block-profile.out")
			Eventually(session).Should(gexec.Exit(0))
			output := string(session.Out.Contents())

			Ω(output).Should(ContainSubstring("--- PASS: TestAlwaysTrue"))
			Ω(output).Should(ContainSubstring("Test Suite Passed"))

			Ω(fm.PathTo("xunit", "block-profile.out")).Should(BeAnExistingFile())
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

	Context("when pointed at a package with tests, but the tests have been excluded via go tags", func() {
		BeforeEach(func() {
			fm.MountFixture("no_tagged_tests")
		})

		It("should exit 0, not run anything, and generate a json report if asked", func() {
			session := startGinkgo(fm.PathTo("no_tagged_tests"), "--no-color", "--json-report=report.json")
			Eventually(session).Should(gexec.Exit(0))
			Ω(session).Should(gbytes.Say("no test files"))

			report := fm.LoadJSONReports("no_tagged_tests", "report.json")[0]
			Ω(report.SuiteSucceeded).Should(BeTrue())
			Ω(report.SpecialSuiteFailureReasons).Should(ConsistOf("Suite did not run go test reported that no test files were found"))
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

		Context("with a specific number of -procs", func() {
			It("should use the specified number of processes", func() {
				session := startGinkgo(fm.PathTo("passing_ginkgo_tests"), "--no-color", "-succinct", "--procs=2")
				Eventually(session).Should(gexec.Exit(0))
				output := string(session.Out.Contents())

				Ω(output).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 4/4 specs - 2 procs [%s]{4} SUCCESS! \d+(\.\d+)?[muµ]s`, regexp.QuoteMeta(denoter)))
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
				Ω(output).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 4/4 specs - %d procs [%s]{4} SUCCESS! \d+(\.\d+)?[muµ]?s`, nodes, regexp.QuoteMeta(denoter)))
				Ω(output).Should(ContainSubstring("Test Suite Passed"))
			})
		})
	})

	Context("when running in parallel and there are specs marked Serial", Label("slow"), func() {
		BeforeEach(func() {
			fm.MountFixture("serial")
		})

		It("runs the serial specs after all the parallel specs have finished", func() {
			By("running a carefully crafted test without the serial decorator")
			session := startGinkgo(fm.PathTo("serial"), "--no-color", "--procs=2", "--randomize-all", "--fail-fast", "--", "--no-serial")
			Eventually(session).Should(gexec.Exit(1))

			By("running a carefully crafted test with the serial decorator")
			session = startGinkgo(fm.PathTo("serial"), "--no-color", "--procs=2", "--randomize-all", "--fail-fast")
			Eventually(session).Should(gexec.Exit(0))
		})
	})

	Context("when running with ordered specs", func() {
		BeforeEach(func() {
			fm.MountFixture("ordered")
		})

		It("always preserve spec order within ordered contexts", func() {
			By("running a carefully crafted test without the ordered decorator")
			session := startGinkgo(fm.PathTo("ordered"), "--no-color", "--procs=2", "-v", "--randomize-all", "--fail-fast", "--", "--no-ordered")
			Eventually(session).Should(gexec.Exit(1))

			By("running a carefully crafted test with the ordered decorator")
			session = startGinkgo(fm.PathTo("ordered"), "--no-color", "--procs=2", "-v", "--randomize-all", "--fail-fast")
			Eventually(session).Should(gexec.Exit(0))
		})
	})

	Context("when running multiple tests", func() {
		BeforeEach(func() {
			fm.MountFixture("passing_ginkgo_tests")
			fm.MountFixture("more_ginkgo_tests")
			fm.MountFixture("no_tagged_tests")
		})

		Context("when all the tests pass", func() {
			Context("with the -r flag", func() {
				It("should run all the tests (in succinct mode) and succeed", func() {
					session := startGinkgo(fm.TmpDir, "--no-color", "-r", ".")
					Eventually(session).Should(gexec.Exit(0))
					output := string(session.Out.Contents())

					outputLines := strings.Split(output, "\n")
					Ω(outputLines[0]).Should(MatchRegexp(`\[\d+\] More_ginkgo_tests Suite - 2/2 specs [%s]{2} SUCCESS! \d+(\.\d+)?[muµ]s PASS`, regexp.QuoteMeta(denoter)))
					Ω(outputLines[1]).Should(ContainSubstring("Skipping ./no_tagged_tests (no test files)"))
					Ω(outputLines[2]).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 4/4 specs [%s]{4} SUCCESS! \d+(\.\d+)?[muµ]s PASS`, regexp.QuoteMeta(denoter)))
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
					Ω(outputLines[1]).Should(ContainSubstring("Skipping ./no_tagged_tests (no test files)"))
					Ω(outputLines[2]).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 4/4 specs [%s]{4} SUCCESS! \d+(\.\d+)?[muµ]s PASS`, regexp.QuoteMeta(denoter)))
					Ω(output).Should(ContainSubstring("Test Suite Passed"))
				})
			})
		})

		Context("when one of the packages has a failing tests", func() {
			BeforeEach(func() {
				fm.MountFixture("failing_ginkgo_tests")
			})

			It("should fail and stop running tests", func() {
				session := startGinkgo(fm.TmpDir, "--no-color", "no_tagged_tests", "passing_ginkgo_tests", "failing_ginkgo_tests", "more_ginkgo_tests")
				Eventually(session).Should(gexec.Exit(1))
				output := string(session.Out.Contents())

				outputLines := strings.Split(output, "\n")
				Ω(outputLines[0]).Should(ContainSubstring("Skipping ./no_tagged_tests (no test files)"))
				Ω(outputLines[1]).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 4/4 specs [%s]{4} SUCCESS! \d+(\.\d+)?[muµ]s PASS`, regexp.QuoteMeta(denoter)))
				Ω(outputLines[2]).Should(MatchRegexp(`\[\d+\] Failing_ginkgo_tests Suite - 2/2 specs`))
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
				session := startGinkgo(fm.TmpDir, "--no-color", "no_tagged_tests", "passing_ginkgo_tests", "does_not_compile", "more_ginkgo_tests")
				Eventually(session).Should(gexec.Exit(1))
				output := string(session.Out.Contents())

				outputLines := strings.Split(output, "\n")
				Ω(outputLines[0]).Should(ContainSubstring("Skipping ./no_tagged_tests (no test files)"))
				Ω(outputLines[1]).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 4/4 specs [%s]{4} SUCCESS! \d+(\.\d+)?[muµ]s PASS`, regexp.QuoteMeta(denoter)))
				Ω(outputLines[2]).Should(ContainSubstring("Failed to compile does_not_compile:"))
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
				session := startGinkgo(fm.TmpDir, "--no-color", "-keep-going", "no_tagged_tests", "passing_ginkgo_tests", "does_not_compile", "failing_ginkgo_tests", "more_ginkgo_tests")
				Eventually(session).Should(gexec.Exit(1))
				output := string(session.Out.Contents())

				outputLines := strings.Split(output, "\n")
				Ω(outputLines[0]).Should(ContainSubstring("Skipping ./no_tagged_tests (no test files)"))
				Ω(outputLines[1]).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 4/4 specs [%s]{4} SUCCESS! \d+(\.\d+)?[muµ]s PASS`, regexp.QuoteMeta(denoter)))
				Ω(outputLines[2]).Should(ContainSubstring("Failed to compile does_not_compile:"))
				Ω(output).Should(MatchRegexp(`\[\d+\] Failing_ginkgo_tests Suite - 2/2 specs`))
				Ω(output).Should(ContainSubstring(fmt.Sprintf("%s [FAILED]", denoter)))
				Ω(output).Should(MatchRegexp(`\[\d+\] More_ginkgo_tests Suite - 2/2 specs [%s]{2} SUCCESS! \d+(\.\d+)?[muµ]s PASS`, regexp.QuoteMeta(denoter)))
				Ω(output).Should(ContainSubstring("Test Suite Failed"))
			})
		})
	})

	Context("when running large suites in parallel", Label("slow"), func() {
		BeforeEach(func() {
			fm.MountFixture("large")
		})

		It("doesn't miss any tests (a sanity test)", func() {
			session := startGinkgo(fm.PathTo("large"), "--no-color", "--procs=3", "--json-report=report.json")
			Eventually(session).Should(gexec.Exit(0))
			report := Reports(fm.LoadJSONReports("large", "report.json")[0].SpecReports)

			expectedNames := []string{}
			for i := 0; i < 2048; i++ {
				expectedNames = append(expectedNames, fmt.Sprintf("%d", i))
			}
			Ω(report.Names()).Should(ConsistOf(expectedNames))

		})
	})
})

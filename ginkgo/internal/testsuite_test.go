package internal_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2/ginkgo/internal"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

func TS(path string, pkgName string, isGinkgo bool, state TestSuiteState) TestSuite {
	return TestSuite{
		Path:        path,
		PackageName: pkgName,
		IsGinkgo:    isGinkgo,
		Precompiled: false,
		State:       state,
	}
}

func PTS(path string, pkgName string, isGinkgo bool, pathToCompiledTest string, state TestSuiteState) TestSuite {
	return TestSuite{
		Path:               path,
		PackageName:        pkgName,
		IsGinkgo:           isGinkgo,
		Precompiled:        true,
		PathToCompiledTest: pathToCompiledTest,
		State:              state,
	}
}

var _ = Describe("TestSuite", func() {
	Describe("Finding Suites", func() {
		var tmpDir string
		var origWd string
		var cliConf types.CLIConfig

		writeFile := func(folder string, filename string, content string, mode os.FileMode) {
			path := filepath.Join(tmpDir, folder)
			err := os.MkdirAll(path, 0700)
			Ω(err).ShouldNot(HaveOccurred())

			path = filepath.Join(path, filename)
			os.WriteFile(path, []byte(content), mode)
		}

		BeforeEach(func() {
			cliConf = types.CLIConfig{}

			var err error
			tmpDir, err = os.MkdirTemp("/tmp", "ginkgo")
			Ω(err).ShouldNot(HaveOccurred())

			origWd, err = os.Getwd()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(os.Chdir(tmpDir)).Should(Succeed())

			//go files in the root directory (no tests)
			writeFile("/", "main.go", "package main", 0666)

			//non-go files in a nested directory
			writeFile("/redherring", "big_test.jpg", "package ginkgo", 0666)

			//ginkgo tests in ignored go files
			writeFile("/ignored", ".ignore_dot_test.go", `import "github.com/onsi/ginkgo/v2"`, 0666)
			writeFile("/ignored", "_ignore_underscore_test.go", `import "github.com/onsi/ginkgo/v2"`, 0666)

			//non-ginkgo tests in a nested directory
			writeFile("/professorplum", "professorplum_test.go", `import "testing"`, 0666)

			//ginkgo tests in a nested directory
			writeFile("/colonelmustard", "colonelmustard_test.go", `import "github.com/onsi/ginkgo/v2"`, 0666)

			//ginkgo tests in a deeply nested directory
			writeFile("/colonelmustard/library", "library_test.go", `import "github.com/onsi/ginkgo/v2"`, 0666)

			//ginkgo tests in a deeply nested directory
			writeFile("/colonelmustard/library/spanner", "spanner_test.go", `import "github.com/onsi/ginkgo/v2/dsl/core"`, 0666)

			//ginkgo tests deeply nested in a vendored dependency
			writeFile("/vendor/mrspeacock/lounge", "lounge_test.go", `import "github.com/onsi/ginkgo/v2"`, 0666)

			//a precompiled ginkgo test
			writeFile("/precompiled-dir", "precompiled.test", `fake-binary-file`, 0777)
			writeFile("/precompiled-dir", "some-other-binary", `fake-binary-file`, 0777)
			writeFile("/precompiled-dir", "windows.test.exe", `fake-binary-file`, 0666)
			writeFile("/precompiled-dir", "windows.exe", `fake-binary-file`, 0666)
			writeFile("/precompiled-dir", "nonexecutable.test", `fake-binary-file`, 0666)
		})

		AfterEach(func() {
			Ω(os.Chdir(origWd)).Should(Succeed())
			os.RemoveAll(tmpDir)
		})

		Context("when passed no args", func() {
			Context("when told to recurse", func() {
				BeforeEach(func() {
					cliConf.Recurse = true
				})

				It("recurses through the current directory, returning all identified tests and skipping vendored, ignored, and precompiled tests", func() {
					suites := FindSuites([]string{}, cliConf, false)
					Ω(suites).Should(ConsistOf(
						TS("./professorplum", "professorplum", false, TestSuiteStateUncompiled),
						TS("./colonelmustard", "colonelmustard", true, TestSuiteStateUncompiled),
						TS("./colonelmustard/library", "library", true, TestSuiteStateUncompiled),
						TS("./colonelmustard/library/spanner", "spanner", true, TestSuiteStateUncompiled),
					))
				})
			})

			Context("when told to recurse and there is a skip-package filter", func() {
				BeforeEach(func() {
					cliConf.Recurse = true
					cliConf.SkipPackage = "professorplum,library,floop"
				})

				It("recurses through the current directory, returning all identified tests and skipping vendored, ignored, and precompiled tests", func() {
					suites := FindSuites([]string{}, cliConf, false)
					Ω(suites).Should(ConsistOf(
						TS("./professorplum", "professorplum", false, TestSuiteStateSkippedByFilter),
						TS("./colonelmustard", "colonelmustard", true, TestSuiteStateUncompiled),
						TS("./colonelmustard/library", "library", true, TestSuiteStateSkippedByFilter),
						TS("./colonelmustard/library/spanner", "spanner", true, TestSuiteStateSkippedByFilter),
					))
				})
			})

			Context("when there are no tests in the current directory", func() {
				BeforeEach(func() {
					cliConf.Recurse = false
				})

				It("returns empty", func() {
					suites := FindSuites([]string{}, cliConf, false)
					Ω(suites).Should(BeEmpty())
				})
			})

			Context("when told not to recurse", func() {
				BeforeEach(func() {
					Ω(os.Chdir("./colonelmustard")).Should(Succeed())
				})

				It("returns tests in the current directory if present", func() {
					suites := FindSuites([]string{}, cliConf, false)
					Ω(suites).Should(ConsistOf(
						TS(".", "colonelmustard", true, TestSuiteStateUncompiled),
					))
				})
			})
		})

		Context("when passed args", func() {
			Context("when told to recurse", func() {
				BeforeEach(func() {
					cliConf.Recurse = true
				})

				It("recurses through the passed-in directories, returning all identified tests and skipping vendored, ignored, and precompiled tests", func() {
					suites := FindSuites([]string{"precompiled-dir", "colonelmustard"}, cliConf, false)
					Ω(suites).Should(ConsistOf(
						TS("./colonelmustard", "colonelmustard", true, TestSuiteStateUncompiled),
						TS("./colonelmustard/library", "library", true, TestSuiteStateUncompiled),
						TS("./colonelmustard/library/spanner", "spanner", true, TestSuiteStateUncompiled),
					))
				})
			})

			Context("when told to recurse and there is a skip-package filter", func() {
				BeforeEach(func() {
					cliConf.Recurse = true
					cliConf.SkipPackage = "library"
				})

				It("recurses through the passed-in directories, returning all identified tests and skipping vendored, ignored, and precompiled tests", func() {
					suites := FindSuites([]string{"precompiled-dir", "professorplum", "colonelmustard"}, cliConf, false)
					Ω(suites).Should(ConsistOf(
						TS("./professorplum", "professorplum", false, TestSuiteStateUncompiled),
						TS("./colonelmustard", "colonelmustard", true, TestSuiteStateUncompiled),
						TS("./colonelmustard/library", "library", true, TestSuiteStateSkippedByFilter),
						TS("./colonelmustard/library/spanner", "spanner", true, TestSuiteStateSkippedByFilter),
					))
				})
			})

			Context("when told not to recurse", func() {
				BeforeEach(func() {
					cliConf.Recurse = false
				})

				It("returns test packages at the passed in arguments", func() {
					suites := FindSuites([]string{"precompiled-dir", "colonelmustard", "professorplum", "ignored"}, cliConf, false)
					Ω(suites).Should(ConsistOf(
						TS("./professorplum", "professorplum", false, TestSuiteStateUncompiled),
						TS("./colonelmustard", "colonelmustard", true, TestSuiteStateUncompiled),
					))
				})
			})

			Context("when told not to recurse, but an arg has /...", func() {
				BeforeEach(func() {
					cliConf.Recurse = false
				})

				It("recurses through the directories it is told to recurse through, returning all identified tests and skipping vendored, ignored, and precompiled tests", func() {
					suites := FindSuites([]string{"precompiled-dir", "colonelmustard/...", "professorplum/...", "ignored/..."}, cliConf, false)
					Ω(suites).Should(ConsistOf(
						TS("./professorplum", "professorplum", false, TestSuiteStateUncompiled),
						TS("./colonelmustard", "colonelmustard", true, TestSuiteStateUncompiled),
						TS("./colonelmustard/library", "library", true, TestSuiteStateUncompiled),
						TS("./colonelmustard/library/spanner", "spanner", true, TestSuiteStateUncompiled),
					))
				})
			})

			Context("when told not to recurse and there is a skip-package filter", func() {
				BeforeEach(func() {
					cliConf.Recurse = false
					cliConf.SkipPackage = "library,plum"
				})

				It("returns skips packages that match", func() {
					suites := FindSuites([]string{"colonelmustard", "professorplum", "colonelmustard/library"}, cliConf, false)
					Ω(suites).Should(ConsistOf(
						TS("./professorplum", "professorplum", false, TestSuiteStateSkippedByFilter),
						TS("./colonelmustard", "colonelmustard", true, TestSuiteStateUncompiled),
						TS("./colonelmustard/library", "library", true, TestSuiteStateSkippedByFilter),
					))
				})
			})

			Context("when pointed at a directory containing a precompiled test suite", func() {
				It("returns nothing", func() {
					suites := FindSuites([]string{"precompiled-dir"}, cliConf, false)
					Ω(suites).Should(BeEmpty())
				})
			})

			Context("when pointed at a precompiled test suite specifically", func() {
				It("returns the precompiled suite", func() {
					path, err := filepath.Abs("./precompiled-dir/precompiled.test")
					Ω(err).ShouldNot(HaveOccurred())
					suites := FindSuites([]string{"precompiled-dir/precompiled.test"}, cliConf, true)
					Ω(suites).Should(ConsistOf(
						PTS("./precompiled-dir", "precompiled", true, path, TestSuiteStateCompiled),
					))
				})
			})

			Context("when pointed at a precompiled test suite on windows", func() {
				It("returns the precompiled suite", func() {
					path, err := filepath.Abs("./precompiled-dir/windows.exe")
					Ω(err).ShouldNot(HaveOccurred())
					suites := FindSuites([]string{"precompiled-dir/windows.exe"}, cliConf, true)
					Ω(suites).Should(ConsistOf(
						PTS("./precompiled-dir", "windows", true, path, TestSuiteStateCompiled),
					))

					path, err = filepath.Abs("./precompiled-dir/windows.test.exe")
					Ω(err).ShouldNot(HaveOccurred())
					suites = FindSuites([]string{"precompiled-dir/windows.test.exe"}, cliConf, true)
					Ω(suites).Should(ConsistOf(
						PTS("./precompiled-dir", "windows", true, path, TestSuiteStateCompiled),
					))
				})
			})

			Context("when pointed at a fake precompiled test", func() {
				It("returns nothing", func() {
					suites := FindSuites([]string{"precompiled-dir/some-other-binary"}, cliConf, true)
					Ω(suites).Should(BeEmpty())

					suites = FindSuites([]string{"precompiled-dir/nonexecutable.test"}, cliConf, true)
					Ω(suites).Should(BeEmpty())
				})
			})

			Context("when pointed at a precompiled test suite specifically but allowPrecompiled is false", func() {
				It("returns nothing", func() {
					suites := FindSuites([]string{"precompiled-dir/some-other-binary"}, cliConf, false)
					Ω(suites).Should(BeEmpty())
				})
			})
		})
	})

	Describe("NamespacedName", func() {
		It("generates a name basd on the relative path to the package", func() {
			plum := TS("./professorplum", "professorplum", false, TestSuiteStateUncompiled)
			library := TS("./colonelmustard/library", "library", true, TestSuiteStateUncompiled)
			root := TS(".", "root", true, TestSuiteStateUncompiled)

			Ω(plum.NamespacedName()).Should(Equal("professorplum"))
			Ω(library.NamespacedName()).Should(Equal("colonelmustard_library"))
			Ω(root.NamespacedName()).Should(Equal("root"))
		})
	})

	Describe("TestSuiteState", func() {
		Describe("Is", func() {
			It("returns true if it matches one of the passed in states", func() {
				Ω(TestSuiteStateCompiled.Is(TestSuiteStateUncompiled, TestSuiteStateCompiled)).Should(BeTrue())
				Ω(TestSuiteStateCompiled.Is(TestSuiteStateUncompiled, TestSuiteStatePassed)).Should(BeFalse())
			})
		})

		Describe("TestSuiteStateFailureStates", func() {
			It("should enumerate the failure states", func() {
				Ω(TestSuiteStateFailureStates).Should(ConsistOf(
					TestSuiteStateFailed,
					TestSuiteStateFailedDueToTimeout,
					TestSuiteStateFailedToCompile,
				))
			})
		})
	})

	Describe("TestSuites", func() {
		var A, B, C, D TestSuite
		var suites TestSuites
		BeforeEach(func() {
			A = TS("/A", "A", true, TestSuiteStateUncompiled)
			B = TS("/B", "B", true, TestSuiteStateUncompiled)
			C = TS("/C", "C", true, TestSuiteStateUncompiled)
			D = TS("/D", "D", true, TestSuiteStateUncompiled)
		})

		JustBeforeEach(func() {
			suites = TestSuites{A, B, C, D}
		})

		Describe("AnyHaveProgrammaticFocus", func() {
			Context("when any suites have programmatic focus", func() {
				BeforeEach(func() {
					B.HasProgrammaticFocus = true
				})

				It("returns true", func() {
					Ω(suites.AnyHaveProgrammaticFocus()).Should(BeTrue())
				})
			})
			Context("when any suites do not have programmatic focus", func() {
				It("returns false", func() {
					Ω(suites.AnyHaveProgrammaticFocus()).Should(BeFalse())
				})
			})
		})

		Describe("ThatAreGinkgoSuites", func() {
			BeforeEach(func() {
				B.IsGinkgo = false
				D.IsGinkgo = false
			})
			It("returns the subset that are Ginkgo suites", func() {
				Ω(suites.ThatAreGinkgoSuites()).Should(Equal(TestSuites{A, C}))
			})
		})

		Describe("CountWithState", func() {
			BeforeEach(func() {
				B.State = TestSuiteStateFailed
				D.State = TestSuiteStateFailedToCompile
			})

			It("returns the number with the matching state", func() {
				Ω(suites.CountWithState(TestSuiteStateFailed)).Should(Equal(1))
				Ω(suites.CountWithState(TestSuiteStateFailed, TestSuiteStateFailedToCompile)).Should(Equal(2))
			})
		})

		Describe("WithState", func() {
			BeforeEach(func() {
				A.State = TestSuiteStatePassed
				C.State = TestSuiteStateSkippedByFilter
			})

			It("returns the suites matching the passed-in states", func() {
				Ω(suites.WithState(TestSuiteStatePassed, TestSuiteStateSkippedByFilter)).Should(Equal(TestSuites{A, C}))
			})
		})

		Describe("WithoutState", func() {
			BeforeEach(func() {
				A.State = TestSuiteStatePassed
				C.State = TestSuiteStateSkippedByFilter
			})

			It("returns the suites _not_ matching the passed-in states", func() {
				Ω(suites.WithoutState(TestSuiteStatePassed, TestSuiteStateSkippedByFilter)).Should(Equal(TestSuites{B, D}))
			})
		})

		Describe("ShuffledCopy", func() {
			It("returns a shuffled copy of the test suites", func() {
				shuffled := suites.ShuffledCopy(17)
				Ω(suites).Should(Equal(TestSuites{A, B, C, D}))
				Ω(shuffled).Should(ConsistOf(A, B, C, D))
				Ω(shuffled).ShouldNot(Equal(suites))
			})
		})
	})
})

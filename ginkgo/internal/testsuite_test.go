package internal_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/ginkgo/ginkgo/internal"
	. "github.com/onsi/gomega"
)

func TS(path string, pkgName string, isGinkgo bool) TestSuite {
	return TestSuite{
		Path:        path,
		PackageName: pkgName,
		IsGinkgo:    isGinkgo,
		Precompiled: false,
	}
}

func PTS(path string, pkgName string, isGinkgo bool, pathToCompiledTest string) TestSuite {
	return TestSuite{
		Path:               path,
		PackageName:        pkgName,
		IsGinkgo:           isGinkgo,
		Precompiled:        true,
		PathToCompiledTest: pathToCompiledTest,
	}
}

var _ = Describe("TestSuite", func() {
	var tmpDir string
	var origWd string
	var cliConf config.GinkgoCLIConfigType

	writeFile := func(folder string, filename string, content string, mode os.FileMode) {
		path := filepath.Join(tmpDir, folder)
		err := os.MkdirAll(path, 0700)
		Ω(err).ShouldNot(HaveOccurred())

		path = filepath.Join(path, filename)
		ioutil.WriteFile(path, []byte(content), mode)
	}

	BeforeEach(func() {
		cliConf = config.GinkgoCLIConfigType{}

		var err error
		tmpDir, err = ioutil.TempDir("/tmp", "ginkgo")
		Ω(err).ShouldNot(HaveOccurred())

		origWd, err = os.Getwd()
		Ω(err).ShouldNot(HaveOccurred())
		Ω(os.Chdir(tmpDir)).Should(Succeed())

		//go files in the root directory (no tests)
		writeFile("/", "main.go", "package main", 0666)

		//non-go files in a nested directory
		writeFile("/redherring", "big_test.jpg", "package ginkgo", 0666)

		//ginkgo tests in ignored go files
		writeFile("/ignored", ".ignore_dot_test.go", `import "github.com/onsi/ginkgo"`, 0666)
		writeFile("/ignored", "_ignore_underscore_test.go", `import "github.com/onsi/ginkgo"`, 0666)

		//non-ginkgo tests in a nested directory
		writeFile("/professorplum", "professorplum_test.go", `import "testing"`, 0666)

		//ginkgo tests in a nested directory
		writeFile("/colonelmustard", "colonelmustard_test.go", `import "github.com/onsi/ginkgo"`, 0666)

		//ginkgo tests in a deeply nested directory
		writeFile("/colonelmustard/library", "library_test.go", `import "github.com/onsi/ginkgo"`, 0666)

		//ginkgo tests deeply nested in a vendored dependency
		writeFile("/vendor/mrspeacock/lounge", "lounge_test.go", `import "github.com/onsi/ginkgo"`, 0666)

		//a precompiled ginkgo test
		writeFile("/precompiled-dir", "precompiled.test", `fake-binary-file`, 0777)
		writeFile("/precompiled-dir", "some-other-binary", `fake-binary-file`, 0777)
		writeFile("/precompiled-dir", "nonexecutable.test", `fake-binary-file`, 0666)
	})

	AfterEach(func() {
		Ω(os.Chdir(origWd)).Should(Succeed())
		os.RemoveAll(tmpDir)
	})

	Describe("Finding Suites", func() {
		Context("when passed no args", func() {
			Context("when told to recurse", func() {
				BeforeEach(func() {
					cliConf.Recurse = true
				})

				It("recurses through the current directory, returning all identified tests and skipping vendored, ignored, and precompiled tests", func() {
					suites, skipped := FindSuites([]string{}, cliConf, false)
					Ω(skipped).Should(BeEmpty())
					Ω(suites).Should(ConsistOf(
						TS("./professorplum", "professorplum", false),
						TS("./colonelmustard", "colonelmustard", true),
						TS("./colonelmustard/library", "library", true),
					))
				})
			})

			Context("when told to recurse and there is a skip-package filter", func() {
				BeforeEach(func() {
					cliConf.Recurse = true
					cliConf.SkipPackage = "professorplum,library,floop"
				})

				It("recurses through the current directory, returning all identified tests and skipping vendored, ignored, and precompiled tests", func() {
					suites, skipped := FindSuites([]string{}, cliConf, false)
					Ω(skipped).Should(ConsistOf(
						"./professorplum",
						"./colonelmustard/library",
					))
					Ω(suites).Should(ConsistOf(
						TS("./colonelmustard", "colonelmustard", true),
					))
				})
			})

			Context("when there are no tests in the current directory", func() {
				BeforeEach(func() {
					cliConf.Recurse = false
				})

				It("returns empty", func() {
					suites, skipped := FindSuites([]string{}, cliConf, false)
					Ω(suites).Should(BeEmpty())
					Ω(skipped).Should(BeEmpty())
				})
			})

			Context("when told not to recurse", func() {
				BeforeEach(func() {
					Ω(os.Chdir("./colonelmustard")).Should(Succeed())
				})

				It("returns tests in the current directory if present", func() {
					suites, skipped := FindSuites([]string{}, cliConf, false)
					Ω(skipped).Should(BeEmpty())
					Ω(suites).Should(ConsistOf(
						TS(".", "colonelmustard", true),
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
					suites, skipped := FindSuites([]string{"precompiled-dir", "colonelmustard"}, cliConf, false)
					Ω(skipped).Should(BeEmpty())
					Ω(suites).Should(ConsistOf(
						TS("./colonelmustard", "colonelmustard", true),
						TS("./colonelmustard/library", "library", true),
					))
				})
			})

			Context("when told to recurse and there is a skip-package filter", func() {
				BeforeEach(func() {
					cliConf.Recurse = true
					cliConf.SkipPackage = "library"
				})

				It("recurses through the passed-in directories, returning all identified tests and skipping vendored, ignored, and precompiled tests", func() {
					suites, skipped := FindSuites([]string{"precompiled-dir", "professorplum", "colonelmustard"}, cliConf, false)
					Ω(skipped).Should(ConsistOf(
						"./colonelmustard/library",
					))
					Ω(suites).Should(ConsistOf(
						TS("./professorplum", "professorplum", false),
						TS("./colonelmustard", "colonelmustard", true),
					))
				})
			})

			Context("when told not to recurse", func() {
				BeforeEach(func() {
					cliConf.Recurse = false
				})

				It("returns test packages at the passed in arguments", func() {
					suites, skipped := FindSuites([]string{"precompiled-dir", "colonelmustard", "professorplum", "ignored"}, cliConf, false)
					Ω(skipped).Should(BeEmpty())
					Ω(suites).Should(ConsistOf(
						TS("./professorplum", "professorplum", false),
						TS("./colonelmustard", "colonelmustard", true),
					))
				})
			})

			Context("when told not to recurse, but an arg has /...", func() {
				BeforeEach(func() {
					cliConf.Recurse = false
				})

				It("recurses through the directories it is told to recurse through, returning all identified tests and skipping vendored, ignored, and precompiled tests", func() {
					suites, skipped := FindSuites([]string{"precompiled-dir", "colonelmustard/...", "professorplum/...", "ignored/..."}, cliConf, false)
					Ω(skipped).Should(BeEmpty())
					Ω(suites).Should(ConsistOf(
						TS("./professorplum", "professorplum", false),
						TS("./colonelmustard", "colonelmustard", true),
						TS("./colonelmustard/library", "library", true),
					))
				})
			})

			Context("when told not to recurse and there is a skip-package filter", func() {
				BeforeEach(func() {
					cliConf.Recurse = false
					cliConf.SkipPackage = "library,plum"
				})

				It("returns skips packages that match", func() {
					suites, skipped := FindSuites([]string{"colonelmustard", "professorplum", "colonelmustard/library"}, cliConf, false)
					Ω(skipped).Should(ConsistOf(
						"./professorplum",
						"./colonelmustard/library",
					))
					Ω(suites).Should(ConsistOf(
						TS("./colonelmustard", "colonelmustard", true),
					))
				})
			})

			Context("when pointed at a directory containing a precompiled test suite", func() {
				It("returns nothing", func() {
					suites, skipped := FindSuites([]string{"precompiled-dir"}, cliConf, false)
					Ω(skipped).Should(BeEmpty())
					Ω(suites).Should(BeEmpty())
				})
			})

			Context("when pointed at a precompiled test suite specifically", func() {
				It("returns the precompiled suite", func() {
					path, err := filepath.Abs("./precompiled-dir/precompiled.test")
					Ω(err).ShouldNot(HaveOccurred())
					suites, skipped := FindSuites([]string{"precompiled-dir/precompiled.test"}, cliConf, true)
					Ω(skipped).Should(BeEmpty())
					Ω(suites).Should(ConsistOf(
						PTS("./precompiled-dir", "precompiled", true, path),
					))
				})
			})

			Context("when pointed at a fake precompiled test", func() {
				It("returns nothing", func() {
					suites, skipped := FindSuites([]string{"precompiled-dir/some-other-binary"}, cliConf, true)
					Ω(skipped).Should(BeEmpty())
					Ω(suites).Should(BeEmpty())

					suites, skipped = FindSuites([]string{"precompiled-dir/nonexecutable.test"}, cliConf, true)
					Ω(skipped).Should(BeEmpty())
					Ω(suites).Should(BeEmpty())
				})
			})

			Context("when pointed at a precompiled test suite specifically but allowPrecompiled is false", func() {
				It("returns nothing", func() {
					suites, skipped := FindSuites([]string{"precompiled-dir/some-other-binary"}, cliConf, false)
					Ω(skipped).Should(BeEmpty())
					Ω(suites).Should(BeEmpty())
				})
			})
		})

		Describe("NamespacedName", func() {
			It("generates a name basd on the relative path to the package", func() {
				plum := TS("./professorplum", "professorplum", false)
				library := TS("./colonelmustard/library", "library", true)
				root := TS(".", "root", true)

				Ω(plum.NamespacedName()).Should(Equal("professorplum"))
				Ω(library.NamespacedName()).Should(Equal("colonelmustard_library"))
				Ω(root.NamespacedName()).Should(Equal("root"))
			})
		})
	})
})

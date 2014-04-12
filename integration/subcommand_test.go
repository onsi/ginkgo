package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var _ = Describe("Subcommand", func() {
	Describe("ginkgo bootstrap", func() {
		It("should generate a bootstrap file, as long as one does not exist", func() {
			pkgPath := tmpPath("foo")
			os.Mkdir(pkgPath, 0777)
			output, err := runGinkgo(pkgPath, "bootstrap")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(output).Should(ContainSubstring("foo_suite_test.go"))

			content, err := ioutil.ReadFile(filepath.Join(pkgPath, "foo_suite_test.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(content).Should(ContainSubstring("func TestFoo(t *testing.T) {"))
			Ω(content).Should(ContainSubstring("RegisterFailHandler"))
			Ω(content).Should(ContainSubstring("RunSpecs"))

			Ω(content).Should(ContainSubstring("\t" + `. "github.com/onsi/ginkgo"`))
			Ω(content).Should(ContainSubstring("\t" + `. "github.com/onsi/gomega"`))

			output, err = runGinkgo(pkgPath, "bootstrap")
			Ω(err).Should(HaveOccurred())
			Ω(output).Should(ContainSubstring("foo_suite_test.go already exists"))
		})

		It("should import nodot declarations when told to", func() {
			pkgPath := tmpPath("foo")
			os.Mkdir(pkgPath, 0777)
			output, err := runGinkgo(pkgPath, "bootstrap", "--nodot")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(output).Should(ContainSubstring("foo_suite_test.go"))

			content, err := ioutil.ReadFile(filepath.Join(pkgPath, "foo_suite_test.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(content).Should(ContainSubstring("func TestFoo(t *testing.T) {"))
			Ω(content).Should(ContainSubstring("RegisterFailHandler"))
			Ω(content).Should(ContainSubstring("RunSpecs"))

			Ω(content).Should(ContainSubstring("var It = ginkgo.It"))
			Ω(content).Should(ContainSubstring("var Ω = gomega.Ω"))

			Ω(content).Should(ContainSubstring("\t" + `"github.com/onsi/ginkgo"`))
			Ω(content).Should(ContainSubstring("\t" + `"github.com/onsi/gomega"`))
		})
	})

	Describe("nodot", func() {
		It("should update the declarations in the bootstrap file", func() {
			pkgPath := tmpPath("foo")
			os.Mkdir(pkgPath, 0777)

			_, err := runGinkgo(pkgPath, "bootstrap", "--nodot")
			Ω(err).ShouldNot(HaveOccurred())

			byteContent, err := ioutil.ReadFile(filepath.Join(pkgPath, "foo_suite_test.go"))
			Ω(err).ShouldNot(HaveOccurred())

			content := string(byteContent)
			content = strings.Replace(content, "var It =", "var MyIt =", -1)
			content = strings.Replace(content, "var Ω = gomega.Ω\n", "", -1)

			err = ioutil.WriteFile(filepath.Join(pkgPath, "foo_suite_test.go"), []byte(content), os.ModePerm)
			Ω(err).ShouldNot(HaveOccurred())

			_, err = runGinkgo(pkgPath, "nodot")
			Ω(err).ShouldNot(HaveOccurred())

			byteContent, err = ioutil.ReadFile(filepath.Join(pkgPath, "foo_suite_test.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(byteContent).Should(ContainSubstring("var MyIt = ginkgo.It"))
			Ω(byteContent).ShouldNot(ContainSubstring("var It = ginkgo.It"))
			Ω(byteContent).Should(ContainSubstring("var Ω = gomega.Ω"))
		})
	})

	Describe("ginkgo generate", func() {
		var pkgPath string

		BeforeEach(func() {
			pkgPath = tmpPath("foo_bar")
			os.Mkdir(pkgPath, 0777)
		})

		Context("with no arguments", func() {
			It("should generate a test file named after the package", func() {
				output, err := runGinkgo(pkgPath, "generate")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(output).Should(ContainSubstring("foo_bar_test.go"))

				content, err := ioutil.ReadFile(filepath.Join(pkgPath, "foo_bar_test.go"))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(content).Should(ContainSubstring(`var _ = Describe("FooBar", func() {`))
				Ω(content).Should(ContainSubstring("\t" + `. "github.com/onsi/ginkgo"`))
				Ω(content).Should(ContainSubstring("\t" + `. "github.com/onsi/gomega"`))

				output, err = runGinkgo(pkgPath, "generate")
				Ω(err).Should(HaveOccurred())
				Ω(output).Should(ContainSubstring("foo_bar_test.go already exists"))
			})
		})

		Context("with an argument of the form: foo", func() {
			It("should generate a test file named after the argument", func() {
				output, err := runGinkgo(pkgPath, "generate", "baz_buzz")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(output).Should(ContainSubstring("baz_buzz_test.go"))

				content, err := ioutil.ReadFile(filepath.Join(pkgPath, "baz_buzz_test.go"))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(content).Should(ContainSubstring(`var _ = Describe("BazBuzz", func() {`))
			})
		})

		Context("with an argument of the form: foo.go", func() {
			It("should generate a test file named after the argument", func() {
				output, err := runGinkgo(pkgPath, "generate", "baz_buzz.go")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(output).Should(ContainSubstring("baz_buzz_test.go"))

				content, err := ioutil.ReadFile(filepath.Join(pkgPath, "baz_buzz_test.go"))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(content).Should(ContainSubstring(`var _ = Describe("BazBuzz", func() {`))

			})
		})

		Context("with an argument of the form: foo_test", func() {
			It("should generate a test file named after the argument", func() {
				output, err := runGinkgo(pkgPath, "generate", "baz_buzz_test")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(output).Should(ContainSubstring("baz_buzz_test.go"))

				content, err := ioutil.ReadFile(filepath.Join(pkgPath, "baz_buzz_test.go"))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(content).Should(ContainSubstring(`var _ = Describe("BazBuzz", func() {`))
			})
		})

		Context("with an argument of the form: foo_test.go", func() {
			It("should generate a test file named after the argument", func() {
				output, err := runGinkgo(pkgPath, "generate", "baz_buzz_test.go")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(output).Should(ContainSubstring("baz_buzz_test.go"))

				content, err := ioutil.ReadFile(filepath.Join(pkgPath, "baz_buzz_test.go"))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(content).Should(ContainSubstring(`var _ = Describe("BazBuzz", func() {`))
			})
		})

		Context("with nodot", func() {
			It("should not import ginkgo or gomega", func() {
				output, err := runGinkgo(pkgPath, "generate", "--nodot")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(output).Should(ContainSubstring("foo_bar_test.go"))

				content, err := ioutil.ReadFile(filepath.Join(pkgPath, "foo_bar_test.go"))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(content).ShouldNot(ContainSubstring("\t" + `. "github.com/onsi/ginkgo"`))
				Ω(content).ShouldNot(ContainSubstring("\t" + `. "github.com/onsi/gomega"`))
			})
		})
	})

	Describe("ginkgo blur", func() {
		It("should unfocus tests", func() {
			pathToTest := tmpPath("focused")
			copyIn("focused_fixture", pathToTest)

			output, err := runGinkgo(pathToTest, "--noColor")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(output).Should(ContainSubstring("3 Passed"))
			Ω(output).Should(ContainSubstring("3 Skipped"))

			output, err = runGinkgo(pathToTest, "blur")
			Ω(err).ShouldNot(HaveOccurred())

			output, err = runGinkgo(pathToTest, "--noColor")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(output).Should(ContainSubstring("6 Passed"))
			Ω(output).Should(ContainSubstring("0 Skipped"))
		})
	})

	Describe("ginkgo version", func() {
		It("should print out the version info", func() {
			output, err := runGinkgo("", "version")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(output).Should(MatchRegexp(`Ginkgo Version \d+\.\d+\.\d+`))
		})
	})

	Describe("ginkgo help", func() {
		It("should print out usage information", func() {
			output, err := runGinkgo("", "help")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(output).Should(MatchRegexp(`Ginkgo Version \d+\.\d+\.\d+`))
			Ω(output).Should(ContainSubstring("ginkgo watch"))
			Ω(output).Should(ContainSubstring("-succinct"))
			Ω(output).Should(ContainSubstring("-nodes"))
			Ω(output).Should(ContainSubstring("ginkgo generate"))
		})
	})
})

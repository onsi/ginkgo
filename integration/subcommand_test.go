package integration_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Subcommand", func() {
	Describe("ginkgo bootstrap", func() {
		var pkg string

		BeforeEach(func() {
			pkg = "foo"
			fm.MkEmpty(pkg)
		})

		It("should generate a bootstrap file, as long as one does not exist", func() {
			session := startGinkgo(fm.PathTo(pkg), "bootstrap")
			Eventually(session).Should(gexec.Exit(0))
			output := session.Out.Contents()

			Ω(output).Should(ContainSubstring("foo_suite_test.go"))

			content := fm.ContentOf(pkg, "foo_suite_test.go")
			Ω(content).Should(ContainSubstring("package foo_test"))
			Ω(content).Should(ContainSubstring("func TestFoo(t *testing.T) {"))
			Ω(content).Should(ContainSubstring("RegisterFailHandler"))
			Ω(content).Should(ContainSubstring("RunSpecs"))

			Ω(content).Should(ContainSubstring("\t" + `. "github.com/onsi/ginkgo/v2"`))
			Ω(content).Should(ContainSubstring("\t" + `. "github.com/onsi/gomega"`))

			session = startGinkgo(fm.PathTo(pkg))
			Eventually(session).Should(gexec.Exit(0))

			session = startGinkgo(fm.PathTo(pkg), "bootstrap")
			Eventually(session).Should(gexec.Exit(1))
			output = session.Err.Contents()
			Ω(output).Should(ContainSubstring("foo_suite_test.go"))
			Ω(output).Should(ContainSubstring("already exists"))
		})

		It("should generate a bootstrap file with a working package name if the folder starts with a numeral", func() {
			fm.MkEmpty("7")
			session := startGinkgo(fm.PathTo("7"), "bootstrap")
			Eventually(session).Should(gexec.Exit(0))

			content := fm.ContentOf("7", "7_suite_test.go")
			pkg := strings.Split(content, "\n")[0]
			Ω(pkg).Should(Equal("package seven_test"))

			session = startGinkgo(fm.PathTo("7"))
			Eventually(session).Should(gexec.Exit(0))
		})

		It("should import nodot declarations when told to", func() {
			session := startGinkgo(fm.PathTo(pkg), "bootstrap", "--nodot")
			Eventually(session).Should(gexec.Exit(0))
			output := session.Out.Contents()

			Ω(output).Should(ContainSubstring("foo_suite_test.go"))

			content := fm.ContentOf(pkg, "foo_suite_test.go")
			Ω(content).Should(ContainSubstring("package foo_test"))
			Ω(content).Should(ContainSubstring("func TestFoo(t *testing.T) {"))
			Ω(content).Should(ContainSubstring("gomega.RegisterFailHandler"))
			Ω(content).Should(ContainSubstring("ginkgo.RunSpecs"))

			Ω(content).Should(ContainSubstring("\t" + `"github.com/onsi/ginkgo/v2"`))
			Ω(content).Should(ContainSubstring("\t" + `"github.com/onsi/gomega"`))

			session = startGinkgo(fm.PathTo(pkg))
			Eventually(session).Should(gexec.Exit(0))
		})

		It("should generate a bootstrap file using a template when told to", func() {
			fm.WriteFile(pkg, ".bootstrap", `package {{.Package}}

			import (
				{{.GinkgoImport}}
				{{.GomegaImport}}

				"testing"
				"binary"
			)

			func Test{{.FormattedName}}(t *testing.T) {
				// This is a {{.Package}} test
			}`)
			session := startGinkgo(fm.PathTo(pkg), "bootstrap", "--template", ".bootstrap")
			Eventually(session).Should(gexec.Exit(0))
			output := session.Out.Contents()

			Ω(output).Should(ContainSubstring("foo_suite_test.go"))

			content := fm.ContentOf(pkg, "foo_suite_test.go")
			Ω(content).Should(ContainSubstring("package foo_test"))
			Ω(content).Should(ContainSubstring(`. "github.com/onsi/ginkgo/v2"`))
			Ω(content).Should(ContainSubstring(`. "github.com/onsi/gomega"`))
			Ω(content).Should(ContainSubstring(`"binary"`))
			Ω(content).Should(ContainSubstring("// This is a foo_test test"))
		})

		It("should generate a bootstrap file using a template that contains functions when told to", func() {
			fm.WriteFile(pkg, ".bootstrap", `package {{.Package}}

			import (
				{{.GinkgoImport}}
				{{.GomegaImport}}

				"testing"
				"binary"
			)

			func Test{{.FormattedName}}(t *testing.T) {
				// This is a {{.Package | repeat 3}} test
			}`)
			session := startGinkgo(fm.PathTo(pkg), "bootstrap", "--template", ".bootstrap")
			Eventually(session).Should(gexec.Exit(0))
			output := session.Out.Contents()

			Ω(output).Should(ContainSubstring("foo_suite_test.go"))

			content := fm.ContentOf(pkg, "foo_suite_test.go")
			Ω(content).Should(ContainSubstring("package foo_test"))
			Ω(content).Should(ContainSubstring(`. "github.com/onsi/ginkgo/v2"`))
			Ω(content).Should(ContainSubstring(`. "github.com/onsi/gomega"`))
			Ω(content).Should(ContainSubstring(`"binary"`))
			Ω(content).Should(ContainSubstring("// This is a foo_testfoo_testfoo_test test"))
		})
	})

	Describe("ginkgo generate", func() {
		var pkg string

		BeforeEach(func() {
			pkg = "foo_bar"
			fm.MkEmpty(pkg)
			Eventually(startGinkgo(fm.PathTo(pkg), "bootstrap")).Should(gexec.Exit(0))
		})

		Context("with no arguments", func() {
			It("should generate a test file named after the package", func() {
				session := startGinkgo(fm.PathTo(pkg), "generate")
				Eventually(session).Should(gexec.Exit(0))
				output := session.Out.Contents()

				Ω(output).Should(ContainSubstring("foo_bar_test.go"))

				By("having the correct content")
				content := fm.ContentOf(pkg, "foo_bar_test.go")
				Ω(content).Should(ContainSubstring("package foo_bar_test"))
				Ω(content).Should(ContainSubstring(`var _ = Describe("FooBar", func() {`))
				Ω(content).Should(ContainSubstring("\t" + `. "github.com/onsi/ginkgo/v2"`))
				Ω(content).Should(ContainSubstring("\t" + `. "github.com/onsi/gomega"`))

				By("compiling correctly (we append to the file to make sure gomega is used)")
				fm.WriteFile(pkg, "foo_bar.go", "package foo_bar\nvar TRUE=true\n")
				fm.AppendToFile(pkg, "foo_bar_test.go", strings.Join([]string{``,
					`var _ = It("works", func() {`,
					`    Expect(foo_bar.TRUE).To(BeTrue())`,
					`})`,
				}, "\n"))
				Eventually(startGinkgo(fm.PathTo(pkg))).Should(gexec.Exit(0))

				By("refusing to overwrite the file if generate is called again")
				session = startGinkgo(fm.PathTo(pkg), "generate")
				Eventually(session).Should(gexec.Exit(1))
				output = session.Err.Contents()

				Ω(output).Should(ContainSubstring("foo_bar_test.go"))
				Ω(output).Should(ContainSubstring("already exists"))
			})
		})

		Context("with template argument", func() {
			It("should generate a test file using a template", func() {
				fm.WriteFile(pkg, ".generate", `package {{.Package}}
				import (
					{{.GinkgoImport}}
					{{.GomegaImport}}

					{{if .ImportPackage}}"{{.PackageImportPath}}"{{end}}
				)

				var _ = Describe("{{.Subject}}", func() {
					// This is a {{.Package}} test
				})`)
				session := startGinkgo(fm.PathTo(pkg), "generate", "--template", ".generate")
				Eventually(session).Should(gexec.Exit(0))
				output := session.Out.Contents()

				Ω(output).Should(ContainSubstring("foo_bar_test.go"))

				content := fm.ContentOf(pkg, "foo_bar_test.go")
				Ω(content).Should(ContainSubstring("package foo_bar_test"))
				Ω(content).Should(ContainSubstring(`. "github.com/onsi/ginkgo/v2"`))
				Ω(content).Should(ContainSubstring(`. "github.com/onsi/gomega"`))
				Ω(content).Should(ContainSubstring(`/foo_bar"`))
				Ω(content).Should(ContainSubstring("// This is a foo_bar_test test"))
			})

			It("should generate a test file using a template that contains functions", func() {
				fm.WriteFile(pkg, ".generate", `package {{.Package}}
				import (
					{{.GinkgoImport}}
					{{.GomegaImport}}

					{{if .ImportPackage}}"{{.PackageImportPath}}"{{end}}
				)

				var _ = Describe("{{.Subject}}", func() {
					// This is a {{.Package | repeat 3 }} test
				})`)
				session := startGinkgo(fm.PathTo(pkg), "generate", "--template", ".generate")
				Eventually(session).Should(gexec.Exit(0))
				output := session.Out.Contents()

				Ω(output).Should(ContainSubstring("foo_bar_test.go"))

				content := fm.ContentOf(pkg, "foo_bar_test.go")
				Ω(content).Should(ContainSubstring("package foo_bar_test"))
				Ω(content).Should(ContainSubstring(`. "github.com/onsi/ginkgo/v2"`))
				Ω(content).Should(ContainSubstring(`. "github.com/onsi/gomega"`))
				Ω(content).Should(ContainSubstring(`/foo_bar"`))
				Ω(content).Should(ContainSubstring("// This is a foo_bar_testfoo_bar_testfoo_bar_test test"))
			})
		})

		Context("with an argument of the form: foo", func() {
			It("should generate a test file named after the argument", func() {
				session := startGinkgo(fm.PathTo(pkg), "generate", "baz_buzz")
				Eventually(session).Should(gexec.Exit(0))
				output := session.Out.Contents()

				Ω(output).Should(ContainSubstring("baz_buzz_test.go"))

				content := fm.ContentOf(pkg, "baz_buzz_test.go")
				Ω(content).Should(ContainSubstring("package foo_bar_test"))
				Ω(content).Should(ContainSubstring(`var _ = Describe("BazBuzz", func() {`))
			})
		})

		Context("with an argument of the form: foo.go", func() {
			It("should generate a test file named after the argument", func() {
				session := startGinkgo(fm.PathTo(pkg), "generate", "baz_buzz.go")
				Eventually(session).Should(gexec.Exit(0))
				output := session.Out.Contents()

				Ω(output).Should(ContainSubstring("baz_buzz_test.go"))

				content := fm.ContentOf(pkg, "baz_buzz_test.go")
				Ω(content).Should(ContainSubstring("package foo_bar_test"))
				Ω(content).Should(ContainSubstring(`var _ = Describe("BazBuzz", func() {`))

			})
		})

		Context("with an argument of the form: foo_test", func() {
			It("should generate a test file named after the argument", func() {
				session := startGinkgo(fm.PathTo(pkg), "generate", "baz_buzz_test")
				Eventually(session).Should(gexec.Exit(0))
				output := session.Out.Contents()

				Ω(output).Should(ContainSubstring("baz_buzz_test.go"))

				content := fm.ContentOf(pkg, "baz_buzz_test.go")
				Ω(content).Should(ContainSubstring("package foo_bar_test"))
				Ω(content).Should(ContainSubstring(`var _ = Describe("BazBuzz", func() {`))
			})
		})

		Context("with an argument of the form: foo-test", func() {
			It("should generate a test file named after the argument", func() {
				session := startGinkgo(fm.PathTo(pkg), "generate", "baz-buzz-test")
				Eventually(session).Should(gexec.Exit(0))
				output := session.Out.Contents()

				Ω(output).Should(ContainSubstring("baz_buzz_test.go"))

				content := fm.ContentOf(pkg, "baz_buzz_test.go")
				Ω(content).Should(ContainSubstring("package foo_bar_test"))
				Ω(content).Should(ContainSubstring(`var _ = Describe("BazBuzz", func() {`))
			})
		})

		Context("with an argument of the form: foo_test.go", func() {
			It("should generate a test file named after the argument", func() {
				session := startGinkgo(fm.PathTo(pkg), "generate", "baz_buzz_test.go")
				Eventually(session).Should(gexec.Exit(0))
				output := session.Out.Contents()

				Ω(output).Should(ContainSubstring("baz_buzz_test.go"))

				content := fm.ContentOf(pkg, "baz_buzz_test.go")
				Ω(content).Should(ContainSubstring("package foo_bar_test"))
				Ω(content).Should(ContainSubstring(`var _ = Describe("BazBuzz", func() {`))
			})
		})

		Context("with multiple arguments", func() {
			It("should generate a test file named after the argument", func() {
				session := startGinkgo(fm.PathTo(pkg), "generate", "baz", "buzz")
				Eventually(session).Should(gexec.Exit(0))
				output := session.Out.Contents()

				Ω(output).Should(ContainSubstring("baz_test.go"))
				Ω(output).Should(ContainSubstring("buzz_test.go"))

				content := fm.ContentOf(pkg, "baz_test.go")
				Ω(content).Should(ContainSubstring("package foo_bar_test"))
				Ω(content).Should(ContainSubstring(`var _ = Describe("Baz", func() {`))

				content = fm.ContentOf(pkg, "buzz_test.go")
				Ω(content).Should(ContainSubstring("package foo_bar_test"))
				Ω(content).Should(ContainSubstring(`var _ = Describe("Buzz", func() {`))
			})
		})

		Context("with nodot", func() {
			It("should not import ginkgo or gomega", func() {
				session := startGinkgo(fm.PathTo(pkg), "generate", "--nodot")
				Eventually(session).Should(gexec.Exit(0))
				output := session.Out.Contents()

				Ω(output).Should(ContainSubstring("foo_bar_test.go"))

				content := fm.ContentOf(pkg, "foo_bar_test.go")
				Ω(content).Should(ContainSubstring("package foo_bar_test"))
				Ω(content).ShouldNot(ContainSubstring("\t" + `. "github.com/onsi/ginkgo/v2"`))
				Ω(content).ShouldNot(ContainSubstring("\t" + `. "github.com/onsi/gomega"`))
				Ω(content).Should(ContainSubstring("\t" + `"github.com/onsi/ginkgo/v2"`))
				Ω(content).Should(ContainSubstring("\t" + `"github.com/onsi/gomega"`))

				By("compiling correctly (we append to the file to make sure gomega is used)")
				fm.WriteFile(pkg, "foo_bar.go", "package foo_bar\nvar TRUE=true\n")
				fm.AppendToFile(pkg, "foo_bar_test.go", strings.Join([]string{``,
					`var _ = ginkgo.It("works", func() {`,
					`    gomega.Expect(foo_bar.TRUE).To(gomega.BeTrue())`,
					`})`,
				}, "\n"))
				Eventually(startGinkgo(fm.PathTo(pkg))).Should(gexec.Exit(0))
			})
		})
	})

	Describe("ginkgo bootstrap/generate", func() {
		var pkg string
		BeforeEach(func() {
			pkg = "some-crazy-thing"
			fm.MkEmpty(pkg)
		})

		Context("when the working directory is empty", func() {
			It("generates correctly named bootstrap and generate files with a package name derived from the directory", func() {
				session := startGinkgo(fm.PathTo(pkg), "bootstrap")
				Eventually(session).Should(gexec.Exit(0))

				content := fm.ContentOf(pkg, "some_crazy_thing_suite_test.go")
				Ω(content).Should(ContainSubstring("package some_crazy_thing_test"))
				Ω(content).Should(ContainSubstring("SomeCrazyThing Suite"))

				session = startGinkgo(fm.PathTo(pkg), "generate")
				Eventually(session).Should(gexec.Exit(0))

				content = fm.ContentOf(pkg, "some_crazy_thing_test.go")
				Ω(content).Should(ContainSubstring("package some_crazy_thing_test"))
				Ω(content).Should(ContainSubstring("SomeCrazyThing"))
			})
		})

		Context("when the working directory contains a file with a package name", func() {
			BeforeEach(func() {
				fm.WriteFile(pkg, "foo.go", "package main\n\nfunc main() {}")
			})

			It("generates correctly named bootstrap and generate files with the package name", func() {
				session := startGinkgo(fm.PathTo(pkg), "bootstrap")
				Eventually(session).Should(gexec.Exit(0))

				content := fm.ContentOf(pkg, "some_crazy_thing_suite_test.go")
				Ω(content).Should(ContainSubstring("package main_test"))
				Ω(content).Should(ContainSubstring("SomeCrazyThing Suite"))

				session = startGinkgo(fm.PathTo(pkg), "generate")
				Eventually(session).Should(gexec.Exit(0))

				content = fm.ContentOf(pkg, "some_crazy_thing_test.go")
				Ω(content).Should(ContainSubstring("package main_test"))
				Ω(content).Should(ContainSubstring("SomeCrazyThing"))
			})
		})
	})

	Describe("Go module and ginkgo bootstrap/generate", func() {
		var (
			pkg         string
			savedGoPath string
		)

		BeforeEach(func() {
			pkg = "myamazingmodule"
			fm.MkEmpty(pkg)
			fm.WriteFile(pkg, "go.mod", "module fake.com/me/myamazingmodule\n")
			savedGoPath = os.Getenv("GOPATH")
			Expect(os.Setenv("GOPATH", "")).To(Succeed())
			Expect(os.Setenv("GO111MODULE", "on")).To(Succeed()) // needed pre-Go 1.13
		})

		AfterEach(func() {
			Expect(os.Setenv("GOPATH", savedGoPath)).To(Succeed())
			Expect(os.Setenv("GO111MODULE", "")).To(Succeed())
		})

		It("generates correctly named bootstrap and generate files with the module name", func() {
			session := startGinkgo(fm.PathTo(pkg), "bootstrap")
			Eventually(session).Should(gexec.Exit(0))

			content := fm.ContentOf(pkg, "myamazingmodule_suite_test.go")
			Expect(content).To(ContainSubstring("package myamazingmodule_test"), string(content))
			Expect(content).To(ContainSubstring("Myamazingmodule Suite"), string(content))

			session = startGinkgo(fm.PathTo(pkg), "generate")
			Eventually(session).Should(gexec.Exit(0))

			content = fm.ContentOf(pkg, "myamazingmodule_test.go")
			Expect(content).To(ContainSubstring("package myamazingmodule_test"), string(content))
			Expect(content).To(ContainSubstring("fake.com/me/myamazingmodule"), string(content))
			Expect(content).To(ContainSubstring("Myamazingmodule"), string(content))
		})
	})

	Describe("ginkgo unfocus", func() {
		It("should unfocus tests", Label("slow"), func() {
			fm.MountFixture("focused")

			session := startGinkgo(fm.PathTo("focused"), "--no-color", "-r")
			Eventually(session).Should(gexec.Exit(types.GINKGO_FOCUS_EXIT_CODE))
			output := session.Out.Contents()

			Ω(string(output)).Should(ContainSubstring("Detected Programmatic Focus"))

			session = startGinkgo(fm.PathTo("focused"), "unfocus")
			Eventually(session).Should(gexec.Exit(0))
			output = session.Out.Contents()
			Ω(string(output)).ShouldNot(ContainSubstring("expected 'package'"))

			session = startGinkgo(fm.PathTo("focused"), "--no-color", "-r")
			Eventually(session).Should(gexec.Exit(0))
			output = session.Out.Contents()
			Ω(string(output)).Should(ContainSubstring("Ginkgo ran 2 suites"))
			Ω(string(output)).Should(ContainSubstring("Test Suite Passed"))
			Ω(string(output)).ShouldNot(ContainSubstring("Detected Programmatic Focus"))

			original := fm.ContentOfFixture("focused", "README.md")
			updated := fm.ContentOf("focused", "README.md")
			Ω(original).Should(Equal(updated))
		})

		It("should ignore the 'vendor' folder", func() {
			fm.MountFixture("focused_with_vendor")

			session := startGinkgo(fm.PathTo("focused_with_vendor"), "unfocus")
			Eventually(session).Should(gexec.Exit(0))

			session = startGinkgo(fm.PathTo("focused_with_vendor"), "--no-color")
			Eventually(session).Should(gexec.Exit(0))
			output := session.Out.Contents()
			Expect(string(output)).To(ContainSubstring("11 Passed"))
			Expect(string(output)).To(ContainSubstring("0 Skipped"))

			originalVendorPath := fm.PathToFixtureFile("focused_with_vendor", "vendor")
			updatedVendorPath := fm.PathTo("focused_with_vendor", "vendor")

			Expect(sameFolder(originalVendorPath, updatedVendorPath)).To(BeTrue())
		})
	})

	Describe("ginkgo version", func() {
		It("should print out the version info", func() {
			session := startGinkgo("", "version")
			Eventually(session).Should(gexec.Exit(0))
			output := session.Out.Contents()

			Ω(output).Should(MatchRegexp(`Ginkgo Version \d+\.\d+\.\d+`))
		})
	})

	Describe("ginkgo help", func() {
		It("should print out usage information", func() {
			session := startGinkgo("", "help")
			Eventually(session).Should(gexec.Exit(0))
			output := string(session.Out.Contents())

			Ω(output).Should(MatchRegexp(`Ginkgo Version \d+\.\d+\.\d+`))
			Ω(output).Should(ContainSubstring("watch"))
			Ω(output).Should(ContainSubstring("generate"))
			Ω(output).Should(ContainSubstring("run"))
		})

		It("should print out usage information for subcommands", func() {
			session := startGinkgo("", "help", "run")
			Eventually(session).Should(gexec.Exit(0))
			output := string(session.Out.Contents())

			Ω(output).Should(ContainSubstring("-succinct"))
			Ω(output).Should(ContainSubstring("-procs"))
		})
	})
})

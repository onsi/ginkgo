package testsuite_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/ginkgo/testsuite"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var _ = Describe("TestSuite", func() {
	var tmpDir string
	var relTmpDir string

	writeFile := func(folder string, filename string, content string) {
		path := filepath.Join(tmpDir, folder)
		err := os.MkdirAll(path, 0700)
		Ω(err).ShouldNot(HaveOccurred())

		path = filepath.Join(path, filename)
		ioutil.WriteFile(path, []byte(content), os.ModePerm)
	}

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("/tmp", "ginkgo")
		Ω(err).ShouldNot(HaveOccurred())

		cwd, err := os.Getwd()
		Ω(err).ShouldNot(HaveOccurred())
		relTmpDir, err = filepath.Rel(cwd, tmpDir)
		relTmpDir = "./" + relTmpDir
		Ω(err).ShouldNot(HaveOccurred())

		//go files in the root directory (no tests)
		writeFile("/", "main.go", "package main")

		//non-go files in a nested directory
		writeFile("/redherring", "big_test.jpg", "package ginkgo")

		//non-ginkgo tests in a nested directory
		writeFile("/proffessorplum", "proffessorplum_test.go", `import "testing"`)

		//ginkgo tests in a nested directory
		writeFile("/colonelmustard", "colonelmustard_test.go", `import "github.com/onsi/ginkgo"`)

		//ginkgo tests in a deeply nested directory
		writeFile("/colonelmustard/library", "library_test.go", `import "github.com/onsi/ginkgo"`)
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Describe("scanning for suites in a directory", func() {
		Context("when there are no tests in the specified directory", func() {
			It("should come up empty", func() {
				suites := SuitesInDir(tmpDir, false)
				Ω(suites).Should(BeEmpty())
			})
		})

		Context("when there are ginkgo tests in the specified directory", func() {
			It("should return an appropriately configured suite", func() {
				suites := SuitesInDir(filepath.Join(tmpDir, "colonelmustard"), false)
				Ω(suites).Should(HaveLen(1))

				Ω(suites[0].Path).Should(Equal(relTmpDir + "/colonelmustard"))
				Ω(suites[0].PackageName).Should(Equal("colonelmustard"))
				Ω(suites[0].IsGinkgo).Should(BeTrue())
			})
		})

		Context("when there are non-ginkgo tests in the specified directory", func() {
			It("should return an appropriately configured suite", func() {
				suites := SuitesInDir(filepath.Join(tmpDir, "proffessorplum"), false)
				Ω(suites).Should(HaveLen(1))

				Ω(suites[0].Path).Should(Equal(relTmpDir + "/proffessorplum"))
				Ω(suites[0].PackageName).Should(Equal("proffessorplum"))
				Ω(suites[0].IsGinkgo).Should(BeFalse())
			})
		})

		Context("when recursively scanning", func() {
			It("should return suites for corresponding test suites, only", func() {
				suites := SuitesInDir(tmpDir, true)
				Ω(suites).Should(HaveLen(3))

				Ω(suites).Should(ContainElement(&TestSuite{
					Path:        relTmpDir + "/colonelmustard",
					PackageName: "colonelmustard",
					IsGinkgo:    true,
				}))
				Ω(suites).Should(ContainElement(&TestSuite{
					Path:        relTmpDir + "/proffessorplum",
					PackageName: "proffessorplum",
					IsGinkgo:    false,
				}))
				Ω(suites).Should(ContainElement(&TestSuite{
					Path:        relTmpDir + "/colonelmustard/library",
					PackageName: "library",
					IsGinkgo:    true,
				}))
			})
		})
	})

	Describe("watching for changes", func() {
		var (
			suite   *TestSuite
			channel chan *TestSuite
		)

		BeforeEach(func() {
			suite = SuitesInDir(filepath.Join(tmpDir, "colonelmustard"), false)[0]
			channel = make(chan *TestSuite, 0)
			suite.Watch(channel)
		})

		Context("when a non-go file is created/modified", func() {
			BeforeEach(func() {
				writeFile("/colonelmustard", "poupon.jpg", "spicy")
			})

			It("should not send a notification down the channel", func() {
				Consistently(channel).ShouldNot(Receive())
			})
		})

		Context("when a go file is modified", func() {
			BeforeEach(func() {
				writeFile("/colonelmustard", "colonelmustard_test.go", `
import "github.com/onsi/ginkgo"
func() {}
`)
			})

			It("should send one notification down the channel", func() {
				Ω(<-channel).Should(Equal(suite))
				Consistently(channel).ShouldNot(Receive())
			})
		})

		Context("when a go file is created", func() {
			BeforeEach(func() {
				writeFile("/colonelmustard", "poupon.go", `
import "mustards"
func() {}
`)
			})

			It("should send one notification down the channel", func() {
				Ω(<-channel).Should(Equal(suite))
				Consistently(channel).ShouldNot(Receive())
			})
		})

		Context("when events are emitted but there are no actual changes", func() {
			BeforeEach(func() {
				writeFile("/colonelmustard", "colonelmustard_test.go", `import "github.com/onsi/ginkgo"`)
			})

			It("should send nothing down the channel", func() {
				Consistently(channel).ShouldNot(Receive())
			})
		})

		Context("when multiple changes happen", func() {
			Context("and the channel is read between changes", func() {
				It("should send a notification when the subsequent change happens", func() {
					writeFile("/colonelmustard", "poupon.go", `
import "mustards"
func() {}
`)
					Ω(<-channel).Should(Equal(suite))
					time.Sleep(10 * time.Millisecond)
					writeFile("/colonelmustard", "poupon.go", `
import "mustards"
func Color() string {
    return "grey"
}
`)
					Ω(<-channel).Should(Equal(suite))
					Consistently(channel).ShouldNot(Receive())
				})
			})

			Context("and the channel is only read *after* the changes have happened", func() {
				It("should not pile on multiple notifications", func() {
					writeFile("/colonelmustard", "poupon.go", `
import "mustards"
func() {}
`)
					writeFile("/colonelmustard", "poupon.go", `
import "mustards"
func Color() string {
    return "grey"
}
`)
					Ω(<-channel).Should(Equal(suite))
					Consistently(channel).ShouldNot(Receive())
				})
			})
		})
	})
})

package internal_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/formatter"
	"github.com/onsi/ginkgo/ginkgo/internal"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	Describe("FileExists", func() {
		var tmpDir string

		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("/tmp", "ginkgo")
			Ω(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			Ω(os.RemoveAll(tmpDir)).Should(Succeed())
		})

		It("returns true if the path exists", func() {
			path := filepath.Join(tmpDir, "foo")
			Ω(os.WriteFile(path, []byte("foo"), 0666)).Should(Succeed())
			Ω(internal.FileExists(path)).Should(BeTrue())
		})

		It("returns false if the path does not exist", func() {
			path := filepath.Join(tmpDir, "foo")
			Ω(internal.FileExists(path)).Should(BeFalse())
		})
	})

	Describe("PluralizedWord", func() {
		It("returns singular when count is 1", func() {
			Ω(internal.PluralizedWord("s", "p", 1)).Should(Equal("s"))
		})

		It("returns plural when count is not 1", func() {
			Ω(internal.PluralizedWord("s", "p", 0)).Should(Equal("p"))
			Ω(internal.PluralizedWord("s", "p", 2)).Should(Equal("p"))
			Ω(internal.PluralizedWord("s", "p", 10)).Should(Equal("p"))
		})
	})

	Describe("FailedSuiteReport", func() {
		var f formatter.Formatter
		BeforeEach(func() {
			f = formatter.New(formatter.ColorModePassthrough)
		})

		It("generates a nicely frormatter report", func() {
			suites := []internal.TestSuite{
				TS("path-A", "package-A", true, internal.TestSuiteStateFailed),
				TS("path-B", "B", true, internal.TestSuiteStateFailedToCompile),
				TS("path-to/package-C", "the-C-package", true, internal.TestSuiteStateFailedDueToTimeout),
				TS("path-D", "D", true, internal.TestSuiteStatePassed),
				TS("path-F", "E", true, internal.TestSuiteStateSkippedByFilter),
				TS("path-F", "E", true, internal.TestSuiteStateSkippedDueToPriorFailures),
			}

			Ω(internal.FailedSuitesReport(suites, f)).Should(HavePrefix(strings.Join([]string{
				"There were failures detected in the following suites:",
				"  {{red}}    package-A {{gray}}path-A{{/}}",
				"  {{red}}            B {{gray}}path-B {{magenta}}[Compilation failure]{{/}}",
				"  {{red}}the-C-package {{gray}}path-to/package-C {{orange}}[Suite did not run because the timeout elapsed]{{/}}",
			}, "\n")))
		})
	})
})

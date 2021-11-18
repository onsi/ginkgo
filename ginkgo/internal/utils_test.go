package internal_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/formatter"
	"github.com/onsi/ginkgo/v2/ginkgo/internal"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	Describe("FileExists", func() {
		var tmpDir string

		BeforeEach(func() {
			tmpDir = GinkgoT().TempDir()
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

	Describe("Copying Files", func() {
		var tmpDirA, tmpDirB string
		var j = filepath.Join

		BeforeEach(func() {
			tmpDirA = GinkgoT().TempDir()
			tmpDirB = GinkgoT().TempDir()

			os.WriteFile(j(tmpDirA, "file_a"), []byte("FILE_A"), 0666)
			os.WriteFile(j(tmpDirA, "file_b"), []byte("FILE_B"), 0777)
			os.WriteFile(j(tmpDirB, "file_c"), []byte("FILE_C"), 0666)
		})

		DescribeTable("it copies files, overwriting existing content and preserve permissions",
			func(src string, dest string) {
				src, dest = j(tmpDirA, src), j(tmpDirB, dest)
				Ω(internal.CopyFile(src, dest)).Should(Succeed())
				expectedContent, err := os.ReadFile(src)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(os.ReadFile(dest)).Should(Equal(expectedContent))
				expectedStat, err := os.Stat(src)
				stat, err := os.Stat(dest)
				Ω(stat.Mode()).Should(Equal(expectedStat.Mode()))
			},
			Entry(nil, "file_a", "file_a"),
			Entry(nil, "file_b", "file_b"),
			Entry(nil, "file_b", "file_c"),
		)

		It("fails when src does not exist", func() {
			err := internal.CopyFile(j(tmpDirA, "file_c"), j(tmpDirB, "file_c"))
			Ω(err).Should(HaveOccurred())
			Ω(os.ReadFile(j(tmpDirB, "file_c"))).Should(Equal([]byte("FILE_C")))
		})

		It("fails when dest's directory does not exist", func() {
			err := internal.CopyFile(j(tmpDirA, "file_a"), j(tmpDirB, "foo", "file_a"))
			Ω(err).Should(HaveOccurred())
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

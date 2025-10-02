package reporters_test

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/joshdk/go-junit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
)

var _ = Describe("JunitReport", func() {
	var report types.Report

	BeforeEach(func() {
		report = types.Report{
			SuiteDescription: "My Suite",
			SuitePath:        "/path/to/suite",
			PreRunStats:      types.PreRunStats{SpecsThatWillRun: 15, TotalSpecs: 20},
			SuiteConfig:      types.SuiteConfig{RandomSeed: 17, ParallelTotal: 1},
			RunTime:          time.Minute,
			SpecReports: types.SpecReports{
				S(types.NodeTypeIt, Label("cat", "dog"), CLabels(Label("dolphin"), Label("gorilla", "cow")), CTS("A", "B"), CLS(cl0, cl1), "C", cl2, types.SpecStateTimedout, STD("some captured stdout\n"), GW("ginkgowriter\noutput\ncleanup!"), SE(types.SpecEventByStart, "a by step", cl0),
					SE(types.SpecEventNodeStart, types.NodeTypeIt, "C", cl2, TL(0)),
					F("failure\nmessage", cl3, types.FailureNodeIsLeafNode, FailureNodeLocation(cl2), types.NodeTypeIt, TL("ginkgowriter\n"), AF(types.SpecStatePanicked, cl4, types.FailureNodeIsLeafNode, FailureNodeLocation(cl2), types.NodeTypeIt, TL("ginkgowriter\noutput\n"), ForwardedPanic("the panic!"))),
					SE(types.SpecEventNodeEnd, types.NodeTypeIt, "C", cl2, TL("ginkgowriter\noutput\n"), time.Microsecond*87230),
					RE("a report entry", cl1, TL("ginkgowriter\noutput\n")),
					RE("a hidden report entry", cl1, TL("ginkgowriter\noutput\n"), types.ReportEntryVisibilityNever),
					AF(types.SpecStateFailed, "a subsequent failure", types.FailureNodeInContainer, FailureNodeLocation(cl3), types.NodeTypeAfterEach, 0, TL("ginkgowriter\noutput\ncleanup!")),
				),
				S(types.NodeTypeIt, "A", cl0, STD("some captured stdout\n"), GW("some GinkgoWriter\noutput is interspersed\nhere and there\n"), Label("cat", "owner:frank", "OWNer:bob"),
					SE(types.SpecEventNodeStart, types.NodeTypeIt, "A", cl0),
					PR("my progress report", LeafNodeText("A"), TL("some GinkgoWriter\n")),
					SE(types.SpecEventByStart, "My Step", cl1, TL("some GinkgoWriter\n")),
					RE("my entry", cl1, types.ReportEntryVisibilityFailureOrVerbose, TL("some GinkgoWriter\noutput is interspersed\n")),
					RE("my hidden entry", cl1, types.ReportEntryVisibilityNever, TL("some GinkgoWriter\noutput is interspersed\n")),
					SE(types.SpecEventByEnd, "My Step", cl1, time.Millisecond*200, TL("some GinkgoWriter\noutput is interspersed\n")),
					SE(types.SpecEventNodeEnd, types.NodeTypeIt, "A", cl0, time.Millisecond*300, TL("some GinkgoWriter\noutput is interspersed\nhere and there\n")),
				),
				S(types.NodeTypeIt, "A", cl0, types.SpecStatePending, CLabels(Label("owner:org")), Label("owner:team")),
				S(types.NodeTypeIt, "A", cl0, types.SpecStatePanicked, CLabels(Label("owner:org")), STD("some captured stdout\n"),
					SE(types.SpecEventNodeStart, types.NodeTypeIt, "A", cl0),
					F("failure\nmessage", cl1, types.FailureNodeIsLeafNode, FailureNodeLocation(cl0), types.NodeTypeIt, ForwardedPanic("the panic")),
					SE(types.SpecEventNodeEnd, types.NodeTypeIt, "A", cl0, time.Millisecond*300, TL("some GinkgoWriter\noutput is interspersed\nhere and there\n")),
				),
				S(types.NodeTypeBeforeSuite, "A", cl0, types.SpecStatePassed),
			},
		}
	})

	Describe("default behavior", func() {
		var generated reporters.JUnitTestSuites

		BeforeEach(func() {
			fname := fmt.Sprintf("./report-%d", GinkgoParallelProcess())
			Ω(reporters.GenerateJUnitReport(report, fname)).Should(Succeed())
			DeferCleanup(os.Remove, fname)

			generated = reporters.JUnitTestSuites{}
			f, err := os.Open(fname)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(xml.NewDecoder(f).Decode(&generated)).Should(Succeed())
		})

		It("generates a Junit report and writes it to disk", func() {
			Ω(generated.Tests).Should(Equal(5))
			Ω(generated.Disabled).Should(Equal(1))
			Ω(generated.Errors).Should(Equal(1))
			Ω(generated.Failures).Should(Equal(1))
			Ω(generated.Time).Should(Equal(60.0))
			Ω(generated.TestSuites).Should(HaveLen(1))

			suite := generated.TestSuites[0]
			Ω(suite.Name).Should(Equal("My Suite"))
			Ω(suite.Package).Should(Equal("/path/to/suite"))
			Ω(suite.Properties.WithName("SuiteSucceeded")).Should(Equal("false"))
			Ω(suite.Properties.WithName("RandomSeed")).Should(Equal("17"))

			Ω(suite.Tests).Should(Equal(5))
			Ω(suite.Disabled).Should(Equal(1))
			Ω(suite.Errors).Should(Equal(1))
			Ω(suite.Failures).Should(Equal(1))
			Ω(suite.Time).Should(Equal(60.0))

			Ω(suite.TestCases).Should(HaveLen(5))

			failingSpec := suite.TestCases[0]
			Ω(failingSpec.Name).Should(Equal("[It] A B C [dolphin, gorilla, cow, cat, dog]"))
			Ω(failingSpec.Classname).Should(Equal("My Suite"))
			Ω(failingSpec.Status).Should(Equal("timedout"))
			Ω(failingSpec.Skipped).Should(BeNil())
			Ω(failingSpec.Error).Should(BeNil())
			Ω(failingSpec.Owner).Should(Equal(""))
			Ω(failingSpec.Failure.Message).Should(Equal("failure\nmessage"))
			Ω(failingSpec.Failure.Type).Should(Equal("timedout"))
			Ω(failingSpec.Failure.Description).Should(MatchLines(
				"[TIMEDOUT] failure",
				"message",
				spr("In [It] at: cl3.go:103 @ %s", FORMATTED_TIME),
				"",
				"[PANICKED] ",
				spr("In [It] at: cl4.go:144 @ %s", FORMATTED_TIME),
				"",
				"the panic!",
				"",
				"Full Stack Trace",
				"  full-trace",
				"  cl-4",
				"",
				"There were additional failures detected after the initial failure. These are visible in the timeline",
				"",
			))
			Ω(failingSpec.SystemOut).Should(Equal("some captured stdout\n"))
			Ω(failingSpec.SystemErr).Should(MatchLines(
				spr("STEP: a by step - cl0.go:12 @ %s", FORMATTED_TIME),
				spr("> Enter [It] C - cl2.go:80 @ %s", FORMATTED_TIME),
				"ginkgowriter",
				"[TIMEDOUT] failure",
				"message",
				spr("In [It] at: cl3.go:103 @ %s", FORMATTED_TIME),
				"output",
				"[PANICKED] ",
				spr("In [It] at: cl4.go:144 @ %s", FORMATTED_TIME),
				"",
				"the panic!",
				"",
				"Full Stack Trace",
				"  full-trace",
				"  cl-4",
				spr("< Exit [It] C - cl2.go:80 @ %s (87ms)", FORMATTED_TIME),
				spr("a report entry - cl1.go:37 @ %s", FORMATTED_TIME),
				spr("a hidden report entry - cl1.go:37 @ %s", FORMATTED_TIME),
				"cleanup!",
				"[FAILED] a subsequent failure",
				spr("In [AfterEach] at: :0 @ %s", FORMATTED_TIME),
				"",
			))

			passingSpec := suite.TestCases[1]
			Ω(passingSpec.Name).Should(Equal("[It] A [cat, owner:frank, OWNer:bob]"))
			Ω(passingSpec.Classname).Should(Equal("My Suite"))
			Ω(passingSpec.Status).Should(Equal("passed"))
			Ω(passingSpec.Skipped).Should(BeNil())
			Ω(passingSpec.Error).Should(BeNil())
			Ω(passingSpec.Failure).Should(BeNil())
			Ω(passingSpec.Owner).Should(Equal("bob"))
			Ω(passingSpec.SystemOut).Should(Equal("some captured stdout\n"))
			Ω(passingSpec.SystemErr).Should(MatchLines(
				spr("> Enter [It] A - cl0.go:12 @ %s", FORMATTED_TIME),
				"some GinkgoWriter",
				"my progress report",
				"  A (Spec Runtime: 5s)",
				"    cl0.go:12",
				spr("STEP: My Step - cl1.go:37 @ %s", FORMATTED_TIME),
				"output is interspersed",
				spr("my entry - cl1.go:37 @ %s", FORMATTED_TIME),
				spr("my hidden entry - cl1.go:37 @ %s", FORMATTED_TIME),
				spr("END STEP: My Step - cl1.go:37 @ %s (200ms)", FORMATTED_TIME),
				"here and there",
				spr("< Exit [It] A - cl0.go:12 @ %s (300ms)", FORMATTED_TIME),
				"",
			))

			pendingSpec := suite.TestCases[2]
			Ω(pendingSpec.Name).Should(Equal("[It] A [owner:org, owner:team]"))
			Ω(pendingSpec.Classname).Should(Equal("My Suite"))
			Ω(pendingSpec.Status).Should(Equal("pending"))
			Ω(pendingSpec.Skipped.Message).Should(Equal("pending"))
			Ω(pendingSpec.Error).Should(BeNil())
			Ω(pendingSpec.Owner).Should(Equal("team"))
			Ω(pendingSpec.Failure).Should(BeNil())
			Ω(pendingSpec.SystemOut).Should(BeEmpty())
			Ω(pendingSpec.SystemErr).Should(BeEmpty())

			panickedSpec := suite.TestCases[3]
			Ω(panickedSpec.Name).Should(Equal("[It] A [owner:org]"))
			Ω(panickedSpec.Classname).Should(Equal("My Suite"))
			Ω(panickedSpec.Status).Should(Equal("panicked"))
			Ω(panickedSpec.Skipped).Should(BeNil())
			Ω(panickedSpec.Owner).Should(Equal("org"))
			Ω(panickedSpec.Error.Message).Should(Equal("the panic"))
			Ω(panickedSpec.Error.Type).Should(Equal("panicked"))
			Ω(panickedSpec.Error.Description).Should(MatchLines(
				"[PANICKED] failure",
				"message",
				spr("In [It] at: cl1.go:37 @ %s", FORMATTED_TIME),
				"",
				"the panic",
				"",
				"Full Stack Trace",
				"  full-trace",
				"  cl-1",
				"",
			))
			Ω(panickedSpec.Failure).Should(BeNil())
			Ω(panickedSpec.SystemOut).Should(Equal("some captured stdout\n"))
			Ω(panickedSpec.SystemErr).Should(MatchLines(
				spr("> Enter [It] A - cl0.go:12 @ %s", FORMATTED_TIME),
				"[PANICKED] failure",
				"message",
				spr("In [It] at: cl1.go:37 @ %s", FORMATTED_TIME),
				"",
				"the panic",
				"",
				"Full Stack Trace",
				"  full-trace",
				"  cl-1",
				spr("< Exit [It] A - cl0.go:12 @ %s (300ms)", FORMATTED_TIME),
				"",
			))

			beforeSuiteSpec := suite.TestCases[4]
			Ω(beforeSuiteSpec.Name).Should(Equal("[BeforeSuite] A"))
			Ω(beforeSuiteSpec.Classname).Should(Equal("My Suite"))
			Ω(beforeSuiteSpec.Status).Should(Equal("passed"))
			Ω(beforeSuiteSpec.Skipped).Should(BeNil())
			Ω(beforeSuiteSpec.Error).Should(BeNil())
			Ω(beforeSuiteSpec.Failure).Should(BeNil())
			Ω(beforeSuiteSpec.SystemOut).Should(BeEmpty())
			Ω(beforeSuiteSpec.SystemErr).Should(BeEmpty())
		})
	})

	Describe("when configured to omit all the omittables", func() {
		var generated reporters.JUnitTestSuites

		BeforeEach(func() {
			fname := fmt.Sprintf("./report-%d", GinkgoParallelProcess())
			Ω(reporters.GenerateJUnitReportWithConfig(report, fname, reporters.JunitReportConfig{
				OmitTimelinesForSpecState: types.SpecStatePassed,
				OmitFailureMessageAttr:    true,
				OmitCapturedStdOutErr:     true,
				OmitSpecLabels:            true,
				OmitLeafNodeType:          true,
				OmitSuiteSetupNodes:       true,
			})).Should(Succeed())
			DeferCleanup(os.Remove, fname)

			generated = reporters.JUnitTestSuites{}
			f, err := os.Open(fname)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(xml.NewDecoder(f).Decode(&generated)).Should(Succeed())
		})

		It("generates a Junit report and writes it to disk", func() {
			Ω(generated.Tests).Should(Equal(4))
			Ω(generated.Disabled).Should(Equal(1))
			Ω(generated.Errors).Should(Equal(1))
			Ω(generated.Failures).Should(Equal(1))
			Ω(generated.Time).Should(Equal(60.0))
			Ω(generated.TestSuites).Should(HaveLen(1))

			suite := generated.TestSuites[0]
			Ω(suite.Name).Should(Equal("My Suite"))
			Ω(suite.Package).Should(Equal("/path/to/suite"))
			Ω(suite.Properties.WithName("SuiteSucceeded")).Should(Equal("false"))
			Ω(suite.Properties.WithName("RandomSeed")).Should(Equal("17"))

			Ω(suite.Tests).Should(Equal(4))
			Ω(suite.Disabled).Should(Equal(1))
			Ω(suite.Errors).Should(Equal(1))
			Ω(suite.Failures).Should(Equal(1))
			Ω(suite.Time).Should(Equal(60.0))

			Ω(suite.TestCases).Should(HaveLen(4))

			failingSpec := suite.TestCases[0]
			Ω(failingSpec.Name).Should(Equal("A B C"))
			Ω(failingSpec.Classname).Should(Equal("My Suite"))
			Ω(failingSpec.Status).Should(Equal("timedout"))
			Ω(failingSpec.Skipped).Should(BeNil())
			Ω(failingSpec.Error).Should(BeNil())
			Ω(failingSpec.Failure.Message).Should(BeEmpty())
			Ω(failingSpec.Failure.Type).Should(Equal("timedout"))
			Ω(failingSpec.Failure.Description).Should(MatchLines(
				"[TIMEDOUT] failure",
				"message",
				spr("In [It] at: cl3.go:103 @ %s", FORMATTED_TIME),
				"",
				"[PANICKED] ",
				spr("In [It] at: cl4.go:144 @ %s", FORMATTED_TIME),
				"",
				"the panic!",
				"",
				"Full Stack Trace",
				"  full-trace",
				"  cl-4",
				"",
				"There were additional failures detected after the initial failure. These are visible in the timeline",
				"",
			))
			Ω(failingSpec.SystemOut).Should(BeEmpty())
			Ω(failingSpec.SystemErr).Should(MatchLines(
				spr("STEP: a by step - cl0.go:12 @ %s", FORMATTED_TIME),
				spr("> Enter [It] C - cl2.go:80 @ %s", FORMATTED_TIME),
				"ginkgowriter",
				"[TIMEDOUT] failure",
				"message",
				spr("In [It] at: cl3.go:103 @ %s", FORMATTED_TIME),
				"output",
				"[PANICKED] ",
				spr("In [It] at: cl4.go:144 @ %s", FORMATTED_TIME),
				"",
				"the panic!",
				"",
				"Full Stack Trace",
				"  full-trace",
				"  cl-4",
				spr("< Exit [It] C - cl2.go:80 @ %s (87ms)", FORMATTED_TIME),
				spr("a report entry - cl1.go:37 @ %s", FORMATTED_TIME),
				spr("a hidden report entry - cl1.go:37 @ %s", FORMATTED_TIME),
				"cleanup!",
				"[FAILED] a subsequent failure",
				spr("In [AfterEach] at: :0 @ %s", FORMATTED_TIME),
				"",
			))

			passingSpec := suite.TestCases[1]
			Ω(passingSpec.Name).Should(Equal("A"))
			Ω(passingSpec.Classname).Should(Equal("My Suite"))
			Ω(passingSpec.Status).Should(Equal("passed"))
			Ω(passingSpec.Skipped).Should(BeNil())
			Ω(passingSpec.Error).Should(BeNil())
			Ω(passingSpec.Failure).Should(BeNil())
			Ω(passingSpec.SystemOut).Should(BeEmpty())
			Ω(passingSpec.SystemErr).Should(BeEmpty())

			pendingSpec := suite.TestCases[2]
			Ω(pendingSpec.Name).Should(Equal("A"))
			Ω(pendingSpec.Classname).Should(Equal("My Suite"))
			Ω(pendingSpec.Status).Should(Equal("pending"))
			Ω(pendingSpec.Skipped.Message).Should(Equal("pending"))
			Ω(pendingSpec.Error).Should(BeNil())
			Ω(pendingSpec.Failure).Should(BeNil())
			Ω(pendingSpec.SystemOut).Should(BeEmpty())
			Ω(pendingSpec.SystemErr).Should(BeEmpty())

			panickedSpec := suite.TestCases[3]
			Ω(panickedSpec.Name).Should(Equal("A"))
			Ω(panickedSpec.Classname).Should(Equal("My Suite"))
			Ω(panickedSpec.Status).Should(Equal("panicked"))
			Ω(panickedSpec.Skipped).Should(BeNil())
			Ω(panickedSpec.Error.Message).Should(BeEmpty())
			Ω(panickedSpec.Error.Type).Should(Equal("panicked"))
			Ω(panickedSpec.Error.Description).Should(MatchLines(
				"[PANICKED] failure",
				"message",
				spr("In [It] at: cl1.go:37 @ %s", FORMATTED_TIME),
				"",
				"the panic",
				"",
				"Full Stack Trace",
				"  full-trace",
				"  cl-1",
				"",
			))
			Ω(panickedSpec.Failure).Should(BeNil())
			Ω(panickedSpec.SystemOut).Should(BeEmpty())
			Ω(panickedSpec.SystemErr).Should(MatchLines(
				spr("> Enter [It] A - cl0.go:12 @ %s", FORMATTED_TIME),
				"[PANICKED] failure",
				"message",
				spr("In [It] at: cl1.go:37 @ %s", FORMATTED_TIME),
				"",
				"the panic",
				"",
				"Full Stack Trace",
				"  full-trace",
				"  cl-1",
				spr("< Exit [It] A - cl0.go:12 @ %s (300ms)", FORMATTED_TIME),
				"",
			))
		})
	})

	Describe("when configured to write the report inside a folder", func() {
		var folderPath string
		var filePath string

		BeforeEach(func() {
			folderPath = filepath.Join(fmt.Sprintf("test_outputs_%d", GinkgoParallelProcess()))
			fileName := fmt.Sprintf("report-%d", GinkgoParallelProcess())
			filePath = filepath.Join(folderPath, fileName)

			Ω(reporters.GenerateJUnitReport(report, filePath)).Should(Succeed())
			DeferCleanup(os.RemoveAll, folderPath)
		})

		It("creates the folder and the report file", func() {
			_, err := os.Stat(folderPath)
			Ω(err).Should(Succeed(), "Parent folder should be created")
			_, err = os.Stat(filePath)
			Ω(err).Should(Succeed(), "Report file should be created")
			reportBytes, err := os.ReadFile(filePath)
			Ω(err).Should(Succeed(), "Report file should be read")
			snaps.MatchSnapshot(GinkgoT(), string(reportBytes))
			suites, err := junit.IngestFile(filePath)
			Ω(err).Should(Succeed(), "Report file should be parsed")

			var summaryOutput bytes.Buffer
			writeSummary := func(in []byte) {
				_, err := summaryOutput.Write(in)
				Ω(err).Should(Succeed(), "writing to summary should succeed")
			}
			for _, suite := range suites {
				writeSummary([]byte(fmt.Sprintf("SUITE: %s, tests: %d \n", suite.Package, len(suite.Tests))))
				for _, test := range suite.Tests {
					writeSummary([]byte(fmt.Sprintf("---- TEST: %s, status: %s\n", test.Name, test.Status)))
				}
				writeSummary([]byte(""))
			}
			snaps.MatchSnapshot(GinkgoT(), summaryOutput.String(), "package summary output match")
		})
	})
})

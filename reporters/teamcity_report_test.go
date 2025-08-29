package reporters_test

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gkampitakis/go-snaps/snaps"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
)

var _ = Describe("TeamCityReport", func() {
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
				S(types.NodeTypeIt, "A", cl0, STD("some captured stdout\n"), GW("some GinkgoWriter\noutput is interspersed\nhere and there\n"),
					SE(types.SpecEventNodeStart, types.NodeTypeIt, "A", cl0),
					PR("my progress report", LeafNodeText("A"), TL("some GinkgoWriter\n")),
					SE(types.SpecEventByStart, "My Step", cl1, TL("some GinkgoWriter\n")),
					RE("my entry", cl1, types.ReportEntryVisibilityFailureOrVerbose, TL("some GinkgoWriter\noutput is interspersed\n")),
					RE("my hidden entry", cl1, types.ReportEntryVisibilityNever, TL("some GinkgoWriter\noutput is interspersed\n")),
					SE(types.SpecEventByEnd, "My Step", cl1, time.Millisecond*200, TL("some GinkgoWriter\noutput is interspersed\n")),
					SE(types.SpecEventNodeEnd, types.NodeTypeIt, "A", cl0, time.Millisecond*300, TL("some GinkgoWriter\noutput is interspersed\nhere and there\n")),
				),
				S(types.NodeTypeIt, "A", cl0, types.SpecStatePending),
				S(types.NodeTypeIt, "A", cl0, types.SpecStatePanicked, STD("some captured stdout\n"),
					SE(types.SpecEventNodeStart, types.NodeTypeIt, "A", cl0),
					F("failure\nmessage", cl1, types.FailureNodeIsLeafNode, FailureNodeLocation(cl0), types.NodeTypeIt, ForwardedPanic("the panic")),
					SE(types.SpecEventNodeEnd, types.NodeTypeIt, "A", cl0, time.Millisecond*300, TL("some GinkgoWriter\noutput is interspersed\nhere and there\n")),
				),
				S(types.NodeTypeBeforeSuite, "A", cl0, types.SpecStatePassed),
			},
		}
	})

	Describe("when configured to write the report inside a folder", func() {
		var folderPath string
		var filePath string

		BeforeEach(func() {
			folderPath = filepath.Join(fmt.Sprintf("test_outputs_%d", GinkgoParallelProcess()))
			fileName := fmt.Sprintf("report-%d", GinkgoParallelProcess())
			filePath = filepath.Join(folderPath, fileName)

			Ω(reporters.GenerateTeamcityReport(report, filePath)).Should(Succeed())
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
		})
	})
})

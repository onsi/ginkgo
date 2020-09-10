package reporters_test

import (
	"reflect"
	"runtime"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/gbytes"
)

type StackTrace string

const DELIMITER = `{{gray}}------------------------------{{/}}`

var cl0 = types.CodeLocation{FileName: "cl0.go", LineNumber: 12, FullStackTrace: "full-trace\ncl-0"}
var cl1 = types.CodeLocation{FileName: "cl1.go", LineNumber: 37, FullStackTrace: "full-trace\ncl-1"}
var cl2 = types.CodeLocation{FileName: "cl2.go", LineNumber: 80, FullStackTrace: "full-trace\ncl-2"}
var cl3 = types.CodeLocation{FileName: "cl3.go", LineNumber: 103, FullStackTrace: "full-trace\ncl-3"}

func CLS(cls ...types.CodeLocation) []types.CodeLocation { return cls }
func CTS(componentTexts ...string) []string              { return componentTexts }

type Location types.CodeLocation
type ForwardedPanic string

// convenience helper to quickly make Failures
func F(options ...interface{}) types.Failure {
	failure := types.Failure{}
	for _, option := range options {
		switch reflect.TypeOf(option) {
		case reflect.TypeOf(""):
			failure.Message = option.(string)
		case reflect.TypeOf(Location{}):
			failure.Location = types.CodeLocation(option.(Location))
		case reflect.TypeOf(ForwardedPanic("")):
			failure.ForwardedPanic = string(option.(ForwardedPanic))
		case reflect.TypeOf(0):
			failure.NodeIndex = option.(int)
		case reflect.TypeOf(types.NodeTypeIt):
			failure.NodeType = option.(types.NodeType)
		}
	}
	return failure
}

type STD string
type GW string

// convenience helper to quickly make summaries
func S(options ...interface{}) types.Summary {
	summary := types.Summary{
		LeafNodeType: types.NodeTypeIt,
		State:        types.SpecStatePassed,
		NumAttempts:  1,
		RunTime:      time.Second,
	}
	for _, option := range options {
		switch reflect.TypeOf(option) {
		case reflect.TypeOf([]string{}):
			summary.NodeTexts = option.([]string)
		case reflect.TypeOf([]types.CodeLocation{}):
			summary.NodeLocations = option.([]types.CodeLocation)
		case reflect.TypeOf(types.NodeTypeIt):
			summary.LeafNodeType = option.(types.NodeType)
		case reflect.TypeOf(types.CodeLocation{}):
			summary.LeafNodeLocation = option.(types.CodeLocation)
		case reflect.TypeOf(types.SpecStatePassed):
			summary.State = option.(types.SpecState)
		case reflect.TypeOf(time.Second):
			summary.RunTime = option.(time.Duration)
		case reflect.TypeOf(types.Failure{}):
			summary.Failure = option.(types.Failure)
		case reflect.TypeOf(0):
			summary.NumAttempts = option.(int)
		case reflect.TypeOf(STD("")):
			summary.CapturedStdOutErr = string(option.(STD))
		case reflect.TypeOf(GW("")):
			summary.CapturedGinkgoWriterOutput = string(option.(GW))
		}
	}
	return summary
}

type ConfigFlags uint8

const (
	Succinct ConfigFlags = 1 << iota
	Verbose
	ReportPassed
	FullTrace
)

func (cf ConfigFlags) Has(flag ConfigFlags) bool { return cf&flag != 0 }

func C(flags ...ConfigFlags) config.DefaultReporterConfigType {
	f := ConfigFlags(0)
	if len(flags) > 0 {
		f = flags[0]
	}
	Ω(f.Has(Verbose) && f.Has(Succinct)).Should(BeFalse(), "Being both verbose and succinct is a configuration error")
	return config.DefaultReporterConfigType{
		NoColor:           true,
		SlowSpecThreshold: SlowSpecThreshold,
		Succinct:          f.Has(Succinct),
		Verbose:           f.Has(Verbose),
		ReportPassed:      f.Has(ReportPassed),
		FullTrace:         f.Has(FullTrace),
	}
}

const SlowSpecThreshold = 3.0

var _ = Describe("DefaultReporter", func() {
	var DENOTER = "•"
	var RETRY_DENOTER = "↺"
	if runtime.GOOS == "windows" {
		DENOTER = "+"
		RETRY_DENOTER = "R"
	}

	var buf *gbytes.Buffer
	verifyExpectedOutput := func(expected []string) {
		if len(expected) == 0 {
			ExpectWithOffset(1, buf.Contents()).Should(BeEmpty())
		} else {
			ExpectWithOffset(1, string(buf.Contents())).Should(Equal(strings.Join(expected, "\n")))
		}
	}

	BeforeEach(func() {
		buf = gbytes.NewBuffer()
		format.CharactersAroundMismatchToInclude = 100
	})

	DescribeTable("Rendering SpecSuiteWillBegin",
		func(conf config.DefaultReporterConfigType, gConf config.GinkgoConfigType, summary types.SuiteSummary, expected ...string) {
			reporter := reporters.NewDefaultReporterUnderTest(conf, buf)
			reporter.SpecSuiteWillBegin(gConf, summary)
			verifyExpectedOutput(expected)
		},
		Entry("Default Behavior",
			C(),
			config.GinkgoConfigType{RandomSeed: 17, ParallelTotal: 1},
			types.SuiteSummary{SuiteDescription: "My Suite", NumberOfSpecsThatWillBeRun: 15, NumberOfTotalSpecs: 20},
			"Running Suite: My Suite",
			"=======================",
			"Random Seed: {{bold}}17{{/}}",
			"",
			"Will run {{bold}}15{{/}} of {{bold}}20{{/}} specs",
			"",
		),
		Entry("When configured to randomize all specs",
			C(),
			config.GinkgoConfigType{RandomSeed: 17, ParallelTotal: 1, RandomizeAllSpecs: true},
			types.SuiteSummary{SuiteDescription: "My Suite", NumberOfSpecsThatWillBeRun: 15, NumberOfTotalSpecs: 20},
			"Running Suite: My Suite",
			"=======================",
			"Random Seed: {{bold}}17{{/}} - will randomize all specs",
			"",
			"Will run {{bold}}15{{/}} of {{bold}}20{{/}} specs",
			"",
		),
		Entry("when configured to run in parallel",
			C(),
			config.GinkgoConfigType{RandomSeed: 17, ParallelTotal: 3},
			types.SuiteSummary{SuiteDescription: "My Suite", NumberOfSpecsThatWillBeRun: 15, NumberOfTotalSpecs: 20},
			"Running Suite: My Suite",
			"=======================",
			"Random Seed: {{bold}}17{{/}}",
			"",
			"Will run {{bold}}15{{/}} of {{bold}}20{{/}} specs",
			"Running in parallel across {{bold}}3{{/}} nodes",
			"",
		),
		Entry("when succinct and in series",
			C(Succinct),
			config.GinkgoConfigType{RandomSeed: 17, ParallelTotal: 1},
			types.SuiteSummary{SuiteDescription: "My Suite", NumberOfSpecsThatWillBeRun: 15, NumberOfTotalSpecs: 20},
			"[17] {{bold}}My Suite{{/}} - 15/20 specs ",
		),
		Entry("when succinct and in parallel",
			C(Succinct),
			config.GinkgoConfigType{RandomSeed: 17, ParallelTotal: 3},
			types.SuiteSummary{SuiteDescription: "My Suite", NumberOfSpecsThatWillBeRun: 15, NumberOfTotalSpecs: 20},
			"[17] {{bold}}My Suite{{/}} - 15/20 specs - 3 nodes ",
		),
	)

	DescribeTable("WillRun",
		func(conf config.DefaultReporterConfigType, summary types.Summary, output ...string) {
			reporter := reporters.NewDefaultReporterUnderTest(conf, buf)
			reporter.WillRun(summary)
			verifyExpectedOutput(output)
		},
		Entry("when not verbose, it emits nothing", C(), S(CTS("A"), CLS(cl0))),
		Entry("pending specs are not emitted", C(Verbose), S(types.SpecStatePending)),
		Entry("skipped specs are not emitted", C(Verbose), S(types.SpecStateSkipped)),
		Entry("setup nodes", C(Verbose),
			S(types.NodeTypeBeforeSuite, cl0),
			DELIMITER,
			"{{bold}}[BeforeSuite]{{/}}",
			"{{gray}}"+cl0.String()+"{{/}}",
			"",
		),
		Entry("top-level it nodes", C(Verbose),
			S(CTS("My Test"), CLS(cl0)),
			DELIMITER,
			"{{bold}}My Test{{/}}",
			"{{gray}}"+cl0.String()+"{{/}}",
			"",
		),
		Entry("nested it nodes", C(Verbose),
			S(CTS("Container", "Nested Container", "My Test"), CLS(cl0, cl1, cl2)),
			DELIMITER,
			"{{/}}Container {{gray}}Nested Container{{/}}",
			"  {{bold}}My Test{{/}}",
			"  {{gray}}"+cl2.String()+"{{/}}",
			"",
		),
	)

	DescribeTable("DidRun",
		func(conf config.DefaultReporterConfigType, summary types.Summary, output ...string) {
			reporter := reporters.NewDefaultReporterUnderTest(conf, buf)
			reporter.DidRun(summary)
			verifyExpectedOutput(output)
		},
		// Passing Tests
		Entry("a passing test",
			C(),
			S(CTS("A"), CLS(cl0)),
			"{{green}}"+DENOTER+"{{/}}",
		),
		Entry("a passing test that was retried",
			C(),
			S(CTS("A", "B"), CLS(cl0, cl1), 2),
			DELIMITER,
			"{{green}}"+RETRY_DENOTER+" [FLAKEY TEST - TOOK 2 ATTEMPTS TO PASS] [1.000 seconds]{{/}}",
			"{{/}}A {{gray}}B{{/}}",
			"{{gray}}"+cl1.String()+"{{/}}",
			DELIMITER,
			"",
		),
		Entry("a passing test that has ginkgo writer output",
			C(),
			S(CTS("A"), CLS(cl0), GW("GINKGO-WRITER-OUTPUT")),
			"{{green}}"+DENOTER+"{{/}}",
		),
		Entry("a passing test that has ginkgo writer output, with ReportPassed configured",
			C(ReportPassed),
			S(CTS("A", "B"), CLS(cl0, cl1), GW("GINKGO-WRITER-OUTPUT\nSHOULD EMIT")),
			DELIMITER,
			"{{green}}"+DENOTER+" [1.000 seconds]{{/}}",
			"{{/}}A {{gray}}B{{/}}",
			"{{gray}}"+cl1.String()+"{{/}}",
			"",
			"  {{gray}}Begin Captured GinkgoWriter Output >>{{/}}",
			"    GINKGO-WRITER-OUTPUT",
			"    SHOULD EMIT",
			"  {{gray}}<< End Captured GinkgoWriter Output{{/}}",
			DELIMITER,
			"",
		),
		Entry("a passing test that has ginkgo writer output, with Verbose configured",
			C(Verbose),
			S(CTS("A"), CLS(cl0), GW("GINKGO-WRITER-OUTPUT\nSHOULD EMIT")),
			DELIMITER,
			"{{green}}"+DENOTER+" [1.000 seconds]{{/}}",
			"A",
			"{{gray}}"+cl0.String()+"{{/}}",
			"",
			"  {{gray}}Begin Captured GinkgoWriter Output >>{{/}}",
			"    GINKGO-WRITER-OUTPUT",
			"    SHOULD EMIT",
			"  {{gray}}<< End Captured GinkgoWriter Output{{/}}",
			DELIMITER,
			"",
		),
		Entry("a slow passing test",
			C(),
			S(CTS("A", "B"), CLS(cl0, cl1), time.Minute, GW("GINKGO-WRITER-OUTPUT")),
			DELIMITER,
			"{{green}}"+DENOTER+" [SLOW TEST] [60.000 seconds]{{/}}",
			"{{/}}A {{gray}}B{{/}}",
			"{{gray}}"+cl1.String()+"{{/}}",
			DELIMITER,
			"",
		),
		Entry("a passing test with captured stdout",
			C(),
			S(CTS("A", "B"), CLS(cl0, cl1), GW("GINKGO-WRITER-OUTPUT"), STD("STD-OUTPUT\nSHOULD EMIT")),
			DELIMITER,
			"{{green}}"+DENOTER+" [1.000 seconds]{{/}}",
			"{{/}}A {{gray}}B{{/}}",
			"{{gray}}"+cl1.String()+"{{/}}",
			"",
			"  {{gray}}Begin Captured StdOut/StdErr Output >>{{/}}",
			"    STD-OUTPUT",
			"    SHOULD EMIT",
			"  {{gray}}<< End Captured StdOut/StdErr Output{{/}}",
			DELIMITER,
			"",
		),
		Entry("a passing suite setup emits nothing",
			C(),
			S(types.NodeTypeBeforeSuite, cl0, GW("GINKGO-WRITER-OUTPUT")),
		),
		Entry("a passing suite setup with verbose always emits",
			C(Verbose),
			S(types.NodeTypeBeforeSuite, cl0, GW("GINKGO-WRITER-OUTPUT")),
			DELIMITER,
			"{{green}}[BeforeSuite] PASSED [1.000 seconds]{{/}}",
			"{{green}}{{bold}}[BeforeSuite]{{/}}",
			"{{gray}}"+cl0.String()+"{{/}}",
			"",
			"  {{gray}}Begin Captured GinkgoWriter Output >>{{/}}",
			"    GINKGO-WRITER-OUTPUT",
			"  {{gray}}<< End Captured GinkgoWriter Output{{/}}",
			DELIMITER,
			"",
		),
		Entry("a passing suite setup with captured stdout always emits",
			C(),
			S(types.NodeTypeBeforeSuite, cl0, STD("STD-OUTPUT")),
			DELIMITER,
			"{{green}}[BeforeSuite] PASSED [1.000 seconds]{{/}}",
			"{{green}}{{bold}}[BeforeSuite]{{/}}",
			"{{gray}}"+cl0.String()+"{{/}}",
			"",
			"  {{gray}}Begin Captured StdOut/StdErr Output >>{{/}}",
			"    STD-OUTPUT",
			"  {{gray}}<< End Captured StdOut/StdErr Output{{/}}",
			DELIMITER,
			"",
		),

		// Pending Tests
		Entry("a pending test when succinct",
			C(Succinct),
			S(CTS("A"), CLS(cl0), types.SpecStatePending, GW("GW-OUTPUT"), STD("STD-OUTPUT")),
			"{{yellow}}P{{/}}",
		),
		Entry("a pending test normally",
			C(),
			S(CTS("A"), CLS(cl0), types.SpecStatePending, GW("GW-OUTPUT")),
			DELIMITER,
			"{{yellow}}P [PENDING]{{/}}",
			"{{/}}A{{/}}",
			"{{gray}}cl0.go:12{{/}}",
			DELIMITER,
			"",
		),
		Entry("a pending test when verbose",
			C(Verbose),
			S(CTS("A", "B"), CLS(cl0, cl1), types.SpecStatePending, GW("GW-OUTPUT"), STD("STD-OUTPUT")),
			DELIMITER,
			"{{yellow}}P [PENDING]{{/}}",
			"A",
			"{{gray}}"+cl0.String()+"{{/}}",
			"  B",
			"  {{gray}}"+cl1.String()+"{{/}}",
			"",
			"  {{gray}}Begin Captured StdOut/StdErr Output >>{{/}}",
			"    STD-OUTPUT",
			"  {{gray}}<< End Captured StdOut/StdErr Output{{/}}",
			DELIMITER,
			"",
		),
		// Skipped Tests
		Entry("a skipped test when succinct",
			C(Succinct),
			S(CTS("A"), CLS(cl0), types.SpecStateSkipped, GW("GW-OUTPUT"), STD("STD-OUTPUT"),
				F("user skipped"),
			),
			"{{cyan}}S{{/}}",
		),
		Entry("a skipped test without a failure message",
			C(),
			S(CTS("A"), CLS(cl0), types.SpecStateSkipped, GW("GW-OUTPUT")),
			"{{cyan}}S{{/}}",
		),
		Entry("a skipped test with a failure message and verbose is *not* configured",
			C(),
			S(CTS("A", "B"), CLS(cl0, cl1), types.SpecStateSkipped, GW("GW-OUTPUT"), STD("STD-OUTPUT"),
				F("user skipped", types.NodeTypeIt, Location(cl2), 1),
			),
			DELIMITER,
			"{{cyan}}S [SKIPPED] [1.000 seconds]{{/}}",
			"{{/}}A {{gray}}{{cyan}}{{bold}}[It] B{{/}}{{/}}",
			"{{gray}}"+cl2.String()+"{{/}}",
			"",
			"  {{gray}}Begin Captured StdOut/StdErr Output >>{{/}}",
			"    STD-OUTPUT",
			"  {{gray}}<< End Captured StdOut/StdErr Output{{/}}",
			"",
			"  {{gray}}Begin Captured GinkgoWriter Output >>{{/}}",
			"    GW-OUTPUT",
			"  {{gray}}<< End Captured GinkgoWriter Output{{/}}",
			"",
			"  {{cyan}}user skipped{{/}}",
			"  {{cyan}}In {{bold}}[It]{{/}}{{cyan}} at: {{bold}}"+cl2.String()+"{{/}}",
			DELIMITER,
			"",
		),
		Entry("a skipped test with a failure message and verbose *is* configured",
			C(Verbose),
			S(CTS("A", "B"), CLS(cl0, cl1), types.SpecStateSkipped, GW("GW-OUTPUT"), STD("STD-OUTPUT"),
				F("user skipped", types.NodeTypeIt, Location(cl2), 1),
			),
			DELIMITER,
			"{{cyan}}S [SKIPPED] [1.000 seconds]{{/}}",
			"A",
			"{{gray}}"+cl0.String()+"{{/}}",
			"  {{cyan}}{{bold}}B [It]{{/}}",
			"  {{gray}}"+cl1.String()+"{{/}}",
			"",
			"  {{gray}}Begin Captured StdOut/StdErr Output >>{{/}}",
			"    STD-OUTPUT",
			"  {{gray}}<< End Captured StdOut/StdErr Output{{/}}",
			"",
			"  {{gray}}Begin Captured GinkgoWriter Output >>{{/}}",
			"    GW-OUTPUT",
			"  {{gray}}<< End Captured GinkgoWriter Output{{/}}",
			"",
			"  {{cyan}}user skipped{{/}}",
			"  {{cyan}}In {{bold}}[It]{{/}}{{cyan}} at: {{bold}}"+cl2.String()+"{{/}}",
			DELIMITER,
			"",
		),

		//Failed tests
		Entry("when a test has failed in an It",
			C(),
			S(CTS("Describe A", "Context B", "The Test"), CLS(cl0, cl1, cl2),
				types.SpecStateFailed, 2,
				GW("GW-OUTPUT\nIS EMITTED"), STD("STD-OUTPUT\nIS EMITTED"),
				F("FAILURE MESSAGE\nWITH DETAILS", Location(cl3), types.NodeTypeIt, 2),
			),
			DELIMITER,
			"{{red}}"+DENOTER+" [FAILED] [1.000 seconds]{{/}}",
			"Describe A",
			"{{gray}}"+cl0.String()+"{{/}}",
			"  Context B",
			"  {{gray}}"+cl1.String()+"{{/}}",
			"    {{red}}{{bold}}The Test [It]{{/}}",
			"    {{gray}}"+cl2.String()+"{{/}}",
			"",
			"  {{gray}}Begin Captured StdOut/StdErr Output >>{{/}}",
			"    STD-OUTPUT",
			"    IS EMITTED",
			"  {{gray}}<< End Captured StdOut/StdErr Output{{/}}",
			"",
			"  {{gray}}Begin Captured GinkgoWriter Output >>{{/}}",
			"    GW-OUTPUT",
			"    IS EMITTED",
			"  {{gray}}<< End Captured GinkgoWriter Output{{/}}",
			"",
			"  {{red}}FAILURE MESSAGE",
			"  WITH DETAILS{{/}}",
			"  {{red}}In {{bold}}[It]{{/}}{{red}} at: {{bold}}"+cl3.String()+"{{/}}",
			DELIMITER,
			"",
		),
		Entry("when a test has failed in a setup/teardown node",
			C(),
			S(CTS("Describe A", "Context B", "The Test"), CLS(cl0, cl1, cl2),
				types.SpecStateFailed, 2,
				GW("GW-OUTPUT\nIS EMITTED"), STD("STD-OUTPUT\nIS EMITTED"),
				F("FAILURE MESSAGE\nWITH DETAILS", Location(cl3), types.NodeTypeJustBeforeEach, 1),
			),
			DELIMITER,
			"{{red}}"+DENOTER+" [FAILED] [1.000 seconds]{{/}}",
			"Describe A",
			"{{gray}}"+cl0.String()+"{{/}}",
			"  {{red}}{{bold}}Context B [JustBeforeEach]{{/}}",
			"  {{gray}}"+cl1.String()+"{{/}}",
			"    The Test",
			"    {{gray}}"+cl2.String()+"{{/}}",
			"",
			"  {{gray}}Begin Captured StdOut/StdErr Output >>{{/}}",
			"    STD-OUTPUT",
			"    IS EMITTED",
			"  {{gray}}<< End Captured StdOut/StdErr Output{{/}}",
			"",
			"  {{gray}}Begin Captured GinkgoWriter Output >>{{/}}",
			"    GW-OUTPUT",
			"    IS EMITTED",
			"  {{gray}}<< End Captured GinkgoWriter Output{{/}}",
			"",
			"  {{red}}FAILURE MESSAGE",
			"  WITH DETAILS{{/}}",
			"  {{red}}In {{bold}}[JustBeforeEach]{{/}}{{red}} at: {{bold}}"+cl3.String()+"{{/}}",
			DELIMITER,
			"",
		),
		Entry("when a test has failed and Succinct is configured",
			C(Succinct),
			S(CTS("Describe A", "Context B", "The Test"), CLS(cl0, cl1, cl2),
				types.SpecStateFailed, 2,
				GW("GW-OUTPUT\nIS EMITTED"), STD("STD-OUTPUT\nIS EMITTED"),
				F("FAILURE MESSAGE\nWITH DETAILS", Location(cl3), types.NodeTypeJustBeforeEach, 1),
			),
			DELIMITER,
			"{{red}}"+DENOTER+" [FAILED] [1.000 seconds]{{/}}",
			"{{/}}Describe A {{gray}}{{red}}{{bold}}[JustBeforeEach] Context B{{/}} {{/}}The Test{{/}}",
			"{{gray}}"+cl3.String()+"{{/}}",
			"",
			"  {{gray}}Begin Captured StdOut/StdErr Output >>{{/}}",
			"    STD-OUTPUT",
			"    IS EMITTED",
			"  {{gray}}<< End Captured StdOut/StdErr Output{{/}}",
			"",
			"  {{gray}}Begin Captured GinkgoWriter Output >>{{/}}",
			"    GW-OUTPUT",
			"    IS EMITTED",
			"  {{gray}}<< End Captured GinkgoWriter Output{{/}}",
			"",
			"  {{red}}FAILURE MESSAGE",
			"  WITH DETAILS{{/}}",
			"  {{red}}In {{bold}}[JustBeforeEach]{{/}}{{red}} at: {{bold}}"+cl3.String()+"{{/}}",
			DELIMITER,
			"",
		),
		Entry("when a test has failed and FullTrace is configured",
			C(FullTrace),
			S(CTS("Describe A", "Context B", "The Test"), CLS(cl0, cl1, cl2),
				types.SpecStateFailed, 2,
				GW("GW-OUTPUT\nIS EMITTED"), STD("STD-OUTPUT\nIS EMITTED"),
				F("FAILURE MESSAGE\nWITH DETAILS", Location(cl3), types.NodeTypeJustBeforeEach, 1),
			),
			DELIMITER,
			"{{red}}"+DENOTER+" [FAILED] [1.000 seconds]{{/}}",
			"Describe A",
			"{{gray}}"+cl0.String()+"{{/}}",
			"  {{red}}{{bold}}Context B [JustBeforeEach]{{/}}",
			"  {{gray}}"+cl1.String()+"{{/}}",
			"    The Test",
			"    {{gray}}"+cl2.String()+"{{/}}",
			"",
			"  {{gray}}Begin Captured StdOut/StdErr Output >>{{/}}",
			"    STD-OUTPUT",
			"    IS EMITTED",
			"  {{gray}}<< End Captured StdOut/StdErr Output{{/}}",
			"",
			"  {{gray}}Begin Captured GinkgoWriter Output >>{{/}}",
			"    GW-OUTPUT",
			"    IS EMITTED",
			"  {{gray}}<< End Captured GinkgoWriter Output{{/}}",
			"",
			"  {{red}}FAILURE MESSAGE",
			"  WITH DETAILS{{/}}",
			"  {{red}}In {{bold}}[JustBeforeEach]{{/}}{{red}} at: {{bold}}"+cl3.String()+"{{/}}",
			"",
			"  {{red}}Full Stack Trace{{/}}",
			"    full-trace",
			"    cl-3",
			DELIMITER,
			"",
		),
		Entry("when a suite setup node has failed",
			C(),
			S(types.NodeTypeSynchronizedBeforeSuite, cl0, types.SpecStateFailed, 1,
				GW("GW-OUTPUT\nIS EMITTED"), STD("STD-OUTPUT\nIS EMITTED"),
				F("FAILURE MESSAGE\nWITH DETAILS", Location(cl3), types.NodeTypeSynchronizedBeforeSuite, 0),
			),
			DELIMITER,
			"{{red}}[SynchronizedBeforeSuite] [FAILED] [1.000 seconds]{{/}}",
			"{{red}}{{bold}}[SynchronizedBeforeSuite]{{/}}",
			"{{gray}}"+cl3.String()+"{{/}}",
			"",
			"  {{gray}}Begin Captured StdOut/StdErr Output >>{{/}}",
			"    STD-OUTPUT",
			"    IS EMITTED",
			"  {{gray}}<< End Captured StdOut/StdErr Output{{/}}",
			"",
			"  {{gray}}Begin Captured GinkgoWriter Output >>{{/}}",
			"    GW-OUTPUT",
			"    IS EMITTED",
			"  {{gray}}<< End Captured GinkgoWriter Output{{/}}",
			"",
			"  {{red}}FAILURE MESSAGE",
			"  WITH DETAILS{{/}}",
			"  {{red}}In {{bold}}[SynchronizedBeforeSuite]{{/}}{{red}} at: {{bold}}"+cl3.String()+"{{/}}",
			DELIMITER,
			"",
		),

		Entry("when a test has panicked and there is no forwarded panic",
			C(),
			S(CTS("Describe A", "Context B", "The Test"), CLS(cl0, cl1, cl2),
				types.SpecStatePanicked, 2,
				GW("GW-OUTPUT\nIS EMITTED"), STD("STD-OUTPUT\nIS EMITTED"),
				F("FAILURE MESSAGE\nWITH DETAILS", Location(cl3), types.NodeTypeJustBeforeEach, 1),
			),
			DELIMITER,
			"{{magenta}}"+DENOTER+"! [PANICKED] [1.000 seconds]{{/}}",
			"Describe A",
			"{{gray}}"+cl0.String()+"{{/}}",
			"  {{magenta}}{{bold}}Context B [JustBeforeEach]{{/}}",
			"  {{gray}}"+cl1.String()+"{{/}}",
			"    The Test",
			"    {{gray}}"+cl2.String()+"{{/}}",
			"",
			"  {{gray}}Begin Captured StdOut/StdErr Output >>{{/}}",
			"    STD-OUTPUT",
			"    IS EMITTED",
			"  {{gray}}<< End Captured StdOut/StdErr Output{{/}}",
			"",
			"  {{gray}}Begin Captured GinkgoWriter Output >>{{/}}",
			"    GW-OUTPUT",
			"    IS EMITTED",
			"  {{gray}}<< End Captured GinkgoWriter Output{{/}}",
			"",
			"  {{magenta}}FAILURE MESSAGE",
			"  WITH DETAILS{{/}}",
			"  {{magenta}}In {{bold}}[JustBeforeEach]{{/}}{{magenta}} at: {{bold}}"+cl3.String()+"{{/}}",
			DELIMITER,
			"",
		),
		Entry("when a test has panicked and there is a forwarded panic",
			C(),
			S(CTS("Describe A", "Context B", "The Test"), CLS(cl0, cl1, cl2),
				types.SpecStatePanicked, 2,
				GW("GW-OUTPUT\nIS EMITTED"), STD("STD-OUTPUT\nIS EMITTED"),
				F("FAILURE MESSAGE\nWITH DETAILS", Location(cl3), types.NodeTypeJustBeforeEach, 1, ForwardedPanic("the panic\nthusly forwarded")),
			),
			DELIMITER,
			"{{magenta}}"+DENOTER+"! [PANICKED] [1.000 seconds]{{/}}",
			"Describe A",
			"{{gray}}"+cl0.String()+"{{/}}",
			"  {{magenta}}{{bold}}Context B [JustBeforeEach]{{/}}",
			"  {{gray}}"+cl1.String()+"{{/}}",
			"    The Test",
			"    {{gray}}"+cl2.String()+"{{/}}",
			"",
			"  {{gray}}Begin Captured StdOut/StdErr Output >>{{/}}",
			"    STD-OUTPUT",
			"    IS EMITTED",
			"  {{gray}}<< End Captured StdOut/StdErr Output{{/}}",
			"",
			"  {{gray}}Begin Captured GinkgoWriter Output >>{{/}}",
			"    GW-OUTPUT",
			"    IS EMITTED",
			"  {{gray}}<< End Captured GinkgoWriter Output{{/}}",
			"",
			"  {{magenta}}FAILURE MESSAGE",
			"  WITH DETAILS{{/}}",
			"  {{magenta}}In {{bold}}[JustBeforeEach]{{/}}{{magenta}} at: {{bold}}"+cl3.String()+"{{/}}",
			"",
			"  {{magenta}}the panic",
			"  thusly forwarded{{/}}",
			"",
			"  {{magenta}}Full Stack Trace{{/}}",
			"    full-trace",
			"    cl-3",
			DELIMITER,
			"",
		),

		Entry("when a test is interrupted",
			C(),
			S(CTS("Describe A", "Context B", "The Test"), CLS(cl0, cl1, cl2),
				types.SpecStateInterrupted, 2,
				GW("GW-OUTPUT\nIS EMITTED"), STD("STD-OUTPUT\nIS EMITTED"),
				F("FAILURE MESSAGE\nWITH DETAILS", Location(cl3), types.NodeTypeJustBeforeEach, 1),
			),
			DELIMITER,
			"{{orange}}"+DENOTER+"! [INTERRUPTED] [1.000 seconds]{{/}}",
			"Describe A",
			"{{gray}}"+cl0.String()+"{{/}}",
			"  {{orange}}{{bold}}Context B [JustBeforeEach]{{/}}",
			"  {{gray}}"+cl1.String()+"{{/}}",
			"    The Test",
			"    {{gray}}"+cl2.String()+"{{/}}",
			"",
			"  {{gray}}Begin Captured StdOut/StdErr Output >>{{/}}",
			"    STD-OUTPUT",
			"    IS EMITTED",
			"  {{gray}}<< End Captured StdOut/StdErr Output{{/}}",
			"",
			"  {{gray}}Begin Captured GinkgoWriter Output >>{{/}}",
			"    GW-OUTPUT",
			"    IS EMITTED",
			"  {{gray}}<< End Captured GinkgoWriter Output{{/}}",
			"",
			"  {{orange}}FAILURE MESSAGE",
			"  WITH DETAILS{{/}}",
			"  {{orange}}In {{bold}}[JustBeforeEach]{{/}}{{orange}} at: {{bold}}"+cl3.String()+"{{/}}",
			DELIMITER,
			"",
		),
	)

	DescribeTable("Rendering SpecSuiteDidEnd",
		func(conf config.DefaultReporterConfigType, summaries []types.Summary, summary types.SuiteSummary, expected ...string) {
			reporter := reporters.NewDefaultReporterUnderTest(conf, buf)
			for _, summary := range summaries {
				reporter.WillRun(summary)
				reporter.DidRun(summary)
			}
			buf.Clear()
			reporter.SpecSuiteDidEnd(summary)
			verifyExpectedOutput(expected)
		},

		Entry("when configured to be succinct",
			C(Succinct),
			[]types.Summary{S()},
			types.SuiteSummary{
				SuiteSucceeded: true,
				RunTime:        time.Minute,
			},
			" {{green}}SUCCESS!{{/}} 1m0s ",
		),
		Entry("the suite passes",
			C(),
			[]types.Summary{S()},
			types.SuiteSummary{
				SuiteSucceeded:             true,
				RunTime:                    time.Minute,
				NumberOfSpecsThatWillBeRun: 20,
				NumberOfTotalSpecs:         30,
				NumberOfPassedSpecs:        18,
				NumberOfSkippedSpecs:       12,
				NumberOfPendingSpecs:       10,
				NumberOfFailedSpecs:        0,
				NumberOfFlakedSpecs:        0,
			},
			"",
			"",
			"{{green}}{{bold}}Ran 18 of 30 Specs in 60.000 seconds{{/}}",
			"{{green}}{{bold}}SUCCESS!{{/}} -- {{green}}{{bold}}18 Passed{{/}} | {{red}}{{bold}}0 Failed{{/}} | {{yellow}}{{bold}}10 Pending{{/}} | {{cyan}}{{bold}}12 Skipped{{/}}",
			"",
		),
		Entry("the suite passes and has flaky specs",
			C(),
			[]types.Summary{S()},
			types.SuiteSummary{
				SuiteSucceeded:             true,
				RunTime:                    time.Minute,
				NumberOfSpecsThatWillBeRun: 20,
				NumberOfTotalSpecs:         30,
				NumberOfPassedSpecs:        18,
				NumberOfSkippedSpecs:       12,
				NumberOfPendingSpecs:       10,
				NumberOfFailedSpecs:        0,
				NumberOfFlakedSpecs:        4,
			},
			"",
			"",
			"{{green}}{{bold}}Ran 18 of 30 Specs in 60.000 seconds{{/}}",
			"{{green}}{{bold}}SUCCESS!{{/}} -- {{green}}{{bold}}18 Passed{{/}} | {{red}}{{bold}}0 Failed{{/}} | {{light-yellow}}{{bold}}4 Flaked{{/}} | {{yellow}}{{bold}}10 Pending{{/}} | {{cyan}}{{bold}}12 Skipped{{/}}",
			"",
		),
		Entry("the suite fails with one failed test",
			C(),
			[]types.Summary{S(CTS("Describe A", "Context B", "The Test"), CLS(cl0, cl1, cl2),
				types.SpecStateFailed, 2,
				F("FAILURE MESSAGE\nWITH DETAILS", Location(cl3), types.NodeTypeJustBeforeEach, 1),
			)},
			types.SuiteSummary{
				SuiteSucceeded:             false,
				RunTime:                    time.Minute,
				NumberOfSpecsThatWillBeRun: 20,
				NumberOfTotalSpecs:         30,
				NumberOfPassedSpecs:        17,
				NumberOfSkippedSpecs:       12,
				NumberOfPendingSpecs:       10,
				NumberOfFailedSpecs:        1,
				NumberOfFlakedSpecs:        4,
			},
			"",
			"{{red}}{{bold}}Ran 18 of 30 Specs in 60.000 seconds{{/}}",
			"{{red}}{{bold}}FAIL!{{/}} -- {{green}}{{bold}}17 Passed{{/}} | {{red}}{{bold}}1 Failed{{/}} | {{light-yellow}}{{bold}}4 Flaked{{/}} | {{yellow}}{{bold}}10 Pending{{/}} | {{cyan}}{{bold}}12 Skipped{{/}}",
			"",
		),
		Entry("the suite fails with multiple failed tests",
			C(),
			[]types.Summary{
				S(CTS("Describe A", "Context B", "The Test"), CLS(cl0, cl1, cl2),
					types.SpecStateFailed, 2,
					F("FAILURE MESSAGE\nWITH DETAILS", Location(cl3), types.NodeTypeJustBeforeEach, 1),
				),
				S(CTS("Describe A", "The Test"), CLS(cl0, cl1),
					types.SpecStatePanicked, 2,
					F("FAILURE MESSAGE\nWITH DETAILS", Location(cl2), types.NodeTypeIt, 1),
				),
				S(CTS("The Test"), CLS(cl0),
					types.SpecStateInterrupted, 2,
					F("FAILURE MESSAGE\nWITH DETAILS", Location(cl1), types.NodeTypeIt, 0),
				),
			},
			types.SuiteSummary{
				SuiteSucceeded:             false,
				RunTime:                    time.Minute,
				NumberOfSpecsThatWillBeRun: 20,
				NumberOfTotalSpecs:         30,
				NumberOfPassedSpecs:        15,
				NumberOfSkippedSpecs:       12,
				NumberOfPendingSpecs:       10,
				NumberOfFailedSpecs:        3,
				NumberOfFlakedSpecs:        4,
			},
			"",
			"",
			"{{red}}{{bold}}Summarizing 3 Failures:{{/}}",
			"  {{red}}[FAIL]{{/}} {{/}}Describe A {{gray}}{{red}}{{bold}}[JustBeforeEach] Context B{{/}} {{/}}The Test{{/}}",
			"  {{gray}}"+cl3.String()+"{{/}}",
			"  {{magenta}}[PANICKED!]{{/}} {{/}}Describe A {{gray}}{{magenta}}{{bold}}[It] The Test{{/}}{{/}}",
			"  {{gray}}"+cl2.String()+"{{/}}",
			"  {{orange}}[INTERRUPTED]{{/}} {{/}}{{orange}}{{bold}}[It] The Test{{/}}{{/}}",
			"  {{gray}}"+cl1.String()+"{{/}}",
			"",
			"{{red}}{{bold}}Ran 18 of 30 Specs in 60.000 seconds{{/}}",
			"{{red}}{{bold}}FAIL!{{/}} -- {{green}}{{bold}}15 Passed{{/}} | {{red}}{{bold}}3 Failed{{/}} | {{light-yellow}}{{bold}}4 Flaked{{/}} | {{yellow}}{{bold}}10 Pending{{/}} | {{cyan}}{{bold}}12 Skipped{{/}}",
			"",
		),
		Entry("the suite fails with failed suite setups",
			C(),
			[]types.Summary{
				S(types.NodeTypeBeforeSuite, cl0, types.SpecStateFailed, 2,
					F("FAILURE MESSAGE\nWITH DETAILS", Location(cl1), types.NodeTypeBeforeSuite, 0),
				),
				S(types.NodeTypeAfterSuite, cl2, types.SpecStateFailed, 2,
					F("FAILURE MESSAGE\nWITH DETAILS", Location(cl3), types.NodeTypeAfterSuite, 0),
				),
			},
			types.SuiteSummary{
				SuiteSucceeded:             false,
				RunTime:                    time.Minute,
				NumberOfSpecsThatWillBeRun: 20,
				NumberOfTotalSpecs:         30,
				NumberOfPassedSpecs:        0,
				NumberOfSkippedSpecs:       30,
				NumberOfPendingSpecs:       10,
				NumberOfFailedSpecs:        0,
				NumberOfFlakedSpecs:        0,
			},
			"",
			"",
			"{{red}}{{bold}}Summarizing 2 Failures:{{/}}",
			"  {{red}}[FAIL]{{/}} {{red}}{{bold}}[BeforeSuite]{{/}}",
			"  {{gray}}"+cl1.String()+"{{/}}",
			"  {{red}}[FAIL]{{/}} {{red}}{{bold}}[AfterSuite]{{/}}",
			"  {{gray}}"+cl3.String()+"{{/}}",
			"",
			"{{red}}{{bold}}Ran 0 of 30 Specs in 60.000 seconds{{/}}",
			"{{red}}{{bold}}FAIL!{{/}} -- {{green}}{{bold}}0 Passed{{/}} | {{red}}{{bold}}0 Failed{{/}} | {{yellow}}{{bold}}10 Pending{{/}} | {{cyan}}{{bold}}30 Skipped{{/}}",
			"",
		),
	)

	Describe("with failOnPending set to true", func() {
		It("notifies the user when the suite failed due to pending tests", func() {
			reporter := reporters.NewDefaultReporterUnderTest(C(), buf)
			reporter.SpecSuiteWillBegin(config.GinkgoConfigType{
				FailOnPending: true,
			}, types.SuiteSummary{SuiteDescription: "My Suite", NumberOfSpecsThatWillBeRun: 20, NumberOfTotalSpecs: 20})
			buf.Clear()
			reporter.SpecSuiteDidEnd(types.SuiteSummary{
				SuiteSucceeded:             false,
				RunTime:                    time.Minute,
				NumberOfSpecsThatWillBeRun: 20,
				NumberOfTotalSpecs:         20,
				NumberOfPassedSpecs:        19,
				NumberOfSkippedSpecs:       0,
				NumberOfPendingSpecs:       1,
				NumberOfFailedSpecs:        0,
				NumberOfFlakedSpecs:        0,
			})
			verifyExpectedOutput([]string{
				"",
				"{{yellow}}{{bold}}Ran 19 of 20 Specs in 60.000 seconds{{/}}",
				"{{yellow}}{{bold}}FAIL! - Detected pending specs and --fail-on-pending is set{{/}} -- {{green}}{{bold}}19 Passed{{/}} | {{red}}{{bold}}0 Failed{{/}} | {{yellow}}{{bold}}1 Pending{{/}} | {{cyan}}{{bold}}0 Skipped{{/}}",
				"",
			})
		})
	})
})

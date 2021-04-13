/*
Ginkgo is a BDD-style testing framework for Golang

The godoc documentation describes Ginkgo's API.  More comprehensive documentation (with examples!) is available at http://onsi.github.io/ginkgo/

Ginkgo's preferred matcher library is [Gomega](http://github.com/onsi/gomega)

Ginkgo on Github: http://github.com/onsi/ginkgo

Ginkgo is MIT-Licensed
*/
package ginkgo

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/formatter"
	"github.com/onsi/ginkgo/internal"
	"github.com/onsi/ginkgo/internal/global"
	"github.com/onsi/ginkgo/internal/parallel_support"
	"github.com/onsi/ginkgo/internal/testingtproxy"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
)

const GINKGO_VERSION = config.VERSION

var flagSet config.GinkgoFlagSet
var deprecationTracker = types.NewDeprecationTracker()

var suiteDidRun = false

func init() {
	var err error
	flagSet, err = config.BuildTestSuiteFlagSet()
	if err != nil {
		panic(err)
	}
	GinkgoWriter = internal.NewWriter(os.Stdout)
}

type GinkgoWriterInterface interface {
	io.Writer

	Print(a ...interface{})
	Printf(format string, a ...interface{})
	Println(a ...interface{})

	TeeTo(writer io.Writer)
	ClearTeeWriters()
}

//GinkgoWriter implements a GinkgoWriterInterface and io.Writer
//When running in verbose mode any writes to GinkgoWriter will be immediately printed
//to stdout.  Otherwise, GinkgoWriter will buffer any writes produced during the current test and flush them to screen
//only if the current test fails.
//
//GinkgoWriter also provides convenience `Print`, `Printf` and `Println` methods.  Running `GinkgoWriter.Print*(...)` is equivalent to `fmt.Fprint*(GinkgoWriter, ...)`
//
//GinkgoWriter also allows you to tee to a custom writer via `GinkgoWriter.TeeTo(writer)`.  Once registered via `TeeTo`, the `writer` will receive _any_ data
//You can unregister all Tee'd Writers with `GinkgoWRiter.ClearTeeWriters()`
//written to `GinkgoWriter` regardless of whether the test succeeded or failed.
var GinkgoWriter GinkgoWriterInterface

//The interface by which Ginkgo receives *testing.T
type GinkgoTestingT interface {
	Fail()
}

//GinkgoRandomSeed returns the seed used to randomize spec execution order.  It is
//useful for seeding your own pseudorandom number generators (PRNGs) to ensure
//consistent executions from run to run, where your tests contain variability (for
//example, when selecting random test data).
func GinkgoRandomSeed() int64 {
	return config.GinkgoConfig.RandomSeed
}

//GinkgoParallelNode returns the parallel node number for the current ginkgo process
//The node number is 1-indexed
func GinkgoParallelNode() int {
	return config.GinkgoConfig.ParallelNode
}

//Some matcher libraries or legacy codebases require a *testing.T
//GinkgoT implements an interface analogous to *testing.T and can be used if
//the library in question accepts *testing.T through an interface
//
// For example, with testify:
// assert.Equal(GinkgoT(), 123, 123, "they should be equal")
//
// Or with gomock:
// gomock.NewController(GinkgoT())
//
// GinkgoT() takes an optional offset argument that can be used to get the
// correct line number associated with the failure.
func GinkgoT(optionalOffset ...int) GinkgoTInterface {
	offset := 3
	if len(optionalOffset) > 0 {
		offset = optionalOffset[0]
	}
	failedFunc := func() bool {
		return CurrentSpecReport().Failed()
	}
	nameFunc := func() string {
		return CurrentSpecReport().FullText()
	}
	return testingtproxy.New(GinkgoWriter, Fail, Skip, failedFunc, nameFunc, offset)
}

//The interface returned by GinkgoT().  This covers most of the methods
//in the testing package's T.
type GinkgoTInterface interface {
	Cleanup(func())
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Helper()
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Name() string
	Parallel()
	Skip(args ...interface{})
	SkipNow()
	Skipf(format string, args ...interface{})
	Skipped() bool
	TempDir() string
}

type SpecReport = types.SpecReport

// CurrentSpecReport returns information about the current running test.
// The returned object is a types.SpecReport which includes helper methods
// to make extracting information about the test easier.
func CurrentSpecReport() SpecReport {
	return global.Suite.CurrentSpecReport()
}

//RunSpecs is the entry point for the Ginkgo test runner.
//You must call this within a Golang testing TestX(t *testing.T) function.
//
//To bootstrap a test suite you can use the Ginkgo CLI:
//
//	ginkgo bootstrap
func RunSpecs(t GinkgoTestingT, description string) bool {
	if suiteDidRun {
		exitIfErr(types.GinkgoErrors.RerunningSuite())
	}
	suiteDidRun = true

	var reporter reporters.Reporter
	var outputInterceptor internal.OutputInterceptor
	if config.GinkgoConfig.ParallelTotal == 1 {
		reporter = reporters.NewDefaultReporter(config.DefaultReporterConfig, formatter.ColorableStdOut)
		outputInterceptor = internal.NewNoopOutputInterceptor()
	} else {
		reporter = parallel_support.NewForwardingReporter(config.DefaultReporterConfig, config.GinkgoConfig.ParallelHost, GinkgoWriter.(*internal.Writer))
		outputInterceptor = internal.NewOutputInterceptor()
	}

	writer := GinkgoWriter.(*internal.Writer)
	if config.DefaultReporterConfig.Verbose && config.GinkgoConfig.ParallelTotal == 1 {
		writer.SetMode(internal.WriterModeStreamAndBuffer)
	} else {
		writer.SetMode(internal.WriterModeBufferOnly)
	}

	configErrors := config.VetConfig(flagSet, config.GinkgoConfig, config.DefaultReporterConfig)
	if len(configErrors) > 0 {
		fmt.Fprintf(formatter.ColorableStdErr, formatter.F("{{red}}Ginkgo detected configuration issues:{{/}}\n"))
		for _, err := range configErrors {
			fmt.Fprintf(formatter.ColorableStdErr, err.Error())
		}
		os.Exit(1)
	}

	err := global.Suite.BuildTree()
	exitIfErr(err)

	passed, hasFocusedTests := global.Suite.Run(description, global.Failer, reporter, writer, outputInterceptor, internal.NewInterruptHandler(), config.GinkgoConfig)

	flagSet.ValidateDeprecations(deprecationTracker)
	if deprecationTracker.DidTrackDeprecations() {
		fmt.Fprintln(formatter.ColorableStdErr, deprecationTracker.DeprecationsReport())
	}

	if !passed {
		t.Fail()
	}

	if passed && hasFocusedTests && strings.TrimSpace(os.Getenv("GINKGO_EDITOR_INTEGRATION")) == "" {
		fmt.Println("PASS | FOCUSED")
		os.Exit(types.GINKGO_FOCUS_EXIT_CODE)
	}
	return passed
}

//Skip notifies Ginkgo that the current spec was skipped.
func Skip(message string, callerSkip ...int) {
	skip := 0
	if len(callerSkip) > 0 {
		skip = callerSkip[0]
	}
	cl := types.NewCodeLocation(skip + 1)
	global.Failer.Skip(message, cl)
	panic(types.GinkgoErrors.UncaughtGinkgoPanic(cl))
}

//Fail notifies Ginkgo that the current spec has failed. (Gomega will call Fail for you automatically when an assertion fails.)
func Fail(message string, callerSkip ...int) {
	skip := 0
	if len(callerSkip) > 0 {
		skip = callerSkip[0]
	}

	cl := types.NewCodeLocation(skip + 1)
	global.Failer.Fail(message, cl)
	panic(types.GinkgoErrors.UncaughtGinkgoPanic(cl))
}

//GinkgoRecover should be deferred at the top of any spawned goroutine that (may) call `Fail`
//Since Gomega assertions call fail, you should throw a `defer GinkgoRecover()` at the top of any goroutine that
//calls out to Gomega
//
//Here's why: Ginkgo's `Fail` method records the failure and then panics to prevent
//further assertions from running.  This panic must be recovered.  Ginkgo does this for you
//if the panic originates in a Ginkgo node (an It, BeforeEach, etc...)
//
//Unfortunately, if a panic originates on a goroutine *launched* from one of these nodes there's no
//way for Ginkgo to rescue the panic.  To do this, you must remember to `defer GinkgoRecover()` at the top of such a goroutine.
func GinkgoRecover() {
	e := recover()
	if e != nil {
		global.Failer.Panic(types.NewCodeLocation(1), e)
	}
}

// pushNode and pushSuiteNodeBuilder are used by the various test construction DSL methods to push nodes onto the suite
// it handles returned errors, emits a detailed error message to help the user learn what they may have done wrong, then exits
func pushNode(node internal.Node) bool {
	err := global.Suite.PushNode(node)
	exitIfErr(err)
	return true
}

func pushSuiteNodeBuilder(suiteNodeBuilder internal.SuiteNodeBuilder) bool {
	err := global.Suite.PushSuiteNodeBuilder(suiteNodeBuilder)
	exitIfErr(err)
	return true
}

//Describe blocks allow you to organize your specs.  A Describe block can contain any number of
//BeforeEach, AfterEach, JustBeforeEach, and It blocks.
//
//In addition you can nest Describe, Context and When blocks.  Describe, Context and When blocks are functionally
//equivalent.  The difference is purely semantic -- you typically Describe the behavior of an object
//or method and, within that Describe, outline a number of Contexts and Whens.
func Describe(text string, body func()) bool {
	return pushNode(internal.NewNode(types.NodeTypeContainer, text, body, types.NewCodeLocation(1), false, false))
}

//You can focus the tests within a describe block using FDescribe
func FDescribe(text string, body func()) bool {
	return pushNode(internal.NewNode(types.NodeTypeContainer, text, body, types.NewCodeLocation(1), true, false))
}

//You can mark the tests within a describe block as pending using PDescribe
func PDescribe(text string, body func()) bool {
	return pushNode(internal.NewNode(types.NodeTypeContainer, text, body, types.NewCodeLocation(1), false, true))
}

//You can mark the tests within a describe block as pending using XDescribe
func XDescribe(text string, body func()) bool {
	return pushNode(internal.NewNode(types.NodeTypeContainer, text, body, types.NewCodeLocation(1), false, true))
}

//Context blocks allow you to organize your specs.  A Context block can contain any number of
//BeforeEach, AfterEach, JustBeforeEach, and It blocks.
//
//In addition you can nest Describe, Context and When blocks.  Describe, Context and When blocks are functionally
//equivalent.  The difference is purely semantic -- you typical Describe the behavior of an object
//or method and, within that Describe, outline a number of Contexts and Whens.
func Context(text string, body func()) bool {
	return pushNode(internal.NewNode(types.NodeTypeContainer, text, body, types.NewCodeLocation(1), false, false))
}

//You can focus the tests within a describe block using FContext
func FContext(text string, body func()) bool {
	return pushNode(internal.NewNode(types.NodeTypeContainer, text, body, types.NewCodeLocation(1), true, false))
}

//You can mark the tests within a describe block as pending using PContext
func PContext(text string, body func()) bool {
	return pushNode(internal.NewNode(types.NodeTypeContainer, text, body, types.NewCodeLocation(1), false, true))
}

//You can mark the tests within a describe block as pending using XContext
func XContext(text string, body func()) bool {
	return pushNode(internal.NewNode(types.NodeTypeContainer, text, body, types.NewCodeLocation(1), false, true))
}

//When blocks allow you to organize your specs.  A When block can contain any number of
//BeforeEach, AfterEach, JustBeforeEach, and It blocks.
//
//In addition you can nest Describe, Context and When blocks.  Describe, Context and When blocks are functionally
//equivalent.  The difference is purely semantic -- you typical Describe the behavior of an object
//or method and, within that Describe, outline a number of Contexts and Whens.
func When(text string, body func()) bool {
	return pushNode(internal.NewNode(types.NodeTypeContainer, "when "+text, body, types.NewCodeLocation(1), false, false))
}

//You can focus the tests within a describe block using FWhen
func FWhen(text string, body func()) bool {
	return pushNode(internal.NewNode(types.NodeTypeContainer, "when "+text, body, types.NewCodeLocation(1), true, false))
}

//You can mark the tests within a describe block as pending using PWhen
func PWhen(text string, body func()) bool {
	return pushNode(internal.NewNode(types.NodeTypeContainer, "when "+text, body, types.NewCodeLocation(1), false, true))
}

//You can mark the tests within a describe block as pending using XWhen
func XWhen(text string, body func()) bool {
	return pushNode(internal.NewNode(types.NodeTypeContainer, "when "+text, body, types.NewCodeLocation(1), false, true))
}

//It blocks contain your test code and assertions.  You cannot nest any other Ginkgo blocks
//within an It block.
func It(text string, body interface{}, _ ...interface{}) bool {
	cl := types.NewCodeLocation(1)
	return pushNode(internal.NewNode(types.NodeTypeIt, text, validateBodyFunc(body, cl), cl, false, false))
}

//You can focus individual Its using FIt
func FIt(text string, body interface{}, _ ...interface{}) bool {
	cl := types.NewCodeLocation(1)
	return pushNode(internal.NewNode(types.NodeTypeIt, text, validateBodyFunc(body, cl), cl, true, false))
}

//You can mark Its as pending using PIt
func PIt(text string, _ ...interface{}) bool {
	return pushNode(internal.NewNode(types.NodeTypeIt, text, nil, types.NewCodeLocation(1), false, true))
}

//You can mark Its as pending using XIt
func XIt(text string, _ ...interface{}) bool {
	return pushNode(internal.NewNode(types.NodeTypeIt, text, nil, types.NewCodeLocation(1), false, true))
}

//Specify blocks are aliases for It blocks and allow for more natural wording in situations
//which "It" does not fit into a natural sentence flow. All the same protocols apply for Specify blocks
//which apply to It blocks.
func Specify(text string, body interface{}, _ ...interface{}) bool {
	cl := types.NewCodeLocation(1)
	return pushNode(internal.NewNode(types.NodeTypeIt, text, validateBodyFunc(body, cl), cl, false, false))
}

//You can focus individual Specifys using FSpecify
func FSpecify(text string, body interface{}, _ ...interface{}) bool {
	cl := types.NewCodeLocation(1)
	return pushNode(internal.NewNode(types.NodeTypeIt, text, validateBodyFunc(body, cl), cl, true, false))
}

//You can mark Specifys as pending using PSpecify
func PSpecify(text string, _ ...interface{}) bool {
	return pushNode(internal.NewNode(types.NodeTypeIt, text, nil, types.NewCodeLocation(1), false, true))
}

//You can mark Specifys as pending using XSpecify
func XSpecify(text string, _ ...interface{}) bool {
	return pushNode(internal.NewNode(types.NodeTypeIt, text, nil, types.NewCodeLocation(1), false, true))
}

//By allows you to better document large Its.
//
//Generally you should try to keep your Its short and to the point.  This is not always possible, however,
//especially in the context of integration tests that capture a particular workflow.
//
//By allows you to document such flows.  By must be called within a runnable node (It, BeforeEach, etc...)
//By will simply log the passed in text to the GinkgoWriter.  If By is handed a function it will immediately run the function.
func By(text string, callbacks ...func()) {
	preamble := "\x1b[1mSTEP\x1b[0m"
	if config.DefaultReporterConfig.NoColor {
		preamble = "STEP"
	}
	fmt.Fprintln(GinkgoWriter, preamble+": "+text)
	if len(callbacks) == 1 {
		callbacks[0]()
	}
	if len(callbacks) > 1 {
		panic("just one callback per By, please")
	}
}

//BeforeSuite blocks are run just once before any specs are run.  When running in parallel, each
//parallel node process will call BeforeSuite.
//
//You may only register *one* BeforeSuite handler per test suite.  You typically do so in your bootstrap file at the top level.
func BeforeSuite(body func()) bool {
	return pushSuiteNodeBuilder(internal.SuiteNodeBuilder{
		NodeType:        types.NodeTypeBeforeSuite,
		CodeLocation:    types.NewCodeLocation(1),
		BeforeSuiteBody: body,
	})
}

//AfterSuite blocks are *always* run after all the specs regardless of whether specs have passed or failed.
//Moreover, if Ginkgo receives an interrupt signal (^C) it will attempt to run the AfterSuite before exiting.
//
//When running in parallel, each parallel node process will call AfterSuite.
//
//You may only register *one* AfterSuite handler per test suite.  You typically do so in your bootstrap file at the top level.
func AfterSuite(body func()) bool {
	return pushSuiteNodeBuilder(internal.SuiteNodeBuilder{
		NodeType:       types.NodeTypeAfterSuite,
		CodeLocation:   types.NewCodeLocation(1),
		AfterSuiteBody: body,
	})
}

//SynchronizedBeforeSuite blocks are primarily meant to solve the problem of setting up singleton external resources shared across
//nodes when running tests in parallel.  For example, say you have a shared database that you can only start one instance of that
//must be used in your tests.  When running in parallel, only one node should set up the database and all other nodes should wait
//until that node is done before running.
//
//SynchronizedBeforeSuite accomplishes this by taking *two* function arguments.  The first is only run on parallel node #1.  The second is
//run on all nodes, but *only* after the first function completes successfully.  Ginkgo also makes it possible to send data from the first function (on Node 1)
//to the second function (on all the other nodes).
//
//The functions have the following signatures.  The first function (which only runs on node 1) has the signature:
//
//	func() []byte
//
//The byte array returned by the first function is then passed to the second function, which has the signature:
//
//	func(data []byte)
//
//Here's a simple pseudo-code example that starts a shared database on Node 1 and shares the database's address with the other nodes:
//
//	var dbClient db.Client
//	var dbRunner db.Runner
//
//	var _ = SynchronizedBeforeSuite(func() []byte {
//		dbRunner = db.NewRunner()
//		err := dbRunner.Start()
//		Ω(err).ShouldNot(HaveOccurred())
//		return []byte(dbRunner.URL)
//	}, func(data []byte) {
//		dbClient = db.NewClient()
//		err := dbClient.Connect(string(data))
//		Ω(err).ShouldNot(HaveOccurred())
//	})
func SynchronizedBeforeSuite(node1Body func() []byte, allNodesBody func([]byte)) bool {
	return pushSuiteNodeBuilder(internal.SuiteNodeBuilder{
		NodeType:                            types.NodeTypeSynchronizedBeforeSuite,
		CodeLocation:                        types.NewCodeLocation(1),
		SynchronizedBeforeSuiteNode1Body:    node1Body,
		SynchronizedBeforeSuiteAllNodesBody: allNodesBody,
	})
}

//SynchronizedAfterSuite blocks complement the SynchronizedBeforeSuite blocks in solving the problem of setting up
//external singleton resources shared across nodes when running tests in parallel.
//
//SynchronizedAfterSuite accomplishes this by taking *two* function arguments.  The first runs on all nodes.  The second runs only on parallel node #1
//and *only* after all other nodes have finished and exited.  This ensures that node 1, and any resources it is running, remain alive until
//all other nodes are finished.
//
//Here's a pseudo-code example that complements that given in SynchronizedBeforeSuite.  Here, SynchronizedAfterSuite is used to tear down the shared database
//only after all nodes have finished:
//
//	var _ = SynchronizedAfterSuite(func() {
//		dbClient.Cleanup()
//	}, func() {
//		dbRunner.Stop()
//	})
func SynchronizedAfterSuite(allNodesBody func(), node1Body func()) bool {
	return pushSuiteNodeBuilder(internal.SuiteNodeBuilder{
		NodeType:                           types.NodeTypeSynchronizedAfterSuite,
		CodeLocation:                       types.NewCodeLocation(1),
		SynchronizedAfterSuiteAllNodesBody: allNodesBody,
		SynchronizedAfterSuiteNode1Body:    node1Body,
	})
}

//BeforeEach blocks are run before It blocks.  When multiple BeforeEach blocks are defined in nested
//Describe and Context blocks the outermost BeforeEach blocks are run first.
func BeforeEach(body interface{}, _ ...interface{}) bool {
	cl := types.NewCodeLocation(1)
	return pushNode(internal.NewNode(types.NodeTypeBeforeEach, "", validateBodyFunc(body, cl), cl, false, false))
}

//JustBeforeEach blocks are run before It blocks but *after* all BeforeEach blocks.  For more details,
//read the [documentation](http://onsi.github.io/ginkgo/#separating_creation_and_configuration_)
func JustBeforeEach(body interface{}, _ ...interface{}) bool {
	cl := types.NewCodeLocation(1)
	return pushNode(internal.NewNode(types.NodeTypeJustBeforeEach, "", validateBodyFunc(body, cl), cl, false, false))
}

//JustAfterEach blocks are run after It blocks but *before* all AfterEach blocks.  For more details,
//read the [documentation](http://onsi.github.io/ginkgo/#separating_creation_and_configuration_)
func JustAfterEach(body interface{}, _ ...interface{}) bool {
	cl := types.NewCodeLocation(1)
	return pushNode(internal.NewNode(types.NodeTypeJustAfterEach, "", validateBodyFunc(body, cl), cl, false, false))
}

//AfterEach blocks are run after It blocks.   When multiple AfterEach blocks are defined in nested
//Describe and Context blocks the innermost AfterEach blocks are run first.
func AfterEach(body interface{}, _ ...interface{}) bool {
	cl := types.NewCodeLocation(1)
	return pushNode(internal.NewNode(types.NodeTypeAfterEach, "", validateBodyFunc(body, cl), cl, false, false))
}

//ReportAfterEach nodes are run for each test, even if the test is skipped or pending.  ReportAfterEach nodes take a function that
//receives a types.SpecReport.  They are called after the test has completed and are passed in the final report for the test.
func ReportAfterEach(body func(SpecReport)) bool {
	cl := types.NewCodeLocation(1)
	return pushNode(internal.NewReportAfterEachNode(body, cl))
}

func exitIfErr(err error) {
	if err != nil {
		fmt.Fprintln(formatter.ColorableStdErr, err.Error())
		os.Exit(1)
	}
}

// Deprecations for v2

// Deprecated Done Channel for asynchronous testing
type Done chan<- interface{}

func validateBodyFunc(body interface{}, cl types.CodeLocation) func() {
	t := reflect.TypeOf(body)
	if t.Kind() != reflect.Func {
		exitIfErr(types.GinkgoErrors.InvalidBodyType(t, cl))
	}

	if t.NumOut() > 0 {
		exitIfErr(types.GinkgoErrors.InvalidBodyType(t, cl))
	}

	if t.NumIn() == 0 {
		return body.(func())
	}

	if t.NumIn() > 1 {
		exitIfErr(types.GinkgoErrors.InvalidBodyType(t, cl))
	}

	if t.In(0) != reflect.TypeOf(make(Done)) {
		exitIfErr(types.GinkgoErrors.InvalidBodyType(t, cl))
	}

	deprecationTracker.TrackDeprecation(types.Deprecations.Async(), cl)

	return func() {
		body.(func(Done))(make(Done))
	}
}

//Deprecated: Custom Ginkgo test reporters are no longer supported
//Please read the documentation at:
//https://github.com/onsi/ginkgo/blob/v2/docs/MIGRATING_TO_V2.md#removed-custom-reporters
//for Ginkgo's new behavior and for a migration path.
type Reporter = reporters.DeprecatedReporter

//Deprecated: Custom Reporters have been removed in v2.  RunSpecsWithDefaultAndCustomReporters will simply call RunSpecs()
//
//Please read the documentation at:
//https://github.com/onsi/ginkgo/blob/v2/docs/MIGRATING_TO_V2.md#removed-custom-reporters
//for Ginkgo's new behavior and for a migration path.
func RunSpecsWithDefaultAndCustomReporters(t GinkgoTestingT, description string, _ []Reporter) bool {
	deprecationTracker.TrackDeprecation(types.Deprecations.CustomReporter())
	return RunSpecs(t, description)
}

//Deprecated: Custom Reporters have been removed in v2.  RunSpecsWithCustomReporters will simply call RunSpecs()
//
//Please read the documentation at:
//https://github.com/onsi/ginkgo/blob/v2/docs/MIGRATING_TO_V2.md#removed-custom-reporters
//for Ginkgo's new behavior and for a migration path.
func RunSpecsWithCustomReporters(t GinkgoTestingT, description string, _ []Reporter) bool {
	deprecationTracker.TrackDeprecation(types.Deprecations.CustomReporter())
	return RunSpecs(t, description)
}

//GinkgoTestDescription represents the information about the current running test returned by CurrentGinkgoTestDescription
//	FullTestText: a concatenation of ComponentTexts and the TestText
//	ComponentTexts: a list of all texts for the Describes & Contexts leading up to the current test
//	TestText: the text in the It node
//	FileName: the name of the file containing the current test
//	LineNumber: the line number for the current test
//	Failed: if the current test has failed, this will be true (useful in an AfterEach)
//
//Deprecated: Use CurrentSpecReport() instead
type DeprecatedGinkgoTestDescription struct {
	FullTestText   string
	ComponentTexts []string
	TestText       string

	FileName   string
	LineNumber int

	Failed   bool
	Duration time.Duration
}
type GinkgoTestDescription = DeprecatedGinkgoTestDescription

//CurrentGinkgoTestDescripton returns information about the current running test.
//Deprecated: Use CurrentSpecReport() instead
func CurrentGinkgoTestDescription() DeprecatedGinkgoTestDescription {
	deprecationTracker.TrackDeprecation(
		types.Deprecations.CurrentGinkgoTestDescription(),
		types.NewCodeLocation(1),
	)
	report := global.Suite.CurrentSpecReport()
	if report.State == types.SpecStateInvalid {
		return GinkgoTestDescription{}
	}

	return DeprecatedGinkgoTestDescription{
		ComponentTexts: report.NodeTexts,
		FullTestText:   strings.Join(report.NodeTexts, " "),
		TestText:       report.NodeTexts[len(report.NodeTexts)-1],
		FileName:       report.LeafNodeLocation.FileName,
		LineNumber:     report.LeafNodeLocation.LineNumber,
		Failed:         report.State.Is(types.SpecStateFailureStates...),
		Duration:       report.RunTime,
	}
}

//deprecated benchmarker
type Benchmarker interface {
	Time(name string, body func(), info ...interface{}) (elapsedTime time.Duration)
	RecordValue(name string, value float64, info ...interface{})
	RecordValueWithPrecision(name string, value float64, units string, precision int, info ...interface{})
}

//deprecated Measure
func Measure(_ ...interface{}) bool {
	deprecationTracker.TrackDeprecation(types.Deprecations.Measure(), types.NewCodeLocation(1))
	return true
}

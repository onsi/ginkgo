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
	"path/filepath"
	"strings"
	"time"

	"github.com/onsi/ginkgo/formatter"
	"github.com/onsi/ginkgo/internal"
	"github.com/onsi/ginkgo/internal/global"
	"github.com/onsi/ginkgo/internal/interrupt_handler"
	"github.com/onsi/ginkgo/internal/parallel_support"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
)

const GINKGO_VERSION = types.VERSION

var flagSet types.GinkgoFlagSet
var deprecationTracker = types.NewDeprecationTracker()
var suiteConfig = types.NewDefaultSuiteConfig()
var reporterConfig = types.NewDefaultReporterConfig()
var suiteDidRun = false
var outputInterceptor internal.OutputInterceptor
var client parallel_support.Client

func init() {
	var err error
	flagSet, err = types.BuildTestSuiteFlagSet(&suiteConfig, &reporterConfig)
	exitIfErr(err)
	GinkgoWriter = internal.NewWriter(os.Stdout)
}

func exitIfErr(err error) {
	if err != nil {
		if outputInterceptor != nil {
			outputInterceptor.Shutdown()
		}
		if client != nil {
			client.Close()
		}
		fmt.Fprintln(formatter.ColorableStdErr, err.Error())
		os.Exit(1)
	}
}

func exitIfErrors(errors []error) {
	if len(errors) > 0 {
		if outputInterceptor != nil {
			outputInterceptor.Shutdown()
		}
		if client != nil {
			client.Close()
		}
		for _, err := range errors {
			fmt.Fprintln(formatter.ColorableStdErr, err.Error())
		}
		os.Exit(1)
	}
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

//GinkgoConfiguration returns the configuration of the currenty running test suite
func GinkgoConfiguration() (types.SuiteConfig, types.ReporterConfig) {
	return suiteConfig, reporterConfig
}

//GinkgoRandomSeed returns the seed used to randomize spec execution order.  It is
//useful for seeding your own pseudorandom number generators (PRNGs) to ensure
//consistent executions from run to run, where your tests contain variability (for
//example, when selecting random test data).
func GinkgoRandomSeed() int64 {
	return suiteConfig.RandomSeed
}

//GinkgoParallelProcess returns the parallel process number for the current ginkgo process
//The process number is 1-indexed
func GinkgoParallelProcess() int {
	return suiteConfig.ParallelProcess
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

	configErrors := types.VetConfig(flagSet, suiteConfig, reporterConfig)
	if len(configErrors) > 0 {
		fmt.Fprintf(formatter.ColorableStdErr, formatter.F("{{red}}Ginkgo detected configuration issues:{{/}}\n"))
		for _, err := range configErrors {
			fmt.Fprintf(formatter.ColorableStdErr, err.Error())
		}
		os.Exit(1)
	}

	var reporter reporters.Reporter
	if suiteConfig.ParallelTotal == 1 {
		reporter = reporters.NewDefaultReporter(reporterConfig, formatter.ColorableStdOut)
		outputInterceptor = internal.NoopOutputInterceptor{}
		client = nil
	} else {
		reporter = reporters.NoopReporter{}
		outputInterceptor = internal.NewOutputInterceptor()
		if os.Getenv("GINKGO_INTERCEPTOR_MODE") == "SWAP" {
			outputInterceptor = internal.NewOSGlobalReassigningOutputInterceptor()
		} else if os.Getenv("GINKGO_INTERCEPTOR_MODE") == "NONE" {
			outputInterceptor = internal.NoopOutputInterceptor{}
		}
		client = parallel_support.NewClient(suiteConfig.ParallelHost)
		if !client.Connect() {
			client = nil
			exitIfErr(types.GinkgoErrors.UnreachableParallelHost(suiteConfig.ParallelHost))
		}
		defer client.Close()
	}

	writer := GinkgoWriter.(*internal.Writer)
	if reporterConfig.Verbose && suiteConfig.ParallelTotal == 1 {
		writer.SetMode(internal.WriterModeStreamAndBuffer)
	} else {
		writer.SetMode(internal.WriterModeBufferOnly)
	}

	if reporterConfig.WillGenerateReport() {
		registerReportAfterSuiteNodeForAutogeneratedReports(reporterConfig)
	}

	err := global.Suite.BuildTree()
	exitIfErr(err)

	suitePath, err := os.Getwd()
	exitIfErr(err)
	suitePath, err = filepath.Abs(suitePath)
	exitIfErr(err)

	passed, hasFocusedTests := global.Suite.Run(description, suitePath, global.Failer, reporter, writer, outputInterceptor, interrupt_handler.NewInterruptHandler(suiteConfig.Timeout, client), client, suiteConfig)
	outputInterceptor.Shutdown()

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

//Skip instructs Ginkgo to skip the current spec
func Skip(message string, callerSkip ...int) {
	skip := 0
	if len(callerSkip) > 0 {
		skip = callerSkip[0]
	}
	cl := types.NewCodeLocationWithStackTrace(skip + 1)
	global.Failer.Skip(message, cl)
	panic(types.GinkgoErrors.UncaughtGinkgoPanic(cl))
}

//Fail notifies Ginkgo that the current spec has failed. (Gomega will call Fail for you automatically when an assertion fails.)
func Fail(message string, callerSkip ...int) {
	skip := 0
	if len(callerSkip) > 0 {
		skip = callerSkip[0]
	}

	cl := types.NewCodeLocationWithStackTrace(skip + 1)
	global.Failer.Fail(message, cl)
	panic(types.GinkgoErrors.UncaughtGinkgoPanic(cl))
}

//AbortSuite instruct Ginkgo to fail the current test and skip all subsequent tests
func AbortSuite(message string, callerSkip ...int) {
	skip := 0
	if len(callerSkip) > 0 {
		skip = callerSkip[0]
	}

	cl := types.NewCodeLocationWithStackTrace(skip + 1)
	global.Failer.AbortSuite(message, cl)
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
		global.Failer.Panic(types.NewCodeLocationWithStackTrace(1), e)
	}
}

// pushNode is used by the various test construction DSL methods to push nodes onto the suite
// it handles returned errors, emits a detailed error message to help the user learn what they may have done wrong, then exits
func pushNode(node internal.Node, errors []error) bool {
	exitIfErrors(errors)
	exitIfErr(global.Suite.PushNode(node))
	return true
}

//Describe blocks allow you to organize your specs.  A Describe block can contain any number of
//BeforeEach, AfterEach, JustBeforeEach, and It blocks.
//
//In addition you can nest Describe, Context and When blocks.  Describe, Context and When blocks are functionally
//equivalent.  The difference is purely semantic -- you typically Describe the behavior of an object
//or method and, within that Describe, outline a number of Contexts and Whens.
func Describe(text string, args ...interface{}) bool {
	return pushNode(internal.NewNode(deprecationTracker, types.NodeTypeContainer, text, args...))
}

//You can focus the tests within a describe block using FDescribe
func FDescribe(text string, args ...interface{}) bool {
	args = append(args, internal.Focus)
	return pushNode(internal.NewNode(deprecationTracker, types.NodeTypeContainer, text, args...))
}

//You can mark the tests within a describe block as pending using PDescribe
func PDescribe(text string, args ...interface{}) bool {
	args = append(args, internal.Pending)
	return pushNode(internal.NewNode(deprecationTracker, types.NodeTypeContainer, text, args...))
}

//You can mark the tests within a describe block as pending using XDescribe
var XDescribe = PDescribe

//Context blocks allow you to organize your specs.  A Context block can contain any number of
//BeforeEach, AfterEach, JustBeforeEach, and It blocks.
//
//In addition you can nest Describe, Context and When blocks.  Describe, Context and When blocks are functionally
//equivalent.  The difference is purely semantic -- you typical Describe the behavior of an object
//or method and, within that Describe, outline a number of Contexts and Whens.
var Context, FContext, PContext, XContext = Describe, FDescribe, PDescribe, XDescribe

//When blocks allow you to organize your specs.  A When block can contain any number of
//BeforeEach, AfterEach, JustBeforeEach, and It blocks.
//
//In addition you can nest Describe, Context and When blocks.  Describe, Context and When blocks are functionally
//equivalent.  The difference is purely semantic -- you typical Describe the behavior of an object
//or method and, within that Describe, outline a number of Contexts and Whens.
var When, FWhen, PWhen, XWhen = Describe, FDescribe, PDescribe, XDescribe

//It blocks contain your test code and assertions.  You cannot nest any other Ginkgo blocks
//within an It block.
func It(text string, args ...interface{}) bool {
	return pushNode(internal.NewNode(deprecationTracker, types.NodeTypeIt, text, args...))
}

//You can focus individual Its using FIt
func FIt(text string, args ...interface{}) bool {
	args = append(args, internal.Focus)
	return pushNode(internal.NewNode(deprecationTracker, types.NodeTypeIt, text, args...))
}

//You can mark Its as pending using PIt
func PIt(text string, args ...interface{}) bool {
	args = append(args, internal.Pending)
	return pushNode(internal.NewNode(deprecationTracker, types.NodeTypeIt, text, args...))
}

//You can mark Its as pending using XIt
var XIt = PIt

//Specify blocks are aliases for It blocks and allow for more natural wording in situations
//which "It" does not fit into a natural sentence flow. All the same protocols apply for Specify blocks
//which apply to It blocks.
var Specify, FSpecify, PSpecify, XSpecify = It, FIt, PIt, XIt

//By allows you to better document large Its.
//
//Generally you should try to keep your Its short and to the point.  This is not always possible, however,
//especially in the context of integration tests that capture a particular workflow.
//
//By allows you to document such flows.  By must be called within a runnable node (It, BeforeEach, etc...)
//By will simply log the passed in text to the GinkgoWriter.  If By is handed a function it will immediately run the function.
func By(text string, callbacks ...func()) {
	value := struct {
		Text     string
		Duration time.Duration
	}{
		Text: text,
	}
	t := time.Now()
	AddReportEntry("By Step", ReportEntryVisibilityNever, Offset(1), &value, t)
	formatter := formatter.NewWithNoColorBool(reporterConfig.NoColor)
	GinkgoWriter.Println(formatter.F("{{bold}}STEP:{{/}} %s {{gray}}%s{{/}}", text, t.Format(types.GINKGO_TIME_FORMAT)))
	if len(callbacks) == 1 {
		callbacks[0]()
		value.Duration = time.Since(t)
	}
	if len(callbacks) > 1 {
		panic("just one callback per By, please")
	}
}

//BeforeSuite blocks are run just once before any specs are run.  When running in parallel, each
//parallel process will call BeforeSuite.
//
//You may only register *one* BeforeSuite handler per test suite.  You typically do so in your bootstrap file at the top level.
func BeforeSuite(body func()) bool {
	return pushNode(internal.NewNode(deprecationTracker, types.NodeTypeBeforeSuite, "", body))
}

//AfterSuite blocks are *always* run after all the specs regardless of whether specs have passed or failed.
//Moreover, if Ginkgo receives an interrupt signal (^C) it will attempt to run the AfterSuite before exiting.
//
//When running in parallel, each parallel process will call AfterSuite.
//
//You may only register *one* AfterSuite handler per test suite.  You typically do so in your bootstrap file at the top level.
func AfterSuite(body func()) bool {
	return pushNode(internal.NewNode(deprecationTracker, types.NodeTypeAfterSuite, "", body))
}

//SynchronizedBeforeSuite blocks are primarily meant to solve the problem of setting up singleton external resources shared across
//tests when running tests in parallel.  For example, say you have a shared database that you can only start one instance of that
//must be used in your tests.  When running in parallel, only one parallel process should set up the database and all other processes should wait
//until that process is done before running.
//
//SynchronizedBeforeSuite accomplishes this by taking *two* function arguments.  The first is only run on parallel process #1.  The second is
//run on all processes, but *only* after the first function completes successfully.  Ginkgo also makes it possible to send data from the first function (on process #1)
//to the second function (on all the other processes).
//
//The functions have the following signatures.  The first function (which only runs on process #1) has the signature:
//
//	func() []byte
//
//The byte array returned by the first function is then passed to the second function, which has the signature:
//
//	func(data []byte)
//
//Here's a simple pseudo-code example that starts a shared database on process #1 and shares the database's address with the other processes:
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
func SynchronizedBeforeSuite(process1Body func() []byte, allProcessBody func([]byte)) bool {
	return pushNode(internal.NewSynchronizedBeforeSuiteNode(process1Body, allProcessBody, types.NewCodeLocation(1)))
}

//SynchronizedAfterSuite blocks complement the SynchronizedBeforeSuite blocks in solving the problem of setting up
//external singleton resources shared across processes when running tests in parallel.
//
//SynchronizedAfterSuite accomplishes this by taking *two* function arguments.  The first runs on all processes.  The second runs only on parallel process #1
//and *only* after all other nodes have finished and exited.  This ensures that process #1, and any resources it is running, remain alive until
//all other processes are finished.
//
//Here's a pseudo-code example that complements that given in SynchronizedBeforeSuite.  Here, SynchronizedAfterSuite is used to tear down the shared database
//only after all test processes have finished:
//
//	var _ = SynchronizedAfterSuite(func() {
//		dbClient.Cleanup()
//	}, func() {
//		dbRunner.Stop()
//	})
func SynchronizedAfterSuite(allProcessBody func(), process1Body func()) bool {
	return pushNode(internal.NewSynchronizedAfterSuiteNode(allProcessBody, process1Body, types.NewCodeLocation(1)))
}

//BeforeEach blocks are run before It blocks.  When multiple BeforeEach blocks are defined in nested
//Describe and Context blocks the outermost BeforeEach blocks are run first.
func BeforeEach(args ...interface{}) bool {
	return pushNode(internal.NewNode(deprecationTracker, types.NodeTypeBeforeEach, "", args...))
}

//JustBeforeEach blocks are run before It blocks but *after* all BeforeEach blocks.  For more details,
//read the [documentation](http://onsi.github.io/ginkgo/#separating_creation_and_configuration_)
func JustBeforeEach(args ...interface{}) bool {
	return pushNode(internal.NewNode(deprecationTracker, types.NodeTypeJustBeforeEach, "", args...))
}

//JustAfterEach blocks are run after It blocks but *before* all AfterEach blocks.  For more details,
//read the [documentation](http://onsi.github.io/ginkgo/#separating_creation_and_configuration_)
func JustAfterEach(args ...interface{}) bool {
	return pushNode(internal.NewNode(deprecationTracker, types.NodeTypeJustAfterEach, "", args...))
}

//AfterEach blocks are run after It blocks.   When multiple AfterEach blocks are defined in nested
//Describe and Context blocks the innermost AfterEach blocks are run first.
func AfterEach(args ...interface{}) bool {
	return pushNode(internal.NewNode(deprecationTracker, types.NodeTypeAfterEach, "", args...))
}

//BeforeAll blocks occur inside Ordered containers and run just once before any tests run.  Multiple BeforeAll blocks can occur in a given Ordered container
//however they cannot be nested inside any other container, even a container inside an Ordered container.
func BeforeAll(args ...interface{}) bool {
	return pushNode(internal.NewNode(deprecationTracker, types.NodeTypeBeforeAll, "", args...))
}

//AfterAll blocks occur inside Ordered containers and run just once after all tests have run.  Multiple AfterAll blocks can occur in a given Ordered container
//however they cannot be nested inside any other container, even a container inside an Ordered container.
func AfterAll(args ...interface{}) bool {
	return pushNode(internal.NewNode(deprecationTracker, types.NodeTypeAfterAll, "", args...))
}

// DeferCleanup can be called within any setup or subject node to register a cleanup callback that Ginkgo will call at the appropriate time to cleanup after the spec.
// DeferCleanup can be passed a function or a function that returns an error (in which case it will assert that the returned error was nil, or it will fail the test).
// You can also pass DeferCleanup a function that takes arguments followed by a list of arguments to pass to the function.  For example:
//
//     BeforeEach(func() {
//         DeferCleanup(os.SetEnv, "FOO", os.GetEnv("FOO"))
//         os.SetEnv("FOO", "BAR")
//     })
//
// will register a cleanup handler that will set the environment variable "FOO" to it's current value (obtained by os.GetEnv("FOO")) after the spec runs and then sets the environment variable "FOO" to "BAR" for the current spec.
//
// When DeferCleanup is called in BeforeEach, JustBeforeEach, It, AfterEach, or JustAfterEach the registered callback will be invoked when the spec completes (i.e. it will behave like an AfterEach block)
// When DeferCleanup is called in BeforeAll or AfterAll the registered callback will be invoked when the ordered container completes (i.e. it will behave like an AfterAll block)
// When DeferCleanup is called in BeforeSuite, SynchronizedBeforeSuite, AfterSuite, or SynchronizedAfterSuite the registered callback will be invoked when the suite completes (i.e. it will behave like an AfterSuite block)
func DeferCleanup(args ...interface{}) {
	fail := func(message string, cl types.CodeLocation) {
		global.Failer.Fail(message, cl)
	}
	pushNode(internal.NewCleanupNode(fail, args...))
}

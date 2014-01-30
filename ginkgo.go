/*
Ginkgo is a BDD-style testing framework for Golang

The godoc documentation describes Ginkgo's API.  More comprehensive documentation (with examples!) is available at http://onsi.github.io/ginkgo/

Ginkgo's preferred matcher library is [Gomega](http://github.com/onsi/gomega)

Ginkgo on Github: http://github.com/onsi/ginkgo

Ginkgo is MIT-Licensed
*/
package ginkgo

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/remote"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/stenographer"
	"github.com/onsi/ginkgo/types"
	"net/http"
	"os"
	"time"
)

const GINKGO_VERSION = config.VERSION

const defaultTimeout = 1

var globalSuite *suite

func init() {
	config.Flags("ginkgo", true)
	globalSuite = newSuite()
}

//The interface by which Ginkgo receives *testing.T
type GinkgoTestingT interface {
	Fail()
}

//Some matcher libraries or legacy codebases require a *testing.T
//GinkgoT implements an interface analogous to *testing.T and can be used if
//the library in question accepts *testing.T through an interface
//
// For example, with testify:
// assert.Equal(GinkgoT(), 123, 123, "they should be equal")
//
// GinkgoT() takes an optional offset argument that can be used to get the
// correct line number associated with the failure.
func GinkgoT(optionalOffset ...int) *ginkgoTestingTProxy {
	offset := 3
	if len(optionalOffset) > 0 {
		offset = optionalOffset[0]
	}
	return newGinkgoTestingTProxy(Fail, offset)
}

//Custom Ginkgo test reporters must implement the Reporter interface.
//
//The custom reporter is passed in a SuiteSummary when the suite begins and ends,
//and an ExmapleSummary just before an example (spec) begins
//and just after an example (spec) ends
type Reporter reporters.Reporter

//Asynchronous specs given a channel of the Done type.  You must close (or send to) the channel
//to tell Ginkgo that your async test is done.
type Done chan<- interface{}

//GinkgoTestDescription represents the information about the current running test returned by CurrentGinkgoTest
//  ComponentTexts: a list of all texts for the Describes & Contexts leading up to the current test
//  FullTestText: a concatenation of ComponentTexts
//  TestText: the text in the actual It or Measure node
//  IsMeasurement: true if the current test is a measurement
//  FileName: the name of the file containing the current test
//  LineNumber: the line number for the current test
type GinkgoTestDescription struct {
	ComponentTexts []string
	FullTestText   string
	TestText       string

	IsMeasurement bool

	FileName   string
	LineNumber int
}

//CurrentGinkgoTestDescripton returns information about the current running test.
func CurrentGinkgoTestDescription() GinkgoTestDescription {
	return globalSuite.currentGinkgoTestDescription()
}

//Measurement tests receive a Benchmarker.
//
//You use the Time() function to time how long the passed in body function takes to run
//You use the RecordValue() function to track arbitrary numerical measurements.
//The optional info argument is passed to the test reporter and can be used, alongside a custom
//reporter, to provide the measurement data with context.
//
//See http://onsi.github.io/ginkgo/#benchmark_tests for more details
type Benchmarker interface {
	Time(name string, body func(), info ...interface{}) (elapsedTime time.Duration)
	RecordValue(name string, value float64, info ...interface{})
}

//RunSpecs is the entry point for the Ginkgo test runner.
//You must call this within a Go Test... function.
//
//To bootstrap a test suite you can use the Ginkgo CLI:
//
//	ginkgo bootstrap
func RunSpecs(t GinkgoTestingT, description string) bool {
	specReporters := []Reporter{buildDefaultReporter()}
	return globalSuite.run(t, description, specReporters, config.GinkgoConfig)
}

//To run your tests with Ginkgo's default reporter and your custom reporter(s), replace
//RunSpecs() with this method.
func RunSpecsWithDefaultAndCustomReporters(t GinkgoTestingT, description string, specReporters []Reporter) bool {
	specReporters = append([]Reporter{buildDefaultReporter()}, specReporters...)
	return globalSuite.run(t, description, specReporters, config.GinkgoConfig)
}

//To run your tests with your custom reporter(s) (and *not* Ginkgo's default reporter), replace
//RunSpecs() with this method.
func RunSpecsWithCustomReporters(t GinkgoTestingT, description string, specReporters []Reporter) bool {
	return globalSuite.run(t, description, specReporters, config.GinkgoConfig)
}

func buildDefaultReporter() Reporter {
	remoteReportingServer := os.Getenv("GINKGO_REMOTE_REPORTING_SERVER")
	if remoteReportingServer == "" {
		stenographer := stenographer.New(!config.DefaultReporterConfig.NoColor)
		return reporters.NewDefaultReporter(config.DefaultReporterConfig, stenographer)
	} else {
		return remote.NewForwardingReporter(remoteReportingServer, &http.Client{}, remote.NewOutputInterceptor())
	}
}

//Fail notifies Ginkgo that the current spec has failed. (Gomega will call Fail for you automatically when an assertion fails.)
func Fail(message string, callerSkip ...int) {
	skip := 0
	if len(callerSkip) > 0 {
		skip = callerSkip[0]
	}
	globalSuite.fail(message, skip)
}

//Describe blocks allow you to organize your specs.  A Describe block can contain any number of
//BeforeEach, AfterEach, JustBeforeEach, It, and Measurement blocks.
//
//In addition you can nest Describe and Context blocks.  Describe and Context blocks are functionally
//equivalent.  The difference is purely semantic -- you typical Describe the behavior of an object
//or method and, within that Describe, outline a number of Contexts.
func Describe(text string, body func()) bool {
	globalSuite.pushContainerNode(text, body, flagTypeNone, types.GenerateCodeLocation(1))
	return true
}

//You can focus the tests within a describe block using FDescribe
func FDescribe(text string, body func()) bool {
	globalSuite.pushContainerNode(text, body, flagTypeFocused, types.GenerateCodeLocation(1))
	return true
}

//You can mark the tests within a describe block as pending using PDescribe
func PDescribe(text string, body func()) bool {
	globalSuite.pushContainerNode(text, body, flagTypePending, types.GenerateCodeLocation(1))
	return true
}

//You can mark the tests within a describe block as pending using XDescribe
func XDescribe(text string, body func()) bool {
	globalSuite.pushContainerNode(text, body, flagTypePending, types.GenerateCodeLocation(1))
	return true
}

//Context blocks allow you to organize your specs.  A Context block can contain any number of
//BeforeEach, AfterEach, JustBeforeEach, It, and Measurement blocks.
//
//In addition you can nest Describe and Context blocks.  Describe and Context blocks are functionally
//equivalent.  The difference is purely semantic -- you typical Describe the behavior of an object
//or method and, within that Describe, outline a number of Contexts.
func Context(text string, body func()) bool {
	globalSuite.pushContainerNode(text, body, flagTypeNone, types.GenerateCodeLocation(1))
	return true
}

//You can focus the tests within a describe block using FContext
func FContext(text string, body func()) bool {
	globalSuite.pushContainerNode(text, body, flagTypeFocused, types.GenerateCodeLocation(1))
	return true
}

//You can mark the tests within a describe block as pending using PContext
func PContext(text string, body func()) bool {
	globalSuite.pushContainerNode(text, body, flagTypePending, types.GenerateCodeLocation(1))
	return true
}

//You can mark the tests within a describe block as pending using XContext
func XContext(text string, body func()) bool {
	globalSuite.pushContainerNode(text, body, flagTypePending, types.GenerateCodeLocation(1))
	return true
}

//It blocks contain your test code and assertions.  You cannot nest any other Ginkgo blocks
//within an It block.
//
//Ginkgo will normally run It blocks synchronously.  To perform asynchronous tests, pass a
//function that accepts a Done channel.  When you do this, you can alos provide an optional timeout.
func It(text string, body interface{}, timeout ...float64) bool {
	globalSuite.pushItNode(text, body, flagTypeNone, types.GenerateCodeLocation(1), parseTimeout(timeout...))
	return true
}

//You can focus individual Its using FIt
func FIt(text string, body interface{}, timeout ...float64) bool {
	globalSuite.pushItNode(text, body, flagTypeFocused, types.GenerateCodeLocation(1), parseTimeout(timeout...))
	return true
}

//You can mark Its as pending using PIt
func PIt(text string, _ ...interface{}) bool {
	globalSuite.pushItNode(text, func() {}, flagTypePending, types.GenerateCodeLocation(1), 0)
	return true
}

//You can mark Its as pending using XIt
func XIt(text string, _ ...interface{}) bool {
	globalSuite.pushItNode(text, func() {}, flagTypePending, types.GenerateCodeLocation(1), 0)
	return true
}

//Measure blocks run the passed in body function repeatedly (determined by the samples argument)
//and accumulate metrics provided to the Benchmarker by the body function.
func Measure(text string, body func(Benchmarker), samples int) bool {
	globalSuite.pushMeasureNode(text, body, flagTypeNone, types.GenerateCodeLocation(1), samples)
	return true
}

//You can focus individual Measures using FMeasure
func FMeasure(text string, body func(Benchmarker), samples int) bool {
	globalSuite.pushMeasureNode(text, body, flagTypeFocused, types.GenerateCodeLocation(1), samples)
	return true
}

//You can mark Maeasurements as pending using PMeasure
func PMeasure(text string, _ ...interface{}) bool {
	globalSuite.pushMeasureNode(text, func(b Benchmarker) {}, flagTypePending, types.GenerateCodeLocation(1), 0)
	return true
}

//You can mark Maeasurements as pending using XMeasure
func XMeasure(text string, _ ...interface{}) bool {
	globalSuite.pushMeasureNode(text, func(b Benchmarker) {}, flagTypePending, types.GenerateCodeLocation(1), 0)
	return true
}

//BeforeEach blocks are run before It blocks.  When multiple BeforeEach blocks are defined in nested
//Describe and Context blocks the outermost BeforeEach blocks are run first.
//
//Like It blocks, BeforeEach blocks can be made asynchronous by providing a body function that accepts
//a Done channel
func BeforeEach(body interface{}, timeout ...float64) bool {
	globalSuite.pushBeforeEachNode(body, types.GenerateCodeLocation(1), parseTimeout(timeout...))
	return true
}

//JustBeforeEach blocks are run before It blocks but *after* all BeforeEach blocks.  For more details,
//read the [documentation](http://onsi.github.io/ginkgo/#separating_creation_and_configuration_)
//
//Like It blocks, BeforeEach blocks can be made asynchronous by providing a body function that accepts
//a Done channel
func JustBeforeEach(body interface{}, timeout ...float64) bool {
	globalSuite.pushJustBeforeEachNode(body, types.GenerateCodeLocation(1), parseTimeout(timeout...))
	return true
}

//AfterEach blocks are run after It blocks.   When multiple AfterEach blocks are defined in nested
//Describe and Context blocks the innermost AfterEach blocks are run first.
//
//Like It blocks, BeforeEach blocks can be made asynchronous by providing a body function that accepts
//a Done channel
func AfterEach(body interface{}, timeout ...float64) bool {
	globalSuite.pushAfterEachNode(body, types.GenerateCodeLocation(1), parseTimeout(timeout...))
	return true
}

func parseTimeout(timeout ...float64) time.Duration {
	if len(timeout) == 0 {
		return time.Duration(defaultTimeout * int64(time.Second))
	} else {
		return time.Duration(timeout[0] * float64(time.Second))
	}
}

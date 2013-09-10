package ginkgo

import (
	"github.com/onsi/ginkgo/config"

	"testing"
	"time"
)

const defaultTimeout = 5

var globalSuite *suite

func init() {
	config.Flags("ginkgo", true)
	globalSuite = newSuite()
}

func RunSpecs(t *testing.T, description string) {
	RunSpecsWithCustomReporter(t, description, newDefaultReporter(config.DefaultReporterConfig))
}

func RunSpecsWithCustomReporter(t *testing.T, description string, reporter Reporter) {
	globalSuite.run(t, description, reporter, config.GinkgoConfig)
}

type Done chan<- interface{}

func Fail(message string, callerSkip ...int) {
	skip := 0
	if len(callerSkip) > 0 {
		skip = callerSkip[0]
	}
	globalSuite.fail(message, skip)
}

//Describes

func Describe(text string, body func()) bool {
	globalSuite.pushContainerNode(text, body, flagTypeNone, generateCodeLocation(1))
	return true
}

func FDescribe(text string, body func()) bool {
	globalSuite.pushContainerNode(text, body, flagTypeFocused, generateCodeLocation(1))
	return true
}

func PDescribe(text string, body func()) bool {
	globalSuite.pushContainerNode(text, body, flagTypePending, generateCodeLocation(1))
	return true
}

func XDescribe(text string, body func()) bool {
	globalSuite.pushContainerNode(text, body, flagTypePending, generateCodeLocation(1))
	return true
}

//Context

func Context(text string, body func()) bool {
	globalSuite.pushContainerNode(text, body, flagTypeNone, generateCodeLocation(1))
	return true
}

func FContext(text string, body func()) bool {
	globalSuite.pushContainerNode(text, body, flagTypeFocused, generateCodeLocation(1))
	return true
}

func PContext(text string, body func()) bool {
	globalSuite.pushContainerNode(text, body, flagTypePending, generateCodeLocation(1))
	return true
}

func XContext(text string, body func()) bool {
	globalSuite.pushContainerNode(text, body, flagTypePending, generateCodeLocation(1))
	return true
}

//It

func It(text string, body interface{}, timeout ...float64) bool {
	globalSuite.pushItNode(text, body, flagTypeNone, generateCodeLocation(1), parseTimeout(timeout...))
	return true
}

func FIt(text string, body interface{}, timeout ...float64) bool {
	globalSuite.pushItNode(text, body, flagTypeFocused, generateCodeLocation(1), parseTimeout(timeout...))
	return true
}

func PIt(text string, body interface{}, timeout ...float64) bool {
	globalSuite.pushItNode(text, body, flagTypePending, generateCodeLocation(1), parseTimeout(timeout...))
	return true
}

func XIt(text string, body interface{}, timeout ...float64) bool {
	globalSuite.pushItNode(text, body, flagTypePending, generateCodeLocation(1), parseTimeout(timeout...))
	return true
}

//Benchmark

func Benchmark(text string, body interface{}, samples int, maximumAllowedTime float64, timeout ...float64) bool {
	globalSuite.pushBenchmarkNode(text, body, flagTypeNone, generateCodeLocation(1), parseTimeout(timeout...), samples, time.Duration(maximumAllowedTime*float64(time.Second)))
	return true
}

func FBenchmark(text string, body interface{}, samples int, maximumAllowedTime float64, timeout ...float64) bool {
	globalSuite.pushBenchmarkNode(text, body, flagTypeFocused, generateCodeLocation(1), parseTimeout(timeout...), samples, time.Duration(maximumAllowedTime*float64(time.Second)))
	return true
}

func PBenchmark(text string, body interface{}, samples int, maximumAllowedTime float64, timeout ...float64) bool {
	globalSuite.pushBenchmarkNode(text, body, flagTypePending, generateCodeLocation(1), parseTimeout(timeout...), samples, time.Duration(maximumAllowedTime*float64(time.Second)))
	return true
}

func XBenchmark(text string, body interface{}, samples int, maximumAllowedTime float64, timeout ...float64) bool {
	globalSuite.pushBenchmarkNode(text, body, flagTypePending, generateCodeLocation(1), parseTimeout(timeout...), samples, time.Duration(maximumAllowedTime*float64(time.Second)))
	return true
}

//Before, JustBefore, and After

func BeforeEach(body interface{}, timeout ...float64) bool {
	globalSuite.pushBeforeEachNode(body, generateCodeLocation(1), parseTimeout(timeout...))
	return true
}

func JustBeforeEach(body interface{}, timeout ...float64) bool {
	globalSuite.pushJustBeforeEachNode(body, generateCodeLocation(1), parseTimeout(timeout...))
	return true
}

func AfterEach(body interface{}, timeout ...float64) bool {
	globalSuite.pushAfterEachNode(body, generateCodeLocation(1), parseTimeout(timeout...))
	return true
}

func parseTimeout(timeout ...float64) time.Duration {
	if len(timeout) == 0 {
		return time.Duration(defaultTimeout * int64(time.Second))
	} else {
		return time.Duration(timeout[0] * float64(time.Second))
	}
}

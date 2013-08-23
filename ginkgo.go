package ginkgo

import (
	"flag"
	"testing"
	"time"
)

var suiteRandomSeed = flag.Int64("seed", time.Now().Unix(), "The seed used to randomize the spec suite.")
var suiteRandomizeAllSpecs = flag.Bool("randomizeAllSpecs", false, "If set, ginkgo will randomize all specs together.  By default, ginkgo only randomizes the top level Describe/Context groups.")
var reporterNoColor = flag.Bool("noColor", false, "If set, suppress color output in default reporter.")
var reporterSlowSpecThreshold = flag.Float64("slowSpecThreshold", 5.0, "(in seconds) Specs that take longer to run than this threshold are flagged as slow by the default reporter (default: 5 seconds).")

var globalSuite *suite

func init() {
	globalSuite = newSuite()
}

func RunSpecs(t *testing.T, description string) {
	reporter := newDefaultReporter(*reporterNoColor, *reporterSlowSpecThreshold)
	RunSpecsWithCustomReporter(t, description, reporter)
}

func RunSpecsWithCustomReporter(t *testing.T, description string, reporter Reporter) {
	globalSuite.run(t, description, *suiteRandomSeed, *suiteRandomizeAllSpecs, reporter)
}

type Done chan<- interface{}

func Fail(message string, callerSkip ...int) {
	skip := 0
	if len(callerSkip) > 0 {
		skip = callerSkip[0]
	}
	globalSuite.fail(message, skip)
}

func Describe(text string, body func()) bool {
	codeLocation, _ := generateCodeLocation(2)
	globalSuite.pushContainerNode(text, body, flagTypeNone, codeLocation)
	return true
}

func FDescribe(text string, body func()) bool {
	codeLocation, _ := generateCodeLocation(2)
	globalSuite.pushContainerNode(text, body, flagTypeFocused, codeLocation)
	return true
}

func PDescribe(text string, body func()) bool {
	codeLocation, _ := generateCodeLocation(2)
	globalSuite.pushContainerNode(text, body, flagTypePending, codeLocation)
	return true
}

func XDescribe(text string, body func()) bool {
	codeLocation, _ := generateCodeLocation(2)
	globalSuite.pushContainerNode(text, body, flagTypePending, codeLocation)
	return true
}

func Context(text string, body func()) bool {
	codeLocation, _ := generateCodeLocation(2)
	globalSuite.pushContainerNode(text, body, flagTypeNone, codeLocation)
	return true
}

func FContext(text string, body func()) bool {
	codeLocation, _ := generateCodeLocation(2)
	globalSuite.pushContainerNode(text, body, flagTypeFocused, codeLocation)
	return true
}

func PContext(text string, body func()) bool {
	codeLocation, _ := generateCodeLocation(2)
	globalSuite.pushContainerNode(text, body, flagTypePending, codeLocation)
	return true
}

func XContext(text string, body func()) bool {
	codeLocation, _ := generateCodeLocation(2)
	globalSuite.pushContainerNode(text, body, flagTypePending, codeLocation)
	return true
}

func It(text string, body interface{}) bool {
	codeLocation, _ := generateCodeLocation(2)
	globalSuite.pushItNode(text, body, flagTypeNone, codeLocation)
	return true
}

func FIt(text string, body interface{}) bool {
	codeLocation, _ := generateCodeLocation(2)
	globalSuite.pushItNode(text, body, flagTypeFocused, codeLocation)
	return true
}

func PIt(text string, body interface{}) bool {
	codeLocation, _ := generateCodeLocation(2)
	globalSuite.pushItNode(text, body, flagTypePending, codeLocation)
	return true
}

func XIt(text string, body interface{}) bool {
	codeLocation, _ := generateCodeLocation(2)
	globalSuite.pushItNode(text, body, flagTypePending, codeLocation)
	return true
}

func BeforeEach(body interface{}) bool {
	codeLocation, _ := generateCodeLocation(2)
	globalSuite.pushBeforeEachNode(body, codeLocation)
	return true
}

func JustBeforeEach(body interface{}) bool {
	codeLocation, _ := generateCodeLocation(2)
	globalSuite.pushJustBeforeEachNode(body, codeLocation)
	return true
}

func AfterEach(body interface{}) bool {
	codeLocation, _ := generateCodeLocation(2)
	globalSuite.pushAfterEachNode(body, codeLocation)
	return true
}

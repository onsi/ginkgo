package godescribe

import (
	"flag"
	"testing"
	"time"
)

var suiteRandomSeed = flag.Int64("seed", time.Now().Unix(), "The seed used to randomize the spec suite.")
var suiteRandomizeAllSpecs = flag.Bool("randomizeAllSpecs", false, "If set, godescribe will randomize all specs together.  By default, godescribe only randomizes the top level Describe/Context groups.")
var reporterNoColor = flag.Bool("noColor", false, "If set, suppress color output in default reporter.")
var reporterSlowSpecThreshold = flag.Float64("slowSpecThreshold", 5.0, "(in seconds) Specs that take longer to run than this threshold are flagged as slow by the default reporter (default: 5 seconds).")

var globalSuite *suite

func init() {
	//set up the global suite
	globalSuite = newSuite()
}

func RunSpecs(t *testing.T, description string) {
	reporter := newDefaultReporter(*reporterNoColor, *reporterSlowSpecThreshold) //todo: color and slow threshold args
	RunSpecsWithCustomReporter(t, description, reporter)
}

func RunSpecsWithCustomReporter(t *testing.T, description string, reporter Reporter) {
	globalSuite.run(t, description, *suiteRandomSeed, *suiteRandomizeAllSpecs, reporter)
}

type Done chan<- interface{} //channel for async callbacks

func Fail(message string, callerSkip ...int) {
	skip := 0
	if len(callerSkip) > 0 {
		skip = callerSkip[0]
	}
	globalSuite.fail(message, skip)
}

//These all just call (private) methods on the global suite

func Describe(text string, body func()) {
	codeLocation, _ := generateCodeLocation(1)
	globalSuite.pushContainerNode(text, body, containerTypeDescribe, flagTypeNone, codeLocation)
}

func FDescribe(text string, body func()) {
	codeLocation, _ := generateCodeLocation(1)
	globalSuite.pushContainerNode(text, body, containerTypeDescribe, flagTypeFocused, codeLocation)
}

func PDescribe(text string, body func()) {
	codeLocation, _ := generateCodeLocation(1)
	globalSuite.pushContainerNode(text, body, containerTypeDescribe, flagTypePending, codeLocation)
}

func Context(text string, body func()) {
	codeLocation, _ := generateCodeLocation(1)
	globalSuite.pushContainerNode(text, body, containerTypeContext, flagTypeNone, codeLocation)
}

func FContext(text string, body func()) {
	codeLocation, _ := generateCodeLocation(1)
	globalSuite.pushContainerNode(text, body, containerTypeContext, flagTypeFocused, codeLocation)
}

func PContext(text string, body func()) {
	codeLocation, _ := generateCodeLocation(1)
	globalSuite.pushContainerNode(text, body, containerTypeContext, flagTypePending, codeLocation)
}

func It(text string, body interface{}) {
	codeLocation, _ := generateCodeLocation(1)
	globalSuite.pushExampleNode(text, body, flagTypeNone, codeLocation)
}

func FIt(text string, body interface{}) {
	codeLocation, _ := generateCodeLocation(1)
	globalSuite.pushExampleNode(text, body, flagTypeFocused, codeLocation)
}

func PIt(text string, body interface{}) {
	codeLocation, _ := generateCodeLocation(1)
	globalSuite.pushExampleNode(text, body, flagTypePending, codeLocation)
}

func BeforeEach(body interface{}) {
	codeLocation, _ := generateCodeLocation(1)
	globalSuite.pushBeforeEachNode(body, codeLocation)
}

func JustBeforeEach(body interface{}) {
	codeLocation, _ := generateCodeLocation(1)
	globalSuite.pushJustBeforeEachNode(body, codeLocation)
}

func AfterEach(body interface{}) {
	codeLocation, _ := generateCodeLocation(1)
	globalSuite.pushAfterEachNode(body, codeLocation)
}

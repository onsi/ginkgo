package ginkgo

import (
	"flag"
	"testing"
	"time"
)

var (
	randomSeed        = flag.Int64("ginkgo.seed", time.Now().Unix(), "The seed used to randomize the spec suite.")
	randomizeAllSpecs = flag.Bool("ginkgo.randomizeAllSpecs", false, "If set, ginkgo will randomize all specs together.  By default, ginkgo only randomizes the top level Describe/Context groups.")

	noColor           = flag.Bool("ginkgo.noColor", false, "If set, suppress color output in default reporter.")
	slowSpecThreshold = flag.Float64("ginkgo.slowSpecThreshold", 5.0, "(in seconds) Specs that take longer to run than this threshold are flagged as slow by the default reporter (default: 5 seconds).")
	noisyPendings     = flag.Bool("ginkgo.noisyPendings", true, "If set, default reporter will shout about pending tests.")
)

type GinkoConfigType struct {
	RandomSeed        int64
	RandomizeAllSpecs bool
}

var GinkgoConfig GinkoConfigType
var globalSuite *suite

func init() {
	globalSuite = newSuite()
}

func RunSpecs(t *testing.T, description string) {
	reporter := newDefaultReporter(defaultReporterConfig{
		noColor:           *noColor,
		slowSpecThreshold: *slowSpecThreshold,
		noisyPendings:     *noisyPendings,
	})
	RunSpecsWithCustomReporter(t, description, reporter)
}

func RunSpecsWithCustomReporter(t *testing.T, description string, reporter Reporter) {
	GinkgoConfig.RandomSeed = *randomSeed
	GinkgoConfig.RandomizeAllSpecs = *randomizeAllSpecs

	globalSuite.run(t, description, reporter, GinkgoConfig)
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
		return time.Duration(5 * time.Second)
	} else {
		return time.Duration(timeout[0] * float64(time.Second))
	}
}

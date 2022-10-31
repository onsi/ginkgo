package internal_integration_test

import (
	"context"
	"io"
	"reflect"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal"
	"github.com/onsi/ginkgo/v2/internal/global"
	"github.com/onsi/ginkgo/v2/internal/parallel_support"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

func TestSuiteTests(t *testing.T) {
	format.TruncatedDiff = false
	RegisterFailHandler(Fail)
	suiteConfig, _ := GinkgoConfiguration()
	suiteConfig.GracePeriod = time.Second
	RunSpecs(t, "Suite Integration Tests", suiteConfig)
}

var conf types.SuiteConfig
var failer *internal.Failer
var writer *internal.Writer
var reporter *FakeReporter
var rt *RunTracker
var cl types.CodeLocation
var interruptHandler *FakeInterruptHandler
var outputInterceptor *FakeOutputInterceptor

var server parallel_support.Server
var client parallel_support.Client
var exitChannels map[int]chan interface{}

var triggerProgressSignal func()

func progressSignalRegistrar(handler func()) context.CancelFunc {
	triggerProgressSignal = handler
	return func() { triggerProgressSignal = nil }
}

func noopProgressSignalRegistrar(_ func()) context.CancelFunc {
	return func() {}
}

var _ = BeforeEach(func() {
	conf = types.SuiteConfig{}
	failer = internal.NewFailer()
	writer = internal.NewWriter(io.Discard)
	writer.SetMode(internal.WriterModeBufferOnly)
	reporter = NewFakeReporter()
	rt = NewRunTracker()
	cl = types.NewCodeLocation(0)
	interruptHandler = NewFakeInterruptHandler()
	outputInterceptor = NewFakeOutputInterceptor()

	conf.ParallelTotal = 1
	conf.ParallelProcess = 1
	conf.RandomSeed = 17
	conf.GracePeriod = 30 * time.Second

	server, client, exitChannels = nil, nil, nil
})

/* Helpers to set up and run test fixtures using the Ginkgo DSL */
func WithSuite(suite *internal.Suite, callback func()) {
	originalSuite, originalFailer := global.Suite, global.Failer
	global.Suite = suite
	global.Failer = failer
	callback()
	global.Suite = originalSuite
	global.Failer = originalFailer
}

func SetUpForParallel(parallelTotal int) {
	conf.ParallelTotal = parallelTotal
	server, client, exitChannels = SetUpServerAndClient(conf.ParallelTotal)
	conf.ParallelHost = server.Address()
}

func RunFixture(description string, callback func()) (bool, bool) {
	suite := internal.NewSuite()
	var success, hasProgrammaticFocus bool
	WithSuite(suite, func() {
		callback()
		Ω(suite.BuildTree()).Should(Succeed())
		success, hasProgrammaticFocus = suite.Run(description, Label("TopLevelLabel"), "/path/to/suite", failer, reporter, writer, outputInterceptor, interruptHandler, client, progressSignalRegistrar, conf)
	})
	return success, hasProgrammaticFocus
}

func F(options ...interface{}) {
	location := cl
	message := "fail"
	for _, option := range options {
		if reflect.TypeOf(option).Kind() == reflect.String {
			message = option.(string)
		} else if reflect.TypeOf(option) == reflect.TypeOf(cl) {
			location = option.(types.CodeLocation)
		}
	}

	failer.Fail(message, location)
	panic("panic to simulate how ginkgo's Fail works")
}

func Abort(options ...interface{}) {
	location := cl
	message := "abort"
	for _, option := range options {
		if reflect.TypeOf(option).Kind() == reflect.String {
			message = option.(string)
		} else if reflect.TypeOf(option) == reflect.TypeOf(cl) {
			location = option.(types.CodeLocation)
		}
	}

	failer.AbortSuite(message, location)
	panic("panic to simulate how ginkgo's AbortSuite works")
}

func FixtureSkip(options ...interface{}) {
	location := cl
	message := "skip"
	for _, option := range options {
		if reflect.TypeOf(option).Kind() == reflect.String {
			message = option.(string)
		} else if reflect.TypeOf(option) == reflect.TypeOf(cl) {
			location = option.(types.CodeLocation)
		}
	}

	failer.Skip(message, location)
	panic("panic to simulate how ginkgo's Skip works")
}

func HaveHighlightedStackLine(cl types.CodeLocation) OmegaMatcher {
	return ContainElement(WithTransform(func(fc types.FunctionCall) types.CodeLocation {
		if fc.Highlight {
			return types.CodeLocation{
				FileName:   fc.Filename,
				LineNumber: int(fc.Line),
			}
		}
		return types.CodeLocation{}
	}, Equal(cl)))
}

func clLine(offset int) types.CodeLocation {
	cl := cl
	cl.LineNumber += offset
	return cl
}

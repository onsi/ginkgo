package internal_integration_test

import (
	"io"
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/internal"
	"github.com/onsi/ginkgo/internal/global"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
)

func TestSuiteTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Suite Integration Tests")
}

var conf types.SuiteConfig
var failer *internal.Failer
var writer *internal.Writer
var reporter *FakeReporter
var rt *RunTracker
var cl types.CodeLocation
var interruptHandler *FakeInterruptHandler
var outputInterceptor *FakeOutputInterceptor

var _ = BeforeEach(func() {
	conf = types.SuiteConfig{}
	failer = internal.NewFailer()
	writer = internal.NewWriter(io.Discard)
	writer.SetMode(internal.WriterModeBufferOnly)
	reporter = &FakeReporter{}
	rt = NewRunTracker()
	cl = types.NewCodeLocation(0)
	interruptHandler = NewFakeInterruptHandler()
	outputInterceptor = NewFakeOutputInterceptor()

	conf.ParallelTotal = 1
	conf.ParallelNode = 1
})

var _ = AfterEach(func() {
	interruptHandler.Stop()
})

/* Helpers to set up and run test fixtures using the Ginkgo DSL */
func WithSuite(suite *internal.Suite, callback func()) {
	originalSuite := global.Suite
	global.Suite = suite
	callback()
	global.Suite = originalSuite
}

func RunFixture(description string, callback func()) (bool, bool) {
	suite := internal.NewSuite()
	var success, hasProgrammaticFocus bool
	WithSuite(suite, func() {
		callback()
		Ω(suite.BuildTree()).Should(Succeed())
		success, hasProgrammaticFocus = suite.Run(description, "/path/to/suite", failer, reporter, writer, outputInterceptor, interruptHandler, conf)
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

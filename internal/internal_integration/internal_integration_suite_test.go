package internal_integration_test

import (
	"io/ioutil"
	"reflect"
	"sync"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/internal"
	"github.com/onsi/ginkgo/internal/global"
)

func TestSuiteTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Suite Integration Tests")
}

var conf config.GinkgoConfigType
var failer *internal.Failer
var writer *internal.Writer
var reporter *FakeReporter
var rt *RunTracker
var cl types.CodeLocation
var interruptHandler *FakeInterruptHandler

var _ = BeforeEach(func() {
	conf = config.GinkgoConfigType{}
	failer = internal.NewFailer()
	writer = internal.NewWriter(ioutil.Discard)
	writer.SetMode(internal.WriterModeBufferOnly)
	reporter = &FakeReporter{}
	rt = NewRunTracker()
	cl = types.NewCodeLocation(0)
	interruptHandler = NewFakeInterruptHandler()

	conf.ParallelTotal = 1
	conf.ParallelNode = 1
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
		success, hasProgrammaticFocus = suite.Run(description, failer, reporter, writer, interruptHandler, conf)
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

/* InterruptHandler */

type FakeInterruptHandler struct {
	triggerInterrupt chan bool

	c           chan interface{}
	lock        *sync.Mutex
	interrupted bool
}

func NewFakeInterruptHandler() *FakeInterruptHandler {
	handler := &FakeInterruptHandler{
		triggerInterrupt: make(chan bool),
		c:                make(chan interface{}),
		lock:             &sync.Mutex{},
		interrupted:      false,
	}
	handler.registerForInterrupts()
	return handler
}

func (handler *FakeInterruptHandler) registerForInterrupts() {
	go func() {
		for {
			<-handler.triggerInterrupt
			handler.lock.Lock()
			handler.interrupted = true
			close(handler.c)
			handler.c = make(chan interface{})
			handler.lock.Unlock()
		}
	}()
}

func (handler *FakeInterruptHandler) Interrupt() {
	handler.triggerInterrupt <- true
}

func (handler *FakeInterruptHandler) Status() internal.InterruptStatus {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	return internal.InterruptStatus{
		Interrupted: handler.interrupted,
		Channel:     handler.c,
	}
}

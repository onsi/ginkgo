package ginkgo

import (
	"fmt"
)

type failFunc func(message string, callerSkip ...int)

func newGinkgoTestingTProxy(fail failFunc) *ginkgoTestingTProxy {
	return &ginkgoTestingTProxy{
		fail:   fail,
		offset: 3,
	}
}

type ginkgoTestingTProxy struct {
	fail   failFunc
	offset int
}

func (t *ginkgoTestingTProxy) Error(args ...interface{}) {
	t.fail(fmt.Sprintln(args...), t.offset)
}

func (t *ginkgoTestingTProxy) Errorf(format string, args ...interface{}) {
	t.fail(fmt.Sprintf(format, args...), t.offset)
}

func (t *ginkgoTestingTProxy) Fail() {
	t.fail("", t.offset)
}

func (t *ginkgoTestingTProxy) FailNow() {
	t.fail("", t.offset)
}

func (t *ginkgoTestingTProxy) Fatal(args ...interface{}) {
	t.fail(fmt.Sprintln(args...), t.offset)
}

func (t *ginkgoTestingTProxy) Fatalf(format string, args ...interface{}) {
	t.fail(fmt.Sprintf(format, args...), t.offset)
}

func (t *ginkgoTestingTProxy) Log(args ...interface{}) {
	fmt.Println(args...)
}

func (t *ginkgoTestingTProxy) Logf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

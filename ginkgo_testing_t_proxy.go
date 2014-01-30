package ginkgo

import (
	"fmt"
)

type failFunc func(message string, callerSkip ...int)

func newGinkgoTestingTProxy(fail failFunc, offset int) *ginkgoTestingTProxy {
	return &ginkgoTestingTProxy{
		fail:   fail,
		offset: offset,
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
	t.fail("failed", t.offset)
}

func (t *ginkgoTestingTProxy) FailNow() {
	t.fail("failed", t.offset)
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

func (t *ginkgoTestingTProxy) Failed() bool {
	return false
}

func (t *ginkgoTestingTProxy) Parallel() {
}

func (t *ginkgoTestingTProxy) Skip(args ...interface{}) {
	fmt.Println(args...)
}

func (t *ginkgoTestingTProxy) Skipf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (t *ginkgoTestingTProxy) SkipNow(args ...interface{}) {
	fmt.Println(args...)
}

func (t *ginkgoTestingTProxy) Skipped() bool {
	return false
}

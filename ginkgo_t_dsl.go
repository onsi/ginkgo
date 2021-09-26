package ginkgo

import "github.com/onsi/ginkgo/internal/testingtproxy"

//Some matcher libraries or legacy codebases require a *testing.T
//GinkgoT implements an interface analogous to *testing.T and can be used if
//the library in question accepts *testing.T through an interface
//
// For example, with testify:
// assert.Equal(GinkgoT(), 123, 123, "they should be equal")
//
// Or with gomock:
// gomock.NewController(GinkgoT())
//
// GinkgoT() takes an optional offset argument that can be used to get the
// correct line number associated with the failure.
func GinkgoT(optionalOffset ...int) GinkgoTInterface {
	offset := 3
	if len(optionalOffset) > 0 {
		offset = optionalOffset[0]
	}
	return testingtproxy.New(GinkgoWriter, Fail, Skip, DeferCleanup, CurrentSpecReport, offset)
}

//The interface returned by GinkgoT().  This covers most of the methods
//in the testing package's T.
type GinkgoTInterface interface {
	Cleanup(func())
	Setenv(kev, value string)
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Helper()
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Name() string
	Parallel()
	Skip(args ...interface{})
	SkipNow()
	Skipf(format string, args ...interface{})
	Skipped() bool
	TempDir() string
}

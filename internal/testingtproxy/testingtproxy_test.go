package testingtproxy_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"

	"github.com/onsi/ginkgo/internal/testingtproxy"
)

type messagedCall struct {
	message    string
	callerSkip []int
}

var _ = Describe("Testingtproxy", func() {
	var t GinkgoTInterface

	var failFunc func(message string, callerSkip ...int)
	var skipFunc func(message string, callerSkip ...int)
	var failedFunc func() bool
	var nameFunc func() string

	var nameToReturn string
	var failedToReturn bool
	var failFuncCall messagedCall
	var skipFuncCall messagedCall
	var offset int
	var buf *gbytes.Buffer

	BeforeEach(func() {
		failFuncCall = messagedCall{}
		skipFuncCall = messagedCall{}
		nameToReturn = ""
		failedToReturn = false
		offset = 3

		failFunc = func(message string, callerSkip ...int) {
			failFuncCall.message = message
			failFuncCall.callerSkip = callerSkip
		}

		skipFunc = func(message string, callerSkip ...int) {
			skipFuncCall.message = message
			skipFuncCall.callerSkip = callerSkip
		}

		failedFunc = func() bool {
			return failedToReturn
		}

		nameFunc = func() string {
			return nameToReturn
		}

		buf = gbytes.NewBuffer()

		t = testingtproxy.New(buf, failFunc, skipFunc, failedFunc, nameFunc, offset)
	})

	It("ignores Cleanup", func() {
		GinkgoT().Cleanup(func() {
			panic("bam!")
		}) //is a no-op
	})

	It("supports Error", func() {
		t.Error("a", 17)
		Ω(failFuncCall.message).Should(Equal("a 17\n"))
		Ω(failFuncCall.callerSkip).Should(Equal([]int{offset}))
	})

	It("supports Errorf", func() {
		t.Errorf("%s %d!", "a", 17)
		Ω(failFuncCall.message).Should(Equal("a 17!"))
		Ω(failFuncCall.callerSkip).Should(Equal([]int{offset}))
	})

	It("supports Fail", func() {
		t.Fail()
		Ω(failFuncCall.message).Should(Equal("failed"))
		Ω(failFuncCall.callerSkip).Should(Equal([]int{offset}))
	})

	It("supports FailNow", func() {
		t.Fail()
		Ω(failFuncCall.message).Should(Equal("failed"))
		Ω(failFuncCall.callerSkip).Should(Equal([]int{offset}))
	})

	It("supports Fatal", func() {
		t.Fatal("a", 17)
		Ω(failFuncCall.message).Should(Equal("a 17\n"))
		Ω(failFuncCall.callerSkip).Should(Equal([]int{offset}))
	})

	It("supports Fatalf", func() {
		t.Fatalf("%s %d!", "a", 17)
		Ω(failFuncCall.message).Should(Equal("a 17!"))
		Ω(failFuncCall.callerSkip).Should(Equal([]int{offset}))
	})

	It("ignores Helper", func() {
		GinkgoT().Helper() //is a no-op
	})

	It("supports Log", func() {
		t.Log("a", 17)
		Ω(string(buf.Contents())).Should(Equal("a 17\n"))
	})

	It("supports Logf", func() {
		t.Logf("%s %d!", "a", 17)
		Ω(string(buf.Contents())).Should(Equal("a 17!\n"))
	})

	It("supports Name", func() {
		nameToReturn = "C.S. Lewis"
		Ω(t.Name()).Should(Equal("C.S. Lewis"))

		Ω(GinkgoT().Name()).Should(ContainSubstring("supports Name"))
	})

	It("ignores Parallel", func() {
		GinkgoT().Parallel() //is a no-op
	})

	It("supports Skip", func() {
		t.Skip("a", 17)
		Ω(skipFuncCall.message).Should(Equal("a 17\n"))
		Ω(skipFuncCall.callerSkip).Should(Equal([]int{offset}))
	})

	It("supports SkipNow", func() {
		t.SkipNow()
		Ω(skipFuncCall.message).Should(Equal("skip"))
		Ω(skipFuncCall.callerSkip).Should(Equal([]int{offset}))
	})

	It("supports Skipf", func() {
		t.Skipf("%s %d!", "a", 17)
		Ω(skipFuncCall.message).Should(Equal("a 17!"))
		Ω(skipFuncCall.callerSkip).Should(Equal([]int{offset}))
	})

	It("always returns false for Skipped", func() {
		Ω(GinkgoT().Skipped()).Should(BeFalse())
	})

	It("returns empty string for TempDir", func() {
		Ω(GinkgoT().TempDir()).Should(Equal(""))
	})
})

package testingtproxy_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"

	"github.com/onsi/ginkgo/v2/internal/testingtproxy"
	"github.com/onsi/ginkgo/v2/types"
)

type messagedCall struct {
	message    string
	callerSkip []int
}

var _ = Describe("Testingtproxy", func() {
	var t GinkgoTInterface

	var failFunc func(message string, callerSkip ...int)
	var skipFunc func(message string, callerSkip ...int)
	var reportFunc func() types.SpecReport

	var failFuncCall messagedCall
	var skipFuncCall messagedCall
	var offset int
	var reportToReturn types.SpecReport
	var buf *gbytes.Buffer

	BeforeEach(func() {
		failFuncCall = messagedCall{}
		skipFuncCall = messagedCall{}
		offset = 3
		reportToReturn = types.SpecReport{}

		failFunc = func(message string, callerSkip ...int) {
			failFuncCall.message = message
			failFuncCall.callerSkip = callerSkip
		}

		skipFunc = func(message string, callerSkip ...int) {
			skipFuncCall.message = message
			skipFuncCall.callerSkip = callerSkip
		}

		reportFunc = func() types.SpecReport {
			return reportToReturn
		}

		buf = gbytes.NewBuffer()

		t = testingtproxy.New(buf, failFunc, skipFunc, DeferCleanup, reportFunc, offset)
	})

	Describe("Cleanup", Ordered, func() {
		var didCleanupAfter bool
		It("supports cleanup", func() {
			Ω(didCleanupAfter).Should(BeFalse())
			t.Cleanup(func() {
				didCleanupAfter = true
			})
		})

		It("ran cleanup after the last test", func() {
			Ω(didCleanupAfter).Should(BeTrue())
		})
	})

	Describe("Setenv", func() {
		Context("when the environment variable does not exist", Ordered, func() {
			const key = "FLOOP_FLARP_WIBBLE_BLARP"

			BeforeAll(func() {
				os.Unsetenv(key)
			})

			It("sets the environment variable", func() {
				t.Setenv(key, "HELLO")
				Ω(os.Getenv(key)).Should(Equal("HELLO"))
			})

			It("cleans up after itself", func() {
				_, exists := os.LookupEnv(key)
				Ω(exists).Should(BeFalse())
			})
		})

		Context("when the environment variable does exist", Ordered, func() {
			const key = "FLOOP_FLARP_WIBBLE_BLARP"
			const originalValue = "HOLA"

			BeforeAll(func() {
				os.Setenv(key, originalValue)
			})

			It("sets it", func() {
				t.Setenv(key, "HELLO")
				Ω(os.Getenv(key)).Should(Equal("HELLO"))
			})

			It("cleans up after itself", func() {
				Ω(os.Getenv(key)).Should(Equal("HOLA"))
			})

			AfterAll(func() {
				os.Unsetenv(key)
			})
		})
	})

	Describe("TempDir", Ordered, func() {
		var tempDirA, tempDirB string

		It("creates temporary directories", func() {
			tempDirA = t.TempDir()
			tempDirB = t.TempDir()
			Ω(tempDirA).Should(BeADirectory())
			Ω(tempDirB).Should(BeADirectory())
			Ω(tempDirA).ShouldNot(Equal(tempDirB))
		})

		It("cleans up after itself", func() {
			Ω(tempDirA).ShouldNot(BeADirectory())
			Ω(tempDirB).ShouldNot(BeADirectory())
		})
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
		reportToReturn.ContainerHierarchyTexts = []string{"C.S."}
		reportToReturn.LeafNodeText = "Lewis"
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

	It("returns the state of the test when asked if it was skipped", func() {
		reportToReturn.State = types.SpecStatePassed
		Ω(t.Skipped()).Should(BeFalse())
		reportToReturn.State = types.SpecStateSkipped
		Ω(t.Skipped()).Should(BeTrue())
	})
})

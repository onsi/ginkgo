package suite_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var dynamicallyGeneratedTests = []string{}

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	dynamicallyGeneratedTests = []string{"Test A", "Test B", "Test C"}
	RunSpecs(t, "Suite")
}

var numBeforeSuiteRuns = 0
var numAfterSuiteRuns = 0
var numDynamicallyGeneratedTests = 0

var _ = BeforeSuite(func() {
	numBeforeSuiteRuns++
})

var _ = AfterSuite(func() {
	numAfterSuiteRuns++
	Ω(numBeforeSuiteRuns).Should(Equal(1))
	Ω(numAfterSuiteRuns).Should(Equal(1))
	Ω(numDynamicallyGeneratedTests).Should(Equal(3), "Expected three test to be dynamically generated")
})

var _ = Describe("Top-level cotnainer node lifecycle", func() {
	for _, test := range dynamicallyGeneratedTests {
		numDynamicallyGeneratedTests += 1
		It(fmt.Sprintf("runs dynamically generated test: %s", test), func() {
			Ω(true).Should(BeTrue())
		})
	}
})

//Fakes
type fakeTestingT struct {
	didFail bool
}

func (fakeT *fakeTestingT) Fail() {
	fakeT.didFail = true
}

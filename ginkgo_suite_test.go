package ginkgo

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"

	"math/rand"
	"testing"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ginkgo")
}

//Helpers

func shuffleStrings(arr []string, seed int64) []string {
	r := rand.New(rand.NewSource(seed))
	permutation := r.Perm(len(arr))
	shuffled := make([]string, len(arr))
	for i, j := range permutation {
		shuffled[i] = arr[j]
	}

	return shuffled
}

//Fakes

type fakeTestingT struct {
	didFail bool
}

func (fakeT *fakeTestingT) Fail() {
	fakeT.didFail = true
}

type fakeReporter struct {
	config config.GinkgoConfigType

	beginSummary            *types.SuiteSummary
	exampleWillRunSummaries []*types.ExampleSummary
	exampleSummaries        []*types.ExampleSummary
	endSummary              *types.SuiteSummary
}

func (fakeR *fakeReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
	fakeR.config = config
	fakeR.beginSummary = summary
}

func (fakeR *fakeReporter) ExampleWillRun(exampleSummary *types.ExampleSummary) {
	fakeR.exampleWillRunSummaries = append(fakeR.exampleWillRunSummaries, exampleSummary)
}

func (fakeR *fakeReporter) ExampleDidComplete(exampleSummary *types.ExampleSummary) {
	fakeR.exampleSummaries = append(fakeR.exampleSummaries, exampleSummary)
}

func (fakeR *fakeReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	fakeR.endSummary = summary
}

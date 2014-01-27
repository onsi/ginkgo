package ginkgo

import (
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

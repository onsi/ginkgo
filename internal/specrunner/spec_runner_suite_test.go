package specrunner_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSpecRunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spec Runner Suite")
}

type fakeTestingT struct {
	didFail bool
}

func (fakeT *fakeTestingT) Fail() {
	fakeT.didFail = true
}

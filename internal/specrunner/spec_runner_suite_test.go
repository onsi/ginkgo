package specrunner_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSpecRunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spec Runner Suite")
}

package labels_fixture_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLabelsFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LabelsFixture Suite")
}

var set1 = Label("dog", "cat", "cow")

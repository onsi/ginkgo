package types_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/internal/test_helpers"
	. "github.com/onsi/gomega"
)

func TestTypes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Types Suite")
}

var anchors test_helpers.Anchors

var _ = BeforeSuite(func() {
	var err error
	anchors, err = test_helpers.LoadAnchors(test_helpers.DOCS, "../")
	Ω(err).ShouldNot(HaveOccurred())
})

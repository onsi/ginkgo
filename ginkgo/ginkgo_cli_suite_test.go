package main

import (
	"testing"

	"github.com/onsi/ginkgo/v2/internal/test_helpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGinkgoCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ginkgo CLI Suite")
}

var anchors test_helpers.Anchors
var _ = BeforeSuite(func() {
	var err error
	anchors, err = test_helpers.LoadAnchors(test_helpers.DOCS, "../")
	Ω(err).ShouldNot(HaveOccurred())
})

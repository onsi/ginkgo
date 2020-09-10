package main

import (
	"testing"

	"github.com/onsi/ginkgo/internal/test_helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGinkgoCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ginkgo CLI Suite")
}

var DOC_ANCHORS = test_helpers.LoadMarkdownHeadingAnchors("../docs/index.md")

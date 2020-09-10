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

var DOC_ANCHORS = test_helpers.LoadMarkdownHeadingAnchors("../docs/index.md")
var DEPRECATION_ANCHORS = test_helpers.LoadMarkdownHeadingAnchors("../docs/MIGRATING_TO_V2.md")

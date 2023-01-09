package nondeterministic_fixture_test

import (
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

func TestNondeterministicFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NondeterministicFixture Suite")
}

var _ = ReportAfterSuite("ensure all specs ran correctly", func(report types.Report) {
	specs := report.SpecReports.WithLeafNodeType(types.NodeTypeIt)
	orderedTexts := []string{}
	textCounts := map[string]int{}
	for _, spec := range specs {
		text := spec.FullText()
		textCounts[text] += 1
		if strings.HasPrefix(text, "ordered") {
			orderedTexts = append(orderedTexts, spec.LeafNodeText)
		}
	}

	By("ensuring there are no duplicates")
	for text, count := range textCounts {
		Ω(count).Should(Equal(1), text)
	}

	By("ensuring ordered specs are strictly preserved")
	Ω(orderedTexts).Should(Equal([]string{"always", "runs", "in", "order"}))
})

package preview_fixture_test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPreviewFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	if os.Getenv("PREVIEW") == "true" {
		report := PreviewSpecs("PreviewFixture Suite", Label("suite-label"))
		for _, spec := range report.SpecReports {
			fmt.Println(spec.State, spec.FullText())
		}
	}
	if os.Getenv("RUN") == "true" {
		RunSpecs(t, "PreviewFixture Suite", Label("suite-label"))
	}
}

var _ = Describe("specs", func() {
	It("A", Label("elephant"), func() {

	})

	It("B", Label("elephant"), func() {

	})

	It("C", func() {

	})

	It("D", func() {

	})
})

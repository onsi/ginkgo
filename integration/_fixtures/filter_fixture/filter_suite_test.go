package filter_fixture_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFilterFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FilterFixture Suite", Label("TopLevelLabel"))
}

var _ = BeforeEach(func() {
	config, _ := GinkgoConfiguration()
	Ω(GinkgoLabelFilter()).Should(Equal(config.LabelFilter))
})

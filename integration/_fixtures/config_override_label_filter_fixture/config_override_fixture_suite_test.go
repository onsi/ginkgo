package config_override_label_filter_fixture_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfigOverrideFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, reporterConfig := GinkgoConfiguration()
	suiteConfig.LabelFilter = "!NORUN"
	RunSpecs(t, "ConfigOverrideFixture Suite", suiteConfig, reporterConfig)
}

var _ = Describe("tests", func() {
	It("never runs", Label("NORUN"), func() {
		Ω(true).Should(BeFalse())
	})

	It("runs", func() {
		Ω(true).Should(BeTrue())
	})
})

package config_override_label_filter_fixture_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfigOverrideFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, reporterConfig := GinkgoConfiguration()
	suiteConfig.MustPassRepeatedly = 10
	RunSpecs(t, "ConfigOverrideFixture Suite", suiteConfig, reporterConfig)
}

var _ = Describe("tests", func() {
	It("suite config overrides decorator", MustPassRepeatedly(2), func() {
		Ω(CurrentSpecReport().MaxMustPassRepeatedly).Should(Equal(10))
	})
})

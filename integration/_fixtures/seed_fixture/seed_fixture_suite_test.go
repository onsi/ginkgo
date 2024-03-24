package seed_fixture_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSeedFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SeedFixture Suite")
}

var _ = BeforeSuite(func() {
	Ω(GinkgoRandomSeed()).Should(Equal(int64(0)))
})

var _ = It("has the expected seed (namely, 0)", func() {
	Ω(GinkgoRandomSeed()).Should(Equal(int64(0)))
})

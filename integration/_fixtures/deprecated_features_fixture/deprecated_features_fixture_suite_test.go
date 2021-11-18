package deprecated_features_fixture_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDeprecatedFeaturesFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DeprecatedFeaturesFixture Suite")
}

var _ = It("tries to perform an async assertion", func(done Done) {
	close(done)
})

var _ = Measure("tries to perform a measurement", func(b Benchmarker) {
	
})
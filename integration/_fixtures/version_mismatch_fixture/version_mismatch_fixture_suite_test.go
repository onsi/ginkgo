package version_mismatch_fixture_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVersionMismatchFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VersionMismatchFixture Suite")
}

var _ = It("has the error message we expect...", func() {

})

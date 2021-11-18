//+build integration

package notaggedtestsfixture

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNoTaggedTestsFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "No Tagged Tests Fixture Suite")
}

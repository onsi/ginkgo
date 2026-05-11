package helpergo_fixture_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHelpergoFixture(t *testing.T) {
	// See github.com/ginkgo/v2/types/PruneStack
	os.Setenv("GINKGO_PRUNE_STACK", "FALSE") // ouch.
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helpergo Suite")
}

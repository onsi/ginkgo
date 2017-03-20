package fixture_suite_test

import (
	. "github.com/onsi/ginkgo"

	"testing"
)

func TestPassingSuiteSetup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FixtureSuite Suite")
}

package malformed_fixture_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMalformedFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MalformedFixture Suite")
}

package malformed_fixture_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMalformedFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MalformedFixture Suite")
}

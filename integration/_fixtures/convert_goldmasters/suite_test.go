package convert_fixtures_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConvertFixtures(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ConvertFixtures Suite")
}

package convert_fixtures_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConvert_fixtures(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Convert_fixtures Suite")
}

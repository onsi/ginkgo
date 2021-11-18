package third_package_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestThirdPackage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ThirdPackage Suite")
}

var _ = It("doesn't cover anything", func() {})

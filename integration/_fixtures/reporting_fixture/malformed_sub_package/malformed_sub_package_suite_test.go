package malformed_sub_package_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMalformedSubPackage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MalformedSubPackage Suite")
}

NO_COMPILE.FORYOU!
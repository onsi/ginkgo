package reporting_sub_package_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestReportingSubPackage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Reporting SubPackage Suite")
}

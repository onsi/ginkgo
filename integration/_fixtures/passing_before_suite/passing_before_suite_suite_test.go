package passing_before_suite_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPassing_before_suite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Passing_before_suite Suite")
}

var a string

var _ = BeforeSuite(func() {
	a = "ran before suite"
	println("BEFORE SUITE")
})

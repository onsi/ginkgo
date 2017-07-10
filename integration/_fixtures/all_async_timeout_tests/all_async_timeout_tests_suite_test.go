package all_async_timeout_tests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAll_Async_timeout_tests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "All_async_timeout_tests Suite")
}

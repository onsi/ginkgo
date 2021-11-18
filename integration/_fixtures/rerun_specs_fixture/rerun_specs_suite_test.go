package rerun_specs_fixture_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRerunSpecs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RerunSpecs Suite")
	RunSpecs(t, "RerunSpecs Suite - part deux")
}

var _ = It("tries twice", func() {

})

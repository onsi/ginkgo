package codelocation_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCodelocation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CodeLocation Suite")
}

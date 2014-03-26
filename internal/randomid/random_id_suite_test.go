package randomid_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRandomid(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Randomid Suite")
}

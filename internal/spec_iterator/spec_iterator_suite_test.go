package spec_iterator_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSpecIterator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SpecIterator Suite")
}

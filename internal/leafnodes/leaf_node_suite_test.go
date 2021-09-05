package leafnodes_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLeafNode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LeafNode Suite")
}

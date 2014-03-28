package leafnode_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRunnableNode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RunnableNode Suite")
}

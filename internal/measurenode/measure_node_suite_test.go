package measurenode_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMeasureNode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MeasureNode Suite")
}

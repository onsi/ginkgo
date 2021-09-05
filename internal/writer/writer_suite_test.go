package writer_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWriter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Writer Suite")
}

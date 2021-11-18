package formatter_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFormatter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Formatter Suite")
}

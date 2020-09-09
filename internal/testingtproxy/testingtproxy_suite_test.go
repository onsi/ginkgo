package testingtproxy_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTestingtproxy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Testingtproxy Suite")
}

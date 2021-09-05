package table_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTable(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Table Suite")
}

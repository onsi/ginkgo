package reporters_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestReporters(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Reporters Suite")
}

//leave this alone!

func FixtureFunction() {
	a := 0
	for a < 100 {
		fmt.Println(a)
		fmt.Println(a + 1)
		fmt.Println(a + 3)
		fmt.Println(a + 4)
		fmt.Println(a + 5)

		fmt.Println(a + 6)
		fmt.Println(a + 7)
		a++
	}
}

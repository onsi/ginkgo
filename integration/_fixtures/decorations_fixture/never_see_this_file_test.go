package decorations_fixture_test

import . "github.com/onsi/ginkgo"

func OffsetIt() {
	It("is offset", Offset(1), func() {
	})
}

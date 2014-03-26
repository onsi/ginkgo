package randomid_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/internal/randomid"
	. "github.com/onsi/gomega"
)

var _ = Describe("New", func() {
	It("should generate a random guid", func() {
		a := randomid.New()
		b := randomid.New()
		Ω(a).ShouldNot(BeEmpty())
		Ω(b).ShouldNot(BeEmpty())
		Ω(a).ShouldNot(Equal(b))

		IDRegexp := "[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}"
		Ω(a).Should(MatchRegexp(IDRegexp))
		Ω(b).Should(MatchRegexp(IDRegexp))
	})
})

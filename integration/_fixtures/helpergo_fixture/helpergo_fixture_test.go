package helpergo_fixture_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
)

var _ = Describe("Exercise different failure modes", func() {

	It("fails assertion in user function", func() {
		GinkgoHelperGo(func(helperFail func(string, ...int)) {
			func() {
				/* fail location */ Ω("a user assertion failure").Should(Equal("nope"))
			}()
			Fail("SHOULD NOT SEE THIS")
		})
	})

	It("fails assertion in helper function before user function", func() {
		/* fail location */ GinkgoHelperGo(func(helperFail func(string, ...int)) {
			g := gomega.NewGomega(helperFail)
			g.Expect("a helper assertion 1 failure").Should(Equal("nope"))
			func() {
				Fail("SHOULD NOT SEE THIS")
			}()
		})
		Fail("SHOULD NOT SEE THIS")
	})

	It("fails assertion in helper function after user function", func() {
		/* fail location */ GinkgoHelperGo(func(helperFail func(string, ...int)) {
			g := gomega.NewGomega(helperFail)
			func() {
				/* It's all fine! */
			}()
			g.Expect("a helper assertion 2 failure").Should(Equal("nope"))
		})
		Fail("SHOULD NOT SEE THIS")
	})

	It("fails panics in user function", func() {
		GinkgoHelperGo(func(helperFail func(string, ...int)) {
			/* test location */ panic("JKL305")
			Fail("SHOULD NOT SEE THIS")
		})
	})

})

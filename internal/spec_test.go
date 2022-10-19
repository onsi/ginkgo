package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal"
	. "github.com/onsi/gomega"
)

var _ = Describe("Spec and Specs", func() {
	Describe("spec.Text", func() {
		Context("when the spec has nodes with texts", func() {
			It("returns the concatenated texts of its nodes (omitting any empty texts)", func() {
				spec := S(N(), N("Oh death,"), N(), N("where is"), N("thy"), N(), N("string?"))
				Ω(spec.Text()).Should(Equal("Oh death, where is thy string?"))
			})
		})

		Context("when the spec has no nodes", func() {
			It("returns the empty string", func() {
				Ω(Spec{}.Text()).Should(BeZero())
			})
		})
	})

	Describe("spec.SubjectID()", func() {
		It("returns the ID of the spec's it node", func() {
			nIt := N(ntIt)
			spec := S(N(ntCon), N(ntCon), N(ntBef), nIt, N(ntAf))
			Ω(spec.SubjectID()).Should(Equal(nIt.ID))
		})
	})

	Describe("spec.FirstNodeWithType", func() {
		Context("when there are matching nodes", func() {
			It("returns the first node matching any of the passed in node types", func() {
				nBef := N(ntBef)
				nIt := N(ntIt)
				spec := S(N(ntCon), N(ntAf), nBef, N(ntBef), nIt, N(ntAf))
				Ω(spec.FirstNodeWithType(ntIt | ntBef)).Should(Equal(nBef))
			})
		})

		Context("when no nodes match", func() {
			It("returns zero", func() {
				spec := S(N(ntCon), N(ntIt), N(ntAf))
				Ω(spec.FirstNodeWithType(ntBef)).Should(BeZero())
			})
		})
	})

	Describe("spec.FlakeAttempts", func() {
		Context("when none of the nodes have FlakeAttempt", func() {
			It("returns 0", func() {
				spec := S(N(ntCon), N(ntCon), N(ntIt))
				Ω(spec.FlakeAttempts()).Should(Equal(0))
			})
		})

		Context("when a node has FlakeAttempt set", func() {
			It("returns that FlakeAttempt", func() {
				spec := S(N(ntCon, FlakeAttempts(3)), N(ntCon), N(ntIt))
				Ω(spec.FlakeAttempts()).Should(Equal(3))

				spec = S(N(ntCon), N(ntCon, FlakeAttempts(2)), N(ntIt))
				Ω(spec.FlakeAttempts()).Should(Equal(2))

				spec = S(N(ntCon), N(ntCon), N(ntIt, FlakeAttempts(4)))
				Ω(spec.FlakeAttempts()).Should(Equal(4))
			})
		})

		Context("when multiple nodes have FlakeAttempt", func() {
			It("returns the inner-most nested FlakeAttempt", func() {
				spec := S(N(ntCon, FlakeAttempts(3)), N(ntCon, FlakeAttempts(4)), N(ntIt, FlakeAttempts(2)))
				Ω(spec.FlakeAttempts()).Should(Equal(2))
			})
		})
	})

	Describe("spec.MustPassRepeatedly", func() {
		Context("when none of the nodes have MustPassRepeatedly", func() {
			It("returns 0", func() {
				spec := S(N(ntCon), N(ntCon), N(ntIt))
				Ω(spec.MustPassRepeatedly()).Should(Equal(0))
			})
		})

		Context("when a node has MustPassRepeatedly set", func() {
			It("returns that MustPassRepeatedly", func() {
				spec := S(N(ntCon, MustPassRepeatedly(3)), N(ntCon), N(ntIt))
				Ω(spec.MustPassRepeatedly()).Should(Equal(3))

				spec = S(N(ntCon), N(ntCon, MustPassRepeatedly(2)), N(ntIt))
				Ω(spec.MustPassRepeatedly()).Should(Equal(2))

				spec = S(N(ntCon), N(ntCon), N(ntIt, MustPassRepeatedly(4)))
				Ω(spec.MustPassRepeatedly()).Should(Equal(4))
			})
		})

		Context("when multiple nodes have MustPassRepeatedly", func() {
			It("returns the inner-most nested MustPassRepeatedly", func() {
				spec := S(N(ntCon, MustPassRepeatedly(3)), N(ntCon, MustPassRepeatedly(4)), N(ntIt, MustPassRepeatedly(2)))
				Ω(spec.MustPassRepeatedly()).Should(Equal(2))
			})
		})
	})

	Describe("specs.HasAnySpecsMarkedPending", func() {
		Context("when there are no specs with any nodes marked pending", func() {
			It("returns false", func() {
				specs := Specs{
					S(N(), N(), N()),
					S(N(), N()),
				}

				Ω(specs.HasAnySpecsMarkedPending()).Should(BeFalse())
			})
		})

		Context("when there is at least one spec with a node marked pending", func() {
			It("returns true", func() {
				specs := Specs{
					S(N(), N(), N()),
					S(N(), N(Pending), N()),
					S(N(), N()),
				}

				Ω(specs.HasAnySpecsMarkedPending()).Should(BeTrue())
			})
		})
	})

	Describe("specs.CountWithoutSkip()", func() {
		It("returns the number of specs that have skip set to false", func() {
			specs := Specs{{Skip: false}, {Skip: true}, {Skip: true}, {Skip: false}, {Skip: false}}
			Ω(specs.CountWithoutSkip()).Should(Equal(3))
		})
	})

	Describe("specs.AtIndices", func() {
		It("returns the subset of specs at the specified indices", func() {
			specs := Specs{S(N()), S(N()), S(N()), S(N())}
			Ω(specs.AtIndices(internal.SpecIndices{1, 3})).Should(Equal(Specs{specs[1], specs[3]}))
		})
	})
})

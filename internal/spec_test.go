package internal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Spec and Specs", func() {
	Describe("spec.Text", func() {
		Context("when the spec has nodes with texts", func() {
			It("returns the concatenated texts of its nodes (omitting any empty texts)", func() {
				spec := S(N(), N("Oh death,"), N(), N("where is"), N("thy"), N(), N("sting?"))
				Ω(spec.Text()).Should(Equal("Oh death, where is thy sting?"))
			})
		})

		Context("when the spec has no nodes", func() {
			It("returns the empty string", func() {
				Ω(Spec{}.Text()).Should(BeZero())
			})
		})
	})

	Describe("spec.FirstNodeWithType", func() {
		Context("when there are matching nodes", func() {
			It("returns the first node matching any of the passed in node types", func() {
				nBef := N(ntBef)
				nIt := N(ntIt)
				spec := S(N(ntCon), N(ntAf), nBef, N(ntBef), nIt, N(ntAf))
				Ω(spec.FirstNodeWithType(ntIt, ntBef)).Should(Equal(nBef))
			})
		})

		Context("when no nodes match", func() {
			spec := S(N(ntCon), N(ntIt), N(ntAf))
			Ω(spec.FirstNodeWithType(ntBef)).Should(BeZero())
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
					S(N(), N(MarkedPending(true)), N()),
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

	Describe("partitioning specs", func() {
		var specs Specs
		BeforeEach(func() {
			sharedBefore := N(ntBef)
			sharedContainer := N(ntCon)
			otherSharedContainer := N(ntCon)
			specs = Specs{
				S(sharedBefore, sharedContainer, N(ntIt)),
				S(sharedBefore, N(ntIt)),
				S(N(ntBef), sharedContainer, N(ntIt)),
				S(sharedBefore, N(ntIt)),
				S(N(ntCon), N(ntIt)),
				S(otherSharedContainer, N(ntBef), N(ntIt)),
				S(otherSharedContainer, N(ntBef), N(ntIt)),
			}
		})

		It(`returns a slice of []Specs - where each entry is a group of specs for which
			the first node that matches on of the passed in nodetypes has the same id`, func() {
			Ω(specs.PartitionByFirstNodeWithType(ntIt)).Should(Equal([]Specs{
				Specs{specs[0]},
				Specs{specs[1]},
				Specs{specs[2]},
				Specs{specs[3]},
				Specs{specs[4]},
				Specs{specs[5]},
				Specs{specs[6]},
			}), "partitioning by It returns one grouping per spec as each spec has a unique It")

			Ω(specs.PartitionByFirstNodeWithType(ntIt, ntCon)).Should(Equal([]Specs{
				Specs{specs[0], specs[2]},
				Specs{specs[1]},
				Specs{specs[3]},
				Specs{specs[4]},
				Specs{specs[5], specs[6]},
			}), "partitioning by Container and It groups specs by common Container first, and It second")

			Ω(specs.PartitionByFirstNodeWithType(ntCon)).Should(Equal([]Specs{
				Specs{specs[0], specs[2]},
				Specs{specs[4]},
				Specs{specs[5], specs[6]},
			}), "partitioning by just Container will not pull in specs that have no container")

			Ω(specs.PartitionByFirstNodeWithType(ntAf)).Should(BeEmpty(),
				"partitioning by a node type that doesn't appear matches against no specs and comes back empty")
		})
	})
})

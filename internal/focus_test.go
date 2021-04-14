package internal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/internal"
	"github.com/onsi/ginkgo/types"
)

var _ = Describe("Focus", func() {
	Describe("ApplyNestedFocusToTree", func() {
		It("unfocuses parent nodes that have a focused child node somewhere in their tree", func() {
			tree := TN(N(ntCon, "root", MarkedFocus(true)), //should lose focus
				TN(N(ntCon, "A", MarkedFocus(true)), //should stay focused
					TN(N(ntIt)),
					TN(N(ntIt)),
				),
				TN(N(ntCon),
					TN(N(ntIt)),
					TN(N(ntIt, "B", MarkedFocus(true))), //should stay focused
				),
				TN(N(ntCon, "C", MarkedFocus(true)), //should lose focus
					TN(N(ntIt)),
					TN(N(ntIt, "D", MarkedFocus(true))), //should stay focused
				),
				TN(N(ntCon, "E", MarkedFocus(true)), //should lose focus
					TN(N(ntIt)),
					TN(N(ntCon),
						TN(N(ntIt)),
						TN(N(ntIt, "F", MarkedFocus(true))), // should stay focused
					),
				),
				TN(N(ntCon, "G", MarkedFocus(true)), //should lose focus
					TN(N(ntIt)),
					TN(N(ntCon, "H", MarkedFocus(true)), //should lose focus
						TN(N(ntIt)),
						TN(N(ntIt, "I", MarkedFocus(true))), //should stay focused
					),
					TN(N(ntCon, "J", MarkedFocus(true)), // should stay focused
						TN(N(ntIt)),
					),
				),
			)

			tree = internal.ApplyNestedFocusPolicyToTree(tree)
			Ω(mustFindNodeWithText(tree, "root").MarkedFocus).Should(BeFalse())
			Ω(mustFindNodeWithText(tree, "A").MarkedFocus).Should(BeTrue())
			Ω(mustFindNodeWithText(tree, "B").MarkedFocus).Should(BeTrue())
			Ω(mustFindNodeWithText(tree, "C").MarkedFocus).Should(BeFalse())
			Ω(mustFindNodeWithText(tree, "D").MarkedFocus).Should(BeTrue())
			Ω(mustFindNodeWithText(tree, "E").MarkedFocus).Should(BeFalse())
			Ω(mustFindNodeWithText(tree, "F").MarkedFocus).Should(BeTrue())
			Ω(mustFindNodeWithText(tree, "G").MarkedFocus).Should(BeFalse())
			Ω(mustFindNodeWithText(tree, "H").MarkedFocus).Should(BeFalse())
			Ω(mustFindNodeWithText(tree, "I").MarkedFocus).Should(BeTrue())
			Ω(mustFindNodeWithText(tree, "J").MarkedFocus).Should(BeTrue())
		})

		It("does not unfocus parent nodes if a focused child is the child of a pending child", func() {
			tree := TN(N(ntCon),
				TN(N(ntCon, "A", MarkedFocus(true)), //should stay focused
					TN(N(ntIt)),
					TN(N(ntCon, "B", MarkedPending(true)), //should stay pending
						TN(N(ntIt)),
						TN(N(ntIt, "C", MarkedFocus(true))), //should stay focused
					),
				),
			)

			tree = internal.ApplyNestedFocusPolicyToTree(tree)
			Ω(mustFindNodeWithText(tree, "A").MarkedFocus).Should(BeTrue())
			Ω(mustFindNodeWithText(tree, "B").MarkedPending).Should(BeTrue())
			Ω(mustFindNodeWithText(tree, "C").MarkedFocus).Should(BeTrue())
		})
	})

	Describe("ApplyFocusToSpecs", func() {
		var specs Specs
		var description string
		var conf types.SuiteConfig

		harvestSkips := func(specs Specs) []bool {
			out := []bool{}
			for _, spec := range specs {
				out = append(out, spec.Skip)
			}
			return out
		}

		BeforeEach(func() {
			description = "Silmarillion Suite"
			conf = types.SuiteConfig{}
		})

		Context("when there are specs with nodes marked pending", func() {
			BeforeEach(func() {
				specs = Specs{
					S(N(), N()),
					S(N(), N()),
					S(N(), N(MarkedPending(true))),
					S(N(), N()),
					S(N(MarkedPending(true))),
				}
			})

			It("skips those specs", func() {
				specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, conf)
				Ω(harvestSkips(specs)).Should(Equal([]bool{false, false, true, false, true}))
				Ω(hasProgrammaticFocus).Should(BeFalse())
			})
		})

		Context("when there are specs with nodes marked focused", func() {
			BeforeEach(func() {
				specs = Specs{
					S(N(), N()),
					S(N(), N()),
					S(N(), N(MarkedFocus(true))),
					S(N()),
					S(N(MarkedFocus(true))),
				}
			})
			It("skips any other specs and notes that it has programmatic focus", func() {
				specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, conf)
				Ω(harvestSkips(specs)).Should(Equal([]bool{true, true, false, true, false}))
				Ω(hasProgrammaticFocus).Should(BeTrue())
			})

			Context("when the specs with nodes marked focused also have nodes marked pending ", func() {
				BeforeEach(func() {
					specs = Specs{
						S(N(), N()),
						S(N(), N()),
						S(N(MarkedPending(true)), N(MarkedFocus(true))),
						S(N()),
					}
				})
				It("does not skip any other specs and notes that it does not have programmatic focus", func() {
					specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, conf)
					Ω(harvestSkips(specs)).Should(Equal([]bool{false, false, true, false}))
					Ω(hasProgrammaticFocus).Should(BeFalse())
				})
			})
		})

		Context("when there are focus strings and/or skip strings configured", func() {
			BeforeEach(func() {
				specs = Specs{
					S(N("blue"), N("dragon")),
					S(N("blue"), N("Dragon")),
					S(N("red dragon"), N()),
					S(N("green dragon"), N()),
					S(N(MarkedPending(true)), N("blue Dragon")),
					S(N("yellow dragon")),
					S(N(MarkedFocus(true), "yellow dragon")),
				}
			})

			Context("when there are focus strings configured", func() {
				BeforeEach(func() {
					conf.FocusStrings = []string{"blue [dD]ra", "(red|green) dragon"}
				})

				It("overrides any programmatic focus, runs only specs that match the focus string, and continues to skip specs with nodes marked pending", func() {
					specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, conf)
					Ω(harvestSkips(specs)).Should(Equal([]bool{false, false, false, false, true, true, true}))
					Ω(hasProgrammaticFocus).Should(BeFalse())
				})

				It("includes the description string in the search", func() {
					conf.FocusStrings = []string{"Silmaril"}
					specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, conf)
					Ω(harvestSkips(specs)).Should(Equal([]bool{false, false, false, false, true, false, false}))
					Ω(hasProgrammaticFocus).Should(BeFalse())
				})
			})

			Context("when there are skip strings configured", func() {
				BeforeEach(func() {
					conf.SkipStrings = []string{"blue [dD]ragon", "red dragon"}
				})

				It("overrides any programmatic focus, and runs specs that don't match the skip strings, and continues to skip specs with nodes marked pending", func() {
					specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, conf)
					Ω(harvestSkips(specs)).Should(Equal([]bool{true, true, true, false, true, false, false}))
					Ω(hasProgrammaticFocus).Should(BeFalse())
				})

				It("includes the description string in the search", func() {
					conf.SkipStrings = []string{"Silmaril"}
					specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, conf)
					Ω(harvestSkips(specs)).Should(Equal([]bool{true, true, true, true, true, true, true}))
					Ω(hasProgrammaticFocus).Should(BeFalse())
				})
			})

			Context("when skip and focus are configured", func() {
				BeforeEach(func() {
					conf.FocusStrings = []string{"blue [dD]ragon", "(red|green) dragon"}
					conf.SkipStrings = []string{"red dragon", "Dragon"}
				})

				It("ORs both together", func() {
					specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, conf)
					Ω(harvestSkips(specs)).Should(Equal([]bool{false, true, true, false, true, true, true}))
					Ω(hasProgrammaticFocus).Should(BeFalse())
				})
			})
		})

		Context("when configured to RegexScansFilePath", func() {
			BeforeEach(func() {
				specs = Specs{
					S(N(CL("file_a"))),
					S(N(CL("file_b"))),
					S(N(CL("file_b"), MarkedPending(true))),
					S(N(CL("c", MarkedFocus(true)))),
				}

				conf.RegexScansFilePath = true
				conf.FocusStrings = []string{"file_"}
				conf.SkipStrings = []string{"_a"}
			})

			It("includes the codelocation filename in the search for focus and skip strings", func() {
				specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, conf)
				Ω(harvestSkips(specs)).Should(Equal([]bool{true, false, true, true}))
				Ω(hasProgrammaticFocus).Should(BeFalse())

			})
		})
	})
})

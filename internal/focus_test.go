package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/internal"
	"github.com/onsi/ginkgo/v2/types"
)

var _ = Describe("Focus", func() {
	Describe("ApplyNestedFocusToTree", func() {
		It("unfocuses parent nodes that have a focused child node somewhere in their tree", func() {
			tree := TN(N(ntCon, "root", Focus), //should lose focus
				TN(N(ntCon, "A", Focus), //should stay focused
					TN(N(ntIt)),
					TN(N(ntIt)),
				),
				TN(N(ntCon),
					TN(N(ntIt)),
					TN(N(ntIt, "B", Focus)), //should stay focused
				),
				TN(N(ntCon, "C", Focus), //should lose focus
					TN(N(ntIt)),
					TN(N(ntIt, "D", Focus)), //should stay focused
				),
				TN(N(ntCon, "E", Focus), //should lose focus
					TN(N(ntIt)),
					TN(N(ntCon),
						TN(N(ntIt)),
						TN(N(ntIt, "F", Focus)), // should stay focused
					),
				),
				TN(N(ntCon, "G", Focus), //should lose focus
					TN(N(ntIt)),
					TN(N(ntCon, "H", Focus), //should lose focus
						TN(N(ntIt)),
						TN(N(ntIt, "I", Focus)), //should stay focused
					),
					TN(N(ntCon, "J", Focus), // should stay focused
						TN(N(ntIt)),
					),
				),
			)

			internal.ApplyNestedFocusPolicyToTree(tree)
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
				TN(N(ntCon, "A", Focus), //should stay focused
					TN(N(ntIt)),
					TN(N(ntCon, "B", Pending), //should stay pending
						TN(N(ntIt)),
						TN(N(ntIt, "C", Focus)), //should stay focused
					),
				),
			)

			internal.ApplyNestedFocusPolicyToTree(tree)
			Ω(mustFindNodeWithText(tree, "A").MarkedFocus).Should(BeTrue())
			Ω(mustFindNodeWithText(tree, "B").MarkedPending).Should(BeTrue())
			Ω(mustFindNodeWithText(tree, "C").MarkedFocus).Should(BeTrue())
		})
	})

	Describe("ApplyFocusToSpecs", func() {
		var specs Specs
		var description string
		var suiteLabels Labels
		var suiteSemVerConstraints SemVerConstraints
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
			suiteLabels = Labels{"SuiteLabel", "TopLevelLabel"}
			conf = types.SuiteConfig{}
		})

		Context("when there are specs with nodes marked pending", func() {
			BeforeEach(func() {
				specs = Specs{
					S(N(), N()),
					S(N(), N()),
					S(N(), N(Pending)),
					S(N(), N()),
					S(N(Pending)),
				}
			})

			It("skips those specs", func() {
				specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, suiteLabels, suiteSemVerConstraints, conf)
				Ω(harvestSkips(specs)).Should(Equal([]bool{false, false, true, false, true}))
				Ω(hasProgrammaticFocus).Should(BeFalse())
			})
		})

		Context("when there are specs with nodes marked focused", func() {
			BeforeEach(func() {
				specs = Specs{
					S(N(), N()),
					S(N(), N()),
					S(N(), N(Focus)),
					S(N()),
					S(N(Focus)),
				}
			})
			It("skips any other specs and notes that it has programmatic focus", func() {
				specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, suiteLabels, suiteSemVerConstraints, conf)
				Ω(harvestSkips(specs)).Should(Equal([]bool{true, true, false, true, false}))
				Ω(hasProgrammaticFocus).Should(BeTrue())
			})

			Context("when the specs with nodes marked focused also have nodes marked pending ", func() {
				BeforeEach(func() {
					specs = Specs{
						S(N(), N()),
						S(N(), N()),
						S(N(Pending), N(Focus)),
						S(N()),
					}
				})
				It("does not skip any other specs and notes that it does not have programmatic focus", func() {
					specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, suiteLabels, suiteSemVerConstraints, conf)
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
					S(N(Pending), N("blue Dragon")),
					S(N("yellow dragon")),
					S(N("yellow dragon")),
				}
			})

			Context("when there are focus strings configured", func() {
				BeforeEach(func() {
					conf.FocusStrings = []string{"blue [dD]ra", "(red|green) dragon"}
				})

				It("overrides any programmatic focus, runs only specs that match the focus string, and continues to skip specs with nodes marked pending", func() {
					specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, suiteLabels, suiteSemVerConstraints, conf)
					Ω(harvestSkips(specs)).Should(Equal([]bool{false, false, false, false, true, true, true}))
					Ω(hasProgrammaticFocus).Should(BeFalse())
				})

				It("includes the description string in the search", func() {
					conf.FocusStrings = []string{"Silmaril"}
					specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, suiteLabels, suiteSemVerConstraints, conf)
					Ω(harvestSkips(specs)).Should(Equal([]bool{false, false, false, false, true, false, false}))
					Ω(hasProgrammaticFocus).Should(BeFalse())
				})
			})

			Context("when there are skip strings configured", func() {
				BeforeEach(func() {
					conf.SkipStrings = []string{"blue [dD]ragon", "red dragon"}
				})

				It("overrides any programmatic focus, and runs specs that don't match the skip strings, and continues to skip specs with nodes marked pending", func() {
					specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, suiteLabels, suiteSemVerConstraints, conf)
					Ω(harvestSkips(specs)).Should(Equal([]bool{true, true, true, false, true, false, false}))
					Ω(hasProgrammaticFocus).Should(BeFalse())
				})

				It("includes the description string in the search", func() {
					conf.SkipStrings = []string{"Silmaril"}
					specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, suiteLabels, suiteSemVerConstraints, conf)
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
					specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, suiteLabels, suiteSemVerConstraints, conf)
					Ω(harvestSkips(specs)).Should(Equal([]bool{false, true, true, false, true, true, true}))
					Ω(hasProgrammaticFocus).Should(BeFalse())
				})
			})
		})

		Context("when configured to focus/skip files", func() {
			BeforeEach(func() {
				specs = Specs{
					S(N(CL("file_a", 1))),               //include because "file_:1" is in FocusFiles
					S(N(CL("file_b", 3, "file_b", 15))), //include because "file_:15-21" is in FocusFiles
					S(N(CL("file_b", 17))),              //skip because "_b:17" is in SkipFiles
					S(N(CL("file_b", 20), Pending)),     //skip because spec is flagged pending
					S(N(CL("c", 3))),                    //skip because "c" is not in FocusFiles
					S(N(CL("d", 17))),                   //include because "d " is in FocusFiles
				}

				conf.FocusFiles = []string{"file_:1,15-21", "d"}
				conf.SkipFiles = []string{"_b:17"}
			})

			It("applies a file-based focus and skip filter", func() {
				specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, suiteLabels, suiteSemVerConstraints, conf)
				Ω(harvestSkips(specs)).Should(Equal([]bool{false, false, true, true, true, false}))
				Ω(hasProgrammaticFocus).Should(BeFalse())
			})
		})

		Context("when configured with a label filter", func() {
			BeforeEach(func() {
				conf.LabelFilter = "(cat || cow) && !fish"
				specs = Specs{
					S(N(ntCon, Label("cat", "dog")), N(ntIt, "A", Label("fish"))),  //skip because fish
					S(N(ntCon, Label("cat", "dog")), N(ntIt, "B", Label("apple"))), //include because has cat and not fish
					S(N(ntCon, Label("dog")), N(ntIt, "C", Label("apple"))),        //skip because no cat or cow
					S(N(ntCon, Label("cow")), N(ntIt, "D", Label("fish"))),         //skip because fish
					S(N(ntCon, Label("cow")), N(ntIt, "E")),                        //include because cow and no fish
					S(N(ntCon, Label("cow")), N(ntIt, "F", Pending)),               //skip because pending
				}
			})

			It("applies the label filters", func() {
				specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, suiteLabels, suiteSemVerConstraints, conf)
				Ω(harvestSkips(specs)).Should(Equal([]bool{true, false, true, true, false, true}))
				Ω(hasProgrammaticFocus).Should(BeFalse())

			})
		})

		Context("when configured with a label set filter", func() {
			BeforeEach(func() {
				conf.LabelFilter = "Feature: consistsOf {A, B} || Feature: containsAny C"
				specs = Specs{
					S(N(ntCon, Label("Feature:A", "dog")), N(ntIt, "A", Label("fish"))),               //skip because fish no feature:B
					S(N(ntCon, Label("Feature:A", "dog")), N(ntIt, "B", Label("apple", "Feature:B"))), //include because has Feature:A and Feature:B
					S(N(ntCon, Label("Feature:A")), N(ntIt, "C", Label("Feature:B", "Feature:D"))),    //skip because it has Feature:D
					S(N(ntCon, Label("Feature:C")), N(ntIt, "D", Label("fish", "Feature:D"))),         //include because it has Feature:C
					S(N(ntCon, Label("cow")), N(ntIt, "E")),                                           //skip because no Feature:
					S(N(ntCon, Label("Feature:A", "Feature:B")), N(ntIt, "F", Pending)),               //skip because pending
				}
			})

			It("applies the label filters", func() {
				specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, suiteLabels, suiteSemVerConstraints, conf)
				Ω(harvestSkips(specs)).Should(Equal([]bool{true, false, true, false, true, true}))
				Ω(hasProgrammaticFocus).Should(BeFalse())

			})
		})

		Context("when configured with a label filter that filters on the suite level label", func() {
			BeforeEach(func() {
				conf.LabelFilter = "cat && TopLevelLabel"
				specs = Specs{
					S(N(ntCon, Label("cat", "dog")), N(ntIt, "A", Label("fish"))), //include because cat and suite has TopLevelLabel
					S(N(ntCon, Label("dog")), N(ntIt, "B", Label("apple"))),       //skip because no cat
				}
			})
			It("honors the suite level label", func() {
				specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, suiteLabels, suiteSemVerConstraints, conf)
				Ω(harvestSkips(specs)).Should(Equal([]bool{false, true}))
				Ω(hasProgrammaticFocus).Should(BeFalse())
			})
		})

		Context("when configured with focus/skip files, focus/skip strings, and label filters", func() {
			BeforeEach(func() {
				specs = Specs{
					S(N("dog", CL("file_a", 1), Label("brown"))),                   //include because "file_:1" is in FocusFiles and "dog" is in FocusStrings and has "brown" label
					S(N("dog", CL("file_a", 1), Label("white"))),                   //skip because does not have "brown" label
					S(N("dog cat", CL("file_b", 3, "file_b", 15), Label("brown"))), //skip because "file_:15-21" is in FocusFiles but "cat" is in SkipStirngs
					S(N("fish", CL("file_b", 17), Label("brown"))),                 //skip because "_b:17" is in SkipFiles, even though "fish" is in FocusStrings
					S(N("biscuit", CL("file_b", 20), Pending, Label("brown"))),     //skip because spec is flagged pending
					S(N("pony", CL("c", 3), Label("brown"))),                       //skip because "c" is not in FocusFiles or FocusStrings
					S(N("goat", CL("d", 17), Label("brown"))),                      //skip because "goat" is in FocusStrings but "d" is not in FocusFiles
				}

				conf.FocusFiles = []string{"file_:1,15-21"}
				conf.SkipFiles = []string{"_b:17"}
				conf.FocusStrings = []string{"goat", "dog", "fish", "biscuit"}
				conf.SkipStrings = []string{"cat"}
				conf.LabelFilter = "brown"
			})

			It("applies all filters", func() {
				specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, suiteLabels, suiteSemVerConstraints, conf)
				Ω(harvestSkips(specs)).Should(Equal([]bool{false, true, true, true, true, true, true}))
				Ω(hasProgrammaticFocus).Should(BeFalse())
			})
		})

		Context("when configured with focus/skip files, focus/skip strings, and label filters and there is a programmatic focus", func() {
			BeforeEach(func() {
				specs = Specs{
					S(N("dog", CL("file_a", 1), Label("brown"))),                   //skip because "file_:1" is in FocusFiles and "dog" is in FocusStrings and has "brown" label but a different spec has a programmatic focus
					S(N("dog", CL("file_a", 17), Label("brown"), Focus)),           //include because "file_:15-21" is in FocusFiles and "dog" is in FocusStrings and has "brown" label
					S(N("dog", CL("file_a", 1), Label("white"), Focus)),            //skip because does not have "brown" label
					S(N("dog cat", CL("file_b", 3, "file_b", 15), Label("brown"))), //skip because "file_:15-21" is in FocusFiles but "cat" is in SkipStirngs
					S(N("fish", CL("file_b", 17), Label("brown"))),                 //skip because "_b:17" is in SkipFiles, even though "fish" is in FocusStrings
					S(N("biscuit", CL("file_b", 20), Pending, Label("brown"))),     //skip because spec is flagged pending
					S(N("pony", CL("c", 3), Label("brown"))),                       //skip because "c" is not in FocusFiles or FocusStrings
					S(N("goat", CL("d", 17), Label("brown"))),                      //skip because "goat" is in FocusStrings but "d" is not in FocusFiles
				}

				conf.FocusFiles = []string{"file_:1,15-21"}
				conf.SkipFiles = []string{"_b:17"}
				conf.FocusStrings = []string{"goat", "dog", "fish", "biscuit"}
				conf.SkipStrings = []string{"cat"}
				conf.LabelFilter = "brown"
			})

			It("applies all filters", func() {
				specs, hasProgrammaticFocus := internal.ApplyFocusToSpecs(specs, description, suiteLabels, suiteSemVerConstraints, conf)
				Ω(harvestSkips(specs)).Should(Equal([]bool{true, false, true, true, true, true, true, true}))
				Ω(hasProgrammaticFocus).Should(BeTrue())
			})
		})
	})
})

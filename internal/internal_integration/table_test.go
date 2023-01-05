package internal_integration_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
)

var _ = Describe("Table driven tests", func() {
	var bodyFunc = func(a, b int) {
		rt.Run(CurrentSpecReport().LeafNodeText)
		if a != b {
			F("fail")
		}
	}

	Describe("constructing tables", func() {
		BeforeEach(func() {
			success, _ := RunFixture("table happy-path", func() {
				DescribeTable("hello", bodyFunc,
					Entry("A", 1, 1),
					Entry("B", 1, 1),
					[]TableEntry{
						Entry("C", 1, 2),
						Entry("D", 1, 1),
					})
			})
			Ω(success).Should(BeFalse())
		})

		It("runs all the entries", func() {
			Ω(rt).Should(HaveTracked("A", "B", "C", "D"))
		})

		It("reports on the tests correctly", func() {
			Ω(reporter.Did.Names()).Should(Equal([]string{"A", "B", "C", "D"}))
			Ω(reporter.Did.Find("C")).Should(HaveFailed("fail", types.NodeTypeIt))
			Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(4), NPassed(3), NFailed(1)))
		})
	})

	Describe("Entry Descriptions", func() {
		Describe("tables with no table-level entry description functions or strings", func() {
			BeforeEach(func() {
				success, _ := RunFixture("table with no table-level entry description function", func() {
					DescribeTable("hello", func(a int, b string, c ...float64) {},
						Entry(nil, 1, "b"),
						Entry(nil, 1, "b", 2.71, 3.141),
						Entry("C", 3, "b", 3.141),
					)
				})
				Ω(success).Should(BeTrue())
			})

			It("renders the parameters for nil-described Entries as It strings", func() {
				Ω(reporter.Did.Names()).Should(Equal([]string{
					"Entry: 1, b",
					"Entry: 1, b, 2.71, 3.141",
					"C",
				}))
			})
		})

		Describe("tables with a table-level entry description function", func() {
			Context("happy path", func() {
				BeforeEach(func() {
					success, _ := RunFixture("table with table-level entry description function", func() {
						DescribeTable("hello",
							func(a int, b string, c ...float64) {},
							func(a int, b string, c ...float64) string {
								return fmt.Sprintf("%d | %s | %v", a, b, c)
							},
							Entry(nil, 1, "b"),
							Entry(nil, 1, "b", 2.71, 3.141),
							Entry("C", 3, "b", 3.141),
						)
					})
					Ω(success).Should(BeTrue())
				})

				It("renders the parameters for nil-described Entries using the provided function as It strings", func() {
					Ω(reporter.Did.Names()).Should(Equal([]string{
						"1 | b | []",
						"1 | b | [2.71 3.141]",
						"C",
					}))
				})
			})

			Context("with more than one entry description function", func() {
				BeforeEach(func() {
					success, _ := RunFixture("table with multiple table-level entry description function", func() {
						DescribeTable("hello",
							func(a int, b string, c ...float64) {},
							func(a int, b string, c ...float64) string {
								return fmt.Sprintf("%d | %s | %v", a, b, c)
							},
							func(a int, b string, c ...float64) string {
								return fmt.Sprintf("%d ~ %s ~ %v", a, b, c)
							},
							Entry(nil, 1, "b"),
							Entry(nil, 1, "b", 2.71, 3.141),
							Entry("C", 3, "b", 3.141),
						)
					})
					Ω(success).Should(BeTrue())
				})

				It("renders the parameters for nil-described Entries using the last provided function as It strings", func() {
					Ω(reporter.Did.Names()).Should(Equal([]string{
						"1 ~ b ~ []",
						"1 ~ b ~ [2.71 3.141]",
						"C",
					}))
				})
			})

			Context("with a parameter mismatch", func() {
				BeforeEach(func() {
					success, _ := RunFixture("table with multiple table-level entry description function", func() {
						DescribeTable("hello",
							func(a int, b string, c ...float64) {},
							func(a int, b string) string {
								return fmt.Sprintf("%d | %s", a, b)
							},
							Entry(nil, 1, "b"),
							Entry(nil, 1, "b", 2.71, 3.141),
							Entry("C", 3, "b", 3.141),
						)
					})
					Ω(success).Should(BeFalse())
				})

				It("fails the entry with a panic", func() {
					Ω(reporter.Did.Find("1 | b")).Should(HavePassed())
					Ω(reporter.Did.Find("")).Should(HavePanicked("Too many parameters passed in to Entry Description function"))
					Ω(reporter.Did.Find("C")).Should(HavePassed())
				})
			})
		})

		Describe("tables with a table-level entry description format strings", func() {
			BeforeEach(func() {
				success, _ := RunFixture("table with table-level entry description format strings", func() {
					DescribeTable("hello",
						func(a int, b string, c float64) {},
						func(a int, b string, c float64) string { return "ignored" },
						EntryDescription("%[2]s | %[1]d | %[3]v"),
						Entry(nil, 1, "a", 1.2),
						Entry(nil, 1, "b", 2.71),
						Entry("C", 3, "b", 3.141),
					)
				})
				Ω(success).Should(BeTrue())
			})

			It("renders the parameters for nil-described Entries using the provided function as It strings", func() {
				Ω(reporter.Did.Names()).Should(Equal([]string{
					"a | 1 | 1.2",
					"b | 1 | 2.71",
					"C",
				}))
			})
		})

		Describe("entries with entry description functions and entry description format strings", func() {
			BeforeEach(func() {
				entryDescriptionBuilder := func(a, b int) string {
					return fmt.Sprintf("%d vs %d", a, b)
				}
				invalidEntryDescriptionBuilder := func(a, b int) {}

				success, _ := RunFixture("table happy-path with custom descriptions", func() {
					DescribeTable("hello",
						bodyFunc,
						EntryDescription("table-level %d, %d"),
						Entry(entryDescriptionBuilder, 1, 1),
						Entry(entryDescriptionBuilder, 2, 2),
						Entry(entryDescriptionBuilder, 1, 2),
						Entry(entryDescriptionBuilder, 3, 3),
						Entry("A", 4, 4),
						Entry(nil, 5, 5),
						Entry(EntryDescription("%dx%d"), 6, 6),
						Entry(invalidEntryDescriptionBuilder, 4, 4),
					)
				})
				Ω(success).Should(BeFalse())
			})

			It("runs all the entries, with the correct names", func() {
				Ω(rt).Should(HaveTracked("1 vs 1", "2 vs 2", "1 vs 2", "3 vs 3", "A", "table-level 5, 5", "6x6"))
			})

			It("catches invalid entry description functions", func() {
				Ω(reporter.Did.Find("")).Should(HavePanicked("Invalid Entry description"))
			})

			It("reports on the tests correctly", func() {
				Ω(reporter.Did.Names()).Should(Equal([]string{"1 vs 1", "2 vs 2", "1 vs 2", "3 vs 3", "A", "table-level 5, 5", "6x6"}))
				Ω(reporter.Did.Find("1 vs 2")).Should(HaveFailed("fail", types.NodeTypeIt))
				Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(8), NPassed(6), NFailed(2)))
			})
		})

	})

	Describe("managing parameters", func() {
		Describe("when table entries are passed incorrect parameters", func() {
			BeforeEach(func() {
				success, _ := RunFixture("table with invalid inputs", func() {
					Describe("container", func() {
						DescribeTable("with variadic parameters", func(a int, b string, c ...float64) { rt.Run(CurrentSpecReport().LeafNodeText) },
							Entry("var-A", 1, "b"),
							Entry("var-B", 1, "b", 3.0, 4.0),
							Entry("var-too-few", 1),
							Entry("var-wrong-type", 1, 2),
							Entry("var-wrong-type-variadic", 1, "b", 3.0, 4, 5.0),
						)
						DescribeTable("without variadic parameters", func(a int, b string) { rt.Run(CurrentSpecReport().LeafNodeText) },
							Entry("nonvar-A", 1, "b"),
							Entry("nonvar-too-few", 1),
							Entry("nonvar-wrong-type", 1, 2),
							Entry("nonvar-too-many", 1, "b", 2),

							Entry(func(a int, b string) string { return "foo" }, 1, 2),
						)
					})
				})
				Ω(success).Should(BeFalse())
			})

			It("runs all the valid entries, but not the invalid entries", func() {
				Ω(rt).Should(HaveTracked("var-A", "var-B", "nonvar-A"))
			})

			It("reports the invalid entries as having panicked", func() {
				Ω(reporter.Did.Find("var-too-few")).Should(HavePanicked("The Table Body function expected 2 parameters but you passed in 1"))
				Ω(reporter.Did.Find("var-wrong-type")).Should(HavePanicked("The Table Body function expected parameter #2 to be of type <string> but you\n  passed in <int>"))
				Ω(reporter.Did.Find("var-wrong-type-variadic")).Should(HavePanicked("The Table Body function expected its variadic parameters to be of type\n  <float64> but you passed in <int>"))

				Ω(reporter.Did.Find("nonvar-too-few")).Should(HavePanicked("The Table Body function expected 2 parameters but you passed in 1"))
				Ω(reporter.Did.Find("nonvar-wrong-type")).Should(HavePanicked("The Table Body function expected parameter #2 to be of type <string> but you\n  passed in <int>"))
				Ω(reporter.Did.Find("nonvar-too-many")).Should(HavePanicked("The Table Body function expected 2 parameters but you passed in 3"))

				Ω(reporter.Did.Find("")).Should(HavePanicked("The Entry Description function expected parameter #2 to be of type <string>"))
			})

			It("reports on the tests correctly", func() {
				Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(10), NPassed(3), NFailed(7)))
			})
		})

		Describe("handling complex types", func() {
			type ComplicatedThings struct {
				Superstructure string
				Substructure   string
			}

			var A, B, C ComplicatedThings

			BeforeEach(func() {
				A = ComplicatedThings{Superstructure: "the sixth sheikh's sixth sheep's sick", Substructure: "emir"}
				B = ComplicatedThings{Superstructure: "the sixth sheikh's sixth sheep's sick", Substructure: "sheep"}
				C = ComplicatedThings{Superstructure: "the sixth sheikh's sixth sheep's sick", Substructure: "si"}
				success, _ := RunFixture("table with complicated inputs`", func() {
					DescribeTable("a more complicated table",
						func(c ComplicatedThings, count int) {
							rt.RunWithData(CurrentSpecReport().LeafNodeText, "thing", c, "count", count)
							Ω(strings.Count(c.Superstructure, c.Substructure)).Should(BeNumerically("==", count))
						},
						Entry("A", A, 0),
						Entry("B", B, 1),
						Entry("C", C, 3),
					)
				})
				Ω(success).Should(BeTrue())
			})

			It("passes the parameters in correctly", func() {
				Ω(rt.DataFor("A")).Should(HaveKeyWithValue("thing", A))
				Ω(rt.DataFor("A")).Should(HaveKeyWithValue("count", 0))
				Ω(rt.DataFor("B")).Should(HaveKeyWithValue("thing", B))
				Ω(rt.DataFor("B")).Should(HaveKeyWithValue("count", 1))
				Ω(rt.DataFor("C")).Should(HaveKeyWithValue("thing", C))
				Ω(rt.DataFor("C")).Should(HaveKeyWithValue("count", 3))
			})
		})

		DescribeTable("it works when nils are passed in", func(a interface{}, b error) {
			Ω(a).Should(BeNil())
			Ω(b).Should(BeNil())
		}, Entry("nils", nil, nil))

		DescribeTable("it supports variadic parameters", func(a int, b string, c ...interface{}) {
			Ω(a).Should(Equal(c[0]))
			Ω(b).Should(Equal(c[1]))
			Ω(c[2]).Should(BeNil())
		}, Entry("variadic arguments", 1, "one", 1, "one", nil))
	})

	Describe("when table entries are marked pending", func() {
		BeforeEach(func() {
			success, _ := RunFixture("table with pending entries", func() {
				DescribeTable("hello", bodyFunc, Entry("A", 1, 1), PEntry("B", 1, 1), Entry("C", 1, 2), Entry("D", 1, 1), Entry("E", Pending, 1, 1))
			})
			Ω(success).Should(BeFalse())
		})

		It("runs all the non-pending entries", func() {
			Ω(rt).Should(HaveTracked("A", "C", "D"))
		})

		It("reports on the tests correctly", func() {
			Ω(reporter.Did.Names()).Should(Equal([]string{"A", "B", "C", "D", "E"}))
			Ω(reporter.Did.Find("B")).Should(BePending())
			Ω(reporter.Did.Find("C")).Should(HaveFailed("fail", types.NodeTypeIt))
			Ω(reporter.Did.Find("E")).Should(BePending())
			Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(5), NPassed(2), NFailed(1), NPending(2)))
		})
	})

	Describe("when table entries are marked focused", func() {
		BeforeEach(func() {
			success, _ := RunFixture("table with focused entries", func() {
				DescribeTable("hello", bodyFunc, Entry("A", 1, 1), Entry("B", 1, 1), FEntry("C", 1, 2), FEntry("D", 1, 1), Entry("E", Focus, 1, 1))
			})
			Ω(success).Should(BeFalse())
		})

		It("runs all the focused entries", func() {
			Ω(rt).Should(HaveTracked("C", "D", "E"))
		})

		It("reports on the tests correctly", func() {
			Ω(reporter.Did.Names()).Should(Equal([]string{"A", "B", "C", "D", "E"}))
			Ω(reporter.Did.Find("A")).Should(HaveBeenSkipped())
			Ω(reporter.Did.Find("B")).Should(HaveBeenSkipped())
			Ω(reporter.Did.Find("C")).Should(HaveFailed("fail", types.NodeTypeIt))
			Ω(reporter.Did.Find("D")).Should(HavePassed(types.NodeTypeIt))
			Ω(reporter.Did.Find("E")).Should(HavePassed(types.NodeTypeIt))
			Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(5), NPassed(2), NFailed(1), NSkipped(2)))
		})
	})

	Describe("when tables are marked pending", func() {
		BeforeEach(func() {
			success, _ := RunFixture("table marked pending", func() {
				Describe("top-level", func() {
					PDescribeTable("hello", bodyFunc, Entry("A", 1, 1), Entry("B", 1, 1))
					DescribeTable("hello", Pending, bodyFunc, Entry("C", 1, 2), Entry("D", 1, 1))
					It("runs", rt.T("runs"))
				})
			})
			Ω(success).Should(BeTrue())
		})

		It("runs all the focused entries", func() {
			Ω(rt).Should(HaveTracked("runs"))
		})

		It("reports on the tests correctly", func() {
			Ω(reporter.Did.Names()).Should(Equal([]string{"A", "B", "C", "D", "runs"}))
			Ω(reporter.Did.Find("A")).Should(BePending())
			Ω(reporter.Did.Find("B")).Should(BePending())
			Ω(reporter.Did.Find("C")).Should(BePending())
			Ω(reporter.Did.Find("D")).Should(BePending())
			Ω(reporter.Did.Find("runs")).Should(HavePassed())

			Ω(reporter.End).Should(BeASuiteSummary(true, NSpecs(5), NPassed(1), NFailed(0), NPending(4)))
		})
	})

	Describe("when tables are marked focused", func() {
		BeforeEach(func() {
			success, _ := RunFixture("table marked focused", func() {
				Describe("top-level", func() {
					FDescribeTable("hello", bodyFunc, Entry("A", 1, 1), Entry("B", 1, 1))
					DescribeTable("hello", Focus, bodyFunc, Entry("C", 1, 2), Entry("D", 1, 1))
					It("does not run", rt.T("does not run"))
				})
			})
			Ω(success).Should(BeFalse())
		})

		It("runs all the focused entries", func() {
			Ω(rt).Should(HaveTracked("A", "B", "C", "D"))
		})

		It("reports on the tests correctly", func() {
			Ω(reporter.Did.Names()).Should(Equal([]string{"A", "B", "C", "D", "does not run"}))
			Ω(reporter.Did.Find("A")).Should(HavePassed())
			Ω(reporter.Did.Find("B")).Should(HavePassed())
			Ω(reporter.Did.Find("C")).Should(HaveFailed("fail"))
			Ω(reporter.Did.Find("D")).Should(HavePassed())
			Ω(reporter.Did.Find("does not run")).Should(HaveBeenSkipped())

			Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(5), NPassed(3), NFailed(1), NSkipped(1)))
		})
	})

	Describe("support for FlakyAttempts decorators", func() {
		BeforeEach(func() {
			success, _ := RunFixture("flaky table", func() {
				var counter int
				var currentSpec string

				BeforeEach(func() {
					if currentSpec != CurrentSpecReport().LeafNodeText {
						counter = 0
						currentSpec = CurrentSpecReport().LeafNodeText
					}
				})

				DescribeTable("contrived flaky table", FlakeAttempts(2),
					func(failUntil int) {
						rt.Run(CurrentSpecReport().LeafNodeText)
						counter += 1
						if counter < failUntil {
							F("fail")
						}
					},
					Entry("A", 1),
					Entry("B", 2),
					Entry("C", 3),
					Entry("D", []interface{}{FlakeAttempts(3), Offset(2)}, 3),
				)
			})
			Ω(success).Should(BeFalse())
		})

		It("honors the flake attempts decorator", func() {
			Ω(rt).Should(HaveTracked("A", "B", "B", "C", "C", "D", "D", "D"))
		})

		It("reports on the specs appropriately", func() {
			Ω(reporter.Did.Find("A")).Should(HavePassed(NumAttempts(1)))
			Ω(reporter.Did.Find("B")).Should(HavePassed(NumAttempts(2)))
			Ω(reporter.Did.Find("C")).Should(HaveFailed(NumAttempts(2)))
			Ω(reporter.Did.Find("D")).Should(HavePassed(NumAttempts(3)))
		})
	})

	Describe("support for MustPassRepeatedly decorators", func() {
		BeforeEach(func() {
			success, _ := RunFixture("Repeated Passes table", func() {
				var counter int
				var currentSpec string

				BeforeEach(func() {
					if currentSpec != CurrentSpecReport().LeafNodeText {
						counter = 0
						currentSpec = CurrentSpecReport().LeafNodeText
					}
				})

				DescribeTable("contrived Repeated Passes table", MustPassRepeatedly(2),
					func(passUntil int) {
						rt.Run(CurrentSpecReport().LeafNodeText)
						counter += 1
						if counter >= passUntil {
							F("fail")
						}
					},
					Entry("A", 1),
					Entry("B", 2),
					Entry("C", 3),
					Entry("D", []interface{}{MustPassRepeatedly(3), Offset(2)}, 3),
				)
			})
			Ω(success).Should(BeFalse())
		})

		It("honors the Repeated attempts decorator", func() {
			Ω(rt).Should(HaveTracked("A", "B", "B", "C", "C", "D", "D", "D"))
		})

		It("reports on the specs appropriately", func() {
			Ω(reporter.Did.Find("A")).Should(HaveFailed(NumAttempts(1)))
			Ω(reporter.Did.Find("B")).Should(HaveFailed(NumAttempts(2)))
			Ω(reporter.Did.Find("C")).Should(HavePassed(NumAttempts(2)))
			Ω(reporter.Did.Find("D")).Should(HaveFailed(NumAttempts(3)))
		})
	})
})

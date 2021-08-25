package internal_integration_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/types"
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
				DescribeTable("hello", bodyFunc, Entry("A", 1, 1), Entry("B", 1, 1), Entry("C", 1, 2), Entry("D", 1, 1))
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

	Describe("constructing tables with dynamic entry description functions", func() {
		BeforeEach(func() {
			entryDescriptionBuilder := func(a, b int) string {
				return fmt.Sprintf("%d vs %d", a, b)
			}

			success, _ := RunFixture("table happy-path with custom descriptions", func() {
				DescribeTable("hello", bodyFunc, Entry(entryDescriptionBuilder, 1, 1), Entry(entryDescriptionBuilder, 2, 2), Entry(entryDescriptionBuilder, 1, 2), Entry(entryDescriptionBuilder, 3, 3))
			})
			Ω(success).Should(BeFalse())
		})

		It("runs all the entries, with the correct names", func() {
			Ω(rt).Should(HaveTracked("1 vs 1", "2 vs 2", "1 vs 2", "3 vs 3"))
		})

		It("reports on the tests correctly", func() {
			Ω(reporter.Did.Names()).Should(Equal([]string{"1 vs 1", "2 vs 2", "1 vs 2", "3 vs 3"}))
			Ω(reporter.Did.Find("1 vs 2")).Should(HaveFailed("fail", types.NodeTypeIt))
			Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(4), NPassed(3), NFailed(1)))
		})
	})

	Describe("managing parameters", func() {
		Describe("when table entries are passed incorrect parameters", func() {
			BeforeEach(func() {
				success, _ := RunFixture("table with invalid inputs", func() {
					DescribeTable("hello", bodyFunc, Entry("A", 1, 1), Entry("B", 1, 1), Entry("C", 1, 2), Entry("D", 1, "aardvark"), Entry("E", 1, 1))
				})
				Ω(success).Should(BeFalse())
			})

			It("runs all the valid entries", func() {
				Ω(rt).Should(HaveTracked("A", "B", "C", "E"))
			})

			It("reports on the tests correctly", func() {
				Ω(reporter.Did.Names()).Should(Equal([]string{"A", "B", "C", "D", "E"}))
				Ω(reporter.Did.Find("C")).Should(HaveFailed("fail", types.NodeTypeIt))
				Ω(reporter.Did.Find("D")).Should(HavePanicked("reflect: Call using string as type int", types.NodeTypeIt))
				Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(5), NPassed(3), NFailed(2)))
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
	})

	Describe("when table entries are marked pending", func() {
		BeforeEach(func() {
			success, _ := RunFixture("table with pending entries", func() {
				DescribeTable("hello", bodyFunc, Entry("A", 1, 1), PEntry("B", 1, 1), Entry("C", 1, 2), Entry("D", 1, 1))
			})
			Ω(success).Should(BeFalse())
		})

		It("runs all the non-pending entries", func() {
			Ω(rt).Should(HaveTracked("A", "C", "D"))
		})

		It("reports on the tests correctly", func() {
			Ω(reporter.Did.Names()).Should(Equal([]string{"A", "B", "C", "D"}))
			Ω(reporter.Did.Find("B")).Should(BePending())
			Ω(reporter.Did.Find("C")).Should(HaveFailed("fail", types.NodeTypeIt))
			Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(4), NPassed(2), NFailed(1), NPending(1)))
		})
	})

	Describe("when table entries are marked focused", func() {
		BeforeEach(func() {
			success, _ := RunFixture("table with focused entries", func() {
				DescribeTable("hello", bodyFunc, Entry("A", 1, 1), Entry("B", 1, 1), FEntry("C", 1, 2), FEntry("D", 1, 1))
			})
			Ω(success).Should(BeFalse())
		})

		It("runs all the focused entries", func() {
			Ω(rt).Should(HaveTracked("C", "D"))
		})

		It("reports on the tests correctly", func() {
			Ω(reporter.Did.Names()).Should(Equal([]string{"A", "B", "C", "D"}))
			Ω(reporter.Did.Find("A")).Should(HaveBeenSkipped())
			Ω(reporter.Did.Find("B")).Should(HaveBeenSkipped())
			Ω(reporter.Did.Find("C")).Should(HaveFailed("fail", types.NodeTypeIt))
			Ω(reporter.Did.Find("D")).Should(HavePassed(types.NodeTypeIt))
			Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(4), NPassed(1), NFailed(1), NSkipped(2)))
		})
	})

	Describe("when tables are marked pending", func() {
		BeforeEach(func() {
			success, _ := RunFixture("table marked pending", func() {
				Describe("top-level", func() {
					PDescribeTable("hello", bodyFunc, Entry("A", 1, 1), Entry("B", 1, 1), Entry("C", 1, 2), Entry("D", 1, 1))
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
					FDescribeTable("hello", bodyFunc, Entry("A", 1, 1), Entry("B", 1, 1), Entry("C", 1, 2), Entry("D", 1, 1))
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
})

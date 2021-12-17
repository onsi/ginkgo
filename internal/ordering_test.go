package internal_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/internal"
	"github.com/onsi/ginkgo/v2/types"
)

type SpecTexts []string

func getTexts(specs Specs, groupedSpecIndices internal.GroupedSpecIndices) SpecTexts {
	out := []string{}
	for _, specIndices := range groupedSpecIndices {
		for _, idx := range specIndices {
			out = append(out, specs[idx].Text())
		}
	}
	return out
}

func (tt SpecTexts) Join() string {
	return strings.Join(tt, "")
}

var _ = Describe("OrderSpecs", func() {
	var conf types.SuiteConfig
	var specs Specs

	BeforeEach(func() {
		conf = types.SuiteConfig{}
		conf.RandomSeed = 1
		conf.ParallelTotal = 1

		con1 := N(ntCon)
		con2 := N(ntCon)
		specs = Specs{
			S(N("A", ntIt)),
			S(N("B", ntIt)),
			S(con1, N("C", ntIt)),
			S(con1, N("D", ntIt)),
			S(con1, N(ntCon), N("E", ntIt)),
			S(N("F", ntIt)),
			S(con2, N("G", ntIt)),
			S(con2, N("H", ntIt)),
		}
	})

	Context("when configured to only randomize top-level specs", func() {
		It("shuffles top level specs only", func() {
			for conf.RandomSeed = 1; conf.RandomSeed < 10; conf.RandomSeed += 1 {
				groupedSpecIndices, serialSpecIndices := internal.OrderSpecs(specs, conf)
				Ω(serialSpecIndices).Should(BeEmpty())

				Ω(getTexts(specs, groupedSpecIndices).Join()).Should(ContainSubstring("CDE"))
				Ω(getTexts(specs, groupedSpecIndices).Join()).Should(ContainSubstring("GH"))
			}

			conf.RandomSeed = 1
			groupedSpecIndices1, _ := internal.OrderSpecs(specs, conf)
			conf.RandomSeed = 2
			groupedSpecIndices2, _ := internal.OrderSpecs(specs, conf)
			Ω(getTexts(specs, groupedSpecIndices1)).ShouldNot(Equal(getTexts(specs, groupedSpecIndices2)))
		})
	})

	Context("when configured to randomize all specs", func() {
		BeforeEach(func() {
			conf.RandomizeAllSpecs = true
		})

		It("shuffles all specs", func() {
			hasCDE := true
			hasGH := true
			for conf.RandomSeed = 1; conf.RandomSeed < 10; conf.RandomSeed += 1 {
				groupedSpecIndices, serialSpecIndices := internal.OrderSpecs(specs, conf)
				Ω(serialSpecIndices).Should(BeEmpty())

				hasCDE, _ = ContainSubstring("CDE").Match(getTexts(specs, groupedSpecIndices).Join())
				hasGH, _ = ContainSubstring("GH").Match(getTexts(specs, groupedSpecIndices).Join())
				if !hasCDE && !hasGH {
					break
				}
			}

			Ω(hasCDE || hasGH).Should(BeFalse(), "after 10 randomizations, we really shouldn't have gotten CDE and GH in order as all specs should be shuffled, not just top-level containers and specs")

			conf.RandomSeed = 1
			groupedSpecIndices1, _ := internal.OrderSpecs(specs, conf)
			conf.RandomSeed = 2
			groupedSpecIndices2, _ := internal.OrderSpecs(specs, conf)
			Ω(getTexts(specs, groupedSpecIndices1)).ShouldNot(Equal(getTexts(specs, groupedSpecIndices2)))
		})
	})

	Context("when passed the same seed", func() {
		It("always generates the same order", func() {
			for _, conf.RandomizeAllSpecs = range []bool{true, false} {
				for conf.RandomSeed = 1; conf.RandomSeed < 10; conf.RandomSeed += 1 {
					groupedSpecIndices, serialSpecIndices := internal.OrderSpecs(specs, conf)
					Ω(serialSpecIndices).Should(BeEmpty())
					for i := 0; i < 10; i++ {
						reshuffledGroupedSpecIndices, serialSpecIndices := internal.OrderSpecs(specs, conf)
						Ω(serialSpecIndices).Should(BeEmpty())

						Ω(getTexts(specs, groupedSpecIndices)).Should(Equal(getTexts(specs, reshuffledGroupedSpecIndices)))
					}
				}
			}
		})
	})

	Context("when specs are in different files and the files are loaded in an undefined order", func() {
		var specsInFileA, specsInFileB Specs
		BeforeEach(func() {
			con1 := N(ntCon, CL("file_A", 10))
			specsInFileA = Specs{
				S(N("A", ntIt, CL("file_A", 1))),
				S(N("B", ntIt, CL("file_A", 5))),
				S(con1, N("C", ntIt, CL("file_A", 15))),
				S(con1, N("D", ntIt, CL("file_A", 20))),
				S(con1, N(ntCon, CL("file_A", 25)), N("E", ntIt, CL("file_A", 30))),
			}

			con2 := N(ntCon, CL("file_B", 10))
			specsInFileB = Specs{
				S(N("F", ntIt, CL("file_B", 1))),
				S(con2, N("G", ntIt, CL("file_B", 15))),
				S(con2, N("H", ntIt, CL("file_B", 20))),
			}

		})

		It("always generates a consistent randomization when given the same seed", func() {
			for _, conf.RandomizeAllSpecs = range []bool{true, false} {
				for conf.RandomSeed = 1; conf.RandomSeed < 10; conf.RandomSeed += 1 {
					specsOrderAB := Specs{}
					specsOrderAB = append(specsOrderAB, specsInFileA...)
					specsOrderAB = append(specsOrderAB, specsInFileB...)

					specsOrderBA := Specs{}
					specsOrderBA = append(specsOrderBA, specsInFileB...)
					specsOrderBA = append(specsOrderBA, specsInFileA...)

					groupedSpecIndicesAB, serialSpecIndices := internal.OrderSpecs(specsOrderAB, conf)
					Ω(serialSpecIndices).Should(BeEmpty())

					groupedSpecIndicesBA, serialSpecIndices := internal.OrderSpecs(specsOrderBA, conf)
					Ω(serialSpecIndices).Should(BeEmpty())

					Ω(getTexts(specsOrderAB, groupedSpecIndicesAB)).Should(Equal(getTexts(specsOrderBA, groupedSpecIndicesBA)))
				}
			}
		})
	})

	Context("when there are ordered specs and randomize-all is true", func() {
		BeforeEach(func() {
			con1 := N(ntCon, Ordered)
			con2 := N(ntCon)
			specs = Specs{
				S(N("A", ntIt)),
				S(N("B", ntIt)),
				S(con1, N("C", ntIt)),
				S(con1, N("D", ntIt)),
				S(con1, N(ntCon), N("E", ntIt)),
				S(N("F", ntIt)),
				S(con2, N("G", ntIt)),
				S(con2, N("H", ntIt)),
			}

			conf.RandomizeAllSpecs = true
		})

		It("never shuffles the specs in ordered specs", func() {
			for conf.RandomSeed = 1; conf.RandomSeed < 10; conf.RandomSeed += 1 {
				groupedSpecIndices, serialSpecIndices := internal.OrderSpecs(specs, conf)
				Ω(serialSpecIndices).Should(BeEmpty())

				Ω(getTexts(specs, groupedSpecIndices).Join()).Should(ContainSubstring("CDE"))
			}
		})
	})

	Context("when there are ordered specs and randomize-all is false and everything is in an enclosing container", func() {
		BeforeEach(func() {
			con0 := N(ntCon)
			con1 := N(ntCon, Ordered)
			con2 := N(ntCon)
			specs = Specs{
				S(con0, N("A", ntIt)),
				S(con0, N("B", ntIt)),
				S(con0, con1, N("C", ntIt)),
				S(con0, con1, N("D", ntIt)),
				S(con0, con1, N(ntCon), N("E", ntIt)),
				S(con0, N("F", ntIt)),
				S(con0, con2, N("G", ntIt)),
				S(con0, con2, N("H", ntIt)),
			}

			conf.RandomizeAllSpecs = false
		})

		It("runs all the specs in order", func() {
			for conf.RandomSeed = 1; conf.RandomSeed < 10; conf.RandomSeed += 1 {
				groupedSpecIndices, serialSpecIndices := internal.OrderSpecs(specs, conf)
				Ω(serialSpecIndices).Should(BeEmpty())

				Ω(getTexts(specs, groupedSpecIndices).Join()).Should(Equal("ABCDEFGH"))
			}
		})
	})

	Context("when there are serial specs", func() {
		BeforeEach(func() {
			con1 := N(ntCon, Ordered, Serial)
			con2 := N(ntCon)
			specs = Specs{
				S(N("A", Serial, ntIt)),
				S(N("B", ntIt)),
				S(con1, N("C", ntIt)),
				S(con1, N("D", ntIt)),
				S(con1, N(ntCon), N("E", ntIt)),
				S(N("F", ntIt)),
				S(con2, N("G", ntIt)),
				S(con2, N("H", ntIt, Serial)),
			}
			conf.RandomizeAllSpecs = true
		})

		Context("and the tests are not running in parallel", func() {
			BeforeEach(func() {
				conf.ParallelTotal = 1
			})

			It("puts all the tests in the parallelizable group and returns an empty serial group", func() {
				for conf.RandomSeed = 1; conf.RandomSeed < 10; conf.RandomSeed += 1 {
					groupedSpecIndices, serialSpecIndices := internal.OrderSpecs(specs, conf)
					Ω(serialSpecIndices).Should(BeEmpty())

					Ω(getTexts(specs, groupedSpecIndices).Join()).Should(ContainSubstring("CDE"))
					Ω(getTexts(specs, groupedSpecIndices)).Should(ConsistOf("A", "B", "C", "D", "E", "F", "G", "H"))
				}

				conf.RandomSeed = 1
				groupedSpecIndices1, _ := internal.OrderSpecs(specs, conf)
				conf.RandomSeed = 2
				groupedSpecIndices2, _ := internal.OrderSpecs(specs, conf)
				Ω(getTexts(specs, groupedSpecIndices1)).ShouldNot(Equal(getTexts(specs, groupedSpecIndices2)))
			})
		})

		Context("and the tests are running in parallel", func() {
			BeforeEach(func() {
				conf.ParallelTotal = 2
			})

			It("puts all parallelizable tests in the parallelizable group and all serial tests in the serial group, preserving ordered test order", func() {
				for conf.RandomSeed = 1; conf.RandomSeed < 10; conf.RandomSeed += 1 {
					groupedSpecIndices, serialSpecIndices := internal.OrderSpecs(specs, conf)

					Ω(getTexts(specs, groupedSpecIndices)).Should(ConsistOf("B", "F", "G"))
					Ω(getTexts(specs, serialSpecIndices).Join()).Should(ContainSubstring("CDE"))
					Ω(getTexts(specs, serialSpecIndices)).Should(ConsistOf("A", "C", "D", "E", "H"))
				}

				conf.RandomSeed = 1
				groupedSpecIndices1, serialSpecIndices1 := internal.OrderSpecs(specs, conf)
				conf.RandomSeed = 2
				groupedSpecIndices2, serialSpecIndices2 := internal.OrderSpecs(specs, conf)
				Ω(getTexts(specs, groupedSpecIndices1)).ShouldNot(Equal(getTexts(specs, groupedSpecIndices2)))
				Ω(getTexts(specs, serialSpecIndices1)).ShouldNot(Equal(getTexts(specs, serialSpecIndices2)))
			})
		})
	})
})

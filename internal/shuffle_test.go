package internal_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/internal"
	"github.com/onsi/ginkgo/types"
)

type SpecTexts []string

func getTexts(specs Specs) SpecTexts {
	out := []string{}
	for _, spec := range specs {
		out = append(out, spec.Text())
	}
	return out
}

func (tt SpecTexts) Join() string {
	return strings.Join(tt, "")
}

var _ = Describe("Shuffle", func() {
	var conf types.SuiteConfig
	var specs Specs

	BeforeEach(func() {
		conf = types.SuiteConfig{}
		conf.RandomSeed = 1

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
				shuffledSpecs := internal.ShuffleSpecs(specs, conf)
				Ω(getTexts(shuffledSpecs).Join()).Should(ContainSubstring("CDE"))
				Ω(getTexts(shuffledSpecs).Join()).Should(ContainSubstring("GH"))
			}

			conf.RandomSeed = 1
			specsSeed1 := internal.ShuffleSpecs(specs, conf)
			conf.RandomSeed = 2
			specsSeed2 := internal.ShuffleSpecs(specs, conf)
			Ω(getTexts(specsSeed1)).ShouldNot(Equal(getTexts(specsSeed2)))
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
				shuffledSpecs := internal.ShuffleSpecs(specs, conf)
				hasCDE, _ = ContainSubstring("CDE").Match(getTexts(shuffledSpecs).Join())
				hasGH, _ = ContainSubstring("GH").Match(getTexts(shuffledSpecs).Join())
				if !hasCDE && !hasGH {
					break
				}
			}

			Ω(hasCDE || hasGH).Should(BeFalse(), "after 10 randomizations, we really shouldn't have gotten CDE and GH in order as all specs should be shuffled, not just top-level containers and specs")

			conf.RandomSeed = 1
			specsSeed1 := internal.ShuffleSpecs(specs, conf)
			conf.RandomSeed = 2
			specsSeed2 := internal.ShuffleSpecs(specs, conf)
			Ω(getTexts(specsSeed1)).ShouldNot(Equal(getTexts(specsSeed2)))
		})
	})

	Context("when passed the same seed", func() {
		It("always generates the same order", func() {
			for _, conf.RandomizeAllSpecs = range []bool{true, false} {
				for conf.RandomSeed = 1; conf.RandomSeed < 10; conf.RandomSeed += 1 {
					shuffledSpecs := internal.ShuffleSpecs(specs, conf)
					for i := 0; i < 10; i++ {
						reshuffledSpecs := internal.ShuffleSpecs(specs, conf)
						Ω(shuffledSpecs).Should(Equal(reshuffledSpecs))
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

					shuffledSpecsAB := internal.ShuffleSpecs(specsOrderAB, conf)
					shuffledSpecsBA := internal.ShuffleSpecs(specsOrderBA, conf)

					Ω(getTexts(shuffledSpecsAB)).Should(Equal(getTexts(shuffledSpecsBA)))
				}
			}
		})
	})
})

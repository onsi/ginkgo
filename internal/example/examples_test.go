package example_test

import (
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/internal/example"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/internal/codelocation"
	"github.com/onsi/ginkgo/internal/containernode"
	"github.com/onsi/ginkgo/internal/leafnodes"
	"github.com/onsi/ginkgo/internal/types"
)

var _ = Describe("Examples", func() {
	var examples *Examples

	newExample := func(text string, flag internaltypes.FlagType) *Example {
		subject := leafnodes.NewItNode(text, func() {}, flag, codelocation.New(0), 0, nil, 0)
		return New(subject, []*containernode.ContainerNode{})
	}

	newMeasureExample := func(text string, flag internaltypes.FlagType) *Example {
		subject := leafnodes.NewMeasureNode(text, func(Benchmarker) {}, flag, codelocation.New(0), 0, nil, 0)
		return New(subject, []*containernode.ContainerNode{})
	}

	newExamples := func(args ...interface{}) *Examples {
		examples := []*Example{}
		for index := 0; index < len(args)-1; index += 2 {
			examples = append(examples, newExample(args[index].(string), args[index+1].(internaltypes.FlagType)))
		}
		return NewExamples(examples)
	}

	exampleTexts := func(examples *Examples) []string {
		texts := []string{}
		for _, example := range examples.Examples() {
			texts = append(texts, example.ConcatenatedString())
		}
		return texts
	}

	willRunTexts := func(examples *Examples) []string {
		texts := []string{}
		for _, example := range examples.Examples() {
			if !(example.Skipped() || example.Pending()) {
				texts = append(texts, example.ConcatenatedString())
			}
		}
		return texts
	}

	skippedTexts := func(examples *Examples) []string {
		texts := []string{}
		for _, example := range examples.Examples() {
			if example.Skipped() {
				texts = append(texts, example.ConcatenatedString())
			}
		}
		return texts
	}

	pendingTexts := func(examples *Examples) []string {
		texts := []string{}
		for _, example := range examples.Examples() {
			if example.Pending() {
				texts = append(texts, example.ConcatenatedString())
			}
		}
		return texts
	}

	Describe("Shuffling specs", func() {
		It("should shuffle the specs using the passed in randomizer", func() {
			examples17 := newExamples("C", noneFlag, "A", noneFlag, "B", noneFlag)
			examples17.Shuffle(rand.New(rand.NewSource(17)))
			texts17 := exampleTexts(examples17)

			examples17Again := newExamples("C", noneFlag, "A", noneFlag, "B", noneFlag)
			examples17Again.Shuffle(rand.New(rand.NewSource(17)))
			texts17Again := exampleTexts(examples17Again)

			examples15 := newExamples("C", noneFlag, "A", noneFlag, "B", noneFlag)
			examples15.Shuffle(rand.New(rand.NewSource(15)))
			texts15 := exampleTexts(examples15)

			examplesUnshuffled := newExamples("C", noneFlag, "A", noneFlag, "B", noneFlag)
			textsUnshuffled := exampleTexts(examplesUnshuffled)

			Ω(textsUnshuffled).Should(Equal([]string{"C", "A", "B"}))

			Ω(texts17).Should(Equal(texts17Again))
			Ω(texts17).ShouldNot(Equal(texts15))
			Ω(texts17).ShouldNot(Equal(textsUnshuffled))
			Ω(texts15).ShouldNot(Equal(textsUnshuffled))

			Ω(texts17).Should(HaveLen(3))
			Ω(texts17).Should(ContainElement("A"))
			Ω(texts17).Should(ContainElement("B"))
			Ω(texts17).Should(ContainElement("C"))

			Ω(texts15).Should(HaveLen(3))
			Ω(texts15).Should(ContainElement("A"))
			Ω(texts15).Should(ContainElement("B"))
			Ω(texts15).Should(ContainElement("C"))
		})
	})

	Describe("Applying focus/skip", func() {
		var description, focusString, skipString string

		BeforeEach(func() {
			description, focusString, skipString = "", "", ""
		})

		JustBeforeEach(func() {
			examples = newExamples("A1", focusedFlag, "A2", noneFlag, "B1", focusedFlag, "B2", pendingFlag)
			examples.ApplyFocus(description, focusString, skipString)
		})

		Context("with neither a focus string nor a skip string", func() {
			It("should apply the programmatic focus", func() {
				Ω(willRunTexts(examples)).Should(Equal([]string{"A1", "B1"}))
				Ω(skippedTexts(examples)).Should(Equal([]string{"A2", "B2"}))
				Ω(pendingTexts(examples)).Should(BeEmpty())
			})
		})

		Context("with a focus regexp", func() {
			BeforeEach(func() {
				focusString = "A"
			})

			It("should override the programmatic focus", func() {
				Ω(willRunTexts(examples)).Should(Equal([]string{"A1", "A2"}))
				Ω(skippedTexts(examples)).Should(Equal([]string{"B1", "B2"}))
				Ω(pendingTexts(examples)).Should(BeEmpty())
			})
		})

		Context("with a focus regexp", func() {
			BeforeEach(func() {
				focusString = "B"
			})

			It("should not override any pendings", func() {
				Ω(willRunTexts(examples)).Should(Equal([]string{"B1"}))
				Ω(skippedTexts(examples)).Should(Equal([]string{"A1", "A2"}))
				Ω(pendingTexts(examples)).Should(Equal([]string{"B2"}))
			})
		})

		Context("with a description", func() {
			BeforeEach(func() {
				description = "C"
				focusString = "C"
			})

			It("should include the description in the focus determination", func() {
				Ω(willRunTexts(examples)).Should(Equal([]string{"A1", "A2", "B1"}))
				Ω(skippedTexts(examples)).Should(BeEmpty())
				Ω(pendingTexts(examples)).Should(Equal([]string{"B2"}))
			})
		})

		Context("with a description", func() {
			BeforeEach(func() {
				description = "C"
				skipString = "C"
			})

			It("should include the description in the focus determination", func() {
				Ω(willRunTexts(examples)).Should(BeEmpty())
				Ω(skippedTexts(examples)).Should(Equal([]string{"A1", "A2", "B1", "B2"}))
				Ω(pendingTexts(examples)).Should(BeEmpty())
			})
		})

		Context("with a skip regexp", func() {
			BeforeEach(func() {
				skipString = "A"
			})

			It("should override the programmatic focus", func() {
				Ω(willRunTexts(examples)).Should(Equal([]string{"B1"}))
				Ω(skippedTexts(examples)).Should(Equal([]string{"A1", "A2"}))
				Ω(pendingTexts(examples)).Should(Equal([]string{"B2"}))
			})
		})

		Context("with both a focus and a skip regexp", func() {
			BeforeEach(func() {
				focusString = "1"
				skipString = "B"
			})

			It("should AND the two", func() {
				Ω(willRunTexts(examples)).Should(Equal([]string{"A1"}))
				Ω(skippedTexts(examples)).Should(Equal([]string{"A2", "B1", "B2"}))
				Ω(pendingTexts(examples)).Should(BeEmpty())
			})
		})
	})

	Describe("skipping measurements", func() {
		BeforeEach(func() {
			examples = NewExamples([]*Example{
				newExample("A", noneFlag),
				newExample("B", noneFlag),
				newExample("C", pendingFlag),
				newMeasureExample("measurementA", noneFlag),
				newMeasureExample("measurementB", pendingFlag),
			})
		})

		It("should skip measurements", func() {
			Ω(willRunTexts(examples)).Should(Equal([]string{"A", "B", "measurementA"}))
			Ω(skippedTexts(examples)).Should(BeEmpty())
			Ω(pendingTexts(examples)).Should(Equal([]string{"C", "measurementB"}))

			examples.SkipMeasurements()

			Ω(willRunTexts(examples)).Should(Equal([]string{"A", "B"}))
			Ω(skippedTexts(examples)).Should(Equal([]string{"measurementA", "measurementB"}))
			Ω(pendingTexts(examples)).Should(Equal([]string{"C"}))
		})
	})

	Describe("when running tests in parallel", func() {
		It("should select out a subset of the tests", func() {
			examplesNode1 := newExamples("A", noneFlag, "B", noneFlag, "C", noneFlag, "D", noneFlag, "E", noneFlag)
			examplesNode2 := newExamples("A", noneFlag, "B", noneFlag, "C", noneFlag, "D", noneFlag, "E", noneFlag)
			examplesNode3 := newExamples("A", noneFlag, "B", noneFlag, "C", noneFlag, "D", noneFlag, "E", noneFlag)

			examplesNode1.TrimForParallelization(3, 1)
			examplesNode2.TrimForParallelization(3, 2)
			examplesNode3.TrimForParallelization(3, 3)

			Ω(willRunTexts(examplesNode1)).Should(Equal([]string{"A", "B"}))
			Ω(willRunTexts(examplesNode2)).Should(Equal([]string{"C", "D"}))
			Ω(willRunTexts(examplesNode3)).Should(Equal([]string{"E"}))

			Ω(examplesNode1.Examples()).Should(HaveLen(2))
			Ω(examplesNode2.Examples()).Should(HaveLen(2))
			Ω(examplesNode3.Examples()).Should(HaveLen(1))

			Ω(examplesNode1.NumberOfOriginalExamples()).Should(Equal(5))
			Ω(examplesNode2.NumberOfOriginalExamples()).Should(Equal(5))
			Ω(examplesNode3.NumberOfOriginalExamples()).Should(Equal(5))
		})

		Context("when way too many nodes are used", func() {
			It("should return 0 examples", func() {
				examplesNode1 := newExamples("A", noneFlag, "B", noneFlag)
				examplesNode2 := newExamples("A", noneFlag, "B", noneFlag)
				examplesNode3 := newExamples("A", noneFlag, "B", noneFlag)

				examplesNode1.TrimForParallelization(3, 1)
				examplesNode2.TrimForParallelization(3, 2)
				examplesNode3.TrimForParallelization(3, 3)

				Ω(willRunTexts(examplesNode1)).Should(Equal([]string{"A"}))
				Ω(willRunTexts(examplesNode2)).Should(Equal([]string{"B"}))
				Ω(willRunTexts(examplesNode3)).Should(BeEmpty())

				Ω(examplesNode1.Examples()).Should(HaveLen(1))
				Ω(examplesNode2.Examples()).Should(HaveLen(1))
				Ω(examplesNode3.Examples()).Should(HaveLen(0))

				Ω(examplesNode1.NumberOfOriginalExamples()).Should(Equal(2))
				Ω(examplesNode2.NumberOfOriginalExamples()).Should(Equal(2))
				Ω(examplesNode3.NumberOfOriginalExamples()).Should(Equal(2))
			})
		})

	})
})

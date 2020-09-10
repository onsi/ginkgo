package types_test

import (
	"reflect"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/formatter"
	"github.com/onsi/ginkgo/types"
)

var _ = Describe("Deprecation Support", func() {
	Describe("Tracking Deprecations", func() {
		var tracker *types.DeprecationTracker

		BeforeEach(func() {
			tracker = types.NewDeprecationTracker()
			formatter.SingletonFormatter.ColorMode = formatter.ColorModePassthrough
		})

		AfterEach(func() {
			formatter.SingletonFormatter.ColorMode = formatter.ColorModeTerminal
		})

		Context("with no tracked deprecations", func() {
			It("reports no tracked deprecations", func() {
				Ω(tracker.DidTrackDeprecations()).Should(BeFalse())
			})
		})

		Context("with tracked dependencies", func() {
			BeforeEach(func() {
				tracker.TrackDeprecation(types.Deprecation{
					Message: "Deprecation 1",
					DocLink: "doclink-1",
				}, types.CodeLocation{FileName: "foo.go", LineNumber: 17})
				tracker.TrackDeprecation(types.Deprecation{
					Message: "Deprecation 1",
					DocLink: "doclink-1",
				}, types.CodeLocation{FileName: "bar.go", LineNumber: 30})
				tracker.TrackDeprecation(types.Deprecation{
					Message: "Deprecation 2",
					DocLink: "doclink-2",
				})
				tracker.TrackDeprecation(types.Deprecation{
					Message: "Deprecation 3",
				}, types.CodeLocation{FileName: "baz.go", LineNumber: 72})
			})

			It("reports tracked deprecations", func() {
				Ω(tracker.DidTrackDeprecations()).Should(BeTrue())
			})

			It("generates a nicely formatted report", func() {
				report := tracker.DeprecationsReport()
				Ω(report).Should(HavePrefix("{{light-yellow}}You're using deprecated Ginkgo functionality:{{/}}\n{{light-yellow}}============================================={{/}}\n"))
				Ω(report).Should(ContainSubstring(strings.Join([]string{
					"  {{yellow}}Deprecation 1{{/}}",
					"  {{bold}}Learn more at:{{/}} {{cyan}}{{underline}}https://github.com/onsi/ginkgo/blob/v2/docs/MIGRATING_TO_V2.md#doclink-1{{/}}",
					"    {{gray}}foo.go:17{{/}}",
					"    {{gray}}bar.go:30{{/}}",
					"",
				}, "\n")))
				Ω(report).Should(ContainSubstring(strings.Join([]string{
					"  {{yellow}}Deprecation 2{{/}}",
					"  {{bold}}Learn more at:{{/}} {{cyan}}{{underline}}https://github.com/onsi/ginkgo/blob/v2/docs/MIGRATING_TO_V2.md#doclink-2{{/}}",
					"",
				}, "\n")))
				Ω(report).Should(ContainSubstring(strings.Join([]string{
					"  {{yellow}}Deprecation 3{{/}}",
					"    {{gray}}baz.go:72{{/}}",
				}, "\n")))
			})

			It("validates that all deprecations point to working documentation", func() {
				v := reflect.ValueOf(types.Deprecations)
				Ω(v.NumMethod()).Should(BeNumerically(">", 0))
				for i := 0; i < v.NumMethod(); i += 1 {
					m := v.Method(i)
					deprecation := m.Call([]reflect.Value{})[0].Interface().(types.Deprecation)

					if deprecation.DocLink != "" {
						Ω(deprecation.DocLink).Should(BeElementOf(DEPRECATION_ANCHORS))
					}
				}
			})
		})
	})

	Describe("DeprecatedSetupSummaryFromSummary", func() {
		It("converts to the v1 summary format", func() {
			cl1 := types.CodeLocation{FileName: "foo.go", LineNumber: 3}
			cl2 := types.CodeLocation{FileName: "bar.go", LineNumber: 5}
			Ω(types.DeprecatedSetupSummaryFromSummary(types.Summary{
				LeafNodeType:               types.NodeTypeBeforeSuite,
				LeafNodeLocation:           cl1,
				State:                      types.SpecStateFailed,
				RunTime:                    time.Hour,
				CapturedGinkgoWriterOutput: "ginkgo-writer-output",
				CapturedStdOutErr:          "std-output",
				Failure: types.Failure{
					Message:        "failure message",
					Location:       cl2,
					ForwardedPanic: "forwarded panic",
					NodeIndex:      2,
					NodeType:       types.NodeTypeBeforeSuite,
				},
			})).Should(Equal(
				&types.SetupSummary{
					ComponentType:  types.SpecComponentTypeBeforeSuite,
					CodeLocation:   cl1,
					State:          types.SpecStateFailed,
					RunTime:        time.Hour,
					CapturedOutput: "std-output\nginkgo-writer-output",
					Failure: types.SpecFailure{
						Message:               "failure message",
						Location:              cl2,
						ForwardedPanic:        "forwarded panic",
						ComponentIndex:        2,
						ComponentType:         types.SpecComponentTypeBeforeSuite,
						ComponentCodeLocation: cl2,
					},
				},
			))
		})
	})

	Describe("DeprecatedSpecSummaryFromSummary", func() {
		It("converts to the v1 summary format", func() {
			cl1 := types.CodeLocation{FileName: "foo.go", LineNumber: 3}
			cl2 := types.CodeLocation{FileName: "bar.go", LineNumber: 5}
			Ω(types.DeprecatedSpecSummaryFromSummary(types.Summary{
				NodeTexts:                  []string{"A", "B"},
				NodeLocations:              []types.CodeLocation{cl1, cl2},
				LeafNodeType:               types.NodeTypeBeforeSuite,
				LeafNodeLocation:           cl1,
				State:                      types.SpecStateFailed,
				RunTime:                    time.Hour,
				CapturedGinkgoWriterOutput: "ginkgo-writer-output",
				CapturedStdOutErr:          "std-output",
				Failure: types.Failure{
					Message:        "failure message",
					Location:       cl2,
					ForwardedPanic: "forwarded panic",
					NodeIndex:      2,
					NodeType:       types.NodeTypeBeforeSuite,
				},
			})).Should(Equal(
				&types.SpecSummary{
					ComponentTexts:         []string{"A", "B"},
					ComponentCodeLocations: []types.CodeLocation{cl1, cl2},
					State:                  types.SpecStateFailed,
					RunTime:                time.Hour,
					CapturedOutput:         "std-output\nginkgo-writer-output",
					IsMeasurement:          false,
					Measurements:           map[string]*types.SpecMeasurement{},
					Failure: types.SpecFailure{
						Message:               "failure message",
						Location:              cl2,
						ForwardedPanic:        "forwarded panic",
						ComponentIndex:        2,
						ComponentType:         types.SpecComponentTypeBeforeSuite,
						ComponentCodeLocation: cl2,
					},
				},
			))
		})
	})
})

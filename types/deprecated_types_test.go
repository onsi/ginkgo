package types_test

import (
	. "github.com/onsi/ginkgo/v2"
	// . "github.com/onsi/gomega"
)

var _ = Describe("DeprecatedTypes", func() {
	// Describe("DeprecatedSetupSummaryFromSpecReport", func() {
	// 	It("converts to the v1 summary format", func() {
	// 		cl1 := types.CodeLocation{FileName: "foo.go", LineNumber: 3}
	// 		cl2 := types.CodeLocation{FileName: "bar.go", LineNumber: 5}
	// 		Ω(types.DeprecatedSetupSummaryFromSpecReport(types.SpecReport{
	// 			LeafNodeType:               types.NodeTypeBeforeSuite,
	// 			LeafNodeLocation:           cl1,
	// 			State:                      types.SpecStateFailed,
	// 			RunTime:                    time.Hour,
	// 			CapturedGinkgoWriterOutput: "ginkgo-writer-output",
	// 			CapturedStdOutErr:          "std-output",
	// 			Failure: types.Failure{
	// 				Message:        "failure message",
	// 				Location:       cl2,
	// 				ForwardedPanic: "forwarded panic",
	// 				NodeIndex:      2,
	// 				NodeType:       types.NodeTypeBeforeSuite,
	// 			},
	// 		})).Should(Equal(
	// 			&types.SetupSummary{
	// 				ComponentType:  types.SpecComponentTypeBeforeSuite,
	// 				CodeLocation:   cl1,
	// 				State:          types.SpecStateFailed,
	// 				RunTime:        time.Hour,
	// 				CapturedOutput: "std-output\nginkgo-writer-output",
	// 				Failure: types.SpecFailure{
	// 					Message:               "failure message",
	// 					Location:              cl2,
	// 					ForwardedPanic:        "forwarded panic",
	// 					ComponentIndex:        2,
	// 					ComponentType:         types.SpecComponentTypeBeforeSuite,
	// 					ComponentCodeLocation: cl2,
	// 				},
	// 			},
	// 		))
	// 	})
	// })

	// Describe("DeprecatedSpecSummaryFromSpecReport", func() {
	// 	It("converts to the v1 summary format", func() {
	// 		cl1 := types.CodeLocation{FileName: "foo.go", LineNumber: 3}
	// 		cl2 := types.CodeLocation{FileName: "bar.go", LineNumber: 5}
	// 		Ω(types.DeprecatedSpecSummaryFromSpecReport(types.SpecReport{
	// 			NodeTexts:                  []string{"A", "B"},
	// 			NodeLocations:              []types.CodeLocation{cl1, cl2},
	// 			LeafNodeType:               types.NodeTypeBeforeSuite,
	// 			LeafNodeLocation:           cl1,
	// 			State:                      types.SpecStateFailed,
	// 			RunTime:                    time.Hour,
	// 			CapturedGinkgoWriterOutput: "ginkgo-writer-output",
	// 			CapturedStdOutErr:          "std-output",
	// 			Failure: types.Failure{
	// 				Message:        "failure message",
	// 				Location:       cl2,
	// 				ForwardedPanic: "forwarded panic",
	// 				NodeIndex:      2,
	// 				NodeType:       types.NodeTypeBeforeSuite,
	// 			},
	// 		})).Should(Equal(
	// 			&types.SpecSummary{
	// 				ComponentTexts:         []string{"A", "B"},
	// 				ComponentCodeLocations: []types.CodeLocation{cl1, cl2},
	// 				State:                  types.SpecStateFailed,
	// 				RunTime:                time.Hour,
	// 				CapturedOutput:         "std-output\nginkgo-writer-output",
	// 				IsMeasurement:          false,
	// 				Measurements:           map[string]*types.SpecMeasurement{},
	// 				Failure: types.SpecFailure{
	// 					Message:               "failure message",
	// 					Location:              cl2,
	// 					ForwardedPanic:        "forwarded panic",
	// 					ComponentIndex:        2,
	// 					ComponentType:         types.SpecComponentTypeBeforeSuite,
	// 					ComponentCodeLocation: cl2,
	// 				},
	// 			},
	// 		))
	// 	})
	// })
})

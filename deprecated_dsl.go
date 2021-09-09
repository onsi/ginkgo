package ginkgo

import (
	"time"

	"github.com/onsi/ginkgo/internal"
	"github.com/onsi/ginkgo/internal/global"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
)

/*
===================================================
   Deprecations for v2
===================================================
*/

// Deprecated Done Channel for asynchronous testing
type Done = internal.Done

//Deprecated: Custom Ginkgo test reporters are no longer supported
//Please read the documentation at:
//https://github.com/onsi/ginkgo/blob/ver2/docs/MIGRATING_TO_V2.md#removed-custom-reporters
//for Ginkgo's new behavior and for a migration path.
type Reporter = reporters.DeprecatedReporter

//Deprecated: Custom Reporters have been removed in v2.  RunSpecsWithDefaultAndCustomReporters will simply call RunSpecs()
//
//Please read the documentation at:
//https://github.com/onsi/ginkgo/blob/ver2/docs/MIGRATING_TO_V2.md#removed-custom-reporters
//for Ginkgo's new behavior and for a migration path.
func RunSpecsWithDefaultAndCustomReporters(t GinkgoTestingT, description string, _ []Reporter) bool {
	deprecationTracker.TrackDeprecation(types.Deprecations.CustomReporter())
	return RunSpecs(t, description)
}

//Deprecated: Custom Reporters have been removed in v2.  RunSpecsWithCustomReporters will simply call RunSpecs()
//
//Please read the documentation at:
//https://github.com/onsi/ginkgo/blob/ver2/docs/MIGRATING_TO_V2.md#removed-custom-reporters
//for Ginkgo's new behavior and for a migration path.
func RunSpecsWithCustomReporters(t GinkgoTestingT, description string, _ []Reporter) bool {
	deprecationTracker.TrackDeprecation(types.Deprecations.CustomReporter())
	return RunSpecs(t, description)
}

//GinkgoTestDescription represents the information about the current running test returned by CurrentGinkgoTestDescription
//	FullTestText: a concatenation of ComponentTexts and the TestText
//	ComponentTexts: a list of all texts for the Describes & Contexts leading up to the current test
//	TestText: the text in the It node
//	FileName: the name of the file containing the current test
//	LineNumber: the line number for the current test
//	Failed: if the current test has failed, this will be true (useful in an AfterEach)
//
//Deprecated: Use CurrentSpecReport() instead
type DeprecatedGinkgoTestDescription struct {
	FullTestText   string
	ComponentTexts []string
	TestText       string

	FileName   string
	LineNumber int

	Failed   bool
	Duration time.Duration
}
type GinkgoTestDescription = DeprecatedGinkgoTestDescription

//CurrentGinkgoTestDescripton returns information about the current running test.
//Deprecated: Use CurrentSpecReport() instead
func CurrentGinkgoTestDescription() DeprecatedGinkgoTestDescription {
	deprecationTracker.TrackDeprecation(
		types.Deprecations.CurrentGinkgoTestDescription(),
		types.NewCodeLocation(1),
	)
	report := global.Suite.CurrentSpecReport()
	if report.State == types.SpecStateInvalid {
		return GinkgoTestDescription{}
	}
	componentTexts := []string{}
	componentTexts = append(componentTexts, report.ContainerHierarchyTexts...)
	componentTexts = append(componentTexts, report.LeafNodeText)

	return DeprecatedGinkgoTestDescription{
		ComponentTexts: componentTexts,
		FullTestText:   report.FullText(),
		TestText:       report.LeafNodeText,
		FileName:       report.LeafNodeLocation.FileName,
		LineNumber:     report.LeafNodeLocation.LineNumber,
		Failed:         report.State.Is(types.SpecStateFailureStates...),
		Duration:       report.RunTime,
	}
}

//deprecated benchmarker
type Benchmarker interface {
	Time(name string, body func(), info ...interface{}) (elapsedTime time.Duration)
	RecordValue(name string, value float64, info ...interface{})
	RecordValueWithPrecision(name string, value float64, units string, precision int, info ...interface{})
}

//deprecated Measure
func Measure(_ ...interface{}) bool {
	deprecationTracker.TrackDeprecation(types.Deprecations.Measure(), types.NewCodeLocation(1))
	return true
}

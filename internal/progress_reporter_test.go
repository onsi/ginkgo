package internal_test

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var leafNodeLocation = types.NewCodeLocation(0)
var nodeLocation = types.NewCodeLocation(0)
var progressStepLocation = types.NewCodeLocation(0)

func SR(leafNodeText string, containerHierarchyTexts ...string) types.SpecReport {
	return types.SpecReport{
		LeafNodeText:            leafNodeText,
		ContainerHierarchyTexts: containerHierarchyTexts,
		LeafNodeLocation:        leafNodeLocation,
	}
}

func PS(name string) internal.ProgressStepCursor {
	if name == "" {
		return internal.ProgressStepCursor{}
	}
	return internal.ProgressStepCursor{
		Name:         name,
		CodeLocation: progressStepLocation,
	}
}

var _ = Describe("ProgressReporter", func() {
	DescribeTable("the report header", func(report types.SpecReport, currentNode Node, currentStep internal.ProgressStepCursor, headerLines ...string) {
		report.StartTime = time.Now().Add(-5 * time.Second)
		currentStep.StartTime = time.Now().Add(-1 * time.Second)
		pr, err := internal.NewProgressReport(report, currentNode, time.Now().Add(-3*time.Second), currentStep)
		Ω(err).ShouldNot(HaveOccurred())
		reportLines := strings.Split(pr.Report("{{orange}}", false), "\n")
		failureMessageLines := []string{}
		for _, line := range reportLines {
			failureMessageLines = append(failureMessageLines, fmt.Sprintf("`%s`", line))
		}
		failureMessage := strings.Join(failureMessageLines, ",\n")

		for i, line := range headerLines {
			Ω(reportLines[i]).Should(MatchRegexp(line), failureMessage)
		}
	},
		Entry("With a suite node",
			SR(""), N(types.NodeTypeBeforeSuite, nodeLocation), PS(""),
			`In {{bold}}{{orange}}\[BeforeSuite\]{{/}} \(Node Runtime: 3[\.\d]*s\)`,
			`  {{gray}}`+nodeLocation.String()+`{{/}}`,
			``,
			`{{bold}}{{underline}}Spec Goroutine{{/}}`,
		),
		Entry("With a top-level spec",
			SR("A Top-Level It"), N(types.NodeTypeIt, "A Top-Level It", leafNodeLocation), PS(""),
			`{{bold}}{{orange}}A Top-Level It{{/}} \(Spec Runtime: 5[\.\d]*s\)`,
			`  {{gray}}`+leafNodeLocation.String()+`{{/}}`,
			`  In {{bold}}{{orange}}\[It\]{{/}} \(Node Runtime: 3[\.\d]*s\)`,
			`    {{gray}}`+leafNodeLocation.String()+`{{/}}`,
			``,
			`{{bold}}{{underline}}Spec Goroutine{{/}}`),
		Entry("With a spec in containers",
			SR("My Spec", "Container A", "Container B", "Container C"), N(types.NodeTypeIt, "My Spec", leafNodeLocation), PS(""),
			`{{/}}Container A {{gray}}Container B {{/}}Container C{{/}} {{bold}}{{orange}}My Spec{{/}} \(Spec Runtime: 5[\.\d]*s\)`,
			`  {{gray}}`+leafNodeLocation.String()+`{{/}}`,
			`  In {{bold}}{{orange}}\[It\]{{/}} \(Node Runtime: 3[\.\d]*s\)`,
			`    {{gray}}`+leafNodeLocation.String()+`{{/}}`,
			``,
			`{{bold}}{{underline}}Spec Goroutine{{/}}`,
		),
		Entry("With no currentNode",
			SR("My Spec", "Container A", "Container B", "Container C"), Node{}, PS(""),
			`{{/}}Container A {{gray}}Container B {{/}}Container C{{/}} {{bold}}{{orange}}My Spec{{/}} \(Spec Runtime: 5[\.\d]*s\)`,
			`  {{gray}}`+leafNodeLocation.String()+`{{/}}`,
			``,
			`{{bold}}{{underline}}Spec Goroutine{{/}}`,
		),
		Entry("With a currentNode that is not an It",
			SR("My Spec", "Container A", "Container B", "Container C"), N(types.NodeTypeBeforeEach, nodeLocation), PS(""),
			`{{/}}Container A {{gray}}Container B {{/}}Container C{{/}} {{bold}}{{orange}}My Spec{{/}} \(Spec Runtime: 5[\.\d]*s\)`,
			`  {{gray}}`+leafNodeLocation.String()+`{{/}}`,
			`  In {{bold}}{{orange}}\[BeforeEach\]{{/}} \(Node Runtime: 3[\.\d]*s\)`,
			`    {{gray}}`+nodeLocation.String()+`{{/}}`,
			``,
			`{{bold}}{{underline}}Spec Goroutine{{/}}`,
		),
		Entry("With a currentNode that is not an It but has text",
			SR(""), N(types.NodeTypeReportAfterSuite, "My Report", nodeLocation, func(Report) {}), PS(""),
			`In {{bold}}{{orange}}\[ReportAfterSuite\]{{/}} {{bold}}{{orange}}My Report{{/}} \(Node Runtime: 3[\.\d]*s\)`,
			`  {{gray}}`+nodeLocation.String()+`{{/}}`,
			``,
			`{{bold}}{{underline}}Spec Goroutine{{/}}`,
		),
		Entry("With a current step",
			SR("My Spec", "Container A", "Container B", "Container C"), N(types.NodeTypeBeforeEach, nodeLocation), PS("Reticulating Splines"),
			`{{/}}Container A {{gray}}Container B {{/}}Container C{{/}} {{bold}}{{orange}}My Spec{{/}} \(Spec Runtime: 5[\.\d]*s\)`,
			`  {{gray}}`+leafNodeLocation.String()+`{{/}}`,
			`  In {{bold}}{{orange}}\[BeforeEach\]{{/}} \(Node Runtime: 3[\.\d]*s\)`,
			`    {{gray}}`+nodeLocation.String()+`{{/}}`,
			`    At {{bold}}{{orange}}\[By Step\] Reticulating Splines{{/}} \(Step Runtime: 1[\.\d]*s\)`,
			`      {{gray}}`+progressStepLocation.String()+`{{/}}`,
			``,
			`{{bold}}{{underline}}Spec Goroutine{{/}}`,
		),
	)

	Describe("The goroutine stack", func() {
		It("is better tested in the internal integration tests because this test package lives in internal which is a key part of the logic for how the goroutine stack is analyzed...", func() {
			//empty
		})

		var pr internal.ProgressReport

		BeforeEach(func() {
			var err error
			pr, err = internal.NewProgressReport(SR(""), Node{}, time.Now(), PS(""))
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("when includeAllGoroutines is false", func() {
			It("does not include any other goroutines", func() {
				report := pr.Report("{{orange}}", false)
				Ω(report).ShouldNot(ContainSubstring("Other Goroutines"))
				Ω(report).ShouldNot(ContainSubstring("main.main()"))
			})
		})

		Context("when includeAllGoroutines is true", func() {
			It("includes all other goroutines", func() {
				report := pr.Report("{{orange}}", true)
				Ω(report).Should(ContainSubstring("Other Goroutines"))
				Ω(report).Should(ContainSubstring("main.main()"))
			})
		})
	})
})

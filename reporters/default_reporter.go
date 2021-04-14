/*
Ginkgo's Default Reporter

A number of command line flags are available to tweak Ginkgo's default output.

These are documented [here](http://onsi.github.io/ginkgo/#running_tests)
*/
package reporters

import (
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/onsi/ginkgo/formatter"
	"github.com/onsi/ginkgo/types"
)

type DefaultReporter struct {
	conf   types.ReporterConfig
	writer io.Writer

	// managing the emission stream
	lastChar                 string
	lastEmissionWasDelimiter bool

	// rendering
	specDenoter  string
	retryDenoter string
	formatter    formatter.Formatter
}

func NewDefaultReporterUnderTest(conf types.ReporterConfig, writer io.Writer) *DefaultReporter {
	reporter := NewDefaultReporter(conf, writer)
	reporter.formatter = formatter.New(formatter.ColorModePassthrough)

	return reporter
}

func NewDefaultReporter(conf types.ReporterConfig, writer io.Writer) *DefaultReporter {
	reporter := &DefaultReporter{
		conf:   conf,
		writer: writer,

		lastChar:                 "\n",
		lastEmissionWasDelimiter: false,

		specDenoter:  "•",
		retryDenoter: "↺",
		formatter:    formatter.NewWithNoColorBool(conf.NoColor),
	}
	if runtime.GOOS == "windows" {
		reporter.specDenoter = "+"
		reporter.retryDenoter = "R"
	}

	return reporter
}

/* The Reporter Interface */

func (r *DefaultReporter) SpecSuiteWillBegin(report types.Report) {
	if r.conf.Succinct {
		r.emit(r.f("[%d] {{bold}}%s{{/}} ", report.SuiteConfig.RandomSeed, report.SuiteDescription))
		r.emit(r.f("- %d/%d specs ", report.PreRunStats.SpecsThatWillRun, report.PreRunStats.TotalSpecs))
		if report.SuiteConfig.ParallelTotal > 1 {
			r.emit(r.f("- %d nodes ", report.SuiteConfig.ParallelTotal))
		}
	} else {
		banner := r.f("Running Suite: %s - %s", report.SuiteDescription, report.SuitePath)
		r.emitBlock(banner)
		r.emitBlock(strings.Repeat("=", len(banner)))

		out := r.f("Random Seed: {{bold}}%d{{/}}", report.SuiteConfig.RandomSeed)
		if report.SuiteConfig.RandomizeAllSpecs {
			out += r.f(" - will randomize all specs")
		}
		r.emitBlock(out)
		r.emit("\n")
		r.emitBlock(r.f("Will run {{bold}}%d{{/}} of {{bold}}%d{{/}} specs", report.PreRunStats.SpecsThatWillRun, report.PreRunStats.TotalSpecs))
		if report.SuiteConfig.ParallelTotal > 1 {
			r.emitBlock(r.f("Running in parallel across {{bold}}%d{{/}} nodes", report.SuiteConfig.ParallelTotal))
		}
	}
}

func (r *DefaultReporter) WillRun(report types.SpecReport) {
	if !r.conf.Verbose || report.State.Is(types.SpecStatePending, types.SpecStateSkipped) {
		return
	}

	r.emitDelimiter()
	if report.LeafNodeType.Is(types.NodeTypesForSuiteSetup...) {
		r.emitBlock(r.f("{{bold}}[%s]{{/}}", report.LeafNodeType.String()))
		r.emitBlock(r.f("{{gray}}%s{{/}}", report.LeafNodeLocation))
	} else {
		lastIndex := len(report.NodeTexts) - 1
		indentation := uint(0)
		if lastIndex > 0 {
			r.emitBlock(r.cycleJoin(report.NodeTexts[0:lastIndex], " "))
			indentation = 1
		}
		if lastIndex >= 0 {
			r.emitBlock(r.fi(indentation, "{{bold}}%s{{/}}", report.NodeTexts[lastIndex]))
			r.emitBlock(r.fi(indentation, "{{gray}}%s{{/}}", report.NodeLocations[lastIndex]))
		}
	}
}

func (r *DefaultReporter) DidRun(report types.SpecReport) {
	var header, highlightColor string
	includeRuntime, emitGinkgoWriterOutput, stream, denoter := true, true, false, r.specDenoter
	succinctLocationBlock := r.conf.Succinct

	hasGW := report.CapturedGinkgoWriterOutput != ""
	hasStd := report.CapturedStdOutErr != ""

	if report.LeafNodeType.Is(types.NodeTypesForSuiteSetup...) {
		denoter = fmt.Sprintf("[%s]", report.LeafNodeType)
	}

	switch report.State {
	case types.SpecStatePassed:
		highlightColor, succinctLocationBlock = "{{green}}", !r.conf.Verbose
		emitGinkgoWriterOutput = (r.conf.ReportPassed || r.conf.Verbose) && hasGW
		if report.LeafNodeType.Is(types.NodeTypesForSuiteSetup...) {
			if r.conf.Verbose || hasStd {
				header = fmt.Sprintf("%s PASSED", denoter)
			} else {
				return
			}
		} else {
			header, stream = denoter, true
			if report.NumAttempts > 1 {
				header, stream = fmt.Sprintf("%s [FLAKEY TEST - TOOK %d ATTEMPTS TO PASS]", r.retryDenoter, report.NumAttempts), false
			}
			if report.RunTime.Seconds() > r.conf.SlowSpecThreshold {
				header, stream = fmt.Sprintf("%s [SLOW TEST]", header), false
			}
		}
		if hasStd || emitGinkgoWriterOutput {
			stream = false
		}
	case types.SpecStatePending:
		highlightColor = "{{yellow}}"
		includeRuntime, emitGinkgoWriterOutput = false, false
		if r.conf.Succinct {
			header, stream = "P", true
		} else {
			header, succinctLocationBlock = "P [PENDING]", !r.conf.Verbose
		}
	case types.SpecStateSkipped:
		highlightColor = "{{cyan}}"
		if r.conf.Succinct || report.Failure.Message == "" {
			header, stream = "S", true
		} else {
			header, succinctLocationBlock = "S [SKIPPED]", !r.conf.Verbose
		}
	case types.SpecStateFailed:
		highlightColor, header = "{{red}}", fmt.Sprintf("%s [FAILED]", denoter)
	case types.SpecStatePanicked:
		highlightColor, header = "{{magenta}}", fmt.Sprintf("%s! [PANICKED]", denoter)
	case types.SpecStateInterrupted:
		highlightColor, header = "{{orange}}", fmt.Sprintf("%s! [INTERRUPTED]", denoter)
	}

	// Emit stream and return
	if stream {
		r.emit(r.f(highlightColor + header + "{{/}}"))
		return
	}

	// Emit header
	r.emitDelimiter()
	if includeRuntime {
		header = r.f("%s [%.3f seconds]", header, report.RunTime.Seconds())
	}
	r.emitBlock(r.f(highlightColor + header + "{{/}}"))

	// Emit Code Location Block
	r.emitBlock(r.codeLocationBlock(report, highlightColor, succinctLocationBlock))

	//Emit Stdout/Stderr Output
	if hasStd {
		r.emitBlock("\n")
		r.emitBlock(r.fi(1, "{{gray}}Begin Captured StdOut/StdErr Output >>{{/}}"))
		r.emitBlock(r.fi(2, "%s", report.CapturedStdOutErr))
		r.emitBlock(r.fi(1, "{{gray}}<< End Captured StdOut/StdErr Output{{/}}"))
	}

	//Emit Captured GinkgoWriter Output
	if emitGinkgoWriterOutput && hasGW {
		r.emitBlock("\n")
		r.emitBlock(r.fi(1, "{{gray}}Begin Captured GinkgoWriter Output >>{{/}}"))
		r.emitBlock(r.fi(2, "%s", report.CapturedGinkgoWriterOutput))
		r.emitBlock(r.fi(1, "{{gray}}<< End Captured GinkgoWriter Output{{/}}"))
	}

	// Emit Failure Message
	if !report.Failure.IsZero() {
		r.emitBlock("\n")
		r.emitBlock(r.fi(1, highlightColor+"%s{{/}}", report.Failure.Message))
		r.emitBlock(r.fi(1, highlightColor+"In {{bold}}[%s]{{/}}"+highlightColor+" at: {{bold}}%s{{/}}\n", report.Failure.NodeType, report.Failure.Location))
		if report.Failure.ForwardedPanic != "" {
			r.emitBlock("\n")
			r.emitBlock(r.fi(1, highlightColor+"%s{{/}}", report.Failure.ForwardedPanic))
		}

		if r.conf.FullTrace || report.Failure.ForwardedPanic != "" {
			r.emitBlock("\n")
			r.emitBlock(r.fi(1, highlightColor+"Full Stack Trace{{/}}"))
			r.emitBlock(r.fi(2, "%s", report.Failure.Location.FullStackTrace))
		}
	}

	r.emitDelimiter()
}

func (r *DefaultReporter) SpecSuiteDidEnd(report types.Report) {
	failures := report.SpecReports.WithState(types.SpecStateFailureStates...)
	if len(failures) > 1 {
		r.emitBlock("\n\n")
		r.emitBlock(r.f("{{red}}{{bold}}Summarizing %d Failures:{{/}}", len(failures)))
		for _, specReport := range failures {
			highlightColor, heading := "{{red}}", "[FAIL]"
			if specReport.State.Is(types.SpecStateInterrupted) {
				highlightColor, heading = "{{orange}}", "[INTERRUPTED]"
			} else if specReport.State.Is(types.SpecStatePanicked) {
				highlightColor, heading = "{{magenta}}", "[PANICKED!]"
			}

			locationBlock := r.codeLocationBlock(specReport, highlightColor, true)
			r.emitBlock(r.fi(1, highlightColor+"%s{{/}} %s", heading, locationBlock))
		}
	}

	//summarize the suite
	if r.conf.Succinct && report.SuiteSucceeded {
		r.emit(r.f(" {{green}}SUCCESS!{{/}} %s ", report.RunTime))
		return
	}

	r.emitBlock("\n")
	color, status := "{{green}}{{bold}}", "SUCCESS!"
	if !report.SuiteSucceeded {
		color, status = "{{red}}{{bold}}", "FAIL!"
		if report.SuiteConfig.FailOnPending && len(failures) == 0 && report.SpecReports.CountWithState(types.SpecStatePending) > 0 {
			color, status = "{{yellow}}{{bold}}", "FAIL! - Detected pending specs and --fail-on-pending is set"
		}
	}

	specs := report.SpecReports.WithLeafNodeType(types.NodeTypeIt) //exclude any suite setup nodes
	r.emitBlock(r.f(color+"Ran %d of %d Specs in %.3f seconds{{/}}",
		specs.CountWithState(types.SpecStatePassed)+specs.CountWithState(types.SpecStateFailureStates...),
		report.PreRunStats.TotalSpecs,
		report.RunTime.Seconds()),
	)
	r.emit(r.f(color+"%s{{/}} -- ", status))

	if len(specs) == 0 && report.SpecReports.WithLeafNodeType(types.NodeTypeBeforeSuite, types.NodeTypeSynchronizedBeforeSuite).CountWithState(types.SpecStateFailureStates...) > 0 {
		r.emit(r.f("{{cyan}}{{bold}}A BeforeSuite node failed so all tests were skipped.{{/}}\n"))
	} else {
		r.emit(r.f("{{green}}{{bold}}%d Passed{{/}} | ", specs.CountWithState(types.SpecStatePassed)))
		r.emit(r.f("{{red}}{{bold}}%d Failed{{/}} | ", specs.CountWithState(types.SpecStateFailureStates...)))
		if specs.CountOfFlakedSpecs() > 0 {
			r.emit(r.f("{{light-yellow}}{{bold}}%d Flaked{{/}} | ", specs.CountOfFlakedSpecs()))
		}
		r.emit(r.f("{{yellow}}{{bold}}%d Pending{{/}} | ", specs.CountWithState(types.SpecStatePending)))
		r.emit(r.f("{{cyan}}{{bold}}%d Skipped{{/}}\n", specs.CountWithState(types.SpecStateSkipped)))
	}
}

/* Emitting to the writer */
func (r *DefaultReporter) emit(s string) {
	if len(s) > 0 {
		r.lastChar = s[len(s)-1:]
		r.lastEmissionWasDelimiter = false
		r.writer.Write([]byte(s))
	}
}

func (r *DefaultReporter) emitBlock(s string) {
	if len(s) > 0 {
		if r.lastChar != "\n" {
			r.emit("\n")
		}
		r.emit(s)
		if r.lastChar != "\n" {
			r.emit("\n")
		}
	}
}

func (r *DefaultReporter) emitDelimiter() {
	if r.lastEmissionWasDelimiter {
		return
	}
	r.emitBlock(r.f("{{gray}}%s{{/}}", strings.Repeat("-", 30)))
	r.lastEmissionWasDelimiter = true
}

/* Rendering text */
func (r *DefaultReporter) f(format string, args ...interface{}) string {
	return r.formatter.F(format, args...)
}

func (r *DefaultReporter) fi(indentation uint, format string, args ...interface{}) string {
	return r.formatter.Fi(indentation, format, args...)
}

func (r *DefaultReporter) cycleJoin(elements []string, joiner string) string {
	return r.formatter.CycleJoin(elements, joiner, []string{"{{/}}", "{{gray}}"})
}

func (r *DefaultReporter) codeLocationBlock(report types.SpecReport, highlightColor string, succinct bool) string {
	out := ""

	if report.LeafNodeType.Is(types.NodeTypesForSuiteSetup...) {
		out = r.f(highlightColor+"{{bold}}[%s]{{/}}\n", report.LeafNodeType)
		if report.Failure.IsZero() {
			out += r.f("{{gray}}%s{{/}}\n", report.LeafNodeLocation)
		} else {
			out += r.f("{{gray}}%s{{/}}\n", report.Failure.Location)
		}
		return out
	}

	if succinct {
		texts := make([]string, len(report.NodeTexts))
		copy(texts, report.NodeTexts)
		var codeLocation = report.NodeLocations[len(report.NodeLocations)-1]
		if !report.Failure.IsZero() {
			codeLocation = report.Failure.Location
			if report.Failure.NodeIndex == -1 {
				texts = append([]string{r.f(highlightColor+"{{bold}}[%s]{{/}}", report.Failure.NodeType)}, texts...)
			} else if report.Failure.NodeIndex < len(texts) {
				i := report.Failure.NodeIndex
				texts[i] = r.f(highlightColor+"{{bold}}[%s] %s{{/}}", report.Failure.NodeType, texts[i])
			}
		}
		out += r.f("%s\n", r.cycleJoin(texts, " "))
		out += r.f("{{gray}}%s{{/}}", codeLocation)

		return out
	}

	indentation := uint(0)
	if !report.Failure.IsZero() && report.Failure.NodeIndex == -1 {
		out += r.fi(indentation, highlightColor+"{{bold}}TOP-LEVEL [%s]{{/}}\n", report.Failure.NodeType)
		out += r.fi(indentation, "{{gray}}%s{{/}}\n", report.Failure.Location)
		indentation += 1
	}

	for i := range report.NodeTexts {
		if !report.Failure.IsZero() && i == report.Failure.NodeIndex {
			out += r.fi(indentation, highlightColor+"{{bold}}%s [%s]{{/}}\n", report.NodeTexts[i], report.Failure.NodeType)
		} else {
			out += r.fi(indentation, "%s\n", report.NodeTexts[i])
		}
		out += r.fi(indentation, "{{gray}}%s{{/}}\n", report.NodeLocations[i])
		indentation += 1
	}

	return out
}

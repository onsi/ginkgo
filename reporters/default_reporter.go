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

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/formatter"
	"github.com/onsi/ginkgo/types"
)

type DefaultReporter struct {
	conf             config.DefaultReporterConfigType
	hasFailOnPending bool
	writer           io.Writer

	failures []types.Summary

	// managing the emission stream
	lastChar                 string
	lastEmissionWasDelimiter bool

	// rendering
	specDenoter  string
	retryDenoter string
	formatter    formatter.Formatter
}

func NewDefaultReporterUnderTest(conf config.DefaultReporterConfigType, writer io.Writer) *DefaultReporter {
	reporter := NewDefaultReporter(conf, writer)
	reporter.formatter = formatter.New(formatter.ColorModePassthrough)

	return reporter
}

func NewDefaultReporter(conf config.DefaultReporterConfigType, writer io.Writer) *DefaultReporter {
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

func (r *DefaultReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary types.SuiteSummary) {
	r.hasFailOnPending = config.FailOnPending
	if r.conf.Succinct {
		r.emit(r.f("[%d] {{bold}}%s{{/}} ", config.RandomSeed, summary.SuiteDescription))
		r.emit(r.f("- %d/%d specs ", summary.NumberOfSpecsThatWillBeRun, summary.NumberOfTotalSpecs))
		if config.ParallelTotal > 1 {
			r.emit(r.f("- %d nodes ", config.ParallelTotal))
		}
	} else {
		banner := r.f("Running Suite: %s", summary.SuiteDescription)
		r.emitBlock(banner)
		r.emitBlock(strings.Repeat("=", len(banner)))

		out := r.f("Random Seed: {{bold}}%d{{/}}", config.RandomSeed)
		if config.RandomizeAllSpecs {
			out += r.f(" - will randomize all specs")
		}
		r.emitBlock(out)
		r.emit("\n")
		r.emitBlock(r.f("Will run {{bold}}%d{{/}} of {{bold}}%d{{/}} specs", summary.NumberOfSpecsThatWillBeRun, summary.NumberOfTotalSpecs))
		if config.ParallelTotal > 1 {
			r.emitBlock(r.f("Running in parallel across {{bold}}%d{{/}} nodes", config.ParallelTotal))
		}
	}
}

func (r *DefaultReporter) WillRun(summary types.Summary) {
	if !r.conf.Verbose || summary.State.Is(types.SpecStatePending, types.SpecStateSkipped) {
		return
	}

	r.emitDelimiter()
	if summary.LeafNodeType.Is(types.NodeTypesForSuiteSetup...) {
		r.emitBlock(r.f("{{bold}}[%s]{{/}}", summary.LeafNodeType.String()))
		r.emitBlock(r.f("{{gray}}%s{{/}}", summary.LeafNodeLocation))
	} else {
		lastIndex := len(summary.NodeTexts) - 1
		indentation := uint(0)
		if lastIndex > 0 {
			r.emitBlock(r.cycleJoin(summary.NodeTexts[0:lastIndex], " "))
			indentation = 1
		}
		if lastIndex >= 0 {
			r.emitBlock(r.fi(indentation, "{{bold}}%s{{/}}", summary.NodeTexts[lastIndex]))
			r.emitBlock(r.fi(indentation, "{{gray}}%s{{/}}", summary.NodeLocations[lastIndex]))
		}
	}
}

func (r *DefaultReporter) DidRun(summary types.Summary) {
	var header, highlightColor string
	includeRuntime, emitGinkgoWriterOutput, stream, denoter := true, true, false, r.specDenoter
	succinctLocationBlock := r.conf.Succinct

	hasGW := summary.CapturedGinkgoWriterOutput != ""
	hasStd := summary.CapturedStdOutErr != ""

	if summary.LeafNodeType.Is(types.NodeTypesForSuiteSetup...) {
		denoter = fmt.Sprintf("[%s]", summary.LeafNodeType)
	}

	switch summary.State {
	case types.SpecStatePassed:
		highlightColor, succinctLocationBlock = "{{green}}", !r.conf.Verbose
		emitGinkgoWriterOutput = (r.conf.ReportPassed || r.conf.Verbose) && hasGW
		if summary.LeafNodeType.Is(types.NodeTypesForSuiteSetup...) {
			if r.conf.Verbose || hasStd {
				header = fmt.Sprintf("%s PASSED", denoter)
			} else {
				return
			}
		} else {
			header, stream = denoter, true
			if summary.NumAttempts > 1 {
				header, stream = fmt.Sprintf("%s [FLAKEY TEST - TOOK %d ATTEMPTS TO PASS]", r.retryDenoter, summary.NumAttempts), false
			}
			if summary.RunTime.Seconds() > r.conf.SlowSpecThreshold {
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
		if r.conf.Succinct || summary.Failure.Message == "" {
			header, stream = "S", true
		} else {
			header, succinctLocationBlock = "S [SKIPPED]", !r.conf.Verbose
		}
	case types.SpecStateFailed:
		highlightColor, header = "{{red}}", fmt.Sprintf("%s [FAILED]", denoter)
		r.failures = append(r.failures, summary)
	case types.SpecStatePanicked:
		highlightColor, header = "{{magenta}}", fmt.Sprintf("%s! [PANICKED]", denoter)
		r.failures = append(r.failures, summary)
	case types.SpecStateInterrupted:
		highlightColor, header = "{{orange}}", fmt.Sprintf("%s! [INTERRUPTED]", denoter)
		r.failures = append(r.failures, summary)
	}

	// Emit stream and return
	if stream {
		r.emit(r.f(highlightColor + header + "{{/}}"))
		return
	}

	// Emit header
	r.emitDelimiter()
	if includeRuntime {
		header = r.f("%s [%.3f seconds]", header, summary.RunTime.Seconds())
	}
	r.emitBlock(r.f(highlightColor + header + "{{/}}"))

	// Emit Code Location Block
	r.emitBlock(r.codeLocationBlock(summary, highlightColor, succinctLocationBlock))

	//Emit Stdout/Stderr Output
	if hasStd {
		r.emitBlock("\n")
		r.emitBlock(r.fi(1, "{{gray}}Begin Captured StdOut/StdErr Output >>{{/}}"))
		r.emitBlock(r.fi(2, "%s", summary.CapturedStdOutErr))
		r.emitBlock(r.fi(1, "{{gray}}<< End Captured StdOut/StdErr Output{{/}}"))
	}

	//Emit Captured GinkgoWriter Output
	if emitGinkgoWriterOutput && hasGW {
		r.emitBlock("\n")
		r.emitBlock(r.fi(1, "{{gray}}Begin Captured GinkgoWriter Output >>{{/}}"))
		r.emitBlock(r.fi(2, "%s", summary.CapturedGinkgoWriterOutput))
		r.emitBlock(r.fi(1, "{{gray}}<< End Captured GinkgoWriter Output{{/}}"))
	}

	// Emit Failure Message
	if !summary.Failure.IsZero() {
		r.emitBlock("\n")
		r.emitBlock(r.fi(1, highlightColor+"%s{{/}}", summary.Failure.Message))
		r.emitBlock(r.fi(1, highlightColor+"In {{bold}}[%s]{{/}}"+highlightColor+" at: {{bold}}%s{{/}}\n", summary.Failure.NodeType, summary.Failure.Location))
		if summary.Failure.ForwardedPanic != "" {
			r.emitBlock("\n")
			r.emitBlock(r.fi(1, highlightColor+"%s{{/}}", summary.Failure.ForwardedPanic))
		}

		if r.conf.FullTrace || summary.Failure.ForwardedPanic != "" {
			r.emitBlock("\n")
			r.emitBlock(r.fi(1, highlightColor+"Full Stack Trace{{/}}"))
			r.emitBlock(r.fi(2, "%s", summary.Failure.Location.FullStackTrace))
		}
	}

	r.emitDelimiter()
}

func (r *DefaultReporter) SpecSuiteDidEnd(summary types.SuiteSummary) {
	if len(r.failures) > 1 {
		r.emitBlock("\n\n")
		r.emitBlock(r.f("{{red}}{{bold}}Summarizing %d Failures:{{/}}", len(r.failures)))
		for _, summary := range r.failures {
			highlightColor, heading := "{{red}}", "[FAIL]"
			if summary.State.Is(types.SpecStateInterrupted) {
				highlightColor, heading = "{{orange}}", "[INTERRUPTED]"
			} else if summary.State.Is(types.SpecStatePanicked) {
				highlightColor, heading = "{{magenta}}", "[PANICKED!]"
			}

			locationBlock := r.codeLocationBlock(summary, highlightColor, true)
			r.emitBlock(r.fi(1, highlightColor+"%s{{/}} %s", heading, locationBlock))
		}
	}

	//summarize the suite
	if r.conf.Succinct && summary.SuiteSucceeded {
		r.emit(r.f(" {{green}}SUCCESS!{{/}} %s ", summary.RunTime))
		return
	}

	r.emitBlock("\n")
	color, status := "{{green}}{{bold}}", "SUCCESS!"
	if !summary.SuiteSucceeded {
		color, status = "{{red}}{{bold}}", "FAIL!"
		if r.hasFailOnPending && len(r.failures) == 0 && summary.NumberOfPendingSpecs > 0 {
			color, status = "{{yellow}}{{bold}}", "FAIL! - Detected pending specs and --fail-on-pending is set"
		}
	}
	r.emitBlock(r.f(color+"Ran %d of %d Specs in %.3f seconds{{/}}", summary.NumberOfSpecsThatRan(), summary.NumberOfTotalSpecs, summary.RunTime.Seconds()))
	r.emit(r.f(color+"%s{{/}} -- ", status))
	r.emit(r.f("{{green}}{{bold}}%d Passed{{/}} | ", summary.NumberOfPassedSpecs))
	r.emit(r.f("{{red}}{{bold}}%d Failed{{/}} | ", summary.NumberOfFailedSpecs))
	if summary.NumberOfFlakedSpecs > 0 {
		r.emit(r.f("{{light-yellow}}{{bold}}%d Flaked{{/}} | ", summary.NumberOfFlakedSpecs))
	}
	r.emit(r.f("{{yellow}}{{bold}}%d Pending{{/}} | ", summary.NumberOfPendingSpecs))
	r.emit(r.f("{{cyan}}{{bold}}%d Skipped{{/}}\n", summary.NumberOfSkippedSpecs))
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

func (r *DefaultReporter) codeLocationBlock(summary types.Summary, highlightColor string, succinct bool) string {
	out := ""

	if summary.LeafNodeType.Is(types.NodeTypesForSuiteSetup...) {
		out = r.f(highlightColor+"{{bold}}[%s]{{/}}\n", summary.LeafNodeType)
		if summary.Failure.IsZero() {
			out += r.f("{{gray}}%s{{/}}\n", summary.LeafNodeLocation)
		} else {
			out += r.f("{{gray}}%s{{/}}\n", summary.Failure.Location)
		}
		return out
	}

	if succinct {
		texts := make([]string, len(summary.NodeTexts))
		copy(texts, summary.NodeTexts)
		var codeLocation = summary.NodeLocations[len(summary.NodeLocations)-1]
		if !summary.Failure.IsZero() {
			codeLocation = summary.Failure.Location
			if summary.Failure.NodeIndex == -1 {
				texts = append([]string{r.f(highlightColor+"{{bold}}[%s]{{/}}", summary.Failure.NodeType)}, texts...)
			} else if summary.Failure.NodeIndex < len(texts) {
				i := summary.Failure.NodeIndex
				texts[i] = r.f(highlightColor+"{{bold}}[%s] %s{{/}}", summary.Failure.NodeType, texts[i])
			}
		}
		out += r.f("%s\n", r.cycleJoin(texts, " "))
		out += r.f("{{gray}}%s{{/}}", codeLocation)

		return out
	}

	indentation := uint(0)
	if !summary.Failure.IsZero() && summary.Failure.NodeIndex == -1 {
		out += r.fi(indentation, highlightColor+"{{bold}}TOP-LEVEL [%s]{{/}}\n", summary.Failure.NodeType)
		out += r.fi(indentation, "{{gray}}%s{{/}}\n", summary.Failure.Location)
		indentation += 1
	}

	for i := range summary.NodeTexts {
		if !summary.Failure.IsZero() && i == summary.Failure.NodeIndex {
			out += r.fi(indentation, highlightColor+"{{bold}}%s [%s]{{/}}\n", summary.NodeTexts[i], summary.Failure.NodeType)
		} else {
			out += r.fi(indentation, "%s\n", summary.NodeTexts[i])
		}
		out += r.fi(indentation, "{{gray}}%s{{/}}\n", summary.NodeLocations[i])
		indentation += 1
	}

	return out
}

package ginkgo

import (
	"fmt"
	"strings"
)

type defaultReporter struct {
	noColor              bool
	slowSpecThreshold    float64
	randomSeed           int64
	randomizeAllExamples bool
	lastExampleWasABlock bool
}

func newDefaultReporter(noColor bool, slowSpecThreshold float64) *defaultReporter {
	return &defaultReporter{
		noColor:           noColor,
		slowSpecThreshold: slowSpecThreshold,
	}
}

const defaultStyle = "\x1b[0m"
const boldStyle = "\x1b[1m"
const redColor = "\x1b[91m"
const greenColor = "\x1b[32m"
const yellowColor = "\x1b[33m"
const cyanColor = "\x1b[36m"
const grayColor = "\x1b[90m"
const blueColor = "\x1b[94m"

func (reporter *defaultReporter) colorize(colorCode string, format string, args ...interface{}) string {
	s := fmt.Sprintf(format, args...)
	if reporter.noColor {
		return s
	} else {
		return fmt.Sprintf("%s%s%s", colorCode, s, defaultStyle)
	}
}

func (reporter *defaultReporter) printBanner(s string, bannerCharacter string) {
	fmt.Println(s)
	fmt.Println(strings.Repeat(bannerCharacter, len(s)))
}

func (reporter *defaultReporter) printNewLine() {
	fmt.Println("")
}

func (reporter *defaultReporter) printDelimiter() {
	fmt.Println(reporter.colorize(grayColor, "%s", strings.Repeat("-", 30)))
}

func (reporter *defaultReporter) indent(indentation int, format string, args ...interface{}) string {
	text := fmt.Sprintf(format, args...)

	stringArray := strings.Split(text, "\n")
	padding := strings.Repeat("  ", indentation)
	for i, s := range stringArray {
		stringArray[i] = fmt.Sprintf("%s%s", padding, s)
	}

	return strings.Join(stringArray, "\n")
}

func (reporter *defaultReporter) print(indentation int, format string, args ...interface{}) {
	fmt.Print(reporter.indent(indentation, format, args...))
}

func (reporter *defaultReporter) println(indentation int, format string, args ...interface{}) {
	fmt.Println(reporter.indent(indentation, format, args...))
}

func (reporter *defaultReporter) RandomizationStrategy(randomSeed int64, randomizeAllExamples bool) {
	reporter.randomSeed = randomSeed
	reporter.randomizeAllExamples = randomizeAllExamples
}

func (reporter *defaultReporter) SpecSuiteWillBegin(summary *SuiteSummary) {
	reporter.printNewLine()
	reporter.printBanner(fmt.Sprintf("Running Suite: %s", summary.SuiteDescription), "=")
	randomSeedReport := fmt.Sprintf("Random Seed: %s", reporter.colorize(boldStyle, "%d", reporter.randomSeed))
	if reporter.randomizeAllExamples {
		randomSeedReport += " - Will randomize all examples"
	}
	reporter.println(0, randomSeedReport)
	reporter.printNewLine()

	reporter.println(0,
		"Will run %s of %s specs",
		reporter.colorize(boldStyle, "%d", summary.NumberOfExamplesThatWillBeRun),
		reporter.colorize(boldStyle, "%d", summary.NumberOfTotalExamples))

	reporter.printNewLine()
}

func (reporter *defaultReporter) ExampleDidComplete(exampleSummary *ExampleSummary) {
	if exampleSummary.State == ExampleStatePassed {
		reporter.printStatus(greenColor, "•", exampleSummary)
	} else if exampleSummary.State == ExampleStatePending {
		reporter.printStatus(yellowColor, "P", exampleSummary)
	} else if exampleSummary.State == ExampleStateSkipped {
		reporter.printStatus(cyanColor, "S", exampleSummary)
	} else if exampleSummary.State == ExampleStateTimedOut {
		reporter.printFailure("•... Timeout", exampleSummary)
	} else if exampleSummary.State == ExampleStatePanicked {
		reporter.printFailure("•! Panic", exampleSummary)
	} else if exampleSummary.State == ExampleStateFailed {
		reporter.printFailure("• Failure", exampleSummary)
	}
}

func (reporter *defaultReporter) SpecSuiteDidEnd(summary *SuiteSummary) {
	reporter.printNewLine()
	color := greenColor
	if summary.NumberOfFailedExamples > 0 {
		color = redColor
	}
	reporter.println(0, reporter.colorize(boldStyle+color, "Ran %d of %d Specs in %.3f seconds", summary.NumberOfExamplesThatWillBeRun, summary.NumberOfTotalExamples, summary.RunTime.Seconds()))

	status := ""
	if summary.NumberOfFailedExamples == 0 {
		status = fmt.Sprintf(reporter.colorize(boldStyle+greenColor, "SUCCESS!"))
	} else {
		status = fmt.Sprintf(reporter.colorize(boldStyle+redColor, "FAIL!"))
	}

	reporter.println(0,
		"%s -- %s | %s | %s | %s",
		status,
		reporter.colorize(greenColor+boldStyle, "%d Passed", summary.NumberOfPassedExamples),
		reporter.colorize(redColor+boldStyle, "%d Failed", summary.NumberOfFailedExamples),
		reporter.colorize(yellowColor+boldStyle, "%d Pending", summary.NumberOfPendingExamples),
		reporter.colorize(cyanColor+boldStyle, "%d Skipped", summary.NumberOfSkippedExamples),
	)
	reporter.printNewLine()
}

func (reporter *defaultReporter) printStatus(color string, message string, exampleSummary *ExampleSummary) {
	if exampleSummary.RunTime.Seconds() >= reporter.slowSpecThreshold {
		if !reporter.lastExampleWasABlock {
			reporter.printNewLine()
			reporter.printDelimiter()
		}
		reporter.println(0, reporter.colorize(color, "%s [SLOW TEST:%.3f seconds]", message, exampleSummary.RunTime.Seconds()))

		for i := 1; i < len(exampleSummary.ComponentTexts); i++ {
			reporter.println(i-1, "%s", exampleSummary.ComponentTexts[i])
			reporter.println(i-1, reporter.colorize(grayColor, "(%s)", exampleSummary.ComponentCodeLocations[i]))
		}

		reporter.printDelimiter()
		reporter.lastExampleWasABlock = true
	} else {
		reporter.print(0, reporter.colorize(color, message))
		reporter.lastExampleWasABlock = false
	}
}

func (reporter *defaultReporter) printFailure(message string, exampleSummary *ExampleSummary) {
	if !reporter.lastExampleWasABlock {
		reporter.printNewLine()
		reporter.printDelimiter()
	}
	reporter.println(0, reporter.colorize(redColor+boldStyle, "%s [%.3f seconds]", message, exampleSummary.RunTime.Seconds()))
	for i := 1; i < len(exampleSummary.ComponentTexts); i++ {
		if i == exampleSummary.Failure.ComponentIndex {
			blockType := ""
			switch exampleSummary.Failure.ComponentType {
			case ExampleComponentTypeBeforeEach:
				blockType = "BeforeEach"
			case ExampleComponentTypeJustBeforeEach:
				blockType = "JustBeforeEach"
			case ExampleComponentTypeAfterEach:
				blockType = "AfterEach"
			case ExampleComponentTypeIt:
				blockType = "It"
			}
			reporter.println(i-1, reporter.colorize(redColor+boldStyle, "%s [%s]", exampleSummary.ComponentTexts[i], blockType))
			reporter.println(i-1, reporter.colorize(grayColor, "(%s)", exampleSummary.ComponentCodeLocations[i]))
		} else {
			reporter.println(i-1, exampleSummary.ComponentTexts[i])
			reporter.println(i-1, reporter.colorize(grayColor, "(%s)", exampleSummary.ComponentCodeLocations[i]))
		}
	}

	indentation := exampleSummary.Failure.ComponentIndex - 1

	reporter.printNewLine()
	if exampleSummary.State == ExampleStatePanicked {
		reporter.println(indentation, reporter.colorize(redColor+boldStyle, exampleSummary.Failure.Message))
		reporter.println(indentation, reporter.colorize(redColor, "%s", exampleSummary.Failure.ForwardedPanic))
		reporter.println(indentation, exampleSummary.Failure.Location.String())
		reporter.printNewLine()
		reporter.println(indentation, reporter.colorize(redColor, "Full Stack Trace"))
		reporter.println(indentation, exampleSummary.Failure.Location.FullStackTrace)
	} else {
		reporter.println(indentation, reporter.colorize(redColor, exampleSummary.Failure.Message))
		reporter.printNewLine()
		reporter.println(indentation, exampleSummary.Failure.Location.String())
	}

	reporter.printDelimiter()
	reporter.lastExampleWasABlock = true
}

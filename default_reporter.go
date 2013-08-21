package godescribe

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

func (reporter *defaultReporter) banner(s string, bannerCharacter string) {
	fmt.Println(s)
	fmt.Println(strings.Repeat(bannerCharacter, len(s)))
}

func (reporter *defaultReporter) RandomizationStrategy(randomSeed int64, randomizeAllExamples bool) {
	reporter.randomSeed = randomSeed
	reporter.randomizeAllExamples = randomizeAllExamples
}

func (reporter *defaultReporter) SpecSuiteWillBegin(summary *SuiteSummary) {
	fmt.Println("")
	reporter.banner(fmt.Sprintf("Running Suite: %s", summary.SuiteDescription), "=")
	fmt.Printf("Random Seed: %s", reporter.colorize(boldStyle, "%d", reporter.randomSeed))
	if reporter.randomizeAllExamples {
		fmt.Print(" - Will randomize all examples")
	}
	fmt.Println("")

	fmt.Printf(
		"Will run %s of %s specs\n",
		reporter.colorize(boldStyle, "%d", summary.NumberOfExamplesThatWillBeRun),
		reporter.colorize(boldStyle, "%d", summary.NumberOfTotalExamples))

	fmt.Println("")
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
	fmt.Println("")
	color := greenColor
	if summary.NumberOfFailedExamples > 0 {
		color = redColor
	}
	fmt.Println(reporter.colorize(boldStyle+color, "Ran %d of %d Specs in %.3f seconds", summary.NumberOfExamplesThatWillBeRun, summary.NumberOfTotalExamples, summary.RunTime.Seconds()))

	status := ""
	if summary.NumberOfFailedExamples == 0 {
		status = fmt.Sprintf(reporter.colorize(boldStyle+greenColor, "SUCCESS!"))
	} else {
		status = fmt.Sprintf(reporter.colorize(boldStyle+redColor, "FAIL!"))
	}

	fmt.Printf(
		"%s -- %s | %s | %s | %s\n",
		status,
		reporter.colorize(greenColor+boldStyle, "%d Passed", summary.NumberOfPassedExamples),
		reporter.colorize(redColor+boldStyle, "%d Failed", summary.NumberOfFailedExamples),
		reporter.colorize(yellowColor+boldStyle, "%d Pending", summary.NumberOfPendingExamples),
		reporter.colorize(cyanColor+boldStyle, "%d Skipped", summary.NumberOfSkippedExamples),
	)
	fmt.Println("")
}

func (reporter *defaultReporter) printStatus(color string, message string, exampleSummary *ExampleSummary) {
	if exampleSummary.RunTime.Seconds() >= reporter.slowSpecThreshold {
		if !reporter.lastExampleWasABlock {
			fmt.Println("")
			reporter.printDelimiter()
		}
		fmt.Print(reporter.colorize(color, "%s [SLOW TEST:%.3f seconds]\n", message, exampleSummary.RunTime.Seconds()))

		for i := 1; i < len(exampleSummary.ComponentTexts); i++ {
			padding := strings.Repeat("  ", i-1)
			fmt.Printf("%s%s\n", padding, exampleSummary.ComponentTexts[i])
			fmt.Println(reporter.colorize(grayColor, "%s(%s)", padding, exampleSummary.ComponentCodeLocations[i]))
		}

		reporter.printDelimiter()
		reporter.lastExampleWasABlock = true
	} else {
		fmt.Print(reporter.colorize(color, message))
		reporter.lastExampleWasABlock = false
	}
}

func (reporter *defaultReporter) printFailure(message string, exampleSummary *ExampleSummary) {
	if !reporter.lastExampleWasABlock {
		fmt.Println("")
		reporter.printDelimiter()
	}
	fmt.Print(reporter.colorize(redColor+boldStyle, "%s [%.3f seconds]\n", message, exampleSummary.RunTime.Seconds()))
	for i := 1; i < len(exampleSummary.ComponentTexts); i++ {
		padding := strings.Repeat("  ", i-1)
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
			fmt.Println(reporter.colorize(redColor+boldStyle, "%s%s [%s]", padding, exampleSummary.ComponentTexts[i], blockType))
			fmt.Println(reporter.colorize(grayColor, "%s(%s)", padding, exampleSummary.ComponentCodeLocations[i]))
		} else {
			fmt.Printf("%s%s\n", padding, exampleSummary.ComponentTexts[i])
			fmt.Println(reporter.colorize(grayColor, "%s(%s)", padding, exampleSummary.ComponentCodeLocations[i]))
		}
	}

	padding := strings.Repeat("  ", exampleSummary.Failure.ComponentIndex-1)

	fmt.Println("")
	if exampleSummary.State == ExampleStatePanicked {
		fmt.Println(reporter.colorize(redColor+boldStyle, "%s%s", padding, exampleSummary.Failure.Message))
		fmt.Println(reporter.colorize(redColor, "%s> %s", padding, exampleSummary.Failure.ForwardedPanic))
	} else {
		fmt.Println(reporter.colorize(redColor, "%s> %s", padding, exampleSummary.Failure.Message))
	}

	fmt.Printf("%s%s\n", padding, exampleSummary.Failure.Location)
	reporter.printDelimiter()
	reporter.lastExampleWasABlock = true
}

func (reporter *defaultReporter) printDelimiter() {
	fmt.Println(reporter.colorize(grayColor, "%s", strings.Repeat("-", 30)))
}

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
	lastExampleFailed    bool
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

func (reporter *defaultReporter) printFailure(exampleSummary *ExampleSummary) {
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
}

func (reporter *defaultReporter) printErrorDelimiter() {
	fmt.Println(reporter.colorize(redColor, "%s", strings.Repeat("-", 10)))
}

func (reporter *defaultReporter) ExampleDidComplete(exampleSummary *ExampleSummary) {
	if exampleSummary.State == ExampleStatePassed {
		fmt.Print(reporter.colorize(greenColor, "•"))
		reporter.lastExampleFailed = false
	} else if exampleSummary.State == ExampleStatePending {
		fmt.Print(reporter.colorize(yellowColor, "P"))
		reporter.lastExampleFailed = false
	} else if exampleSummary.State == ExampleStateSkipped {
		fmt.Print(reporter.colorize(cyanColor, "S"))
		reporter.lastExampleFailed = false
	} else if exampleSummary.State == ExampleStatePanicked {
		if !reporter.lastExampleFailed {
			fmt.Println("")
			reporter.printErrorDelimiter()
		}
		fmt.Print(reporter.colorize(redColor+boldStyle, "•! Panic [%.3f seconds]\n", exampleSummary.RunTime.Seconds()))
		reporter.lastExampleFailed = true
		reporter.printFailure(exampleSummary)
		reporter.printErrorDelimiter()
	} else if exampleSummary.State == ExampleStateFailed {
		if !reporter.lastExampleFailed {
			fmt.Println("")
			reporter.printErrorDelimiter()
		}
		fmt.Print(reporter.colorize(redColor+boldStyle, "• Failure [%.3f seconds]\n", exampleSummary.RunTime.Seconds()))
		reporter.lastExampleFailed = true
		reporter.printFailure(exampleSummary)
		reporter.printErrorDelimiter()
	}
}

func (reporter *defaultReporter) SpecSuiteDidEnd(summary *SuiteSummary) {
	fmt.Println("")
	fmt.Println(reporter.colorize(boldStyle, "Finished in %.3f seconds", summary.RunTime.Seconds()))
}

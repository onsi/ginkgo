/*
The stenographer is used by Ginkgo's reporters to generate output.

Move along, nothing to see here.
*/

package stenographer

import (
	"fmt"
	"github.com/onsi/ginkgo/types"
	"strings"
)

const defaultStyle = "\x1b[0m"
const boldStyle = "\x1b[1m"
const redColor = "\x1b[91m"
const greenColor = "\x1b[32m"
const yellowColor = "\x1b[33m"
const cyanColor = "\x1b[36m"
const grayColor = "\x1b[90m"
const lightGrayColor = "\x1b[37m"

type cursorStateType int

const (
	cursorStateTop cursorStateType = iota
	cursorStateStreaming
	cursorStateMidBlock
	cursorStateEndBlock
)

type Stenographer interface {
	AnnounceSuite(description string, randomSeed int64, randomizingAll bool)
	AnnounceAggregatedParallelRun(nodes int)
	AnnounceParallelRun(node int, nodes int, specsToRun int, totalSpecs int)
	AnnounceNumberOfSpecs(specsToRun int, total int)
	AnnounceSpecRunCompletion(summary *types.SuiteSummary)

	AnnounceExampleWillRun(example *types.ExampleSummary)

	AnnounceCapturedOutput(example *types.ExampleSummary)

	AnnounceSuccesfulExample(example *types.ExampleSummary)
	AnnounceSuccesfulSlowExample(example *types.ExampleSummary, succinct bool)
	AnnounceSuccesfulMeasurement(example *types.ExampleSummary, succinct bool)

	AnnouncePendingExample(example *types.ExampleSummary, noisy bool, succinct bool)
	AnnounceSkippedExample(example *types.ExampleSummary)

	AnnounceExampleTimedOut(example *types.ExampleSummary, succinct bool)
	AnnounceExamplePanicked(example *types.ExampleSummary, succinct bool)
	AnnounceExampleFailed(example *types.ExampleSummary, succinct bool)
}

func New(color bool) Stenographer {
	return &consoleStenographer{
		color:       color,
		cursorState: cursorStateTop,
	}
}

type consoleStenographer struct {
	color       bool
	cursorState cursorStateType
}

var alternatingColors = []string{defaultStyle, grayColor}

func (s *consoleStenographer) AnnounceSuite(description string, randomSeed int64, randomizingAll bool) {
	s.printNewLine()
	s.printBanner(fmt.Sprintf("Running Suite: %s", description), "=")
	s.print(0, "Random Seed: %s", s.colorize(boldStyle, "%d", randomSeed))
	if randomizingAll {
		s.print(0, " - Will randomize all examples")
	}
	s.printNewLine()
}

func (s *consoleStenographer) AnnounceParallelRun(node int, nodes int, specsToRun int, totalSpecs int) {
	s.println(0,
		"Parallel test node %s/%s. Assigned %s of %s specs.",
		s.colorize(boldStyle, "%d", node),
		s.colorize(boldStyle, "%d", nodes),
		s.colorize(boldStyle, "%d", specsToRun),
		s.colorize(boldStyle, "%d", totalSpecs),
	)
	s.printNewLine()
}

func (s *consoleStenographer) AnnounceAggregatedParallelRun(nodes int) {
	s.println(0,
		"Running in parallel across %s nodes",
		s.colorize(boldStyle, "%d", nodes),
	)
	s.printNewLine()
}

func (s *consoleStenographer) AnnounceNumberOfSpecs(specsToRun int, total int) {
	s.println(0,
		"Will run %s of %s specs",
		s.colorize(boldStyle, "%d", specsToRun),
		s.colorize(boldStyle, "%d", total),
	)

	s.printNewLine()
}

func (s *consoleStenographer) AnnounceSpecRunCompletion(summary *types.SuiteSummary) {
	s.printNewLine()
	color := greenColor
	if !summary.SuiteSucceeded {
		color = redColor
	}
	s.println(0, s.colorize(boldStyle+color, "Ran %d of %d Specs in %.3f seconds", summary.NumberOfExamplesThatWillBeRun, summary.NumberOfTotalExamples, summary.RunTime.Seconds()))

	status := ""
	if summary.SuiteSucceeded {
		status = s.colorize(boldStyle+greenColor, "SUCCESS!")
	} else {
		status = s.colorize(boldStyle+redColor, "FAIL!")
	}

	s.println(0,
		"%s -- %s | %s | %s | %s",
		status,
		s.colorize(greenColor+boldStyle, "%d Passed", summary.NumberOfPassedExamples),
		s.colorize(redColor+boldStyle, "%d Failed", summary.NumberOfFailedExamples),
		s.colorize(yellowColor+boldStyle, "%d Pending", summary.NumberOfPendingExamples),
		s.colorize(cyanColor+boldStyle, "%d Skipped", summary.NumberOfSkippedExamples),
	)
	s.printNewLine()
}

func (s *consoleStenographer) AnnounceExampleWillRun(example *types.ExampleSummary) {
	if s.cursorState == cursorStateStreaming {
		s.printNewLine()
		s.printDelimiter()
	}
	for i, text := range example.ComponentTexts[1 : len(example.ComponentTexts)-1] {
		s.print(0, s.colorize(alternatingColors[i%2], text)+" ")
	}

	indentation := 0
	if len(example.ComponentTexts) > 2 {
		indentation = 1
		s.printNewLine()
	}
	index := len(example.ComponentTexts) - 1
	s.print(indentation, s.colorize(boldStyle, example.ComponentTexts[index]))
	s.printNewLine()
	s.print(indentation, s.colorize(lightGrayColor, example.ComponentCodeLocations[index].String()))
	s.printNewLine()
	s.cursorState = cursorStateMidBlock
}

func (s *consoleStenographer) AnnounceCapturedOutput(example *types.ExampleSummary) {
	if example.CapturedOutput == "" {
		return
	}

	if s.cursorState == cursorStateStreaming {
		s.printNewLine()
		s.printDelimiter()
	} else if s.cursorState == cursorStateMidBlock {
		s.printNewLine()
	}
	s.println(0, example.CapturedOutput)
	s.cursorState = cursorStateMidBlock
}

func (s *consoleStenographer) AnnounceSuccesfulExample(example *types.ExampleSummary) {
	s.print(0, s.colorize(greenColor, "•"))
	s.cursorState = cursorStateStreaming
}

func (s *consoleStenographer) AnnounceSuccesfulSlowExample(example *types.ExampleSummary, succinct bool) {
	s.printBlockWithMessage(
		s.colorize(greenColor, "• [SLOW TEST:%.3f seconds]", example.RunTime.Seconds()),
		"",
		example,
		succinct,
	)
}

func (s *consoleStenographer) AnnounceSuccesfulMeasurement(example *types.ExampleSummary, succinct bool) {
	s.printBlockWithMessage(
		s.colorize(greenColor, "• [MEASUREMENT]"),
		s.measurementReport(example),
		example,
		succinct,
	)
}

func (s *consoleStenographer) AnnouncePendingExample(example *types.ExampleSummary, noisy bool, succinct bool) {
	if noisy {
		s.printBlockWithMessage(
			s.colorize(yellowColor, "P [PENDING]"),
			"",
			example,
			succinct,
		)
	} else {
		s.print(0, s.colorize(greenColor, "P"))
		s.cursorState = cursorStateStreaming
	}
}

func (s *consoleStenographer) AnnounceSkippedExample(example *types.ExampleSummary) {
	s.print(0, s.colorize(cyanColor, "S"))
	s.cursorState = cursorStateStreaming
}

func (s *consoleStenographer) AnnounceExampleTimedOut(example *types.ExampleSummary, succinct bool) {
	s.printFailure("•... Timeout", example, succinct)
}

func (s *consoleStenographer) AnnounceExamplePanicked(example *types.ExampleSummary, succinct bool) {
	s.printFailure("•! Panic", example, succinct)
}

func (s *consoleStenographer) AnnounceExampleFailed(example *types.ExampleSummary, succinct bool) {
	s.printFailure("• Failure", example, succinct)
}

func (s *consoleStenographer) printBlockWithMessage(header string, message string, example *types.ExampleSummary, succinct bool) {
	if s.cursorState == cursorStateStreaming {
		s.printNewLine()
		s.printDelimiter()
	} else if s.cursorState == cursorStateMidBlock {
		s.printNewLine()
	}

	s.println(0, header)

	indentation := s.printCodeLocationBlock(example, false, succinct)

	if message != "" {
		s.printNewLine()
		s.println(indentation, message)
	}

	s.printDelimiter()
	s.cursorState = cursorStateEndBlock
}

func (s *consoleStenographer) printFailure(message string, example *types.ExampleSummary, succinct bool) {
	if s.cursorState == cursorStateStreaming {
		s.printNewLine()
		s.printDelimiter()
	} else if s.cursorState == cursorStateMidBlock {
		s.printNewLine()
	}

	s.println(0, s.colorize(redColor+boldStyle, "%s [%.3f seconds]", message, example.RunTime.Seconds()))

	indentation := s.printCodeLocationBlock(example, true, succinct)

	s.printNewLine()
	if example.State == types.ExampleStatePanicked {
		s.println(indentation, s.colorize(redColor+boldStyle, example.Failure.Message))
		s.println(indentation, s.colorize(redColor, "%v", example.Failure.ForwardedPanic))
		s.println(indentation, example.Failure.Location.String())
		s.printNewLine()
		s.println(indentation, s.colorize(redColor, "Full Stack Trace"))
		s.println(indentation, example.Failure.Location.FullStackTrace)
	} else {
		s.println(indentation, s.colorize(redColor, example.Failure.Message))
		s.printNewLine()
		s.println(indentation, example.Failure.Location.String())
	}

	s.printDelimiter()
	s.cursorState = cursorStateEndBlock
}

func (s *consoleStenographer) printCodeLocationBlock(example *types.ExampleSummary, failure bool, succinct bool) int {
	indentation := 0
	startIndex := 1

	if len(example.ComponentTexts) == 1 {
		startIndex = 0
	}

	for i := startIndex; i < len(example.ComponentTexts); i++ {
		if failure && i == example.Failure.ComponentIndex {
			blockType := ""
			switch example.Failure.ComponentType {
			case types.ExampleComponentTypeBeforeEach:
				blockType = "BeforeEach"
			case types.ExampleComponentTypeJustBeforeEach:
				blockType = "JustBeforeEach"
			case types.ExampleComponentTypeAfterEach:
				blockType = "AfterEach"
			case types.ExampleComponentTypeIt:
				blockType = "It"
			case types.ExampleComponentTypeMeasure:
				blockType = "Measurement"
			}
			if succinct {
				s.print(0, s.colorize(redColor+boldStyle, "[%s] %s ", blockType, example.ComponentTexts[i]))
			} else {
				s.println(indentation, s.colorize(redColor+boldStyle, "%s [%s]", example.ComponentTexts[i], blockType))
				s.println(indentation, s.colorize(grayColor, "(%s)", example.ComponentCodeLocations[i]))
			}
		} else {
			if succinct {
				s.print(0, s.colorize(alternatingColors[i%2], "%s ", example.ComponentTexts[i]))
			} else {
				s.println(indentation, example.ComponentTexts[i])
				s.println(indentation, s.colorize(grayColor, "(%s)", example.ComponentCodeLocations[i]))
			}
		}

		indentation++
	}

	if succinct {
		if len(example.ComponentTexts) > 0 {
			s.printNewLine()
			s.print(0, s.colorize(lightGrayColor, "(%s)", example.ComponentCodeLocations[len(example.ComponentCodeLocations)-1]))
		}
		s.printNewLine()
		indentation = 1
	} else {
		indentation--
	}

	return indentation
}

func (s *consoleStenographer) measurementReport(example *types.ExampleSummary) string {
	if len(example.Measurements) == 0 {
		return "Found no measurements"
	}

	message := []string{}

	message = append(message, fmt.Sprintf("Ran %s samples:", s.colorize(boldStyle, "%d", example.NumberOfSamples)))
	i := 0
	for _, measurement := range example.Measurements {
		if i > 0 {
			message = append(message, "\n")
		}
		info := ""
		if measurement.Info != nil {
			message = append(message, fmt.Sprintf("%v", measurement.Info))
		}

		message = append(message, fmt.Sprintf("%s:\n%s  %s: %s%s\n  %s: %s%s\n  %s: %s%s ± %s%s",
			s.colorize(boldStyle, "%s", measurement.Name),
			info,
			measurement.SmallestLabel,
			s.colorize(greenColor, "%.3f", measurement.Smallest),
			measurement.Units,
			measurement.LargestLabel,
			s.colorize(redColor, "%.3f", measurement.Largest),
			measurement.Units,
			measurement.AverageLabel,
			s.colorize(cyanColor, "%.3f", measurement.Average),
			measurement.Units,
			s.colorize(cyanColor, "%.3f", measurement.StdDeviation),
			measurement.Units,
		))
		i++
	}

	return strings.Join(message, "\n")
}

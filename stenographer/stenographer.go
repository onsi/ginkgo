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

type Stenographer struct {
	color       bool
	cursorState cursorStateType
}

func New(color bool) *Stenographer {
	return &Stenographer{
		color:       color,
		cursorState: cursorStateTop,
	}
}

func (s *Stenographer) AnnounceSuite(description string, randomSeed int64, randomizingAll bool) {
	s.printNewLine()
	s.printBanner(fmt.Sprintf("Running Suite: %s", description), "=")
	s.print(0, "Random Seed: %s", s.colorize(boldStyle, "%d", randomSeed))
	if randomizingAll {
		s.print(0, " - Will randomize all examples")
	}
	s.printNewLine()
}

func (s *Stenographer) AnnounceParallelRun(node int, nodes int, specsToRun int, totalSpecs int) {
	s.println(0,
		"Parallel test node %s/%s. Assigned %s of %s specs.",
		s.colorize(boldStyle, "%d", node),
		s.colorize(boldStyle, "%d", nodes),
		s.colorize(boldStyle, "%d", specsToRun),
		s.colorize(boldStyle, "%d", totalSpecs),
	)
	s.printNewLine()
}

func (s *Stenographer) AnnounceNumberOfSpecs(specsToRun int, total int) {
	s.println(0,
		"Will run %s of %s specs",
		s.colorize(boldStyle, "%d", specsToRun),
		s.colorize(boldStyle, "%d", total),
	)

	s.printNewLine()
}

func (s *Stenographer) AnnounceSpecRunCompletion(summary *types.SuiteSummary) {
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

func (s *Stenographer) AnnounceExampleWillRun(example *types.ExampleSummary) {
	if s.cursorState == cursorStateStreaming {
		s.printNewLine()
		s.printDelimiter()
	}
	colors := []string{defaultStyle, grayColor}
	for i, text := range example.ComponentTexts[1 : len(example.ComponentTexts)-1] {
		s.print(0, s.colorize(colors[i%2], text)+" ")
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

func (s *Stenographer) AnnounceSuccesfulExample(example *types.ExampleSummary) {
	s.print(0, s.colorize(greenColor, "•"))
	s.cursorState = cursorStateStreaming
}

func (s *Stenographer) AnnounceSuccesfulSlowExample(example *types.ExampleSummary) {
	s.printBlockWithMessage(
		s.colorize(greenColor, "• [SLOW TEST:%.3f seconds]", example.RunTime.Seconds()),
		"",
		example,
	)
}

func (s *Stenographer) AnnounceSuccesfulMeasurement(example *types.ExampleSummary) {
	s.printBlockWithMessage(
		s.colorize(greenColor, "• [MEASUREMENT]"),
		s.measurementReport(example),
		example,
	)
}

func (s *Stenographer) AnnouncePendingExample(example *types.ExampleSummary, noisy bool) {
	if noisy {
		s.printBlockWithMessage(
			s.colorize(yellowColor, "P [PENDING]"),
			"",
			example,
		)
	} else {
		s.print(0, s.colorize(greenColor, "P"))
		s.cursorState = cursorStateStreaming
	}
}

func (s *Stenographer) AnnounceSkippedExample(example *types.ExampleSummary) {
	s.print(0, s.colorize(cyanColor, "S"))
	s.cursorState = cursorStateStreaming
}

func (s *Stenographer) AnnounceExampleTimedOut(example *types.ExampleSummary) {
	s.printFailure("•... Timeout", example)
}

func (s *Stenographer) AnnounceExamplePanicked(example *types.ExampleSummary) {
	s.printFailure("•! Panic", example)
}

func (s *Stenographer) AnnounceExampleFailed(example *types.ExampleSummary) {
	s.printFailure("• Failure", example)
}

func (s *Stenographer) printBlockWithMessage(header string, message string, example *types.ExampleSummary) {
	if s.cursorState == cursorStateStreaming {
		s.printNewLine()
		s.printDelimiter()
	} else if s.cursorState == cursorStateMidBlock {
		s.printNewLine()
	}

	s.println(0, header)

	startIndex := 1
	if len(example.ComponentTexts) == 1 {
		startIndex = 0
	}

	indentation := 0
	for i := startIndex; i < len(example.ComponentTexts); i++ {
		s.println(indentation, "%s", example.ComponentTexts[i])
		s.println(indentation, s.colorize(grayColor, "(%s)", example.ComponentCodeLocations[i]))
		indentation++
	}

	if message != "" {
		s.printNewLine()
		s.println(indentation-1, message)
	}

	s.printDelimiter()
	s.cursorState = cursorStateEndBlock
}

func (s *Stenographer) printFailure(message string, example *types.ExampleSummary) {
	if s.cursorState == cursorStateStreaming {
		s.printNewLine()
		s.printDelimiter()
	} else if s.cursorState == cursorStateMidBlock {
		s.printNewLine()
	}

	s.println(0, s.colorize(redColor+boldStyle, "%s [%.3f seconds]", message, example.RunTime.Seconds()))

	startIndex := 1
	if example.Failure.ComponentIndex == 0 {
		startIndex = 0
	}

	indentation := 0
	for i := startIndex; i < len(example.ComponentTexts); i++ {
		if i == example.Failure.ComponentIndex {
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
			s.println(indentation, s.colorize(redColor+boldStyle, "%s [%s]", example.ComponentTexts[i], blockType))
			s.println(indentation, s.colorize(grayColor, "(%s)", example.ComponentCodeLocations[i]))
		} else {
			s.println(indentation, example.ComponentTexts[i])
			s.println(indentation, s.colorize(grayColor, "(%s)", example.ComponentCodeLocations[i]))
		}

		indentation++
	}

	indentation--
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

func (s *Stenographer) measurementReport(example *types.ExampleSummary) string {
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

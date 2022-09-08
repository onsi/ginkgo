package internal

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2/formatter"
	"github.com/onsi/ginkgo/v2/types"
)

type ProgressSignalRegistrar func(func()) context.CancelFunc

func RegisterForProgressSignal(handler func()) context.CancelFunc {
	signalChannel := make(chan os.Signal, 1)
	if len(PROGRESS_SIGNALS) > 0 {
		signal.Notify(signalChannel, PROGRESS_SIGNALS...)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-signalChannel:
				handler()
			case <-ctx.Done():
				signal.Stop(signalChannel)
				return
			}
		}
	}()

	return cancel
}

type Goroutines []Goroutine

type Goroutine struct {
	ID    uint64
	State string
	Stack FunctionCalls
}

func (g Goroutine) Report(color string) string {
	if !g.Stack.HasHighlights() {
		color = "{{gray}}"
	}
	out := &strings.Builder{}
	fmt.Fprintf(out, "%sgoroutine %d [%s]{{/}}\n", color, g.ID, g.State)

	for _, functionCall := range g.Stack {
		fmt.Fprintf(out, "%s\n", functionCall.Report(color))
	}

	return out.String()
}

type FunctionCalls []FunctionCall

func (fcs FunctionCalls) HasHighlights() bool {
	for _, fc := range fcs {
		if fc.Highlight {
			return true
		}
	}
	return false
}

type FunctionCall struct {
	Function  string
	Filename  string
	Line      int64
	Highlight bool
}

func (fc FunctionCall) Report(color string) string {
	if fc.Highlight {
		out := fmt.Sprintf("{{bold}}%s> %s\n    %s:%d{{/}}", color, fc.Function, fc.Filename, fc.Line)
		src := sourceAt(fc.Filename, int(fc.Line), 2, color)
		if src != "" {
			out += "\n" + src
		}
		return out
	} else {
		return fmt.Sprintf("{{gray}}  %s\n    %s:%d{{/}}", fc.Function, fc.Filename, fc.Line)
	}
}

type ProgressStepCursor struct {
	Name         string
	CodeLocation types.CodeLocation
	StartTime    time.Time
}

type ProgressReport struct {
	CurrentSpecReport    types.SpecReport
	CurrentNode          Node
	CurrentNodeStartTime time.Time
	CurrentStep          ProgressStepCursor

	SpecGoroutine         Goroutine
	HighlightedGoroutines Goroutines
	OtherGoroutines       Goroutines
}

var cycleJoiner = formatter.New(formatter.ColorModePassthrough)

func (pc ProgressReport) Report(color string, includeAllGoroutines bool) string {
	now := time.Now()
	out := &strings.Builder{}
	indent := ""
	if pc.CurrentSpecReport.LeafNodeText != "" {
		if len(pc.CurrentSpecReport.ContainerHierarchyTexts) > 0 {
			out.WriteString(cycleJoiner.CycleJoin(pc.CurrentSpecReport.ContainerHierarchyTexts, " ", []string{"{{/}}", "{{gray}}"}))
			out.WriteString(" ")
		}

		fmt.Fprintf(out, "{{bold}}%s%s{{/}} (Spec Runtime: %s)\n", color, pc.CurrentSpecReport.LeafNodeText, now.Sub(pc.CurrentSpecReport.StartTime))
		fmt.Fprintf(out, "  {{gray}}%s{{/}}\n", pc.CurrentSpecReport.LeafNodeLocation)
		indent += "  "
	}
	if !pc.CurrentNode.IsZero() {
		fmt.Fprintf(out, "%sIn {{bold}}%s[%s]{{/}}", indent, color, pc.CurrentNode.NodeType)
		if pc.CurrentNode.Text != "" && !pc.CurrentNode.NodeType.Is(types.NodeTypeIt) {
			fmt.Fprintf(out, " {{bold}}%s%s{{/}}", color, pc.CurrentNode.Text)
		}
		fmt.Fprintf(out, " (Node Runtime: %s)\n", now.Sub(pc.CurrentNodeStartTime))
		fmt.Fprintf(out, "%s  {{gray}}%s{{/}}\n", indent, pc.CurrentNode.CodeLocation)
		indent += "  "
	}
	if pc.CurrentStep.Name != "" {
		fmt.Fprintf(out, "%sAt {{bold}}%s[By Step] %s{{/}} (Step Runtime: %s)\n", indent, color, pc.CurrentStep.Name, now.Sub(pc.CurrentStep.StartTime))
		fmt.Fprintf(out, "%s  {{gray}}%s{{/}}\n", indent, pc.CurrentStep.CodeLocation)
	}

	out.WriteString("\n{{bold}}{{underline}}Spec Goroutine{{/}}\n")
	out.WriteString(pc.SpecGoroutine.Report(color))

	if len(pc.HighlightedGoroutines) > 0 {
		out.WriteString("\n{{bold}}{{underline}}Goroutines of Interest{{/}}\n")
		for _, goroutine := range pc.HighlightedGoroutines {
			out.WriteString(goroutine.Report(color))
			out.WriteString("\n")
		}
	}

	if includeAllGoroutines && len(pc.OtherGoroutines) > 0 {
		out.WriteString("\n{{gray}}{{bold}}{{underline}}Other Goroutines{{/}}\n")
		for _, goroutine := range pc.OtherGoroutines {
			out.WriteString(goroutine.Report(color))
			out.WriteString("\n")
		}
	}

	return out.String()
}

func NewProgressReport(report types.SpecReport, currentNode Node, currentNodeStartTime time.Time, currentStep ProgressStepCursor) (ProgressReport, error) {
	pc := ProgressReport{
		CurrentSpecReport:    report,
		CurrentNode:          currentNode,
		CurrentNodeStartTime: currentNodeStartTime,
		CurrentStep:          currentStep,
	}

	goroutines, err := extractRunningGoroutines()
	if err != nil {
		return pc, err
	}

	// now we want to try to find goroutines of interest.  these will be goroutines that have any function calls with code in packagesOfInterest:
	packagesOfInterest := map[string]bool{}
	addPackageFor := func(filename string) {
		if filename != "" {
			packagesOfInterest[packageFromFilename(filename)] = true
		}
	}
	for _, location := range report.ContainerHierarchyLocations {
		addPackageFor(location.FileName)
	}
	addPackageFor(report.LeafNodeLocation.FileName)
	addPackageFor(currentNode.CodeLocation.FileName)
	addPackageFor(currentStep.CodeLocation.FileName)

	//First, we find the SpecGoroutine - this will be the goroutine that includes `runNode`
	specGoRoutineIdx := -1
	runNodeFunctionCallIdx := -1
OUTER:
	for goroutineIdx, goroutine := range goroutines {
		for functionCallIdx, functionCall := range goroutine.Stack {
			if strings.Contains(functionCall.Function, "ginkgo/v2/internal.(*Suite).runNode.func") {
				specGoRoutineIdx = goroutineIdx
				runNodeFunctionCallIdx = functionCallIdx
				break OUTER
			}
		}
	}

	//Now, we find the first non-Ginkgo function call
	if specGoRoutineIdx > -1 {
		for runNodeFunctionCallIdx >= 0 {
			if strings.Contains(goroutines[specGoRoutineIdx].Stack[runNodeFunctionCallIdx].Function, "ginkgo/v2/internal") {
				runNodeFunctionCallIdx--
				continue
			}
			//found it!  lets add its package of interest
			addPackageFor(goroutines[specGoRoutineIdx].Stack[runNodeFunctionCallIdx].Filename)
			break
		}
	}

	// Now we go through all goroutines and highlight any lines with packages in `packagesOfInterest`
	// Any goroutines with highlighted lines end up in the HighlightGoRoutines
	for goroutineIdx, goroutine := range goroutines {
		hasHighlights := false
		isGinkgoEntryPoint := false
		for functionCallIdx, functionCall := range goroutine.Stack {
			if strings.Contains(functionCall.Function, "ginkgo/v2.RunSpecs") {
				isGinkgoEntryPoint = true
				break
			}
			if packagesOfInterest[packageFromFilename(functionCall.Filename)] {
				goroutine.Stack[functionCallIdx].Highlight = true
				hasHighlights = true
			}
		}
		if goroutineIdx == specGoRoutineIdx {
			pc.SpecGoroutine = goroutine
		} else if hasHighlights && !isGinkgoEntryPoint {
			pc.HighlightedGoroutines = append(pc.HighlightedGoroutines, goroutine)
		} else {
			pc.OtherGoroutines = append(pc.OtherGoroutines, goroutine)
		}
	}

	return pc, nil
}

func extractRunningGoroutines() (Goroutines, error) {
	var stack []byte
	for size := 64 * 1024; ; size *= 2 {
		stack = make([]byte, size)
		if n := runtime.Stack(stack, true); n < size {
			stack = stack[:n]
			break
		}
	}

	r := bufio.NewReader(bytes.NewReader(stack))
	out := Goroutines{}
	idx := -1
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}

		line = strings.TrimSuffix(line, "\n")

		//skip blank lines
		if line == "" {
			continue
		}

		//parse headers for new goroutine frames
		if strings.HasPrefix(line, "goroutine") {
			out = append(out, Goroutine{})
			idx = len(out) - 1

			line = strings.TrimPrefix(line, "goroutine ")
			line = strings.TrimSuffix(line, ":")
			fields := strings.SplitN(line, " ", 2)
			if len(fields) != 2 {
				return nil, types.GinkgoErrors.FailedToParseStackTrace(fmt.Sprintf("Invalid goroutine frame header: %s", line))
			}
			out[idx].ID, err = strconv.ParseUint(fields[0], 10, 64)
			if err != nil {
				return nil, types.GinkgoErrors.FailedToParseStackTrace(fmt.Sprintf("Invalid goroutine ID: %s", fields[1]))
			}

			out[idx].State = strings.TrimSuffix(strings.TrimPrefix(fields[1], "["), "]")
			continue
		}

		//if we are here we must be at a function call entry in the stack
		functionCall := FunctionCall{
			Function: strings.TrimPrefix(line, "created by "), // no need to track 'created by'
		}

		line, err = r.ReadString('\n')
		line = strings.TrimSuffix(line, "\n")
		if err == io.EOF {
			return nil, types.GinkgoErrors.FailedToParseStackTrace(fmt.Sprintf("Invalid function call: %s -- missing file name and line number", functionCall.Function))
		}
		line = strings.TrimLeft(line, " \t")
		fields := strings.SplitN(line, ":", 2)
		if len(fields) != 2 {
			return nil, types.GinkgoErrors.FailedToParseStackTrace(fmt.Sprintf("Invalid filename nad line number: %s", line))
		}
		functionCall.Filename = fields[0]
		line = strings.Split(fields[1], " ")[0]
		functionCall.Line, err = strconv.ParseInt(line, 10, 64)
		if err != nil {
			return nil, types.GinkgoErrors.FailedToParseStackTrace(fmt.Sprintf("Invalid function call line number: %s\n%s", line, err.Error()))
		}
		out[idx].Stack = append(out[idx].Stack, functionCall)
	}

	return out, nil
}

var _SOURCE_CACHE = map[string][]string{}

func sourceAt(filename string, lineNumber int, span int, color string) string {
	if filename == "" {
		return ""
	}
	var lines []string
	var ok bool
	if lines, ok = _SOURCE_CACHE[filename]; !ok {
		data, err := os.ReadFile(filename)
		if err != nil {
			return ""
		}
		lines = strings.Split(string(data), "\n")
		_SOURCE_CACHE[filename] = lines
	}
	idx := lineNumber - span - 1
	out := []string{}
	for idx < lineNumber+span {
		if idx >= 0 && idx <= len(lines)-1 {
			if idx == lineNumber-1 {
				out = append(out, "      {{bold}}"+color+"> "+lines[idx]+"{{/}}")
			} else {
				out = append(out, "      | "+lines[idx])
			}
		}
		idx++
	}
	return strings.Join(out, "\n")
}

func packageFromFilename(fname string) string {
	return filepath.Dir(fname)
}

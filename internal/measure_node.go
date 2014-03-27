package internal

import (
	"github.com/onsi/ginkgo/internal/codelocation"
	internaltypes "github.com/onsi/ginkgo/internal/types"
	"github.com/onsi/ginkgo/types"
	"reflect"
)

type measureNode struct {
	text         string
	body         func(internaltypes.Benchmarker)
	flag         FlagType
	codeLocation types.CodeLocation
	samples      int
	benchmarker  *benchmarker
}

func newMeasureNode(text string, body interface{}, flag FlagType, codeLocation types.CodeLocation, samples int) *measureNode {
	bodyValue := reflect.ValueOf(body)
	wrappedBody := func(b internaltypes.Benchmarker) {
		bodyValue.Call([]reflect.Value{reflect.ValueOf(b)})
	}

	return &measureNode{
		text:         text,
		body:         wrappedBody,
		flag:         flag,
		codeLocation: codeLocation,
		samples:      samples,
		benchmarker:  newBenchmarker(),
	}
}

func (node *measureNode) run() (outcome runOutcome, failure failureData) {
	defer func() {
		if e := recover(); e != nil {
			outcome = runOutcomePanicked
			failure = failureData{
				message:        "Test Panicked",
				codeLocation:   codelocation.New(2),
				forwardedPanic: e,
			}
		}
	}()

	node.body(node.benchmarker)
	outcome = runOutcomeCompleted

	return
}

func (node *measureNode) measurementsReport() map[string]*types.ExampleMeasurement {
	return node.benchmarker.measurementsReport()
}

func (node *measureNode) nodeType() nodeType {
	return nodeTypeMeasure
}

func (node *measureNode) getText() string {
	return node.text
}

func (node *measureNode) getFlag() FlagType {
	return node.flag
}

func (node *measureNode) getCodeLocation() types.CodeLocation {
	return node.codeLocation
}

package measurenode

import (
	"github.com/onsi/ginkgo/internal/failer"
	"github.com/onsi/ginkgo/internal/types"
	"github.com/onsi/ginkgo/types"
	"reflect"
)

type MeasureNode struct {
	text           string
	body           func(*benchmarker)
	flag           internaltypes.FlagType
	codeLocation   types.CodeLocation
	componentIndex int
	samples        int
	benchmarker    *benchmarker
	failer         *failer.Failer
}

func New(text string, body interface{}, flag internaltypes.FlagType, codeLocation types.CodeLocation, samples int, failer *failer.Failer, componentIndex int) *MeasureNode {
	bodyValue := reflect.ValueOf(body)
	wrappedBody := func(b *benchmarker) {
		bodyValue.Call([]reflect.Value{reflect.ValueOf(b)})
	}

	return &MeasureNode{
		text:           text,
		body:           wrappedBody,
		flag:           flag,
		codeLocation:   codeLocation,
		samples:        samples,
		benchmarker:    newBenchmarker(),
		failer:         failer,
		componentIndex: componentIndex,
	}
}

func (node *MeasureNode) Run() (outcome types.ExampleState, failure types.ExampleFailure) {
	defer func() {
		if e := recover(); e != nil {
			node.failer.Panic(node.codeLocation, e)
		}

		failure, outcome = node.failer.Drain(types.ExampleComponentTypeMeasure, node.componentIndex, node.codeLocation)
	}()

	node.body(node.benchmarker)

	return
}

func (node *MeasureNode) MeasurementsReport() map[string]*types.ExampleMeasurement {
	return node.benchmarker.measurementsReport()
}

func (node *MeasureNode) Type() types.ExampleComponentType {
	return types.ExampleComponentTypeMeasure
}

func (node *MeasureNode) Text() string {
	return node.text
}

func (node *MeasureNode) Flag() internaltypes.FlagType {
	return node.flag
}

func (node *MeasureNode) CodeLocation() types.CodeLocation {
	return node.codeLocation
}

func (node *MeasureNode) Samples() int {
	return node.samples
}

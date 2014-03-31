package leafnodes

import (
	"github.com/onsi/ginkgo/internal/failer"
	"github.com/onsi/ginkgo/types"
	"reflect"
)

type MeasureNode struct {
	runner *runner

	text        string
	flag        types.FlagType
	samples     int
	benchmarker *benchmarker
}

func NewMeasureNode(text string, body interface{}, flag types.FlagType, codeLocation types.CodeLocation, samples int, failer *failer.Failer, componentIndex int) *MeasureNode {
	benchmarker := newBenchmarker()

	wrappedBody := func() {
		reflect.ValueOf(body).Call([]reflect.Value{reflect.ValueOf(benchmarker)})
	}

	return &MeasureNode{
		runner: newRunner(wrappedBody, codeLocation, 0, failer, types.ExampleComponentTypeMeasure, componentIndex),

		text:        text,
		flag:        flag,
		samples:     samples,
		benchmarker: benchmarker,
	}
}

func (node *MeasureNode) Run() (outcome types.ExampleState, failure types.ExampleFailure) {
	return node.runner.run()
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

func (node *MeasureNode) Flag() types.FlagType {
	return node.flag
}

func (node *MeasureNode) CodeLocation() types.CodeLocation {
	return node.runner.codeLocation
}

func (node *MeasureNode) Samples() int {
	return node.samples
}

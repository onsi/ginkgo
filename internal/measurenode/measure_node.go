package measurenode

import (
	"github.com/onsi/ginkgo/internal/codelocation"
	"github.com/onsi/ginkgo/internal/types"
	"github.com/onsi/ginkgo/types"
	"reflect"
)

type MeasureNode struct {
	text         string
	body         func(*benchmarker)
	flag         internaltypes.FlagType
	codeLocation types.CodeLocation
	samples      int
	benchmarker  *benchmarker
}

func New(text string, body interface{}, flag internaltypes.FlagType, codeLocation types.CodeLocation, samples int) *MeasureNode {
	bodyValue := reflect.ValueOf(body)
	wrappedBody := func(b *benchmarker) {
		bodyValue.Call([]reflect.Value{reflect.ValueOf(b)})
	}

	return &MeasureNode{
		text:         text,
		body:         wrappedBody,
		flag:         flag,
		codeLocation: codeLocation,
		samples:      samples,
		benchmarker:  newBenchmarker(),
	}
}

func (node *MeasureNode) Run() (outcome internaltypes.Outcome, failure internaltypes.FailureData) {
	defer func() {
		if e := recover(); e != nil {
			outcome = internaltypes.OutcomePanicked
			failure = internaltypes.FailureData{
				Message:        "Test Panicked",
				CodeLocation:   codelocation.New(2),
				ForwardedPanic: e,
			}
		}
	}()

	node.body(node.benchmarker)
	outcome = internaltypes.OutcomeCompleted

	return
}

func (node *MeasureNode) MeasurementsReport() map[string]*types.ExampleMeasurement {
	return node.benchmarker.measurementsReport()
}

func (node *MeasureNode) Type() internaltypes.NodeType {
	return internaltypes.NodeTypeMeasure
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

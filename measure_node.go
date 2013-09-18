package ginkgo

type measureNode struct {
	text         string
	body         func(Benchmarker)
	flag         flagType
	codeLocation CodeLocation
	samples      int
	benchmarker  *benchmarker
}

func newMeasureNode(text string, body func(Benchmarker), flag flagType, codeLocation CodeLocation, samples int) *measureNode {
	return &measureNode{
		text:         text,
		body:         body,
		flag:         flag,
		codeLocation: codeLocation,
		samples:      samples,
		benchmarker:  newBenchmarker(),
	}
}

func (node *measureNode) run() (outcome runOutcome, failure failureData) {
	node.body(node.benchmarker)

	return runOutcomeCompleted, failureData{}
}

func (node *measureNode) measurementsReport() map[string]*ExampleMeasurement {
	return node.benchmarker.measurementsReport()
}

func (node *measureNode) nodeType() nodeType {
	return nodeTypeMeasure
}

func (node *measureNode) getText() string {
	return node.text
}

func (node *measureNode) getFlag() flagType {
	return node.flag
}

func (node *measureNode) getCodeLocation() CodeLocation {
	return node.codeLocation
}

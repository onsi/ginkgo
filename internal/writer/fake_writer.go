package writer

type FakeGinkgoWriter struct {
	EventStream []string
}

func NewFake() *FakeGinkgoWriter {
	return &FakeGinkgoWriter{
		EventStream: []string{},
	}
}

func (writer *FakeGinkgoWriter) AddEvent(event string) {
	writer.EventStream = append(writer.EventStream, event)
}

func (writer *FakeGinkgoWriter) Truncate() {
	writer.EventStream = append(writer.EventStream, "TRUNCATE")
}

func (writer *FakeGinkgoWriter) DumpOut() {
	writer.EventStream = append(writer.EventStream, "DUMP")
}

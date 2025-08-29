package reporters

type GoJSONEventWriter struct {
	enc encoder
}

func NewGoJSONEventWriter(enc encoder) *GoJSONEventWriter {
	return &GoJSONEventWriter{
		enc: enc,
	}
}

func (r *GoJSONEventWriter) writeEvent(e *test2jsonEvent) error {
	return r.enc.Encode(e)
}

func (r *GoJSONEventWriter) WriteSuiteStart(report *report) error {
	e := &test2jsonEvent{
		Time:        &report.o.StartTime,
		Action:      Test2JSONStart,
		Package:     report.goPkg,
		Output:      nil,
		FailedBuild: "",
	}
	return r.writeEvent(e)
}

func (r *GoJSONEventWriter) WriteSuiteResult(report *report) error {
	var action Test2JSONAction
	switch {
	case report.o.PreRunStats.SpecsThatWillRun == 0:
		action = Test2JSONSkip
	case report.o.SuiteSucceeded:
		action = Test2JSONPass
	default:
		action = Test2JSONFail
	}
	e := &test2jsonEvent{
		Time:        &report.o.EndTime,
		Action:      action,
		Package:     report.goPkg,
		Output:      nil,
		FailedBuild: "",
		Elapsed:     ptr(report.elapsed),
	}
	return r.writeEvent(e)
}

func (r *GoJSONEventWriter) WriteSuiteLeafNodeOut(report *report, specReport *specReport) error {
	events := []*test2jsonEvent{}

	combinedOutput := specReport.o.CombinedOutput()
	if combinedOutput != "" {
		events = append(events, &test2jsonEvent{
			Time:        &specReport.o.EndTime,
			Action:      Test2JSONOutput,
			Package:     report.goPkg,
			Output:      ptr(combinedOutput),
			FailedBuild: "",
		})
	}

	for _, ev := range events {
		err := r.writeEvent(ev)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *GoJSONEventWriter) WriteSpecStart(report *report, specReport *specReport) error {
	e := &test2jsonEvent{
		Time:        &specReport.o.StartTime,
		Action:      Test2JSONRun,
		Test:        specReport.testName,
		Package:     report.goPkg,
		Output:      nil,
		FailedBuild: "",
	}
	return r.writeEvent(e)
}

func (r *GoJSONEventWriter) WriteSpecOut(report *report, specReport *specReport) error {
	events := []*test2jsonEvent{}
	combinedOutput := specReport.o.CombinedOutput()
	if combinedOutput != "" {
		events = append(events, &test2jsonEvent{
			Time:        &specReport.o.EndTime,
			Action:      Test2JSONOutput,
			Test:        specReport.testName,
			Package:     report.goPkg,
			Output:      ptr(combinedOutput),
			FailedBuild: "",
		})
	}
	if specReport.o.Failure.Message != "" {
		events = append(events, &test2jsonEvent{
			Time:        &specReport.o.EndTime,
			Action:      Test2JSONOutput,
			Test:        specReport.testName,
			Package:     report.goPkg,
			Output:      ptr(failureToOutput(specReport.o.Failure)),
			FailedBuild: "",
		})
	}
	for _, ev := range events {
		err := r.writeEvent(ev)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *GoJSONEventWriter) WriteSpecResult(report *report, specReport *specReport) error {
	e := &test2jsonEvent{
		Time:        &specReport.o.EndTime,
		Action:      specReport.action,
		Test:        specReport.testName,
		Package:     report.goPkg,
		Elapsed:     ptr(specReport.elapsed),
		Output:      nil,
		FailedBuild: "",
	}
	return r.writeEvent(e)
}

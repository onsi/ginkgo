package reporters

type GoJSONEventWriter struct {
	enc encoder
}

func NewGoJSONEventWriter(enc encoder) *GoJSONEventWriter {
	return &GoJSONEventWriter{
		enc: enc,
	}
}

func (r *GoJSONEventWriter) writeEvent(e *gojsonEvent) error {
	return r.enc.Encode(e)
}

func (r *GoJSONEventWriter) WriteSuiteStart(report *report) error {
	e := &gojsonEvent{
		Time:        &report.o.StartTime,
		Action:      GoJSONStart,
		Package:     report.goPkg,
		Output:      nil,
		FailedBuild: "",
	}
	return r.writeEvent(e)
}

func (r *GoJSONEventWriter) WriteSuiteResult(report *report) error {
	var action GoJSONAction
	switch {
	case report.o.PreRunStats.SpecsThatWillRun == 0:
		action = GoJSONSkip
	case report.o.SuiteSucceeded:
		action = GoJSONPass
	default:
		action = GoJSONFail
	}
	e := &gojsonEvent{
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
	events := []*gojsonEvent{}

	combinedOutput := specReport.o.CombinedOutput()
	if combinedOutput != "" {
		events = append(events, &gojsonEvent{
			Time:        &specReport.o.EndTime,
			Action:      GoJSONOutput,
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
	e := &gojsonEvent{
		Time:        &specReport.o.StartTime,
		Action:      GoJSONRun,
		Test:        specReport.testName,
		Package:     report.goPkg,
		Output:      nil,
		FailedBuild: "",
	}
	return r.writeEvent(e)
}

func (r *GoJSONEventWriter) WriteSpecOut(report *report, specReport *specReport) error {
	events := []*gojsonEvent{}
	combinedOutput := specReport.o.CombinedOutput()
	if combinedOutput != "" {
		events = append(events, &gojsonEvent{
			Time:        &specReport.o.EndTime,
			Action:      GoJSONOutput,
			Test:        specReport.testName,
			Package:     report.goPkg,
			Output:      ptr(combinedOutput),
			FailedBuild: "",
		})
	}
	if specReport.o.Failure.Message != "" {
		events = append(events, &gojsonEvent{
			Time:        &specReport.o.EndTime,
			Action:      GoJSONOutput,
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
	e := &gojsonEvent{
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

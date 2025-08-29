package reporters

import (
	"github.com/onsi/ginkgo/v2/types"
)

type GoJSONReporter struct {
	ev *GoJSONEventWriter
}


func NewGoJSONReporter(enc encoder) *GoJSONReporter {
	return &GoJSONReporter{
		ev: NewGoJSONEventWriter(enc),
	}
}

func (r *GoJSONReporter) Write(originalReport types.Report) error {
	// suite start events
	report := newReport(originalReport)
	err := report.Fill()
	if err != nil {
		return err
	}
	r.ev.WriteSuiteStart(report)
	for _, originalSpecReport := range originalReport.SpecReports {
		specReport := newSpecReport(originalSpecReport)
		err := specReport.Fill()
		if err != nil {
			return err
		}
		if specReport.o.LeafNodeType == types.NodeTypeIt {
			// handle any It leaf node as a spec
			r.ev.WriteSpecStart(report, specReport)
			r.ev.WriteSpecOut(report, specReport)
			r.ev.WriteSpecResult(report, specReport)
		} else {
			// handle any other leaf node as generic output
			r.ev.WriteSuiteLeafNodeOut(report, specReport)
		}
	}
	r.ev.WriteSuiteResult(report)
	return nil
}

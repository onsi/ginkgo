package reporters

import (
	"errors"

	"github.com/onsi/ginkgo/v2/types"
	"golang.org/x/tools/go/packages"
)

// types.Report
type report struct {
	o types.Report
	// Extra calculated fields
	goPkg string
	elapsed float64
}

func newReport(in types.Report) *report {
	return &report{
		o: in,
	}
}

func (r *report) Fill() error {
	// NOTE: could the types.Report include the go package name?
	goPkg, err := suitePathToPkg(r.o.SuitePath)
	if err != nil {
		return err
	}
	r.goPkg = goPkg
	r.elapsed = r.o.RunTime.Seconds()
	return nil
}

// types.SpecReport
type specReport struct {
	o types.SpecReport
	// extra calculated fields
	testName string
	action GoJSONAction
	elapsed float64
}

func newSpecReport(in types.SpecReport) *specReport {
	return &specReport{
		o: in,
	}
}

func (sr *specReport) Fill() error {
	sr.elapsed = sr.o.RunTime.Seconds()
	sr.testName = sr.o.FullText()
	sr.action = specStateToAction(sr.o.State)
	return nil
}


func suitePathToPkg(dir string) (string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedFiles | packages.NeedSyntax,
	}
	pkgs, err := packages.Load(cfg, dir)
	if err != nil {
		return "", err
	}
	if len(pkgs) != 1 {
		return "", errors.New("error")
	}
	return pkgs[0].ID, nil
}

func specStateToAction(state types.SpecState) GoJSONAction {
	switch state {
	case types.SpecStateInvalid:
		return GoJSONFail
	case types.SpecStatePending:
		return GoJSONSkip
	case types.SpecStateSkipped:
		return GoJSONSkip
	case types.SpecStatePassed:
		return GoJSONPass
	case types.SpecStateFailed:
		return GoJSONFail
	case types.SpecStateAborted:
		return GoJSONFail
	case types.SpecStatePanicked:
		return GoJSONFail
	case types.SpecStateInterrupted:
		return GoJSONFail
	case types.SpecStateTimedout:
		return GoJSONFail
	default:
		panic("unexpected state should not happen")
	}
}

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
	action Test2JSONAction
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

func specStateToAction(state types.SpecState) Test2JSONAction {
	switch state {
	case types.SpecStateInvalid:
		return Test2JSONFail
	case types.SpecStatePending:
		return Test2JSONSkip
	case types.SpecStateSkipped:
		return Test2JSONSkip
	case types.SpecStatePassed:
		return Test2JSONPass
	case types.SpecStateFailed:
		return Test2JSONFail
	case types.SpecStateAborted:
		return Test2JSONFail
	case types.SpecStatePanicked:
		return Test2JSONFail
	case types.SpecStateInterrupted:
		return Test2JSONFail
	case types.SpecStateTimedout:
		return Test2JSONFail
	default:
		panic("unexpected state should not happen")
	}
}

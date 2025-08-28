package reporters

import "github.com/onsi/ginkgo/v2/types"

func failureToOutput(failure types.Failure) string {
	return failure.Message
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

func testNameFromSpecReport(specReport types.SpecReport) (string, error) {
	return specReport.FullText(), nil
}

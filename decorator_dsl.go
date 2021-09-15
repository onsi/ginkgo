package ginkgo

import (
	"github.com/onsi/ginkgo/internal"
)

//Offset(uint) is a decorator that allows you to change the stack-frame offset used when computing the line number of the node in question.
type Offset = internal.Offset

//FlakeAttempts(uint N) is a decorator that allows you to mark individual tests or test containers as flaky.  Ginkgo will run them up to `N` times until they pass.
type FlakeAttempts = internal.FlakeAttempts

//Focus is a decorator that allows you to mark a test or container as focused.  Identical to FIt and FDescribe.
const Focus = internal.Focus

//Pending is a decorator that allows you to mark a test or container as pending.  Identical to PIt and PDescribe.
const Pending = internal.Pending

//Serial is a decorator that allows you to mark a test or container as serial.  These tests will never run in parallel with other tests.
const Serial = internal.Serial

//Label decorates specs with Labels.  Multiple labels can be passed to Label and these can be arbitrary strings but must not include the following characters: "&|!,()/".
//Labels can be pplied to container and test nodes, but not setup nodes.  You can provide multiple Labels to a given node and a spec's labels is the union of all labels in its node hierarchy.
func Label(labels ...string) Labels {
	return Labels(labels)
}

//Labels are the type for spec Label decorations.  Use Label(...) to construct Labels.
type Labels = internal.Labels

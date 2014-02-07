/*
The gomocktestreporter package provides a Ginkgo friendly implementation of [Gomock's](https://code.google.com/p/gomock/) `TestReporter` interface.

More details and a code example are [here](http://onsi.github.io/ginkgo/#integrating_with_gomock).
*/
package gomocktestreporter

import (
	"fmt"
	"github.com/onsi/ginkgo"
)

type GomockTestReporter struct{}

func New() GomockTestReporter {
	return GomockTestReporter{}
}

func (g GomockTestReporter) Errorf(format string, args ...interface{}) {
	ginkgo.Fail(fmt.Sprintf(format, args...), 3)
}

func (g GomockTestReporter) Fatalf(format string, args ...interface{}) {
	ginkgo.Fail(fmt.Sprintf(format, args...), 3)
}

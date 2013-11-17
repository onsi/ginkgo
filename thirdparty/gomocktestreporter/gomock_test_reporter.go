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
	ginkgo.Fail(fmt.Sprintf(format, args), 3)
}

func (g GomockTestReporter) Fatalf(format string, args ...interface{}) {
	ginkgo.Fail(fmt.Sprintf(format, args), 3)
}

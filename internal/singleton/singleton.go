package singleton

import (
	"github.com/onsi/ginkgo/internal/failer"
	"github.com/onsi/ginkgo/internal/suite"
)

var GlobalSuite *suite.Suite
var GlobalFailer *failer.Failer

func init() {
	GlobalFailer = failer.New()
	GlobalSuite = suite.New(GlobalFailer)
}

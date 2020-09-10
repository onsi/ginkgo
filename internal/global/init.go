package global

import (
	"github.com/onsi/ginkgo/internal"
)

var Suite *internal.Suite
var Failer *internal.Failer

func init() {
	InitializeGlobals()
}

func InitializeGlobals() {
	Failer = internal.NewFailer()
	Suite = internal.NewSuite()
}

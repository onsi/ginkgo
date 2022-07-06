package reporters_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/reporters"
	. "github.com/onsi/gomega"
)

var _ = Describe("JUnitTestCase", func() {
	It("JUnitTestCase is initialized", func() {
		report := reporters.JUnitTestCase{
			Name:      "name",
			Timestamp: "2006-01-02T15:04:05",
			Classname: "classname",
			Status:    "passed",
			Time:      17,
			SystemOut: "gw",
			SystemErr: "gw",
		}

		Ω(report.Name).Should(Equal(string("name")))
		Ω(report.Timestamp).Should(Equal(string("2006-01-02T15:04:05")))
		Ω(report.Classname).Should(Equal(string("classname")))
		Ω(report.Status).Should(Equal(string("passed")))
		Ω(report.Time).Should(Equal(17.0))
	})
})

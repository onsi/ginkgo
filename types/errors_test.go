package types_test

import (
	"reflect"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/onsi/ginkgo/formatter"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("GinkgoErrors", func() {
	BeforeEach(func() {
		formatter.SingletonFormatter.ColorMode = formatter.ColorModePassthrough
	})

	AfterEach(func() {
		formatter.SingletonFormatter.ColorMode = formatter.ColorModeTerminal
	})

	DescribeTable("error render cases", func(err error, expected ...string) {
		Expect(err.Error()).To(HavePrefix(strings.Join(expected, "\n")))
	},
		Entry("an error with only a heading",
			types.GinkgoError{
				Heading: "Error! Error!",
			},
			"{{bold}}{{red}}Error! Error!{{/}}",
			"",
		),
		Entry("an error with all the things",
			types.GinkgoError{
				Heading:      "Error! Error!",
				CodeLocation: types.CodeLocation{FileName: "foo.go", LineNumber: 17},
				Message:      "An error occured.\nWelp!",
				DocLink:      "the-doc-section",
			},
			"{{bold}}{{red}}Error! Error!{{/}}",
			"{{gray}}foo.go:17{{/}}",
			"  An error occured.",
			"  Welp!",
			"",
			"  {{bold}}Learn more at:{{/}}",
			"  {{cyan}}{{underline}}http://onsi.github.io/ginkgo/#the-doc-section{{/}}",
		),
	)

	It("validates that all errors point to working documentation", func() {
		v := reflect.ValueOf(types.GinkgoErrors)
		Ω(v.NumMethod()).Should(BeNumerically(">", 0))
		for i := 0; i < v.NumMethod(); i += 1 {
			m := v.Method(i)
			args := []reflect.Value{}
			for j := 0; j < m.Type().NumIn(); j += 1 {
				args = append(args, reflect.Zero(m.Type().In(j)))
			}

			ginkgoError := m.Call(args)[0].Interface().(types.GinkgoError)

			if ginkgoError.DocLink != "" {
				Ω(ginkgoError.DocLink).Should(BeElementOf(DOC_ANCHORS))
			}
		}
	})
})

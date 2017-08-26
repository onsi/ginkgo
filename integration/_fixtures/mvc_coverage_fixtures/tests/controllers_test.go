package controllers_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/integration/_fixtures/mvc_coverage_fixtures/controllers"
	. "github.com/onsi/gomega"
)

var _ = Describe("MvcCoverageTests", func() {

	It("Should test GET", func() {
		var controller = controllers.MvcController{}

		Expect(controller.Get()).To(Equal(`{"1":"1"","2":"2"","3":"3"","4":"4"","5":"5"","6":"6"","7":"7"","8":"8"","9":"9""}`))
	})

	It("Should test POST", func() {
		var controller = controllers.MvcController{}

		Expect(controller.Post()).To(Equal("362880"))
	})
})

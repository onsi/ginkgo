package controllers_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/integration/_fixtures/mvc_coverage_fixtures/models"
	. "github.com/onsi/gomega"
)

var _ = Describe("MvcCoverageTests", func() {

	It("Should test Name", func() {
		var model = models.Model{}

		Expect(model.Name()).To(Equal(`I'm a model`))
	})
})

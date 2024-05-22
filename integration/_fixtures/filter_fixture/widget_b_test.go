package filter_fixture_test

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("WidgetB", func() {
	It("cat", func() {

	})

	It("dog", Label("slow"), func() {

	})

	It("fish", Label("Feature:Alpha"), func() {

	})

	It("cat fish", func() {

	})

	It("dog fish", Label("Feature:Beta"), func() {

	})
})

var _ = Describe("More WidgetB", func() {
	It("cat", func() {

	})

	It("dog", func() {

	})

	It("cat fish", func() {

	})

	It("dog fish", func() {

	})
})

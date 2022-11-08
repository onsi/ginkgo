package internal_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProgressReport", func() {
	Describe("The goroutine stack", func() {
		It("is better tested in the internal integration tests because this test package lives in internal which is a key part of the logic for how the goroutine stack is analyzed...", func() {
			//empty
		})

	})

	Context("when includeAll is false", func() {
		It("does not include any other goroutines", func() {
			pr, err := internal.NewProgressReport(false, types.SpecReport{}, Node{}, time.Now(), types.SpecEvent{}, "", types.TimelineLocation{}, []string{}, []string{}, false)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(pr.OtherGoroutines()).Should(HaveLen(0))
		})
	})

	Context("when includeAll is true", func() {
		It("includes all other goroutines", func() {
			pr, err := internal.NewProgressReport(false, types.SpecReport{}, Node{}, time.Now(), types.SpecEvent{}, "", types.TimelineLocation{}, []string{}, []string{}, true)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(pr.OtherGoroutines()).ShouldNot(HaveLen(0))
		})
	})
})

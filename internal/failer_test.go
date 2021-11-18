package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("Failer", func() {
	var failer *internal.Failer
	var clA types.CodeLocation
	var clB types.CodeLocation

	BeforeEach(func() {
		clA = CL("file_a.go")
		clB = CL("file_b.go")
		failer = internal.NewFailer()
	})

	Context("with no failures", func() {
		It("should return success when drained", func() {
			state, failure := failer.Drain()
			Ω(state).Should(Equal(types.SpecStatePassed))
			Ω(failure).Should(BeZero())
		})
	})

	Describe("when told of a failure", func() {
		BeforeEach(func() {
			failer.Fail("something failed", clA)
		})

		It("should record the failure", func() {
			state, failure := failer.Drain()
			Ω(state).Should(Equal(types.SpecStateFailed))
			Ω(failure).Should(Equal(types.Failure{
				Message:  "something failed",
				Location: clA,
			}))
		})

		Context("when told of anotehr failure", func() {
			It("discards the second failure, preserving the original", func() {
				failer.Fail("something else failed", clB)

				state, failure := failer.Drain()
				Ω(state).Should(Equal(types.SpecStateFailed))
				Ω(failure).Should(Equal(types.Failure{
					Message:  "something failed",
					Location: clA,
				}))
			})
		})
	})

	Describe("when told to skip", func() {
		Context("when no failure has occured", func() {
			It("registers the test as skipped", func() {
				failer.Skip("something skipped", clA)
				state, failure := failer.Drain()
				Ω(state).Should(Equal(types.SpecStateSkipped))
				Ω(failure).Should(Equal(types.Failure{
					Message:  "something skipped",
					Location: clA,
				}))
			})
		})

		Context("when a failure has already occured", func() {
			BeforeEach(func() {
				failer.Fail("something failed", clA)
			})

			It("does not modify the failure", func() {
				failer.Skip("something skipped", clB)
				state, failure := failer.Drain()
				Ω(state).Should(Equal(types.SpecStateFailed))
				Ω(failure).Should(Equal(types.Failure{
					Message:  "something failed",
					Location: clA,
				}))
			})
		})
	})

	Describe("when told to abort", func() {
		Context("when no failure has occured", func() {
			It("registers the test as aborted", func() {
				failer.AbortSuite("something aborted", clA)
				state, failure := failer.Drain()
				Ω(state).Should(Equal(types.SpecStateAborted))
				Ω(failure).Should(Equal(types.Failure{
					Message:  "something aborted",
					Location: clA,
				}))
			})
		})

		Context("when a failure has already occured", func() {
			BeforeEach(func() {
				failer.Fail("something failed", clA)
			})

			It("does not modify the failure", func() {
				failer.AbortSuite("something aborted", clA)
				state, failure := failer.Drain()
				Ω(state).Should(Equal(types.SpecStateFailed))
				Ω(failure).Should(Equal(types.Failure{
					Message:  "something failed",
					Location: clA,
				}))
			})
		})
	})

	Describe("when told to panic", func() {
		BeforeEach(func() {
			failer.Panic(clA, 17)
		})

		It("should record the panic", func() {
			state, failure := failer.Drain()
			Ω(state).Should(Equal(types.SpecStatePanicked))
			Ω(failure).Should(Equal(types.Failure{
				Message:        "Test Panicked",
				Location:       clA,
				ForwardedPanic: "17",
			}))
		})

		Context("when told of another panic", func() {
			It("discards the second panic, preserving the original", func() {
				failer.Panic(clB, 23)

				state, failure := failer.Drain()
				Ω(state).Should(Equal(types.SpecStatePanicked))
				Ω(failure).Should(Equal(types.Failure{
					Message:        "Test Panicked",
					Location:       clA,
					ForwardedPanic: "17",
				}))
			})
		})
	})

	Context("when drained", func() {
		BeforeEach(func() {
			failer.Fail("something failed", clA)
			state, _ := failer.Drain()
			Ω(state).Should(Equal(types.SpecStateFailed))
		})

		It("resets the failer such that subsequent drains pass", func() {
			state, failure := failer.Drain()
			Ω(state).Should(Equal(types.SpecStatePassed))
			Ω(failure).Should(BeZero())
		})

		It("allows subsequent failures to be recorded", func() {
			failer.Fail("something else failed", clB)
			state, failure := failer.Drain()
			Ω(state).Should(Equal(types.SpecStateFailed))
			Ω(failure).Should(Equal(types.Failure{
				Message:  "something else failed",
				Location: clB,
			}))
		})
	})
})

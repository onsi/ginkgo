package command_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/ginkgo/command"
)

var _ = Describe("Abort", func() {
	It("panics when called", func() {
		details := command.AbortDetails{
			ExitCode:  1,
			Error:     fmt.Errorf("boom"),
			EmitUsage: true,
		}
		Ω(func() {
			command.Abort(details)
		}).Should(PanicWith(details))
	})

	Describe("AbortWith", func() {
		It("aborts with a formatted error", func() {
			Ω(func() {
				command.AbortWith("boom %d %s", 17, "bam!")
			}).Should(PanicWith(command.AbortDetails{
				ExitCode:  1,
				Error:     fmt.Errorf("boom 17 bam!"),
				EmitUsage: false,
			}))
		})
	})

	Describe("AbortWithUsage", func() {
		It("aborts with a formatted error and sets usage to true", func() {
			Ω(func() {
				command.AbortWithUsage("boom %d %s", 17, "bam!")
			}).Should(PanicWith(command.AbortDetails{
				ExitCode:  1,
				Error:     fmt.Errorf("boom 17 bam!"),
				EmitUsage: true,
			}))
		})
	})

	Describe("AbortIfError", func() {
		Context("with a nil error", func() {
			It("does not abort", func() {
				Ω(func() {
					command.AbortIfError("boom boom?", nil)
				}).ShouldNot(Panic())
			})
		})

		Context("with a non-nil error", func() {
			It("does aborts, tacking on the message", func() {
				Ω(func() {
					command.AbortIfError("boom boom?", fmt.Errorf("kaboom!"))
				}).Should(PanicWith(command.AbortDetails{
					ExitCode:  1,
					Error:     fmt.Errorf("boom boom?\nkaboom!"),
					EmitUsage: false,
				}))
			})
		})
	})

	Describe("AbortIfErrors", func() {
		Context("with an empty errors", func() {
			It("does not abort", func() {
				Ω(func() {
					command.AbortIfErrors("boom boom?", []error{})
				}).ShouldNot(Panic())
			})
		})

		Context("with non-nil errors", func() {
			It("does aborts, tacking on the messages", func() {
				Ω(func() {
					command.AbortIfErrors("boom boom?", []error{fmt.Errorf("kaboom!\n"), fmt.Errorf("kababoom!!\n")})
				}).Should(PanicWith(command.AbortDetails{
					ExitCode:  1,
					Error:     fmt.Errorf("boom boom?\nkaboom!\nkababoom!!\n"),
					EmitUsage: false,
				}))
			})
		})
	})
})

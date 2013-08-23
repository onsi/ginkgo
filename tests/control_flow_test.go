package gospec_tests

import (
	"fmt"
	. "github.com/onsi/godescribe"
	"time"
)

func init() {
	Describe("A test", func() {
		var a int
		var b string

		BeforeEach(func() {
			a = 5
		})

		JustBeforeEach(func() {
			b = fmt.Sprintf("%d", a)
		})

		It("should pass", func() {
			if a != 5 {
				Fail("Wanted 5!")
			}
		})

		It("even if its empty", func() {})

		It("should have just before eaches", func() {
			if b != "5" {
				Fail("Wanted 5!")
			}
		})

		It("should pass (eventually)", func(done Done) {
			go func() {
				time.Sleep(1 * time.Second)
				if a != 5 {
					Fail("Wanted 5!")
				}
				done <- true
			}()
		})

		It("should timeout", func(done Done) {
			go func() {
				time.Sleep(10 * time.Second)
				if a != 5 {
					Fail("Wanted 5!")
				}
				done <- true
			}()
		})

		PIt("should be pending", func() {})

		It("should panic", func() {
			panic("AAAH!!")
		})

		Context("Afters too, and they can fail", func() {
			It("should be fine until its not", func() {})

			AfterEach(func() {
				Fail("Oops!")
			})
		})

		Context("When stuff fails", func() {
			BeforeEach(func(done Done) {
				go func() {
					time.Sleep(10 * time.Millisecond)
					a = 4
					done <- true
				}()
			})

			PIt("should have just before eaches", func() {
				if b != "4" {
					Fail("Wanted 5!")
				}
			})

			It("should fail", func() {
				if a != 5 {
					Fail("Wanted 5!")
				}
			})

			It("should fail (eventually)", func(done Done) {
				go func() {
					time.Sleep(1.0 * time.Second)
					if a != 5 {
						Fail("Wanted 5!")
					}
					done <- true
				}()
			})
		})
	})
}

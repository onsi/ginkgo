package gospec_tests

import (
	. "github.com/onsi/godescribe"
	"time"
)

func init() {
	Describe("A test", func() {
		var a int

		BeforeEach(func() {
			a = 5
		})

		It("should pass", func() {
			if a != 5 {
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

		PIt("should be pending", func() {})

		It("should panic", func() {
			panic("AAAH!!")
		})

		Context("When stuff fails", func() {
			BeforeEach(func(done Done) {
				go func() {
					time.Sleep(10 * time.Millisecond)
					a = 4
					done <- true
				}()
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

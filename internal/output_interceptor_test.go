package internal_test

import (
	"fmt"
	"os"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/onsi/ginkgo/internal"
)

var _ = Describe("OutputInterceptor", func() {
	var interceptor internal.OutputInterceptor

	sharedInterceptorTests := func() {
		It("intercepts output", func() {
			for i := 0; i < 2048; i++ { //we loop here to stress test and make sure we aren't leaking any file descriptors
				interceptor.StartInterceptingOutput()
				fmt.Println("hi stdout")
				fmt.Fprintln(os.Stderr, "hi stderr")
				output := interceptor.StopInterceptingAndReturnOutput()
				Ω(output).Should(Equal("hi stdout\nhi stderr\n"))
			}
		})

		It("can forward intercepted output to a buffer", func() {
			buffer := gbytes.NewBuffer()
			interceptor.StartInterceptingOutputAndForwardTo(buffer)
			fmt.Println("hi stdout")
			fmt.Fprintln(os.Stderr, "hi stderr")
			output := interceptor.StopInterceptingAndReturnOutput()
			Ω(output).Should(Equal("hi stdout\nhi stderr\n"))
			Ω(buffer).Should(gbytes.Say("hi stdout\nhi stderr\n"))
		})

		It("is stable across multiple shutdowns", func() {
			numRoutines := runtime.NumGoroutine()
			for i := 0; i < 2048; i++ { //we loop here to stress test and make sure we aren't leaking any file descriptors
				interceptor.StartInterceptingOutput()
				fmt.Println("hi stdout")
				fmt.Fprintln(os.Stderr, "hi stderr")
				output := interceptor.StopInterceptingAndReturnOutput()
				Ω(output).Should(Equal("hi stdout\nhi stderr\n"))
				interceptor.Shutdown()
			}
			Eventually(runtime.NumGoroutine).Should(BeNumerically("~", numRoutines, 10))
		})
	}

	Context("the OutputInterceptor for this OS", func() {
		BeforeEach(func() {
			interceptor = internal.NewOutputInterceptor()
			DeferCleanup(interceptor.Shutdown)
		})
		sharedInterceptorTests()
	})

	Context("the OSGlobalReassigningOutputInterceptor used on windows", func() {
		BeforeEach(func() {
			interceptor = internal.NewOSGlobalReassigningOutputInterceptor()
			DeferCleanup(interceptor.Shutdown)
		})

		sharedInterceptorTests()
	})

})

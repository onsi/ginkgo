package internal_test

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/onsi/ginkgo/v2/internal"
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

		It("can bail out if stdout and stderr are tied up by an external process", func() {
			// See GitHub issue #851: https://github.com/onsi/ginkgo/issues/851
			interceptor.StartInterceptingOutput()
			cmd := exec.Command("sleep", "60")
			//by threading stdout and stderr through, the sleep process will hold them open and prevent the interceptor from stopping:
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
			Ω(cmd.Start()).Should(Succeed())
			fmt.Println("hi stdout")
			fmt.Fprintln(os.Stderr, "hi stderr")

			// we try to stop here and see that we bail out eventually:
			outputChan := make(chan string)
			go func() {
				outputChan <- interceptor.StopInterceptingAndReturnOutput()
			}()
			var output string
			Eventually(outputChan, internal.BAILOUT_TIME*2).Should(Receive(&output))
			Ω(output).Should(Equal(internal.BAILOUT_MESSAGE))

			//subsequent attempts should be fine
			interceptor.StartInterceptingOutput()
			fmt.Println("hi stdout, again")
			fmt.Fprintln(os.Stderr, "hi stderr, again")
			output = interceptor.StopInterceptingAndReturnOutput()
			Ω(output).Should(Equal("hi stdout, again\nhi stderr, again\n"))

			cmd.Process.Kill()

			interceptor.StartInterceptingOutput()
			fmt.Println("hi stdout, once more")
			fmt.Fprintln(os.Stderr, "hi stderr, once more")
			output = interceptor.StopInterceptingAndReturnOutput()
			Ω(output).Should(Equal("hi stdout, once more\nhi stderr, once more\n"))
		})

		It("doesn't get stuck if it's paused and resumed before starting an external process that attaches to stdout/stderr", func() {
			// See GitHub issue #851: https://github.com/onsi/ginkgo/issues/851
			interceptor.StartInterceptingOutput()
			interceptor.PauseIntercepting()
			cmd := exec.Command("sleep", "60")
			//by threading stdout and stderr through, the sleep process will hold them open and prevent the interceptor from stopping:
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
			Ω(cmd.Start()).Should(Succeed())

			interceptor.ResumeIntercepting()
			fmt.Println("hi stdout")
			fmt.Fprintln(os.Stderr, "hi stderr")
			output := interceptor.StopInterceptingAndReturnOutput()

			Ω(output).Should(Equal("hi stdout\nhi stderr\n"))
			Ω(output).ShouldNot(ContainSubstring(internal.BAILOUT_MESSAGE))
			cmd.Process.Kill()
		})

		It("can start/stop/pause/resume correctly", func() {
			interceptor.StartInterceptingOutput()
			fmt.Fprint(os.Stdout, "O-A")
			fmt.Fprint(os.Stderr, "E-A")
			interceptor.PauseIntercepting()
			fmt.Fprint(os.Stdout, "O-B")
			fmt.Fprint(os.Stderr, "E-B")
			interceptor.ResumeIntercepting()
			fmt.Fprint(os.Stdout, "O-C")
			fmt.Fprint(os.Stderr, "E-C")
			interceptor.ResumeIntercepting() //noop
			fmt.Fprint(os.Stdout, "O-D")
			fmt.Fprint(os.Stderr, "E-D")
			interceptor.PauseIntercepting()
			fmt.Fprint(os.Stdout, "O-E")
			fmt.Fprint(os.Stderr, "E-E")
			interceptor.PauseIntercepting() //noop
			fmt.Fprint(os.Stdout, "O-F")
			fmt.Fprint(os.Stderr, "E-F")
			interceptor.ResumeIntercepting()
			fmt.Fprint(os.Stdout, "O-G")
			fmt.Fprint(os.Stderr, "E-G")
			interceptor.StartInterceptingOutput() //noop
			fmt.Fprint(os.Stdout, "O-H")
			fmt.Fprint(os.Stderr, "E-H")
			interceptor.PauseIntercepting()
			output := interceptor.StopInterceptingAndReturnOutput()
			Ω(output).Should(Equal("O-AE-AO-CE-CO-DE-DO-GE-GO-HE-H"))
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

package main_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal"
	. "github.com/onsi/gomega"
)

func TestInterceptorFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "InterceptorFixture Suite")
}

var _ = Describe("Ensuring the OutputInterceptor handles the edge case where an external process keeps the interceptor's pipe open", func() {
	var interceptor internal.OutputInterceptor

	sharedBehavior := func() {
		It("can avoid getting stuck, but also doesn't block the external process", func() {
			interceptor.StartInterceptingOutput()
			cmd := exec.Command("./interceptor")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			Ω(cmd.Start()).Should(Succeed())

			By("Bailing out because the pipe is stuck")
			outputChan := make(chan string)
			go func() {
				outputChan <- interceptor.StopInterceptingAndReturnOutput()
			}()
			var output string
			Eventually(outputChan, internal.BAILOUT_TIME*2).Should(Receive(&output))
			Ω(output).Should(Equal(internal.BAILOUT_MESSAGE))

			By("Not subsequently bailing out because the new pipe isn't tied to an external process")
			interceptor.StartInterceptingOutput()
			output = interceptor.StopInterceptingAndReturnOutput()
			Ω(output).ShouldNot(ContainSubstring("STDOUT"), "we're no longer capturing this output")
			Ω(output).ShouldNot(ContainSubstring("STDERR"), "we're no longer capturing this output")
			Ω(output).ShouldNot(ContainSubstring(internal.BAILOUT_MESSAGE), "we didn't have to bail out")

			expected := ""
			for i := 0; i < 300; i++ {
				expected += fmt.Sprintf("FILE %d\n", i)
			}
			Eventually(func(g Gomega) string {
				out, err := os.ReadFile("file-output")
				g.Ω(err).ShouldNot(HaveOccurred())
				return string(out)
			}, 5*time.Second).Should(Equal(expected))

			os.Remove("file-output")
		})

		It("works successfully if the user pauses then resumes around starging an external process", func() {
			interceptor.StartInterceptingOutput()
			interceptor.PauseIntercepting()

			cmd := exec.Command("./interceptor")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			Ω(cmd.Start()).Should(Succeed())

			interceptor.ResumeIntercepting()
			output := interceptor.StopInterceptingAndReturnOutput()
			Ω(output).ShouldNot(ContainSubstring(internal.BAILOUT_MESSAGE), "we didn't have to bail out")

			interceptor.StartInterceptingOutput()
			output = interceptor.StopInterceptingAndReturnOutput()
			Ω(output).ShouldNot(ContainSubstring(internal.BAILOUT_MESSAGE), "we still didn't have to bail out")

			expected := ""
			for i := 0; i < 300; i++ {
				expected += fmt.Sprintf("FILE %d\n", i)
			}
			Eventually(func(g Gomega) string {
				out, err := os.ReadFile("file-output")
				g.Ω(err).ShouldNot(HaveOccurred())
				return string(out)
			}, 5*time.Second).Should(Equal(expected))

			os.Remove("file-output")
		})
	}

	Context("the dup2 interceptor", func() {
		BeforeEach(func() {
			interceptor = internal.NewOutputInterceptor()
		})

		sharedBehavior()
	})

	Context("the global reassigning interceptor", func() {
		BeforeEach(func() {
			interceptor = internal.NewOSGlobalReassigningOutputInterceptor()
		})

		sharedBehavior()
	})
})

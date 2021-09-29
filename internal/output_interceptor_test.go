package internal_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/internal"
)

var _ = Describe("OutputInterceptor", func() {
	It("intercepts output", func() {
		interceptor := internal.NewOutputInterceptor()
		for i := 0; i < 2048; i++ { //we loop here to stress test and make sure we aren't leaking any file descriptors
			interceptor.StartInterceptingOutput()
			fmt.Println("hi stdout")
			fmt.Fprintln(os.Stderr, "hi stderr")
			output := interceptor.StopInterceptingAndReturnOutput()
			Ω(output).Should(Equal("hi stdout\nhi stderr\n"))
		}
	})
})

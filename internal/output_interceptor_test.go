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
		interceptor.StartInterceptingOutput()
		fmt.Println("hi stdout")
		fmt.Fprintln(os.Stderr, "hi stderr")
		output := interceptor.StopInterceptingAndReturnOutput()
		Ω(output).Should(Equal("hi stdout\nhi stderr\n"))
	})
})

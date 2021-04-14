package internal_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/onsi/ginkgo/internal"
	"github.com/onsi/ginkgo/types"
)

var _ = Describe("Counter", func() {
	var counter func() (int, error)
	var conf types.SuiteConfig

	BeforeEach(func() {
		conf = types.SuiteConfig{}
	})

	JustBeforeEach(func() {
		counter = internal.MakeNextIndexCounter(conf)
	})

	Context("when running in series", func() {
		BeforeEach(func() {
			conf.ParallelTotal = 1
		})

		It("returns a counter that grows by one with each invocation", func() {
			for i := 0; i < 10; i += 1 {
				Ω(counter()).Should(Equal(i))
			}
		})
	})

	Context("when running in parallel", func() {
		var server *ghttp.Server
		BeforeEach(func() {
			conf.ParallelTotal = 2
			server = ghttp.NewServer()
			conf.ParallelHost = server.URL()
		})

		AfterEach(func() {
			server.Close()
		})

		Context("when the server returns with a counter", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/counter"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, internal.Counter{Index: 1138}),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/counter"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, internal.Counter{Index: 27}),
					),
				)
			})

			It("forwards along the returned counter", func() {
				Ω(counter()).Should(Equal(1138))
				Ω(counter()).Should(Equal(27))
			})
		})

		Context("when an error occurs trying to get", func() {
			BeforeEach(func() {
				server.Close()
			})

			It("returns an error", func() {
				index, err := counter()
				Ω(index).Should(Equal(-1))
				Ω(err).Should(HaveOccurred())
			})
		})

		Context("when the server returns a non-ok status code", func() {
			BeforeEach(func() {
				server.AppendHandlers(ghttp.RespondWithJSONEncoded(http.StatusInternalServerError, internal.Counter{Index: 1138}))
			})

			It("returns an error", func() {
				index, err := counter()
				Ω(index).Should(Equal(-1))
				Ω(err).Should(MatchError("unexpected status code 500"))
			})
		})

		Context("when the server returns garbage json", func() {
			BeforeEach(func() {
				server.AppendHandlers(ghttp.RespondWith(http.StatusOK, "∫"))
			})

			It("returns an error", func() {
				index, err := counter()
				Ω(index).Should(Equal(-1))
				Ω(err).Should(HaveOccurred())
			})

		})
	})
})

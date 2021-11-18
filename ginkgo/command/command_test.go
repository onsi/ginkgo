package command_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/formatter"
	"github.com/onsi/ginkgo/v2/ginkgo/command"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Command", func() {
	var c command.Command
	var rt *RunTracker

	BeforeEach(func() {
		rt = NewRunTracker()
		fs, err := types.NewGinkgoFlagSet(
			types.GinkgoFlags{
				{Name: "contrabulaturally", KeyPath: "C", Usage: "with irridiacy"},
			},
			&(struct{ C int }{C: 17}),
			types.GinkgoFlagSections{},
		)
		Ω(err).ShouldNot(HaveOccurred())
		c = command.Command{
			Name:          "enflabulate",
			Flags:         fs,
			Usage:         "flooper enflabulate <args>",
			ShortDoc:      "Enflabulate all the mandribles",
			Documentation: "Coherent quasistatic protocols will be upended if contrabulaturally is greater than 23.",
			DocLink:       "fabulous-enflabulence",
			Command:       rt.C("enflabulate"),
		}
	})

	Describe("Run", func() {
		Context("when flags fails to parse", func() {
			It("aborts with usage", func() {
				Ω(func() {
					c.Run([]string{"-not-a-flag=oops"}, []string{"additional", "args"})
				}).Should(PanicWith(SatisfyAll(
					HaveField("ExitCode", 1),
					HaveField("Error", HaveOccurred()),
					HaveField("EmitUsage", BeTrue()),
				)))

				Ω(rt).Should(HaveTrackedNothing())
			})
		})

		Context("when flags parse", func() {
			It("runs the command", func() {
				c.Run([]string{"-contrabulaturally=16", "and-an-arg", "and-another"}, []string{"additional", "args"})
				Ω(rt).Should(HaveRun("enflabulate"))

				Ω(rt.DataFor("enflabulate")["Args"]).Should(Equal([]string{"and-an-arg", "and-another"}))
				Ω(rt.DataFor("enflabulate")["AdditionalArgs"]).Should(Equal([]string{"additional", "args"}))

			})
		})
	})

	Describe("Usage", func() {
		BeforeEach(func() {
			formatter.SingletonFormatter.ColorMode = formatter.ColorModePassthrough
		})

		It("emits a nicely formatted usage", func() {
			buf := gbytes.NewBuffer()
			c.EmitUsage(buf)

			expected := strings.Join([]string{
				"{{bold}}flooper enflabulate <args>{{/}}",
				"{{gray}}--------------------------{{/}}",
				"Enflabulate all the mandribles",
				"",
				"Coherent quasistatic protocols will be upended if contrabulaturally is greater",
				"than 23.",
				"",
				"{{bold}}Learn more at:{{/}} {{cyan}}{{underline}}http://onsi.github.io/ginkgo/#fabulous-enflabulence{{/}}",
				"",
				"  --contrabulaturally{{/}} [int] {{gray}}{{/}}",
				"    {{light-gray}}with irridiacy{{/}}",
				"", "",
			}, "\n")

			Ω(string(buf.Contents())).Should(Equal(expected))
		})
	})
})

package formatter_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/formatter"
	. "github.com/onsi/gomega"
)

var _ = Describe("Formatter", func() {
	var colorMode formatter.ColorMode
	var f formatter.Formatter

	BeforeEach(func() {
		colorMode = formatter.ColorModeTerminal
		os.Unsetenv("GINKGO_CLI_COLOR_RED")
		os.Unsetenv("GINKGO_CLI_COLOR_ORANGE")
		os.Unsetenv("GINKGO_CLI_COLOR_CORAL")
		os.Unsetenv("GINKGO_CLI_COLOR_MAGENTA")
		os.Unsetenv("GINKGO_CLI_COLOR_GREEN")
		os.Unsetenv("GINKGO_CLI_COLOR_DARK_GREEN")
		os.Unsetenv("GINKGO_CLI_COLOR_YELLOW")
		os.Unsetenv("GINKGO_CLI_COLOR_LIGHT_YELLOW")
		os.Unsetenv("GINKGO_CLI_COLOR_CYAN")
		os.Unsetenv("GINKGO_CLI_COLOR_LIGHT_GRAY")
		os.Unsetenv("GINKGO_CLI_COLOR_BLUE")
	})

	JustBeforeEach(func() {
		f = formatter.New(colorMode)
	})

	Context("with ColorModeNone", func() {
		BeforeEach(func() {
			colorMode = formatter.ColorModeNone
		})

		It("strips out color information", func() {
			Ω(f.F("{{green}}{{bold}}hi there{{/}}")).Should(Equal("hi there"))
		})
	})

	Context("with ColorModeTerminal", func() {
		BeforeEach(func() {
			colorMode = formatter.ColorModeTerminal
		})

		It("renders the color information using terminal escape codes", func() {
			Ω(f.F("{{green}}{{bold}}hi there{{/}}")).Should(Equal("\x1b[38;5;10m\x1b[1mhi there\x1b[0m"))
		})
	})

	Context("with ColorModePassthrough", func() {
		BeforeEach(func() {
			colorMode = formatter.ColorModePassthrough
		})

		It("leaves the color information as is, allowing us to test statements more easily", func() {
			Ω(f.F("{{green}}{{bold}}hi there{{/}}")).Should(Equal("{{green}}{{bold}}hi there{{/}}"))
		})
	})

	DescribeTable("with environment overrides",
		func(envVars map[string]string, input, expected string) {
			for envVar, value := range envVars {
				os.Setenv(envVar, value)
			}
			f := formatter.New(colorMode)
			Ω(f.F(input)).Should(Equal(expected))
			for envVar := range envVars {
				os.Unsetenv(envVar)
			}
		},

		Entry("uses default for too low codes", map[string]string{
			"GINKGO_CLI_COLOR_RED": "-1",
		}, "{{red}}hi there{{/}}", "\x1b[38;5;9mhi there\x1b[0m"),

		Entry("uses default for too high codes", map[string]string{
			"GINKGO_CLI_COLOR_RED": "256",
		}, "{{red}}hi there{{/}}", "\x1b[38;5;9mhi there\x1b[0m"),

		Entry("supports literal alias for 8bit color", map[string]string{
			"GINKGO_CLI_COLOR_RED": "red",
		}, "{{red}}hi there{{/}}", "\x1b[38;5;1mhi there\x1b[0m"),

		Entry("supports number alias for 8bit color", map[string]string{
			"GINKGO_CLI_COLOR_RED": "1",
		}, "{{red}}hi there{{/}}", "\x1b[38;5;1mhi there\x1b[0m"),

		Entry("supports 16bit colors (bright)", map[string]string{
			"GINKGO_CLI_COLOR_RED": "9",
		}, "{{red}}hi there{{/}}", "\x1b[38;5;9mhi there\x1b[0m"),

		Entry("supports 16bit color literal aliases (bright)", map[string]string{
			"GINKGO_CLI_COLOR_RED": "bright-red",
		}, "{{red}}hi there{{/}}", "\x1b[38;5;9mhi there\x1b[0m"),

		Entry("supports extended 256 colors", map[string]string{
			"GINKGO_CLI_COLOR_RED": "16",
		}, "{{red}}hi there{{/}}", "\x1b[38;5;16mhi there\x1b[0m"),
	)

	Describe("NewWithNoColorBool", func() {
		Context("when the noColor bool is true", func() {
			It("strips out color information", func() {
				f = formatter.NewWithNoColorBool(true)
				Ω(f.F("{{green}}{{bold}}hi there{{/}}")).Should(Equal("hi there"))
			})
		})

		Context("when the noColor bool is false", func() {
			It("renders the color information using terminal escape codes", func() {
				f = formatter.NewWithNoColorBool(false)
				Ω(f.F("{{green}}{{bold}}hi there{{/}}")).Should(Equal("\x1b[38;5;10m\x1b[1mhi there\x1b[0m"))
			})
		})
	})

	Describe("F", func() {
		It("transforms the color information and sprintfs", func() {
			Ω(f.F("{{green}}hi there {{cyan}}%d {{yellow}}%s{{/}}", 3, "wise men")).Should(Equal("\x1b[38;5;10mhi there \x1b[38;5;14m3 \x1b[38;5;11mwise men\x1b[0m"))
		})

		It("avoids sprintf if there are no additional arguments", func() {
			Ω(f.F("{{green}}hi there {{cyan}}%d {{yellow}}%s{{/}}")).Should(Equal("\x1b[38;5;10mhi there \x1b[38;5;14m%d \x1b[38;5;11m%s\x1b[0m"))
		})
	})

	Describe("Fi", func() {
		It("transforms the color information, sprintfs, and applies an indentation", func() {
			Ω(f.Fi(2, "{{green}}hi there\n{{cyan}}%d {{yellow}}%s{{/}}", 3, "wise men")).Should(Equal(
				"    \x1b[38;5;10mhi there\n    \x1b[38;5;14m3 \x1b[38;5;11mwise men\x1b[0m",
			))
		})
	})

	DescribeTable("Fiw",
		func(indentation int, maxWidth int, input string, expected ...string) {
			Ω(f.Fiw(uint(indentation), uint(maxWidth), input)).Should(Equal(strings.Join(expected, "\n")))
		},
		Entry("basic case", 0, 0, "a really long string is fine", "a really long string is fine"),
		Entry("indentation is accounted for in width",
			1, 10,
			"1234 678",
			"  1234 678",
		),
		Entry("indentation is accounted for in width",
			1, 10,
			"1234 6789",
			"  1234",
			"  6789",
		),
		Entry("when there is a nice long sentence",
			0, 10,
			"12 456 890 1234 5",
			"12 456 890",
			"1234 5",
		),
		Entry("when a word in a sentence intersects the boundary",
			0, 10,
			"12 456 8901 123 45",
			"12 456",
			"8901 123",
			"45",
		),
		Entry("when a word in a sentence is just too long",
			0, 10,
			"12 12345678901 12 12345 678901 12345678901",
			"12",
			"12345678901",
			"12 12345",
			"678901",
			"12345678901",
		),
	)

	Describe("CycleJoin", func() {
		It("combines elements, cycling through styles as it goes", func() {
			Ω(f.CycleJoin([]string{"a", "b", "c"}, "|", []string{"{{red}}", "{{green}}"})).Should(Equal(
				"\x1b[38;5;9ma|\x1b[38;5;10mb|\x1b[38;5;9mc\x1b[0m",
			))
		})
	})
})

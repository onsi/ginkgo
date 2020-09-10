package command_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/onsi/ginkgo/formatter"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/onsi/ginkgo/ginkgo/command"
)

var _ = Describe("Program", func() {
	var program command.Program
	var rt *RunTracker
	var buf *gbytes.Buffer

	BeforeEach(func() {
		rt = NewRunTracker()
		defaultCommand := command.Command{Name: "alpha", Usage: "alpha usage", ShortDoc: "such usage!", Command: rt.C("alpha")}

		fs, err := config.NewGinkgoFlagSet(
			config.GinkgoFlags{
				{Name: "decay-rate", KeyPath: "Rate", Usage: "set the decay rate, in years"},
				{DeprecatedName: "old", KeyPath: "Old"},
			},
			&(struct {
				Rate float64
				Old  bool
			}{Rate: 17.0}),
			config.GinkgoFlagSections{},
		)
		Ω(err).ShouldNot(HaveOccurred())
		commands := []command.Command{
			{Name: "beta", Flags: fs, Usage: "beta usage", ShortDoc: "such usage!", Command: rt.C("beta")},
			{Name: "gamma", Command: rt.C("gamma")},
			{Name: "zeta", Command: rt.C("zeta", func() {
				command.Abort(command.AbortDetails{Error: fmt.Errorf("Kaboom!"), ExitCode: 17})
			})},
		}

		deprecatedCommands := []command.DeprecatedCommand{
			{Name: "delta", Deprecation: types.Deprecation{Message: "delta is for deprecated"}},
		}

		formatter.SingletonFormatter.ColorMode = formatter.ColorModePassthrough

		buf = gbytes.NewBuffer()
		program = command.Program{
			Name:               "omicron",
			Heading:            "Omicron v2.0.0",
			Commands:           commands,
			DefaultCommand:     defaultCommand,
			DeprecatedCommands: deprecatedCommands,

			Exiter: func(code int) {
				rt.RunWithData("exit", "Code", code)
			},
			OutWriter: buf,
			ErrWriter: buf,
		}
	})

	Context("when called with no subcommand", func() {
		BeforeEach(func() {
			program.RunAndExit([]string{"omicron"}) //os.Args always includes the name of the program as the first element
		})

		It("runs the default command", func() {
			Ω(rt).Should(HaveTracked("alpha", "exit"))
			Ω(rt).Should(HaveRunWithData("exit", "Code", 0))
			Ω(buf.Contents()).Should(BeEmpty())
		})
	})

	Context("when called with the default command's name", func() {
		BeforeEach(func() {
			program.RunAndExit([]string{"omicron", "alpha", "args1", "args2"})
		})

		It("runs the default command", func() {
			Ω(rt).Should(HaveTracked("alpha", "exit"))
			Ω(rt).Should(HaveRunWithData("alpha", "Args", []string{"args1", "args2"}))
			Ω(rt).Should(HaveRunWithData("exit", "Code", 0))
			Ω(buf.Contents()).Should(BeEmpty())
		})
	})

	Context("when called with a subcommand", func() {
		BeforeEach(func() {
			program.RunAndExit([]string{"omicron", "beta"})
		})

		It("runs that subcommand", func() {
			Ω(rt).Should(HaveTracked("beta", "exit"))
			Ω(rt).Should(HaveRunWithData("exit", "Code", 0))
			Ω(buf.Contents()).Should(BeEmpty())
		})
	})

	Context("when called with an unkown subcommand", func() {
		BeforeEach(func() {
			program.RunAndExit([]string{"omicron", "xi"})
		})

		It("calls the default command with arguments", func() {
			Ω(rt).Should(HaveTracked("alpha", "exit"))
			Ω(rt).Should(HaveRunWithData("alpha", "Args", []string{"xi"}))
			Ω(rt).Should(HaveRunWithData("exit", "Code", 0))
			Ω(buf.Contents()).Should(BeEmpty())
		})
	})

	Context("when passed arguments and additional arguments", func() {
		BeforeEach(func() {
			program.RunAndExit([]string{"omicron", "gamma", "arg1", "-arg2", "--", "addArg1", "addArg2"})
		})
		It("passes both in", func() {
			Ω(rt).Should(HaveTracked("gamma", "exit"))
			Ω(rt).Should(HaveRunWithData("gamma", "Args", []string{"arg1", "-arg2"}, "AdditionalArgs", []string{"addArg1", "addArg2"}))
			Ω(rt).Should(HaveRunWithData("exit", "Code", 0))
			Ω(buf.Contents()).Should(BeEmpty())
		})
	})

	DescribeTable("Emitting help when asked", func(args []string) {
		program.RunAndExit(args)
		Ω(rt).Should(HaveTracked("exit"))
		Ω(rt).Should(HaveRunWithData("exit", "Code", 0))
		//HavePrefix to avoid trailing whitespace causing failures
		Ω(string(buf.Contents())).Should(HavePrefix(strings.Join([]string{
			"Omicron v2.0.0",
			"{{gray}}--------------{{/}}",
			"For usage information, run {{bold}}omicron help <COMMAND>{{/}}:",
			"  {{bold}}omicron{{/}} or omicron {{bold}}alpha{{/}} - {{gray}}alpha usage{{/}}",
			"    such usage!",
			"  {{bold}}beta{{/}} - {{gray}}beta usage{{/}}",
			"    such usage!",
			"  {{bold}}gamma{{/}} - {{gray}}{{/}}",
			"  {{bold}}zeta{{/}} - {{gray}}{{/}}",
		}, "\n")))
	},
		Entry("with help", []string{"omicron", "help"}),
		Entry("with -help", []string{"omicron", "-help"}),
		Entry("with --help", []string{"omicron", "--help"}),
		Entry("with -h", []string{"omicron", "-h"}),
		Entry("with --h", []string{"omicron", "--h"}),
	)

	DescribeTable("Emitting help for the default command", func(args []string) {
		program.RunAndExit(args)
		Ω(rt).Should(HaveTracked("exit"))
		Ω(rt).Should(HaveRunWithData("exit", "Code", 0))
		Ω(string(buf.Contents())).Should(HavePrefix(strings.Join([]string{
			"{{bold}}alpha usage{{/}}",
			"{{gray}}-----------{{/}}",
			"such usage!",
		}, "\n")))
	},
		Entry("with help omicron", []string{"omicron", "help", "omicron"}),
		Entry("with -help omicron", []string{"omicron", "-help", "omicron"}),
		Entry("with --help omicron", []string{"omicron", "--help", "omicron"}),
		Entry("with -h omicron", []string{"omicron", "-h", "omicron"}),
		Entry("with --h omicron", []string{"omicron", "--h", "omicron"}),
		Entry("with help alpha", []string{"omicron", "help", "alpha"}),
		Entry("with -help alpha", []string{"omicron", "-help", "alpha"}),
		Entry("with --help alpha", []string{"omicron", "--help", "alpha"}),
		Entry("with -h alpha", []string{"omicron", "-h", "alpha"}),
		Entry("with --h alpha", []string{"omicron", "--h", "alpha"}),
		Entry("with alpha -help", []string{"omicron", "alpha", "-help"}),
		Entry("with alpha --help", []string{"omicron", "alpha", "--help"}),
		Entry("with alpha -h", []string{"omicron", "alpha", "-h"}),
		Entry("with alpha --h", []string{"omicron", "alpha", "--h"}),
	)

	DescribeTable("Emitting help for a known subcommand", func(args []string) {
		program.RunAndExit(args)
		Ω(rt).Should(HaveTracked("exit"))
		Ω(rt).Should(HaveRunWithData("exit", "Code", 0))
		Ω(string(buf.Contents())).Should(HavePrefix(strings.Join([]string{
			"{{bold}}beta usage{{/}}",
			"{{gray}}----------{{/}}",
			"such usage!",
			"",
			"  --decay-rate{{/}} [float] {{gray}}{{/}}",
			"    {{light-gray}}set the decay rate, in years{{/}}",
		}, "\n")))
	},
		Entry("with help beta", []string{"omicron", "help", "beta"}),
		Entry("with -help beta", []string{"omicron", "-help", "beta"}),
		Entry("with --help beta", []string{"omicron", "--help", "beta"}),
		Entry("with -h beta", []string{"omicron", "-h", "beta"}),
		Entry("with --h beta", []string{"omicron", "--h", "beta"}),
		Entry("with beta -help", []string{"omicron", "beta", "-help"}),
		Entry("with beta --help", []string{"omicron", "beta", "--help"}),
		Entry("with beta -h", []string{"omicron", "beta", "-h"}),
		Entry("with beta --h", []string{"omicron", "beta", "--h"}),
	)

	DescribeTable("Emitting help for an unknown subcommand", func(args []string) {
		program.RunAndExit(args)
		Ω(rt).Should(HaveTracked("exit"))
		Ω(rt).Should(HaveRunWithData("exit", "Code", 1))
		Ω(string(buf.Contents())).Should(HavePrefix(strings.Join([]string{
			"{{red}}Unknown Command: {{bold}}xi{{/}}",
			"",
			"Omicron v2.0.0",
			"{{gray}}--------------{{/}}",
			"For usage information, run {{bold}}omicron help <COMMAND>{{/}}:",
			"  {{bold}}omicron{{/}} or omicron {{bold}}alpha{{/}} - {{gray}}alpha usage{{/}}",
			"    such usage!",
			"  {{bold}}beta{{/}} - {{gray}}beta usage{{/}}",
			"    such usage!",
			"  {{bold}}gamma{{/}} - {{gray}}{{/}}",
			"  {{bold}}zeta{{/}} - {{gray}}{{/}}",
		}, "\n")))
	},
		Entry("with help xi", []string{"omicron", "help", "xi"}),
		Entry("with -help xi", []string{"omicron", "-help", "xi"}),
		Entry("with --help xi", []string{"omicron", "--help", "xi"}),
		Entry("with -h xi", []string{"omicron", "-h", "xi"}),
		Entry("with --h xi", []string{"omicron", "--h", "xi"}),
		Entry("with xi -help", []string{"omicron", "xi", "-help"}),
		Entry("with xi --help", []string{"omicron", "xi", "--help"}),
		Entry("with xi -h", []string{"omicron", "xi", "-h"}),
		Entry("with xi --h", []string{"omicron", "xi", "--h"}),
	)
	Context("when called with a deprecated command", func() {
		BeforeEach(func() {
			program.RunAndExit([]string{"omicron", "delta"})
		})

		It("lets the user know the command is deprecated", func() {
			Ω(rt).Should(HaveTracked("exit"))
			Ω(rt).Should(HaveRunWithData("exit", "Code", 0))
			Ω(string(buf.Contents())).Should(HavePrefix(strings.Join([]string{
				"{{light-yellow}}You're using deprecated Ginkgo functionality:{{/}}",
				"{{light-yellow}}============================================={{/}}",
				"  {{yellow}}delta is for deprecated{{/}}",
			}, "\n")))
		})
	})

	Context("when a deprecated flag is used", func() {
		BeforeEach(func() {
			program.RunAndExit([]string{"omicron", "beta", "-old"})
		})

		It("lets the user know a deprecated flag was used", func() {
			Ω(rt).Should(HaveTracked("beta", "exit"))
			Ω(rt).Should(HaveRunWithData("exit", "Code", 0))
			Ω(string(buf.Contents())).Should(HavePrefix(strings.Join([]string{
				"{{light-yellow}}You're using deprecated Ginkgo functionality:{{/}}",
				"{{light-yellow}}============================================={{/}}",
				"  {{yellow}}--old is deprecated{{/}}",
			}, "\n")))
		})
	})

	Context("when an unkown flag is used", func() {
		BeforeEach(func() {
			program.RunAndExit([]string{"omicron", "beta", "-zanzibar"})
		})

		It("emits usage for the associated subcommand", func() {
			Ω(rt).Should(HaveTracked("exit"))
			Ω(rt).Should(HaveRunWithData("exit", "Code", 1))
			Ω(string(buf.Contents())).Should(HavePrefix(strings.Join([]string{
				"{{red}}{{bold}}omicron beta{{/}} {{red}}failed{{/}}",
				"  flag provided but not defined: -zanzibar",
				"",
				"{{bold}}beta usage{{/}}",
				"{{gray}}----------{{/}}",
				"such usage!",
				"",
				"  --decay-rate{{/}} [float] {{gray}}{{/}}",
				"    {{light-gray}}set the decay rate, in years{{/}}",
			}, "\n")))
		})
	})

	Context("when a subcommand aborts", func() {
		BeforeEach(func() {
			program.RunAndExit([]string{"omicron", "zeta"})
		})

		It("emits information about the error", func() {
			Ω(rt).Should(HaveTracked("zeta", "exit"))
			Ω(rt).Should(HaveRunWithData("exit", "Code", 17))
			Ω(string(buf.Contents())).Should(HavePrefix(strings.Join([]string{
				"{{red}}{{bold}}omicron zeta{{/}} {{red}}failed{{/}}",
				"  Kaboom!",
			}, "\n")))
		})
	})

})

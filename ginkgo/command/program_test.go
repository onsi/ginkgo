package command_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/formatter"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/onsi/ginkgo/v2/ginkgo/command"
)

var _ = Describe("Program", func() {
	var program command.Program
	var rt *RunTracker
	var buf *gbytes.Buffer

	BeforeEach(func() {
		rt = NewRunTracker()
		defaultCommand := command.Command{Name: "alpha", Usage: "alpha usage", ShortDoc: "such usage!", Command: rt.C("alpha")}

		fs, err := types.NewGinkgoFlagSet(
			types.GinkgoFlags{
				{Name: "decay-rate", KeyPath: "Rate", Usage: "set the decay rate, in years"},
				{DeprecatedName: "old", KeyPath: "Old"},
			},
			&(struct {
				Rate float64
				Old  bool
			}{Rate: 17.0}),
			types.GinkgoFlagSections{},
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

	Context("when called with an unknown subcommand", func() {
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

	DescribeTable("Emitting help when asked",
		func(args []string) {
			program.RunAndExit(args)
			Ω(rt).Should(HaveTracked("exit"))
			Ω(rt).Should(HaveRunWithData("exit", "Code", 0))
			//HavePrefix to avoid trailing whitespace causing failures
			Ω(string(buf.Contents())).Should(HavePrefix(strings.Join([]string{
				"Omicron v2.0.0",
				"{{gray}}--------------{{/}}",
				"For usage information for a command, run {{bold}}omicron help COMMAND{{/}}.",
				"For usage information for the default command, run {{bold}}omicron help omicron{{/}} or {{bold}}omicron help alpha{{/}}.",
				"",
				"The following commands are available:",
				"  {{bold}}omicron{{/}} or omicron {{bold}}alpha{{/}} - {{gray}}alpha usage{{/}}",
				"    such usage!",
				"  {{bold}}beta{{/}} - {{gray}}beta usage{{/}}",
				"    such usage!",
				"  {{bold}}gamma{{/}} - {{gray}}{{/}}",
				"  {{bold}}zeta{{/}} - {{gray}}{{/}}",
			}, "\n")))
		},
		func(args []string) string {
			return fmt.Sprintf("with %s", args[1])
		},
		Entry(nil, []string{"omicron", "help"}),
		Entry(nil, []string{"omicron", "-help"}),
		Entry(nil, []string{"omicron", "--help"}),
		Entry(nil, []string{"omicron", "-h"}),
		Entry(nil, []string{"omicron", "--h"}),
	)

	DescribeTable("Emitting help for the default command",
		func(args []string) {
			program.RunAndExit(args)
			Ω(rt).Should(HaveTracked("exit"))
			Ω(rt).Should(HaveRunWithData("exit", "Code", 0))
			Ω(string(buf.Contents())).Should(HavePrefix(strings.Join([]string{
				"{{bold}}alpha usage{{/}}",
				"{{gray}}-----------{{/}}",
				"such usage!",
			}, "\n")))
		},
		func(args []string) string {
			return fmt.Sprintf("with %s %s", args[1], args[2])
		},
		Entry(nil, []string{"omicron", "help", "omicron"}),
		Entry(nil, []string{"omicron", "-help", "omicron"}),
		Entry(nil, []string{"omicron", "--help", "omicron"}),
		Entry(nil, []string{"omicron", "-h", "omicron"}),
		Entry(nil, []string{"omicron", "--h", "omicron"}),
		Entry(nil, []string{"omicron", "help", "alpha"}),
		Entry(nil, []string{"omicron", "-help", "alpha"}),
		Entry(nil, []string{"omicron", "--help", "alpha"}),
		Entry(nil, []string{"omicron", "-h", "alpha"}),
		Entry(nil, []string{"omicron", "--h", "alpha"}),
		Entry(nil, []string{"omicron", "alpha", "-help"}),
		Entry(nil, []string{"omicron", "alpha", "--help"}),
		Entry(nil, []string{"omicron", "alpha", "-h"}),
		Entry(nil, []string{"omicron", "alpha", "--h"}),
	)

	DescribeTable("Emitting help for a known subcommand",
		func(args []string) {
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
		func(args []string) string {
			return fmt.Sprintf("with %s %s", args[1], args[2])
		},
		Entry(nil, []string{"omicron", "help", "beta"}),
		Entry(nil, []string{"omicron", "-help", "beta"}),
		Entry(nil, []string{"omicron", "--help", "beta"}),
		Entry(nil, []string{"omicron", "-h", "beta"}),
		Entry(nil, []string{"omicron", "--h", "beta"}),
		Entry(nil, []string{"omicron", "beta", "-help"}),
		Entry(nil, []string{"omicron", "beta", "--help"}),
		Entry(nil, []string{"omicron", "beta", "-h"}),
		Entry(nil, []string{"omicron", "beta", "--h"}),
	)

	DescribeTable("Emitting help for an unknown subcommand",
		func(args []string) {
			program.RunAndExit(args)
			Ω(rt).Should(HaveTracked("exit"))
			Ω(rt).Should(HaveRunWithData("exit", "Code", 1))
			Ω(string(buf.Contents())).Should(HavePrefix(strings.Join([]string{
				"{{red}}Unknown Command: {{bold}}xi{{/}}",
				"",
				"Omicron v2.0.0",
				"{{gray}}--------------{{/}}",
				"For usage information for a command, run {{bold}}omicron help COMMAND{{/}}.",
				"For usage information for the default command, run {{bold}}omicron help omicron{{/}} or {{bold}}omicron help alpha{{/}}.",
				"",
				"The following commands are available:",
				"  {{bold}}omicron{{/}} or omicron {{bold}}alpha{{/}} - {{gray}}alpha usage{{/}}",
				"    such usage!",
				"  {{bold}}beta{{/}} - {{gray}}beta usage{{/}}",
				"    such usage!",
				"  {{bold}}gamma{{/}} - {{gray}}{{/}}",
				"  {{bold}}zeta{{/}} - {{gray}}{{/}}",
			}, "\n")))
		},
		func(args []string) string {
			return fmt.Sprintf("with %s %s", args[1], args[2])
		},
		Entry(nil, []string{"omicron", "help", "xi"}),
		Entry(nil, []string{"omicron", "-help", "xi"}),
		Entry(nil, []string{"omicron", "--help", "xi"}),
		Entry(nil, []string{"omicron", "-h", "xi"}),
		Entry(nil, []string{"omicron", "--h", "xi"}),
		Entry(nil, []string{"omicron", "xi", "-help"}),
		Entry(nil, []string{"omicron", "xi", "--help"}),
		Entry(nil, []string{"omicron", "xi", "-h"}),
		Entry(nil, []string{"omicron", "xi", "--h"}),
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

	Context("when an unknown flag is used", func() {
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

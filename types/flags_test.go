package types_test

import (
	"flag"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/formatter"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Flags", func() {
	BeforeEach(func() {
		format.CharactersAroundMismatchToInclude = 1000
		formatter.SingletonFormatter.ColorMode = formatter.ColorModePassthrough
	})

	Describe("GinkgoFlags", func() {
		Describe("CopyAppend", func() {
			It("concatenates the flags together, making a copy as it does so", func() {
				A := types.GinkgoFlags{{Name: "A"}, {Name: "B"}, {Name: "C"}}
				B := types.GinkgoFlags{{Name: "1"}, {Name: "2"}, {Name: "3"}}

				Ω(A.CopyAppend(B...)).Should(Equal(types.GinkgoFlags{{Name: "A"}, {Name: "B"}, {Name: "C"}, {Name: "1"}, {Name: "2"}, {Name: "3"}}))
				Ω(A).Should(HaveLen(3))
				Ω(B).Should(HaveLen(3))
			})
		})

		Describe("WithPrefix", func() {
			It("attaches the passed in prefi to the Name, DeprecatedName and ExportAs fields", func() {
				flags := types.GinkgoFlags{
					{Name: "A"},
					{DeprecatedName: "B"},
					{ExportAs: "C"},
				}
				prefixed := flags.WithPrefix("hello")
				Ω(prefixed).Should(Equal(types.GinkgoFlags{
					{Name: "hello.A"},
					{DeprecatedName: "hello.B"},
					{ExportAs: "hello.C"},
				}))
			})
		})

		Describe("SubsetWithNames", func() {
			It("returns the subset of flags with matching names", func() {
				A := types.GinkgoFlags{{Name: "A", Usage: "Hey A"}, {Name: "B", Usage: "Hey B"}, {Name: "C", Usage: "Hey C"}}
				subset := A.SubsetWithNames("A", "C", "D")
				Ω(subset).Should(Equal(types.GinkgoFlags{{Name: "A", Usage: "Hey A"}, {Name: "C", Usage: "Hey C"}}))
			})
		})
	})

	Describe("GinkgoFlagSections", func() {
		Describe("Lookup", func() {
			var sections types.GinkgoFlagSections
			BeforeEach(func() {
				sections = types.GinkgoFlagSections{
					{Key: "A", Heading: "Aft"},
					{Key: "B", Heading: "Starboard"},
				}
			})

			It("looks up the flag section with the passed in key", func() {
				section, found := sections.Lookup("A")
				Ω(found).Should(BeTrue())
				Ω(section).Should(Equal(sections[0]))
			})

			It("returns an empty flag section when the key is not found", func() {
				section, found := sections.Lookup("C")
				Ω(found).Should(BeFalse())
				Ω(section).Should(BeZero())
			})
		})
	})

	Describe("The zero GinkgoFlagSet", func() {
		It("returns true for IsZero", func() {
			Ω(types.GinkgoFlagSet{}.IsZero()).Should(BeTrue())
		})

		It("returns the passed in args when asked to parse", func() {
			args := []string{"-a=1", "-b=2", "-c=3"}
			Ω(types.GinkgoFlagSet{}.Parse(args)).Should(Equal(args))
		})

		It("does not validate any deprecations", func() {
			deprecationTracker := types.NewDeprecationTracker()
			types.GinkgoFlagSet{}.ValidateDeprecations(deprecationTracker)
			Ω(deprecationTracker.DidTrackDeprecations()).Should(BeFalse())
		})

		It("emits an empty string for usage", func() {
			Ω(types.GinkgoFlagSet{}.Usage()).Should(Equal(""))
		})
	})

	Describe("GinkgoFlagSet", func() {
		type StructA struct {
			StringProperty  string
			Int64Property   int64
			Float64Property float64
		}
		type StructB struct {
			IntProperty         int
			BoolProperty        bool
			StringSliceProperty []string
			DeprecatedProperty  string
		}
		var A StructA
		var B StructB
		var flags types.GinkgoFlags
		var bindings map[string]interface{}
		var sections types.GinkgoFlagSections
		var flagSet types.GinkgoFlagSet

		BeforeEach(func() {
			A = StructA{
				StringProperty:  "the default string",
				Int64Property:   1138,
				Float64Property: 3.141,
			}
			B = StructB{
				IntProperty:         2009,
				BoolProperty:        true,
				StringSliceProperty: []string{"once", "upon", "a time"},
				DeprecatedProperty:  "n/a",
			}
			bindings = map[string]interface{}{
				"A": &A,
				"B": &B,
			}
			sections = types.GinkgoFlagSections{
				{Key: "candy", Style: "{{red}}", Heading: "Candy Section", Description: "So sweet."},
				{Key: "dairy", Style: "{{blue}}", Heading: "Dairy Section", Description: "Abbreviated section", Succinct: true},
			}
			flags = types.GinkgoFlags{
				{Name: "string-flag", SectionKey: "candy", Usage: "string-usage", UsageArgument: "name", UsageDefaultValue: "Gerald", KeyPath: "A.StringProperty", DeprecatedName: "stringFlag"},
				{Name: "int-64-flag", SectionKey: "candy", Usage: "int-64-usage", KeyPath: "A.Int64Property", DeprecatedName: "int64Flag", DeprecatedDocLink: "no-more-camel-case"},
				{Name: "float-64-flag", SectionKey: "dairy", Usage: "float-64-usage", KeyPath: "A.Float64Property"},
				{Name: "int-flag", SectionKey: "invalid", Usage: "int-usage", KeyPath: "B.IntProperty"},
				{Name: "bool-flag", SectionKey: "candy", Usage: "bool-usage", KeyPath: "B.BoolProperty"},
				{Name: "string-slice-flag", SectionKey: "dairy", Usage: "string-slice-usage", KeyPath: "B.StringSliceProperty"},
				{SectionKey: "candy", DeprecatedName: "deprecated-flag", KeyPath: "B.DeprecatedProperty", Usage: "deprecated-usage"},
			}
		})

		Describe("Creation Failure Cases", func() {
			Context("when passed an unsuppoted type in the map", func() {
				BeforeEach(func() {
					type UnsupportedStructB struct {
						IntProperty         int
						BoolProperty        bool
						StringSliceProperty []string
						DeprecatedProperty  int32 //not supported
					}

					bindings = map[string]interface{}{
						"A": &A,
						"B": &UnsupportedStructB{},
					}
				})

				It("errors", func() {
					flagSet, err := types.NewGinkgoFlagSet(flags, bindings, sections)
					Ω(flagSet.IsZero()).Should(BeTrue())
					Ω(err).Should(HaveOccurred())
				})
			})

			Context("when the flags point to an invalid keypath in the map", func() {
				BeforeEach(func() {
					flags = append(flags, types.GinkgoFlag{Name: "welp-flag", Usage: "welp-usage", KeyPath: "A.WelpProperty"})
				})

				It("errors", func() {
					flagSet, err := types.NewGinkgoFlagSet(flags, bindings, sections)
					Ω(flagSet.IsZero()).Should(BeTrue())
					Ω(err).Should(HaveOccurred())
				})
			})
		})

		Describe("A stand-alone GinkgoFlagSet", func() {
			BeforeEach(func() {
				var err error
				flagSet, err = types.NewGinkgoFlagSet(flags, bindings, sections)
				Ω(flagSet.IsZero()).Should(BeFalse())
				Ω(err).ShouldNot(HaveOccurred())
			})

			Describe("Parsing flags", func() {
				It("maintains default values when no flags are parsed", func() {
					args, err := flagSet.Parse([]string{})
					Ω(err).ShouldNot(HaveOccurred())
					Ω(args).Should(Equal([]string{}))

					Ω(A.StringProperty).Should(Equal("the default string"))
					Ω(B.IntProperty).Should(Equal(2009))
				})

				It("updates the bindings when flags are parsed, returning any additional arguments", func() {
					args, err := flagSet.Parse([]string{
						"-string-flag", "a new string",
						"-int-64-flag=1139",
						"--float-64-flag", "2.71",
						"-int-flag=1984",
						"-bool-flag=false",
						"-string-slice-flag", "there lived",
						"-string-slice-flag", "three dragons",
						"extra-1",
						"extra-2",
					})
					Ω(err).ShouldNot(HaveOccurred())
					Ω(args).Should(Equal([]string{"extra-1", "extra-2"}))

					Ω(A.StringProperty).Should(Equal("a new string"))
					Ω(A.Int64Property).Should(Equal(int64(1139)))
					Ω(A.Float64Property).Should(Equal(2.71))
					Ω(B.IntProperty).Should(Equal(1984))
					Ω(B.BoolProperty).Should(Equal(false))
					Ω(B.StringSliceProperty).Should(Equal([]string{"once", "upon", "a time", "there lived", "three dragons"}))
				})

				It("updates the bindings when deprecated flags are set", func() {
					_, err := flagSet.Parse([]string{
						"-stringFlag", "deprecated but works",
						"-int64Flag=1234",
						"-deprecated-flag", "does not fail",
					})
					Ω(err).ShouldNot(HaveOccurred())

					Ω(A.StringProperty).Should(Equal("deprecated but works"))
					Ω(A.Int64Property).Should(Equal(int64(1234)))
					Ω(B.DeprecatedProperty).Should(Equal("does not fail"))
				})

				It("reports accurately on flags that were set", func() {
					_, err := flagSet.Parse([]string{
						"-string-flag", "a new string",
						"--float-64-flag", "2.71",
					})
					Ω(err).ShouldNot(HaveOccurred())

					Ω(flagSet.WasSet("string-flag")).Should(BeTrue())
					Ω(flagSet.WasSet("int-64-flag")).Should(BeFalse())
					Ω(flagSet.WasSet("float-64-flag")).Should(BeTrue())
				})
			})

			Describe("Validating Deprecations", func() {
				var deprecationTracker *types.DeprecationTracker
				BeforeEach(func() {
					deprecationTracker = types.NewDeprecationTracker()
				})

				Context("when no deprecated flags were invoked", func() {
					It("doesn't track any deprecations", func() {
						flagSet.Parse([]string{
							"--string-flag", "ok",
							"--int-flag", "1983",
						})
						flagSet.ValidateDeprecations(deprecationTracker)

						Ω(deprecationTracker.DidTrackDeprecations()).Should(BeFalse())
					})
				})

				Context("when deprecated flags were invoked", func() {
					It("tracks any detected deprecations with the passed in deprecation tracker", func() {
						flagSet.Parse([]string{
							"--stringFlag", "deprecated version",
							"--string-flag", "ok",
							"--int64Flag", "427",
						})
						flagSet.ValidateDeprecations(deprecationTracker)

						Ω(deprecationTracker.DidTrackDeprecations()).Should(BeTrue())
						report := deprecationTracker.DeprecationsReport()
						Ω(report).Should(ContainSubstring("--int64Flag is deprecated, use --int-64-flag instead"))
						Ω(report).Should(ContainSubstring("https://onsi.github.io/ginkgo/MIGRATING_TO_V2#no-more-camel-case"))
						Ω(report).Should(ContainSubstring("--stringFlag is deprecated, use --string-flag instead"))
					})
				})
			})

			Describe("Emitting Usage information", func() {
				It("emits information by section", func() {
					expectedUsage := []string{
						"{{red}}{{bold}}{{underline}}Candy Section{{/}}", //Candy section
						"So sweet.", //with heading
						"  {{red}}--string-flag{{/}} [name] {{gray}}(default: Gerald){{/}}", //flag with usage argument and default value
						"    {{light-gray}}string-usage{{/}}",
						"  {{red}}--int-64-flag{{/}} [int] {{gray}}{{/}}",
						"    {{light-gray}}int-64-usage{{/}}",
						"  {{red}}--bool-flag{{/}} {{gray}}{{/}}",
						"    {{light-gray}}bool-usage{{/}}",
						"",
						"{{blue}}{{bold}}{{underline}}Dairy Section{{/}}", //Dairy section is Succinct...
						"Abbreviated section",
						"  {{blue}}--float-64-flag, --string-slice-flag{{/}}", //so flags are just enumerated, without documentation
						"",
						"  --int-flag{{/}} [int] {{gray}}{{/}}",
						"    {{light-gray}}int-usage{{/}}",
						"",
						"",
					}
					Ω(flagSet.Usage()).Should(Equal(strings.Join(expectedUsage, "\n")))
				})
			})
		})

		Describe("A GinkgoFlagSet attached to an existing golang flagset", func() {
			var goFlagSet *flag.FlagSet

			BeforeEach(func() {
				var err error
				goFlagSet = flag.NewFlagSet("go-set", flag.ContinueOnError)
				goStringFlag := ""
				goIntFlag := 0
				goFlagSet.StringVar(&goStringFlag, "go-string-flag", "bob", "sets via `go`")
				goFlagSet.IntVar(&goIntFlag, "go-int-flag", 0, "an integer, please")
				flagSet, err = types.NewAttachedGinkgoFlagSet(goFlagSet, flags, bindings, sections, types.GinkgoFlagSection{
					Heading: "The go flags...",
				})
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("attaches its flags to go flag set, including deprecated flags", func() {
				registeredFlags := map[string]*flag.Flag{}
				goFlagSet.VisitAll(func(flag *flag.Flag) {
					registeredFlags[flag.Name] = flag
				})

				Ω(registeredFlags["string-flag"].Usage).Should(Equal("string-usage"))
				Ω(registeredFlags["int-64-flag"].Usage).Should(Equal("int-64-usage"))
				Ω(registeredFlags["float-64-flag"].Usage).Should(Equal("float-64-usage"))
				Ω(registeredFlags["int-flag"].Usage).Should(Equal("int-usage"))
				Ω(registeredFlags["bool-flag"].Usage).Should(Equal("bool-usage"))
				Ω(registeredFlags["string-slice-flag"].Usage).Should(Equal("string-slice-usage"))
				Ω(registeredFlags["string-flag"].DefValue).Should(Equal("the default string"))
				Ω(registeredFlags["int-64-flag"].DefValue).Should(Equal("1138"))
				Ω(registeredFlags["float-64-flag"].DefValue).Should(Equal("3.141"))
				Ω(registeredFlags["int-flag"].DefValue).Should(Equal("2009"))
				Ω(registeredFlags["bool-flag"].DefValue).Should(Equal("true"))
				Ω(registeredFlags["stringFlag"].Usage).Should(Equal("[DEPRECATED] use --string-flag instead"))
				Ω(registeredFlags["int64Flag"].Usage).Should(Equal("[DEPRECATED] use --int-64-flag instead"))
				Ω(registeredFlags["deprecated-flag"].Usage).Should(Equal("[DEPRECATED] deprecated-usage"))
			})

			It("overrides the goFlagSet's usage", func() {
				buf := gbytes.NewBuffer()
				goFlagSet.SetOutput(buf)
				goFlagSet.Parse([]string{"--oops"})

				Ω(string(buf.Contents())).Should(Equal("flag provided but not defined: -oops\n" + flagSet.Usage() + "\n"))
			})

			It("includes the go FlagSets flags in their own section", func() {
				expectedUsage := []string{
					"{{red}}{{bold}}{{underline}}Candy Section{{/}}",
					"So sweet.",
					"  {{red}}--string-flag{{/}} [name] {{gray}}(default: Gerald){{/}}",
					"    {{light-gray}}string-usage{{/}}",
					"  {{red}}--int-64-flag{{/}} [int] {{gray}}{{/}}",
					"    {{light-gray}}int-64-usage{{/}}",
					"  {{red}}--bool-flag{{/}} {{gray}}{{/}}",
					"    {{light-gray}}bool-usage{{/}}",
					"",
					"{{blue}}{{bold}}{{underline}}Dairy Section{{/}}",
					"Abbreviated section",
					"  {{blue}}--float-64-flag, --string-slice-flag{{/}}",
					"",
					"  --int-flag{{/}} [int] {{gray}}{{/}}",
					"    {{light-gray}}int-usage{{/}}",
					"",
					"{{bold}}{{underline}}The go flags...{{/}}", //separate go flags section at the bottom, includes flags that are in the go FlagSet but not in the GinkgoFLagSet
					"  -go-int-flag int",
					"    	an integer, please",
					"  -go-string-flag go", //Note the processing of `go` using flag.UnquoteUsage()
					"    	sets via go",
					"",
				}
				Ω(flagSet.Usage()).Should(Equal(strings.Join(expectedUsage, "\n")))
			})
		})
	})

	Describe("GenerateFlagArgs", func() {
		type StructA struct {
			StringProperty   string
			Int64Property    int64
			Float64Property  float64
			UnsupportedInt32 int32
		}
		type StructB struct {
			IntProperty         int
			BoolProperty        bool
			StringSliceProperty []string
			DeprecatedProperty  string
		}
		var A StructA
		var B StructB
		var flags types.GinkgoFlags
		var bindings map[string]interface{}

		BeforeEach(func() {
			A = StructA{
				StringProperty:  "the default string",
				Int64Property:   1138,
				Float64Property: 3.141,
			}
			B = StructB{
				IntProperty:         2009,
				BoolProperty:        true,
				StringSliceProperty: []string{"once", "upon", "a time"},
				DeprecatedProperty:  "n/a",
			}
			bindings = map[string]interface{}{
				"A": &A,
				"B": &B,
			}
			flags = types.GinkgoFlags{
				{Name: "string-flag", KeyPath: "A.StringProperty", DeprecatedName: "stringFlag"},
				{Name: "int-64-flag", KeyPath: "A.Int64Property"},
				{Name: "float-64-flag", KeyPath: "A.Float64Property"},
				{Name: "int-flag", KeyPath: "B.IntProperty", ExportAs: "alias-int-flag"},
				{Name: "bool-flag", KeyPath: "B.BoolProperty", ExportAs: "alias-bool-flag"},
				{Name: "string-slice-flag", KeyPath: "B.StringSliceProperty"},
				{DeprecatedName: "deprecated-flag", KeyPath: "B.DeprecatedProperty"},
			}
		})

		It("generates an array of flag arguments that, if parsed, reproduce the values in the passed-in bindings", func() {
			args, err := types.GenerateFlagArgs(flags, bindings)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(args).Should(Equal([]string{
				"--string-flag=the default string",
				"--int-64-flag=1138",
				"--float-64-flag=3.141000",
				"--alias-int-flag=2009",
				"--alias-bool-flag",
				"--string-slice-flag=once",
				"--string-slice-flag=upon",
				"--string-slice-flag=a time",
			}))
		})

		It("errors if there is a keypath issue", func() {
			flags[0] = types.GinkgoFlag{Name: "unsupported-type", KeyPath: "A.UnsupportedInt32"}
			args, err := types.GenerateFlagArgs(flags, bindings)
			Ω(err).Should(MatchError("unsupported type int32"))
			Ω(args).Should(BeEmpty())

			flags[0] = types.GinkgoFlag{Name: "bad-keypath", KeyPath: "A.StringProoperty"}
			args, err = types.GenerateFlagArgs(flags, bindings)
			Ω(err).Should(MatchError("could not load KeyPath: A.StringProoperty"))
			Ω(args).Should(BeEmpty())
		})
	})
})

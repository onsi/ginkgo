package types_test

import (
	"os"
	"reflect"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/formatter"
	"github.com/onsi/ginkgo/v2/types"
)

var _ = Describe("Deprecation Support", func() {
	Describe("Tracking Deprecations", func() {
		var tracker *types.DeprecationTracker

		BeforeEach(func() {
			tracker = types.NewDeprecationTracker()
			formatter.SingletonFormatter.ColorMode = formatter.ColorModePassthrough
		})

		AfterEach(func() {
			formatter.SingletonFormatter.ColorMode = formatter.ColorModeTerminal
		})

		Context("with no tracked deprecations", func() {
			It("reports no tracked deprecations", func() {
				Ω(tracker.DidTrackDeprecations()).Should(BeFalse())
			})
		})

		Context("with tracked dependencies", func() {
			BeforeEach(func() {
				tracker.TrackDeprecation(types.Deprecation{
					Message: "Deprecation 1",
					DocLink: "doclink-1",
				}, types.CodeLocation{FileName: "foo.go", LineNumber: 17})
				tracker.TrackDeprecation(types.Deprecation{
					Message: "Deprecation 1",
					DocLink: "doclink-1",
				}, types.CodeLocation{FileName: "bar.go", LineNumber: 30})
				tracker.TrackDeprecation(types.Deprecation{
					Message: "Deprecation 2",
					DocLink: "doclink-2",
				})
				tracker.TrackDeprecation(types.Deprecation{
					Message: "Deprecation 3",
				}, types.CodeLocation{FileName: "baz.go", LineNumber: 72})
			})

			It("reports tracked deprecations", func() {
				Ω(tracker.DidTrackDeprecations()).Should(BeTrue())
			})

			It("generates a nicely formatted report", func() {
				report := tracker.DeprecationsReport()
				Ω(report).Should(HavePrefix("{{light-yellow}}You're using deprecated Ginkgo functionality:{{/}}\n{{light-yellow}}============================================={{/}}\n"))
				Ω(report).Should(ContainSubstring(strings.Join([]string{
					"  {{yellow}}Deprecation 1{{/}}",
					"  {{bold}}Learn more at:{{/}} {{cyan}}{{underline}}https://onsi.github.io/ginkgo/MIGRATING_TO_V2#doclink-1{{/}}",
					"    {{gray}}foo.go:17{{/}}",
					"    {{gray}}bar.go:30{{/}}",
					"",
				}, "\n")))
				Ω(report).Should(ContainSubstring(strings.Join([]string{
					"  {{yellow}}Deprecation 2{{/}}",
					"  {{bold}}Learn more at:{{/}} {{cyan}}{{underline}}https://onsi.github.io/ginkgo/MIGRATING_TO_V2#doclink-2{{/}}",
					"",
				}, "\n")))
				Ω(report).Should(ContainSubstring(strings.Join([]string{
					"  {{yellow}}Deprecation 3{{/}}",
					"    {{gray}}baz.go:72{{/}}",
				}, "\n")))
			})

			It("validates that all deprecations point to working documentation", func() {
				v := reflect.ValueOf(types.Deprecations)
				Ω(v.NumMethod()).Should(BeNumerically(">", 0))
				for i := 0; i < v.NumMethod(); i += 1 {
					m := v.Method(i)
					deprecation := m.Call([]reflect.Value{})[0].Interface().(types.Deprecation)

					if deprecation.DocLink != "" {
						Ω(anchors.DocAnchors["MIGRATING_TO_V2.md"]).Should(ContainElement(deprecation.DocLink))
					}
				}
			})
		})

		Context("when ACK_GINKGO_DEPRECATIONS is set", func() {
			var origEnv string
			BeforeEach(func() {
				origEnv = os.Getenv("ACK_GINKGO_DEPRECATIONS")
				os.Setenv("ACK_GINKGO_DEPRECATIONS", "v1.18.3-boop")
			})

			AfterEach(func() {
				os.Setenv("ACK_GINKGO_DEPRECATIONS", origEnv)
			})

			It("does not track deprecations with lower version numbers", func() {
				tracker.TrackDeprecation(types.Deprecation{Message: "Deprecation A", Version: "0.19.2"})
				tracker.TrackDeprecation(types.Deprecation{Message: "Deprecation B", Version: "1.17.4"})
				tracker.TrackDeprecation(types.Deprecation{Message: "Deprecation C", Version: "1.18.2"})
				tracker.TrackDeprecation(types.Deprecation{Message: "Deprecation D", Version: "1.18.3"})
				tracker.TrackDeprecation(types.Deprecation{Message: "Deprecation E", Version: "1.18.4"})
				tracker.TrackDeprecation(types.Deprecation{Message: "Deprecation F", Version: "1.19.2"})
				tracker.TrackDeprecation(types.Deprecation{Message: "Deprecation G", Version: "2.0.0"})
				tracker.TrackDeprecation(types.Deprecation{Message: "Deprecation H"})

				report := tracker.DeprecationsReport()
				Ω(report).ShouldNot(ContainSubstring("Deprecation A"))
				Ω(report).ShouldNot(ContainSubstring("Deprecation B"))
				Ω(report).ShouldNot(ContainSubstring("Deprecation C"))
				Ω(report).ShouldNot(ContainSubstring("Deprecation D"))
				Ω(report).Should(ContainSubstring("Deprecation E"))
				Ω(report).Should(ContainSubstring("Deprecation F"))
				Ω(report).Should(ContainSubstring("Deprecation G"))
				Ω(report).Should(ContainSubstring("Deprecation H"))

				Ω(report).Should(ContainSubstring("ACK_GINKGO_DEPRECATIONS="))
			})
		})
	})
})

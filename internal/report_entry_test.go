package internal_test

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

type SomeStruct struct {
	Label string
	Count int
}

type StringerStruct struct {
	Label string
	Count int
}

func (s StringerStruct) String() string {
	return fmt.Sprintf("%s %d", s.Label, s.Count)
}

type ColorableStringerStruct struct {
	Label string
	Count int
}

func (s ColorableStringerStruct) String() string {
	return fmt.Sprintf("%s %d", s.Label, s.Count)
}

func (s ColorableStringerStruct) ColorableString() string {
	return fmt.Sprintf("{{red}}%s {{green}}%d{{/}}", s.Label, s.Count)
}

func reportEntryJSONRoundTrip(reportEntry internal.ReportEntry) internal.ReportEntry {
	data, err := json.Marshal(reportEntry)
	ExpectWithOffset(1, err).ShouldNot(HaveOccurred())
	var out internal.ReportEntry
	ExpectWithOffset(1, json.Unmarshal(data, &out)).Should(Succeed())
	return out
}

var _ = Describe("ReportEntry and ReportEntries", func() {
	var reportEntry internal.ReportEntry
	var err error

	Describe("ReportEntry with no passed-in value", func() {
		BeforeEach(func() {
			reportEntry, err = internal.NewReportEntry("name", cl)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns a correctly configured ReportEntry", func() {
			Ω(reportEntry.Visibility).Should(Equal(types.ReportEntryVisibilityAlways))
			Ω(reportEntry.Name).Should(Equal("name"))
			Ω(reportEntry.Time).Should(BeTemporally("~", time.Now(), time.Second))
			Ω(reportEntry.Location).Should(Equal(cl))
			Ω(reportEntry.GetRawValue()).Should(BeNil())
		})

		It("has an empty StringRepresentation", func() {
			Ω(reportEntry.StringRepresentation()).Should(BeZero())
		})

		It("round-trips through JSON correctly", func() {
			rtEntry := reportEntryJSONRoundTrip(reportEntry)
			Ω(rtEntry.Visibility).Should(Equal(types.ReportEntryVisibilityAlways))
			Ω(rtEntry.Name).Should(Equal("name"))
			Ω(rtEntry.Time).Should(BeTemporally("~", time.Now(), time.Second))
			Ω(rtEntry.Location).Should(Equal(cl))
			Ω(rtEntry.GetRawValue()).Should(BeNil())
			Ω(rtEntry.StringRepresentation()).Should(BeZero())
		})
	})

	Context("with a string passed-in value", func() {
		BeforeEach(func() {
			reportEntry, err = internal.NewReportEntry("name", cl, "bob")
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns a correctly configured ReportEntry", func() {
			Ω(reportEntry.GetRawValue()).Should(Equal("bob"))
		})

		It("has the correct StringRepresentation", func() {
			Ω(reportEntry.StringRepresentation()).Should(Equal("bob"))
		})

		It("round-trips through JSON correctly", func() {
			rtEntry := reportEntryJSONRoundTrip(reportEntry)
			Ω(rtEntry.GetRawValue()).Should(Equal("bob"))
			Ω(rtEntry.StringRepresentation()).Should(Equal("bob"))
		})
	})

	Context("with a numerical passed-in value", func() {
		BeforeEach(func() {
			reportEntry, err = internal.NewReportEntry("name", cl, 17)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns a correctly configured ReportEntry", func() {
			Ω(reportEntry.GetRawValue()).Should(Equal(17))
		})

		It("has the correct StringRepresentation", func() {
			Ω(reportEntry.StringRepresentation()).Should(Equal("17"))
		})

		It("round-trips through JSON correctly", func() {
			rtEntry := reportEntryJSONRoundTrip(reportEntry)
			Ω(rtEntry.GetRawValue()).Should(Equal(float64(17)))
			Ω(rtEntry.StringRepresentation()).Should(Equal("17"))
		})
	})

	Context("with a struct passed-in value", func() {
		BeforeEach(func() {
			reportEntry, err = internal.NewReportEntry("name", cl, SomeStruct{"bob", 17})
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns a correctly configured ReportEntry", func() {
			Ω(reportEntry.GetRawValue()).Should(Equal(SomeStruct{"bob", 17}))
		})

		It("has the correct StringRepresentation", func() {
			Ω(reportEntry.StringRepresentation()).Should(Equal("{Label:bob Count:17}"))
		})

		It("round-trips through JSON correctly", func() {
			rtEntry := reportEntryJSONRoundTrip(reportEntry)
			Ω(rtEntry.GetRawValue()).Should(Equal(map[string]any{"Label": "bob", "Count": float64(17)}))
			Ω(rtEntry.StringRepresentation()).Should(Equal("{Label:bob Count:17}"))
		})

		It("can be rehydrated into the correct struct, manually", func() {
			rtEntry := reportEntryJSONRoundTrip(reportEntry)
			var s SomeStruct
			Ω(json.Unmarshal([]byte(rtEntry.Value.AsJSON), &s)).Should(Succeed())
			Ω(s).Should(Equal(SomeStruct{"bob", 17}))
		})
	})

	Context("with a stringer passed-in value", func() {
		BeforeEach(func() {
			reportEntry, err = internal.NewReportEntry("name", cl, StringerStruct{"bob", 17})
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns a correctly configured ReportEntry", func() {
			Ω(reportEntry.GetRawValue()).Should(Equal(StringerStruct{"bob", 17}))
		})

		It("has the correct StringRepresentation", func() {
			Ω(reportEntry.StringRepresentation()).Should(Equal("bob 17"))
		})

		It("round-trips through JSON correctly", func() {
			rtEntry := reportEntryJSONRoundTrip(reportEntry)
			Ω(rtEntry.GetRawValue()).Should(Equal(map[string]any{"Label": "bob", "Count": float64(17)}))
			Ω(rtEntry.StringRepresentation()).Should(Equal("bob 17"))
		})
	})

	Context("with a ColorableStringer passed-in value", func() {
		BeforeEach(func() {
			reportEntry, err = internal.NewReportEntry("name", cl, ColorableStringerStruct{"bob", 17})
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns a correctly configured ReportEntry", func() {
			Ω(reportEntry.GetRawValue()).Should(Equal(ColorableStringerStruct{"bob", 17}))
		})

		It("has the correct StringRepresentation", func() {
			Ω(reportEntry.StringRepresentation()).Should(Equal("{{red}}bob {{green}}17{{/}}"))
		})

		It("round-trips through JSON correctly", func() {
			rtEntry := reportEntryJSONRoundTrip(reportEntry)
			Ω(rtEntry.GetRawValue()).Should(Equal(map[string]any{"Label": "bob", "Count": float64(17)}))
			Ω(rtEntry.StringRepresentation()).Should(Equal("{{red}}bob {{green}}17{{/}}"))
		})
	})

	Context("with multiple passed-in values", func() {
		It("errors", func() {
			reportEntry, err = internal.NewReportEntry("name", cl, 1, "2")
			Ω(err).Should(MatchError(types.GinkgoErrors.TooManyReportEntryValues(cl, "2")))
		})
	})

	Context("with the Offset decoration", func() {
		It("computes a new offset code location", func() {
			reportEntry, err = internal.NewReportEntry("name", cl, Offset(1))
			Ω(reportEntry.GetRawValue()).Should(BeNil())
			expectedCL := types.NewCodeLocation(2) // NewReportEntry has a BaseOffset of 2
			Ω(reportEntry.Location.FileName).Should(Equal(expectedCL.FileName))
		})
	})
	Context("with a CodeLocation", func() {
		It("uses the passed-in codelocation", func() {
			customCl := types.NewCustomCodeLocation("foo")
			reportEntry, err = internal.NewReportEntry("name", cl, customCl)
			Ω(reportEntry.GetRawValue()).Should(BeNil())
			Ω(reportEntry.Location).Should(Equal(customCl))
		})
	})
	Context("with a ReportEntryVisibility", func() {
		It("uses the passed in visibility", func() {
			reportEntry, err = internal.NewReportEntry("name", cl, types.ReportEntryVisibilityFailureOrVerbose)
			Ω(reportEntry.GetRawValue()).Should(BeNil())
			Ω(reportEntry.Visibility).Should(Equal(types.ReportEntryVisibilityFailureOrVerbose))
		})
	})

	Context("with a time", func() {
		It("uses the passed in time", func() {
			t := time.Date(1984, 3, 7, 0, 0, 0, 0, time.Local)
			reportEntry, err = internal.NewReportEntry("name", cl, t)
			Ω(reportEntry.GetRawValue()).Should(BeNil())
			Ω(reportEntry.Time).Should(Equal(t))
		})
	})

	Describe("ReportEntries.HasVisibility", func() {
		It("is true when the ReportEntries have the requested visibilities", func() {
			entries := types.ReportEntries{
				types.ReportEntry{Visibility: types.ReportEntryVisibilityAlways},
				types.ReportEntry{Visibility: types.ReportEntryVisibilityAlways},
			}

			Ω(entries.HasVisibility(types.ReportEntryVisibilityNever, types.ReportEntryVisibilityAlways)).Should(BeTrue())
			Ω(entries.HasVisibility(types.ReportEntryVisibilityNever, types.ReportEntryVisibilityFailureOrVerbose)).Should(BeFalse())
		})
	})

	Describe("ReportEntries.WithVisibility", func() {
		It("returns the subset of report entries with the requested visibilities", func() {
			entries := types.ReportEntries{
				types.ReportEntry{Name: "A", Visibility: types.ReportEntryVisibilityAlways},
				types.ReportEntry{Name: "B", Visibility: types.ReportEntryVisibilityFailureOrVerbose},
				types.ReportEntry{Name: "C", Visibility: types.ReportEntryVisibilityNever},
			}
			Ω(entries.WithVisibility(types.ReportEntryVisibilityAlways, types.ReportEntryVisibilityFailureOrVerbose)).Should(Equal(
				types.ReportEntries{
					types.ReportEntry{Name: "A", Visibility: types.ReportEntryVisibilityAlways},
					types.ReportEntry{Name: "B", Visibility: types.ReportEntryVisibilityFailureOrVerbose},
				},
			))

		})
	})

	Describe("mini-integration test - validating that the DSL correctly wires into the suite", func() {
		Context("when passed a value", func() {
			It("works!", func() {
				AddReportEntry("A Test ReportEntry", ColorableStringerStruct{"bob", 17}, types.ReportEntryVisibilityFailureOrVerbose)
			})

			ReportAfterEach(func(report SpecReport) {
				config, _ := GinkgoConfiguration()
				if !config.DryRun && report.State.Is(types.SpecStatePassed) {
					Ω(report.ReportEntries[0].StringRepresentation()).Should(Equal("{{red}}bob {{green}}17{{/}}"))
				}
			})
		})

		Context("when passed a pointer that subsequently changes", func() {
			var obj *ColorableStringerStruct

			BeforeEach(func() {
				obj = &ColorableStringerStruct{"bob", 17}
			})

			It("works!", func() {
				AddReportEntry("A Test ReportEntry", obj, types.ReportEntryVisibilityFailureOrVerbose)
			})

			AfterEach(func() {
				obj.Label = "alice"
				obj.Count = 42
			})

			ReportAfterEach(func(report SpecReport) {
				config, _ := GinkgoConfiguration()
				if !config.DryRun && report.State.Is(types.SpecStatePassed) {
					Ω(report.ReportEntries[0].StringRepresentation()).Should(Equal("{{red}}alice {{green}}42{{/}}"))
				}
			})
		})
	})
})

package internal_test

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/internal"
	"github.com/onsi/ginkgo/types"
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
			Ω(reportEntry.Value).Should(BeNil())
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
			Ω(rtEntry.Representation).Should(Equal(""))
			Ω(rtEntry.Value).Should(BeNil())
			Ω(rtEntry.StringRepresentation()).Should(BeZero())
		})
	})

	Context("with a string passed-in value", func() {
		BeforeEach(func() {
			reportEntry, err = internal.NewReportEntry("name", cl, "bob")
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns a correctly configured ReportEntry", func() {
			Ω(reportEntry.Value).Should(Equal("bob"))
		})

		It("has an empty StringRepresentation", func() {
			Ω(reportEntry.StringRepresentation()).Should(Equal("bob"))
		})

		It("round-trips through JSON correctly", func() {
			rtEntry := reportEntryJSONRoundTrip(reportEntry)
			Ω(rtEntry.Representation).Should(Equal(""))
			Ω(rtEntry.Value).Should(Equal("bob"))
			Ω(rtEntry.StringRepresentation()).Should(Equal("bob"))
		})
	})

	Context("with a numerical passed-in value", func() {
		BeforeEach(func() {
			reportEntry, err = internal.NewReportEntry("name", cl, 17)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns a correctly configured ReportEntry", func() {
			Ω(reportEntry.Value).Should(Equal(17))
		})

		It("has an empty StringRepresentation", func() {
			Ω(reportEntry.StringRepresentation()).Should(Equal("17"))
		})

		It("round-trips through JSON correctly", func() {
			rtEntry := reportEntryJSONRoundTrip(reportEntry)
			Ω(rtEntry.Representation).Should(Equal(""))
			Ω(rtEntry.Value).Should(Equal(float64(17)))
			Ω(rtEntry.StringRepresentation()).Should(Equal("17"))
		})
	})

	Context("with a struct passed-in value", func() {
		BeforeEach(func() {
			reportEntry, err = internal.NewReportEntry("name", cl, SomeStruct{"bob", 17})
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns a correctly configured ReportEntry", func() {
			Ω(reportEntry.Value).Should(Equal(SomeStruct{"bob", 17}))
		})

		It("has an empty StringRepresentation", func() {
			Ω(reportEntry.StringRepresentation()).Should(Equal("{Label:bob Count:17}"))
		})

		It("round-trips through JSON correctly", func() {
			rtEntry := reportEntryJSONRoundTrip(reportEntry)
			Ω(rtEntry.Representation).Should(Equal(""))
			Ω(rtEntry.Value).Should(Equal(map[string]interface{}{"Label": "bob", "Count": float64(17)}))
			Ω(rtEntry.StringRepresentation()).Should(Equal("map[Count:17 Label:bob]"))
		})
	})

	Context("with a stringer passed-in value", func() {
		BeforeEach(func() {
			reportEntry, err = internal.NewReportEntry("name", cl, StringerStruct{"bob", 17})
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns a correctly configured ReportEntry", func() {
			Ω(reportEntry.Value).Should(Equal(StringerStruct{"bob", 17}))
		})

		It("has an empty StringRepresentation", func() {
			Ω(reportEntry.StringRepresentation()).Should(Equal("{{red}}bob {{green}}17{{/}}"))
		})

		It("round-trips through JSON correctly", func() {
			rtEntry := reportEntryJSONRoundTrip(reportEntry)
			Ω(rtEntry.Representation).Should(Equal("{{red}}bob {{green}}17{{/}}"))
			Ω(rtEntry.Value).Should(Equal(map[string]interface{}{"Label": "bob", "Count": float64(17)}))
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
			Ω(reportEntry.Value).Should(BeNil())
			expectedCL := types.NewCodeLocation(2) // NewReportEntry has a BaseOffset of 2
			Ω(reportEntry.Location.FileName).Should(Equal(expectedCL.FileName))
		})
	})
	Context("with a CodeLocation", func() {
		It("uses the passed-in codelocation", func() {
			customCl := types.NewCustomCodeLocation("foo")
			reportEntry, err = internal.NewReportEntry("name", cl, customCl)
			Ω(reportEntry.Value).Should(BeNil())
			Ω(reportEntry.Location).Should(Equal(customCl))
		})
	})
	Context("with a ReportEntryVisibility", func() {
		It("uses the passed in visibility", func() {
			reportEntry, err = internal.NewReportEntry("name", cl, types.ReportEntryVisibilityFailureOnly)
			Ω(reportEntry.Value).Should(BeNil())
			Ω(reportEntry.Visibility).Should(Equal(types.ReportEntryVisibilityFailureOnly))
		})
	})

	Describe("ReportEntries.HasVisibility", func() {
		It("is true when the ReportEntries have the requested visibilities", func() {
			entries := types.ReportEntries{
				types.ReportEntry{Visibility: types.ReportEntryVisibilityAlways},
				types.ReportEntry{Visibility: types.ReportEntryVisibilityAlways},
			}

			Ω(entries.HasVisibility(types.ReportEntryVisibilityNever, types.ReportEntryVisibilityAlways)).Should(BeTrue())
			Ω(entries.HasVisibility(types.ReportEntryVisibilityNever, types.ReportEntryVisibilityFailureOnly)).Should(BeFalse())
		})
	})

	Describe("ReportEntries.WithVisibility", func() {
		It("returns the subset of report entries with the requested visibilities", func() {
			entries := types.ReportEntries{
				types.ReportEntry{Name: "A", Visibility: types.ReportEntryVisibilityAlways},
				types.ReportEntry{Name: "B", Visibility: types.ReportEntryVisibilityFailureOnly},
				types.ReportEntry{Name: "C", Visibility: types.ReportEntryVisibilityNever},
			}
			Ω(entries.WithVisibility(types.ReportEntryVisibilityAlways, types.ReportEntryVisibilityFailureOnly)).Should(Equal(
				types.ReportEntries{
					types.ReportEntry{Name: "A", Visibility: types.ReportEntryVisibilityAlways},
					types.ReportEntry{Name: "B", Visibility: types.ReportEntryVisibilityFailureOnly},
				},
			))

		})
	})

	Describe("mini-integration test - validating that the DSL correctly wires into the suite", func() {
		Context("when passed a value", func() {
			It("works!", func() {
				AddReportEntry("A Test ReportEntry", StringerStruct{"bob", 17}, types.ReportEntryVisibilityFailureOnly)
			})

			ReportAfterEach(func(report SpecReport) {
				Ω(report.ReportEntries[0].StringRepresentation()).Should(Equal("{{red}}bob {{green}}17{{/}}"))
			})
		})
		Context("when passed a pointer that subsequently changes", func() {
			var obj *StringerStruct

			BeforeEach(func() {
				obj = &StringerStruct{"bob", 17}
			})

			It("works!", func() {
				AddReportEntry("A Test ReportEntry", obj, types.ReportEntryVisibilityFailureOnly)
			})

			AfterEach(func() {
				obj.Label = "alice"
				obj.Count = 42
			})

			ReportAfterEach(func(report SpecReport) {
				Ω(report.ReportEntries[0].StringRepresentation()).Should(Equal("{{red}}alice {{green}}42{{/}}"))
			})
		})
	})
})

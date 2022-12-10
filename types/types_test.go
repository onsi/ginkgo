package types_test

import (
	"encoding/json"
	"sort"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/types"
)

var _ = Describe("Types", func() {
	Describe("Report", func() {
		Describe("Add", func() {
			It("concatenates spec reports, combines success, and computes a new RunTime", func() {
				t := time.Now()
				reportA := types.Report{
					SuitePath:                  "foo",
					SuiteSucceeded:             true,
					StartTime:                  t.Add(-time.Minute),
					EndTime:                    t.Add(2 * time.Minute),
					SpecialSuiteFailureReasons: []string{"blame jim", "blame alice"},
					SpecReports: types.SpecReports{
						types.SpecReport{NumAttempts: 3},
						types.SpecReport{NumAttempts: 4},
					},
				}

				reportB := types.Report{
					SuitePath:                  "bar",
					SuiteSucceeded:             false,
					StartTime:                  t.Add(-2 * time.Minute),
					EndTime:                    t.Add(time.Minute),
					SpecialSuiteFailureReasons: []string{"blame bob", "blame jim"},
					SpecReports: types.SpecReports{
						types.SpecReport{NumAttempts: 5},
						types.SpecReport{NumAttempts: 6},
					},
				}

				composite := reportA.Add(reportB)
				Ω(composite).Should(Equal(types.Report{
					SuitePath:                  "foo",
					SuiteSucceeded:             false,
					StartTime:                  t.Add(-2 * time.Minute),
					EndTime:                    t.Add(2 * time.Minute),
					RunTime:                    4 * time.Minute,
					SpecialSuiteFailureReasons: []string{"blame jim", "blame alice", "blame bob"},
					SpecReports: types.SpecReports{
						types.SpecReport{NumAttempts: 3},
						types.SpecReport{NumAttempts: 4},
						types.SpecReport{NumAttempts: 5},
						types.SpecReport{NumAttempts: 6},
					},
				}))

			})
		})
	})

	Describe("ProgressReport", func() {
		It("can return the correct subset of Goroutines when asked", func() {
			specGoroutine := types.Goroutine{ID: 7, IsSpecGoroutine: true, Stack: []types.FunctionCall{{Highlight: true}}}
			highlightedGoroutineA := types.Goroutine{ID: 8, Stack: []types.FunctionCall{{Highlight: true}}}
			highlightedGoroutineB := types.Goroutine{ID: 9, Stack: []types.FunctionCall{{Highlight: true}}}
			otherGoroutineA := types.Goroutine{ID: 10, Stack: []types.FunctionCall{{Highlight: false}}}
			otherGoroutineB := types.Goroutine{ID: 11, Stack: []types.FunctionCall{{Highlight: false}}}

			pr := types.ProgressReport{
				Goroutines: []types.Goroutine{
					otherGoroutineA,
					highlightedGoroutineA,
					specGoroutine,
					highlightedGoroutineB,
					otherGoroutineB,
				},
			}

			Ω(pr.SpecGoroutine()).Should(Equal(specGoroutine))
			Ω(pr.HighlightedGoroutines()).Should(Equal([]types.Goroutine{highlightedGoroutineA, highlightedGoroutineB}))
			Ω(pr.OtherGoroutines()).Should(Equal([]types.Goroutine{otherGoroutineA, otherGoroutineB}))
		})

		It("can return a copy sans GinkgoWriter output", func() {

			pr := types.ProgressReport{
				LeafNodeText:               "hi",
				CapturedGinkgoWriterOutput: "foo",
				TimelineLocation:           types.TimelineLocation{Offset: 10},
			}

			Ω(pr.WithoutCapturedGinkgoWriterOutput()).Should(Equal(types.ProgressReport{LeafNodeText: "hi", TimelineLocation: types.TimelineLocation{Offset: 10}}))

		})
	})

	Describe("NodeType", func() {
		Describe("Is", func() {
			It("returns true when the NodeType is in the passed-in list", func() {
				Ω(types.NodeTypeContainer.Is(types.NodeTypeIt | types.NodeTypeContainer)).Should(BeTrue())
			})

			It("returns false when the NodeType is not in the passed-in list", func() {
				Ω(types.NodeTypeContainer.Is(types.NodeTypeIt | types.NodeTypeBeforeEach)).Should(BeFalse())
			})
		})

		DescribeTable("Representation and Encoding", func(nodeType types.NodeType, expectedString string) {
			Ω(nodeType.String()).Should(Equal(expectedString))

			marshalled, err := json.Marshal(nodeType)
			Ω(err).ShouldNot(HaveOccurred())
			var unmarshalled types.NodeType
			json.Unmarshal(marshalled, &unmarshalled)
			Ω(unmarshalled).Should(Equal(nodeType))
		},
			func(nodeType types.NodeType, expectedString string) string {
				return expectedString
			},
			Entry(nil, types.NodeTypeContainer, "Container"),
			Entry(nil, types.NodeTypeIt, "It"),
			Entry(nil, types.NodeTypeBeforeEach, "BeforeEach"),
			Entry(nil, types.NodeTypeJustBeforeEach, "JustBeforeEach"),
			Entry(nil, types.NodeTypeAfterEach, "AfterEach"),
			Entry(nil, types.NodeTypeJustAfterEach, "JustAfterEach"),
			Entry(nil, types.NodeTypeBeforeAll, "BeforeAll"),
			Entry(nil, types.NodeTypeAfterAll, "AfterAll"),
			Entry(nil, types.NodeTypeBeforeSuite, "BeforeSuite"),
			Entry(nil, types.NodeTypeSynchronizedBeforeSuite, "SynchronizedBeforeSuite"),
			Entry(nil, types.NodeTypeAfterSuite, "AfterSuite"),
			Entry(nil, types.NodeTypeSynchronizedAfterSuite, "SynchronizedAfterSuite"),
			Entry(nil, types.NodeTypeReportBeforeEach, "ReportBeforeEach"),
			Entry(nil, types.NodeTypeReportAfterEach, "ReportAfterEach"),
			Entry(nil, types.NodeTypeReportBeforeSuite, "ReportBeforeSuite"),
			Entry(nil, types.NodeTypeReportAfterSuite, "ReportAfterSuite"),
			Entry(nil, types.NodeTypeCleanupInvalid, "DeferCleanup"),
			Entry(nil, types.NodeTypeCleanupAfterEach, "DeferCleanup (Each)"),
			Entry(nil, types.NodeTypeCleanupAfterAll, "DeferCleanup (All)"),
			Entry(nil, types.NodeTypeCleanupAfterSuite, "DeferCleanup (Suite)"),
			Entry(nil, types.NodeTypeInvalid, "INVALID NODE TYPE"),
		)
	})

	Describe("FailureNodeContext", func() {
		DescribeTable("Representation and Encoding", func(context types.FailureNodeContext) {
			marshalled, err := json.Marshal(context)
			Ω(err).ShouldNot(HaveOccurred())
			var unmarshalled types.FailureNodeContext
			json.Unmarshal(marshalled, &unmarshalled)
			Ω(unmarshalled).Should(Equal(context))
		},
			Entry("LeafNode", types.FailureNodeIsLeafNode),
			Entry("TopLevel", types.FailureNodeAtTopLevel),
			Entry("InContainer", types.FailureNodeInContainer),
			Entry("Invalid", types.FailureNodeContextInvalid),
		)
	})

	Describe("SpecState", func() {
		DescribeTable("Representation and Encoding", func(specState types.SpecState, expectedString string) {
			Ω(specState.String()).Should(Equal(expectedString))

			marshalled, err := json.Marshal(specState)
			Ω(err).ShouldNot(HaveOccurred())
			var unmarshalled types.SpecState
			json.Unmarshal(marshalled, &unmarshalled)
			Ω(unmarshalled).Should(Equal(specState))
		},
			Entry("Pending", types.SpecStatePending, "pending"),
			Entry("Skipped", types.SpecStateSkipped, "skipped"),
			Entry("Passed", types.SpecStatePassed, "passed"),
			Entry("Failed", types.SpecStateFailed, "failed"),
			Entry("Panicked", types.SpecStatePanicked, "panicked"),
			Entry("Aborted", types.SpecStateAborted, "aborted"),
			Entry("Interrupted", types.SpecStateInterrupted, "interrupted"),
			Entry("Invalid", types.SpecStateInvalid, "INVALID SPEC STATE"),
		)
	})

	Describe("SpecReport Helper Functions", func() {
		Describe("CombinedOutput", func() {
			Context("with no GinkgoWriter or StdOutErr output", func() {
				It("comes back empty", func() {
					Ω(types.SpecReport{}.CombinedOutput()).Should(Equal(""))
				})
			})

			Context("with only StdOutErr output", func() {
				It("returns that output", func() {
					Ω(types.SpecReport{
						CapturedStdOutErr: "hello",
					}.CombinedOutput()).Should(Equal("hello"))
				})
			})

			Context("with only GinkgoWriter output", func() {
				It("returns that output", func() {
					Ω(types.SpecReport{
						CapturedGinkgoWriterOutput: "hello",
					}.CombinedOutput()).Should(Equal("hello"))
				})
			})

			Context("with both", func() {
				It("returns both concatenated", func() {
					Ω(types.SpecReport{
						CapturedGinkgoWriterOutput: "gw",
						CapturedStdOutErr:          "std",
					}.CombinedOutput()).Should(Equal("std\ngw"))
				})
			})
		})

		Describe("Labels", Label("TestA", "TestB"), func() {
			It("returns a concatenated, deduped, set of labels", Label("TestB", "TestC"), func() {
				Ω(CurrentSpecReport().Labels()).Should(Equal([]string{"TestA", "TestB", "TestC"}))
			})
		})

		Describe("MatchesLabelFilter", Label("dog", "cat"), func() {
			It("returns an error when passed an invalid filter query", func() {
				matches, err := CurrentSpecReport().MatchesLabelFilter("(welp")
				Ω(err).Should(HaveOccurred())
				Ω(matches).Should(BeFalse())
			})

			It("returns whether or not the query matches", Label("catfish"), func() {
				Ω(CurrentSpecReport().MatchesLabelFilter("dog")).Should(BeTrue())
				Ω(CurrentSpecReport().MatchesLabelFilter("cow || cat")).Should(BeTrue())
				Ω(CurrentSpecReport().MatchesLabelFilter("/fish/")).Should(BeTrue())
				Ω(CurrentSpecReport().MatchesLabelFilter("dog && !/fish/")).Should(BeFalse())
			})
		})

		It("can report on whether state is a failed state", func() {
			Ω(types.SpecReport{State: types.SpecStatePending}.Failed()).Should(BeFalse())
			Ω(types.SpecReport{State: types.SpecStateSkipped}.Failed()).Should(BeFalse())
			Ω(types.SpecReport{State: types.SpecStatePassed}.Failed()).Should(BeFalse())
			Ω(types.SpecReport{State: types.SpecStateFailed}.Failed()).Should(BeTrue())
			Ω(types.SpecReport{State: types.SpecStatePanicked}.Failed()).Should(BeTrue())
			Ω(types.SpecReport{State: types.SpecStateAborted}.Failed()).Should(BeTrue())
			Ω(types.SpecReport{State: types.SpecStateInterrupted}.Failed()).Should(BeTrue())
		})

		It("can return a concatenated set of texts", func() {
			Ω(CurrentSpecReport().FullText()).Should(Equal("Types SpecReport Helper Functions can return a concatenated set of texts"))
		})

		It("can return the name of the file it's spec is in", func() {
			cl := types.NewCodeLocation(0)
			Ω(CurrentSpecReport().FileName()).Should(Equal(cl.FileName))
		})

		It("can return the linenumber of the file it's spec is in", func() {
			cl := types.NewCodeLocation(0)
			Ω(CurrentSpecReport().LineNumber()).Should(Equal(cl.LineNumber - 1))
		})

		It("can return it's failure's message", func() {
			report := types.SpecReport{
				Failure: types.Failure{Message: "why this failed"},
			}
			Ω(report.FailureMessage()).Should(Equal("why this failed"))
		})

		It("can return it's failure's code location", func() {
			cl := types.NewCodeLocation(0)
			report := types.SpecReport{
				Failure: types.Failure{Location: cl},
			}
			Ω(report.FailureLocation()).Should(Equal(cl))
		})
	})

	Describe("SpecReports", func() {
		Describe("Encoding to JSON", func() {
			var report types.SpecReport
			BeforeEach(func() {
				report = types.SpecReport{
					ContainerHierarchyTexts: []string{"A", "B"},
					ContainerHierarchyLocations: []types.CodeLocation{
						types.NewCodeLocation(0),
						types.NewCodeLocationWithStackTrace(0),
						types.NewCustomCodeLocation("welp"),
					},
					LeafNodeType:               types.NodeTypeIt,
					LeafNodeLocation:           types.NewCodeLocation(0),
					LeafNodeText:               "C",
					State:                      types.SpecStateFailed,
					StartTime:                  time.Date(2012, 06, 19, 05, 32, 12, 0, time.UTC),
					EndTime:                    time.Date(2012, 06, 19, 05, 33, 12, 0, time.UTC),
					RunTime:                    time.Minute,
					ParallelProcess:            2,
					NumAttempts:                3,
					CapturedGinkgoWriterOutput: "gw",
					CapturedStdOutErr:          "std",
					Failure: types.Failure{
						Message:                   "boom",
						Location:                  types.NewCodeLocation(1),
						ForwardedPanic:            "bam",
						FailureNodeContext:        types.FailureNodeInContainer,
						FailureNodeType:           types.NodeTypeBeforeEach,
						FailureNodeLocation:       types.NewCodeLocation(0),
						FailureNodeContainerIndex: 1,
					},
				}
			})

			Context("with a failure", func() {
				It("round-trips correctly", func() {
					marshalled, err := json.Marshal(report)
					Ω(err).ShouldNot(HaveOccurred())
					unmarshalled := types.SpecReport{}
					err = json.Unmarshal(marshalled, &unmarshalled)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(unmarshalled).Should(Equal(report))
				})
			})

			Context("without a failure", func() {
				BeforeEach(func() {
					report.Failure = types.Failure{}
				})
				It("round-trips correctly and doesn't include the Failure struct", func() {
					marshalled, err := json.Marshal(report)
					Ω(string(marshalled)).ShouldNot(ContainSubstring("Failure"))
					Ω(err).ShouldNot(HaveOccurred())
					unmarshalled := types.SpecReport{}
					err = json.Unmarshal(marshalled, &unmarshalled)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(unmarshalled).Should(Equal(report))
				})
			})
		})

		Describe("WithLeafNodeType", func() {
			It("returns reports with the matching LeafNodeTypes", func() {
				reports := types.SpecReports{
					{LeafNodeType: types.NodeTypeIt, NumAttempts: 2},
					{LeafNodeType: types.NodeTypeIt, NumAttempts: 3},
					{LeafNodeType: types.NodeTypeBeforeSuite, NumAttempts: 4},
					{LeafNodeType: types.NodeTypeAfterSuite, NumAttempts: 5},
					{LeafNodeType: types.NodeTypeSynchronizedAfterSuite, NumAttempts: 6},
				}

				Ω(reports.WithLeafNodeType(types.NodeTypeIt | types.NodeTypeAfterSuite)).Should(Equal(types.SpecReports{
					{LeafNodeType: types.NodeTypeIt, NumAttempts: 2},
					{LeafNodeType: types.NodeTypeIt, NumAttempts: 3},
					{LeafNodeType: types.NodeTypeAfterSuite, NumAttempts: 5},
				}))
			})
		})

		Describe("WithState", func() {
			It("returns reports with the matching SpecStates", func() {
				reports := types.SpecReports{
					{State: types.SpecStatePassed, NumAttempts: 2},
					{State: types.SpecStatePassed, NumAttempts: 3},
					{State: types.SpecStateFailed, NumAttempts: 4},
					{State: types.SpecStatePending, NumAttempts: 5},
					{State: types.SpecStateSkipped, NumAttempts: 6},
				}

				Ω(reports.WithState(types.SpecStatePassed | types.SpecStatePending)).Should(Equal(types.SpecReports{
					{State: types.SpecStatePassed, NumAttempts: 2},
					{State: types.SpecStatePassed, NumAttempts: 3},
					{State: types.SpecStatePending, NumAttempts: 5},
				}))
			})
		})

		Describe("CountWithState", func() {
			It("returns the number with the matching SpecStates", func() {
				reports := types.SpecReports{
					{State: types.SpecStatePassed, NumAttempts: 2},
					{State: types.SpecStatePassed, NumAttempts: 3},
					{State: types.SpecStateFailed, NumAttempts: 4},
					{State: types.SpecStatePending, NumAttempts: 5},
					{State: types.SpecStateSkipped, NumAttempts: 6},
				}

				Ω(reports.CountWithState(types.SpecStatePassed | types.SpecStatePending)).Should(Equal(3))
			})
		})

		Describe("CountOfFlakedSpecs", func() {
			It("returns the number of passing specs with NumAttempts > 1", func() {
				reports := types.SpecReports{
					{State: types.SpecStatePassed, NumAttempts: 2, MaxFlakeAttempts: 2},
					{State: types.SpecStatePassed, NumAttempts: 2, MaxFlakeAttempts: 2},
					{State: types.SpecStatePassed, NumAttempts: 1, MaxFlakeAttempts: 2},
					{State: types.SpecStatePassed, NumAttempts: 1, MaxFlakeAttempts: 2},
					{State: types.SpecStateFailed, NumAttempts: 2, MaxFlakeAttempts: 2},
				}

				Ω(reports.CountOfFlakedSpecs()).Should(Equal(2))
			})
		})

		Describe("CountOfRepeatedSpecs", func() {
			It("returns the number of failed specs with NumAttempts > 1", func() {
				reports := types.SpecReports{
					{State: types.SpecStatePassed, NumAttempts: 2, MaxMustPassRepeatedly: 2},
					{State: types.SpecStatePassed, NumAttempts: 2, MaxMustPassRepeatedly: 2},
					{State: types.SpecStatePassed, NumAttempts: 1, MaxMustPassRepeatedly: 2},
					{State: types.SpecStatePassed, NumAttempts: 1, MaxMustPassRepeatedly: 2},
					{State: types.SpecStateFailed, NumAttempts: 2, MaxMustPassRepeatedly: 2},
				}

				Ω(reports.CountOfRepeatedSpecs()).Should(Equal(1))
			})
		})
	})

	Describe("Timelines", func() {
		var report types.SpecReport
		var primaryFailure types.Failure
		var timeoutFailure, panicFailure, embeddedFailure types.AdditionalFailure
		var reportEntryAlways, reportEntryNever types.ReportEntry
		var progressReport1, progressReport2 types.ProgressReport
		var byStartSpecEvent, nodeStartSpecEvent types.SpecEvent

		BeforeEach(func() {
			embeddedFailure = types.AdditionalFailure{
				Failure: types.Failure{
					Message:          "embedded-failure",
					TimelineLocation: types.TimelineLocation{Order: 100},
				},
				State: types.SpecStateFailed,
			}
			primaryFailure = types.Failure{
				Message:           "primary-failure",
				TimelineLocation:  types.TimelineLocation{Order: 80},
				AdditionalFailure: &embeddedFailure,
			}
			timeoutFailure = types.AdditionalFailure{
				Failure: types.Failure{
					Message:          "timeout-failure",
					TimelineLocation: types.TimelineLocation{Order: 120},
					AdditionalFailure: &types.AdditionalFailure{
						Failure: types.Failure{
							Message:          "additional-embeded-failure",
							TimelineLocation: types.TimelineLocation{Order: 125},
						},
						State: types.SpecStateFailed,
					},
				},
				State: types.SpecStateTimedout,
			}
			panicFailure = types.AdditionalFailure{
				Failure: types.Failure{Message: "panic-failure",
					TimelineLocation: types.TimelineLocation{Order: 140},
				},
				State: types.SpecStatePanicked,
			}
			reportEntryAlways = types.ReportEntry{
				Name:             "report-entry-always",
				Visibility:       types.ReportEntryVisibilityAlways,
				TimelineLocation: types.TimelineLocation{Order: 130},
			}
			reportEntryNever = types.ReportEntry{
				Name:             "report-entry-never",
				Visibility:       types.ReportEntryVisibilityNever,
				TimelineLocation: types.TimelineLocation{Order: 50},
			}
			progressReport1 = types.ProgressReport{
				Message:          "progress-report-1",
				TimelineLocation: types.TimelineLocation{Order: 60},
			}
			progressReport2 = types.ProgressReport{
				Message:          "progress-report-2",
				TimelineLocation: types.TimelineLocation{Order: 90},
			}
			byStartSpecEvent = types.SpecEvent{
				SpecEventType:    types.SpecEventByStart,
				Message:          "by-start-spec-event",
				TimelineLocation: types.TimelineLocation{Order: 95},
			}
			nodeStartSpecEvent = types.SpecEvent{
				SpecEventType:    types.SpecEventNodeStart,
				Message:          "node-start-spec-event",
				NodeType:         types.NodeTypeBeforeEach,
				TimelineLocation: types.TimelineLocation{Order: 30},
			}

			report = types.SpecReport{
				Failure:            primaryFailure,
				AdditionalFailures: []types.AdditionalFailure{timeoutFailure, panicFailure},
				ReportEntries:      []types.ReportEntry{reportEntryAlways, reportEntryNever},
				ProgressReports:    []types.ProgressReport{progressReport1, progressReport2},
				SpecEvents:         []types.SpecEvent{byStartSpecEvent, nodeStartSpecEvent},
			}
		})

		It("can return a timeline comprised of the various TimelineEvents in the spec, that can be sorted by timelinelocation", func() {
			timeline := report.Timeline()
			sort.Sort(timeline)
			Ω(timeline).Should(Equal(types.Timeline{
				nodeStartSpecEvent,
				reportEntryNever,
				progressReport1,
				primaryFailure,
				progressReport2,
				byStartSpecEvent,
				embeddedFailure,
				timeoutFailure,
				reportEntryAlways,
				panicFailure,
			}))
		})

		It("can filter the timeline to remove entries that will not be displayed", func() {
			timeline := report.Timeline()
			sort.Sort(timeline)

			timeline = timeline.WithoutHiddenReportEntries()
			Ω(timeline).Should(Equal(types.Timeline{
				nodeStartSpecEvent,
				progressReport1,
				primaryFailure,
				progressReport2,
				byStartSpecEvent,
				embeddedFailure,
				timeoutFailure,
				reportEntryAlways,
				panicFailure,
			}))

			timeline = timeline.WithoutVeryVerboseSpecEvents()
			Ω(timeline).Should(Equal(types.Timeline{
				progressReport1,
				primaryFailure,
				progressReport2,
				byStartSpecEvent,
				embeddedFailure,
				timeoutFailure,
				reportEntryAlways,
				panicFailure,
			}))
		})
	})

	Describe("SpecEvent", func() {
		DescribeTable("IsOnlyVisibleAtVeryVerbose", func(specEventType types.SpecEventType, isOnlyVisibleAtVeryVerbose bool) {
			Ω(types.SpecEvent{SpecEventType: specEventType}.IsOnlyVisibleAtVeryVerbose()).Should(Equal(isOnlyVisibleAtVeryVerbose))
		},
			Entry(nil, types.SpecEventByStart, false),
			Entry(nil, types.SpecEventByEnd, true),
			Entry(nil, types.SpecEventNodeStart, true),
			Entry(nil, types.SpecEventNodeEnd, true),
			Entry(nil, types.SpecEventSpecRepeat, false),
			Entry(nil, types.SpecEventSpecRetry, false),
		)

		DescribeTable("SpecEventType: Representation and Encoding", func(specEventType types.SpecEventType, expectedString string) {
			Ω(specEventType.String()).Should(Equal(expectedString))

			marshalled, err := json.Marshal(specEventType)
			Ω(err).ShouldNot(HaveOccurred())
			var unmarshalled types.SpecEventType
			json.Unmarshal(marshalled, &unmarshalled)
			Ω(unmarshalled).Should(Equal(specEventType))
		},
			Entry(nil, types.SpecEventInvalid, "INVALID SPEC EVENT"),
			Entry(nil, types.SpecEventByStart, "By"),
			Entry(nil, types.SpecEventByEnd, "By (End)"),
			Entry(nil, types.SpecEventNodeStart, "Node"),
			Entry(nil, types.SpecEventNodeEnd, "Node (End)"),
			Entry(nil, types.SpecEventSpecRepeat, "Repeat"),
			Entry(nil, types.SpecEventSpecRetry, "Retry"),
		)
	})
})

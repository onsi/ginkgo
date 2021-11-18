package internal_test

import (
	"fmt"
	"os"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"

	"github.com/onsi/ginkgo/v2/internal"
)

var _ = Describe("UniqueNodeID", func() {
	It("returns a unique id every time it's called", func() {
		Ω(internal.UniqueNodeID()).ShouldNot(Equal(internal.UniqueNodeID()))
	})
})

var _ = Describe("Partitioning Decorations", func() {
	It("separates out decorations and non-decorations", func() {
		type Foo struct {
			A int
		}
		decorations, remaining := internal.PartitionDecorations(
			Offset(3),
			Foo{3},
			types.NewCustomCodeLocation("hey there"),
			"hey there",
			Focus,
			2.0,
			Pending,
			Serial,
			Ordered,
			nil,
			1,
			[]interface{}{Focus, Pending, []interface{}{Offset(2), Serial, FlakeAttempts(2)}, Ordered, Label("a", "b", "c")},
			[]interface{}{1, 2, 3.1, nil},
			[]string{"a", "b", "c"},
			Label("A", "B", "C"),
			Label("D"),
			[]interface{}{},
			FlakeAttempts(1),
			true,
		)

		Ω(decorations).Should(Equal([]interface{}{
			Offset(3),
			types.NewCustomCodeLocation("hey there"),
			Focus,
			Pending,
			Serial,
			Ordered,
			[]interface{}{Focus, Pending, []interface{}{Offset(2), Serial, FlakeAttempts(2)}, Ordered, Label("a", "b", "c")},
			Label("A", "B", "C"),
			Label("D"),
			FlakeAttempts(1),
		}))

		Ω(remaining).Should(Equal([]interface{}{
			Foo{3},
			"hey there",
			2.0,
			nil,
			1,
			[]interface{}{1, 2, 3.1, nil},
			[]string{"a", "b", "c"},
			[]interface{}{},
			true,
		}))
	})
})

var _ = Describe("Combining Labels", func() {
	It("can combine labels and produce the unique union", func() {
		Ω(internal.UnionOfLabels(Label("a", "b", "c"), Label("b", "c", "d"), Label("e", "a", "f"))).Should(Equal(Label("a", "b", "c", "d", "e", "f")))
	})
})

var _ = Describe("Constructing nodes", func() {
	var dt *types.DeprecationTracker
	var didRun bool
	var body func()
	BeforeEach(func() {
		dt = types.NewDeprecationTracker()
		didRun = false
		body = func() { didRun = true }
	})

	ExpectAllWell := func(errors []error) {
		ExpectWithOffset(1, errors).Should(BeEmpty())
		ExpectWithOffset(1, dt.DidTrackDeprecations()).Should(BeFalse())
	}

	Describe("happy path", func() {
		It("creates a node with a non-zero id", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, cl, Focus, Label("A", "B", "C"))
			Ω(node.ID).Should(BeNumerically(">", 0))
			Ω(node.NodeType).Should(Equal(ntIt))
			Ω(node.Text).Should(Equal("text"))
			node.Body()
			Ω(didRun).Should(BeTrue())
			Ω(node.CodeLocation).Should(Equal(cl))
			Ω(node.MarkedFocus).Should(BeTrue())
			Ω(node.MarkedPending).Should(BeFalse())
			Ω(node.NestingLevel).Should(Equal(-1))
			Ω(node.Labels).Should(Equal(Labels{"A", "B", "C"}))
			ExpectAllWell(errors)
		})
	})

	Describe("Assigning CodeLocation", func() {
		Context("with nothing explicitly specified ", func() {
			It("assumes a base-offset of 2", func() {
				cl := types.NewCodeLocation(1)
				node, errors := internal.NewNode(dt, ntIt, "text", body)
				Ω(node.CodeLocation.FileName).Should(Equal(cl.FileName))
				ExpectAllWell(errors)
			})
		})

		Context("specifying code locations", func() {
			It("uses the last passed-in code location", func() {
				cl2 := types.NewCustomCodeLocation("hi")
				node, errors := internal.NewNode(dt, ntIt, "text", body, cl, cl2)
				Ω(node.CodeLocation).Should(Equal(cl2))
				ExpectAllWell(errors)
			})
		})

		Context("specifying offets", func() {
			It("takes the offset and adds it to the base-offset of 2 to compute the code location", func() {
				cl := types.NewCodeLocation(2)
				cl2 := types.NewCustomCodeLocation("hi")
				node, errors := internal.NewNode(dt, ntIt, "text", body, cl2, Offset(1))
				//note that Offset overrides cl2
				Ω(node.CodeLocation.FileName).Should(Equal(cl.FileName))
				ExpectAllWell(errors)
			})
		})
	})

	Describe("ignoring deprecated timeouts", func() {
		It("ignores any float64s", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, 3.141, 2.71)
			node.Body()
			Ω(didRun).Should(BeTrue())
			ExpectAllWell(errors)
		})
	})

	Describe("the Focus and Pending decorations", func() {
		It("the node is neither Focused nor Pending by default", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body)
			Ω(node.MarkedFocus).Should(BeFalse())
			Ω(node.MarkedPending).Should(BeFalse())
			ExpectAllWell(errors)
		})
		It("marks the node as focused", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, Focus)
			Ω(node.MarkedFocus).Should(BeTrue())
			Ω(node.MarkedPending).Should(BeFalse())
			ExpectAllWell(errors)
		})
		It("marks the node as pending", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, Pending)
			Ω(node.MarkedFocus).Should(BeFalse())
			Ω(node.MarkedPending).Should(BeTrue())
			ExpectAllWell(errors)
		})
		It("errors when both Focus and Pending are set", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, cl, Focus, Pending)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidDeclarationOfFocusedAndPending(cl, ntIt)))
		})
		It("allows containers to be marked", func() {
			node, errors := internal.NewNode(dt, ntCon, "text", body, Focus)
			Ω(node.MarkedFocus).Should(BeTrue())
			Ω(node.MarkedPending).Should(BeFalse())
			ExpectAllWell(errors)

			node, errors = internal.NewNode(dt, ntCon, "text", body, Pending)
			Ω(node.MarkedFocus).Should(BeFalse())
			Ω(node.MarkedPending).Should(BeTrue())
			ExpectAllWell(errors)
		})
		It("does not allow non-container/it nodes to be marked", func() {
			node, errors := internal.NewNode(dt, ntBef, "", body, cl, Focus)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidDecoratorForNodeType(cl, ntBef, "Focus")))

			node, errors = internal.NewNode(dt, ntAf, "", body, cl, Pending)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidDecoratorForNodeType(cl, ntAf, "Pending")))

			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})
	})

	Describe("the Serial decoration", func() {
		It("the node is not Serial by default", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body)
			Ω(node.MarkedSerial).Should(BeFalse())
			ExpectAllWell(errors)
		})
		It("marks the node as Serial", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, Serial)
			Ω(node.MarkedSerial).Should(BeTrue())
			ExpectAllWell(errors)
		})
		It("allows containers to be marked", func() {
			node, errors := internal.NewNode(dt, ntCon, "text", body, Serial)
			Ω(node.MarkedSerial).Should(BeTrue())
			ExpectAllWell(errors)
		})
		It("does not allow non-container/it nodes to be marked", func() {
			node, errors := internal.NewNode(dt, ntBef, "", body, cl, Serial)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidDecoratorForNodeType(cl, ntBef, "Serial")))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})
	})

	Describe("the Ordered decoration", func() {
		It("the node is not Ordered by default", func() {
			node, errors := internal.NewNode(dt, ntCon, "", body)
			Ω(node.MarkedOrdered).Should(BeFalse())
			ExpectAllWell(errors)
		})
		It("marks the node as Ordered", func() {
			node, errors := internal.NewNode(dt, ntCon, "", body, Ordered)
			Ω(node.MarkedOrdered).Should(BeTrue())
			ExpectAllWell(errors)
		})
		It("does not allow non-container nodes to be marked", func() {
			node, errors := internal.NewNode(dt, ntBef, "", body, cl, Ordered)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidDecoratorForNodeType(cl, ntBef, "Ordered")))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())

			node, errors = internal.NewNode(dt, ntIt, "not even Its", body, cl, Ordered)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidDecoratorForNodeType(cl, ntIt, "Ordered")))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})
	})

	Describe("The FlakeAttempts decoration", func() {
		It("is zero by default", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body)
			Ω(node).ShouldNot(BeZero())
			Ω(node.FlakeAttempts).Should(Equal(0))
			ExpectAllWell(errors)
		})
		It("sets the FlakeAttempts field", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, FlakeAttempts(2))
			Ω(node.FlakeAttempts).Should(Equal(2))
			ExpectAllWell(errors)
		})
		It("can be applied to containers", func() {
			node, errors := internal.NewNode(dt, ntCon, "text", body, FlakeAttempts(2))
			Ω(node.FlakeAttempts).Should(Equal(2))
			ExpectAllWell(errors)
		})
		It("cannot be applied to non-container/it nodes", func() {
			node, errors := internal.NewNode(dt, ntBef, "", body, cl, FlakeAttempts(2))
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidDecoratorForNodeType(cl, ntBef, "FlakeAttempts")))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})
	})

	Describe("The Label decoration", func() {
		It("has no labels by default", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body)
			Ω(node).ShouldNot(BeZero())
			Ω(node.Labels).Should(Equal(Labels{}))
			ExpectAllWell(errors)
		})

		It("can track labels", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, Label("A", "B", "C"))
			Ω(node.Labels).Should(Equal(Labels{"A", "B", "C"}))
			ExpectAllWell(errors)
		})

		It("appends and dedupes all labels together, even if nested", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, Label("A", "B", "C"), Label("D", "E", "C"), []interface{}{Label("F"), []interface{}{Label("G", "H", "A", "F")}})
			Ω(node.Labels).Should(Equal(Labels{"A", "B", "C", "D", "E", "F", "G", "H"}))
			ExpectAllWell(errors)
		})

		It("can be applied to containers", func() {
			node, errors := internal.NewNode(dt, ntCon, "text", body, Label("A", "B", "C"))
			Ω(node.Labels).Should(Equal(Labels{"A", "B", "C"}))
			ExpectAllWell(errors)
		})

		It("cannot be applied to non-container/it nodes", func() {
			node, errors := internal.NewNode(dt, ntBef, "", body, cl, Label("A", "B", "C"))
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidDecoratorForNodeType(cl, ntBef, "Label")))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})

		It("validates labels", func() {
			node, errors := internal.NewNode(dt, ntIt, "", body, cl, Label("A", "B&C", "C,D", "C,D ", "  "))
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidLabel("B&C", cl), types.GinkgoErrors.InvalidLabel("C,D", cl), types.GinkgoErrors.InvalidLabel("C,D ", cl), types.GinkgoErrors.InvalidEmptyLabel(cl)))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})
	})

	Describe("passing in functions", func() {
		It("works when a single function is passed in", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, cl)
			node.Body()
			Ω(didRun).Should(BeTrue())
			ExpectAllWell(errors)
		})

		It("allows deprecated async functions and registers a deprecation warning", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", func(done Done) {
				didRun = true
				Ω(done).ShouldNot(BeNil())
				close(done)
			}, cl)
			node.Body()
			Ω(didRun).Should(BeTrue())
			Ω(errors).Should(BeEmpty())
			Ω(dt.DeprecationsReport()).Should(ContainSubstring(types.Deprecations.Async().Message))
		})

		It("errors if more than one function is provided", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, body, cl)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.MultipleBodyFunctions(cl, ntIt)))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})

		It("errors if the function has a return value", func() {
			f := func() string { return "" }
			node, errors := internal.NewNode(dt, ntIt, "text", f, cl)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidBodyType(reflect.TypeOf(f), cl, ntIt)))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})

		It("errors if the function takes more than one argument", func() {
			f := func(Done, string) {}
			node, errors := internal.NewNode(dt, ntIt, "text", f, cl)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidBodyType(reflect.TypeOf(f), cl, ntIt)))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})

		It("errors if the function takes one argument and that argument is not the deprecated Done channel", func() {
			f := func(chan interface{}) {}
			node, errors := internal.NewNode(dt, ntIt, "text", f, cl)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidBodyType(reflect.TypeOf(f), cl, ntIt)))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})

		It("errors if no function is passed in", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", cl)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.MissingBodyFunction(cl, ntIt)))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})

		It("is ok if no function is passed in but it is marked pending", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", cl, Pending)
			Ω(node.IsZero()).Should(BeFalse())
			ExpectAllWell(errors)
		})
	})

	Describe("non-recognized decorations", func() {
		It("errors when a non-recognized decoration is provided", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", cl, body, Focus, "aardvark", 5)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(
				types.GinkgoErrors.UnknownDecorator(cl, ntIt, "aardvark"),
				types.GinkgoErrors.UnknownDecorator(cl, ntIt, 5),
			))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})
	})

	Describe("when decorations are nested in slices", func() {
		It("unrolls them first", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", []interface{}{body, []interface{}{Focus, FlakeAttempts(3), Label("A")}, FlakeAttempts(2), Label("B"), Label("C", "D")})
			Ω(node.FlakeAttempts).Should(Equal(2))
			Ω(node.MarkedFocus).Should(BeTrue())
			Ω(node.Labels).Should(Equal(Labels{"A", "B", "C", "D"}))
			node.Body()
			Ω(didRun).Should(BeTrue())
			ExpectAllWell(errors)
		})
	})
})

var _ = Describe("Node", func() {
	Describe("The other node constructors", func() {
		Describe("NewSynchronizedBeforeSuiteNode", func() {
			It("returns a correctly configured node", func() {
				var ranProc1, ranAllProcs bool
				proc1Body := func() []byte { ranProc1 = true; return nil }
				allProcsBody := func(_ []byte) { ranAllProcs = true }
				node, errors := internal.NewSynchronizedBeforeSuiteNode(proc1Body, allProcsBody, cl)
				Ω(errors).Should(BeEmpty())
				Ω(node.ID).Should(BeNumerically(">", 0))
				Ω(node.NodeType).Should(Equal(types.NodeTypeSynchronizedBeforeSuite))

				node.SynchronizedBeforeSuiteProc1Body()
				Ω(ranProc1).Should(BeTrue())

				node.SynchronizedBeforeSuiteAllProcsBody(nil)
				Ω(ranAllProcs).Should(BeTrue())

				Ω(node.CodeLocation).Should(Equal(cl))
				Ω(node.NestingLevel).Should(Equal(0))
			})
		})

		Describe("NewSynchronizedAfterSuiteNode", func() {
			It("returns a correctly configured node", func() {
				var ranProc1, ranAllProcs bool
				allProcsBody := func() { ranAllProcs = true }
				proc1Body := func() { ranProc1 = true }

				node, errors := internal.NewSynchronizedAfterSuiteNode(allProcsBody, proc1Body, cl)
				Ω(errors).Should(BeEmpty())
				Ω(node.ID).Should(BeNumerically(">", 0))
				Ω(node.NodeType).Should(Equal(types.NodeTypeSynchronizedAfterSuite))

				node.SynchronizedAfterSuiteAllProcsBody()
				Ω(ranAllProcs).Should(BeTrue())

				node.SynchronizedAfterSuiteProc1Body()
				Ω(ranProc1).Should(BeTrue())

				Ω(node.CodeLocation).Should(Equal(cl))
				Ω(node.NestingLevel).Should(Equal(0))
			})
		})

		Describe("NewReportBeforeEachNode", func() {
			It("returns a correctly configured node", func() {
				var didRun bool
				body := func(types.SpecReport) { didRun = true }

				node, errors := internal.NewReportBeforeEachNode(body, cl)
				Ω(errors).Should(BeEmpty())
				Ω(node.ID).Should(BeNumerically(">", 0))
				Ω(node.NodeType).Should(Equal(types.NodeTypeReportBeforeEach))

				node.ReportEachBody(types.SpecReport{})
				Ω(didRun).Should(BeTrue())

				Ω(node.CodeLocation).Should(Equal(cl))
				Ω(node.NestingLevel).Should(Equal(-1))
			})
		})

		Describe("NewReportAfterEachNode", func() {
			It("returns a correctly configured node", func() {
				var didRun bool
				body := func(types.SpecReport) { didRun = true }

				node, errors := internal.NewReportAfterEachNode(body, cl)
				Ω(errors).Should(BeEmpty())
				Ω(node.ID).Should(BeNumerically(">", 0))
				Ω(node.NodeType).Should(Equal(types.NodeTypeReportAfterEach))

				node.ReportEachBody(types.SpecReport{})
				Ω(didRun).Should(BeTrue())

				Ω(node.CodeLocation).Should(Equal(cl))
				Ω(node.NestingLevel).Should(Equal(-1))
			})
		})

		Describe("NewReportAfterSuiteNode", func() {
			It("returns a correctly configured node", func() {
				var didRun bool
				body := func(types.Report) { didRun = true }
				node, errors := internal.NewReportAfterSuiteNode("my custom report", body, cl)
				Ω(errors).Should(BeEmpty())
				Ω(node.Text).Should(Equal("my custom report"))
				Ω(node.ID).Should(BeNumerically(">", 0))
				Ω(node.NodeType).Should(Equal(types.NodeTypeReportAfterSuite))

				node.ReportAfterSuiteBody(types.Report{})
				Ω(didRun).Should(BeTrue())

				Ω(node.CodeLocation).Should(Equal(cl))
				Ω(node.NestingLevel).Should(Equal(0))
			})
		})

		Describe("NewCleanupNode", func() {
			var capturedFailure string
			var capturedCL types.CodeLocation

			var failFunc = func(msg string, cl types.CodeLocation) {
				capturedFailure = msg
				capturedCL = cl
			}

			BeforeEach(func() {
				capturedFailure = ""
				capturedCL = types.CodeLocation{}
			})

			Context("when passed no function", func() {
				It("errors", func() {
					node, errs := internal.NewCleanupNode(failFunc, cl)
					Ω(node.IsZero()).Should(BeTrue())
					Ω(errs).Should(ConsistOf(types.GinkgoErrors.DeferCleanupInvalidFunction(cl)))
					Ω(capturedFailure).Should(BeZero())
					Ω(capturedCL).Should(BeZero())
				})
			})

			Context("when passed a function that returns too many values", func() {
				It("errors", func() {
					node, errs := internal.NewCleanupNode(failFunc, cl, func() (int, error) {
						return 0, nil
					})
					Ω(node.IsZero()).Should(BeTrue())
					Ω(errs).Should(ConsistOf(types.GinkgoErrors.DeferCleanupInvalidFunction(cl)))
					Ω(capturedFailure).Should(BeZero())
					Ω(capturedCL).Should(BeZero())
				})
			})

			Context("when passed a function that does not return", func() {
				It("creates a body that runs the function and never calls the fail handler", func() {
					didRun := false
					node, errs := internal.NewCleanupNode(failFunc, cl, func() {
						didRun = true
					})
					Ω(node.CodeLocation).Should(Equal(cl))
					Ω(node.NodeType).Should(Equal(types.NodeTypeCleanupInvalid))
					Ω(errs).Should(BeEmpty())

					node.Body()
					Ω(didRun).Should(BeTrue())
					Ω(capturedFailure).Should(BeZero())
					Ω(capturedCL).Should(BeZero())
				})
			})

			Context("when passed a function that returns nil", func() {
				It("creates a body that runs the function and does not call the fail handler", func() {
					didRun := false
					node, errs := internal.NewCleanupNode(failFunc, cl, func() error {
						didRun = true
						return nil
					})
					Ω(node.CodeLocation).Should(Equal(cl))
					Ω(node.NodeType).Should(Equal(types.NodeTypeCleanupInvalid))
					Ω(errs).Should(BeEmpty())

					node.Body()
					Ω(didRun).Should(BeTrue())
					Ω(capturedFailure).Should(BeZero())
					Ω(capturedCL).Should(BeZero())
				})
			})

			Context("when passed a function that returns an error", func() {
				It("creates a body that runs the function and does not call the fail handler", func() {
					didRun := false
					node, errs := internal.NewCleanupNode(failFunc, cl, func() error {
						didRun = true
						return fmt.Errorf("welp")
					})
					Ω(node.CodeLocation).Should(Equal(cl))
					Ω(node.NodeType).Should(Equal(types.NodeTypeCleanupInvalid))
					Ω(errs).Should(BeEmpty())

					node.Body()
					Ω(didRun).Should(BeTrue())
					Ω(capturedFailure).Should(Equal("DeferCleanup callback returned error: welp"))
					Ω(capturedCL).Should(Equal(cl))
				})
			})

			Context("when passed a function that takes arguments, and those arguments", func() {
				It("creates a body that runs the function and passes in those arguments", func() {
					var inA, inB, inC = "A", 2, "C"
					var receivedA, receivedC string
					var receivedB int
					node, errs := internal.NewCleanupNode(failFunc, cl, func(a string, b int, c string) error {
						receivedA, receivedB, receivedC = a, b, c
						return nil
					}, inA, inB, inC)
					inA, inB, inC = "floop", 3, "flarp"
					Ω(node.CodeLocation).Should(Equal(cl))
					Ω(node.NodeType).Should(Equal(types.NodeTypeCleanupInvalid))
					Ω(errs).Should(BeEmpty())

					node.Body()
					Ω(receivedA).Should(Equal("A"))
					Ω(receivedB).Should(Equal(2))
					Ω(receivedC).Should(Equal("C"))
					Ω(capturedFailure).Should(BeZero())
					Ω(capturedCL).Should(BeZero())
				})
			})

			Context("controlling the cleanup's code location", func() {
				It("computes its own when one is not provided", func() {
					node, errs := func() (internal.Node, []error) {
						return internal.NewCleanupNode(failFunc, func() error {
							return fmt.Errorf("welp")
						})
					}()
					localCL := types.NewCodeLocation(0)
					localCL.LineNumber -= 1
					Ω(node.CodeLocation).Should(Equal(localCL))
					Ω(node.NodeType).Should(Equal(types.NodeTypeCleanupInvalid))
					Ω(errs).Should(BeEmpty())

					node.Body()
					Ω(capturedFailure).Should(Equal("DeferCleanup callback returned error: welp"))
					Ω(capturedCL).Should(Equal(localCL))
				})

				It("can accept an Offset", func() {
					node, errs := func() (internal.Node, []error) {
						return func() (internal.Node, []error) {
							return internal.NewCleanupNode(failFunc, Offset(1), func() error {
								return fmt.Errorf("welp")
							})
						}()
					}()
					localCL := types.NewCodeLocation(0)
					localCL.LineNumber -= 1
					Ω(node.CodeLocation).Should(Equal(localCL))
					Ω(node.NodeType).Should(Equal(types.NodeTypeCleanupInvalid))
					Ω(errs).Should(BeEmpty())

					node.Body()
					Ω(capturedFailure).Should(Equal("DeferCleanup callback returned error: welp"))
					Ω(capturedCL).Should(Equal(localCL))

				})

				It("can accept a code location", func() {
					node, errs := internal.NewCleanupNode(failFunc, cl, func() error {
						return fmt.Errorf("welp")
					})
					Ω(node.CodeLocation).Should(Equal(cl))
					Ω(node.NodeType).Should(Equal(types.NodeTypeCleanupInvalid))
					Ω(errs).Should(BeEmpty())

					node.Body()
					Ω(capturedFailure).Should(Equal("DeferCleanup callback returned error: welp"))
					Ω(capturedCL).Should(Equal(cl))
				})
			})
		})
	})

	Describe("IsZero()", func() {
		It("returns true if the node is zero", func() {
			Ω(Node{}.IsZero()).Should(BeTrue())
		})

		It("returns false if the node is non-zero", func() {
			node, errors := internal.NewNode(nil, ntIt, "hummus", func() {}, cl)
			Ω(errors).Should(BeEmpty())
			Ω(node.IsZero()).Should(BeFalse())
		})
	})
})

var _ = Describe("Nodes", func() {
	Describe("CopyAppend", func() {
		var n1, n2, n3, n4 Node

		BeforeEach(func() {
			n1, n2, n3, n4 = N(), N(), N(), N()
		})

		It("appends the passed in nodes and returns the result", func() {
			result := Nodes{n1, n2}.CopyAppend(n3, n4)
			Ω(result).Should(Equal(Nodes{n1, n2, n3, n4}))
		})

		It("makes a copy, leaving the original untouched", func() {
			original := Nodes{n1, n2}
			original.CopyAppend(n3, n4)
			Ω(original).Should(Equal(Nodes{n1, n2}))
		})
	})

	Describe("SplitAround", func() {
		var nodes Nodes

		BeforeEach(func() {
			nodes = Nodes{N(), N(), N(), N(), N()}
		})

		Context("when the pivot is a member of nodes", func() {
			Context("when the pivot is not at one of the ends", func() {
				It("returns the correct left and right nodes", func() {
					left, right := nodes.SplitAround(nodes[2])
					Ω(left).Should(Equal(Nodes{nodes[0], nodes[1]}))
					Ω(right).Should(Equal(Nodes{nodes[3], nodes[4]}))
				})
			})

			Context("when the pivot is the first member", func() {
				It("returns an empty left nodes and the complete right nodes", func() {
					left, right := nodes.SplitAround(nodes[0])
					Ω(left).Should(BeEmpty())
					Ω(right).Should(Equal(Nodes{nodes[1], nodes[2], nodes[3], nodes[4]}))

				})
			})

			Context("when the pivot is the last member", func() {
				It("returns an empty right nodes and the complete left nodes", func() {
					left, right := nodes.SplitAround(nodes[4])
					Ω(left).Should(Equal(Nodes{nodes[0], nodes[1], nodes[2], nodes[3]}))
					Ω(right).Should(BeEmpty())
				})
			})
		})

		Context("when the pivot is not in nodes", func() {
			It("returns an empty right nodes and the complete left nodes", func() {
				left, right := nodes.SplitAround(N())
				Ω(left).Should(Equal(nodes))
				Ω(right).Should(BeEmpty())
			})
		})
	})

	Describe("FirstNodeWithType", func() {
		var nodes Nodes

		BeforeEach(func() {
			nodes = Nodes{N(ntCon), N("bef1", ntBef), N("bef2", ntBef), N(ntIt), N(ntAf)}
		})

		Context("when there is a matching node", func() {
			It("returns the first node that matches one of the requested node types", func() {
				Ω(nodes.FirstNodeWithType(ntAf | ntIt | ntBef).Text).Should(Equal("bef1"))
			})
		})
		Context("when there is no matching node", func() {
			It("returns an empty node", func() {
				Ω(nodes.FirstNodeWithType(ntJusAf)).Should(BeZero())
			})
		})
	})

	Describe("Filtering By NodeType", func() {
		var nCon, nBef1, nBef2, nIt, nAf Node
		var nodes Nodes

		BeforeEach(func() {
			nCon = N(ntCon)
			nBef1 = N(ntBef)
			nBef2 = N(ntBef)
			nIt = N(ntIt)
			nAf = N(ntAf)
			nodes = Nodes{nCon, nBef1, nBef2, nIt, nAf}
		})

		Describe("WithType", func() {
			Context("when there are matching nodes", func() {
				It("returns them while preserving order", func() {
					Ω(nodes.WithType(ntIt | ntBef)).Should(Equal(Nodes{nBef1, nBef2, nIt}))
				})
			})

			Context("when there are no matching nodes", func() {
				It("returns an empty Nodes{}", func() {
					Ω(nodes.WithType(ntJusAf)).Should(BeEmpty())
				})
			})
		})

		Describe("WithoutType", func() {
			Context("when there are matching nodes", func() {
				It("does not include them in the result", func() {
					Ω(nodes.WithoutType(ntIt | ntBef)).Should(Equal(Nodes{nCon, nAf}))
				})
			})

			Context("when no nodes match", func() {
				It("doesn't elide any nodes", func() {
					Ω(nodes.WithoutType(ntJusAf)).Should(Equal(nodes))
				})
			})
		})

		Describe("WithoutNode", func() {
			Context("when operating on an empty nodes list", func() {
				It("does nothing", func() {
					nodes = Nodes{}
					Ω(nodes.WithoutNode(N(ntIt))).Should(BeEmpty())

				})
			})
			Context("when the node is in the nodes list", func() {
				It("returns a copy of the nodes list without the node in it", func() {
					Ω(nodes.WithoutNode(nBef2)).Should(Equal(Nodes{nCon, nBef1, nIt, nAf}))
					Ω(nodes).Should(Equal(Nodes{nCon, nBef1, nBef2, nIt, nAf}))
				})
			})

			Context("when the node is not in the nodes list", func() {
				It("returns an unadulterated copy of the nodes list", func() {
					Ω(nodes.WithoutNode(N(ntBef))).Should(Equal(Nodes{nCon, nBef1, nBef2, nIt, nAf}))
					Ω(nodes).Should(Equal(Nodes{nCon, nBef1, nBef2, nIt, nAf}))
				})
			})
		})

		Describe("Filter", func() {
			It("returns a copy of the nodes list containing nodes that pass the filter", func() {
				filtered := nodes.Filter(func(n Node) bool {
					return n.NodeType.Is(types.NodeTypeBeforeEach | types.NodeTypeIt)
				})
				Ω(filtered).Should(Equal(Nodes{nBef1, nBef2, nIt}))
				Ω(nodes).Should(Equal(Nodes{nCon, nBef1, nBef2, nIt, nAf}))

				filtered = nodes.Filter(func(n Node) bool {
					return false
				})
				Ω(filtered).Should(BeEmpty())
			})
		})
	})

	Describe("SortedByDescendingNestingLevel", func() {
		var n0A, n0B, n1A, n1B, n1C, n2A, n2B Node
		var nodes Nodes
		BeforeEach(func() {
			n0A = N(NestingLevel(0))
			n0B = N(NestingLevel(0))
			n1A = N(NestingLevel(1))
			n1B = N(NestingLevel(1))
			n1C = N(NestingLevel(1))
			n2A = N(NestingLevel(2))
			n2B = N(NestingLevel(2))
			nodes = Nodes{n0A, n0B, n1A, n1B, n1C, n2A, n2B}
		})

		It("returns copy sorted by descending nesting level, preserving order within nesting level", func() {
			Ω(nodes.SortedByDescendingNestingLevel()).Should(Equal(Nodes{n2A, n2B, n1A, n1B, n1C, n0A, n0B}))
			Ω(nodes).Should(Equal(Nodes{n0A, n0B, n1A, n1B, n1C, n2A, n2B}), "original nodes should not have been modified")
		})
	})

	Describe("SortedByAscendingNestingLevel", func() {
		var n0A, n0B, n1A, n1B, n1C, n2A, n2B Node
		var nodes Nodes
		BeforeEach(func() {
			n0A = N(NestingLevel(0))
			n0B = N(NestingLevel(0))
			n1A = N(NestingLevel(1))
			n1B = N(NestingLevel(1))
			n1C = N(NestingLevel(1))
			n2A = N(NestingLevel(2))
			n2B = N(NestingLevel(2))
			nodes = Nodes{n2A, n1A, n1B, n0A, n2B, n0B, n1C}
		})

		It("returns copy sorted by ascending nesting level, preserving order within nesting level", func() {
			Ω(nodes.SortedByAscendingNestingLevel()).Should(Equal(Nodes{n0A, n0B, n1A, n1B, n1C, n2A, n2B}))
			Ω(nodes).Should(Equal(Nodes{n2A, n1A, n1B, n0A, n2B, n0B, n1C}), "original nodes should not have been modified")
		})
	})

	Describe("WithinNestingLevel", func() {
		var n0, n1, n2a, n2b, n3 Node
		var nodes Nodes
		BeforeEach(func() {
			n0 = N(NestingLevel(0))
			n1 = N(NestingLevel(1))
			n2a = N(NestingLevel(2))
			n3 = N(NestingLevel(3))
			n2b = N(NestingLevel(2))
			nodes = Nodes{n0, n1, n2a, n3, n2b}
		})

		It("returns nodes, in order, with nesting level equal to or less than the requested level", func() {
			Ω(nodes.WithinNestingLevel(-1)).Should(BeEmpty())
			Ω(nodes.WithinNestingLevel(0)).Should(Equal(Nodes{n0}))
			Ω(nodes.WithinNestingLevel(1)).Should(Equal(Nodes{n0, n1}))
			Ω(nodes.WithinNestingLevel(2)).Should(Equal(Nodes{n0, n1, n2a, n2b}))
			Ω(nodes.WithinNestingLevel(3)).Should(Equal(Nodes{n0, n1, n2a, n3, n2b}))
		})
	})

	Describe("Reverse", func() {
		It("reverses the nodes", func() {
			nodes := Nodes{N("A"), N("B"), N("C"), N("D"), N("E")}
			Ω(nodes.Reverse().Texts()).Should(Equal([]string{"E", "D", "C", "B", "A"}))
		})

		It("works with empty nodes", func() {
			nodes := Nodes{}
			Ω(nodes.Reverse()).Should(Equal(Nodes{}))
		})
	})

	Describe("Texts", func() {
		var nodes Nodes
		BeforeEach(func() {
			nodes = Nodes{N("the first node"), N(""), N("2"), N("c"), N("")}
		})

		It("returns a string slice containing the individual node text strings in order", func() {
			Ω(nodes.Texts()).Should(Equal([]string{"the first node", "", "2", "c", ""}))
		})
	})

	Describe("Labels and UnionOfLabels", func() {
		var nodes Nodes
		BeforeEach(func() {
			nodes = Nodes{N(Label("A", "B")), N(Label("C")), N(), N(Label("A")), N(Label("D")), N(Label("B", "D", "E"))}
		})

		It("Labels returns a slice containing the labels for each node in order", func() {
			Ω(nodes.Labels()).Should(Equal([][]string{
				{"A", "B"},
				{"C"},
				{},
				{"A"},
				{"D"},
				{"B", "D", "E"},
			}))
		})

		It("UnionOfLabels returns a single slice of labels harvested from all nodes and deduped", func() {
			Ω(nodes.UnionOfLabels()).Should(Equal([]string{"A", "B", "C", "D", "E"}))
		})
	})

	Describe("CodeLocation", func() {
		var nodes Nodes
		var cl1, cl2 types.CodeLocation
		BeforeEach(func() {
			cl1 = types.NewCodeLocation(0)
			cl2 = types.NewCodeLocation(0)
			nodes = Nodes{N(cl1), N(cl2), N()}
		})

		It("returns a types.CodeLocation sice containing the individual node code locations in order", func() {
			Ω(nodes.CodeLocations()).Should(Equal([]types.CodeLocation{cl1, cl2, cl}))
		})
	})

	Describe("BestTextFor", func() {
		var nIt, nBef1, nBef2 Node
		var nodes Nodes
		BeforeEach(func() {
			nIt = N("an it", ntIt, NestingLevel(2))
			nBef1 = N(ntBef, NestingLevel(2))
			nBef2 = N(ntBef, NestingLevel(4))
			nodes = Nodes{
				N("the root container", ntCon, NestingLevel(0)),
				N("the inner container", ntCon, NestingLevel(1)),
				nBef1,
				nIt,
				nBef2,
			}
		})

		Context("when the passed in node has text", func() {
			It("returns that text", func() {
				Ω(nodes.BestTextFor(nIt)).Should(Equal("an it"))
			})
		})

		Context("when the node has no text", func() {
			Context("and there is a node one-nesting-level-up with text", func() {
				It("returns that node's text", func() {
					Ω(nodes.BestTextFor(nBef1)).Should(Equal("the inner container"))
				})
			})

			Context("and there is no node one-nesting-level up with text", func() {
				It("returns empty string", func() {
					Ω(nodes.BestTextFor(nBef2)).Should(Equal(""))
				})
			})
		})
	})

	Describe("ContainsNodeID", func() {
		Context("when there is a node with the matching ID", func() {
			It("returns true", func() {
				nodes := Nodes{N(), N(), N()}
				Ω(nodes.ContainsNodeID(nodes[1].ID)).Should(BeTrue())
			})
		})

		Context("when there is no node with matching ID", func() {
			It("returns false", func() {
				nodes := Nodes{N(), N(), N()}
				Ω(nodes.ContainsNodeID(nodes[2].ID + 1)).Should(BeFalse())
			})
		})
	})

	Describe("HasNodeMarkedPending", func() {
		Context("when there is a node marked pending", func() {
			It("returns true", func() {
				nodes := Nodes{N(), N(), N(Pending), N()}
				Ω(nodes.HasNodeMarkedPending()).Should(BeTrue())
			})
		})

		Context("when there is no node marked pending", func() {
			It("returns false", func() {
				nodes := Nodes{N(), N(), N()}
				Ω(nodes.HasNodeMarkedPending()).Should(BeFalse())
			})
		})
	})

	Describe("HasNodeMarkedFocus", func() {
		Context("when there is a node marked focus", func() {
			It("returns true", func() {
				nodes := Nodes{N(), N(), N(Focus), N()}
				Ω(nodes.HasNodeMarkedFocus()).Should(BeTrue())
			})
		})

		Context("when there is no node marked focus", func() {
			It("returns false", func() {
				nodes := Nodes{N(), N(), N()}
				Ω(nodes.HasNodeMarkedFocus()).Should(BeFalse())
			})
		})
	})

	Describe("HasNodeMarkedSerial", func() {
		Context("when there is a node marked serial", func() {
			It("returns true", func() {
				nodes := Nodes{N(), N(), N(Serial), N()}
				Ω(nodes.HasNodeMarkedSerial()).Should(BeTrue())
			})
		})

		Context("when there is no node marked serial", func() {
			It("returns false", func() {
				nodes := Nodes{N(), N(), N()}
				Ω(nodes.HasNodeMarkedSerial()).Should(BeFalse())
			})
		})
	})

	Describe("FirstNodeMarkedOrdered", func() {
		Context("when there are nodes marked ordered", func() {
			It("returns the first one", func() {
				nodes := Nodes{N(), N("A", ntCon, Ordered), N("B", ntCon, Ordered), N()}
				Ω(nodes.FirstNodeMarkedOrdered().Text).Should(Equal("A"))
			})
		})

		Context("when there is no node marked ordered", func() {
			It("returns zero", func() {
				nodes := Nodes{N(), N(), N()}
				Ω(nodes.FirstNodeMarkedOrdered()).Should(BeZero())
			})
		})
	})
})

var _ = Describe("Iteration Performance", Serial, Label("performance"), func() {
	BeforeEach(func() {
		if os.Getenv("PERF") == "" {
			Skip("")
		}
	})

	It("compares the performance of iteration using range vs counters", func() {
		experiment := gmeasure.NewExperiment("iteration")

		size := 1000
		nodes := make(Nodes, size)
		for i := 0; i < size; i++ {
			nodes[i] = N(ntAf)
		}
		nodes[size-1] = N(ntIt)

		experiment.SampleDuration("range", func(idx int) {
			numIts := 0
			for _, node := range nodes {
				if node.NodeType.Is(ntIt) {
					numIts += 1
				}
			}
		}, gmeasure.SamplingConfig{N: 1024}, gmeasure.Precision(time.Nanosecond))

		experiment.SampleDuration("range-index", func(idx int) {
			numIts := 0
			for i := range nodes {
				if nodes[i].NodeType.Is(ntIt) {
					numIts += 1
				}
			}
		}, gmeasure.SamplingConfig{N: 1024}, gmeasure.Precision(time.Nanosecond))

		experiment.SampleDuration("counter", func(idx int) {
			numIts := 0
			for i := 0; i < len(nodes); i++ {
				if nodes[i].NodeType.Is(ntIt) {
					numIts += 1
				}
			}
		}, gmeasure.SamplingConfig{N: 1024}, gmeasure.Precision(time.Nanosecond))

		AddReportEntry(experiment.Name, gmeasure.RankStats(gmeasure.LowerMedianIsBetter, experiment.GetStats("range"), experiment.GetStats("range-index"), experiment.GetStats("counter")))

	})

	It("compares the performance of slice construction by growing slices vs pre-allocating slices vs counting twice", func() {
		experiment := gmeasure.NewExperiment("filtering")

		size := 1000
		nodes := make(Nodes, size)
		for i := 0; i < size; i++ {
			if i%100 == 0 {
				nodes[i] = N(ntIt)
			} else {
				nodes[i] = N(ntAf)
			}
		}

		largeStats := []gmeasure.Stats{}
		smallStats := []gmeasure.Stats{}

		experiment.SampleDuration("grow-slice (large)", func(idx int) {
			out := Nodes{}
			for i := range nodes {
				if nodes[i].NodeType.Is(ntAf) {
					out = append(out, nodes[i])
				}
			}
		}, gmeasure.SamplingConfig{N: 1024}, gmeasure.Precision(time.Nanosecond))
		largeStats = append(largeStats, experiment.GetStats("grow-slice (large)"))

		experiment.SampleDuration("grow-slice (small)", func(idx int) {
			out := Nodes{}
			for i := range nodes {
				if nodes[i].NodeType.Is(ntIt) {
					out = append(out, nodes[i])
				}
			}
		}, gmeasure.SamplingConfig{N: 1024}, gmeasure.Precision(time.Nanosecond))
		smallStats = append(smallStats, experiment.GetStats("grow-slice (small)"))

		experiment.SampleDuration("pre-allocate (large)", func(idx int) {
			out := make(Nodes, 0, len(nodes))
			for i := range nodes {
				if nodes[i].NodeType.Is(ntAf) {
					out = append(out, nodes[i])
				}
			}
		}, gmeasure.SamplingConfig{N: 1024}, gmeasure.Precision(time.Nanosecond))
		largeStats = append(largeStats, experiment.GetStats("pre-allocate (large)"))

		experiment.SampleDuration("pre-allocate (small)", func(idx int) {
			out := make(Nodes, 0, len(nodes))
			for i := range nodes {
				if nodes[i].NodeType.Is(ntIt) {
					out = append(out, nodes[i])
				}
			}
		}, gmeasure.SamplingConfig{N: 1024}, gmeasure.Precision(time.Nanosecond))
		smallStats = append(smallStats, experiment.GetStats("pre-allocate (small)"))

		experiment.SampleDuration("pre-count (large)", func(idx int) {
			count := 0
			for i := range nodes {
				if nodes[i].NodeType.Is(ntAf) {
					count++
				}
			}

			out := make(Nodes, count)
			j := 0
			for i := range nodes {
				if nodes[i].NodeType.Is(ntAf) {
					out[j] = nodes[i]
					j++
				}
			}
		}, gmeasure.SamplingConfig{N: 1024}, gmeasure.Precision(time.Nanosecond))
		largeStats = append(largeStats, experiment.GetStats("pre-count (large)"))

		experiment.SampleDuration("pre-count (small)", func(idx int) {
			count := 0
			for i := range nodes {
				if nodes[i].NodeType.Is(ntIt) {
					count++
				}
			}

			out := make(Nodes, count)
			j := 0
			for i := range nodes {
				if nodes[i].NodeType.Is(ntIt) {
					out[j] = nodes[i]
					j++
				}
			}
		}, gmeasure.SamplingConfig{N: 1024}, gmeasure.Precision(time.Nanosecond))
		smallStats = append(smallStats, experiment.GetStats("pre-count (small)"))

		AddReportEntry("Large Slice", gmeasure.RankStats(gmeasure.LowerMedianIsBetter, largeStats...))
		AddReportEntry("Small Slice", gmeasure.RankStats(gmeasure.LowerMedianIsBetter, smallStats...))
	})
})

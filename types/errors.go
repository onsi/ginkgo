package types

import (
	"reflect"

	"github.com/onsi/ginkgo/formatter"
)

type GinkgoError struct {
	Heading      string
	Message      string
	DocLink      string
	CodeLocation CodeLocation
}

func (g GinkgoError) Error() string {
	out := formatter.F("{{bold}}{{red}}%s{{/}}\n", g.Heading)
	if (g.CodeLocation != CodeLocation{}) {
		out += formatter.F("{{gray}}%s{{/}}\n", g.CodeLocation)
	}
	if g.Message != "" {
		out += formatter.Fiw(1, formatter.COLS, g.Message)
		out += "\n\n"
	}
	if g.DocLink != "" {
		out += formatter.Fiw(1, formatter.COLS, "{{bold}}Learn more at:{{/}} {{cyan}}{{underline}}http://onsi.github.io/ginkgo/#%s{{/}}\n", g.DocLink)
	}

	return out
}

type ginkgoErrors struct{}

var GinkgoErrors = ginkgoErrors{}

func (g ginkgoErrors) UncaughtGinkgoPanic(cl CodeLocation) error {
	return GinkgoError{
		Heading: "Your Test Panicked",
		Message: `When you, or your assertion library, calls Ginkgo's Fail(),
Ginkgo panics to prevent subsequent assertions from running.

Normally Ginkgo rescues this panic so you shouldn't see it.

However, if you make an assertion in a goroutine, Ginkgo can't capture the panic.
To circumvent this, you should call

	defer GinkgoRecover()

at the top of the goroutine that caused this panic.

Alternatively, you may have made an assertion outside of the Ginkgo
testing tree (e.g. in a func init()) - please move your assertion to
an appropriate Ginkgo node (e.g. a BeforeSuite or as part of a specific test).`,
		DocLink:      "marking-specs-as-failed",
		CodeLocation: cl,
	}
}

func (g ginkgoErrors) PushingNodeInRunPhase(nodeType NodeType, cl CodeLocation) error {
	return GinkgoError{
		Heading: "Ginkgo detected an issue with your test structure",
		Message: formatter.F(
			`It looks like you are trying to add a {{bold}}[%s]{{/}} node
to the Ginkgo test tree in a leaf node {{bold}}after{{/}} the tests started running.

To enable randomization and parallelization Ginkgo requires the test tree
to be fully construted up front.  In practice, this means that you can
only create nodes like {{bold}}[%s]{{/}} at the top-level or within the
body of a {{bold}}Describe{{/}}, {{bold}}Context{{/}}, or {{bold}}When{{/}}.`, nodeType, nodeType),
		CodeLocation: cl,
		DocLink:      "understanding-ginkgos-lifecycle",
	}
}

func (g ginkgoErrors) CaughtPanicDuringABuildPhase(cl CodeLocation) error {
	return GinkgoError{
		Heading: "Assertion or Panic detected during tree construction",
		Message: formatter.F(
			`Ginkgo detected a panic while constructing the test tree.
You may be trying to make an assertion in the body of a container node
(i.e. {{bold}}Describe{{/}}, {{bold}}Context{{/}}, or {{bold}}When{{/}}).

Please ensure all assertions are inside leaf nodes such as {{bold}}BeforeEach{{/}},
{{bold}}It{{/}}, etc.`),
		CodeLocation: cl,
		DocLink:      "do-not-make-assertions-in-container-node-functions",
	}
}

func (g ginkgoErrors) SetupNodeInNestedContext(nodeType NodeType, cl CodeLocation) error {
	return GinkgoError{
		Heading: "Ginkgo detected an issue with your test structure",
		Message: formatter.F(
			`It looks like you are trying to add a {{bold}}[%s]{{/}} node within a container node.

{{bold}}%s{{/}} can only be called at the top level.`, nodeType, nodeType),
		CodeLocation: cl,
		DocLink:      "global-setup-and-teardown-beforesuite-and-aftersuite",
	}
}

func (g ginkgoErrors) SetupNodeDuringRunPhase(nodeType NodeType, cl CodeLocation) error {
	return GinkgoError{
		Heading: "Ginkgo detected an issue with your test structure",
		Message: formatter.F(
			`It looks like you are trying to add a {{bold}}[%s]{{/}} node within a leaf node after the test started running.

{{bold}}%s{{/}} can only be called at the top level.`, nodeType, nodeType),
		CodeLocation: cl,
		DocLink:      "global-setup-and-teardown-beforesuite-and-aftersuite",
	}
}

func (g ginkgoErrors) MultipleBeforeSuiteNodes(nodeType NodeType, cl CodeLocation, earlierNodeType NodeType, earlierCodeLocation CodeLocation) error {
	return ginkgoErrorMultipleSuiteNodes("setup", nodeType, cl, earlierNodeType, earlierCodeLocation)
}

func (g ginkgoErrors) MultipleAfterSuiteNodes(nodeType NodeType, cl CodeLocation, earlierNodeType NodeType, earlierCodeLocation CodeLocation) error {
	return ginkgoErrorMultipleSuiteNodes("teardown", nodeType, cl, earlierNodeType, earlierCodeLocation)
}

func ginkgoErrorMultipleSuiteNodes(setupOrTeardown string, nodeType NodeType, cl CodeLocation, earlierNodeType NodeType, earlierCodeLocation CodeLocation) error {
	return GinkgoError{
		Heading: "Ginkgo detected an issue with your test structure",
		Message: formatter.F(
			`It looks like you are trying to add a {{bold}}[%s]{{/}} node but
you already have a {{bold}}[%s]{{/}} node defined at: {{gray}}%s{{/}}.

Ginkgo only allows you to define one suite %s node.`, nodeType, earlierNodeType, earlierCodeLocation, setupOrTeardown),
		CodeLocation: cl,
		DocLink:      "global-setup-and-teardown-beforesuite-and-aftersuite",
	}
}

func (g ginkgoErrors) InvalidBodyType(t reflect.Type, cl CodeLocation) error {
	return GinkgoError{
		Heading: "Invalid Parameter",
		Message: formatter.F(`Ginkgo nodes must be passed {{bold}}func(){{/}} - i.e. functions that take nothing and return nothing.
You passed {{bold}}%s{{/}} instead.`, t),
		CodeLocation: cl,
	}
}

func (g ginkgoErrors) RerunningSuite() error {
	return GinkgoError{
		Heading: "Rerunning Suite",
		Message: formatter.F(`It looks like you are calling RunSpecs more than once. Ginkgo does not support rerunning suites.  If you want to rerun a suite try {{bold}}ginkgo --repeat=N{{/}} or {{bold}}ginkgo --until-it-fails{{/}}`),
		DocLink: "repeating-test-runs-and-managing-flakey-tests",
	}
}

var sharedParallelErrorMessage = "It looks like you are trying to run tests in parallel with go test.\nThis is unsupported and you should use the ginkgo CLI instead."

func (g ginkgoErrors) InvalidParallelTotalConfiguration() error {
	return GinkgoError{
		Heading: "-ginkgo.parallel.total must be >= 1",
		Message: sharedParallelErrorMessage,
		DocLink: "parallel-specs",
	}
}

func (g ginkgoErrors) InvalidParallelNodeConfiguration() error {
	return GinkgoError{
		Heading: "-ginkgo.parallel.node is one-indexed and must be <= ginkgo.parallel.total",
		Message: sharedParallelErrorMessage,
		DocLink: "parallel-specs",
	}
}

func (g ginkgoErrors) MissingParallelHostConfiguration() error {
	return GinkgoError{
		Heading: "-ginkgo.parallel.host is missing",
		Message: sharedParallelErrorMessage,
		DocLink: "parallel-specs",
	}
}

func (g ginkgoErrors) UnreachableParallelHost(host string) error {
	return GinkgoError{
		Heading: "Could not reach ginkgo.parallel.host:" + host,
		Message: sharedParallelErrorMessage,
		DocLink: "parallel-specs",
	}
}

func (g ginkgoErrors) DryRunInParallelConfiguration() error {
	return GinkgoError{
		Heading: "Ginkgo only performs -dryRun in serial mode.",
		Message: "Please try running ginkgo -dryRun again, but without -p or -nodes to ensure the test is running in series.",
	}
}

func (g ginkgoErrors) ConflictingVerboseSuccinctConfiguration() error {
	return GinkgoError{
		Heading: "Conflicting reporter verbosity settings -v and --succinct.",
		Message: "You can't set both -v and --succinct.  Please pick one!",
	}
}

func (g ginkgoErrors) InvalidGoFlagCount() error {
	return GinkgoError{
		Heading: "Use of go test -count",
		Message: "Ginkgo does not support using go test -count to rerun test suites.  Please use the ginkgo cli and `ginkgo -until-it-fails` or `ginkgo -repeat`.",
	}
}

func (g ginkgoErrors) InvalidGoFlagParallel() error {
	return GinkgoError{
		Heading: "Use of go test -parallel",
		Message: "Go test's implementation of parallelization does not actually parallelize Ginkgo tests.  Please use the ginkgo cli and `ginkgo -p` or `ginkgo -nodes=N` instad.",
	}
}

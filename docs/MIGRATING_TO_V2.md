---
layout: default
title: Migrating to Ginkgo V2
---

[Ginkgo 2.0](#711) is a major release that adds substantial new functionality and removes/moves existing functionality.

This document serves as a changelog and migration guide for users migrating from Ginkgo 1.x to 2.0.  The intent is that the migration will take minimal effort.

# Additions and Improvements

## Major Additions and Improvements

### Interrupt Behavior
Interrupt behavior is substantially improved, sending an interrupt signal will now:
	- immediately cause the current test to unwind.  Ginkgo will run any `AfterEach` blocks, then immediately skip all remaining tests, then run the `AfterSuite` block.  
	- emit information about which node Ginkgo was running when the interrupt signal was recevied.
	- emit as much information as possible about the interrupted test (e.g. `GinkgoWriter` contents, `stdout` and `stderr` context).

Previously, sending a second interrupt signal would cause Ginkgo to exit immediately.  With the improved interrupt behavior this is no longer necessary.

## Minor Additions and Improvements
- `BeforeSuite` and `AfterSuite` no longer run if all tests in a suite are skipped.
- Ginkgo can now catch several common user gotchas and emit a helpful error.
- Error output is clearer and consistently links to relevant sections in the documentation.
- Test randomization is now more stable as tests are now sorted deterministcally on file_name:line_number first (previously they were sorted on test text which could not guarantee a stable sort).

# Changes

## Major Changes
These are major changes that will need user intervention to migrate succesfully.

### Removed: Async Testing
As described in the [Ginkgo 2.0 Proposal](https://docs.google.com/document/d/1h28ZknXRsTLPNNiOjdHIO-F2toCzq4xoZDXbfYaBdoQ/edit#heading=h.mzgqmkg24xoo) the Ginkgo 1.x implementation of asynchronous testing using a `Done` channel was a confusing source of test-pollution.  It is removed in Ginkgo 2.0.

In Ginkgo 2.0 tests of the form:

```
It("...", func(done Done) {
	// user test code to run asynchronously
	close(done) //signifies the test is done
}, timeout)
```

will emit a deprecation warning and will run **synchronously**.  This means the `timeout` will not be enforced and the status of the `Done` channel will be ignored - a test that hangs will hang indefinitely.

#### Migration Strategy:
We recommend users make targeted use of Gomega's [Asynchronous Assertions](https://onsi.github.io/gomega/#making-asynchronous-assertions) to better test asynchronous behavior.

As a first migration pass that produces **equivalent behavior** users can replace asynchronous tests with:

```
It("...", func() {
	done := make(chan interface{})
	go func() {
		// user test code to run asynchronously
		close(done) //signifies the code is done
	}()
	Eventually(done, timeout).Should(BeClosed())
})
```

### Removed: Measure
As described in the [Ginkgo 2.0 Proposal](https://docs.google.com/document/d/1h28ZknXRsTLPNNiOjdHIO-F2toCzq4xoZDXbfYaBdoQ/edit#heading=h.2ezhpn4gmcgs) the Ginkgo 1.x implementation of benchmarking using `Measure` nodes was a source of tightly-coupled complexity.  It is removed in Ginkgo 2.0.

In Ginkgo 2.0 tests of the form:
```
Measure(..., func(b Benchmarker) {
	// user benchmark code
})
```

will emit a deprecation warning and **will no longer run**.

#### Migration Strategy:
A new Gomega benchmarking subpackage is being developed to replace Ginkgo's benchmarking capabilities with a more mature, decoupled, and useful implementation.  This section will be updated once the Gomega package is ready.

### Changed: Reporter Interface
Objects satisfying the `Reporter` interface can be passed to Ginkgo to report information about tests.  The `Reporter` interface has changed in Ginkgo 2.0.

The Ginkgo 1.x `Reporter` interface:
```
type V1Reporter interface {
	SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary)
	BeforeSuiteDidRun(setupSummary *types.SetupSummary)
	SpecWillRun(specSummary *types.SpecSummary)
	SpecDidComplete(specSummary *types.SpecSummary)
	AfterSuiteDidRun(setupSummary *types.SetupSummary)
	SpecSuiteDidEnd(summary *types.SuiteSummary)
}
```

is now simpler in Ginkgo 2.0:
```
type Reporter interface {
	SpecSuiteWillBegin(config config.GinkgoConfigType, summary types.SuiteSummary)
	WillRun(specSummary types.Summary)
	DidRun(specSummary types.Summary)
	SpecSuiteDidEnd(summary types.SuiteSummary)
}
```

In addition, there have been changes to the data types passed into these methods.  The original types are preserved in [http://github.com/onsi/ginkgo/blob/v2/types/deprecation_support.go] and aliased to their original names.

#### Migration Strategy:
Most users will not need to worry about this change.  For users with custom reporters you will see a compilation failure when you try to pass your `V1Reporter` to Ginkgo.  This can be fixed immediately by wrapping your custom reporter with `reporters.ReporterFromV1Reporter()`:

```
import "github.com/onsi/ginkgo/reporters"

func TestXXX(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "XXX", []Reporter{
		reporters.ReporterFromV1Reporter(customV1Reporter)
	})
}
```

Your test suite will now compile and Ginkgo will send appropriately formatted data to your reporter.  However, Ginkgo will emit a deprecation warning that you are using a V1 reporter.  To silence the warning you will need to rewrite your reporter to match the new interface.  You can look at the implementation of `reporters.ReporterFromV1Reporter` and `reporters.DefaultReporter` for guidance, open an issue for help, or reach out for help on the Ginkgo slack channel.

Alternatively, you can avoid using a custom reporter and instead ask Ginkgo to emit a JSON formatted report (TODO: update with link once implemented) that you can then post-process with your tooling.

### Changed: Profiling Support
Ginkgo V1 was incorrectly handling Go test's various profiling flags (e.g. -cpuprofile, -memprofile).  This has been fixed in V2.  In fact, V2 can capture profiles for multiple packages (e.g. ginkgo -r -cpuprofile=profile.out will work).

When generating profiles for `-cpuprofile=FILE`, `-blockprofile=FILE`, `-memprofile=FILE`, `-mutexprofile=FILE`, and `-execution-trace=FILE` (Ginkgo's alias for `go test -test.trace`) the following rules apply:

- If `-output-dir` is not set: each profile generates a file named `$FILE` in the directory of each package under test.
- If `-output-dir` is set: each profile generates a file in the specified `-output-dir` directory. named `PACKAGE_NAME_$FILE`

When generating cover profiles using `-cover` and `-coverprofile=FILE`, the following rules apply:

- By default, a single cover profile is generated at $FILE (or `coverprofile.out` if `-cover-profile` is not set but `-cover` is set).  This includes the merged results of all the cover profiels reported by each suite.
- If `-output-dir` is set: the $FILE is placed in the specified `-output-dir` directory.
- If `-keep-separate-coverprofiles` is set, the individual coverprofiles generated by each package are not merged and, instead, a file named $FILE will appear in each package directory.  If `-output-dir` is set these files are copied into the `-output-dir` directory and namespaced with `PACKAGE_NAME_$FILE`.


### Changed: Command Line Flags
All camel case flags (e.g. `-randomizeAllSpecs`) are replaced with kebab case flags (e.g. `-randomize-all-specs`) in Ginkgo 2.0.  The camel case versions continue to work but emit a deprecation warning.

#### Migration Strategy:
Users should update any scripts they have that invoke the `ginkgo` cli from camel case to kebab case (:camel: :arrow_right: :oden:).

### Removed: `-stream`
`-stream` was originally introduce in Ginkgo 1.x to force parallel test processes to emit output simultaneously in order to help debug hanging test issues.  With improvements to Ginkgo's interrupt handling and parallel test reporting this behavior is no longer necessary and has been removed.

### Removed: `-notify`
`-notify` instructed Ginkgo to emit desktop notifications on linux and MacOS.  This feature was rarely used and has been removed.

### Removed: `-noisyPendings` and `-noisySkippings`
Both these flags tweaked the reporter's behavior for pending and skipped tests.  A similar, and more intuitive, outcome is possible using `--succinct` and `-v`.

#### Migration Strategy:
Users should remove -stream from any scripts they have that invoke the `ginkgo` cli.

### Removed: `ginkgo convert`
The `ginkgo convert` subcommand in V1 could convert an existing set of Go tests into a Ginkgo test suite, wrapping each `TestX` function in an `It`.  This subcommand added complexity to the codebase and was infrequently used.  It has been removed.  Users who want to convert tests suites over to Ginkgo will need to do so by hand.

## Minor Changes
These are minor changes that will be transparent for most users.

- `"top level"` is no longer the first element in `types.Summary.NodeTexts`.  This will only affect users who write custom reporters.

- The output format of Ginkgo's Default Reporter has changed in numerous subtle ways to improve readability and the user experience.  Users who were scraping Ginkgo output programatically may need to change their scripts or use the new JSON formatted report option (TODO: update with link once JSON reporting is implemented).

- Removed `ginkgo blur` alias.  Use `ginkgo unfocus` instead.



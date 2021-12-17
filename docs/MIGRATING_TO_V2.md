---
layout: default
title: Migrating to Ginkgo V2
---
{% raw  %}
[Ginkgo 2.0](https://github.com/onsi/ginkgo/issues/711) is a major release that adds substantial new functionality and removes/moves existing functionality.  **Please provide feedback, concerns, questions, etc. on issue [#711](https://github.com/onsi/ginkgo/issues/711).**

This document serves as a changelog and migration guide for users migrating from Ginkgo 1.x to 2.0.  The intent is that the migration will take minimal user effort.

The work on 2.0 is tracked in this [Pivotal Tracker backlog](https://www.pivotaltracker.com/n/projects/2493102).  The original 2.0 proposal is [here](https://docs.google.com/document/d/1h28ZknXRsTLPNNiOjdHIO-F2toCzq4xoZDXbfYaBdoQ/edit#heading=h.mzgqmkg24xoo) however this is starting to grow stale.  The tracker backlog is the most authoritative version of the roadmap.

# Using the Beta

A release candidate for Ginkgo 2.0 is now available.  The GA release should come out in the Fall of 2021 but we'd love to get feedback and usage of the RC underway.

Currently, 2.0 lives on a branch called `ver2` and major version number of Ginkgo has been bumped to v2.  If you are using `go mod` you'll need to do the following to use Ginkgo 2.0 in an existing or new project:

1. Get the `2.0.0-rc3` tag of `ginkgo/v2`:
	```bash
	go get github.com/onsi/ginkgo/v2@v2.0.0-rc3
	```

2. Install the V2 CLI.  Running this may require you to run a few addition `go get`s - just follow the go toolchain's instructions until you successfully get ginkgo v2 compiled:
	```bash
	go install github.com/onsi/ginkgo/v2/ginkgo
	ginkgo version //should print out "Ginkgo Version 2.0.0-rc3"
	```

3. Update all your import statements from `import github.com/onsi/ginkgo` to `import github.com/onsi/ginkgo/v2`.  You can use your text editor to replace all instances of `"github.com/onsi/ginkgo` with `"github.com/onsi/ginkgo/v2`

And that's it!

Please share any feedback about the RC on the [Ginkgo 2.0](https://github.com/onsi/ginkgo/issues/711) issue.  Updated V2 documentation is being maintained [here](https://github.com/onsi/ginkgo/blob/ver2/docs/index.md) though this migration guide has all the details around new features and backward incompatible changes.  Updating to V2 will require you to make some changes to your test suites however the intent is that this work should be relatively minimal for most users.

# Additions and Improvements

## Major Additions and Improvements

### Interrupt Behavior
Interrupt behavior is substantially improved, sending an interrupt signal will now:
	- immediately cause the current test to unwind.  Ginkgo will run any `AfterEach` blocks, then immediately skip all remaining tests, then run the `AfterSuite` block.  
	- emit information about which node Ginkgo was running when the interrupt signal was recevied.
	- emit as much information as possible about the interrupted test (e.g. `GinkgoWriter` contents, `stdout` and `stderr` context).
	- emit a stack trace of every running goroutine at the moment of interruption.

Previously, sending a second interrupt signal would cause Ginkgo to exit immediately.  With the improved interrupt behavior this is no longer necessary and Ginkgo will not exit until the test suite has unwound and completed.

### Timeout Behavior
In Ginkgo V1.x, Ginkgo's timeout was managed by `go test`.  This meant that timeouts exited the test suite abruptly with no opportunity for custom reporters or clean up code (e.g. `AfterEach`, `AfterSuite`) to run.  This is fixed in V2.  Ginkgo now manages its own timeout and when a timeout triggers the test winds down gracefully.  In fact, a timeout is now functionally equivalen to a user-initiated interrupt.

In addition, in V1.x when running multiple test suites Ginkgo would give each suite the full timeout allotment (so `ginkgo -r -timeout=1h` would give _each_ test suite one hour to complete).  In V2 the timeout now applies to the entire test suite run so that `ginkgo -r -timeout=1h` is now guaranteed to exit after (about) one hour.

**Finally, the default timeout has been reduced from `24h` down to `1h`.**  Users with long-running tests may need to adjust the timeout in their CI scripts.

### Spec Decorators
Specs can now be decorated with a series of new spec decorators.  These decorators enable fine-grained control over certain aspects of the spec's creation and lifecycle. 

To support decorators, the signature for Ginkgo's container, setup, and It nodes have been changed to:

```go
func Describe(text string, args ...interface{})
func It(text string, args ...interface{})
func BeforeEach(args ...interface{})
```
Note that this change is backwards compatible with v1.X.

Ginkgo supports passing in decorators _and_ arbitrarily nested slices of decorators.  Ginkgo will unroll any slices and process the flattened list of decorators.  This makes it easier to pass around and combine groups of decorators.  In addition, decorators can be passed into the table-related DSL: `DescribeTable` and `Entry`.

Here's a list of new decorators.  They are documented in more detail in the [Node Decorator Reference](https://github.com/onsi/ginkgo/blob/ver2/docs/index.md#node-decorators-overview) section of the documentation.

#### Serial Decorator
Specs can now be decorated with the `Serial` decorator.  Specs decorated as `Serial` will never run in parallel with other specs.  Instead, Ginkgo will run them on a single test process _after_ all the parallel tests have finished running.

#### Ordered Decorator
Spec containers (i.e. `Describe` and `Context` blocks) can now be decorated with the `Ordered` decorator.  Specs within `Ordered` containers will always run in the order they appear and will never be randomized.  In addition, when running in parallel, specs in an `Ordered` containers will always run on the same process to ensure spec order is preserved.  When a spec in an `Ordered` container fails, all subsequent specs in the container are skipped.

`Ordered` containers also support `BeforeAll` and `AfterAll` setup nodes.  These nodes will run just once - the `BeforeAll` will run before any ordered tests in the container run; the `AfterAll` will run after all the ordered tests in the container are finished.

Ordered containers are documented in more details in the [Ordered Container](https://github.com/onsi/ginkgo/blob/ver2/docs/index.md#ordered-containers) section of the documentation.

#### OncePerOrdered Decorator
The `OncePerOrdered` decorator can be applied to setup nodes and causes them to run just once around ordered containers.  More details in the [Setup around Ordered Containers: the OncePerOrdered Decorator](https://github.com/onsi/ginkgo/blob/ver2/docs/index.md#setup-around-ordered-containers-the-onceperordered-decorator) section of the documentation.

#### Label Decorator
Specs can now be decorated with the `Label` decorator (see [Spec Labels](#spec-labels) below for details):

```go
Describe("a labelled container", Label("red", "white"), Label("blue"), func() {
	It("a labelled test", Label("yellow"), func() {

	})
})
```

the labels associated with a given spec is the union of the labels attached to that spec's `It` and any of the `It`'s containers.  So `"a labelled test"` will have the labels `red`, `white`, `blue`, and `yellow`.

Labels can be arbitrary strings however they cannot include any of the following characters: `"&|!,()/"`.

#### Focus Decorator
In addition to `FDescribe` and `FIt`, specs can now be focused using the new `Focus` decorator:

```go
Describe("a focused container", Focus, func() {
  ....
})
```

#### Pending Decorator
In addition to `PDescribe` and `PIt`, specs can now be focused using the new `Pending` decorator:

```go
Describe("a focused container", Pending, func() {
  ....
})
```

#### Offset Decorator
The `Offset(uint)` decorator allows the user to change the stack-frame offset used to compute the location of the test node.  This is useful when building shared test behaviors.  For example:

```go
SharedBehaviorIt := func() {
	It("does something common and complicated", Offset(1), func() {
		...
	})
}

Describe("thing A", func() {
	SharedBehaviorIt()
})

Describe("thing B", func() {
	SharedBehaviorIt()
})
```

now, if the `It` defined in `SharedBehaviorIt` the location reported by Ginkgo will point to the line where `SharedBehaviorIt` is *invoked*.

#### FlakeAttempts Decorator
The `FlakeAttempts(uint)` decorator allows the user to flag specific tests or groups of tests as potentially flaky.  Ginkgo will run tests up to the number of times specified in `FlakeAttempts` until they pass.  For example:

```go
Describe("flaky tests", FlakeAttempts(3), func() {
	It("is flaky", func() {
		...
	})

	It("is also flaky", func() {
		...
	})

	It("is _really_ flaky", FlakeAttempts(5) func() {
		...
	})

	It("is _not_ flaky", FlakeAttempts(1), func() {
		...
	})
})
```

With this setup, `"is flaky"` and `"is also flaky"` will run up to 3 times.  `"is _really_ flaky"` will run up to 5 times.  `"is _not_ flaky"` will run only once.

Note that if `ginkgo --flake-attempts=N` is set the value passed in by the CLI will override all the decorated values.  Every test will now run up to `N` times.

### Spec Labels
Users can now label specs using the [`Label` decorator](#label-decorator).  Labels provide more fine-grained control for organizing specs and running specific subsets of labelled specs.  Labels are arbitrary strings however they cannot contain the characters `"&|!,()/"`.  A given spec inherits the labels of all its containers and any labels attached to the spec's `It`, for example:

```go
Describe("Extracting widgets", Label("integration", "extracting widgets"), func() {
	It("can extract widgets from the external database", Label("network", "slow"), func() {
		//has labels [integration, extracting widgets, network, slow]
	})

	It("can delete extracted widgets", Label("network"), func() {
		//has labels [integration, extracting widgets, network]
	})

	It("can create new widgets locally", Label("local"), func() {
		//has labels [integration, extracting widgets, local]
	})
})


Describe("Editing widgets", Label("integration", "editing widgets"), func() {
	It("can edit widgets in the external database", Label("network", "slow"), func() {
		//has labels [integration, editing widgets, network, slow]
	})

	It("errors if the widget does not exist", Label("network"), func() {
		//has labels [integration, editing widgets, network]
	})
})
```

In addition an entire suite can be decorated by passing a `Label` decorator to `RunSpecs`:

```go
RunSpecs(t, "My Suite", Label("top-level-label", "labels-all-specs"))
```

You can filter by label using the new `ginkgo --label-filter` flag.  Label filter accepts a simple filter language that supports the following:

- The `&&` and `||` logical binary operators representing AND and OR operations.
- The `!` unary operator representing the NOT operation.
- The `,` binary operator equivalent to `||`.
- The `()` for grouping expressions.
- All other characters will match as label literals.  Label matches are case intensive and trailing and leading whitespace is trimmed.
- Regular expressions can be provided using `/REGEXP/` notation.

For example:

- `ginkgo --label-filter=integration` will match any specs with the `integration` label.
- `ginkgo --label-filter=!slow` will avoid any tests labelled `slow`.
- `ginkgo --label-filter=(local || network) && !slow` will run any specs labelled `local` and `network` but without the `slow` label.
- `ginkgo --label-filter=/widgets/ && !slow` will run any specs with a label that matches the regular expression `widgets` but does not include the `slow` label.  This would match both the `extracting widgets` and `editing widgets` labels in our example above.

### DeferCleanup

`DeferCleanup` allows users to move cleanup code out of `AfterEach/AfterAll/AfterSuite` and closer to the setup code that needs to be cleaned up.  Based on the context in which it is called, `DeferCleanup` will effectively register a dynamic `AfterEach/AfterAll/AfterSuite` node to clean up after the test/test group/suite.  The [docs](https://github.com/onsi/ginkgo/blob/ver2/docs/index.md#cleaning-up-after-tests) have more detailed examples.

`DeferCleanup` allows `GinkgoT()` to more fully implement the `testing.T` interface.  `Cleanup`, `TempDir`, and `Setenv` are now all supported.

### Aborting the Test Suite
Users can now signal that the entire test suite should abort via `AbortSuite(message string, skip int)`.  This will fail the current test and skip all subsequent tests.

### Improved: --fail-fast
`ginkgo --fail-fast` now interrupts all test processes when a failure occurs and the tests are running in parallel.

### CLI Flags
Ginkgo's CLI flags have been rewritten to provide clearer, better-organized documentation.  In addition, Ginkgo v1 was mishandling several go cli flags.  This is now resolved with clear distinctions between flags intended for compilation time and run-time.  As a result, users can now generate `memprofile`s and `cpuprofile`s using the Ginkgo CLI.  Ginkgo 2.0 will automatically merge profiles generated by running tests in parallel (i.e. across multiple processes) and will allow you to choose between having profiles stored in individual package directories, or collected in one place using the `-output-dir` flag.  See [Changed: Profiling Support](#improved-profiling-support) for more details.

### Expanded GinkgoWriter Functionality
The `GinkgoWriter` is used to write output that is only made visible if a test fails, or if the user runs in verbose mode with `ginkgo -v`.

In Ginkgo 2.0 `GinkgoWriter` now has:
	- Three new convenience methods `GinkgoWriter.Print(a ...interface{})`, `GinkgoWriter.Println(a ...interface{})`, `GinkgoWriter.Printf(format string, a ...interface{})`  These are equivalent to calling the associated `fmt.Fprint*` functions and passing in `GinkgoWriter`.
	- The ability to tee to additional writers.  `GinkgoWriter.TeeTo(writer)` will send any future data written to `GinkgoWriter` to the passed in `writer`.  You can attach multiple `io.Writer`s for `GinkgoWriter` to tee to.  You can remove all attached writers with `GinkgoWriter.ClearTeeWriters()`.

	Note that _all_ data written to `GinkgoWriter` is immediately forwarded to attached tee writers regardless of where a test passes or fails.

### Improved: Reporting Infrastructure
Ginkgo V2 provides an improved reporting infrastructure that [replaces and improves upon Ginkgo V1's support for custom reporters](#removed-custom-reporters).  Here are a few distinct use-cases that the new reporting infrastructure supports:

#### Generating machine-readable reports
Ginkgo now natively supports generating and aggregating reports in a number of machine-readable formats - and these reports can be generated and managed by simply passing `ginkgo` command line flags.

Ginkgo V2 introduces a new JSON format that faithfully captures all avialable information about a Ginkgo test suite.  JSON reports can be generated via `ginkgo --json-report=out.json`.  The resulting JSON file encodes an array of `types.Report`.  Each entry in that array lists detailed information about the test suite and includes a list of `types.SpecReport` that captures detailed information about each spec.  These types are documented [here](https://github.com/onsi/ginkgo/blob/ver2/types/types.go).

Ginkgo also supports generating JUnit reports with `ginkgo --junit-report=out.xml` and Teamcity reports with `ginkgo --teamcity-report=out.teamcity`.  In addition, Ginkgo V2's JUnit reporter has been improved and is now more conformant with the JUnit specification.

Ginkgo follows the following rules when generating reports using these new `--FORMAT-report` flags:
- By default, a single report file per format is generated at the passed-in file name.  This single report merges all the reports generated by each individual suite.
- If `-output-dir` is set: the report files are placed in the specified `-output-dir` directory.
- If `-keep-separate-reports` is set, the individual reports generated by each test suite are not merged.  Instead, individual report files will appear in each package directory.  If `-output-dir` is _also_ set these individual files are copied into the `-output-dir` directory and namespaced with `PACKAGE_NAME_{REPORT}`.

#### Generating Custom Reports when a test suite completes
Ginkgo now provides a new node, `ReportAfterSuite`, with the following properties and constraints:
- `ReportAfterSuite` nodes are passed a function that takes a `types.Report`:
  ```go
  var _ = ReportAfterSuite(func(report types.Report) {
  	// do stuff with report
  })
  ```
- These functions are called exactly once at the end of the test suite after any `AfterSuite` nodes have run.  When running in parallel the functions are called on the primary Ginkgo process after all other processes have finished and is guaranteed to have an aggregated copy of `types.Report` that includes all `SpecReport`s from all the Ginkgo parallel processes.
- If a failure occurs in `ReportAfterSuite` it is reported in reports sent to subsequent `ReportAfterSuite`s.  In particular, it is reported as part of Ginkgo's default output and is in included in any reports generated by the `--FORMAT-report` flags described above.
- `ReportAfterSuite` nodes **cannot** be interrupted by the user.  This is to ensure the integrity of generated reports... so be careful what kind of code you put in there!
- Multiple `ReportAfterSuite` nodes can be registered per test suite, but all must be defined at the top-level of the suite.

`ReportAfterSuite` is useful for users who want to emit a custom-formatted report or register the results of the test run with an external service.

#### Capturing report information about each spec as the test suite runs
Ginkgo also provides a new node, `ReportAfterEach`, with the following properties and constraints:
- `ReportAfterEach` nodes are passed a function that takes a `types.SpecReport`:
  ```go
  var _ = ReportAfterEach(func(specReport types.SpecReport) {
  	// do stuff with specReport
  })
  ```
- `ReportAfterEach` nodes are called after a spec completes (i.e. after any `AfterEach` nodes have run).  `ReportAfterEach` nodes are **always** called - even if the test has failed, is marked pending, or is skipped.  In this way, the user is guaranteed to have access to a report about every spec defined in a suite.
- If a failure occurs in `ReportAfterEach`, the spec in question is marked as failed.  Any subsequently defined `ReportAfterEach` block will receive an updated report that includes the failure.  In general, though, assertions about your code should go in `AfterEach` nodes.
- `ReportAfterEach` nodes **cannot** be interrupted by the user.  This is to ensure the integrity of generated reports... so be careful what kind of code you put in there!
- `ReportAfterEach` nodes can be placed in any container node in the suite's hierarchy - in this way the follow the same semantics as `AfterEach` blocks.
- When running in parallel, `ReportAfterEach` nodes will run on the Ginkgo process that is running the spec being reported on.  This means that multiple `ReportAfterEach` blocks can be running concurrently on independent processes.

`ReportAfterEach` is useful if you need to stream or emit up-to-date information about the test suite as it runs.

Ginkgo also provides `ReportBeforeEach` which is called before the test runs and receives a preliminary `types.SpecReport` - the state of this report will indicate whether the test will be skipped or is marked pending.

### New: Report Entries
Ginkgo V2 supports attaching arbitrary data to individual spec reports.  These are called `ReportEntries` and appear in the various report-related data structures (e.g. `Report` in `ReportAfterSuite` and `SpecReport` in `ReportAfterEach`) as well as the machine-readable reports generated by `--json-report`, `--junit-report`, etc.  `ReportEntries` are also emitted to the console by Ginkgo's reporter and you can specify a visibility policy to control when this output is displayed.

You attach data to a spec report via

```go
AddReportEntry(name string, args ...interface{})
```

`AddReportEntry` can be called from any runnable node (e.g. `It`, `BeforeEach`, `BeforeSuite`) - but not from the body of a container node (e.g. `Describe`, `Context`).

`AddReportEntry` generates `ReportEntry` and attaches it to the current running spec.  `ReportEntry` includes the passed in `name` as well as the time and source location at which `AddReportEntry` was called.  Users can also attach a single object of arbitrary type to the `ReportEntry` by passing it into `AddReportEntry` - this object is wrapped and stored under `ReportEntry.Value` and is always included in the suite's JSON report.

You can access the report entries attached to a spec by getting the `CurrentSpecReport()` or registering a `ReportAfterEach()` - the returned report will include the attached `ReportEntries`.  You can fetch the value associated with the `ReportEntry` by calling `entry.GetRawValue()`.  When called in-process this returns the object that was passed to `AddReportEntry`.  When called after hydrating a report from JSON `entry.GetRawValue()` will include a parsed JSON `interface{}` - if you want to hydrate the JSON yourself into an object of known type you can `json.Unmarshal([]byte(entry.Value.AsJSON), &object)`.

#### Supported Args
`AddReportEntry` supports the `Offset` and `CodeLocation` decorators.  These will control the source code location associated with the generated `ReportEntry`.  You can also pass in a `time.Time` to override the `ReportEntry`'s timestamp. It also supports passing in a `ReportEntryVisibility` enum to control the report's visibility (see below).

#### Controlling Output
By default, Ginkgo's console reporter will emit any `ReportEntry` attached to a spec.  It will emit the `ReportEntry` name, location, and time.  If the `ReportEntry` value is non-nil it will also emit a representation of the value.  If the value implements `fmt.Stringer` or `types.ColorableStringer` then `value.String()` or `value.ColorableString()` (which takes precedence) is used to generate the representation, otherwise Ginkgo uses `fmt.Sprintf("%#v", value)`. 

You can modify this default behavior by passing in one of the `ReportEntryVisibility` enum to `AddReportEntry`:

- `ReportEntryVisibilityAlways`: the default behavior - the `ReportEntry` is always emitted.
- `ReportEntryVisibilityFailureOrVerbose`: the `ReportEntry` is only emitted if the spec fails or is run with `-v` (similar to `GinkgoWriter`s behavior).
- `ReportEntryVisibilityNever`: the `ReportEntry` is never emitted though it appears in any generated machine-readable reports (e.g. by setting `--json-report`).

The console reporter passes the string representation of the `ReportEntry.Value` through Ginkgo's `formatter`.  This allows you to generate colorful console output using the color codes documented in `github.com/onsi/ginkgo/formatter/formatter.go`.  For example:

```go
type StringerStruct struct {
	Label string
	Count int
}

// ColorableString for ReportEntry to use
func (s StringerStruct) ColorableString() string {
	return fmt.Sprintf("{{red}}%s {{yellow}}{{bold}}%d{{/}}", s.Label, s.Count)
}

// non-colorable String() is used by go's string formatting support but ignored by ReportEntry
func (s StringerStruct) String() string {
	return fmt.Sprintf("%s %d", s.Label, s.Count)
}

It("is reported", func() {
	AddReportEntry("Report", StringerStruct{Label: "Mahomes", Count: 15})
})
```

Will emit a report that has the word "Mahomes" in red and the number 15 in bold and yellow.

Lastly, it is possible to pass a pointer into `AddReportEntry`.  Ginkgo will compute the string representation of the passed in pointer at the last possible moment - so any changes to the object _after_ it is reported will be captured in the final report.  This is useful for building libraries on top of `AddReportEntry` - users can simply register objects when they're created and any subsequent mutations will appear in the generated report.

### New: Table-level Entry Descriptions

Table `Entry`s can now opt-into table-level descriptions.  Simply pass `nil` as the first argument into `Entry`.  By default, Ginkgo will generate an `Entry` description from the `Entry`s parameters.  You can also provide a string-returning function to `DescribeTable` which will be used to generate the description for these entries.  There's also a new `EntryDescription` decorator that can be passed in to `DescribeTable` - `EntryDescription` wraps a format string that can be used to format the parameters associated with each `Entry` to generate it's description.

For example:

```go
var _ = Describe("Math", func() {
    DescribeTable("addition",
        func(a, b, c int) {
            Expect(a+b).To(Equal(c))
        },
        EntryDescription("%d + %d = %d")
        Entry(nil, 1, 2, 3),
        Entry(nil, -1, 2, 1),
        Entry("zeros", 0, 0, 0),
        Entry(EntryDescription("%[3]d = %[1]d + %[2]d"), 2, 3, 5)
        Entry(func(a, b, c int) string {fmt.Sprintf("%d = %d", a + b, c)}, 4, 3, 7)
    )
})
```

Will generate entries named: `1 + 2 = 3`, `-1 + 2 = 1`, `zeros`, `5 = 2 + 3`, and `7 = 7`.

### Improved: Profiling Support
Ginkgo V1 was incorrectly handling Go test's various profiling flags (e.g. -cpuprofile, -memprofile).  This has been fixed in V2.  In fact, V2 can capture profiles for multiple packages (e.g. ginkgo -r -cpuprofile=profile.out will work).

When generating profiles for `-cpuprofile=FILE`, `-blockprofile=FILE`, `-memprofile=FILE`, `-mutexprofile=FILE`, and `-execution-trace=FILE` (Ginkgo's alias for `go test -test.trace`) the following rules apply:

- If `-output-dir` is not set: each profile generates a file named `$FILE` in the directory of each package under test.
- If `-output-dir` is set: each profile generates a file in the specified `-output-dir` directory. named `PACKAGE_NAME_$FILE`

### Improved: Cover Support
Coverage reporting is much improved in 2.0:

- `ginkgo -cover -p` now emits code coverage after the test completes, just like `ginkgo -cover` does in series.
- When running across mulitple packages (e.g. `ginkgo -r -cover`) ginkgo will now emit a composite coverage statistic that represents the total coverage across all test suites run.  (Note that this is disabled if you set `-keep-separate-coverprofiles`).

In addition, Ginkgo now follows the following rules when generating cover profiles using `-cover` and/or `-coverprofile=FILE`:

- By default, a single cover profile is generated at `FILE` (or `coverprofile.out` if `-cover-profile` is not set but `-cover` is set).  This includes the merged results of all the cover profiles reported by each suite.
- If `-output-dir` is set: the `FILE` is placed in the specified `-output-dir` directory.
- If `-keep-separate-coverprofiles` is set, the individual coverprofiles generated by each package are not merged and, instead, a file named `FILE` will appear in each package directory.  If `-output-dir` is _also_ set these files are copied into the `-output-dir` directory and namespaced with `PACKAGE_NAME_{FILE}`.

### New: --repeat
Ginkgo can now repeat a test suite N additional times by running `ginkgo --repeat=N`.  This is similar to `go test -count=N+1` and is a variant of `ginkgo --until-it-fails` that can be run in CI environments to repeat test runs to suss out flakey tests.

Ginkgo requires the tests to succeed during each repetition in order to consider the test run a success.

### New: --focus-file and --skip-file
You can now tell Ginkgo to only run specs that match (or don't match) a given file filter.  You can filter by filename as well as file:line.  See the [Filtering Specs](onsi.github.io/ginkgo/#filtering-specs) documentation for more details.

### Improved: windows support for capturing stdout and stderr
In V1 Ginkgo would run windows tests in parallel with the `--stream` option.  This would result in hard-to-understand interleaved output.  The reason behind this design choice was that it proved challenging to intercept all stdout and stderr output on Windows.  V2 implements a best-effort output interception scheme for windows that entails reassigning the global `os.Stdout` and `os.Stderr` variables.  While not as bullet-proof as the Unix `syscall.Dup2` based implementation, this is likely good enough for most usecases and allows Ginkgo support on Windows to come into parity with unix.

## Minor Additions and Improvements
- `BeforeSuite` and `AfterSuite` no longer run if all tests in a suite are skipped.
- The entire suite is skipped if `Skip()` is called in `BeforeSuite`.
- Any output generated by `SynchronizedBeforeSuite`'s proc-1 function will now be immediately streamed to stdout, even when running in parallel.  This is useful to debug complex long-running `SynchronizedBeforeSuite` setups.
- Ginkgo's performance should be improved now when running multiple suites in a context where the go mod dependencies have not been fetched yet (e.g. on CI).  Previously, Ginkgo would compile suites in parallel resulting in substantial slowdown when fetching the dependencies in parallel.  In V2, Ginkgo compiles the first suite in series before compiling the remaining suites in parallel.
- Ginkgo can now catch several common user gotchas and emit a helpful error.
- Tables can now accept slices of []TableEntry in addition to individual entries.  This allows for entries to be reused across different tables.
- Error output is clearer and consistently links to relevant sections in the documentation.
- `By` now emits a timestamp.  It also registers a `ReportEntry` that appears in the suite report as structured data.  If passed a callback, `By` will now time the callback and include the duration in the suite report.
- Test randomization is now more stable as tests are now sorted deterministcally on file_name:line_number first (previously they were sorted on test text which could not guarantee a stable sort).
- A new "very verbose" setting is now available.  Setting `-vv` implies `-v` but also causes skipped tests to be emitted.
- Ginkgo's OutputInterceptor (the component that intercepts stdout/stderr when running in parallel) should now be more performant and better handle edge cases.  It can be paused and resumed with PauseOutputInterception() and ResumeOutputInterception() and disabled entirely with --output-interceptor-mode=none.

# Changes

## Major Changes
These are major changes that will need user intervention to migrate succesfully.

### Removed: Async Testing
As described in the [Ginkgo 2.0 Proposal](https://docs.google.com/document/d/1h28ZknXRsTLPNNiOjdHIO-F2toCzq4xoZDXbfYaBdoQ/edit#heading=h.mzgqmkg24xoo) the Ginkgo 1.x implementation of asynchronous testing using a `Done` channel was a confusing source of test-pollution.  It is removed in Ginkgo 2.0.

In Ginkgo 2.0 tests of the form:

```go
It("...", func(done Done) {
	// user test code to run asynchronously
	close(done) //signifies the test is done
}, timeout)
```

will emit a deprecation warning and will run **synchronously**.  This means the `timeout` will not be enforced and the status of the `Done` channel will be ignored - a test that hangs will hang indefinitely.

#### Migration Strategy:
We recommend users make targeted use of Gomega's [Asynchronous Assertions](https://onsi.github.io/gomega/#making-asynchronous-assertions) to better test asynchronous behavior.

As a first migration pass that produces **equivalent behavior** users can replace asynchronous tests with:

```go
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
```go
Measure(..., func(b Benchmarker) {
	// user benchmark code
})
```

will emit a deprecation warning and **will no longer run**.

#### Migration Strategy:
Gomega now provides a benchmarking subpackage called `gmeasure`.  Users should migrate to `gmeasure` by replacing `Measure` nodes with `It` nodes that create `gmeasure.Experiment`s and record values/durations.  To generate output in Ginkgo reports add the `experiment` as a `ReportEntry` via `AddReportEntry(experiment.Name, experiment)`.

### Removed: Custom Reporters
Ginkgo 2.0 removes support for Ginkgo 1.X's custom reporters - they behaved poorly when running in parallel and represented unnecessary and error-prone boiler plate for users who simply wanted to produce machine-readable reports.  Instead, the reporting infrastructure has been significantly improved to enable simpler support for the most common use-cases and custom reporting needs.

Please read through the [Improved: Reporting Infrastructure](#improved-reporting-infrastructure) section to learn more.  For users with custom reporters, follow the migration guide below.

#### Migration Strategy:
In Ginkgo 2.0 both `RunSpecsWithDefaultAndCustomReporters` and `RunSpecsWithCustomReporters` have been deprecated.  Users must call `RunSpecs` instead.

If you were using custom reporters to generate JUnit or Teamcity reports, simply call `RunSpecs` and [invoke your tests with the new `--junit-report` and/or `--teamcity-report` flags](#generating-machine-readable-reports).  Note that unlike the 1.X JUnit and Teamcity reporters, these flags generate unified reports for all test suites run (though you can adjust this with the `--keep-separate-reports` flag) and take care of aggregating reports from parallel processes for you.

If you've written your own custom reporter, [add a `ReportAfterSuite` node](#generating-custom-reports-when-a-test-suite-completes) and process the `types.Report` that it provides you.  If you'd like to continue using your custom reporter you can simply call `reporters.ReportViaDeprecatedReporter(reporter, report)` in `ReportAfterSuite` - though we recommend actually changing your code's logic to use the `types.Report` object directly as `reporters.ReportViaDeprecatedReporter` will be removed in a future release of Ginkgo 2.X.  Unlike 1.X custom reporters which are called concurrently by independent parallel processes when running in parallel, `ReportAFterSuite` is called exactly once per suite and is guaranteed to have aggregated information from all parallel processes.

Alternatively, you can use the new `--json-report` flag to produce a machine readable JSON-format report that you can post-process after the test completes.

Finally, if you still need the real-time reporting capabilities that 1.X's custom reporters provided you can use [`ReportBeforeEach` and `ReportAfterEach`](#capturing-report-information-about-each-spec-as-the-test-suite-runs) to get information about each spec as it completes.

### Changed: First-class Support for Table Testing
The table extension has been moved into the core Ginkgo DSL and the table functionality has been improved while maintaining backward compatibility.  Users no longer need to `import "github.com/onsi/ginkgo/v2/extenstions/table"`.  Instead the table DSL is automatically pulled in by importing `"github.com/onsi/ginkgo/v2"`.

#### Migration Strategy:
Remove `"github.com/onsi/ginkgo/v2/extensions/table` imports.  Code that was dot-importing both Ginkgo and the table extension should automatically work.  If you were not dot-importing you will need to replace references to `table.DescribeTable` and `table.Entry` with `ginkgo.DescribeTable` and `ginkgo.Entry`.


### Changed: CurrentGinkgoTestDescription()
`CurrentGinkgoTestDescription()` has been deprecated and will be removed in a future release.  The method was returning a processed object that included a subset of information available about the running test.

It has been replaced with `CurrentSpecReport()` which returns the full-fledge `types.SpecReport` used by Ginkgo's reporting infrastructure.  To help users migrate, `types.SpecReport` now includes a number of helper methods to make it easier to extract information about the running test.

#### Migration Strategy:
Replace any calls to `CurrentGinkgoTestDescription()` with `CurrentSpecReport()` and use the struct fields or helper methods on the returned `types.SpecReport` to get the information you need about the current test.

### Changed: availability of Ginkgo's configuration
In v1 Ginkgo's configuration could be accessed by importing the `config` package and accessing the globally available `GinkgoConfig` and `DefaultReporterConfig` objects.  This is no longer supported in V2.

V1 also allowed mutating the global config objects which could lead to strange behavior if done within a test.  This too is no longer supported in V2.

#### Migration Strategy:
Instead, configuration can be accessed using the DSL's `GinkgoConfiguration()` function.  This will return a `types.SuiteConfig` and `types.ReporterConfig`.  Users generally don't need to access this configuration - the most commonly used fields by end users are already made available via `GinkgoRandomSeed()` and `GinkgoParallelProcess()`.

It is generally recommended that users use the CLI to configure Ginkgo as some aspects of configuration must apply to the CLI as well as the suite under tests - nonetheless there are contexts where it is necessary to change Ginkgo's configuration programatically.  V2 supports this by allowing users to pass updated configuration into `RunSpecs`:

```go
func TestMySuite(t *testing.T)  {
	RegisterFailHanlder(gomega.Fail)
	// fetch the current config
	suiteConfig, reporterConfig := GinkgoConfiguration()
	// adjust it
	suiteConfig.SkipStrings = []string{"NEVER-RUN"}
	reporterConfig.FullTrace = true
	// pass it in to RunSpecs
	RunSpecs(t, "My Suite", suiteConfig, reporterConfig)
}
```

### Renamed: GinkgoParallelNode
`GinkgoParallelNode` has been renamed to `GinkgoParallelProcess` to reduce confusion around the word `node` and better capture Ginkgo's parallelization mechanism.

#### Migration strategy:
Change all instance of `GinkgoParallelNode()` to `GinkgoParallelProcess()`

### Changed: Command Line Flags
All camel case flags (e.g. `-randomizeAllSpecs`) are replaced with kebab case flags (e.g. `-randomize-all-specs`) in Ginkgo 2.0.  The camel case versions continue to work but emit a deprecation warning.

#### Migration Strategy:
Users should update any scripts they have that invoke the `ginkgo` cli from camel case to kebab case (:camel: :arrow_right: :oden:).

### Removed: -stream
`-stream` was originally introduce in Ginkgo 1.x to force parallel test processes to emit output simultaneously in order to help debug hanging test issues.  With improvements to Ginkgo's interrupt handling and parallel test reporting this behavior is no longer necessary and has been removed.

### Removed: -notify
`-notify` instructed Ginkgo to emit desktop notifications on linux and MacOS.  This feature was rarely used and has been removed.

### Removed: -noisyPendings and -noisySkippings
Both these flags tweaked the reporter's behavior for pending and skipped tests but never worked quite right.  Now the user can specify between four verbosity levels.  `--succinct`, no verbosity setting, `-v`, and `-vv`.  Specifically, when run with `-vv`  skipped tests will emit their titles and code locations - otherwise skipped tests are silent.

### Changed: -slowSpecThreshold
`-slowSpecThreshold` is now `-slow-spec-threshold` and takes a `time.Duration` (e.g. `5s` or `3m`) instead of a `float64` number of seconds.

### Renamed: -reportPassed
`-reportPassed` is now `--always-emit-ginkgo-writer` which better captures the intent of the flag; namely to always emit any GinkgoWriter content, even if the spec has passed.

### Removed: -debug
The `-debug` flag has been removed.  It functioned primarily as a band-aid to Ginkgo V1's poor handling of stuck parallel tests. The new [interrupt behavior](#interrupt-behavior) in V2 resolves the root issues behind the `-debug` flag.

### Removed: -regexScansFilePath
`-regexScansFilePath` allowed users to have the `-focus` and `-skip` regular expressions apply to filenames.  It is now removed in favor of `-focus-file` and `-skip-file` which provide more granular and explicit control over focusing/skipping files and line numbers.

#### Migration Strategy:
Users should remove -stream from any scripts they have that invoke the `ginkgo` cli.

### Removed: ginkgo nodot
The `ginkgo nodot` subcommand in V1, along with the `--nodot` flags for `ginkgo bootstrap` and `ginkgo generate` were provided to allow users to avoid a `.` import of Ginkgo and Gomega but still have access to the exported variables and types at the top-level.  This was implemented by defining top-level aliases that pointed to the objects and types in the imported Ginkgo and Gomega libraries in the user's bootstrap file.  In practice most users either dot-import Ginkgo and Gomega, or they don't and use the imported package name to refer to objects and types instead.  V2 removes the support generating and maintaining these alias lists.  `--nodot` remainds for `ginkgo bootstrap` and `ginkgo generate` and it simply avoids dot-importing Ginkgo and Gomega.

As a result of this change custom bootstrap and generate templates may need to be updated:

1. `ginkgo generate` templates should no longer reference `{{.IncludeImports}}`.  Instead they should `import {{.GinkgoImport}}` and `import {{.GomegaImport}}`.
2. Both `ginkgo generate` and `ginkgo boostrap` templates can use `{{.GinkgoPackage}}` and `{{.GomegaPackage}}` to correctly reference any names exported by Ginkgo or Gomega.  For example:

	```go

	import (
		{{.GinkgoImport}}
		{{.GomegaImport}}
	}

	var _ = {{.GinkgoPackage}}It("is templated", func() {
		{{.GomegaPackage}}Expect(foo).To({{.GomegaPackage}}Equal(bar))
	})

	```

	will generate the correct output if `--nodot` is specified by the user.


### Removed: ginkgo convert
The `ginkgo convert` subcommand in V1 could convert an existing set of Go tests into a Ginkgo test suite, wrapping each `TestX` function in an `It`.  This subcommand added complexity to the codebase and was infrequently used.  It has been removed.  Users who want to convert tests suites over to Ginkgo will need to do so by hand.

## Minor Changes
These are minor changes that will be transparent for most users.

- `"top level"` is no longer the first element in `types.SpecReport.NodeTexts`.  This will only affect users who write custom reporters.

- The output format of Ginkgo's Default Reporter has changed in numerous subtle ways to improve readability and the user experience.  Users who were scraping Ginkgo output programatically may need to change their scripts or use the new JSON formatted report option.

- When running in series and verbose mode (i.e. `ginkgo -v`) GinkgoWriter output is emitted in real-time (existing behavior) but also emitted in the failure message for failed tests.  This allows for consistent failure messages regardless of verbosity settings and also makes it possible for the resulting JSON report to include captured GinkgoWriter information.

- Removed `ginkgo blur` alias.  Use `ginkgo unfocus` instead.

{% endraw  %}

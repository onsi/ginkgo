---
layout: default
title: Ginkgo
---
[Ginkgo](http://github.com/onsi/ginkgo) is a BDD-style Golang testing framework built to help you efficiently write expressive and comprehensive tests.  It is best paired with the [Gomega](http://github.com/onsi/gomega) matcher library but is designed to be matcher-agnostic.  

These docs are written assuming you'll be using Gomega with Ginkgo.  They also assume you know your way around Go and have a good mental model for how Go organizes packages under `$GOPATH`.

---

##Getting Ginkgo

Just `go get` it:

    $ go get github.com/onsi/ginkgo
    $ go get github.com/onsi/gomega

To install the Ginkgo CLI (*recommended, but optional*):

    $ go install github.com/onsi/ginkgo/ginkgo

this installs the `ginkgo` executable under `$GOPATH/bin` -- you'll want that on your `$PATH`.

---

##Getting Started: Writing Your First Test
Ginkgo hooks into Go's existing `testing` infrastructure.  This allows you to run a Ginkgo suite using `go test`.

> This also means that Ginkgo tests can live alongside traditional Golang `testing` tests.  Both `go test` and `ginkgo` will run all the tests in your suite.

###Bootstrapping a Suite
To write Ginkgo tests for a package you must first bootstrap a Ginkgo test suite.  Say you have a package named `books`:

    $ cd path/to/books
    $ ginkgo bootstrap

This will generate a file named `books_suite_test.go` containing:

    package books_test

    import (
        . "github.com/onsi/ginkgo"
        . "github.com/onsi/gomega"

        "testing"
    )

    func TestBooks(t *testing.T) {
        RegisterFailHandler(Fail)
        RunSpecs(t, "Books Suite")
    }

Let's break this down:

- Go allows us to specify the `books_test` package alongside the `books` package.  Using `books_test` instead of `books` allows us to respect the encapsulation of the `books` package: your tests will need to import `books` and access it from the outside, like any other package.  This is preferred to reaching into the package and testing its internals and leads to more behavioral tests.  You can, of course, opt out of this -- just change `package books_test` to `package books`
- `TestBooks` is a `testing` test.  The Go test runner will run this function when you run `go test`.
- `RegisterFailHandler(Fail)`: A Ginkgo test signals failure by calling Ginkgo's `Fail(description string)` function.  We pass this function to Gomega using `RegisterFailHandler`.  This is the sole connection point between Ginkgo and Gomega.
- `RunSpecs(t *testing.T, suiteDescription string)` tells Ginkgo to start the test suite.  Ginkgo will automatically fail the `testing.T` if any of your specs fail.

At this point you can run your suite:

    $ ginkgo #or go test

    === RUN TestBootstrap

    Running Suite: Books Suite
    ==========================
    Random Seed: 1378936983

    Will run 0 of 0 specs


    Ran 0 of 0 Specs in 0.000 seconds
    SUCCESS! -- 0 Passed | 0 Failed | 0 Pending | 0 Skipped

    --- PASS: TestBootstrap (0.00 seconds)
    PASS
    ok      books   0.019s

###Adding Specs to a Suite
An empty test suite is not very interesting.  While you can start to add tests directly into `books_suite_test.go` you'll probably prefer to separate your tests into separate files (especially for packages with multiple files).  Let's add a test file for our `book.go` model:

    $ ginkgo generate book

This will generate a file named `book_test.go` containing:

    package books_test

    import (
        . "/path/to/books"
        . "github.com/onsi/ginkgo"
        . "github.com/onsi/gomega"
    )

    var _ = Describe("Book", func() {

    })

Let's break this down:

- We import the `ginkgo` and `gomega` packages into the global name space.  While incredibly convenient, this is not - strictly speaking - necessary.
- Similarly, we import the `books` package since we are using the special `books_test` package to isolate our tests from our code.  For convenience we import the `books` package into the namespace.  You can opt out of either these decisions by editing the generated test file.
- We add a *top-level* describe container using Ginkgo's `Describe(text string, body func()) bool` function.  The `var _ = ...` trick allows us to evaluate the Describe at the top level without having to wrap it in a `func init() {}`

The function in the `Describe` will contain our specs.  Let's add a few now to test loading books from JSON:

    var _ = Describe("Book", func() {
        var (
            longBook  Book
            shortBook Book
        )

        BeforeEach(func() {
            longBook = Book{
                Title:  "Les Miserables",
                Author: "Victor Hugo",
                Pages:  1488,
            }

            shortBook = Book{
                Title:  "Fox In Socks",
                Author: "Dr. Seuss",
                Pages:  24,
            }
        })

        Describe("Categorizing book length", func() {
            Context("With more than 300 pages", func() {
                It("should be a novel", func() {
                    Expect(longBook.CategoryByLength()).To(Equal("NOVEL"))
                })
            })

            Context("With fewer than 300 pages", func() {
                It("should be a short story", func() {
                    Expect(shortBook.CategoryByLength()).To(Equal("SHORT STORY"))
                })
            })
        })
    })

Let's break this down:

- Ginkgo makes extensive use of closures to allow you to build descriptive test suites.
- You should make use of `Describe` and `Context` containers to expressively organize the behavior of your code.
- You can use `BeforeEach` to set up state for your specs.  You use `It` to specify a single spec.  
- In order to share state between a `BeforeEach` and an `It` you use closure variables, typically defined at the top of the most relevant `Describe` or `Context` container.
- We use Gomega's `Expect` syntax to make expectations on the `CategoryByLength()` method.

Assuming a `Book` model with this behavior, running the tests will yield:

    $ ginkgo # or go test
    === RUN TestBootstrap

    Running Suite: Books Suite
    ==========================
    Random Seed: 1378938274

    Will run 2 of 2 specs

    ••
    Ran 2 of 2 Specs in 0.000 seconds
    SUCCESS! -- 2 Passed | 0 Failed | 0 Pending | 0 Skipped

    --- PASS: TestBootstrap (0.00 seconds)
    PASS
    ok      books   0.025s

Success!

###Marking Specs as Failed

While you typically want to use a matcher library, like [Gomega](https://github.com/onsi/gomega), to make assertions in your specs, Ginkgo provides a simple, global, `Fail` function that allows you to mark a spec as failed.  Just call:

    Fail("Failure reason")

and Ginkgo will take the rest.  More details about `Fail` and about using matcher libraries other than Gomega can be found in the [Using Other Matcher Libraries](#using_other_matcher_libraries) section.
 
---

##Structuring Your Specs

Ginkgo makes it easy to write expressive specs that describe the behavior of your code in an organized manner.  You use `Describe` and `Context` containers to organize your `It` specs and you use `BeforeEach` and `AfterEach` to build up and tear down common set up amongst your tests.

### Individual Specs: `It`
You can add a single spec by placing an `It` block within a `Describe` or `Context` container block:

    var _ = Describe("Book", func() {
        It("can be loaded from JSON", func() {
            book := NewBookFromJSON(`{
                "title":"Les Miserables",
                "author":"Victor Hugo",
                "pages":1488
            }`)

            Expect(book.Title).To(Equal("Les Miserables"))
            Expect(book.Author).To(Equal("Victor Hugo"))
            Expect(book.Pages).To(Equal(1488))
        })
    })

### Extracting Common Setup: `BeforeEach`
You can remove duplication and share common setup across tests using `BeforeEach` blocks:

    var _ = Describe("Book", func() {
        var book Book

        BeforeEach(func() {
            book = NewBookFromJSON(`{
                "title":"Les Miserables",
                "author":"Victor Hugo",
                "pages":1488
            }`)
        })
        
        It("can be loaded from JSON", func() {
            Expect(book.Title).To(Equal("Les Miserables"))
            Expect(book.Author).To(Equal("Victor Hugo"))
            Expect(book.Pages).To(Equal(1488))
        })

        It("can extract the author's last name", func() {
            Expect(book.AuthorLastName()).To(Equal("Hugo"))
        })
    })

The `BeforeEach` is run before each spec thereby ensuring that each spec has a pristine copy of the state.  Common state is shared using closure variables (`var book Book` in this case).  You can also perform clean up in `AfterEach` blocks.

It is also common to place assertions within `BeforeEach` and `AfterEach` blocks.  These assertions can, for example, assert that no errors occured while preparing the state for the spec.

### Organizing Specs With Containers: `Describe` and `Context`

Ginkgo allows you to expressively organize the specs in your suite using `Describe` and `Context` containers:

    var _ = Describe("Book", func() {
        var (
            book Book
            err error
        )

        BeforeEach(func() {
            book, err = NewBookFromJSON(`{
                "title":"Les Miserables",
                "author":"Victor Hugo",
                "pages":1488
            }`)
        })

        Describe("loading from JSON", func() {
            Context("when the JSON parses succesfully", func() {
                It("should populate the fields correctly", func() {
                    Expect(book.Title).To(Equal("Les Miserables"))
                    Expect(book.Author).To(Equal("Victor Hugo"))
                    Expect(book.Pages).To(Equal(1488))
                })

                It("should not error", func() {
                    Expect(err).NotTo(HaveOccured())
                })
            })

            Context("when the JSON fails to parse", func() {
                BeforeEach(func() {
                    book, err = NewBookFromJSON(`{
                        "title":"Les Miserables",
                        "author":"Victor Hugo",
                        "pages":1488oops
                    }`)
                })

                It("should return the zero-value for the book", func() {
                    Expect(book).To(BeZero())
                })

                It("should error", func() {
                    Expect(err).To(HaveOccured())
                })
            })
        })
        
        Describe("Extracting the author's last name", func() {
            It("should correctly identify and return the last name", func() {
                Expect(book.AuthorLastName()).To(Equal("Hugo"))
            })
        })
    })

You use `Describe` blocks to describe the individual behaviors of your code and `Context` blocks to excercise those behaviors under different circumstances.  In this example we `Describe` loading a book from JSON and specify two `Context`s: when the JSON parses succesfully and when the JSON fails to parse.  Semantic differences aside, the two container types have identical behavior.

When nesting `Describe`/`Context` blocks the `BeforeEach` blocks for all the container nodes surrounding an `It` are run from outermost to innermost when the `It` is executed.  The same is true for `AfterEach` block thoug they run from innermost to outermost.  Note: the `BeforeEach` and `AfterEach` blocks run for **each** `It` block.  This ensures a pristine state for each spec.

> In general, the only code within a container block should be an `It` block or a `BeforeEach`/`JustBeforeEach`/`AfterEach` block, or closure variable declarations.  It is generally a mistake to make an assertion in a container block.

> It is also a mistake to *initialize* a closure variable in a container block.  If one of your `It`s mutates that variable, subsequent `It`s will receive the mutated value.  This is a case of test pollution and can be hard to track down.  **Always initialize your variables in `BeforeEach` blocks.**

### Separating Creation and Configuration: `JustBeforeEach`

The above example illustrates a common antipattern in BDD-style testing.  Our top level `BeforeEach` creates a new book using valid JSON, but a lower level `Context` excercises the case where a book is created with *invalid* JSON.  This causes us to recreate and override the original book.  Thankfully, with Ginkgo's `JustBeforeEach` blocks, this code duplication is unnecessary.

`JustBeforeEach` blocks are guaranteed to be run *after* all the `BeforeEach` blocks have run and *just before* the `It` block has run.  We can use this fact to clean up the Book specs:

    var _ = Describe("Book", func() {
        var (
            book Book
            err error
            json string
        )

        BeforeEach(func() {
            json = `{
                "title":"Les Miserables",
                "author":"Victor Hugo",
                "pages":1488
            }`
        })

        JustBeforeEach(func() {
            book, err = NewBookFromJSON(json)
        })

        Describe("loading from JSON", func() {
            Context("when the JSON parses succesfully", func() {
                It("should populate the fields correctly", func() {
                    Expect(book.Title).To(Equal("Les Miserables"))
                    Expect(book.Author).To(Equal("Victor Hugo"))
                    Expect(book.Pages).To(Equal(1488))
                })

                It("should not error", func() {
                    Expect(err).NotTo(HaveOccured())
                })
            })

            Context("when the JSON fails to parse", func() {
                BeforeEach(func() {
                    json = `{
                        "title":"Les Miserables",
                        "author":"Victor Hugo",
                        "pages":1488oops
                    }`
                })

                It("should return the zero-value for the book", func() {
                    Expect(book).To(BeZero())
                })

                It("should error", func() {
                    Expect(err).To(HaveOccured())
                })
            })
        })
        
        Describe("Extracting the author's last name", func() {
            It("should correctly identify and return the last name", func() {
                Expect(book.AuthorLastName()).To(Equal("Hugo"))
            })
        })
    })

Now the actual book creation only occurs once for every `It`, and the failing JSON context can simply assign invalid json to the `json` variable in a `BeforeEach`.

Abstractly, `JustBeforeEach` allows you to decouple **creation** from **configuration**.  Creation occurs in the `JustBeforeEach` using configuration specified and modified by a chain of `BeforeEach`s.

> You can have multiple `JustBeforeEach`es at different levels of nesting.  Ginkgo will first run all the `BeforeEach`es from the outside in, then it will run the `JustBeforeEach`es from the outside in.  While powerful, this can lead to confusing test suites -- so use nested `JustBeforeEach`es judiciously.
>
> Some parting words: `JustBeforeEach` is a powerful tool that can be easily abused.  Use it well.

### Global Setup and Teardown: Before and After the Suite

Sometimes you want to run some set up code once before the entire test suite and some clean up code once after the entire test suite.  For example, perhaps you need to spin up and tear down an external database.

A convenient pattern for doing this is to use the bootstrap test file:

    package books_test

    import (
        . "github.com/onsi/ginkgo"
        . "github.com/onsi/gomega"

        "testing"
    )

    func TestBooks(t *testing.T) {
        RegisterFailHandler(Fail)

        //Code to set up your infrastructure
        RunSpecs(t, "Books Suite")
        //Code to tear down your infrastructure
    }

---

## The Spec Runner

This section will discuss some of the properties Ginkgo's spec runner.

### Pending Specs

You can mark an individual spec or container as Pending.  This will prevent the spec (or specs within the container) from running.  You do this by adding a `P` or an `X` in front of your `Describe`, `Context` and `It`:

    PDescribe("some behavior", func() { ... })
    PContext("some scenario", func() { ... })
    PIt("some assertion", func() { ... })

    XDescribe("some behavior", func() { ... })
    XContext("some scenario", func() { ... })
    XIt("some assertion", func() { ... })

> By default, Ginkgo will print out a description for each pending spec.  You can suppress this by setting the `--noisyPendings=false` flag.

> By default, Ginkgo will not fail a suite for having pending specs.  You can pass the `--failOnPending` flag to reverse this behavior.

### Focused Specs

It is often convenient, when developing to be able to run a subset of specs.  Ginkgo has two mechanisms for allowing you to focus specs:

1. You can focus individual specs or whole containers of specs *programatically* by adding an `F` in front of your `Describe`, `Context`, and `It`:

        FDescribe("some behavior", func() { ... })
        FContext("some scenario", func() { ... })
        FIt("some assertion", func() { ... })

    doing so instructs Ginkgo to only run those specs.  To run all specs, you'll need to go back and remove all the `F`s.

2. You can pass in a regular expression with the `--focus=REGEXP` flag.  Ginkgo will only run specs that match this regular expression.

> The programatic approach and the `--focus=REGEXP` approach are mutually exclusive.  Using the command line flag will override the programmatic focus.

### Spec Permutation

By default, Ginkgo will randomize the order in which your specs are run.  This can help suss out test pollution early on in a suite's development.

Ginkgo's default behavior is to only permute the order of top-level containers -- the specs *within* those containers continue to run in the order in which they are specified in the test file.  This is helpful when developing specs as it mitigates the coginitive overload of having specs continuously change the order in which they run.

To randomize *all* specs in a suite, you can pass the `--randomizeAllSpecs` flag.  This is useful on CI and can greatly help fight the scourge of test pollution.

Ginkgo uses the current time to seed the randomization.  It prints out the seed near the beginning of the test output.  If you notice test intermittent test failures that you think may be due to test pollution, you can use the seed from a failing suite to exactly reproduce the spec order for that suite.  To do this pass the `--seed=SEED` flag.

### Parallel Specs

Ginkgo has support for running specs in parallel.  It does this by spawning separate `go test` processes and dividing the specs evenly among these processes.  This is important for a BDD test framework, as the shared context of the closures does not parallelize well in-process.

To run a Ginkgo suite in parallel you must use the `ginkgo` CLI.  To run N processes in parallel invoke:

    ginkgo -nodes=N

When running in parallel mode the test runner will not present any output until all the nodes have completed running.

If your tests spin up or connect to external processes you'll need to make sure that those connections are safe in a parallel context.  One way to ensure this would be, for example, to spin up a separate instance of an external resource for each Ginkgo process.  For example, let's say your tests spin up and hit a local web server.  You could bring up a different server bound to a different port for each of your parallel processes:

    package books_test

    import (
        . "github.com/onsi/ginkgo"
        . "github.com/onsi/gomega"
        "github.com/onsi/ginkgo/config"

        "testing"
    )

    func TestBooks(t *testing.T) {
        RegisterFailHandler(Fail)

        port := 4000 + config.GinkgoConfig.ParallelNode
        startServer(port)

        RunSpecs(t, "Books Suite")
    }

The `github.com/onsi/ginkgo/config` package provides your suite with access to the command line configuration parameters passed into Ginkgo.  The `config.GinkgoConfig.ParallelNode` parameter is the index for the current node (starts with `1`, goes up to `N`).  Similarly `config.GinkgoConfig.ParallelTotal` is the total number of nodes running in parallel.

---

## Asynchronous Tests

Go does concurrency well.  Ginkgo provides support for testing asynchronicity effectively.

Consider this example:

    It("should post to the channel, eventually", func() {
        c := make(chan string, 0)

        go DoSomething(c)
        Expect(<-c).To(ContainSubstring("Done!"))
    })

This test will block until a response is received over the channel `c`.  A deadlock or timeout is a common failure mode for tests like this, a common pattern in such situations is to add a select statement at the bottom of the function and include a `<-time.After(X)` channel to specify a timeout.

Ginkgo has this pattern built in.  The `body` functions in all non-container blocks (`It`s, `BeforeEach`es, `AfterEach`es, `JustBeforeEach`es, and `Benchmark`s) can take an optional `done Done` argument:

    It("should post to the channel, eventually", func(done Done) {
        c := make(chan string, 0)

        go DoSomething(c)
        Expect(<-c).To(ContainSubstring("Done!"))
        close(done)
    }, 0.2)

`Done` is a `chan interface{}`.  When Ginkgo detects that the `done Done` argument has been requested it runs the `body` function as a goroutine, wrapping it with the necessary logic to apply a timeout assertion.  You must either close the `done` channel, or send something (anything) to it to tell Ginkgo that your test has ended.  If your test doesn't end after a timeout period, Ginkgo will fail the test and move on the next one.

The default timeout is 1 second. You can modify this timeout by passing a `float64` (in seconds) after the `body` function.  In this example we set hte timeout to 0.2 seconds.

> Gomega has additional support for making rich assertions on asynchronous code.  Make sure to check out how `Eventually` works in Gomega.

---

## The Ginkgo CLI

The Ginkgo CLI can be installed by running

    $ go install github.com/onsi/ginkgo/ginkgo

It offers a number of conveniences beyond what `go test` provides out of the box and is recommended, but not necessary.

### Running Tests

To run the suite in the current directory, simply run:

    $ ginkgo #or go test

There are a number of command line flags that can be passed to the `ginkgo` test runner.  Additionally, most of these can also be passed to `go test`, though you will need to prepend `ginkgo.` to each flag when using `go test`:

- `--seed=SEED`

    The random seed to use when permuting the specs.

- `--randomizeAllSpecs`

    If present, all specs will be permuted.  By default Ginkgo will only permute the order of the top level containers.

- `--skipMeasurements`

    If present, Ginkgo will skip any `Measure` specs you've defined.

- `--failOnPending`

    If present, Ginkgo will mark a test suite as failed if it has any pending specs.

- `--focus=REGEXP`
    
    If provided, Ginkgo will only run specs with descriptions that match the regular expression REGEXP.

- `--noColor`
    
    If provided, Ginkgo's default reporter will not print out in color.

- `--slowSpecThreshold=TIME_IN_SECONDS`

    By default, Ginkgo's default reporter will flag tests that take longer than 5 seconds to run -- this does not fail the suite, it simply notifies you of slow running specs.  You can change this threshold using this flag.

- `--noisyPendings=false`

    By default, Ginkgo's defautlt reporter will provide detailed output for pending specs.  You can set --noisyPendings=false to supress this behavior.

- `--v`

    If present, Ginkgo's default reporter will print out the text and location for each spec before running it.

Additional flags supported by the `ginkgo` command:

- `--nodes=NODE_TOTAL`

    Use this to parallelize the suite across NODE_TOTAL processes.

- `-r`
    
    Set `-r` to have the `ginkgo` CLI recursively run all test suites under the current directory.  Useful for running all the tests across all your packages.

- `-i`
    
    Set `-i` to have the `ginkgo` CLI invoke the mysterious `go test -i` before running your tests.

- `-race`
    
    Set `-race` to have the `ginkgo` CLI run your tests with the race detector on.

- `-cover`

    Set `-cover` to have the `ginkgo` CLI run your tests with coverage analysis turned on (a Golang 1.2+ feature).  Ginkgo will generate coverage profiles under the current directory named `PACKAGE.coverprofile` for each set of package tests that is run.


Flags for `go test` only:

- `--ginkgo.parallel.node=NODE_INDEX`
    
    For parallel tests, this specifies the node index for this test run.  `parallel.node` and `parallel.total` are used in conjunction to select a subset of tests to run.  You generally don't need to set this, instead pass the `--nodes=NODE_TOTAL` flag to the `ginkgo` CLI.

- `--ginkgo.parallel.total=NODE_TOTAL`
    
    For parallel tests, this specifies the total number of nodes for this test run.  `parallel.node` and `parallel.total` are used in conjunction to select a subset of tests to run.  You generally don't need to set this, instead pass the `--nodes=NODE_TOTAL` flag to the `ginkgo` CLI.

### Generators

- To bootstrap a Ginkgo test suite for the package in the current directory, run:

        $ ginkgo bootstrap

    This will generate a file named `PACKAGE_suite_test.go` where PACKAGE is the name of the current directory.

- To add a test file to a package, run:

        $ ginkgo generate <SUBJECT>

    This will generate a file named `SUBJECT_test.go`.  If you don't specify SUBJECT, it will generate a file named `PACKAGE_test.go` where PACKAGE is the name of the current directory.

> Note that you don't have to use either of these generators.  They're just convenient helpers to get you up and running quickly.

### Other Subcommands

- For help:

        $ ginkgo help

- To get the current version of Ginkgo:

        $ ginkgo version

---

## Benchmark Tests

Ginkgo allows you to measure the performance of your code using `Measure` blocks.   `Measure` blocks can go wherever an `It` block can go -- each `Measure` generates a new spec.  The closure function passed to `Measure` must take a `Benchmarker` argument.  The `Benchmarker` is used to measure runtimes and record arbitrary numerical values.  You must also pass `Measure` an integer after your closure function, this represents the number of samples of your code `Measure` will perform.

For example:

    Measure("it should do something hard efficiently", func(b Benchmarker) {
        runtime := b.Time("runtime", func() {
            output := SomethingHard()
            Expect(output).To(Equal(17))            
        })

        Ω(runtime.Seconds()).Should(BeNumerically("<", 0.2), "SomethingHard() shouldn't take too long.")

        b.RecordValue("disk usage (in MB)", HowMuchDiskSpaceDidYouUse())
    }, 10)

will run the closure function 10 times, aggregating data for "runtime" and "disk usage".  Ginkgo's reporter will then print out a summary of each of these metrics containing some simple statistics:

    • [MEASUREMENT]
    Suite
        it should do something hard efficiently

        Ran 10 samples:
        runtime:
          Fastest Time: 0.01s
          Slowest Time: 0.08s
          Average Time: 0.05s ± 0.02s
        
        disk usage (in MB):
          Smallest: 3.0
           Largest: 5.2
           Average: 3.9 ± 0.4

With `Measure` you can write expressive, exploratory, specs to measure the performance of various parts of your code (or external components, if you use Ginkgo to write integration tests).  As you collect your data, you can leave the `Measure` specs in place to monitor performance and fail the suite should components start growing slow and bloated.

`Measure`s can live alongside `It`s within a test suite.  If you want to run just the `It`s you can pass the `--skipMeasurements` flag to `ginkgo`.

> You can also mark `Measure`s as pending with `PMeasure` and `XMeasure` or focus them with `FMeasure`.

### Measuring Time

The `Benchmarker` passed into your closure function provides the

    Time(name string, body func(), info ...Interface{}) time.Duration

 method.  `Time` runs the passed in `body` function and records, and returns, its runtime.  The resulting measurements for each sample are aggregated and some simple statistics are computed.  These stats appear in the spec output under the `name` you pass in.  Note that `name` must be unique within the scope of the `Measure` node.

 You can also pass arbitrary information via the optional `info` argument.  This will be passed along to the reporter along with the agreggated runtimes that `Time` measures.  The default reporter presents a string representation of `info`, but you can write a custom reporter to perform something more structured.  For example, you might run several measurements of the same code, but vary some parameter between runs.  You could encode the value of that parameter in `info`, and then have a custom reporter that uses `info` and the statistics provided by Ginkgo to generate a CSV file - or perhaps even a plot.

 If you want to assert that `body` ran within some threshold time, you can make an assertion against `Time`'s return value.

### Recording Arbitrary Values

The `Benchmarker` also provides the

    RecordValue(name string, value float64, info ...Interface{})

method.  `RecordValue` allows you to record arbitrary numerical data.  These results are aggregated and some simple statistics are computed.  These stats appear in the spec output under the `name` you pass in.  Note that `name` must be unique within the scope of the `Measure` node.

The optional `info` parameter can be used to pass structured data to a custom reporter.  See [Measuring Time](#measuring_time) above for more details.

---

## Shared Example Patterns

Ginkgo doesn't have any have any explicit support for Shared Examples (also known as Shared Behaviors) but there are a few patterns that you can use to reuse tests across your suite.

### Locally-scoped Shared Behaviors

It is often the case that a number of `Context`s within a suite describe slightly different set ups that result in the roughly the same behavior.  Rather than repeat the `It`s for across these `Context`s you can pull out a function that lives within the same closure that `Context`s live in, that defines these shared `It`s.  For example:

    Describe("my api client", func() {
        var client APIClient
        var fakeServer FakeServer
        var response chan APIResponse

        BeforeEach(func() {
            response = make(chan APIResponse, 1)
            fakeServer = NewFakeServer()
            client = NewAPIClient(fakeServer)
            client.Get("/some/endpoint", response)
        })

        Describe("failure modes", func() {
            Context("when the server does not return a 200", func() {
                BeforeEach(func() {
                    fakeServer.Respond(404)
                })

                AssertFailedBehavior()
            })

            Context("when the server returns unparseable JSON", func() {
                BeforeEach(func() {
                    fakeServer.Succeed("{I'm not JSON!")
                })

                AssertFailedBehavior()                
            })

            Context("when the request errors", func() {
                BeforeEach(func() {
                    fakeServer.Error(errors.New("oops!"))
                })

                AssertFailedBehavior()
            })

            AssertFailedBehavior := func() {                
                It("should not include JSON in the response", func() {
                    Ω((<-response).JSON).Should(BeZero())
                })

                It("should not report success", func() {
                    Ω((<-response).Success).Should(BeFalse())
                })
            }
        })
    })

Note that the `AssertFailedBehavior` function is called within the body of the `Context` container block.  The `It`s defined by this function get added to the enclosing container.  Since the function shares the same closure scope we don't need to pass the `response` channel in.

### Global Shared Behaviors

The pattern outlined above works well when the shared behavior is intended to be used within a fixed scope.  If you want to build shared behavior that can be used across different test files (or even different packages) you'll need to tweak the pattern to make it possible to pass inputs in.  We can extend the example outlined above to illustrate how this might work:

    package sharedbehaviors

    type FailedResponseBehaviorInputs struct {
        response chan APIResponse
    }

    func SharedFailedResponseBehavior(inputs *FailedResponseBehaviorInputs) {                
        It("should not include JSON in the response", func() {
            Ω((<-(inputs.response)).JSON).Should(BeZero())
        })

        It("should not report success", func() {
            Ω((<-(inputs.response)).Success).Should(BeFalse())
        })
    }

Users of the shared behavior must generate and populate a `FailedResponseBehaviorInputs` and pass it in to `SharedFailedResponseBehavior()`.  Why do things this way?  Two reasons:

1. Having a stuct to encapsulate the input variables (like `FailedResponseBehaviorInputs`) allows you to clearly stipulate the contract between the the specs and the shared behavior.  The shared behavior *needs* these inputs in order to function correctly.

2. More importantly, inputs like the `response` channel are generally created and/or set in `BeforeEach` blocks.  However the shared behavior function `SharedFailedResponseBehavior` must be called within a container block and will not have access to any variabels specified in a `BeforeEach` as the `BeforeEach` hasn't run yet.  To get around this, we instantiate a `FailedResponseBehaviorInputs` and pass a pointer to it to the `SharedFailedResponseBehavior` -- in the `BeforeEach` we manipulate the fields of the `FailedResponseBehaviorInputs`, ensuring that their values get communicated to the `It`s generated by the shared behavior.

Here's what the calling test might look like:

    Describe("my api client", func() {
        var client APIClient
        var fakeServer FakeServer
        var response chan APIResponse
        sharedBehaviorInputs := FailedResponseBehaviorInputs{}

        BeforeEach(func() {
            sharedBehaviorInputs.response = make(chan APIResponse, 1)
            fakeServer = NewFakeServer()
            client = NewAPIClient(fakeServer)
            client.Get("/some/endpoint", sharedBehaviorInputs.response)
        })

        Describe("failure modes", func() {
            Context("when the server does not return a 200", func() {
                BeforeEach(func() {
                    fakeServer.Respond(404)
                })

                SharedFailedResponseBehavior(&sharedBehaviorInputs)
            })

            Context("when the server returns unparseable JSON", func() {
                BeforeEach(func() {
                    fakeServer.Succeed("{I'm not JSON!")
                })

                SharedFailedResponseBehavior(&sharedBehaviorInputs)
            })

            Context("when the request errors", func() {
                BeforeEach(func() {
                    fakeServer.Error(errors.New("oops!"))
                })

                SharedFailedResponseBehavior(&sharedBehaviorInputs)
            })
        })
    })

---

## Ginkgo and Gomega on TravisCI

Here's a sample `.travis.yml` for golang versions under 1.2:

    language: go
    go:
      - 1.1.2
      - tip

    install:
      - go get -v ./...
      - go get -v github.com/onsi/ginkgo
      - go get -v github.com/onsi/gomega
      - go install -v github.com/onsi/ginkgo/ginkgo

    script: PATH=$HOME/gopath/bin:$PATH ginkgo -r  -i --randomizeAllSpecs --failOnPending --skipMeasurements

Notice the manual steps to get ginkgo and gomega.  This is because version of go before 1.2 do not fetch test-only dependencies

For golang > 1.2:

    language: go
    go:
      - 1.1.2
      - tip

    install:
      - go get -v -t ./...
      - go get -v github.com/onsi/ginkgo
      - go get -v github.com/onsi/gomega
      - go install -v github.com/onsi/ginkgo/ginkgo

    script: PATH=$HOME/gopath/bin:$PATH ginkgo -r  -i --randomizeAllSpecs --failOnPending --skipMeasurements --cover

In both of these examples we're using the `ginkgo` CLI to recursively run all the tests in our package, we're also passing in a number of [flags](#running_tests) that are particularly pertinent for a CI environment.  You can, of course, use `go test` instead, in which case you can skip the `go install -v github.com/onsi/ginkgo/ginkgo` command.

---

## Writing Custom Reporters

While Ginkgo's default reporter offers a comprehensive set of features, Ginkgo makes it easy to write and run multiple custom reporters at once.  There are many usecases for this - you might implement a custom reporter to support a special output format for your CI setup, or you might implement a custom reporter to [aggregate data](#measuring_time) from Ginkgo's `Measure` nodes and produce HTML or CSV reports (or even plots!)

In Ginkgo a reporter must satisfy the `Reporter` interface:

    type Reporter interface {
        SpecSuiteWillBegin(config config.GinkgoConfigType, summary *SuiteSummary)
        ExampleWillRun(exampleSummary *ExampleSummary)
        ExampleDidComplete(exampleSummary *ExampleSummary)
        SpecSuiteDidEnd(summary *SuiteSummary)
    }

The method names should be self-explanatory.  Be sure to dig into the `SuiteSummary` and `ExampleSummary` objects to get a sense of what data is available to your reporter.  If you're writing a custom reporter to ingest benchmarking data generated by `Measure` nodes you'll want to look at the `ExampleMeasurement` struct that is provided by `ExampleSummary.Measurements`.

Once you've created your custom reporter you may pass an instance of it to Ginkgo by replacing the `RunSpecs` command in your test suite bootstrap with either: 

    RunSpecsWithDefaultAndCustomReporters(t *testing.T, description string, reporters []Reporter)

or

    RunSpecsWithCustomReporters(t *testing.T, description string, reporters []Reporter)

`RunSpecsWithDefaultAndCustomReporters` will run your custom reporters alongside Ginkgo's default reporter.  `RunSpecsWithCustomReporters` will only run the custom reporters you pass in.

---

## Using Other Matcher Libraries

Ginkgo provides a single (global) function to signify spec failure:

    Fail(message string, callerSkip ...int)

`Fail` takes a failure message to present to the user.  In order to determine the best line of code to report to the user alongsid the failure, `Fail` takes an optional `callerSkip` integer that is used to index into the call stack.  If set to `0`, the file and line number of the *caller* of `Fail` will be presented to the user, if set to `1` the file and line number of the *caller of the caller* of `Fail` will be presented to the user, etc.

Gomega uses `Fail` to communicate failures to Ginkgo.  If you want to use a different matcher library you'll need to figure out how to get it to communicate failures to Ginkgo using the `Fail` function.


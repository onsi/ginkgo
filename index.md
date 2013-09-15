---
layout: default
title: Ginkgo
sidebar: Ginkgo
---
# Ginkgo: a BDD Testing Framework for Golang

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

    package books

    import (
        . "github.com/onsi/ginkgo"
        . "github.com/onsi/gomega"

        "testing"
    )

    func TestBootstrap(t *testing.T) {
        RegisterFailHandler(Fail)
        RunSpecs(t, "Books Suite")
    }

Let's break this down:

- `TestBootstrap` is a `testing` test.  The Go test runner will run this function when you run `go test`.
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

    package books

    import (
        . "github.com/onsi/ginkgo"
        . "github.com/onsi/gomega"
    )

    var _ = Describe("Book", func() {

    })

Let's break this down:

- We import the `ginkgo` and `gomega` packages into the global name space.  While incredibly convenient, this is not - strictly speaking - necessary.
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

    package books

    import (
        . "github.com/onsi/ginkgo"
        . "github.com/onsi/gomega"

        "testing"
    )

    func TestBootstrap(t *testing.T) {
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

If your tests spin up or connect to external processes you'll need to make sure that those connections are safe in a parallel context.  One way to ensure this would be, for example, to spin up a separate instance of an eternal resource for each Ginkgo process.  For example, let's say your tests spin up and hit a local web server.  You could bring up a different server bound to a different port for each of your parallel processes:

    package books

    import (
        . "github.com/onsi/ginkgo"
        . "github.com/onsi/gomega"
        "github.com/onsi/ginkgo/config"

        "testing"
    )

    func TestBootstrap(t *testing.T) {
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

- `--skipBenchmarks`

    If present, Ginkgo will skip any `Benchmark` specs you've defined.

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

Additional flags supported by the `ginkgo` command:

- `--nodes=NODE_TOTAL`

    Use this to parallelize the suite across NODE_TOTAL processes.

- `-r`
    
    Set `-r` to have the `ginkgo` CLI recursively run all test suites under the current directory.  Useful for running all the tests across all your packages.

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

Ginkgo allows you to measure the performance of your code using `Benchmark` blocks.   `Benchmark` blocks can go wherever an `It` block can go -- each `Benchmark` generates a new spec:

    Benchmark("it should do something hard efficiently", func() {
        output := SomethingHard()
        Expect(output).To(Equal(17))
    }, 10, 0.1)

`Benchmark` takes two additional arguments after the `body` function.  The first is `N`, the number of samples to perform.  The second is the maximum time that a sample can take (in seconds).  Ginkgo will run the `body` function `N` times, timing each sample.  It will then present the fastest time, the slowest time, and the average time along with the standard deviation.  If any of the samples takes longer than the maximum time, Ginkgo will mark the spec as failed.  In addition, if any of the individual samples happens to fail then `Benchmark` will abort the sampling and fail the spec.

In this way you can write expressive, exploratory, specs to measure the performance of various parts of your code (or external components, if you use Ginkgo to write integration tests).  As you collect your data, you can leave the `Benchmark` specs in place to monitor performance and fail the suite should components start getting slow and bloated.

`Benchmark`s can live alongside `It`s within a test suite.  If you want to run just the `It`s you can pass the `--skipBenchmarks` flag to `ginkgo`.

> `Benchmark`s also support the async testing mode that `It`s support.  Just pass a `done Done` argument to the `body` function.  You can specify the async timeout with an additional, third, numerical argument.

> You can also mark `Benchmark`s as pending with `PBenchmark` and `XBenchmark` or focus them with `FBenchmark`.

> The combination of `Benchmark` and asyncronous testing support makes Ginkgo an ideal testing framework for black-box integration testing components "from the outside".

---

## Shared Example Patterns

Coming Soon!

---

## Ginkgo and Gomega on TravisCI

Coming Soon!

---

## Writing Custom Reporters

Coming Soon!

---

## Using Other Matcher Libraries

Coming Soon!
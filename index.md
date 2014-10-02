---
layout: default
title: Ginkgo
---
[Ginkgo](http://github.com/onsi/ginkgo) is a BDD-style Golang testing framework built to help you efficiently write expressive and comprehensive tests.  It is best paired with the [Gomega](http://github.com/onsi/gomega) matcher library but is designed to be matcher-agnostic.  

These docs are written assuming you'll be using Gomega with Ginkgo.  They also assume you know your way around Go and have a good mental model for how Go organizes packages under `$GOPATH`.

---

##Getting Ginkgo

Just `go get` it:

    $ go get github.com/onsi/ginkgo/ginkgo
    $ go get github.com/onsi/gomega

this fetches ginkgo and installs the `ginkgo` executable under `$GOPATH/bin` -- you'll want that on your `$PATH`.

**Ginkgo is compatible with Golang v1.2+  Ubuntu users: apt-get currently installs v1.0 -- you'll need to upgrade!**

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
- We import the `ginkgo` and `gomega` packages into the test's top-level namespace by performing a dot-import.  If you'd rather not do this, check out the [Avoiding Dot Imports](#avoiding-dot-imports) section below.
- `TestBooks` is a `testing` test.  The Go test runner will run this function when you run `go test` or `ginkgo`.
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

- We import the `ginkgo` and `gomega` packages into the top-level namespace.  While incredibly convenient, this is not - strictly speaking - necessary.  If youd like to avoid this check out the [Avoiding Dot Imports](#avoiding-dot-imports) section below.
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

and Ginkgo will take care of the rest.

`Fail` (and therefore Gomega, since it uses fail) will record a failure for the current space *and* panic.  This allows Ginkgo to stop the current spec in its tracks - no subsequent assertions (or any code for that matter) will be called.  Ordinarily Ginkgo will rescue this panic itself and move on to the next test.

However, if your test launches a *goroutine* that calls `Fail` (or, equivalently, invokes a failing Gomega assertion), there's no way for Ginkgo to rescue the panic that `Fail` invokes.  This will cause the test suite to panic and no subsequent tests will run.  To get around this you must rescue the panic using `GinkgoRecover`.  Here's an example:

    It("panics in a goroutine", func(done Done) {
        go func() {
            defer GinkgoRecover()

            Ω(doSomething()).Should(BeTrue())

            close(done)
        }()
    })

Now, if `doSomething()` returns false, Gomega will call `Fail` which will panic but the `defer`red `GinkgoRecover()` will recover said panic and prevent the test suite from exploding.

More details about `Fail` and about using matcher libraries other than Gomega can be found in the [Using Other Matcher Libraries](#using-other-matcher-libraries) section.

###Logging Output

Ginkgo provides a globally available `io.Writer` called `GinkgoWriter` that you can write to.  `GinkgoWriter` aggregates input while a test is running and only dumps it to stdout if the test fails.  When running in verbose mode (`ginkgo -v` or `go test -ginkgo.v`) `GinkgoWriter` always immediately redirects its input to stdout.

When a Ginkgo test suite is interrupted (via `^C`) Ginkgo will emit any content written to the `GinkgoWriter`.  This makes it easier to debug stuck tests.  This is particularly useful when paired with `--progress` which instruct Ginkgo to emit notifications to the `GinkgoWriter` as it runs through your `BeforeEach`es, `It`s, `AfterEach`es, etc...

###IDE Support

Ginkgo works best from the command-line, and [`ginkgo watch`](#watching-for-changes) makes it easy to rerun tests on the command line whenever changes are detected.

There are a set of [completions](https://github.com/onsi/ginkgo-sublime-completions) available for [Sublime Text](http://www.sublimetext.com/): just use [Package Control](https://sublime.wbond.net/) to install `Ginkgo Completions`

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
                    Expect(err).NotTo(HaveOccurred())
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
                    Expect(err).To(HaveOccurred())
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

If you'd like to get information, at runtime about the current test, you can use `CallGinkgoTestDescription()` from within any `It` or `BeforeEach`/`AfterEach`/`JustBeforeEach` block.  The `GinkgoTestDescription` returned by this call has a variety of information about the currently running test including the filename, line number, text in the `It` block, and text in the surrounding container blocks.

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
                    Expect(err).NotTo(HaveOccurred())
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
                    Expect(err).To(HaveOccurred())
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

### Global Setup and Teardown: `BeforeSuite` and `AfterSuite`

Sometimes you want to run some set up code once before the entire test suite and some clean up code once after the entire test suite.  For example, perhaps you need to spin up and tear down an external database.

Ginkgo provides `BeforeSuite` and `AfterSuite` to accomplish this.  You typically define these at the top-level in the bootstrap file.  For example, say you need to set up an external database:

    package books_test

    import (
        . "github.com/onsi/ginkgo"
        . "github.com/onsi/gomega"

        "your/db"

        "testing"
    )

    var dbRunner *db.Runner
    var dbClient *db.Client

    func TestBooks(t *testing.T) {
        RegisterFailHandler(Fail)

        RunSpecs(t, "Books Suite")
    }

    var _ = BeforeSuite(func() {
        dbRunner = db.NewRunner()
        err := dbRunner.Start()
        Expect(err).NotTo(HaveOccurred())

        dbClient = db.NewClient()
        err = dbClient.Connect(dbRunner.Address())
        Expect(err).NotTo(HaveOccurred())
    })

    var _ = AfterSuite(func() {
        dbClient.Cleanup()
        dbRunner.Stop()
    })

The `BeforeSuite` function is run before any specs are run.  If a failure occurs in the `BeforeSuite` then none of the specs are run and the test suite ends.

The `AfterSuite` function is run after all the specs have run, regardless of whether any tests have failed.  Since the `AfterSuite` typically includes code to clean up persistent state ginkgo will *also* run `AfterSuite` when you send the running test suite an interrupt signal (`^C`).  To abort the `AfterSuite` send another interrupt signal.

Both `BeforeSuite` and `AfterSuite` can be run asynchronously by passing a function that takes a `Done` parameter.

You are only allowed to define `BeforeSuite` and `AfterSuite` *once* in a test suite (you shouldn't need more than one!) 

Finally, when running in parallel, each parallel process will run `BeforeSuite` and `AfterSuite` functions.  [Look here](#parallel-specs) for more on running tests in parallel.

---

## The Spec Runner

### Pending Specs

You can mark an individual spec or container as Pending.  This will prevent the spec (or specs within the container) from running.  You do this by adding a `P` or an `X` in front of your `Describe`, `Context`, `It`, and `Measure`:

    PDescribe("some behavior", func() { ... })
    PContext("some scenario", func() { ... })
    PIt("some assertion")
    PMeasure("some measurement")

    XDescribe("some behavior", func() { ... })
    XContext("some scenario", func() { ... })
    XIt("some assertion")
    XMeasure("some measurement")

> You don't need to remove the `func() { ... }` when you mark an `It` or `Measure` as pending.  Ginkgo will happily ignore any arguments after the string.

> By default, Ginkgo will print out a description for each pending spec.  You can suppress this by setting the `--noisyPendings=false` flag.

> By default, Ginkgo will not fail a suite for having pending specs.  You can pass the `--failOnPending` flag to reverse this behavior.

### Focused Specs

It is often convenient, when developing to be able to run a subset of specs.  Ginkgo has two mechanisms for allowing you to focus specs:

1. You can focus individual specs or whole containers of specs *programatically* by adding an `F` in front of your `Describe`, `Context`, and `It`:

        FDescribe("some behavior", func() { ... })
        FContext("some scenario", func() { ... })
        FIt("some assertion", func() { ... })

    doing so instructs Ginkgo to only run those specs.  To run all specs, you'll need to go back and remove all the `F`s.

2. You can pass in a regular expression with the `--focus=REGEXP` and/or `--skip=REGEXP` flags.  Ginkgo will only run specs that match the focus regular expression and don't match the skip regular expression.

When Ginkgo detects that a passing test suite has a programmatically focused test it causes the suite to exit with a non-zero status code.  This is to help detect erroneously committed focused tests on CI systems.  When passed a command-line focus/skip flag Ginkgo exits with status code 0 - if you want to focus tests on your CI system you should explicitly pass in a -focus or -skip flag.

Nested programmatically focused specs follow a simple rule: if a leaf-node is marked focused, any of its ancestor nodes that are marked focus will be unfocused.  With this rule, sibling leaf nodes (regardless of relative-depth) that are focused will run regardless of the focus of a shared ancestor; and non-focused siblings will not run regardless of the focus of the shared ancestor or the relative depths of the siblings.  More simply:

    FDescribe("outer describe", func() {
        It("A", func() { ... })
        It("B", func() { ... })
    })

will run both `It`s but

    FDescribe("outer describe", func() {
        It("A", func() { ... })
        FIt("B", func() { ... })
    })

will only run `B`.  This behavior tends to map more closely to what the developer actually intends when iterating on a test suite.

> The programatic approach and the `--focus=REGEXP`/`--skip=REGEXP` approach are mutually exclusive.  Using the command line flags will override the programmatic focus.

> When using the command line flags you can specify one or both of `--focus` and `--skip`.  If both are specified the constraints will be `AND`ed together.

> You can unfocus programatically focused tests by running `ginkgo unfocus`.  This will strip the `F`s off of any `FDescribe`, `FContext`, and `FIt`s that your tests in the current directory may have.

> If you want to skip entire packages (when running `ginkgo` recursively with the `-r` flag) you can pass a comma-separated list  to `--skipPackage=PACKAGES,TO,SKIP`.  Any packages with *paths* that contain one of the entries in this comma separated list will be skipped.

### Spec Permutation

By default, Ginkgo will randomize the order in which your specs are run.  This can help suss out test pollution early on in a suite's development.

Ginkgo's default behavior is to only permute the order of top-level containers -- the specs *within* those containers continue to run in the order in which they are specified in the test file.  This is helpful when developing specs as it mitigates the coginitive overload of having specs continuously change the order in which they run.

To randomize *all* specs in a suite, you can pass the `--randomizeAllSpecs` flag.  This is useful on CI and can greatly help fight the scourge of test pollution.

Ginkgo uses the current time to seed the randomization.  It prints out the seed near the beginning of the test output.  If you notice test intermittent test failures that you think may be due to test pollution, you can use the seed from a failing suite to exactly reproduce the spec order for that suite.  To do this pass the `--seed=SEED` flag.

When running multiple spec suites Ginkgo defaults to running the suites in the order they would be listed on the file-system.  You can permute the suites by passing `ginkgo --randomizeSuites`

### Parallel Specs

Ginkgo has support for running specs in parallel.  It does this by spawning separate `go test` processes and dividing the specs evenly among these processes.  This is important for a BDD test framework, as the shared context of the closures does not parallelize well in-process.

To run a Ginkgo suite in parallel you *must* use the `ginkgo` CLI.  Simply pass in the `-p` flag:

    ginkgo -p

this will automatically detect the optimal number of test nodes to spawn (see the note below.

To specify the number of nodes to spawn, use `-nodes`:

    ginkgo -nodes=N

> You do not need to specify both `-p` and `-nodes`.  Setting `-nodes` to anything greater than 1 implies a parallelized test run.

> The number of nodes used with `-p` is `runtime.NumCPU()` if `runtime.NumCPU() <= 4`, otherwise it is `runtime.NumCPU() - 1` based on a rigorous science based heuristic best characterized as "my gut sense based on a few months of experience"

The test runner collates output from the running processes into one coherent output.  This is done, under the hood, via a client-server model: as each client suite completes a test, the test output and status is sent to the server which then prints to screen.  This collates the output of simultaneous test runners to one coherent (i.e. non-interleaved), aggregated, output.

It is sometimes necessary/preferable to view the output of the individual parallel test suites in real-time.  To do this you can set `-stream`:

    ginkgo -nodes=N -stream

When run with the `-stream` flag the test runner simply pipes the output from each individual node as it runs (it prepends each line of output with the node # that the output came from).  This results in less coherent output (lines from different nodes will be interleaved) but can be valuable when debugging flakey/hanging test suites.

> On windows, parallel tests default to `-stream` because Ginkgo can't capture logging to stdout/stderr (necessary for aggregation) on windows.

#### Managing External Processes in Parallel Test Suites

If your tests spin up or connect to external processes you'll need to make sure that those connections are safe in a parallel context.  One way to ensure this would be, for example, to spin up a separate instance of an external resource for each Ginkgo process.  For example, let's say your tests spin up and hit a database.  You could bring up a different database server bound to a different port for each of your parallel processes:

    package books_test

    import (
        . "github.com/onsi/ginkgo"
        . "github.com/onsi/gomega"
        "github.com/onsi/ginkgo/config"

        "your/db"

        "testing"
    )

    var dbRunner *db.Runner
    var dbClient *db.Client


    func TestBooks(t *testing.T) {
        RegisterFailHandler(Fail)

        RunSpecs(t, "Books Suite")
    }

    var _ = BeforeSuite(func() {
        port := 4000 + config.GinkgoConfig.ParallelNode

        dbRunner = db.NewRunner()
        err := dbRunner.Start(port)
        Expect(err).NotTo(HaveOccurred())

        dbClient = db.NewClient()
        err = dbClient.Connect(dbRunner.Address())
        Expect(err).NotTo(HaveOccurred())
    })

    var _ = AfterSuite(func() {
        dbClient.Cleanup()
        dbRunner.Stop()
    })


The `github.com/onsi/ginkgo/config` package provides your suite with access to the command line configuration parameters passed into Ginkgo.  The `config.GinkgoConfig.ParallelNode` parameter is the index for the current node (starts with `1`, goes up to `N`).  Similarly `config.GinkgoConfig.ParallelTotal` is the total number of nodes running in parallel.

#### Managing *Singleton* External Processes in Parallel Test Suites

When possible, you should make every effort to start up a new instance of an external resource for every parallel node.  This helps avoid test-pollution by strictly separating each parallel node.

Sometimes (rarely) this is not possible.  Perhaps, for reasons beyond your control, you can only start one instance of a service on your machine.  Ginkgo provides a workaround for this with `SynchronizedBeforeSuite` and `SynchronizedAfterSuite`.

The idea here is simple.  With `SynchronizedBeforeSuite` Ginkgo gives you a way to run some preliminary setup code on just one parallel node (Node 1) and other setup code on all nodes.  Ginkgo synchronizes these functions and guarantees that node 1 will complete its preliminary setup before the other nodes run their setup code.  Moreover, Ginkgo makes it possible for the preliminary setup code on the first node to pass information on to the setup code on the other nodes.

Here's what our earlier database example looks like using `SynchronizedBeforeSuite`:

    var _ = SynchronizedBeforeSuite(func() []byte {
        port := 4000 + config.GinkgoConfig.ParallelNode

        dbRunner = db.NewRunner()
        err := dbRunner.Start(port)
        Expect(err).NotTo(HaveOccurred())

        return []byte(dbRunner.Address())
    }, func(data []byte) {
        dbAddress := string(data)

        dbClient = db.NewClient()
        err = dbClient.Connect(dbAddress)
        Expect(err).NotTo(HaveOccurred())
    })

`SynchronizedBeforeSuite` must be passed two functions.  The first must return `[]byte` and the second must accept `[]byte`.  When running with multiple nodes the *first* function is only run on node 1.  When this function completes, all nodes (including node 1) proceed to run the *second* function and will receive the data returned by the first function.  In this example, we use this data-passing mechanism to forward the database's address (set up on node 1) to all nodes.

To clean up correctly, you should use `SynchronizedAfterSuite`.  Continuing our example:

    var _ = SynchronizedAfterSuite(func() {
        dbClient.Cleanup()
    }, func() {
        dbRunner.Stop()
    })

With `SynchronizedAfterSuite` the *first* function is run on *all* nodes (including node 1).  The *second* function is only run on node 1.  Moreover, the second function is only run when all other nodes have finished running.  This is important, since node 1 is responsible for setting up and tearing down the singleton resources it must wait for the other nodes to end before tearing down the resources they depend on.

Finally, all of these function can be passed an additional `Done` parameter to run asynchronously.  When running asynchronously, an optional timeout can be provided as a third parameter to `SynchronizedBeforeSuite` and `SynchronizedAfterSuite`.  The same timeout is applied to both functions.

> Note an important subtelty:  The `dbRunner` variable is *only* populated on Node 1.  No other node should attempt to touch the data in that variable (it will be nil on the other nodes).  The `dbClient` variable, which is populated in the second `SynchronizedBeforeSuite` function is, of course, available across all nodes.

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

To run the suites in other directories, simply run:

    $ ginkgo /path/to/package /path/to/other/package ...

To pass arguments/custom flags down to your test suite:

    $ ginkgo -- <PASS-THROUGHS>

Note: the `--` is important!  Only arguments following `--` will be passed to your suite.

Of course, ginkgo takes a number of flags.  These must be specified *before* you specify the packages to run.  Here's a summary of the call syntax:

    $ ginkgo <FLAGS> <PACKAGES> -- <PASS-THROUGHS>

Here are the flags that Ginkgo accepts:

**Controlling which test suites run:**
   
- `-r`
    
    Set `-r` to have the `ginkgo` CLI recursively run all test suites under the target directories.  Useful for running all the tests across all your packages.

- `-skipPackage=PACKAGES,TO,SKIP`

    When running with `-r` you can pass `-skipPackage` a comma-separated list of entries.  Any packages with *paths* that contain one of the entries in this comma separated list will be skipped.

**Running in parallel:**

- `-p`

    Set `-p` to parallelize the test suite across an auto-detected number of nodes.

- `--nodes=NODE_TOTAL`

    Use this to parallelize the suite across NODE_TOTAL processes.  You don't need to specify `-p` as well (though you can!)

- `-stream`

    By default, when parallelizing a suite, the test runner aggregates data from each parallel node and produces coherent output as the tests run.  Setting `stream` to `true` will, instead, stream output from all parallel nodes in real-time, prepending each log line with the node # that emitted it.  This leads to incoherent (interleaved) output, but is useful when debugging flakey/hanging test suites.

**Modifying output:**

- `--noColor`
    
    If provided, Ginkgo's default reporter will not print out in color.

- `--succinct`
    
    Succinct silences much of Ginkgo's more verbose output.  Test suites that succeed basically get printed out on just one line!  Succinct is turned off, by default, when running tests for one package.  It is turned on by default when Ginkgo runs multiple test packages.

- `--v`

    If present, Ginkgo's default reporter will print out the text and location for each spec before running it.  Also, the GinkgoWriter will flush its output to stdout in realtime.

- `--noisyPendings=false`

    By default, Ginkgo's default reporter will provide detailed output for pending specs.  You can set --noisyPendings=false to supress this behavior.

- `--trace`

    If present, Ginkgo will print out full stack traces for each failure, not just the line number at which the failure occurs.
    
- `--progress`

	If present, Ginkgo will emit the progress to the `GinkgoWriter` as it enters and runs each `BeforeEach`, `AfterEach`, `It`, etc... node.  This is useful for debugging stuck tests (e.g. where exactly is the test getting stuck?) and for making tests that emit many logs to the `GinkgoWriter` more readable (e.g. which logs were emitted in the `BeforeEach`?  Which were emitted by the `It`?).  Combine with `--v` to emit the `--progress` output to stdout.

**Controlling randomization:**

- `--seed=SEED`

    The random seed to use when permuting the specs.

- `--randomizeAllSpecs`

    If present, all specs will be permuted.  By default Ginkgo will only permute the order of the top level containers.

- `--randomizeSuites`

    If present and running multiple spec suites, the order in which the specs run will be randomized.

**Focusing and Skipping specs:**

- `--skipMeasurements`

    If present, Ginkgo will skip any `Measure` specs you've defined.

- `--focus=REGEXP`
    
    If provided, Ginkgo will only run specs with descriptions that match the regular expression REGEXP.

- `--skip=REGEXP`

    If provided, Ginkgo will only run specs with descriptions that do not match the regular expression REGEXP.

**Running the race detector and code coverage tools:**

- `-race`
    
    Set `-race` to have the `ginkgo` CLI run your tests with the race detector on.

- `-cover`

    Set `-cover` to have the `ginkgo` CLI run your tests with coverage analysis turned on (a Golang 1.2+ feature).  Ginkgo will generate coverage profiles under the current directory named `PACKAGE.coverprofile` for each set of package tests that is run.

** Build flags: **

- `-tags`
    
    Set `-tags` to pass in build tags down to the compilation step.

** Failure behavior: **

- `--failOnPending`

    If present, Ginkgo will mark a test suite as failed if it has any pending specs.

- `--failFast`

    If present, Ginkgo will stop the suite right after the first spec failure.  When set during parallel tests, a failure on one node does *not* stop other nodes from running - only the node that failed will stop running.

** Watch flags: **

- `--depth=DEPTH`

    When watching packages, Ginkgo also watches those package's dependencies for changes.  The default value for `--depth` is `1` meaning that only the immediate dependencies of a package are monitored.  You can adjust this up to monitor dependencies-of-dependencies, or set it to `0` to only monitor the package itself, not its dependencies.

** Miscellaneous: **

- `-dryRun`

    If present, Ginkgo will walk your test suite and report output *without* actually running your tests.  This is best paried with `-v` to preview which tests will run.  Ther ordering of the tests honors the randomization strategy specified by `--seed` and `--randomizeAllSpecs`.

- `-keepGoing`

    By default, when running multiple tests (with -r or a list of packages) Ginkgo will abort when a test fails.  To have Ginkgo run subsequent test suites after a failure you can set -keepGoing.

- `-untilItFails`
    
    If set to `true`, Ginkgo will keep running your tests until a failure occurs.  This can be useful to help suss out race conditions or flakey tests.  It's best to run this with `--randomizeAllSpecs` and --randomizeSuites` to permute test order between iterations.

- `-notify`

    Set `-notify` to receive desktop notifications when a test suite completes.  This is especially useful with the `watch` subcommand.  Currently `-notify` is only supported on OS X.  You'll also need to `brew install terminal-notifier` to receive notifications.

- `--slowSpecThreshold=TIME_IN_SECONDS`

    By default, Ginkgo's default reporter will flag tests that take longer than 5 seconds to run -- this does not fail the suite, it simply notifies you of slow running specs.  You can change this threshold using this flag.

### Watching For Changes

The Ginkgo CLI provides a `watch` subcommand that takes (almost) all the flags that the main `ginkgo` command takes.  With `ginkgo watch` ginkgo will monitor the package in the current directory and trigger tests when changes are detected.

You can also run `ginkgo watch -r` to monitor all packages recursively.

For each monitored packaged, Ginkgo will also monitor that package's dependencies and trigger the monitored package's test suite when a change in a dependency is detected.  By default, `ginkgo watch` monitors a package's immediate dependencies.  You can adjust this using the `-depth` flag.  Set `-depth` to `0` to disable monitoring dependencies and set `-depth` to something greater than `1` to monitor deeper down the dependency graph.

Passing the `-notify` flag on OS X will trigger desktop notifications when `ginkgo watch` triggers and completes a test run.

### Generators

- To bootstrap a Ginkgo test suite for the package in the current directory, run:

        $ ginkgo bootstrap

    This will generate a file named `PACKAGE_suite_test.go` where PACKAGE is the name of the current directory.

- To add a test file to a package, run:

        $ ginkgo generate <SUBJECT>

    This will generate a file named `SUBJECT_test.go`.  If you don't specify SUBJECT, it will generate a file named `PACKAGE_test.go` where PACKAGE is the name of the current directory.

By default, these generators will dot-import both Ginkgo and Gomega.  To avoid dot imports, you can pass `--nodot` to both subcommands.  This is discussed more fully in the [next section](#avoiding-dot-imports).

> Note that you don't have to use either of these generators.  They're just convenient helpers to get you up and running quickly.

###Avoiding Dot Imports

Ginkgo and Gomega provide a DSL and, by default, the `ginkgo bootstrap` and `ginkgo generate` commands import both packages into the top-level namespace using dot imports.

There are certain, rare, cases where you need to avoid this.  For example, your code may define methods with names that conflict with the methods defined in Ginkgo and/or Goemga.  In such cases you can either import your code into its own namespace (i.e. drop the `.` in front of your package import).  Or, you can drop the `.` in front of Ginkgo and/or Gomega.  The latter comes at the cost of constantly having to preface your `Describe`s and `It`s with `ginkgo.` and your `Expect`s and `ContainSubstring`s with `gomega.`.

There is a *third* option that the ginkgo CLI provides, however.  If you need to (or simply want to!) avoid dot imports you can:

    ginkgo bootstrap --nodot

and

    ginkgo generate --nodot <filename>

This will create a bootstrap file that *explicitly* imports all the exported identifiers in Ginkgo and Gomega into the top level namespace.  This happens at the bottom of your bootstrap file and generates code that looks something like:

    import (
        github.com/onsi/ginkgo
        ...
    )

    ...

    // Declarations for Ginkgo DSL
    var Describe = ginkgo.Describe
    var Context = ginkgo.Context
    var It = ginkgo.It
    // etc...

This allows you to write tests using `Describe`, `Context`, and `It` without dot imports and without the `ginkgo.` prefix.  Crucially, it also allows you to redefine any conflicting identifiers (or even cook up your own semantics!).  For example:

    var _ = ginkgo.Describe
    var When = ginkgo.Context
    var Then = ginkgo.It

This will avoid importing `Describe` and will rename `Context` to `When` and `It` to `Then`.

As new matchers are added to Gomega you may need to update the set of imports identifiers.  You can do this by entering the directory containing your bootstrap file and running:

    ginkgo nodot

this will update the imports, preserving any renames that you've provided.
 
###Converting Existing Tests

If you have an existing XUnit test suite that you'd like to convert to a Ginkgo suite, you can use the `ginkgo convert` command:

    ginkgo convert github.com/your/package

This will generate a Ginkgo bootstrap file and convert any `TestX...(t *testing.T)` XUnit style tsts into simply (flat) Ginkgo tests.  It also substitutes `GinkgoT()` for any references to `*testing.T` in your code.  `ginkgo convert` usually gets things right the first time round, but you may need to go in and tweak your tests after the fact.

Also: `ginkgo convert` will **overwrite** your existing test files, so make sure you don't have any uncommitted changes before trying `ginkgo convert`!

`ginkgo convert` is the brainchild of [Tim Jarratt](http://github.com/tjarratt)

### Other Subcommands

- To unfocus any programmatically focused tests in the current directory (and subdirectories):

        $ ginkgo unfocus

- For help:

        $ ginkgo help

    For help on a particular subcommand:

        $ ginkgo help <COMMAND>

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

The optional `info` parameter can be used to pass structured data to a custom reporter.  See [Measuring Time](#measuring-time) above for more details.

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

## Ginkgo and Continuous Integration

Ginkgo comes with a number of [flags](#running-tests) that you probably want to turn on when running in a Continuous Integration environment.  The following is recommended:

    ginkgo -r --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --progress

- `-r` will recursively find and run all spec suites in the current directory
- `--randomizeAllSpecs` and `--randomizeSuites` will shuffle both the order in which specs within a suite run, and the order in which different suites run.  This can be *great* for identifying test pollution.  You can always rerun a given ordering later by passing the `--seed` flag a matching seed.
- `--failOnPending` causes the test suite to fail if there are any pending tests (typically these should not be committed but should signify work in progress).
- `--cover` generates `.coverprofile`s and coverage statistics for each test suite.
- `--trace` prints out a full stack trace when failures occur.  This makes debugging based on CI logs easier.
- `--race` runs the tests with the race detector turned on.
- `--progress` emits test progress to the GinkgoWriter.  Makes identifying where failures occur a little easier.

It is *not* recommended that you run tests in parallel on CI with `-p`.  Many CI systems run on multi-core machines that report very many (e.g. 32 nodes).  Parallelizing on such a high scale typically yields *longer* test run times (particularly since your tests are probably running inside some sort of cpu-share limited container: you don't actually have free reign of all 32 cores).  To run tests in parallel on CI you're probably better off providing an explicit number of parallel nodes with `-nodes`.

### Sample .travis.yml

For Travis CI, you could use something like this:

    language: go
    go:
      - 1.1.2
      - tip

    install:
      - go get -v github.com/onsi/ginkgo/ginkgo
      - go get -v github.com/onsi/gomega
      - go get -v -t ./...
      - export PATH=$PATH:$HOME/gopath/bin

    script: ginkgo -r --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race

---

## Writing Custom Reporters

While Ginkgo's default reporter offers a comprehensive set of features, Ginkgo makes it easy to write and run multiple custom reporters at once.  There are many usecases for this - you might implement a custom reporter to support a special output format for your CI setup, or you might implement a custom reporter to [aggregate data](#measuring-time) from Ginkgo's `Measure` nodes and produce HTML or CSV reports (or even plots!)

In Ginkgo a reporter must satisfy the `Reporter` interface:

    type Reporter interface {
        SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary)
        BeforeSuiteDidRun(setupSummary *types.SetupSummary)
        SpecWillRun(specSummary *types.SpecSummary)
        SpecDidComplete(specSummary *types.SpecSummary)
        AfterSuiteDidRun(setupSummary *types.SetupSummary)
        SpecSuiteDidEnd(summary *types.SuiteSummary)
    }

The method names should be self-explanatory.  Be sure to dig into the `SuiteSummary` and `SpecSummary` objects to get a sense of what data is available to your reporter.  If you're writing a custom reporter to ingest benchmarking data generated by `Measure` nodes you'll want to look at the `ExampleMeasurement` struct that is provided by `ExampleSummary.Measurements`.

Once you've created your custom reporter you may pass an instance of it to Ginkgo by replacing the `RunSpecs` command in your test suite bootstrap with either: 

    RunSpecsWithDefaultAndCustomReporters(t *testing.T, description string, reporters []Reporter)

or

    RunSpecsWithCustomReporters(t *testing.T, description string, reporters []Reporter)

`RunSpecsWithDefaultAndCustomReporters` will run your custom reporters alongside Ginkgo's default reporter.  `RunSpecsWithCustomReporters` will only run the custom reporters you pass in.

If you wish to run your tests in parallel you should not use `RunSpecsWithCustomReporters` as the default reporter plays an important role in streaming test output to the ginkgo CLI.

---

## Third Party Integrations

### Using Other Matcher Libraries

Most matcher library accept the `*testing.T` object.  Unfortunately, since this is a concrete type is can be tricky to pass in an equivalent that will work with Ginkgo.

It is, typically, not difficult to replace `*testing.T` in such libraries with an interface that `*testing.T` satisfies.  For example [testify](https://github.com/stretchr/testify) accepts `t` via an interface.  In such cases you can pass `GinkgoT()`.  This generates an object that mimics `*testing.T` and communicates to Ginkgo directly.

For example, to get testify working:

    package foo_test

    import (
        . "github.com/onsi/ginkgo"

        "github.com/stretchr/testify/assert"
    )

    var _ = Describe(func("foo") {
        It("should testify to its correctness", func)({
            assert.Equal(GinkgoT(), foo{}.Name(), "foo")
        })
    })

> Note that passing the `*testing.T` from Ginkgo's bootstrap `Test...()` function will cause the suite to abort as soon as the first failure is encountered.  Don't do this.  You need to communicate failure to Ginkgo's single (global) `Fail` function

### Integrating with Gomock

Ginkgo does not provide a mocking/stubbing framework.  It's the author's opinion that mocks and stubs can be avoided completely by embracing dependency injection and always injecting Go interfaces.  Real dependencies are then injected in production code, and fake dependencies are injected under test.  Building and maintaining such fakes tends to be straightforward and can allow for clearer and more expressive tests than mocks.

With that said, it is relatively straightforward to use a mocking framework such as [Gomock](https://code.google.com/p/gomock/).  `GinkgoT()` implements Gomock's `TestReporter` interface.  Here's how you use it (for example):

    import (
      "code.google.com/p/gomock/gomock"

      . github.com/onsi/ginkgo
      . github.com/onsi/gomega
    )

    var _ = Describe("Consumer", func() {
        var (
            mockCtrl *gomock.Controller
            mockThing *mockthing.MockThing
            consumer *Consumer
        )

        BeforeEach(func() {
            mockCtrl = gomock.NewController(GinkgoT())
            mockThing = mockthing.NewMockThing(mockCtrl)
            consumer = NewConsumer(mockThing)
        })

        AfterEach(func() {
            mockCtrl.Finish()
        })

        It("should consume things", func() {
            mockThing.EXPECT().OmNom()
            consumer.Consume()
        })
    })

When using Gomock you may want to run `ginkgo` with the `-trace` flag to print out stack traces for failures which will help you trace down where, in your code, invalid calls occured.

### Generating JUnit XML Output

Ginkgo provides a [custom reporter](#writing-custom-reporters) for generating JUnit compatible XML output.  Here's a sample bootstrap file that instantiates a JUnit reporter and passes it to the test runner:

    package foo_test

    import (
        . "github.com/onsi/ginkgo"
        . "github.com/onsi/gomega"

        "github.com/onsi/ginkgo/reporters"
        "testing"
    )

    func TestFoo(t *testing.T) {
        RegisterFailHandler(Fail)
        junitReporter := reporters.NewJUnitReporter("junit.xml")
        RunSpecsWithDefaultAndCustomReporters(t, "Foo Suite", []Reporter{junitReporter})
    }    

This will generate a file name "junit.xml" in the directory containing your test.  This xml file is compatible with the latest version of the Jenkins JUnit plugin.

If you want to run your tests in parallel you'll need to make your JUnit xml filename a function of the parallel node number.  You can do this like so:

    junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("junit_%d.xml", config.GinkgoConfig.ParallelNode))

Note that you'll need to import `fmt` and `github.com/onsi/ginkgo/config` to get this to work.  This will generate an xml file for each parallel node.  The Jenkins JUnit plugin (for example) automatically aggregates data from across all these files.

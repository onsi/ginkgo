---
layout: default
title: Ginkgo
---
[Ginkgo](https://github.com/onsi/ginkgo) is a Go testing framework built to help you efficiently write expressive and comprehensive tests using [Behavior-Driven Development](https://en.wikipedia.org/wiki/Behavior-driven_development) ("BDD") style.  It is best paired with the [Gomega](https://github.com/onsi/gomega) matcher library but is designed to be matcher-agnostic.

These docs are written assuming you'll be using Gomega with Ginkgo.  They also assume you know your way around Go and have a good mental model for how Go organizes packages under `$GOPATH`.

---

## Support Policy

Ginkgo provides support for versions of Go that are noted by the [Go release policy](https://golang.org/doc/devel/release.html#policy) i.e. N and N-1 major versions.

---

## Getting Ginkgo

Just `go get` it:

    $ go get github.com/onsi/ginkgo/ginkgo
    $ go get github.com/onsi/gomega/...

This fetches ginkgo and installs the `ginkgo` executable under `$GOPATH/bin` -- you'll want that on your `$PATH`.

**Ginkgo is tested against Go v1.6 and newer**
To install Go, follow the [installation instructions](https://golang.org/doc/install)

The above commands also install the entire gomega library. If you want to fetch only the packages needed by your tests, import the packages you need and use `go get -t`. 

For example, import the gomega package in your test code:

    import "github.com/onsi/gomega"

Use `go get -t` to retrieve the packages referenced in your test code:

    $ cd /path/to/my/app
    $ go get -t ./...

---

## Getting Started: Writing Your First Test
Ginkgo hooks into Go's existing `testing` infrastructure.  This allows you to run a Ginkgo suite using `go test`.

> This also means that Ginkgo tests can live alongside traditional Go `testing` tests.  Both `go test` and `ginkgo` will run all the tests in your suite.

### Bootstrapping a Suite
To write Ginkgo tests for a package you must first bootstrap a Ginkgo test suite.  Say you have a package named `books`:

    $ cd path/to/books
    $ ginkgo bootstrap

This will generate a file named `books_suite_test.go` containing:

```go
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
```

Let's break this down:

- Go allows us to specify the `books_test` package alongside the `books` package.  Using `books_test` instead of `books` allows us to respect the encapsulation of the `books` package: your tests will need to import `books` and access it from the outside, like any other package.  This is preferred to reaching into the package and testing its internals and leads to more behavioral tests.  You can, of course, opt out of this -- just change `package books_test` to `package books`
- We import the `ginkgo` and `gomega` packages into the test's top-level namespace by performing a dot-import.  If you'd rather not do this, check out the [Avoiding Dot Imports](#avoiding-dot-imports) section below.
- `TestBooks` is a `testing` test.  The Go test runner will run this function when you run `go test` or `ginkgo`.
- `RegisterFailHandler(Fail)`: A Ginkgo test signals failure by calling Ginkgo's `Fail(description string)` function.  We pass this function to Gomega using `RegisterFailHandler`.  This is the sole connection point between Ginkgo and Gomega.
- `RunSpecs(t *testing.T, suiteDescription string)` tells Ginkgo to start the test suite.  Ginkgo will automatically fail the `testing.T` if any of your specs fail.  You should only ever call `RunSpecs` once.

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

### Adding Specs to a Suite
An empty test suite is not very interesting.  While you can start to add tests directly into `books_suite_test.go` you'll probably prefer to separate your tests into separate files (especially for packages with multiple files).  Let's add a test file for our `book.go` model:

    $ ginkgo generate book

This will generate a file named `book_test.go` containing:

```go
package books_test

import (
    "/path/to/books"
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

var _ = Describe("Book", func() {

})
```

Let's break this down:

- We import the `ginkgo` and `gomega` packages into the top-level namespace.  While convenient this is, of course, not necessary.
- Similarly, we import the `books` package since we are using the special `books_test` package to isolate our tests from our code.  For convenience we import the `books` package into the namespace.  You can opt out of either these decisions by editing the generated test file.
- We add a *top-level* describe container using Ginkgo's `Describe(text string, body func()) bool` function.  The `var _ = ...` trick allows us to evaluate the Describe at the top level without having to wrap it in a `func init() {}`

The function in the `Describe` will contain our specs.  Let's add a few now to test loading books from JSON:

```go
var _ = Describe("Book", func() {
    var (
        longBook  Book
        shortBook Book
    )

    BeforeEach(func() {
        longBook = Book{
            Title:  "Les Miserables",
            Author: "Victor Hugo",
            Pages:  2783,
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
```

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

### Marking Specs as Failed

While you typically want to use a matcher library, like [Gomega](https://github.com/onsi/gomega), to make assertions in your specs, Ginkgo provides a simple, global, `Fail` function that allows you to mark a spec as failed.  Just call:

```go
Fail("Failure reason")
```

and Ginkgo will take care of the rest.

`Fail` (and therefore Gomega, since it uses fail) will record a failure for the current space *and* panic.  This allows Ginkgo to stop the current spec in its tracks - no subsequent assertions (or any code for that matter) will be called.  Ordinarily Ginkgo will rescue this panic itself and move on to the next test.

However, if your test launches a *goroutine* that calls `Fail` (or, equivalently, invokes a failing Gomega assertion), there's no way for Ginkgo to rescue the panic that `Fail` invokes.  This will cause the test suite to panic and no subsequent tests will run.  To get around this you must rescue the panic using `GinkgoRecover`.  Here's an example:

```go
It("panics in a goroutine", func(done Done) {
    go func() {
        defer GinkgoRecover()

        Ω(doSomething()).Should(BeTrue())

        close(done)
    }()
})
```

Now, if `doSomething()` returns false, Gomega will call `Fail` which will panic but the `defer`red `GinkgoRecover()` will recover said panic and prevent the test suite from exploding.

More details about `Fail` and about using matcher libraries other than Gomega can be found in the [Using Other Matcher Libraries](#using-other-matcher-libraries) section.

### Logging Output

Ginkgo provides a globally available `io.Writer` called `GinkgoWriter` that you can write to.  `GinkgoWriter` aggregates input while a test is running and only dumps it to stdout if the test fails or is [interrupted](#interrupting-and-aborting-test-runs) (via `^C`).  When running in verbose mode (`ginkgo -v` or `go test -ginkgo.v`) `GinkgoWriter` always immediately redirects its input to stdout.

`GinkgoWriter` includes three convenience methods:

- `GinkgoWriter.Print(a ...interface{})` is equivalent to `fmt.Fprint(GinkgoWriter, a...)`
- `GinkgoWriter.Println(a ...interface{})` is equivalent to `fmt.Fprintln(GinkgoWriter, a...)`
- `GinkgoWriter.Printf(format string, a ...interface{})` is equivalent to `fmt.Fprintf(GinkgoWriter, format, a...)`

You can also attach additional `io.Writer`s for `GinkgoWriter` to tee to via `GinkgoWriter.TeeTo(writer)`.  Any data written to `GinkgoWriter` will immediately be sent to attached tee writers.  All attached Tee writers can be cleared wtih `GinkgoWriter.ClearTeeWriters()`.

> Note that data is **immediately** written to writers registered via `GinkgoWriter.TooTo(writer)` regardless of whether the test has succeeded or passed.

### IDE Support

Ginkgo works best from the command-line, and [`ginkgo watch`](#watching-for-changes) makes it easy to rerun tests on the command line whenever changes are detected.

There are a set of [completions](https://github.com/onsi/ginkgo-sublime-completions) available for [Sublime Text](https://www.sublimetext.com/) (just use [Package Control](https://sublime.wbond.net/) to install `Ginkgo Completions`) and for [VSCode](https://code.visualstudio.com/) (use the extensions installer and install vscode-ginkgo).

IDE authors can set the `GINKGO_EDITOR_INTEGRATION` environment variable to any non-empty value to enable coverage to be displayed for focused specs. By default, Ginkgo will fail with a non-zero exit code if specs are focused to ensure they do not pass in CI.

---

## Structuring Your Specs

Ginkgo makes it easy to write expressive specs that describe the behavior of your code in an organized manner.  You use `Describe` and `Context` containers to organize your `It` specs and you use `BeforeEach` and `AfterEach` to build up and tear down common set up amongst your tests.

### Individual Specs: `It`
You can add a single spec by placing an `It` block within a `Describe` or `Context` container block:

```go
var _ = Describe("Book", func() {
    It("can be loaded from JSON", func() {
        book := NewBookFromJSON(`{
            "title":"Les Miserables",
            "author":"Victor Hugo",
            "pages":2783
        }`)

        Expect(book.Title).To(Equal("Les Miserables"))
        Expect(book.Author).To(Equal("Victor Hugo"))
        Expect(book.Pages).To(Equal(2783))
    })
})
```

> `It`s may also be placed at the top-level though this is uncommon.

#### The `Specify` Alias

In order to ensure that your specs read naturally, the `Specify`, `PSpecify`, `XSpecify`, and `FSpecify` blocks are available as aliases to use in situations where the corresponding `It` alternatives do not seem to read as natural language. `Specify` blocks behave identically to `It` blocks and can be used wherever `It` blocks (and `PIt`, `XIt`, and `FIt` blocks) are used.

An example of a good substitution of `Specify` for `It` would be the following:

```go
Describe("The foobar service", func() {
  Context("when calling Foo()", func() {
    Context("when no ID is provided", func() {
      Specify("an ErrNoID error is returned", func() {
      })
    })
  })
})
```

### Extracting Common Setup: `BeforeEach`
You can remove duplication and share common setup across tests using `BeforeEach` blocks:

```go
var _ = Describe("Book", func() {
    var book Book

    BeforeEach(func() {
        book = NewBookFromJSON(`{
            "title":"Les Miserables",
            "author":"Victor Hugo",
            "pages":2783
        }`)
    })

    It("can be loaded from JSON", func() {
        Expect(book.Title).To(Equal("Les Miserables"))
        Expect(book.Author).To(Equal("Victor Hugo"))
        Expect(book.Pages).To(Equal(2783))
    })

    It("can extract the author's last name", func() {
        Expect(book.AuthorLastName()).To(Equal("Hugo"))
    })
})
```

The `BeforeEach` is run before each spec thereby ensuring that each spec has a pristine copy of the state.  Common state is shared using closure variables (`var book Book` in this case).  You can also perform clean up in `AfterEach` blocks.

It is also common to place assertions within `BeforeEach` and `AfterEach` blocks.  These assertions can, for example, assert that no errors occured while preparing the state for the spec.

### Organizing Specs With Containers: `Describe` and `Context`

Ginkgo allows you to expressively organize the specs in your suite using `Describe` and `Context` containers:

```go
var _ = Describe("Book", func() {
    var (
        book Book
        err error
    )

    BeforeEach(func() {
        book, err = NewBookFromJSON(`{
            "title":"Les Miserables",
            "author":"Victor Hugo",
            "pages":2783
        }`)
    })

    Describe("loading from JSON", func() {
        Context("when the JSON parses succesfully", func() {
            It("should populate the fields correctly", func() {
                Expect(book.Title).To(Equal("Les Miserables"))
                Expect(book.Author).To(Equal("Victor Hugo"))
                Expect(book.Pages).To(Equal(2783))
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
                    "pages":2783oops
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
```

You use `Describe` blocks to describe the individual behaviors of your code and `Context` blocks to exercise those behaviors under different circumstances.  In this example we `Describe` loading a book from JSON and specify two `Context`s: when the JSON parses succesfully and when the JSON fails to parse.  Semantic differences aside, the two container types have identical behavior.

When nesting `Describe`/`Context` blocks the `BeforeEach` blocks for all the container nodes surrounding an `It` are run from outermost to innermost when the `It` is executed.  The same is true for `AfterEach` block though they run from innermost to outermost.  Note: the `BeforeEach` and `AfterEach` blocks run for **each** `It` block.  This ensures a pristine state for each spec.

> In general, the only code within a container block should be an `It` block or a `BeforeEach`/`JustBeforeEach`/`JustAfterEach`/`AfterEach` block, or closure variable declarations.  It is generally a mistake to make an assertion in a container block.

> It is also a mistake to *initialize* a closure variable in a container block.  If one of your `It`s mutates that variable, subsequent `It`s will receive the mutated value.  This is a case of test pollution and can be hard to track down.  **Always initialize your variables in `BeforeEach` blocks.**

### Separating Creation and Configuration: `JustBeforeEach`

The above example illustrates a common antipattern in BDD-style testing.  Our top level `BeforeEach` creates a new book using valid JSON, but a lower level `Context` exercises the case where a book is created with *invalid* JSON.  This causes us to recreate and override the original book.  Thankfully, with Ginkgo's `JustBeforeEach` blocks, this code duplication is unnecessary.

`JustBeforeEach` blocks are guaranteed to be run *after* all the `BeforeEach` blocks have run and *just before* the `It` block has run.  We can use this fact to clean up the Book specs:

```go
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
            "pages":2783
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
                Expect(book.Pages).To(Equal(2783))
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
                    "pages":2783oops
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
```

Now the actual book creation only occurs once for every `It`, and the failing JSON context can simply assign invalid json to the `json` variable in a `BeforeEach`.

Abstractly, `JustBeforeEach` allows you to decouple **creation** from **configuration**.  Creation occurs in the `JustBeforeEach` using configuration specified and modified by a chain of `BeforeEach`s.

> You can have multiple `JustBeforeEach`es at different levels of nesting.  Ginkgo will first run all the `BeforeEach`es from the outside in, then it will run the `JustBeforeEach`es from the outside in.  While powerful, this can lead to confusing test suites -- so use nested `JustBeforeEach`es judiciously.
>
> Some parting words: `JustBeforeEach` is a powerful tool that can be easily abused.  Use it well.

### Separating Diagnostics Collection and Teardown: `JustAfterEach`

It is sometimes useful to have some code which is executed just after each `It` block, but **before** Teardown (which might destroy useful state) - for example to to perform diagnostic operations if the test failed.

We can use this in the example above to check if the test failed and if so output the actual book:

```go
    JustAfterEach(func() {
        if CurrentSpecReport().Failed() {
            fmt.Printf("Collecting diags just after failed test in %s\n", CurrentSpecReport().SpecText())
            fmt.Printf("Actual book was %v\n", book)
        }
    })
```

> You can have multiple `JustAfterEach`es at different levels of nesting.  Ginkgo will first run all the `JustAfterEach`es from the inside out, then it will run the `AfterEach`es from the inside out.  While powerful, this can lead to confusing test suites -- so use nested `JustAfterEach`es judiciously.
>
> Like `JustBeforeEach`, `JustAfterEach` is a powerful tool that can be easily abused.  Use it well.

### Global Setup and Teardown: `BeforeSuite` and `AfterSuite`

Sometimes you want to run some set up code once before the entire test suite and some clean up code once after the entire test suite.  For example, perhaps you need to spin up and tear down an external database.

Ginkgo provides `BeforeSuite` and `AfterSuite` to accomplish this.  You typically define these at the top-level in the bootstrap file.  For example, say you need to set up an external database:

```go
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
```

The `BeforeSuite` function is run before any specs are run.  If a failure occurs in the `BeforeSuite` then none of the specs are run and the test suite ends.

The `AfterSuite` function is run after all the specs have run, regardless of whether any tests have failed.

Both `BeforeSuite` and `AfterSuite` can be run asynchronously by passing a function that takes a `Done` parameter.

You are only allowed to define `BeforeSuite` and `AfterSuite` *once* in a test suite (you shouldn't need more than one!)

Finally, when running in parallel, each parallel process will run `BeforeSuite` and `AfterSuite` functions.  [Look here](#parallel-specs) for more on running tests in parallel.

### Documenting Complex `It`s: `By`

As a rule, you should try to keep your `It`s, `BeforeEach`es, etc. short and to the point.  Sometimes this is not possible, particularly when testing complex workflows in integration-style tests.  In these cases your test blocks begin to hide a narrative that is hard to glean by looking at code alone.  Ginkgo provides `by` to help in these situations.  Here's a hokey example:

```go
var _ = Describe("Browsing the library", func() {
    BeforeEach(func() {
        By("Fetching a token and logging in")

        authToken, err := authClient.GetToken("gopher", "literati")
        Exepect(err).NotTo(HaveOccurred())

        err := libraryClient.Login(authToken)
        Exepect(err).NotTo(HaveOccurred())
    })

    It("should be a pleasant experience", func() {
        By("Entering an aisle")

        aisle, err := libraryClient.EnterAisle()
        Expect(err).NotTo(HaveOccurred())

        By("Browsing for books")

        books, err := aisle.GetBooks()
        Expect(err).NotTo(HaveOccurred())
        Expect(books).To(HaveLen(7))

        By("Finding a particular book")

        book, err := books.FindByTitle("Les Miserables")
        Expect(err).NotTo(HaveOccurred())
        Expect(book.Title).To(Equal("Les Miserables"))

        By("Check the book out")

        err := libraryClient.CheckOut(book)
        Expect(err).NotTo(HaveOccurred())
        books, err := aisle.GetBooks()
        Expect(books).To(HaveLen(6))
        Expect(books).NotTo(ContainElement(book))
    })
})
```

The string passed to `By` is emitted via the [`GinkgoWriter`](#logging-output).  If a test succeeds you won't see any output beyond Ginkgo's green dot.  If a test fails, however, you will see each step printed out up to the step immediately preceding the failure.  Running with `ginkgo -v` always emits all steps.

`By` takes an optional function of type `func()`.  When passed such a function `By` will immediately call the function.  This allows you to organize your `It`s into groups of steps but is purely optional.

`By` also adds a `ReportEntry` to the running spec.  This `ReportEntry` will appear in the structured suite report generated by `--json-format`.  If passed a function, `By` will measure the runtime of the function and attach the resulting duration to the report as well.

### `Ordered` Containers

By default, Ginkgo treats indiviudal specs as standalone units that can be randomized and executed in parallel with each other.  This encourages clean, independent, parallelizable tests.

Sometimes, however, it is necessary to ensure that a set of tests run in order and are treated as a coherent unit.  Ginkgo supports this via `Ordered` containers.  Here's the complex example from above laid out as an `Ordered` set of tests:

```go
var _ = Describe("Browing the Library", Ordered, func() {
    var aisle Aisle
    var books Books
    var book Book
    var err error

    BeforeAll(func() {
        authToken, err := authClient.GetToken("gopher", "literati")
        Exepect(err).NotTo(HaveOccurred())

        err := libraryClient.Login(authToken)
        Exepect(err).NotTo(HaveOccurred())
    })

    It("should enter an aisle", func() {
        aisle, err = libraryClient.EnterAisle()
        Expect(aisle).NotTo(BeZero())
    })

    It("should browse for books", func() {
        books, err = aisle.GetBooks()
        Expect(books).To(HaveLen(7))
    })

    It("should find a particular book", func() {
        book, err = books.FindByTitle("Les Miserables")
        Expect(book.Title).To(Equal("Les Miserables"))
    })

    It("should check the book out", func() {
        err = libraryClient.CheckOut(book)
        Expect(err).NotTo(HaveOccurred())

        books, err = aisle.GetBooks()
        Expect(books).To(HaveLen(6))
        Expect(books).NotTo(ContainElement(book))
    })

    AfterEach(func() {
        Expect(err).NotTo(HaveOccurred())
    })

    AfterAll(func() {
        libraryClient.Reset()
    })
})
```

Ginkgo will guarantee that these specs will always run in the order they appear and that they will not be parallelized with respect to one another.  The `BeforeAll` setup node will run once before the first spec runs and the `AfterAll` setup node will run once after the last spec runs.  `BeforeAll` and `AfterAll` nodes are only supported within `Ordered` containers and cannot be nested within other containers, even if those containers appear in an `Ordered` container.  Note that the `AfterEach` runs after every spec.

Since the specs are guaranteed to share order and run on the same process they modify the shared closure variables and produce a single coherent narrative.  Because this is a common pattern, when a spec in an `Ordered` container fails all subsequent specs are skipped.

While you can nest additional containers within an `Ordered` container you cannot nest additional `Ordered` containers.  While `Ordered` containers are useful pragmattic tools they can also be abused - in general we recommend keeping your tests separable and parallelizable where possible.

### Cleaning Up After Tests

The various examples so far have illustrated how `AfterEach`, `AfterAll`, and `AfterSuite` can be used to perform cleanup operations after tests, ordered containers or tests, and test suites have run, repsectively.

While powerful, the `AfterX` class of nodes have a tendency to separate cleanup code from set up code.  For example:

```go
var oldDeployTarget
BeforeEach(func() {
    oldDeployTarget = os.Getenv("DEPLOY_TARGET")
    err := os.Setenv("DEPLOY_TARGET", "TEST")
    Expect(err).NotTo(HaveOccurred())
})

It(...)
It(...)
It(...)

AfterEach(func() {
    err := os.Setenv("DEPLOY_TARGET", oldDeployTarget)
    Expect(err).NotTo(HaveOccurred())
})
```

sets the `DEPLOY_TARGET` environment variable before each test, then resets its value after each test, using `oldDeployTarget` to hold onto the previous value.  As written the clean up code is separated from the set up code and a shared variable must be used to communicate between the two.

Ginkgo provides the `DeferCleanup()` function to solve for this usecase and bring test setup closer to test cleanup.  Here's what your example looks like with `DeferCleanup()`:

```go
BeforeEach(func() {
    oldDeployTarget := os.Getenv("DEPLOY_TARGET")
    err := os.Setenv("DEPLOY_TARGET", "TEST")
    Expect(err).NotTo(HaveOccurred())
    DeferCleanup(func() {
        err := os.Setenv("DEPLOY_TARGET", oldDeployTarget)
        Expect(err).NotTo(HaveOccurred())
    })
})

It(...)
It(...)
It(...)
```

You can think of `DeferCleanup` as generating a dynamic `AfterEach` node when it is invoked.  The callback passed to `DeferCleanup` is guaranteed to run _after_ the test has completed and has identical semantics to the `AfterEach` in the previous example.

In fact, `DeferCleanup`s behavior depends on the context in which it is called:

- When called in a `BeforeEach`, `JustBeforeEach`, `It`, `AfterEach`, and `JustAfterEach`, `DeferCleanup` will behave like an `AfterEach` and run after the test completes.
- When called in a `BeforeAll` or `AfterAll`, `DeferCleanup` will behave like an `AfterAll` and run after the last test in an `Ordered` container.
- When called in a `BeforeSuite`, `SynchronizedBeforeSuite`, `AfterSuite`, and `SynchronizedAfterSuite`, `DeferCleanup` will behave like an `AfterSuite` that runs after all the tests (and any registered `AfterSuite` nodes) have run.
- `DeferCleanup()` cannot be called in any reporting nodes (e.g. `ReportAfterEach`), within the body of a container node outside of a setup or subject node, or within another `DeferCleanup` node.

As shown above `DeferCleanup` can be passed a function that takes no arguments and returns no value.  You can also pass a function that returns a single value.  `DeferCleanup` interprets this value as an error and fails the test if the error is non-nil.  This allows us to rewrite our example as:

```go
BeforeEach(func() {
    oldDeployTarget := os.Getenv("DEPLOY_TARGET")
    err := os.Setenv("DEPLOY_TARGET", "TEST")
    Expect(err).NotTo(HaveOccurred())
    DeferCleanup(func() error {
        return os.Setenv("DEPLOY_TARGET", oldDeployTarget)
    })
})

It(...)
It(...)
It(...)
```

`DeferCleanup` has one more trick up it's sleeve.  You can also pass in a function that accepts arguments, then pass those arguments in directly to `DeferCleanup`.  These arguments will be captured and passed to the function when cleanup is invoked.  This allows us to rewrite our example once more as:

```go
BeforeEach(func() {
    DeferCleanup(os.Setenv, "DEPLOY_TARGET", os.Getenv("DEPLOY_TARGET"))
    err := os.Setenv("DEPLOY_TARGET", "TEST")
    Expect(err).NotTo(HaveOccurred())
})

It(...)
It(...)
It(...)
```

here `DeferCleanup` is capturing the original value of `DEPLOY_TARGET` as returned by `os.Getenv("DEPLOY_TARGET")` then passing it into `os.Setenv` when cleanup is triggered after each test and asserting that the error returned by `os.Setenv` is `nil`.

### Getting information about the running test

If you'd like to get information, at runtime about the current test, you can use `CurrentSpecReport()` from within any runnable block.  The `types.SpecReport` returned by this call has a variety of information about the currently running test and is documented [here](https://pkg.go.dev/github.com/onsi/ginkgo/types#SpecReport).

You can also use a `ReportAfterEach` node to get a final report on a spec.  [More details here](#capturing-report-information-about-each-spec-as-the-test-suite-runs).

### Table Driven Tests

Ginkgo provides an expressive DSL for writing table driven tests:

```go
package table_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

var _ = Describe("Math", func() {
    DescribeTable("the > inequality",
        func(x int, y int, expected bool) {
            Expect(x > y).To(Equal(expected))
        },
        Entry("x > y", 1, 0, true),
        Entry("x == y", 0, 0, false),
        Entry("x < y", 0, 1, false),
    )
})
```

Let's break this down `DescribeTable` takes a description, a function to run for each test case, and a set of table entries.

The function you pass in to `DescribeTable` can accept arbitrary arguments.  The parameters passed in to the individual `Entry` calls will be passed in to the function (type mismatches will result in a runtime panic).

The indiviudal `Entry` calls construct a `TableEntry` that is passed into `DescribeTable`.  A `TableEntry` consists of a description (the first call to `Entry`) and an arbitrary set of parameters to be passed into the function registered with `DescribeTable`.

It's important to understand the life-cycle of the table.  The Table DSL is a thin wrapper around Ginkgo's cire DSL.  `DescribeTable` generates a single Ginkgo `Describe`, within this `Describe` each `Entry` generates a Ginkgo `It`.  This all happens *before* the tests run (at testing tree construction time).  The result is that the table expands into a number of `It`s (one for each `Entry`) that are subject to all of Ginkgo's test-running semantics: `It`s can be randomized and parallelized across multiple nodes.

To be clear, the above test is *exactly* equivalent to:

```go
package table_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

var _ = Describe("Math", func() {
    Describe("the > inequality",
        It("x > y", func() {
            Expect(1 > 0).To(Equal(true))
        })

        It("x == y", func() {
            Expect(0 > 0).To(Equal(false))
        })

        It("x < y", func() {
            Expect(0 > 1).To(Equal(false))
        })
    )
})
```

You should be aware of the Ginkgo test lifecycle - particularly around [dynamically generating tests](#patterns-for-dynamically-generating-tests) - when using `DescribeTable`.

#### Focusing and Pending Tables and Entries

Entire tables can be focused or marked pending by simply swapping out `DescribeTable` with `FDescribeTable` (to focus) or `PDescribeTable` (to mark pending).  Similarly, individual entries can be focused/pended out with `FEntry` and `PEntry`.  This is particularly useful when debugging tests.

#### Generating Entry Descriptions

There are a number of mechanisms for generating Entry descriptions.  As the example above shows, entries can provide their own descriptions.  There is also support for table-level Entry descriptions.  Entries can opt into having their names auto-generated by passing in `nil` for the Entry description:

```go
var _ = Describe("Math", func() {
    DescribeTable("addition",
        func(a, b, c int) {
            Expect(a+b).To(Equal(c))
        },
        Entry(nil, 1, 2, 3),
        Entry(nil, -1, 2, 1),
    )
})
```

This will generate entries named `Entry: 1, 2, 3` and `Entry: -1, 2, 1`.

You can customize the table-level Entry description by providing a function that returns a string:

```go
var _ = Describe("Math", func() {
    DescribeTable("addition",
        func(a, b, c int) {
            Expect(a+b).To(Equal(c))
        },
        func(a, b, c int) string {
            return fmt.Sprintf("%d + %d = %d", a, b, c)
        }
        Entry(nil, 1, 2, 3),
        Entry(nil, -1, 2, 1),
    )
})
```

This will generate entries named `1 + 2 = 3`, and `-1 + 2 = 1`.

There's also a convience decorator called `EntryDescription` to specify Entry descriptions as format strings:


```go
var _ = Describe("Math", func() {
    DescribeTable("addition",
        func(a, b, c int) {
            Expect(a+b).To(Equal(c))
        },
        EntryDescription("%d + %d = %d")
        Entry(nil, 1, 2, 3),
        Entry(nil, -1, 2, 1),
    )
})
```

This will generate entries named `1 + 2 = 3`, and `-1 + 2 = 1`.


Note that only Entries that explicitly set their description to `nil` will have their descriptions auto-generated.

In addition to `nil` and strings you can also pass a string-returning function or an `EntryDescription` as the first argument to `Entry`.  Doing so will cause the entry's description to be generated by the passed-in function or `EntryDescription` format string.

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


#### Managing Complex Parameters

While passing arbitrary parameters to `Entry` is convenient it can make the test cases difficult to parse at a glance.  For more complex tables it may make more sense to define a new type and pass it around instead.  For example:

```go
package table_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

var _ = Describe("Substring matching", func() {
    type SubstringCase struct {
        String    string
        Substring string
        Count     int
    }

    DescribeTable("counting substring matches",
        func(c SubstringCase) {
            Ω(strings.Count(c.String, c.Substring)).Should(BeNumerically("==", c.Count))
        },
        Entry("with no matching substring", SubstringCase{
            String:    "the sixth sheikh's sixth sheep's sick",
            Substring: "emir",
            Count:     0,
        }),
        Entry("with one matching substring", SubstringCase{
            String:    "the sixth sheikh's sixth sheep's sick",
            Substring: "sheep",
            Count:     1,
        }),
        Entry("with many matching substring", SubstringCase{
            String:    "the sixth sheikh's sixth sheep's sick",
            Substring: "si",
            Count:     3,
        }),
    )
})
```

Note that this pattern uses the same DSL, it's simply a way to manage the parameters flowing between the `Entry` cases and the callback registered with `DescribeTable`.

---

## The Spec Runner

### Pending Specs

You can mark an individual spec or container as Pending.  This will prevent the spec (or specs within the container) from running.  You do this by adding a `P` or an `X` in front of your `Describe`, `Context`, `It`, and `Measure`:

```go
PDescribe("some behavior", func() { ... })
PContext("some scenario", func() { ... })
PIt("some assertion")
PMeasure("some measurement")

XDescribe("some behavior", func() { ... })
XContext("some scenario", func() { ... })
XIt("some assertion")
XMeasure("some measurement")
```

> You don't need to remove the `func() { ... }` when you mark an `It` or `Measure` as pending.  Ginkgo will happily ignore any arguments after the string.

> By default, Ginkgo will not fail a suite for having pending specs.  You can pass the `--failOnPending` flag to reverse this behavior.

Using the `P` and `X` prefixes marks specs as pending at compile time.  If you need to skip a spec at *runtime* (perhaps due to a constraint that can only be known at runtime) you may call `Skip` in your test:

```go
It("should do something, if it can", func() {
    if !someCondition {
        Skip("special condition wasn't met")
    }

    // assertions go here
})
```

Note that `Skip(...)` causes the closure to exit so there is no need to return.

### Filtering Specs

It is often convenient to be able to run a subset of specs.  Ginkgo has several mechanisms for allowing you to filter specs:

#### Programattic Filtering

You can focus individual specs or containers of specs *programatically* by adding an `F` in front of your `Describe`, `Context`, and `It` or by using the `Focus` decorator:

```go
FDescribe("some behavior", func() { ... })
FContext("some scenario", func() { ... })
FIt("some assertion", func() { ... })
It("some other assertion", Focus, func() { ... })
```

doing so instructs Ginkgo to only run those specs.  To run all specs, you'll need to go back and remove all the `F`s and `Focus`es.

Nested programmatically focused specs follow a simple rule: if a leaf-node is marked focused, any of its ancestor nodes that are marked focus will be unfocused.  With this rule, sibling leaf nodes (regardless of relative-depth) that are focused will run regardless of the focus of a shared ancestor; and non-focused siblings will not run regardless of the focus of the shared ancestor or the relative depths of the siblings.  More simply:

```go
FDescribe("outer describe", func() {
    It("A", func() { ... })
    It("B", func() { ... })
})
```

will run both `It`s but

```go
FDescribe("outer describe", func() {
    It("A", func() { ... })
    FIt("B", func() { ... })
})
```

will only run `B`.  This behavior tends to map more closely to what the developer actually intends when iterating on a test suite.

When Ginkgo detects that a passing test suite has programmatically focused tests it causes the suite to exit with a non-zero status code.  This is to help detect erroneously committed focused tests on CI systems.  

> You can unfocus programatically focused tests by running `ginkgo unfocus`.  This will strip the `F`s off of any `FDescribe`, `FContext`, and `FIt`s that your tests in the current directory may have.

#### Spec Labels

Users can label specs using the [`Label` decoration](#the-label-decoration).  Labels provide fine-grained control for organizing specs and running specific subsets of labelled specs.  Labels are arbitrary strings however they cannot contain the characters `"&|!,()/"`.  A given spec inherits the labels of all its containers and any labels attached to the spec's `It`, for example:

```
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

You can filter by label using the `ginkgo --label-filter` flag.  Label filter accepts a simple filter language that supports the following:

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

To list the labels used in a given package you can use the `ginkgo labels` command.  This does a simple/naive scan of your test files for calls to `Label` and returns any labels it finds.

#### Command-line Filtering

Ginkgo allows you to filter specs via the command line.  This command-line based filtering will always override programatic filtering, however Pending tests can never be forced to run from the command line, they must be unmarked as pending in source, first.  Unlike programattic filtering, command-line filtering does not alter Ginkgo's exit code.

There are two command-line filtering mechanisms provide - filtering by filename, and filtering by spec text.

The `--focus=REGEXP` and `--skip=REGEXP` flags allow you to focus and/or skip specs on the basis of their spec text.  The spec text is the fully concatenated string comprised of the texts of the spec's containers and the spec itself.  For example:

```go
Describe("Measuring widgets", func() {
    Context("when they are short", func() {
        It("returns zero", func() {

        })
    })
})

```

will have the spec text `"Measuring widgets when they are short returns zero"`.

When `--focus` and/or `--skip` are provided Ginkgo will _only_ run specs with texts that match the focus regex **and** _don't_ match the skip regex.  You can provide `--focus` and `--skip` multiple times.  The `--focus` filters will be ORed together and the `--skip` tests will be ORed together.  For example, say you have the following specs:

```go
It("likes dogs", func() {...})
It("likes purple dogs", func() {...})
It("likes cats", func() {...})
It("likes dog fish", func() {...})
It("likes cat fish", func() {...})
It("likes fish", func() {...})
```

then `ginkgo --focus=dog --focus=fish --skip=cat --skip=purple` will only run `"likes dogs"`, `"likes dog fish"`, and `"likes fish"`.

The `--focus-file` and `--skip-file` flags allow you to focus and/or skip specs on the basis of they file they are in.  When provided Gingo will only run specs that are in files that _do_ match the `--focus-file` filter *and* _don't_ match the `--skip-file` filter.  You can provide multiple `--focus-file` and `--skip-file` flags.  The `--focus-file`s will be ORed together and the `--skip-file`s will be ORed together.

The argument passed to `--focus-file`/`--skip-file` is a file filter and takes one of the following forms:

- `FILE_REGEX` - will match specs in files who's absolute path matches the FILE_REGEX.  So `ginkgo --focus-file=foo` will match specs in files like `foo_test.go` or `/foo/bar_test.go`.
- `FILE_REGEX:LINE` - will match specs in files that match FILE_REGEX where at least one node in the Spec (e.g. a `Describe` node, or an `It` node) is called at line number `LINE`.
- `FILE_REGEX:LINE1-LINE2` - will match specs in files that match FILE_REGEX where at least one node in the Spec (e.g. a `Describe` node, or an `It` node) is called at a line within the range of `[LINE1:LINE2)`. 

You can specify multiple comma-separated `LINE` and `LINE1-LINE2` arguments in a single `--focus-file/--skip-file` (e.g. `--focus-file=foo:1,2,10-12` will apply filters for line 1, line 2, and the range [10-12)).  To specify multiple files, pass in multiple `--focus-file` or `--skip-file` flags.

When `-label-filter`, -focus`, `-skip`, `-focus-file`, and `-skip-file` are all provided they are all ANDed together.  Meaning a given spec MUST be in a file:line matching the focus-file filter, AND MUST NOT be in a file:line matching the skip-file filter AND MUST have text matching the focus filter AND MUST NOT have text matching the skip filter AND MUST have labels matching the label filter.

> If you want to skip entire packages (when running `ginkgo` recursively with the `-r` flag) you should use `--skip-package` instead of `--skip-file`.  `--skip-package` takes a comma-separated list of packages - any packages with *paths* that contain one of the entries in this comma separated list will not be compiled and will be skipped entirely.  Simply using `--skip-file` does not prevent package compilation and you can end up compiling and running packages that skip all their tests.

### Spec Permutation

By default, Ginkgo will randomize the order in which your specs are run.  This can help suss out test pollution early on in a suite's development.

Ginkgo's default behavior is to only permute the order of top-level containers -- the specs *within* those containers continue to run in the order in which they are specified in the test file.  This is helpful when developing specs as it mitigates the coginitive overload of having specs continuously change the order in which they run.

To randomize *all* specs in a suite, you can pass the `--randomize-all` flag.  This is useful on CI and can greatly help fight the scourge of test pollution.  Note that specs in containers marked as `Ordered` are never randomized.

Ginkgo uses the current time to seed the randomization.  It prints out the seed near the beginning of the test output.  If you notice test intermittent test failures that you think may be due to test pollution, you can use the seed from a failing suite to exactly reproduce the spec order for that suite.  To do this pass the `--seed=SEED` flag.

When running multiple spec suites Ginkgo defaults to running the suites in the order they would be listed on the file-system.  You can permute the suites by passing `ginkgo --randomizeSuites`

### Repeating Test Runs and Managing Flakey Tests

It is sometimes useful to rerun a test suite repeatedly - for example, to ensure there are no flakey tests or race conditions.  The `ginkgo` CLI provides two flags that support this:

- `ginkgo --until-it-fails`: will rerun the target test suites repeatedly until a failure occurs.  This flag is best paired with `--randomize-all` and `--randomize-suites` to ensure all tests are permuted between each run.  `--until-it-fails` can be useful when debugging a flaky test in a development environment.

- `ginkgo --repeat=N`: will repeat the test suite up to `N` times or until a failure occurs, whichever comes first.  Like `--until-it-fails`, `--repeat` is best paired with `--randomize-all` and `--randomize-suites`.  `--repeat` can be useful in CI environments to root out potential flakey tests.

> Note, you should not call `RunSpecs` more than once in your test suite.  Ginkgo will exit with an error if you attempt to do so - use `--repeat` or `--until-it-fails` instead.

It is strongly recommended that flakey tests be identified quickly and fixed.  Sometimes, however, flakey tests can't be prevented and retries are necessary.  The `ginkgo` CLI supports this at a global level with `ginkgo --flake-attempts=N`.  Any individual test that fails will be run up-to `N` times.  If it succeeds in any of those tries the test suite is considered successful.  For a more granular approach you can decorate individual tests or test containers with [the `FlakeAttempts(N)` decoration](#the-flakeattempts-decoration)

### Parallel Specs

Ginkgo has support for running specs in parallel.  It does this by spawning separate `go test` processes and serving specs to each process off of a shared queue.  This is important for a BDD test framework, as the shared context of the closures does not parallelize well in-process.

To run a Ginkgo suite in parallel you *must* use the `ginkgo` CLI.  Simply pass in the `-p` flag:

    ginkgo -p

this will automatically detect the optimal number of test nodes to spawn (see the note below).

To specify the number of nodes to spawn, use `-nodes`:

    ginkgo -nodes=N

> You do not need to specify both `-p` and `-nodes`.  Setting `-nodes` to anything greater than 1 implies a parallelized test run.

> The number of nodes used with `-p` is `runtime.NumCPU()` if `runtime.NumCPU() <= 4`, otherwise it is `runtime.NumCPU() - 1` based on a rigorous science based heuristic best characterized as "my gut sense based on a few months of experience"

The test runner collates output from the running processes into one coherent output.  This is done, under the hood, via a client-server model: as each client suite completes a test, the test output and status is sent to the server which then prints to screen.  This collates the output of simultaneous test runners to one coherent (i.e. non-interleaved), aggregated, output.

#### Marking Specs as Serial

By default all specs are distributed in parallel across Ginkgo's running test processes.  However there can be contexts where specific specs need to run in series and must not be run in parallel with other specs.  You can tell Ginkgo about these specs using the [Serial decorator](#the-serial-decoration).  Specs marked `Serial` will run on process #1 after all other processes have finished running parallel specs and exited.

In addition, specs that appear in containers decorated with `Ordered` are guaranteed to run in order.  When running in parallel Ginkgo will ensure that specs in `Ordered` containers are all scheduled on the same parallel process, in order.  Note that these specs can still run in parallel with _other_ specs, just not with each other.  To ensure they run in order and in series, use both the `Ordered` and `Serial` decorator on the container.

#### Managing External Processes in Parallel Test Suites

If your tests spin up or connect to external processes you'll need to make sure that those connections are safe in a parallel context.  One way to ensure this would be, for example, to spin up a separate instance of an external resource for each Ginkgo process.  For example, let's say your tests spin up and hit a database.  You could bring up a different database server bound to a different port for each of your parallel processes:

```go
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
    port := 4000 + GinkgoParallelProcess()

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
```


Ginkgo provides access to the index of the current parallel process via `GinkgoParallelProcess()` - it is one-indexed.  To fetch the total number of nodes you can get the suite configuration from `suiteConfig, _ := GinkgoConfiguration()` and access `suiteConfig.ParallelTotal`.

#### Managing *Singleton* External Processes in Parallel Test Suites

When possible, you should make every effort to start up a new instance of an external resource for every parallel node.  This helps avoid test-pollution by strictly separating each parallel node.

Sometimes (rarely) this is not possible.  Perhaps, for reasons beyond your control, you can only start one instance of a service on your machine.  Ginkgo provides a workaround for this with `SynchronizedBeforeSuite` and `SynchronizedAfterSuite`.

The idea here is simple.  With `SynchronizedBeforeSuite` Ginkgo gives you a way to run some preliminary setup code on just one parallel node (Node 1) and other setup code on all nodes.  Ginkgo synchronizes these functions and guarantees that node 1 will complete its preliminary setup before the other nodes run their setup code.  Moreover, Ginkgo makes it possible for the preliminary setup code on the first node to pass information on to the setup code on the other nodes.

Here's what our earlier database example looks like using `SynchronizedBeforeSuite`:

```go
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
```

`SynchronizedBeforeSuite` must be passed two functions.  The first must return `[]byte` and the second must accept `[]byte`.  When running with multiple nodes the *first* function is only run on node 1.  When this function completes, all nodes (including node 1) proceed to run the *second* function and will receive the data returned by the first function.  In this example, we use this data-passing mechanism to forward the database's address (set up on node 1) to all nodes.

To clean up correctly, you should use `SynchronizedAfterSuite`.  Continuing our example:

```go
var _ = SynchronizedAfterSuite(func() {
    dbClient.Cleanup()
}, func() {
    dbRunner.Stop()
})
```

With `SynchronizedAfterSuite` the *first* function is run on *all* nodes (including node 1).  The *second* function is only run on node 1.  Moreover, the second function is only run when all other nodes have finished running.  This is important, since node 1 is responsible for setting up and tearing down the singleton resources it must wait for the other nodes to end before tearing down the resources they depend on.

Finally, all of these function can be passed an additional `Done` parameter to run asynchronously.  When running asynchronously, an optional timeout can be provided as a third parameter to `SynchronizedBeforeSuite` and `SynchronizedAfterSuite`.  The same timeout is applied to both functions.

> Note an important subtelty:  The `dbRunner` variable is *only* populated on Node 1.  No other node should attempt to touch the data in that variable (it will be nil on the other nodes).  The `dbClient` variable, which is populated in the second `SynchronizedBeforeSuite` function is, of course, available across all nodes.

### Node Decoration Reference
We've seen a number of node decorations detailed throughout this documentation.  This reference collects them all in one place.

#### Node Decorations Overview
The Ginkgo container nodes (`Describe`, `Context`, `When`), Ginkgo subject nodes (`It`, `Specify`), and Ginkgo setup nodes (`BeforeEach`, `AfterEach`, `JustBeforeEach`, `JustAfterEach`) can all be decorated.  Decorations are specially typed arguments passed into the node constructors.  They can appear anywhere in the `args ...interface{}` list in the constructor signatures:

```go
func Describe(text string, args ...interface{})
func It(text string, args ...interface{})
func BeforeEach(args ...interface{})
```

Ginkgo will vet the passed in decorations and exit with a clear error message if it detects any invalid configurations. 

Moreover, Ginkgo also supports passing in arbitrarily nested slices of decorators.  Ginkgo will unroll these slices and process the flattened list.  This makes it easier to pass around groups of decorators.  For example, this is valid:

```go
flakyDecorations := []interface{}{Label("flaky"), FlakeAttempts(3)}

var _ = Describe("a bunch of flaky controller tests", flakyDecorations, Label("controller"), func() {
    ...
}
```
The resulting tests will be decorated with `FlakeAttempts(3)` and the two labels `flaky` and `controller`.

#### The `Serial` Decoration
The `Serial` decoration applies to container nodes and subject nodes only.  It is an error to try to apply the `Serial` decoration to a setup node.

`Serial` allows the user to mark specs and containers of specs as only eligible to run in serial.  Ginkgo will guarantee that these specs never run in parallel with other specs.

If a container is marked as `Serial` then all the specs defined in that container will be marked as `Serial`.

You cannot mark specs and containers as `Serial` if they appear in an `Ordered` container.  Instead, mark the `Ordered` container as `Serial`.

#### The `Ordered` Decoration
The `Ordered` decoration applies to container nodes only.  It is an error to try to apply the `Ordered` decoration to a setup or subject node.  It is an error to nest an `Ordered` container within another `Ordered` container - however you may nest an `Ordered` container within a non-ordered container and vice versa.

`Ordered` allows the user to [mark containers of specs as ordered](#ordered-containers).  Ginkgo will guarantee that the container's specs will run in the order they appear in and will never run in parallel with one another (though they may run in parallel with other specs unless the `Serial` decoration is also applied to the `Ordered` container).

When a spec in an `Ordered` container fails, all subsequent specs in the ordered container are skipped.

#### The `Label` Decoration
The `Label` decoration applies to container nodes and subject nodes only.  It is an error to try to apply the `Label` decoration to a setup node.

`Label` allows the user to annotate specs and containers of specs with labels.  The `Label` decoration takes a variadic set of strings allowing you to apply multiple labels simultaneously.  Labels are arbitrary strings that do not include the characters `"&|!,()/"`.  Specs can have as many labels as you'd like and the set of labels for a given set is the union of all the labels of the container nodes and the subject node.

Labels can be used to control which subset of tests to run.  This is done by providing the `--label-filter` flag to the `ginkgo` cli.  More details can be found at [Spec Labels](#spec-labels).

#### The `Focus` and `Pending` Decoration
The `Focus` and `Pending` decorations apply to container nodes and subject nodes only.  It is an error to try to `Focus` or `Pending` a setup node.

Using these decorators is identical to using the `FX` or `PX` form of the node constructor.  For example:

```go
FDescribe("container", func() {
    It("runs", func() {})
    PIt("is pending", func() {})
})
```

and

```go
Describe("container", Focus, func() {
    It("runs", func() {})
    It("is pending", Pending, func() {})
})
```

are equivalent.

It is an error to decorate a node as both `Pending` and `Focus`:
```go
It("is invalid", Focus, Pending, func() {}) //this will cause Ginkgo to exit with an error
```

The `Focus` and `Pending` decorations are propagated through the test hierarchy as described in [Pending Specs](#pending-specs) and [Filtering Specs](#filtering-specs)

#### The `Offset` Decoration
The `Offset(uint)` decoration applies to all decorable nodes.  The `Offset(uint)` decoration allows the user to change the stack-frame offset used to compute the location of the test node.  This is useful when building shared test behaviors.  For example:

```
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

`Offset`s only apply to the node that they decorate.  Setting the `Offset` for a container node does not affect the `Offset`s computed in its child nodes.

If multiple `Offset`s are provided on a given node, only the last one is used.

#### The `CodeLocation` Decoration
In addition to `Offset`, users can decorate nodes with a `types.CodeLocation`.  `CodeLocation`s are the structs Ginkgo uses to capture location information.  You can, for example, set a custom location using `types.NewCustomCodeLocation(message string)`.  Now when the location of the node is emitted the passed in `message` will be printed out instead of the usual `file:line` location.

Passing a `types.CodeLocation` decoraiton in has the same semantics as passing `Offset` in: it only applies to the node in question.

#### The `FlakeAttempts` Decoration
The `FlakeAttempts(uint)` decoration applies container and subject nodes.  It is an error to apply `FlakeAttempts` to a setup node.

`FlakeAttempts` allows the user to flag specific tests or groups of tests as potentially flaky.  Ginkgo will run tests up to the number of times specified in `FlakeAttempts` until they pass.  For example:

```
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

With this setup, `"is flaky"` and `"is also flaky"` will run up to 3 times.  `"is _really_ flaky"` will run up to 5 times.  `"is _not_ flaky"` will run only once.  Note that if multiple `FlakeAttempts` appear in a spec's hierarchy, the most deeply nested `FlakeAttempts` wins.  If multiple `FlakeAttempts` are passed into a given node, the last one wins.

If `ginkgo --flake-attempts=N` is set the value passed in by the CLI will override all the decorated values.  Every test will now run up to `N` times.

### Interrupting and Aborting Test Runs

Users can end a test run by sending a `SIGINT` or `SIGTERM` signal (or, in a terminal, just hit `^C`).  When an interrupt signal is received Ginkgo will:

- Immediately interrupt the current test.
- Run any `AfterEach` and `JustAfterEach` components associated with the current test.
- Emit as much information about the interrupted test as possible, this includes:
    - Anything written to the `GinkgoWriter`
    - The location of the Ginkgo component that was running at the time of interrupt
    - A full stacktrace dump of all running goroutines at the time of interrupt
- Skip any subsequent tests
- Run the `AfterSuite` component if present
- Exit, marking the test suite as failed

Should any of the components run after receiving an interrupt get stuck, subsequent interrupt signals can be sent by the user to interrupt them as well.

### Profiling your Test Suites

Go supports a rich set of profiling features to gather information about your running test suite.  Ginkgo exposes all of these and manages them for you when you are running multiple test suites and/or test suites in parallel.

Ginkg supports `-race` to analyze race conditions, `-cover` to compute code coverage, `-vet` to evaluate and vet your code, `-cpuprofile` to profile CPU performacne, `-memprofile` to profile memory usage, `-blockprofile` to profile blocking goroutines, and `-mutexprofile` to profile locking around mutexes.

`ginkgo -race` runs the race detector and emits any detected race conditions as the test runs.  If any are detected the test is failed.

`ginkgo -vet` allows you to configure the set of checks that are applied when your code is compiled.  `ginkgo` defaults to the set of default checks that `go test` uses and you can specify additional checks by passing a comma-separated list to `-vet`.  The set of available checks can be found by running `go doc cmd/vet`.

#### Computing Coverage

`ginkgo -cover` will compute and emit code coverage.  When running multiple suites Ginkgo will emit coverage for each suite and then emit a composite coverage across all running suites.  As with `go test` the default behavior for a given test suite is to measure the coverage it provides for the code in the test suite's package - however you can extend coverage to additional packages using `-coverpkg`.  You can also specify the `-covermode` to be one of `set` ("was this code called at all?"), `count` (how many times was it called?) and `atomic` (same as count, but threadsafe and expensive).  If you run `ginkgo -race -cover` the `-covermode` is automatically set to `atomic`.

When run with `-cover`, Ginkgo will generate a single `coverprofile.out` file that captures the coverage statistics of all test suites run.  You can change the name of this file by specifiying `-coverprofile=filename`.  If you would like to keep separate coverprofiles for each test suite use the `-keep-separate-coverprofiles` option.

Ginkgo also honors the `-output-dir` flag when generating coverprofiles.  If you specify `-output-dir` the generated coverprofile will be placed in the requested directory.  If you also specify `-keep-separate-coverprofiles` individual package coverprofiles will be placed in the requested directory and namespaced with a prefix that contains the nmae of the package in question.

#### Other Profiles

Running `ginkgo` with any of `-cpuprofile=X`, `-memprofile=X`, `-blockprofile=X`, and `-mutexprofile=X` will generate corresponding profile files for each test package run.  Doing so will also preserve the test binary generated by Ginkgo to enable users to use `go tool pprof <BINARY> <PROFILE>` to analyze the profile.

By default, the test binary and various profile files are stored in the individual directories of any test suites that Ginkgo runs.  If you specify `-output-dir`, however, then these assets are moved to the requested directory and namespaced with a prefix that contains the name of the package in question.

---

## Understanding Ginkgo's Lifecycle

Users of Ginkgo sometimes get tripped up by Ginkgo's lifecycle.  This section provides a mental model to help you reason about what code runs when.

Ginkgo endeavors to carefully control the order in which specs run and provides seamless support for running a given test suite in parallel across [multiple processes](#parallel-specs).  To accomplish this, Ginkgo needs to know a suite's entire testing tree (i.e. the nested set of `Describe`s, `Context`s, `BeforeEach`es, `It`s, etc.) **up front**.  Ginkgo uses this tree to construct an ordered, ([deterministically randomized](#spec-permutation)), list of tests to run.

This means that all the tests must be defined _before_ Ginkgo can run the suite.  Once the suite is running it is an error to attempt to define a new test (e.g. calling `It` within an existing `It` block).

Of course, it is still possible (in fact, common) to dynamically generate test suites based on configuration.  However you must generate these tests _at the right time_ in the Ginkgo lifecycle.  This nuance sometimes trips users up.

Let's look at a typical Ginkgo test suite.  What follows is a test suite for a `books` package that spans multiple files:


```go
// books_suite_test.go

package books_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    "github.com/onsi/books"

    "testing"
)

func TestBooks(t *testing.T) {     // L1
    RegisterFailHandler(Fail)      // L2
    RunSpecs(t, "Books Suite")     // L3
}                                  // L4
```

```go
// reading_test.go

package books_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    "github.com/onsi/books"

    "testing"
)

var _ = Describe("When reading a book", func() {                        // L5
    var book *books.Book                                                // L6

    BeforeEach(func() {                                                 // L7
        book = books.New("The Chronicles of Narnia", 300)               // L8
        Expect(book.CurrentPage()).To(Equal(1))                         // L9
        Expect(book.NumPages()).To(Equal(300))                          // L10
    })                                                                  // L11

    It("should increment the page number", func() {                     // L12
        err := book.Read(3)                                             // L13
        Expect(err).NotTo(HaveOccurred())                               // L14
        Expect(book.CurrentPage()).To(Equal(4))                         // L15
    })                                                                  // L16

    Context("when the reader finishes the book", func() {               // L17
        It("should not allow them to read more pages", func() {         // L18
            err := book.Read(300)                                       // L19
            Expect(err).NotTo(HaveOccurred())                           // L20
            Expect(book.IsFinished()).To(BeTrue())                      // L21
            err = book.Read(1)                                          // L22
            Expect(err).To(HaveOccurred())                              // L23
        })                                                              // L24
    })                                                                  // L25
})                                                                      // L26
```

```go
// isbn_test.go

package books_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    "github.com/onsi/books"

    "testing"
)

var _ = Describe("Looking up ISBN numbers", func() {                                                   // L27
    Context("When the book can be found", func() {                                                     // L28
        It("returns the correct ISBN number", func() {                                                 // L29
            Expect(books.ISBNFor("The Chronicles of Narnia", "C.S. Lewis")).To(Equal("9780060598242")) // L30
        })                                                                                             // L31
    })                                                                                                 // L32

    Context("When the book can't be found", func() {                                                   // L33
        It("returns an error", func() {                                                                // L34
            isbn, err := books.ISBNFor("The Chronicles of Blarnia", "C.S. Lewis")                      // L35
            Expect(isbn).To(BeZero())                                                                  // L36
            Expect(err).To(HaveOccurred())                                                             // L37
        })                                                                                             // L38
    })                                                                                                 // L39
})                                                                                                     // L40
```

Here's what happens when this test is run via the `ginkgo` cli, in order:

1. `ginkgo` runs `go test -c` to compile the test binary
2. `ginkgo` launches the test binary (this is equivalent to simply running `go test` but gives Ginkgo the ability to launch multiple test processes without paying the cost of repeat compilation)
3. The test binary loads into memory and all top-level functions are defined and invoked.  Specifically, that means:
    1. The `TestBooks` function is defined (Line `L1`)
    2. The `Describe` at `L5` is invoked and passed in the string "When reading a book", and the anonymous function that contains the tests nested in this describe.
    3. The `Describe` at `L27` is invoked and passed in the string "Looking up ISBN numbers", and the anonymous function that contains the tests nested in this describe.

    Note that the anonymous functions passed into the `Describe`s are *not* invoked at this time.  At this point, Ginkgo simply knows that there are two top-level containers in this suite.
4. The `go test` runtime starts running tests by invoking `TestBooks()` (`L1`)
5. Ginkgo's `Fail` handler is registered with `gomega` via `RegisterFailHandler` (`L2`) - this is necessary because `ginkgo` and `gomega` are not tightly coupled and alternative matcher libraries can be used with Ginkgo.
6. `RunSpecs` is called (`L3`).  This does a number of things:
    
    **Test Tree Construction Phase**:
    1. Ginkgo iterates through every top-level container(i.e. the two `Describe`s at `L5` and `L27`) and invokes their anonymous functions.
    2. When the function at `L5` is invoked it:
        - Defines a closure variable named `book` (`L6`)
        - Registers a `BeforeEach` passing it an anonymous function (`L7`).  This function is registered and saved as part of the testing tree and **is not run yet**.
        - Registers an `It` with a description and anonymous function. (`L12`).  This function is registered and saved as part of the testing tree and **is not run yet**.
        - Adds a nested `Context` (`L17`).  The anonymous function passed into the `Context` is **immediately** invoked to continue building the tree.  This registers the `It` at line `L18`.
        - At this point the anonymous function at `L5` exists **and is never called again**.
    3. The function for the next top-level describe at `L27` is also invoked, and behaves similarly.
    4. At this point all top-level containers have been invoked and the testing tree looks like:
        ```
        [
          ["When Reading A Book", BeforeEach, It "should increment the page number"],
          ["When Reading A Book", BeforeEach, "When the reader finishes the book", It "should not allow them to read more pages"],          
          ["Looking up ISBN numbers", "When the book can be found", It "returns the correct ISBN number"],
          ["Looking up ISBN numbers", "When the book can't be found", It "returns an error"],
        ]
        ```
        Here, the `BeforeEach` and `It` nodes in the tree all contain references to their respective anonymous functions.  Note that there are four tests, each corresponding to one of the four `It`s defined in the test.

    Having constructed the testing tree, Ginkgo can now randomize it deterministically based on the random seed.  This simply amounts to shuffling the list of tests shown above.

    **Test Tree Invocation Phase**:
    To run the tests, Ginkgo now simply walks the shuffled testing tree.  Invoking the anonymous functions attached to any registered `BeforeEach` and `It` in order.  For example, when running the test defined at `L18` Ginkgo will first invoke the anonymous function passed into the `BeforeEach` at `L7` and then the anonymous function passed into the `It` at `L18`.

    > Note, again, that the parent closure defined at `L5` is _not_ reinvoked.  The functions passed into `Describe`s and `Context`s are _only_ invoked during the **Tree Construction Phase**

    > It is an error to define new `It`s, `BeforeEach`es, etc. during the Test Tree Invocation Phase.

7. `RunSpecs` keeps track of running tests and test failures, updating any attached reporters as the test runs.  When the test completes `RunSpecs` exits and the `TestBooks` function exits.


That was a lot of detail but it all boils down to a fairly simple flow.  To summarize:

There are two phases during a Ginkgo test run.

In the **Test Tree Construction Phase** the anonymous functions passed into any containers (i.e. `Describe`s and `Context`s) are invoked.  These functions define closure variables and call child nodes (`It`s, `BeforeEach`es, and `AfterEach`es etc.) to definte the testing tree.  The anonymous functions passed into child nodes are **not** called during the Test Tree Construction Phase. Once constructed the tree is randomized.

In the **Test Tree Invocation Phase** the child node functions are invoked in the correct order.  Note that the container functions are not invoked during this phase.

### Avoiding test pollution

Because the anonymous functions passed into container functions are _not_ reinvoked during the Test Tree Invocation Phase you should not rely on variable initializations in container functions to be called repeatedly.  You *must*, instead, reinitialize any variables that can be mutated by tests in your `BeforeEach` functions.

Consider, for example, this variation of the `"When reading a book"` `Describe` container defined in `L5` above:

```
var _ = Describe("When reading a book", func() {                                        //L1'
    var book *books.Book                                                                //L2'
    book = books.New("The Chronicles of Narnia", 300) // create book in parent closure  //L3'

    It("should increment the page number", func() {                                     //L4'
        err := book.Read(3)                                                             //L5'
        Expect(err).NotTo(HaveOccurred())                                               //L6'
        Expect(book.CurrentPage()).To(Equal(4))                                         //L7'
    })                                                                                  //L8'

    Context("when the reader finishes the book", func() {                               //L9'
        It("should not allow them to read more pages", func() {                         //L10'
            err := book.Read(300)                                                       //L11'
            Expect(err).NotTo(HaveOccurred())                                           //L12'
            Expect(book.IsFinished()).To(BeTrue())                                      //L13'
            err = book.Read(1)                                                          //L14'
            Expect(err).To(HaveOccurred())                                              //L15'
        })                                                                              //L16'
    })                                                                                  //L17'
})                                                                                      //L18'
```

In this variation not only is the book variable shared between both `It`s.  The exact same book instance is shared between both `It`s.  This will lead to confusing test pollution behavior that will vary depending on which order the `It`s are called in.  For example, if the `"should not allow them to read more pages"` test (`L10'`)is invoked first, then the book will already be finished (`L13'`)when the `"should increment the page number"` (`L4'`) test is called resulting in an aberrant test failure.

The correct solution to test pollution like this is to always initialize variables in `BeforeEach` blocks.  This ensures test state is clean between each test run.

### Do not make assertions in container node functions

A related, common, error is to make assertions in the anonymous functions passed into container node.  Assertions must _only_ be made in the functions of child nodes as only those functions run during the Test Tree Invocation Phase.

So, avoid code like this:

```
var _ = Describe("When reading a book", func() {
    var book *books.Book
    book = books.New("The Chronicles of Narnia", 300)
    Expect(book.CurrentPage()).To(Equal(1))
    Expect(book.NumPages()).To(Equal(300))     

    It("...")
})
```

If those assertions fail, they will do so during the Test Tree Construction Phase - not the Test Tree Invocation Phase when Ginkgo is tracking and reporting failures.  Instead, initialize variables and make correctness assertions like these inside a `BeforeEach` block.

### Patterns for dynamically generating tests

A common pattern (and one closely related to the [shared example patterns outlined below](#shared-example-patterns)) is to dynamically generate a test suite based on external input (e.g. a file or an environment variable).

Imagine, for example, a file named `isbn.json` that includes a set of known ISBN lookups:

```json
// isbn.json
[
  {"title": "The Chronicles of Narnia", "author": "C.S. Lewis", "isbn": "9780060598242"},
  {"title": "Ender's Game", "author": "Orson Scott Card", "isbn": "9780765378484"},  
  {"title": "Ender's Game", "author": "Victor Hugo", "isbn": "9780140444308"},  
]
```

You might want to generate a collection of tests, one for each book.  A recommended pattern for this is:

```go
// isbn_test.go

package books_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    "github.com/onsi/books"

    "testing"
)

var _ = Describe("Looking up ISBN numbers", func() {
    testConfigData := loadTestISBNs("isbn.json")           

    Context("When the book can be found", func() {
        for _, d := range testConfigData {
            d := d //necessary to ensure the correct value is passed to the closure
            It("returns the correct ISBN number for " + d.Title, func() {                                                
                Expect(books.ISBNFor(d.Title, d.Author)).To(Equal(d.ISBN))
            })                                                                                            
        }
    })                                                                                                
})                                                                                                    
```

Here `data` is defined and initialized with the contents of the `isbn.json` file during the Test Tree Construction Phase and is then used to define a set of `It`s.

If you have test configuration data like this that you want to share across multiple top-level `Describes` or test files you can either load it in each `Describe` (as shown here) or load it once into a globally shared variable.  The recommended pattern for the latter is to load such variables just prior to the invocation of `RunSpecs`:


```go
// books_suite_test.go

package books_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    "github.com/onsi/books"

    "testing"
)

var testConfigData TestConfigData

func TestBooks(t *testing.T) {
    RegisterFailHandler(Fail) 
    testConfigData = loadTestISBNs("isbn.json")
    RunSpecs(t, "Books Suite")
}                             
```

Here, the `testConfigData` can be referenced in any subsequent `Describe` or `Context` closure and is guaranteed to be initialized as the Test Tree Construction Phase does not begin until `RunSpecs` is invoked.

If you must make an assertion on the `testConfigData` you can do so in a `BeforeSuite` as follows:

```go
func TestBooks(t *testing.T) {
    RegisterFailHandler(Fail) 
    testConfigData = loadTestISBNs("isbn.json")
    RunSpecs(t, "Books Suite")
}

var _ = BeforeSuite(func() {
    Expect(testConfigData).NotTo(BeEmpty())
})
```                           

This works because `BeforeSuite` functions run during the Test Tree Invocation phase and only run once.

Lastly - the most common mistake folks encounter when dynamically generating tests is to set up test configuration data in a node that only runs during the Test Tree Invocation Phase.  For example:

```go

var _ = Describe("Looking up ISBN numbers", func() {
    var testConfigData TestConfigData

    BeforeEach(func() {
        testConfigData = loadTestISBNs("isbn.json") // WRONG!
    })

    Context("When the book can be found", func() {
        for _, d := range testConfigData {
            d := d //necessary to ensure the correct value is passed to the closure
            It("returns the correct ISBN number for " + d.Title, func() {                                                
                Expect(books.ISBNFor(d.Title, d.Author)).To(Equal(d.ISBN))
            })                                                                                            
        }
    })                                                                                                
})  
```

This will generate zero tests as `testConfigData` will be zero during the Test Tree Construction Phase.

---

## Asynchronous Tests

Go does concurrency well.  Ginkgo provides support for testing asynchronicity effectively.

Consider this example:

```go
It("should post to the channel, eventually", func() {
    c := make(chan string, 0)

    go DoSomething(c)
    Expect(<-c).To(ContainSubstring("Done!"))
})
```
This test will block until a response is received over the channel `c`.  A deadlock or timeout is a common failure mode for tests like this, a common pattern in such situations is to add a select statement at the bottom of the function and include a `<-time.After(X)` channel to specify a timeout.

Ginkgo has this pattern built in.  The `body` functions in all non-container blocks (`It`s, `BeforeEach`es, `AfterEach`es, `JustBeforeEach`es, `JustAfterEach`es, and `Benchmark`s) can take an optional `done Done` argument:

```go
It("should post to the channel, eventually", func(done Done) {
    c := make(chan string, 0)

    go DoSomething(c)
    Expect(<-c).To(ContainSubstring("Done!"))
    close(done)
}, 0.2)
```

`Done` is a `chan interface{}`.  When Ginkgo detects that the `done Done` argument has been requested it runs the `body` function as a goroutine, wrapping it with the necessary logic to apply a timeout assertion.  You must either close the `done` channel, or send something (anything) to it to tell Ginkgo that your test has ended.  If your test doesn't end after a timeout period, Ginkgo will fail the test and move on the next one.

The default timeout is 1 second. You can modify this timeout by passing a `float64` (in seconds) after the `body` function.  In this example we set the timeout to 0.2 seconds.

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

Note: the `--` is important!  Only arguments following `--` will be passed to your suite. To parse arguments/custom flags in your test suite, declare a variable and initialize it at the package-level:

```go
var myFlag string
func init() {
    flag.StringVar(&myFlag, "myFlag", "defaultvalue", "myFlag is used to control my behavior")
}
```

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


- `--vv`

    If present, Ginkgo's default reporter will be even more verbose and will emit texts and locations for skipped tests.  Also implies `-v`.


- `--reportPassed`

    If present, Ginkgo's default reporter will print detailed output for passed specs.

- `--reportFile=<file path>`

    Create report output file in specified path (relative or absolute). It will also override a pre-defined path of `ginkgo.Reporter`, and parent directories will be created, if not exist.

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

    Set `-cover` to have the `ginkgo` CLI run your tests with coverage analysis turned on (a Go 1.2+ feature).  Ginkgo will generate coverage profiles under the current directory named `PACKAGE.coverprofile` for each set of package tests that is run.

- `-coverpkg=<PKG1>,<PKG2>`

    Like `-cover`, `-coverpkg` runs your tests with coverage analysis turned on.  However, `-coverpkg` allows you to specify the packages to run the analysis on.  This allows you to get coverage on packages outside of the current package, which is useful for integration tests.  Note that it will not run coverage on the current package by default, you always need to specify all packages you want coverage for.
    The package name should be fully resolved, eg `github.com/onsi/ginkgo/reporters/stenographer`

- `-coverprofile=<FILENAME>`

    Renames coverage results file to a provided filename
    
- `-outputdir=<DIRECTORY>`
   
    Moves coverage results to a specified directory <br />
    When combined with `-coverprofile` will also append them together
    
**Build flags:**

- `-tags`

    Set `-tags` to pass in build tags down to the compilation step.

- `-compilers`

    When compiling multiple test suites (e.g. with `ginkgo -r`) Ginkgo will use `runtime.NumCPU()` to determine the number of compile processes to spin up.  On some environments this is a bad idea.  You can specify th enumber of compilers manually with this flag.

**Failure behavior:**

- `--failOnPending`

    If present, Ginkgo will mark a test suite as failed if it has any pending specs.

- `--failFast`

    If present, Ginkgo will stop the suite right after the first spec failure.

**Watch flags:**

- `--depth=DEPTH`

    When watching packages, Ginkgo also watches those package's dependencies for changes.  The default value for `--depth` is `1` meaning that only the immediate dependencies of a package are monitored.  You can adjust this up to monitor dependencies-of-dependencies, or set it to `0` to only monitor the package itself, not its dependencies.

- `--watchRegExp=WATCH_REG_EXP`

    When watching packages, Ginkgo only monitors files matching the watch regular expression for changes.  The default value is `\.go$` meaning only go files are watched for changes.

**Flaky test mitigation:**

- `--flakeAttempts=ATTEMPTS`

    If a test fails, Gingko can rerun it immediately. Set this flag to a value
    greater than 1 to enable retries. As long as one of the retries succeeds,
    Ginkgo will not consider the test suite to have been failed.

    This flag is dangerous! Don't be tempted to use it to cover up bad tests!

**Miscellaneous:**

- `-dryRun`

    If present, Ginkgo will walk your test suite and report output *without* actually running your tests.  This is best paired with `-v` to preview which tests will run.  Ther ordering of the tests honors the randomization strategy specified by `--seed` and `--randomizeAllSpecs`.  Note that `-dryRun` does not work with parallel tests and Ginkgo will emit an error if you try to use `-dryRun` with `-p` or `-nodes`.

- `-keepGoing`

    By default, when running multiple tests (with -r or a list of packages) Ginkgo will abort when a test fails.  To have Ginkgo run subsequent test suites after a failure you can set -keepGoing.

- `-untilItFails`

    If set to `true`, Ginkgo will keep running your tests until a failure occurs.  This can be useful to help suss out race conditions or flakey tests.  It's best to run this with `--randomizeAllSpecs` and `--randomizeSuites` to permute test order between iterations.

- `--slowSpecThreshold=TIME_IN_SECONDS`

    By default, Ginkgo's default reporter will flag tests that take longer than 5 seconds to run -- this does not fail the suite, it simply notifies you of slow running specs.  You can change this threshold using this flag.

- `-timeout=DURATION`

    Ginkgo will fail the test suite if it takes longer than `DURATION` to run.  The default value is 1 hour.

- `--afterSuiteHook=HOOK_COMMAND`

    Ginko has the ability to run a command hook after a suite test completes.  You simply give it the command to run and it will do string replacement to pass data into the command.  Example: --afterSuiteHook=”echo  (ginkgo-suite-name) suite tests have [(ginkgo-suite-passed)]”  This suite hook will replace (ginkgo-suite-name) and (ginkgo-suite-passed) with the suite name and pass/fail status respectively, then echo that to the terminal.

- `-requireSuite`

    If you create a package with Ginkgo test files, but forget to run `ginkgo bootstrap` your tests will never run and the suite will always pass. Ginkgo does notify with the message `Found no test suites, did you forget to run "ginkgo bootstrap"?` but doesn't fail. This flag causes Ginkgo to mark the suite as failed if there are test files but nothing that references `RunSpecs.`

### Watching For Changes

The Ginkgo CLI provides a `watch` subcommand that takes (almost) all the flags that the main `ginkgo` command takes.  With `ginkgo watch` ginkgo will monitor the package in the current directory and trigger tests when changes are detected.

You can also run `ginkgo watch -r` to monitor all packages recursively.

For each monitored packaged, Ginkgo will also monitor that package's dependencies and trigger the monitored package's test suite when a change in a dependency is detected.  By default, `ginkgo watch` monitors a package's immediate dependencies.  You can adjust this using the `-depth` flag.  Set `-depth` to `0` to disable monitoring dependencies and set `-depth` to something greater than `1` to monitor deeper down the dependency graph.

### Precompiling Tests

Ginkgo has strong support for writing integration-style acceptance tests.  These tests are useful tools to validate that (for example) a complex distributed system is functioning correctly.  It is often convenient to distribute these acceptance tests as standalone binaries.

Ginkgo allows you to build such binaries with:

    ginkgo build path/to/package

This will produce a precompiled binary called `package.test`.  You can then invoke `package.test` directly to run the test suite.  Under the hood `ginkgo` is simply calling `go test -c -o` to compile the `package.test` binary.

Calling `package.test` directly will run the tests in *series*.  To run the tests in parallel you'll need the `ginkgo` cli to orchestrate the parallel nodes.  You can run:

    ginkgo -p path/to/package.test

to do this.  Since the ginkgo CLI is a single binary you can provide a parallelizable (and therefore *fast*) set of integration-style acceptance tests simply by distributing two binaries.

> The `build` subcommand accepts a subset of the flags that `ginkgo` and `ginkgo watch` take.  These flags are constrained to compile-time concerns such as `--cover` and `--race`.  You can learn more via `ginkgo help build`.

> You can cross-compile and target different platforms using the standard `GOOS` and `GOARCH` environment variables.  So `GOOS=linux GOARCH=amd64 ginkgo build path/to/package` run on OS X will create a `package.test` binary that runs on linux.

### Generators

- To bootstrap a Ginkgo test suite for the package in the current directory, run:

        $ ginkgo bootstrap

    This will generate a file named `PACKAGE_suite_test.go` where PACKAGE is the name of the current directory.

- To add a test file to a package, run:

        $ ginkgo generate <SUBJECT>

    This will generate a file named `SUBJECT_test.go`.  If you don't specify SUBJECT, it will generate a file named `PACKAGE_test.go` where PACKAGE is the name of the current directory.

By default, these generators will dot-import both Ginkgo and Gomega.  To avoid dot imports, you can pass `--nodot` to both subcommands.

You can also pass in a custom template using `--template`.  Take a look at the code under `ginkgo/ginkgo/generators` to see what's available in your template.

> Note that you don't have to use either of these generators.  They're just convenient helpers to get you up and running quickly.

### Creating an Outline of Tests

If you want to see an outline of the Ginkgo tests in a file, you can use the `ginkgo outline` command. The following uses the `book_test.go` example from [Getting Started: Writing Your First Test](#getting-started-writing-your-first-test):

    ginkgo outline book_test.go

This generates an outline in a comma-separated-values (CSV) format. Column headers are on the first line, followed by Ginkgo containers, specs, and other identifiers, in the order they appear in the file:

    Name,Text,Start,End,Spec,Focused,Pending
    Describe,Book,124,973,false,false,false
    BeforeEach,,217,507,false,false,false
    Describe,Categorizing book length,513,970,false,false,false
    Context,With more than 300 pages,567,753,false,false,false
    It,should be a novel,624,742,true,false,false
    Context,With fewer than 300 pages,763,963,false,false,false
    It,should be a short story,821,952,true,false,false

The columns are:

- Name (string): The name of a container, spec, or other identifier in the core DSL.
- Text (string): The description of a container or spec. (If it is not a literal, it is undefined in the outline.)
- Start (int): Position of the first character in the container or spec.
- End (int): Position of the character immediately after the container or spec.
- Spec (bool): True, if the identifier is a spec.
- Focused (bool): True, if focused. (Conforms to the rules in [Filtering Specs](#filtering-specs).)
- Pending (bool): True, if pending. (Conforms to the rules in [Pending Specs](#pending-specs).)

You can set a different output format with the `-format` flag. Accepted formats are `csv`, `indent`, and `json`. The `ident` format is like `csv`, but uses identation to show the nesting of containers and specs. Both the `csv` and `json` formats can be read by another program, e.g., an editor plugin that displays a tree view of Ginkgo tests in a file, or presents a menu for the user to quickly navigate to a container or spec.

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

```go
Measure("it should do something hard efficiently", func(b Benchmarker) {
    runtime := b.Time("runtime", func() {
        output := SomethingHard()
        Expect(output).To(Equal(17))
    })

    Ω(runtime.Seconds()).Should(BeNumerically("<", 0.2), "SomethingHard() shouldn't take too long.")

    b.RecordValue("disk usage (in MB)", HowMuchDiskSpaceDidYouUse())
}, 10)
```

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

```go
Time(name string, body func(), info ...Interface{}) time.Duration
```

 method.  `Time` runs the passed in `body` function and records, and returns, its runtime.  The resulting measurements for each sample are aggregated and some simple statistics are computed.  These stats appear in the spec output under the `name` you pass in.  Note that `name` must be unique within the scope of the `Measure` node.

 You can also pass arbitrary information via the optional `info` argument.  This will be passed along to the reporter along with the agreggated runtimes that `Time` measures.  The default reporter presents a string representation of `info`, but you can write a custom reporter to perform something more structured.  For example, you might run several measurements of the same code, but vary some parameter between runs.  You could encode the value of that parameter in `info`, and then have a custom reporter that uses `info` and the statistics provided by Ginkgo to generate a CSV file - or perhaps even a plot.

 If you want to assert that `body` ran within some threshold time, you can make an assertion against `Time`'s return value.

### Recording Arbitrary Values

The `Benchmarker` also provides the

```go
RecordValue(name string, value float64, info ...Interface{})
```

method.  `RecordValue` allows you to record arbitrary numerical data.  These results are aggregated and some simple statistics are computed.  These stats appear in the spec output under the `name` you pass in.  Note that `name` must be unique within the scope of the `Measure` node.

The optional `info` parameter can be used to pass structured data to a custom reporter.  See [Measuring Time](#measuring-time) above for more details.

---

## Shared Example Patterns

Ginkgo doesn't have any explicit support for Shared Examples (also known as Shared Behaviors) but there are a few patterns that you can use to reuse tests across your suite.

### Locally-scoped Shared Behaviors

It is often the case that within a particular suite, there will be a number of different `Context`s that assert the exact same behavior, in that they have identical `It`s within them.  The only difference between these `Context`s is the set up done in their respective `BeforeEach`s.  Rather than repeat the `It`s for these `Context`s, here are two ways to extract the code and avoid repeating yourself.

#### Pattern 1: Extract a function that defines the shared `It`s

Here, we will pull out a function that lives within the same closure that `Context`s live in, that defines the `It`s that are common to those `Context`s.  For example:

```go
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
        AssertFailedBehavior := func() {
            It("should not include JSON in the response", func() {
                Ω((<-response).JSON).Should(BeZero())
            })

            It("should not report success", func() {
                Ω((<-response).Success).Should(BeFalse())
            })
        }

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
    })
})
```

Note that the `AssertFailedBehavior` function is called within the body of the `Context` container block.  The `It`s defined by this function get added to the enclosing container.  Since the function shares the same closure scope we don't need to pass the `response` channel in.

You can put as many `It`s as you wanted into the shared behavior `AssertFailedBehavior` above, and can even nest `It`s within `Context`s inside of `AssertFailedBehavior`.  Although it may not always be the best idea to DRY your test suites excessively, this pattern gives you the ability do so as you see fit.  One drawback of this approach, however, is that you cannot focus or pend a shared behavior group, or examples/contexts within the group.  In other words, you don't get `FAssertFailedBehavior` or `XAssertFailedBehavior` for free.

#### Pattern 2: Extract functions that return closures, and pass the results to `It`s

To understand this pattern, let's just redo the above example with this pattern:

```go
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
        AssertNoJSONInResponse := func() func() {
            return func() {
                Ω((<-response).JSON).Should(BeZero())
            }
        }

        AssertDoesNotReportSuccess := func() func() {
            return func() {
                Ω((<-response).Success).Should(BeFalse())
            }
        }
        Context("when the server does not return a 200", func() {
            BeforeEach(func() {
                fakeServer.Respond(404)
            })

            It("should not include JSON in the response", AssertNoJSONInResponse())
            It("should not report success", AssertDoesNotReportSuccess())
        })

        Context("when the server returns unparseable JSON", func() {
            BeforeEach(func() {
                fakeServer.Succeed("{I'm not JSON!")
            })

            It("should not include JSON in the response", AssertNoJSONInResponse())
            It("should not report success", AssertDoesNotReportSuccess())
        })

        Context("when the request errors", func() {
            BeforeEach(func() {
                fakeServer.Error(errors.New("oops!"))
            })

            It("should not include JSON in the response", AssertNoJSONInResponse())
            It("should not report success", AssertDoesNotReportSuccess())
        })
    })
})
```

Note that this solution is still very compact, especially because there are only two shared `It`s for each `Context`.  There is slightly more repetition here, but it's also slightly more explicit.  The main benefit of this pattern is you can focus and pend individual `It`s in individual `Context`s.

### Global Shared Behaviors

The patterns outlined above work well when the shared behavior is intended to be used within a fixed scope.  If you want to build shared behavior that can be used across different test files (or even different packages) you'll need to tweak the pattern to make it possible to pass inputs in.  We can extend both examples outlined above to illustrate how this might work:

#### Pattern 1

```go
package sharedbehaviors

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

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
```

#### Pattern 2

```go
package sharedbehaviors

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

type FailedResponseBehaviorInputs struct {
    response chan APIResponse
}

func AssertNoJSONInResponese(inputs *FailedResponseBehaviorInputs) func() {
    return func() {
        Ω((<-(inputs.response)).JSON).Should(BeZero())
    }
}

func AssertDoesNotReportSuccess(inputs *FailedResponseBehaviorInputs) func() {
    return func() {
        Ω((<-(inputs.response)).Success).Should(BeFalse())
    }
}
```

Users of the shared behavior must generate and populate a `FailedResponseBehaviorInputs` and pass it in to either `SharedFailedResponseBehavior` or `AssertNoJSONInResponese` and `AssertDoesNotReportSuccess`.  Why do things this way?  Two reasons:

1. Having a stuct to encapsulate the input variables (like `FailedResponseBehaviorInputs`) allows you to clearly stipulate the contract between the the specs and the shared behavior.  The shared behavior *needs* these inputs in order to function correctly.

2. More importantly, inputs like the `response` channel are generally created and/or set in `BeforeEach` blocks.  However the shared behavior functions must be called within a container block and will not have access to any variables specified in a `BeforeEach` as the `BeforeEach` hasn't run yet.  To get around this, we instantiate a `FailedResponseBehaviorInputs` and pass a pointer to it to the shared behavior functions -- in the `BeforeEach` we manipulate the fields of the `FailedResponseBehaviorInputs`, ensuring that their values get communicated to the `It`s generated by the shared behavior.

Here's what the calling test would look like after dot-importing the `sharedbehaviors` package (for brevity we'll combine patterns 1 and 2 in this example):

```go
Describe("my api client", func() {
    var client APIClient
    var fakeServer FakeServer
    var response chan APIResponse
    sharedInputs := FailedResponseBehaviorInputs{}

    BeforeEach(func() {
        sharedInputs.response = make(chan APIResponse, 1)
        fakeServer = NewFakeServer()
        client = NewAPIClient(fakeServer)
        client.Get("/some/endpoint", sharedInputs.response)
    })

    Describe("failure modes", func() {
        Context("when the server does not return a 200", func() {
            BeforeEach(func() {
                fakeServer.Respond(404)
            })

            // Pattern 1
            SharedFailedResponseBehavior(&sharedInputs)
        })

        Context("when the server returns unparseable JSON", func() {
            BeforeEach(func() {
                fakeServer.Succeed("{I'm not JSON!")
            })

            // Pattern 2
            It("should not include JSON in the response", AssertNoJSONInResponse(&sharedInputs))
            It("should not report success", AssertDoesNotReportSuccess(&sharedInputs))
        })
    })
})
```

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
        - 1.9
        - tip

    install:
        - go get -v github.com/onsi/ginkgo/ginkgo
        - go get -v github.com/onsi/gomega
        - go get -v -t ./...
        - export PATH=$PATH:$HOME/gopath/bin

    script: ginkgo -r --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --compilers=2

Note that we've added `--compilers=2` -- this resolves an issue where Travis kills the Ginkgo process early for being too greedy.  By default, Ginkgo will run `runtime.NumCPU()` compilers which, on Travis, can be as many as `32` compilers!  Similarly, if you want to run your tests in parallel on Travis, make sure to specify `--nodes=N` instead of `-p`.

---

## Reporting Infrastucture
Ginkgo provides a reporting infrastructure that supports a number of distinct use-cases:

### Generating machine-readable reports
Ginkgo natively supports generating and aggregating reports in a number of machine-readable formats - and these reports can be generated and managed by simply passing `ginkgo` command line flags.

A JSON-format report that faithfully captures all available information about a Ginkgo test suite can be generated via `ginkgo --json-report=out.json`.  The resulting JSON file encodes an array of `types.Report`.  Each entry in that array lists detailed information about the test suite and includes a list of `types.SpecReport` that captures detailed information about each spec.  These types are documented [here](https://github.com/onsi/ginkgo/blob/ver2/types/types.go).

Ginkgo also supports generating JUnit reports with `ginkgo --junit-report=out.xml` and Teamcity reports with `ginkgo --teamcity-report=out.teamcity`.  In addition, Ginkgo V2's JUnit reporter has been improved and is now more conformant with the JUnit specification.

Ginkgo follows the following rules when generating reports using these new `--FORMAT-report` flags:
- By default, a single report file per format is generated at the passed-in file name.  This single report merges all the reports generated by each individual suite.
- If `-output-dir` is set: the report files are placed in the specified `-output-dir` directory.
- If `-keep-separate-reports` is set, the individual reports generated by each test suite are not merged.  Instead, individual report files will appear in each package directory.  If `-output-dir` is _also_ set these individual files are copied into the `-output-dir` directory and namespaced with `PACKAGE_NAME_{REPORT}`.

### Generating Custom Reports when a test suite completes
Ginkgo provides a `ReportAfterSuite` new node with the following properties and constraints:
- `ReportAfterSuite` nodes are passed a function that takes a `types.Report`:
  ```
  var _ = ReportAfterSuite(func(report types.Report) {
    // do stuff with report
  })
  ```
- These functions are called exactly once at the end of the test suite after any `AfterSuite` nodes have run.  When running in parallel the functions are called on the primary Ginkgo process after all other processes have finished and is guaranteed to have an aggregated copy of `types.Report` that includes all `SpecReport`s from all the Ginkgo parallel processes.
- If a failure occurs in `ReportAfterSuite` it is reported in reports sent to subsequent `ReportAfterSuite`s.  In particular, it is reported as part of Ginkgo's default output and is in included in any reports generated by the `--FORMAT-report` flags described above.
- `ReportAfterSuite` nodes **cannot** be interrupted by the user.  This is to ensure the integrity of generated reports... so be careful what kind of code you put in there!
- Multiple `ReportAfterSuite` nodes can be registered per test suite, but all must be defined at the top-level of the suite.

`ReportAfterSuite` is useful for users who want to emit a custom-formatted report or register the results of the test run with an external service.

### Capturing report information about each spec as the test suite runs
Ginkgo also provides a `ReportAfterEach` node with the following properties and constraints:
- `ReportAfterEach` nodes are passed a function that takes a `types.SpecReport`:
  ```
  var _ = ReportAfterEach(func(specReport types.SpecReport) {
    // do stuff with specReport
  })
  ```
- `ReportAfterEach` nodes are called after a spec completes (i.e. after any `AfterEach` nodes have run).  `ReportAfterEach` nodes are **always** called - even if the test has failed, is marked pending, or is skipped.  In this way, the user is guaranteed to have access to a report about every spec defined in a suite.
- If a failure occurs in `ReportAfterEach`, the spec in question is marked as failed.  Any subsequently defined `ReportAfterEach` block will receive an updated report that includes the failure.  In general, though, assertions about your code should go in `AfterEach` nodes.
- `ReportAfterEach` nodes **cannot** be interrupted by the user.  This is to ensure the integrity of generated reports... so be careful what kind of code you put in there!
- `ReportAfterEach` nodes can be placed in any container node in the suite's hierarchy - in this way the follow the same semantics as `AfterEach` blocks.
- When running in parallel, `ReportAfterEach` nodes will run on the Ginkgo process that is running the spec being reported on.  This means that multiple `ReportAfterEach` blocks can be running concurrently on independent processes.

`ReportAfterEach` is useful if you need to stream or emit up-to-date information about the test suite as it runs.  Ginkgo also provides `ReportBeforeEach` which is called before the test runs and receives a preliminary `types.SpecReport` - the state of this report will indicate whether the test will be skipped or is marked pending.

### Attaching Data to Reports
Ginkgo supports attaching arbitrary data to individual spec reports.  These are called `ReportEntries` and appear in the various report-related data structures (e.g. `Report` in `ReportAfterSuite` and `SpecReport` in `ReportAfterEach`) as well as the machine-readable reports generated by `--json-report`, `--junit-report`, etc.  `ReportEntries` are also emitted to the console by Ginkgo's reporter and you can specify a visibility policy to control when this output is displayed.

You attach data to a spec report via

```
AddReportEntry(name string, args ...interface{})
```

`AddReportEntry` can be called from any runnable node (e.g. `It`, `BeforeEach`, `BeforeSuite`) - but not from the body of a container node (e.g. `Describe`, `Context`).

`AddReportEntry` generates `ReportEntry` and attaches it to the current running spec.  `ReportEntry` includes the passed in `name` as well as the time and source location at which `AddReportEntry` was called.  Users can also attach a single object of arbitrary type to the `ReportEntry` by passing it into `AddReportEntry` - this object is wrapped and stored under `ReportEntry.Value` and is always included in the suite's JSON report.

You can access the report entries attached to a spec by getting the `CurrentSpecReport()` or registering a `ReportAfterEach()` - the returned report will include the attached `ReportEntries`.  You can fetch the value associated with the `ReportEntry` by calling `entry.GetRawValue()`.  When called in-process this returns the object that was passed to `AddReportEntry`.  When called after hydrating a report from JSON `entry.GetRawValue()` will include a parsed JSON `interface{}` - if you want to hydrate the JSON yourself into an object of known type you can `json.Unmarshal([]byte(entry.Value.AsJSON), &object)`.

#### Supported Args
`AddReportEntry` supports the `Offset` and `CodeLocation` decorations.  These will control the source code location associated with the generated `ReportEntry`.  You can also pass in a `time.Time` argument to override the timestamp associated with the `ReportEntry` - this can be helpful if you want to ensure a consistent timestamp between your code and the `ReportEntry`.

You can also pass in a `ReportEntryVisibility` enum to control the report's visibility.  This is discussed in more detail below.

If you pass multiple arguments of the same type (e.g. two `Offset`s), the last argument in wins.  This does mean you cannot attach an object with one of the types discussed in this section as the `ReportEntry.Value`.  To get by this you'll need to define a custom type.  For example, if you want the `Value` to be a `time.Time` timestamp you can use a custom type such as

`type Timestamp time.Time`

#### Controlling Output
By default, Ginkgo's console reporter will emit any `ReportEntry` attached to a spec.  It will emit the `ReportEntry` name, location, and time.  If the `ReportEntry` value is non-nil it will also emit a representation of the value.  If the value implements `fmt.Stringer` or `types.ColorableStringer` then `value.String()` or `value.ColorableString()` (which takes precedence) is used to generate the representation, otherwise Ginkgo uses `fmt.Sprintf("%#v", value)`. 

You can modify this default behavior by passing in one of the `ReportEntryVisibility` enum to `AddReportEntry`:

- `ReportEntryVisibilityAlways`: the default behavior - the `ReportEntry` is always emitted.
- `ReportEntryVisibilityFailureOrVerbose`: the `ReportEntry` is only emitted if the spec fails or the tests are run with `-v` (similar to `GinkgoWriter`s behavior).
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

---

## Third Party Integrations

### Using Other Matcher Libraries

Most matcher library accept the `*testing.T` object.  Unfortunately, since this is a concrete type is can be tricky to pass in an equivalent that will work with Ginkgo.

It is, typically, not difficult to replace `*testing.T` in such libraries with an interface that `*testing.T` satisfies.  For example [testify](https://github.com/stretchr/testify) accepts `t` via an interface.  In such cases you can pass `GinkgoT()`.  This generates an object that mimics `*testing.T` and communicates to Ginkgo directly.

For example, to get testify working:

```go
package foo_test

import (
    . "github.com/onsi/ginkgo"

    "github.com/stretchr/testify/assert"
)

var _ = Describe(func("foo") {
    It("should testify to its correctness", func(){
        assert.Equal(GinkgoT(), foo{}.Name(), "foo")
    })
})
```

> Note that passing the `*testing.T` from Ginkgo's bootstrap `Test...()` function will cause the suite to abort as soon as the first failure is encountered.  Don't do this.  You need to communicate failure to Ginkgo's single (global) `Fail` function

### Integrating with Gomock

Ginkgo does not provide a mocking/stubbing framework.  It's the author's opinion that mocks and stubs can be avoided completely by embracing dependency injection and always injecting Go interfaces.  Real dependencies are then injected in production code, and fake dependencies are injected under test.  Building and maintaining such fakes tends to be straightforward and can allow for clearer and more expressive tests than mocks.

With that said, it is relatively straightforward to use a mocking framework such as [Gomock](https://code.google.com/p/gomock/).  `GinkgoT()` implements Gomock's `TestReporter` interface.  Here's how you use it (for example):

```go
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
```

When using Gomock you may want to run `ginkgo` with the `-trace` flag to print out stack traces for failures which will help you trace down where, in your code, invalid calls occured.

### Generating JUnit XML Output

Ginkgo provides first-class support for generating conformant JUnit XML reports.  Simply run `ginkgo --junit-report=out.xml` to generate a JUnit report at `out.xml`.  This report will include information from all test suites that Ginkgo ran.  To keep reports separate you can use `ginkgo --junit-report=out.xml --keep-separate-reports`.  You can also instruct Ginkgo to place the reports in a given directory by specifying `--output-dir`.
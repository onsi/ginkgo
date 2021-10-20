---
layout: default
title: Ginkgo
---
{% raw  %}
[Ginkgo](https://github.com/onsi/ginkgo) is a Go testing framework built to help you efficiently write expressive and comprehensive tests.  It is best paired with the [Gomega](https://github.com/onsi/gomega) matcher library.

The narrative docs you are reading here are supplemented by the [godoc](https://pkg.go.dev/github.com/onsi/ginkgo) API-level docs.  We suggest starting here to build a mental model for how Ginkgo works and understand how the Ginkgo DSL can be used to solve real-world testing scenarios.  These docs are written assuming you are familiar with Go and the Go toolchain and that you are using Ginkgo V2 (V1 is no longer supported - see [here](https://onsi.github.com/ginkgo/MIGRATING_TO_V2) for the migration guide).

## Why Ginkgo?

This section captures some of Ginkgo's history and motivation.  If you just want to dive straight in, feel free to [jump ahead](#getting-started)!

Like all software projects, Ginkgo was written at a particular time and place to solve a particular set of problems.

The first commit to Ginkgo was made by [@onsi](https://github.com/onsi/) on August 19th, 2013 (to put that timeframe in perspective, it's roughly three months before [Go 1.2](https://golang.org/doc/go1.2) was released!)  Ginkgo came together in the highly collaborative environment fostered by Pivotal, a software company and consultancy that advocated for outcome-oriented software development built by balanced teams that embrace test-driven development.

Specifically, Pivotal was one of the lead contributers to Cloud Foundry.  A sprawling distributed system, originaly written in Ruby, that was slowly migrating towards the emerging distributed systems language of choice: Go.  At the time (and, arguably, to this day) the landscape of Go's testing infrastructure was somewhat anemic.  For engineers coming from the rich ecosystems of testing frameworks such as [Jasmine](https://jasmine.github.io), [rspec](https://rspec.info), and [Cedar](https://github.com/cedarbdd/cedar) there was a need for a comprehensive testing framework with a mature set of matchers in Go.

The need was twofold: organizational and technical. As a growing organization Pivotal needed a shared testing framework to be used across its many teams writing Go.  Engineers jumping from one team to another needed to be able to hit the ground running; we needed fewer testing bikesheds and more shared testing patterns.  And our test-driven development culture put a premium on tests as first-class citizens: they needed to be easy to write, easy to read, and easy to maintain.

Moreover, the _nature_ of the code being built - complex distributed systems - required a testing framework that could provide for the needs unique to unit-testing and integration-testing such a system.  We needed to make testing [asynchronous behavior](https://onsi.github.io/gomega/#making-asynchronous-assertions) ubiqutous and straightforward.  We needed to have [parallelizable integration tests](#parallelism) to ensure our test run-times didn't get out of control.  We needed a test framework that helped us [suss out](#spec-permutation) flaky tests and fix them.

This was the context that led to Ginkgo.  Over the years the Go testing ecosystem has grown and evolved - sometimes [bringing](https://go.dev/blog/subtests) it [closer](https://golang.org/doc/go1.17#testing) to Ginkgo.  Throughout, the community's reactions to Ginkgo have been... interesting.  Some enjoy the expressive framework and rich set of matchers - for many the DSL is familiar and the `ginkgo` CLI is productive.  Others have found the DSL off-putting, arguing that Ginkgo is not "the Go way."  That's OK; the world is plenty large enough for options to abound :)

Happy Testing!

---

## Getting Started

In this section we dive right in and cover installing Ginkgo, Gomega, and the `ginkgo` CLI.  We bootstrap a test suite, write our first test, and run it.

### Installing Ginkgo

Ginkgo uses [go modules](https://go.dev/blog/using-go-modules).  To add Ginkgo to your project, assuming you have a `go.mod` file setup, just `go get` it:

```bash
$> go get github.com/onsi/ginkgo/ginkgo
$> go get github.com/onsi/gomega/...
```

This fetches Ginkgo and installs the `ginkgo` executable under `$GOBIN` - you'll want that on your `$PATH`.  It also fetches the core Gomega matcher library and its set of supporting libraries.

You should now be able to run `ginkgo version` at the command line and see the Ginkgo CLI emit a version number.

### Your First Ginkgo Suite

Ginkgo hooks into Go's existing `testing` infrastructure.  That means that Ginkgo specs live in `X_test.go` files, just like standard go tests.  However, instead of using `func TestX(t *testing.T) {}` to write your tests you use the Ginkgo and Gomega DSLs.  In most Ginkgo suites there is only one `TestX` function - the entry point for Ginkgo.

Let's bootstrap a Ginkgo suite to see what that looks like.

#### Bootstrapping a Suite

Say you have a package named `books` that you'd like to add a Ginkgo suite to.  To bootstrap the suite run:

```bash
$> cd path/to/books
$> ginkgo bootstrap
Generating ginkgo test suite bootstrap for books in:
    books_suite_test.go
```

This will generate a file named `books_suite_test.go` containing:

```go
package books_test

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "testing"
)

func TestBooks(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Books Suite")
}
```

Let's break this down:

First, `ginkgo bootstrap` generates a new test file and places it in the `books_test` package.  That small detail is acutally quite important so let's take a brief detour to discuss test packages.

Normally, Go only allows one package to live in a given directory (in our case, it would be a package named `books`).  There is, however, one exception to this rule: a package ending in `_test` is allowed to live in the same directory as the package being tested.  Doing so instructs Go to compile the test suite as a **separate package**.  This means your test suite will **not** have access to the internals of the `books` package and will need to `import` the `books` package to access its external interface.  Ginkgo defaults to setting up `_test` package as this encourages you to only test the external behavior of your package, not its internal implementation details.

You can, of course, override this - simply change `package books_test` to `package books` or run `ginkgo bootstrap -internal`.  However we recommend sticking with the `_test` package idiom as a best practice.

OK back to our boostrap file.  We import the `ginkgo` and `gomega` packages into the test's top-level namespace by performing a `.` dot-import.  Since Ginkgo and Gomega are DSLs this makes the tests more natural to read.  You can, of course, avoid the dot-import via `ginkgo boostrap --nodot`.  Throughout this documentation we'll be using the dot-import form.

Next we define a single `testing` test: `func TestBooks(t *testing.T)`.  This is the entry point for Ginkgo - the go test runner will run this function when you run `go test` or `ginkgo`.

Inside the `TestBooks` function are two lines:

`RegisterFailHandler(Fail)` is the single line of glue code connecting Ginkgo to Gomega.  If we were to avoid dot-imports this would reas as `gomega.RegisterFailHandler(ginkgo.Fail)` - and what we're doing here is telling our matcher library (Gomega) which function to call (Ginkgo's `Fail`) in the event a failure is detected.

Finally the `RunSpecs()` call tells Ginkgo to start the test suite, passing it the `*testing.T` instance and a description of the suite.  You should only ever call `RunSpecs` once and you can let Ginkgo worry about calling `*testing.T` for you.

You can now run your suite using the `ginkgo` command:

```bash
$> ginkgo

Running Suite: Books Suite - path/to/books
==========================================================
Random Seed: 1634745148

Will run 0 of 0 specs

Ran 0 of 0 Specs in 0.000 seconds
SUCCESS! -- 0 Passed | 0 Failed | 0 Pending | 0 Skipped
PASS

Ginkgo ran 1 suite in Xs
Test Suite Passed
```

We've successfuly set up and run our first suite.  Of course that suite is empty, which isn't very interesting.  Let's add some specs.

#### Adding Specs to a Suite

While you can add all your specs directly into `books_suite_test.go` you'll generally prefer to place your specs in separate files.  This is particularly ture if you have packages with multiple files that need to be tested.  Let's say we have a `book.go` model and we'd like to test its behavior.  We can generate a test file like so:

```bash
$> ginkgo generate book
Generating ginkgo test for Book in:
  book_test.go
```

This will generate a test file named `book_test.go` containing:

```go
package books_test

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    "path/to/books"
)

var _ = Describe("Book", func() {

})
```

As with the bootstrapped suite file, this test file is in the separate `books_test` package and dot-imports both `ginkgo` and `gomega`.  Since we're testing the external interface of `books` Ginkgo adds an `import` statement to pull the `books` package into the test.

Ginkgo then adds an empty top-level `Describe` container node.  `Describe` is part of the Ginkgo DSL and takes a description and a closure function. `Describe("book", func() { })` generates a container that will contain specs that describe the behavior of `"Book"`.

> By default, Go does not allow you to invoke bare functions at the top-level of a file.  Ginkgo gets around this by having its node DSL functions return a value that is intended to be discarded.  This allows us to write `var _ = Describe(...)` at the top-level which satisfies Go's top-level policies.

Let's add a few specs, now, to describe our book model's ability to categorize books:

```go
var _ = Describe("Book", func() {
    var shortBook, longBook book.Book

    BeforeEach(func() {
        longBook = book.Book{
            Title:  "Les Miserables",
            Author: "Victor Hugo",
            Pages:  2783,
        }

        shortBook = book.Book{
            Title:  "Fox In Socks",
            Author: "Dr. Seuss",
            Pages:  24,
        }
    })

    Describe("Categorizing books", func() {
        Context("With more than 300 pages", func() {
            It("should be a novel", func() {
                Expect(longBook.Category()).To(Equal(book.CategoryNovel))
            })
        })

        Context("With fewer than 300 pages", func() {
            It("should be a short story", func() {
                Expect(shortBook.Category()).To(Equal(book.CategoryShortStory))
            })
        })
    })
})
```

There's a lot going on here so let's break it down slowly.

Ginkgo makes extensive use of closures to allow you to build descriptive test suites.  You use **container nodes** like `Describe` and `Context` to expressively organize the different aspects of your code that you are testing.  In this case we are describing our book model's capability to categorize books and we are testing two different contexts - the behavior for large books `"With more than 300 pages"` and small books `"With fewer than 300 pages"`.

You use **setup nodes** like `BeforeEach` to set up the state of your specs.  In this case, we are instantiating two new book models, a `longBook` and a `shortBook`.

Finally, you use **subject nodes** like `It` to write a spec that makes assertions about your code.  In this case, we are ensuring that `book.Category()` returns the correct category enum based on the length of the book.  We make these assertions using Gomega's matchers and `Expect` syntax.  You can learn much more about [Gomega here](https://onsi.github.io/gomega/#making-assertions) - the examples used in these docs should be self-explanatory.

Because there are two subject nodes, Ginkgo will identify two specs to run.  It will run the setup node (`BeforeEach`) before each of these specs and then run the closure passed to the subject node.  In order to share state between the setup node and subject node we define two closure variables `shortBook` and `longBook`.  This is a common pattern and the main way that tests are organized in Ginkgo.

Assuming a `book.Book` model with this behavior we can run the tests:

```bash
$> ginkgo
Running Suite: Books Suite - path/to/books
==========================================================
Random Seed: 1634748172

Will run 2 of 2 specs
••

Ran 2 of 2 Specs in 0.000 seconds
SUCCESS! -- 2 Passed | 0 Failed | 0 Pending | 0 Skipped
PASS

Ginkgo ran 1 suite in Xs
Test Suite Passed
```

Success!  We've written and run our first Ginkgo suite.  The next sections will delve into the mechanisms Ginkgo provides for writing specs and for running specs.

## Writing Specs
  ### Nomenclature
  suite, package, etc.
  ### Spec Containers: `Describe`, `Context`, `When`
  ### Spec Subjects: `It`, `Specify`
    #### Documenting Complex `It`s: `By`
  ### Spec Setup: `BeforeEach` and `JustBeforeEach`
  ### Spec Cleanup: `AfterEach`, `JustAfterEach`
  ### Spec Cleanup: `DeferCleanup`
  ### Marking Specs as Failed
  ### Logging Output
  ### Suite Setup and Cleanup: `BeforeSuite` and `AfterSuite`
  ### Table Tests
    #### Generating Entry Descriptions
    #### Managing Complex Parameters (link to patterns)

## Running Specs
  ### A Mental Model
  #### Spec Permutation
  #### Parallelism
  #### Gotchas: Avoiding test pollution
  #### Gotchas: Do not make assertions in container node functions
  #### `ginkgo` CLI vs `go test`
  ### Pending Specs
  ### Filtering Specs
    #### Programattic Filtering
    #### Spec Labels
    #### Command-line Filtering
  ### Serial Specs
  ### Ordered Containers
  ### Repeating Test Runs and Managing Flakey Tests
  ### Interrupting and Aborting Test Runs

## Reporting and Profiling Suites
  ### Generating machine-readable reports
  ### Generating Custom Reports when a test suite completes
  ### Capturing report information about each spec as the test suite runs
  ### Attaching Data to Reports
    #### Supported Args
    #### Controlling Output
  ### Profiling your Test Suites
    #### Computing Coverage
    #### Other Profiles

## Ginkgo and Gomega Patterns
  ### Configuring Suites Programatically
  ### Custom Command-Line Flags
  ### Dynamically Generating Specs
  ### Benchmarking Code
  ### Managing External Resources in Parallel Test Suites
  ### Locally-scoped Shared Behaviors
    #### Pattern 1: Extract a function that defines the shared `It`s
    #### Pattern 2: Extract functions that return closures, and pass the results to `It`s
  ### Global Shared Behaviors
    #### Pattern 1
    #### Pattern 2
  ### Table Patterns
    #### Managing Complex Parameters

## Decorator Reference
  #### Node Decorations Overview
  #### The `Serial` Decoration
  #### The `Ordered` Decoration
  #### The `Label` Decoration
  #### The `Focus` and `Pending` Decoration
  #### The `Offset` Decoration
  #### The `CodeLocation` Decoration
  #### The `FlakeAttempts` Decoration

## `ginkgo` CLI Reference
  ### Running Tests
  ### Watching For Changes
  ### Precompiling Tests
  ### Generators
  ### Creating an Outline of Tests
  ### Other Subcommands

## Third-Party Integrations
  ### Sample .travis.yml
  ### Providing a `testing.T`
  ### Using Other Matcher Libraries
  ### Integrating with Gomock
  ### IDE Support

{% endraw  %}
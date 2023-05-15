---
layout: default
title: Ginkgo
---
{% raw  %}
![Ginkgo](./images/ginkgo.png)

[Ginkgo](https://github.com/onsi/ginkgo) is a testing framework for Go designed to help you write expressive tests.  It is best paired with the [Gomega](https://github.com/onsi/gomega) matcher library.  When combined, Ginkgo and Gomega provide a rich and expressive DSL ([Domain-specific Language](https://en.wikipedia.org/wiki/Domain-specific_language)) for writing tests.

Ginkgo is sometimes described as a "Behavior Driven Development" (BDD) framework.  In reality, Ginkgo is a general purpose testing framework in active use across a wide variety of testing contexts: unit tests, integration tests, acceptance test, performance tests, etc.

The narrative docs you are reading here are supplemented by the [godoc](https://pkg.go.dev/github.com/onsi/ginkgo) API-level docs.  We suggest starting here to build a mental model for how Ginkgo works and understand how the Ginkgo DSL can be used to solve real-world testing scenarios.  These docs are written assuming you are familiar with Go and the Go toolchain and that you are using Ginkgo V2 (V1 is no longer supported - see [here](https://onsi.github.io/ginkgo/MIGRATING_TO_V2) for the migration guide).

## Why Ginkgo?

This section captures some of Ginkgo's history and motivation.  If you just want to dive straight in, feel free to [jump ahead](#getting-started)!

Like all software projects, Ginkgo was written at a particular time and place to solve a particular set of problems.

The first commit to Ginkgo was made by [@onsi](https://github.com/onsi/) on August 19th, 2013 (to put that timeframe in perspective, it's roughly three months before [Go 1.2](https://golang.org/doc/go1.2) was released!)  Ginkgo came together in the highly collaborative environment fostered by Pivotal, a software company and consultancy that advocated for outcome-oriented software development built by balanced teams that embrace test-driven development.

Specifically, Pivotal was one of the lead contributors to Cloud Foundry.  A sprawling distributed system, originally written in Ruby, that was slowly migrating towards the emerging distributed systems language of choice: Go.  At the time (and, arguably, to this day) the landscape of Go's testing infrastructure was somewhat anemic.  For engineers coming from the rich ecosystems of testing frameworks such as [Jasmine](https://jasmine.github.io), [rspec](https://rspec.info), and [Cedar](https://github.com/cedarbdd/cedar) there was a need for a comprehensive testing framework with a mature set of matchers in Go.

The need was twofold: organizational and technical. As a growing organization Pivotal would benefit from a shared testing framework to be used across its many teams writing Go.  Engineers jumping from one team to another needed to be able to hit the ground running; we needed fewer testing bikesheds and more shared testing patterns.  And our test-driven development culture put a premium on tests as first-class citizens: they needed to be easy to write, easy to read, and easy to maintain.

Moreover, the _nature_ of the code being built - complex distributed systems - required a testing framework that could provide for the needs unique to unit-testing and integration-testing such a system.  We needed to make testing [asynchronous behavior](https://onsi.github.io/gomega/#making-asynchronous-assertions) ubiquitous and straightforward.  We needed to have [parallelizable integration tests](#spec-parallelization) to ensure our test run-times didn't get out of control.  We needed a test framework that helped us [suss out](#spec-randomization) flaky tests and fix them.

This was the context that led to Ginkgo.  Over the years the Go testing ecosystem has grown and evolved - sometimes [bringing](https://go.dev/blog/subtests) it [closer](https://golang.org/doc/go1.17#testing) to Ginkgo.  Throughout, the community's reactions to Ginkgo have been... interesting.  Some enjoy the expressive framework and rich set of matchers - for many the DSL is familiar and the CLI is productive.  Others have found the DSL off-putting, arguing that Ginkgo is not "the Go way" and that Go developers should eschew third party libraries in general.  That's OK; the world is plenty large enough for options to abound :)

Happy Testing!

## Getting Started

In this section we cover installing Ginkgo, Gomega, and the `ginkgo` CLI.  We bootstrap a Ginkgo suite, write our first spec, and run it.

### Installing Ginkgo

Ginkgo uses [go modules](https://go.dev/blog/using-go-modules).  To add Ginkgo to your project, assuming you have a `go.mod` file setup, just `go install` it:

```bash
go install github.com/onsi/ginkgo/v2/ginkgo
go get github.com/onsi/gomega/...
```

This fetches Ginkgo and installs the `ginkgo` executable under `$GOBIN` - you'll want that on your `$PATH`.  It also fetches the core Gomega matcher library and its set of supporting libraries.  Note that the current supported major version of Ginkgo is `v2`.

You should now be able to run `ginkgo version` at the command line and see the Ginkgo CLI emit a version number.

**Note** you _must_ make sure the version of the `ginkgo` cli you install is the same as the version of Ginkgo in your `go.mod` file.  You can do this by running `go install github.com/onsi/ginkgo/v2/ginkgo` from your package.

#### Upgrading Ginkgo

To upgrade Ginkgo run:

```bash
go get github.com/onsi/ginkgo/v2
go install github.com/onsi/ginkgo/v2/ginkgo
```

To pick a particular version: 

```bash
go get github.com/onsi/ginkgo/v2@v2.m.p
go install github.com/onsi/ginkgo/v2/ginkgo
```

### Support Policy

Ginkgo adheres to semantic versioning - the intent is for there to be no breaking changes along the `2.m.p` line with new functionality landing as minor releases and bug-fixes landing as patch releases (fixes are never back-ported).  We work hard to maintain this policy however exceptions (while rare and typically minor) are possible, especially for brand new/emerging features.

The current version of Ginkgo is guaranteed to be compatible with the currently supported versions of Go that are noted by the Go release policy i.e. N and N-1 major versions.

### Your First Ginkgo Suite

Ginkgo hooks into Go's existing `testing` infrastructure.  That means that Ginkgo specs live in `*_test.go` files, just like standard go tests.  However, instead of using `func TestX(t *testing.T) {}` to write your tests you use the Ginkgo and Gomega DSLs.  

We call a collection of Ginkgo specs in a given package a **Ginkgo suite**; and we use the word **spec** to talk about individual Ginkgo tests contained in the suite.  Though they're functionally interchangeable, we'll use the word "spec" instead of "test" to make a distinction between Ginkgo tests and traditional `testing` tests.

In most Ginkgo suites there is only one `TestX` function - the entry point for Ginkgo.  Let's bootstrap a Ginkgo suite to see what that looks like.

### Bootstrapping a Suite

Say you have a package named `books` that you'd like to add a Ginkgo suite to.  To bootstrap the suite run:

```bash
cd path/to/books
ginkgo bootstrap
Generating ginkgo test suite bootstrap for books in:
  books_suite_test.go
```

This will generate a file named `books_suite_test.go` in the `books` directory containing:

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

First, `ginkgo bootstrap` generates a new test file and places it in the `books_test` package.  That small detail is actually quite important so let's take a brief detour to discuss how Go organizes code in general, and test packages in particular.

#### Mental Model: Go modules, packages, and tests

Go code is organized into [**modules**](https://go.dev/blog/using-go-modules).  A module is typically associated with a version controlled repository and is comprised of a series of versioned **packages**.  Each package is typically associated with a single directory within the module's file tree containing a series of source code files.  When testing Go code, unit tests for a package typically reside within the same directory as the package and are named `*_test.go`.  Ginkgo follows this convention.  It's also possible to construct test-only packages comprised solely of `*_test.go` files.  For example, module-level integration tests typically live in their own test-only package directory and exercise the various packages of the module as a whole.  As Ginkgo simply builds on top of Go's existing test infrastructure, this usecase is supported and encouraged as well.

Normally, Go only allows one package to live in a given directory (in our case, it would be a package named `books`).  There is, however, one exception to this rule: a package ending in `_test` is allowed to live in the same directory as the package being tested.  Doing so instructs Go to compile the package's test suite as a **separate package**.  This means your test suite will **not** have access to the internals of the `books` package and will need to `import` the `books` package to access its external interface.  Ginkgo defaults to setting up the suite as a `*_test` package to encourage you to only test the external behavior of your package, not its internal implementation details.

OK back to our bootstrap file.  After the `package books_test` declaration we import the `ginkgo` and `gomega` packages into the test's top-level namespace by performing a `.` dot-import.  Since Ginkgo and Gomega are DSLs this makes the tests more natural to read.  If you prefer, you can avoid the dot-import via `ginkgo bootstrap --nodot`.  Throughout this documentation we'll assume dot-imports.

Next we define a single `testing` test: `func TestBooks(t *testing.T)`.  This is the entry point for Ginkgo - the go test runner will run this function when you run `go test` or `ginkgo`.

Inside the `TestBooks` function are two lines:

`RegisterFailHandler(Fail)` is the single line of glue code connecting Ginkgo to Gomega.  If we were to avoid dot-imports this would read as `gomega.RegisterFailHandler(ginkgo.Fail)` - what we're doing here is telling our matcher library (Gomega) which function to call (Ginkgo's `Fail`) in the event a failure is detected.

Finally the `RunSpecs()` call tells Ginkgo to start the test suite, passing it the `*testing.T` instance and a description of the suite.  You should only ever call `RunSpecs` once and you can let Ginkgo worry about calling `*testing.T` for you.

With the bootstrap file in place, you can now run your suite using the `ginkgo` command:

```bash
ginkgo

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

Under the hood, `ginkgo` is simply calling `go test`.  While you _can_ run `go test` instead of the `ginkgo` CLI, Ginkgo has several capabilities that can only be accessed via `ginkgo`.  We generally recommend users embrace the `ginkgo` CLI and treat it as a first-class member of their testing toolchain.

Alright, we've successfully set up and run our first suite.  Of course that suite is empty, which isn't very interesting.  Let's add some specs.

#### Adding Specs to a Suite
While you can add all your specs directly into `books_suite_test.go` you'll generally prefer to place your specs in separate files.  This is particularly true if you have packages with multiple files that need to be tested.  Let's say our `book` package includes a `book.go` model and we'd like to test its behavior.  We can generate a test file like so:

```bash
ginkgo generate book
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

var _ = Describe("Books", func() {

})
```

As with the bootstrapped suite file, this test file is in the separate `books_test` package and dot-imports both `ginkgo` and `gomega`.  Since we're testing the external interface of `books` Ginkgo adds an `import` statement to pull the `books` package into the test.

Ginkgo then adds an empty top-level `Describe` container node.  `Describe` is part of the Ginkgo DSL and takes a description and a closure function. `Describe("book", func() { })` generates a container that will contain specs that describe the behavior of `"Books"`.

> By default, Go does not allow you to invoke bare functions at the top-level of a file.  Ginkgo gets around this by having its node DSL functions return a value that is intended to be discarded.  This allows us to write `var _ = Describe(...)` at the top-level which satisfies Go's top-level policies.

Let's add a few specs, now, to describe our book model's ability to categorize books:

```go
var _ = Describe("Books", func() {
  var foxInSocks, lesMis *books.Book

  BeforeEach(func() {
    lesMis = &books.Book{
      Title:  "Les Miserables",
      Author: "Victor Hugo",
      Pages:  2783,
    }

    foxInSocks = &books.Book{
      Title:  "Fox In Socks",
      Author: "Dr. Seuss",
      Pages:  24,
    }
  })

  Describe("Categorizing books", func() {
    Context("with more than 300 pages", func() {
      It("should be a novel", func() {
        Expect(lesMis.Category()).To(Equal(books.CategoryNovel))
      })
    })

    Context("with fewer than 300 pages", func() {
      It("should be a short story", func() {
        Expect(foxInSocks.Category()).To(Equal(books.CategoryShortStory))
      })
    })
  })
})
```

There's a lot going on here so let's break it down.

Ginkgo makes extensive use of closures to allow us to build a descriptive spec hierarchy.  This hierarchy is constructed using three kinds of **nodes**:

We use **container nodes** like `Describe` and `Context` to organize the different aspects of code that we are testing hierarchically.  In this case we are describing our book model's ability to categorize books across two different contexts - the behavior for large books `"With more than 300 pages"` and small books `"With fewer than 300 pages"`.

We use **setup nodes** like `BeforeEach` to set up the state of our specs.  In this case, we are instantiating two new book models: `lesMis` and `foxInSocks`.

Finally, we use **subject nodes** like `It` to write a spec that makes assertions about the subject under test.  In this case, we are ensuring that `book.Category()` returns the correct category `enum` based on the length of the book.  We make these assertions using Gomega's `Equal` matcher and `Expect` syntax.  You can learn much more about [Gomega here](https://onsi.github.io/gomega/#making-assertions) - the examples used throughout these docs should be self-explanatory.

The container, setup, and subject nodes form a **spec tree**.  Ginkgo uses the tree to construct a flattened list of specs where each spec can have multiple setup nodes but will only have one subject node.

Because there are two subject nodes, Ginkgo will identify two specs to run.  For each spec, Ginkgo will run the closures attached to any associated setup nodes and then run the closure attached to the subject node.  In order to share state between the setup node and subject node we define closure variables like `lesMis` and `foxInSocks`.  This is a common pattern and the main way that tests are organized in Ginkgo.

Assuming a `book.Book` model with this behavior we can run the tests:

```bash
ginkgo
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

Success!  We've written and run our first Ginkgo suite.  From here we can grow our test suite as we iterate on our code.

The next sections will delve into the many mechanisms Ginkgo provides for writing and running specs.

## Writing Specs

Ginkgo makes it easy to write expressive specs that describe the behavior of your code in an organized manner.  We've seen that Ginkgo suites are hierarchical collections of specs comprised of container nodes, setup nodes, and subject nodes organized into a spec tree.  In this section we dig into the various kinds of nodes available in Ginkgo and their properties.

### Spec Subjects: It
Every Ginkgo spec has exactly one subject node.  You can add a single spec to a suite by adding a new subject node using `It(<description>, <closure>)`.  Here's a spec to validate that we can extract the author's last name from a `Book` model:

```go
var _ = Describe("Books", func() {
  It("can extract the author's last name", func() {
    book = &books.Book{
      Title: "Les Miserables",
      Author: "Victor Hugo",
      Pages: 2783,
    }

    Expect(book.AuthorLastName()).To(Equal("Hugo"))
  })
})
```

As you can see, the description documents the intent of the spec while the closure includes assertions about our code's behavior.

We can add multiple specs to a `Describe` container:

```go
var _ = Describe("Books", func() {
  It("can extract the author's last name", func() {
    book = &books.Book{
      Title: "Les Miserables",
      Author: "Victor Hugo",
      Pages: 2783,
    }

    Expect(book.AuthorLastName()).To(Equal("Hugo"))
  })

  It("can fetch a summary of the book from the library service", func(ctx SpecContext) {
    book = &books.Book{
      Title: "Les Miserables",
      Author: "Victor Hugo",
      Pages: 2783,
    }

    summary, err := library.FetchSummary(ctx, book)
    Expect(err).NotTo(HaveOccurred())
    Expect(summary).To(ContainSubstring("Jean Valjean"))
  }, SpecTimeout(time.Second))
})
```

Our new spec connects with a library service to fetch a summary of the book and asserts that the request succeeds with a meaningful response.  This example previews a few advanced concepts that you'll learn about later in these docs: Ginkgo supports [decorators](#mental-model-spec-decorators) like [SpecTimeout](#the-spectimeout-and-nodetimeout-decorators) to annotate and modify the behavior of specs; and Ginkgo allows you to test potentially long-running code by writing [interruptible](#spec-timeouts-and-interruptible-nodes) specs that accept a `SpecContext` or `context.Context`.  Now, if more than a second elapses, _or_ an interrupt signal is received, Ginkgo will signal `library.FetchSummary` to clean up by cancelling `ctx`.

Ginkgo provides an alias for `It` called `Specify`.  `Specify` is functionally identical to `It` but can help your specs read more naturally.

### Extracting Common Setup: BeforeEach
You can remove duplication and share common setup across specs using `BeforeEach(<closure>)` setup nodes.  Let's add specs to our `Book` suite that cover extracting the author's first name and a few natural edge cases:

```go
var _ = Describe("Books", func() {
  var book *books.Book

  BeforeEach(func() {
    book = &books.Book{
      Title: "Les Miserables",
      Author: "Victor Hugo",
      Pages: 2783,
    }
    Expect(book.IsValid()).To(BeTrue())
  })

  It("can extract the author's last name", func() {
    Expect(book.AuthorLastName()).To(Equal("Hugo"))
  })

  It("interprets a single author name as a last name", func() {
    book.Author = "Hugo"
    Expect(book.AuthorLastName()).To(Equal("Hugo"))
  })

  It("can extract the author's first name", func() {
    Expect(book.AuthorFirstName()).To(Equal("Victor"))
  })

  It("returns no first name when there is a single author name", func() {
    book.Author = "Hugo"
    Expect(book.AuthorFirstName()).To(BeZero()) //BeZero asserts the value is the zero-value for its type.  In this case: ""
  })
})
```

We now have four subject nodes so Ginkgo will run four specs.  The common setup for each spec is captured in the `BeforeEach` node.  When running each spec Ginkgo will first run the `BeforeEach` closure and then the subject closure.

Information is shared between closures via closure variables. It is idiomatic for these closure variables to be declared within the container node closure and initialized in the setup node closure.  Doing so ensures that each spec has a pristine, correctly initialized, copy of the shared variable.

In this example, the `single author name` specs _mutate_ the shared `book` closure variable.  These mutations do not pollute the other specs because the `BeforeEach` closure reinitializes `book`.

This detail is really important.  Ginkgo requires, by default, that specs be fully **independent**.  This allows Ginkgo to shuffle the order of specs and run specs in parallel.  We'll cover this in more detail later on but for now embrace this takeaway: **"Declare in container nodes, initialize in setup nodes"**.

One last point - Ginkgo allows assertions to be made in both setup nodes and subject nodes.  In fact, it's a common pattern to make assertions in setup nodes to validate that the spec setup is correct _before_ making behavioral assertions in subject nodes.  In our (admittedly contrived) example here, we are asserting that the `book` we've instantiated is valid with `Expect(book.IsValid()).To(BeTrue())`.

### Organizing Specs With Container Nodes
Ginkgo allows you to hierarchically organize the specs in your suite using container nodes.  Ginkgo provides three synonymous nouns for creating container nodes: `Describe`, `Context`, and `When`.  These three are functionally identical and are provided to help the spec narrative flow.  You usually `Describe` different capabilities of your code and explore the behavior of each capability across different `Context`s.

Our `book` suite is getting longer and would benefit from some hierarchical organization.  Let's organize what we have so far using container nodes:

```go
var _ = Describe("Books", func() {
  var book *books.Book

  BeforeEach(func() {
    book = &books.Book{
      Title: "Les Miserables",
      Author: "Victor Hugo",
      Pages: 2783,
    }
    Expect(book.IsValid()).To(BeTrue())
  })

  Describe("Extracting the author's first and last name", func() {
    Context("When the author has both names", func() {
      It("can extract the author's last name", func() {        
        Expect(book.AuthorLastName()).To(Equal("Hugo"))
      })

      It("can extract the author's first name", func() {
        Expect(book.AuthorFirstName()).To(Equal("Victor"))
      })      
    })

    Context("When the author only has one name", func() {
      BeforeEach(func() {
        book.Author = "Hugo"
      })  

      It("interprets the single author name as a last name", func() {
        Expect(book.AuthorLastName()).To(Equal("Hugo"))
      })

      It("returns empty for the first name", func() {
        Expect(book.AuthorFirstName()).To(BeZero())
      })
    })

  })
})
```

Using container nodes helps clarify the intent behind our suite.  The author name specs are now clearly grouped together and we're exploring the behavior of our code in different contexts.  Most importantly, we're able to scope additional setup nodes to those contexts to refine our spec setup.

When Ginkgo runs a spec it runs through all the `BeforeEach` closures that appear in that spec's hierarchy from the outer-most to the inner-most.  If multiple `BeforeEach` nodes appear at the same nesting level they will be run in the order in which they appear in the test file.  For the `both names` specs, Ginkgo will run the outermost `BeforeEach` closure before the subject node closure.  For the `one name` specs, Ginkgo will run the outermost `BeforeEach` closure and then the innermost `BeforeEach` closure which sets `book.Author = "Hugo"`.

Organizing our specs in this way can also help us reason about our spec coverage.  What additional contexts are we missing?  What edge cases should we worry about?  Let's add a few:

```go
var _ = Describe("Books", func() {
  var book *books.Book

  BeforeEach(func() {
    book = &books.Book{
      Title: "Les Miserables",
      Author: "Victor Hugo",
      Pages: 2783,
    }
    Expect(book.IsValid()).To(BeTrue())
  })

  Describe("Extracting the author's first and last name", func() {
    Context("When the author has both names", func() {
      It("can extract the author's last name", func() {        
        Expect(book.AuthorLastName()).To(Equal("Hugo"))
      })

      It("can extract the author's first name", func() {
        Expect(book.AuthorFirstName()).To(Equal("Victor"))
      })      
    })

    Context("When the author only has one name", func() {
      BeforeEach(func() {
        book.Author = "Hugo"
      })  

      It("interprets the single author name as a last name", func() {
        Expect(book.AuthorLastName()).To(Equal("Hugo"))
      })

      It("returns empty for the first name", func() {
        Expect(book.AuthorFirstName()).To(BeZero())
      })
    })

    Context("When the author has a middle name", func() {
      BeforeEach(func() {
        book.Author = "Victor Marie Hugo"
      })  

      It("can extract the author's last name", func() {        
        Expect(book.AuthorLastName()).To(Equal("Hugo"))
      })

      It("can extract the author's first name", func() {
        Expect(book.AuthorFirstName()).To(Equal("Victor"))
      })      
    })

    Context("When the author has no name", func() {
      It("should not be a valid book and returns empty for first and last name", func() {
        book.Author = ""
        Expect(book.IsValid()).To(BeFalse())
        Expect(book.AuthorLastName()).To(BeZero())
        Expect(book.AuthorFirstName()).To(BeZero())
      })
    })
  })
})
```

That should cover most edge cases.  As you can see we have flexibility in how we structure our specs.  Some developers prefer single assertions in `It` nodes where possible.  Others prefer consolidating multiple assertions into a single `It` as we do in the `no name` context.  Both approaches are supported and perfectly reasonable.

Let's keep going and add spec out some additional behavior.  Let's test how our `book` model handles JSON encoding/decoding.  Since we're describing new behavior we'll add a new `Describe` container node:


```go
var _ = Describe("Books", func() {
  var book *books.Book

  BeforeEach(func() {
    book = &books.Book{
      Title: "Les Miserables",
      Author: "Victor Hugo",
      Pages: 2783,
    }
    Expect(book.IsValid()).To(BeTrue())
  })

  Describe("Extracting the author's first and last name", func() { ... })

  Describe("JSON encoding and decoding", func() {
    It("survives the round trip", func() {
      encoded, err := book.AsJSON()
      Expect(err).NotTo(HaveOccurred())

      decoded, err := books.NewBookFromJSON(encoded)
      Expect(err).NotTo(HaveOccurred())

      Expect(decoded).To(Equal(book))
    })

    Describe("some JSON decoding edge cases", func() {
      var err error

      When("the JSON fails to parse", func() {
        BeforeEach(func() {
          book, err = NewBookFromJSON(`{
            "title":"Les Miserables",
            "author":"Victor Hugo",
            "pages":2783oops
          }`)
        })

        It("returns a nil book", func() {
          Expect(book).To(BeNil())
        })

        It("errors", func() {
          Expect(err).To(MatchError(books.ErrInvalidJSON))
        })
      })

      When("the JSON is incomplete", func() {
        BeforeEach(func() {
          book, err = NewBookFromJSON(`{
            "title":"Les Miserables",
            "author":"Victor Hugo",
          }`)
        })

        It("returns a nil book", func() {
          Expect(book).To(BeNil())
        })

        It("errors", func() {
          Expect(err).To(MatchError(books.ErrIncompleteJSON))
        })
      })      
    })
  })
})
```

In this way we can continue to grow our suite while clearly delineating the structure of our specs using a spec tree hierarchy.  Note that we use the `When` container variant in this example as it reads cleanly.  Remember that `Describe`, `Context`, and `When` are functionally equivalent aliases.

### Mental Model: How Ginkgo Traverses the Spec Hierarchy 

We've delved into the three basic Ginkgo node types: container nodes, setup nodes, and subject nodes.  Before we move on let's build a mental model for how Ginkgo traverses and runs specs in a little more detail.

When Ginkgo runs a suite it does so in _two phases_.  The **Tree Construction Phase** followed by the **Run Phase**.

During the Tree Construction Phase Ginkgo enters all container nodes by invoking their closures to construct the spec tree.  During this phase Ginkgo is capturing and saving off the various setup and subject node closures it encounters in the tree _without running them_.  Only container node closures run during this phase and Ginkgo does not expect to encounter any assertions as no specs are running yet.

Let's paint a picture of what that looks like in practice.  Consider the following set of book specs:

```go
var _ = Describe("Books", func() {
  var book *books.Book

  BeforeEach(func() {
    //Closure A
    book = &books.Book{
      Title: "Les Miserables",
      Author: "Victor Hugo",
      Pages: 2783,
    }
    Expect(book.IsValid()).To(BeTrue())
  })

  Describe("Extracting names", func() {
    When("author has both names", func() {
      It("extracts the last name", func() {        
        //Closure B
        Expect(book.AuthorLastName()).To(Equal("Hugo"))
      })

      It("extracts the first name", func() {
        //Closure C
        Expect(book.AuthorFirstName()).To(Equal("Victor"))
      })      
    })

    When("author has one name", func() {
      BeforeEach(func() {
        //Closure D
        book.Author = "Hugo"
      })  

      It("extracts the last name", func() {
        //Closure E
        Expect(book.AuthorLastName()).To(Equal("Hugo"))
      })

      It("returns empty first name", func() {
        //Closure F
        Expect(book.AuthorFirstName()).To(BeZero())
      })
    })

  })
})
```

We could represent the spec tree that Ginkgo generates as follows:

```
Describe: "Books"
  |_BeforeEach: <Closure-A>
  |_Describe: "Extracting names"
    |_When: "author has both names"
      |_It: "extracts the last name", <Closure-B>
      |_It: "extracts the first name", <Closure-C>
    |_When: "author has one name"
      |_BeforeEach: <Closure-D>
      |_It: "extracts the last name", <Closure-E>
      |_It: "returns empty first name", <Closure-F>
```

Note that Ginkgo is saving off just the setup and subject node closures.

Once the spec tree is constructed Ginkgo walks the tree to generate a flattened list of specs.  For our example, the resulting spec list would look a bit like:

```
{
  Texts: ["Books", "Extracting names", "author has both names", "extracts the last name"],
  Closures: <BeforeEach-Closure-A>, <It-Closure-B>
},
{
  Texts: ["Books", "Extracting names", "author has both names", "extracts the first name"],
  Closures: <BeforeEach-Closure-A>, <It-Closure-C>
},
{
  Texts: ["Books", "Extracting names", "author has one name", "extracts the last name"],
  Closures: <BeforeEach-Closure-A>, <BeforeEach-Closure-D>, <It-Closure-E>
},
{
  Texts: ["Books", "Extracting names", "author has one name", "returns empty first name"],
  Closures: <BeforeEach-Closure-A>, <BeforeEach-Closure-D>, <It-Closure-F>
}
```

As you can see each generated spec has exactly one subject node, and the appropriate set of setup nodes.  During the Run Phase Ginkgo runs through each spec in the spec list sequentially.  When running a spec Ginkgo invokes the setup and subject nodes closures in the correct order and tracks any failed assertions.  Note that container node closures are _never_ invoked during the run phase.

Given this mental model, here are a few common gotchas to avoid:

#### Nodes only belong in Container Nodes

Since the spec tree is constructed by traversing container nodes all Ginkgo nodes **must** only appear at the top-level of the suite _or_ nested within a container node.  They cannot appear within a subject node or setup node.  The following is invalid:

```go
/* === INVALID === */
var _ = It("has a color", func() {
  Context("when blue", func() { // NO! Nodes can only be nested in containers
    It("is blue", func() { // NO! Nodes can only be nested in containers

    })
  })
})
```

Ginkgo will emit a warning if it detects this.

#### No Assertions in Container Nodes

Because container nodes are invoked to construct the spec tree, but never when running specs, assertions _must_ be in subject nodes or setup nodes.  Never in container nodes.  The following is invalid:

```go
/* === INVALID === */
var _ = Describe("book", func() {
  var book *Book
  Expect(book.Title()).To(BeFalse()) // NO!  Place in a setup node instead.

  It("tests something", func() {...})
})
```

Ginkgo will emit a warning if it detects this.

#### Avoid Spec Pollution: Don't Initialize Variables in Container Nodes

We've covered this already but it bears repeating: **"Declare in container nodes, initialize in setup nodes"**.  Since container nodes are only invoked once during the tree construction phase you should declare closure variables in container nodes but always initialize them in setup nodes.  The following is 
invalid can potentially infuriating to debug:

```go
/* === INVALID === */
var _ = Describe("book", func() {
  book := &books.Book{ // No!
    Title:  "Les Miserables",
    Author: "Victor Hugo",
    Pages:  2783,
  }

  It("is invalid with no author", func() {
    book.Author = "" // bam! we've changed the closure variable and it will never be reset.
    Expect(book.IsValid()).To(BeFalse())
  })

  It("is valid with an author", func() {
    Expect(book.IsValid()).To(BeTrue()) // this will fail if it runs after the previous test
  })
})
```

you should do this instead:

```go
var _ = Describe("book", func() {
  var book *books.Book // declare in container nodes

  BeforeEach(func() {
    book = &books.Book {  //initialize in setup nodes
      Title:  "Les Miserables",
      Author: "Victor Hugo",
      Pages:  2783,
    }    
  })

  It("is invalid with no author", func() {
    book.Author = ""
    Expect(book.IsValid()).To(BeFalse())
  })

  It("is valid with an author", func() {
    Expect(book.IsValid()).To(BeTrue())
  })
})

```

Ginkgo currently has no mechanism in place to detect this failure mode, you'll need to stick to "declare in container nodes, initialize in setup nodes" to avoid spec pollution.

### Separating Creation and Configuration: JustBeforeEach

Let's get back to our growing Book suite and explore a few more Ginkgo nodes.  So far we've met the `BeforeEach` setup node, let's introduce its closely related cousin: `JustBeforeEach`.

`JustBeforeEach` is intended to solve a very specific problem but should be used with care as it can add complexity to a test suite.  Consider the following section of our JSON decoding book tests:

```go
Describe("some JSON decoding edge cases", func() {
  var book *books.Book
  var err error

  When("the JSON fails to parse", func() {
    BeforeEach(func() {
      book, err = NewBookFromJSON(`{
        "title":"Les Miserables",
        "author":"Victor Hugo",
        "pages":2783oops
      }`)
    })

    It("returns a nil book", func() {
      Expect(book).To(BeNil())
    })

    It("errors", func() {
      Expect(err).To(MatchError(books.ErrInvalidJSON))
    })
  })

  When("the JSON is incomplete", func() {
    BeforeEach(func() {
      book, err = NewBookFromJSON(`{
        "title":"Les Miserables",
        "author":"Victor Hugo",
      }`)
    })

    It("returns a nil book", func() {
      Expect(book).To(BeNil())
    })

    It("errors", func() {
      Expect(err).To(MatchError(books.ErrIncompleteJSON))
    })
  })      
})
```

In each case we're creating a new `book` from an invalid snippet of JSON, ensuring the `book` is `nil` and checking that the correct error was returned.  There's some degree of deduplication that could be attained here.  We could try to pull out a shared `BeforeEach` like so:

```go
/* === INVALID === */
Describe("some JSON decoding edge cases", func() {
  var book *books.Book
  var err error
  BeforeEach(func() {
    book, err = NewBookFromJSON(???)
    Expect(book).To(BeNil())
  })

  When("the JSON fails to parse", func() {
    It("errors", func() {
      Expect(err).To(MatchError(books.ErrInvalidJSON))
    })
  })

  When("the JSON is incomplete", func() {
    It("errors", func() {
      Expect(err).To(MatchError(books.ErrIncompleteJSON))
    })
  })      
})
```

but there's no way using `BeforeEach` and `It` nodes to configure the json we use to create the book differently for each `When` container _before_ we invoke `NewBookFromJSON`.  That's where `JustBeforeEach` comes in.  As the name suggests, `JustBeforeEach` nodes run _just before_ the subject node but _after_ any other `BeforeEach` nodes.  We can leverage this behavior to write:

```go
Describe("some JSON decoding edge cases", func() {
  var book *books.Book
  var err error
  var json string
  JustBeforeEach(func() {
    book, err = NewBookFromJSON(json)
    Expect(book).To(BeNil())
  })

  When("the JSON fails to parse", func() {
    BeforeEach(func() {
      json = `{
        "title":"Les Miserables",
        "author":"Victor Hugo",
        "pages":2783oops
      }`
    })

    It("errors", func() {
      Expect(err).To(MatchError(books.ErrInvalidJSON))
    })
  })

  When("the JSON is incomplete", func() {
    BeforeEach(func() {
      json = `{
        "title":"Les Miserables",
        "author":"Victor Hugo",
      }`
    })
    
    It("errors", func() {
      Expect(err).To(MatchError(books.ErrIncompleteJSON))
    })
  })      
})
```

When Ginkgo runs these specs it will _first_ run the `BeforeEach` setup closures, thereby population the `json` variable, and _then_ run the `JustBeforeEach` setup closure, thereby decoding the correct JSON string.

Abstractly, `JustBeforeEach` allows you to decouple **creation** from **configuration**.  Creation occurs in the `JustBeforeEach` using configuration specified and modified by a chain of `BeforeEach`s.

As with `BeforeEach` you can have multiple `JustBeforeEach` nodes at different levels of container nesting.  Ginkgo will first run all the `BeforeEach` closures from the outside in, then all the `JustBeforeEach` closures from the outside in.  While powerful and flexible overuse of `JustBeforeEach` (and nest `JustBeforeEach`es in particular!) can lead to confusing suites to be sure to use `JustBeforeEach` judiciously!

### Spec Cleanup: AfterEach and DeferCleanup

The setup nodes we've seen so far all run _before_ the spec's subject closure.  Ginkgo also provides setup nodes that run _after_ the spec's subject: `AfterEach` and `JustAfterEach`.  These are used to clean up after specs and can be particularly helpful in complex integration suites where some external system must be restored to its original state after each spec.

Here's a simple (if contrived!) example to get us started.  Let's suspend disbelief and imagine that our `book` model tracks the weight of books... and that the units used to display the weight can be specified with an environment variable.  Let's spec this out:

```go
Describe("Reporting book weight", func() {
  var book *books.Book

  BeforeEach(func() {
    book = &books.Book{
      Title: "Les Miserables",
      Author: "Victor Hugo",
      Pages: 2783,
      Weight: 500,
    }
  })

  Context("with no WEIGHT_UNITS environment set", func() {
    BeforeEach(func() {
      err := os.Clearenv("WEIGHT_UNITS")
      Expect(err).NotTo(HaveOccurred())
    })

    It("reports the weight in grams", func() {
      Expect(book.HumanReadableWeight()).To(Equal("500g"))
    })
  })

  Context("when WEIGHT_UNITS is set to oz", func() {
    BeforeEach(func() {
      err := os.Setenv("WEIGHT_UNITS", "oz")      
      Expect(err).NotTo(HaveOccurred())
    })

    It("reports the weight in ounces", func() {
      Expect(book.HumanReadableWeight()).To(Equal("17.6oz"))
    })
  })

  Context("when WEIGHT_UNITS is invalid", func() {
    BeforeEach(func() {
      err := os.Setenv("WEIGHT_UNITS", "smoots")
      Expect(err).NotTo(HaveOccurred())
    })

    It("errors", func() {
      weight, err := book.HumanReadableWeight()
      Expect(weight).To(BeZero())
      Expect(err).To(HaveOccurred())
    })
  })
})
```

These specs are... _OK_.  But we've got a subtle issue: we're not cleaning up when we override the value of `WEIGHT_UNITS`.  This is an example of spec pollution and can lead to subtle failures in unrelated specs.

Let's fix this up using an `AfterEach`:

```go
Describe("Reporting book weight", func() {
  var book *books.Book

  BeforeEach(func() {
    book = &books.Book{
      Title: "Les Miserables",
      Author: "Victor Hugo",
      Pages: 2783,
      Weight: 500,
    }
  })

  AfterEach(func() {
    err := os.Clearenv("WEIGHT_UNITS")
    Expect(err).NotTo(HaveOccurred())
  })

  Context("with no WEIGHT_UNITS environment set", func() {
    BeforeEach(func() {
      err := os.Clearenv("WEIGHT_UNITS")
      Expect(err).NotTo(HaveOccurred())
    })

    It("reports the weight in grams", func() {
      Expect(book.HumanReadableWeight()).To(Equal("500g"))
    })
  })

  Context("when WEIGHT_UNITS is set to oz", func() {
    BeforeEach(func() {
      err := os.Setenv("WEIGHT_UNITS", "oz")      
      Expect(err).NotTo(HaveOccurred())
    })

    It("reports the weight in ounces", func() {
      Expect(book.HumanReadableWeight()).To(Equal("17.6oz"))
    })
  })

  Context("when WEIGHT_UNITS is invalid", func() {
    BeforeEach(func() {
      err := os.Setenv("WEIGHT_UNITS", "smoots")
      Expect(err).NotTo(HaveOccurred())
    })

    It("errors", func() {
      weight, err := book.HumanReadableWeight()
      Expect(weight).To(BeZero())
      Expect(err).To(HaveOccurred())
    })
  })
})
```

Now we're guaranteed to clear out `WEIGHT_UNITS` after each spec as Ginkgo will run the `AfterEach` node's closure after the subject node for each spec... 

...but we've still got a subtle issue.  By clearing it out in our `AfterEach` we're assuming that `WEIGHT_UNITS` is not set when the specs run.  But perhaps it is?  What we really want to do is restore `WEIGHT_UNITS` to its original value.  We can solve this by recording the original value first:

```go
Describe("Reporting book weight", func() {
  var book *books.Book
  var originalWeightUnits string

  BeforeEach(func() {
    book = &books.Book{
      Title: "Les Miserables",
      Author: "Victor Hugo",
      Pages: 2783,
      Weight: 500,
    }
    originalWeightUnits = os.Getenv("WEIGHT_UNITS")
  })

  AfterEach(func() {
    err := os.Setenv("WEIGHT_UNITS", originalWeightUnits)
    Expect(err).NotTo(HaveOccurred())
  })
  ...
})
```

That's better.  The specs will now clean up after themselves correctly.

One quick note before we move on, you may have caught that `book.HumanReadableWeight()` returns _two_ values - a `weight` string and an `error`.  This is common pattern and Gomega has first class support for it.  The assertion:

```go
Expect(book.HumanReadableWeight()).To(Equal("17.6oz"))
```

is actually making two assertions under the hood.  That the first value returned by `book.HumanReadableWeight` is equal to `"17.6oz"` and that any subsequent values (i.e. the returned `error`) are `nil`.  This elegantly inlines error handling and can make your specs more readable.

#### Cleaning up our Cleanup code: DeferCleanup

Setup and cleanup patterns like the one above are common in Ginkgo suites.  While powerful, however, `AfterEach` nodes have a tendency to separate cleanup code from setup code.  We had to create an `originalWeightUnits` closure variable to keep track of the original environment variable in the `BeforeEach` and pass it to the `AfterEach` - this feels noisy and potential error-prone.

Ginkgo provides the `DeferCleanup()` function to help solve for this usecase and bring spec setup closer to spec cleanup.  Here's what our example looks like with `DeferCleanup()`:

```go
Describe("Reporting book weight", func() {
  var book *books.Book

  BeforeEach(func() {
    ...
    originalWeightUnits := os.Getenv("WEIGHT_UNITS")
    DeferCleanup(func() {      
      err := os.Setenv("WEIGHT_UNITS", originalWeightUnits)
      Expect(err).NotTo(HaveOccurred())
    })
  })
  ...
})
```

As you can see, `DeferCleanup()` can be called inside any setup or subject nodes.  This allows us to bring our intended cleanup closer to our setup code and avoid extracting a separate closure variable.  At first glance this code might seem confusing - as we discussed [above](#nodes-only-belong-in-container-nodes) Ginkgo does not allow you to define nodes within setup or subject nodes.  `DeferCleanup` is not a Ginkgo node, however, but rather a convenience function that knows how to track cleanup code and run it at the right time in the spec's lifecycle.

> Under the hood `DeferCleanup` is generating a dynamic `AfterEach` node and adding it to the running spec.  This detail isn't important - you can simply assume that code in `DeferCleanup` has the identical runtime semantics to code in an `AfterEach`.

`DeferCleanup` has a few more tricks up its sleeve.

As shown above `DeferCleanup` can be passed a function that takes no arguments and returns no value.  You can also pass a function that returns values.  `DeferCleanup` ignores all these return value except for the last.  If the last return value is a non-nil error - a common go pattern - `DeferCleanup` will fail the spec.  This allows us to rewrite our example as:

```go
Describe("Reporting book weight", func() {
  var book *books.Book

  BeforeEach(func() {
    ...
    originalWeightUnits := os.Getenv("WEIGHT_UNITS")
    DeferCleanup(func() error {      
      return os.Setenv("WEIGHT_UNITS", originalWeightUnits)
    })
  })
  ...
})
```

You can also pass in a function that accepts arguments, then pass those arguments in directly to `DeferCleanup`. These arguments will be captured and passed to the function when cleanup is invoked.  This allows us to rewrite our example once more as:

```go
Describe("Reporting book weight", func() {
  var book *books.Book

  BeforeEach(func() {
    ...
    DeferCleanup(os.Setenv, "WEIGHT_UNITS", os.Getenv("WEIGHT_UNITS"))
  })
  ...
})
```

here `DeferCleanup` is capturing the original value of `WEIGHT_UNITS` as returned by `os.Getenv("WEIGHT_UNITS")` then passing both it into `os.Setenv` when cleanup is triggered after each spec and asserting that the error returned by `os.Setenv` is `nil`.  We've reduced our cleanup code to a single line!

#### Separating Diagnostics Collection and Teardown: JustAfterEach

We haven't discussed it but Ginkgo also provides a `JustAfterEach` setup node.  `JustAfterEach` closures runs _just after_ the subject node and before any `AfterEach` closures.  This can be useful if you need to collect diagnostic information about your spec _before_ invoking the clean up code in `AfterEach`.  Here's a quick example:

```go
Describe("Saving books to a database", func() {
  AfterEach(func() {
    dbClient.Clear() //clear out the database between tests
  })

  JustAfterEach(func() {
    if CurrentSpecReport().Failed() {
      AddReportEntry("db-dump", dbClient.Dump())
    }
  })

  It("saves the book", func() {
    err := dbClient.Save(book)
    Expect(err).NotTo(HaveOccurred())
  })

})
```

We're, admittedly, jumping ahead a bit here by introducing a few new concepts that we'll dig into more later.  The `JustAfterEach` closure in this container will always run after the subject closure but before the `AfterEach` closure.  When it runs it will check if the current spec has failed (`CurrentSpecReport().Failed()`) and, if a failure was detected, it will download a dump of the database `dbClient.Dump()` and attach it to the spec's report `AddReportEntry()`.  It's important that this runs before the `dbClient.Clear()` invocation in `AfterEach` - so we use a `JustAfterEach`.  Of course, we could have inlined this diagnostic behavior into our `AfterEach`.

As with `JustBeforeEach`, `JustAfterEach` can be nested in multiple containers.  Doing so can have powerful results but might lead to confusing test suites -- so use nested `JustAfterEach`es judiciously.

### Suite Setup and Cleanup: BeforeSuite and AfterSuite

The setup nodes we've explored so far have all applied at the spec level.  They run Before**Each** or After**Each** spec in their associated container node.

It is common, however, to need to perform setup and cleanup at the level of the Ginkgo suite.  This is setup that should be performed just once - before any specs run, and cleanup that should be performed just once, when all the specs have finished.  Such code is particularly common in integration tests that need to prepare environments or spin up external resources.

Ginkgo supports suite-level setup and cleanup through two specialized **suite setup** nodes: `BeforeSuite` and `AfterSuite`.  These suite setup nodes **must** be called at the top-level of the suite and cannot be nested in containers.  Also there can be at most one `BeforeSuite` node and one `AfterSuite` node per suite.  It is idiomatic to place the suite setup nodes in the Ginkgo bootstrap suite file.

Let's continue to build out our book tests.  Books can be stored and retrieved from an external database and we'd like to test this behavior.  To do that, we'll need to spin up a database and set up a client to access it.  We can do that `BeforeEach` spec - but doing so would be prohibitively expensive and slow.  Instead, it would be more efficient to spin up the database just once when the suite starts.  Here's how we'd do it in our `books_suite_test.go` file:

```go
package books_test

import (
  . "github.com/onsi/ginkgo/v2"
  . "github.com/onsi/gomega"

  "path/to/db"

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
  Expect(dbRunner.Start()).To(Succeed())

  dbClient = db.NewClient()
  Expect(dbClient.Connect(dbRunner.Address())).To(Succeed())
})

var _ = AfterSuite(func() {
  Expect(dbRunner.Stop()).To(Succeed())
})

var _ = AfterEach(func() {
   Expect(dbClient.Clear()).To(Succeed())
})
```

Ginkgo will run our `BeforeSuite` closure at the beginning of the [run phase](#mental-model-how-ginkgo-traverses-the-spec-hierarchy) - i.e. after the spec tree has been constructed but before any specs have run.  This closure will instantiate a new `*db.Runner` - this is hypothetical code that knows how to spin up an instance of a database - and ask the runner to `Start()` a database.

It will then instantiate a `*db.Client` and connect it to the database.  Since `dbRunner` and `dbClient` are closure variables defined at the top-level all specs in our suite will have access to them and can trust that they have been correctly initialized.

Our specs will be manipulating the database in all sorts of ways.  However, since we're only spinning the database up once we run the risk of spec pollution if one spec does something that puts the database in a state that will influence an independent spec.  To avoid that, it's a common pattern to introduce a top-level `AfterEach` to clear out our database.  This `AfterEach` closure will run after each spec and clear out the database ensuring a pristine state for the spec.  This is often much faster than instantiating a new copy of the database!

Finally, the `AfterSuite` closure will run after all the tests to tear down the running database via `dbRunner.Stop()`.  We can, alternatively, use `DeferCleanup` to achieve the same effect:

```go
var _ = BeforeSuite(func() {
  dbRunner = db.NewRunner()
  Expect(dbRunner.Start()).To(Succeed())
  DeferCleanup(dbRunner.Stop)

  dbClient = db.NewClient()
  Expect(dbClient.Connect(dbRunner.Address())).To(Succeed())
})
```

`DeferCleanup` is context-aware and knows that it's being called in a `BeforeSuite`.  The registered cleanup code will only run after all the specs have completed, just like `AfterSuite`.

One quick note before we move on.  We've introduced Gomega's [`Succeed()`](https://onsi.github.io/gomega/#handling-errors) matcher here.  `Succeed()` simply asserts that a passed-in error is `nil`.  The following two assertions are equivalent:

```go
err := dbRunner.Start()
Expect(err).NotTo(HaveOccurred())

/* is equivalent to */

Expect(dbRunner.Start()).To(Succeed())
```

The `Succeed()` form is more succinct and reads clearly.

> We won't get into it here but make sure to keep reading to understand how Ginkgo manages [suite parallelism](#spec-parallelization) and provides [SynchronizedBeforeSuite and SynchronizedAfterSuite](#parallel-suite-setup-and-cleanup-synchronizedbeforesuite-and-synchronizedaftersuite) suite setup nodes.

### Mental Model: How Ginkgo Handles Failure
So far we've focused on how Ginkgo specs are constructed using nested nodes and how node closures are called in order when specs run.

...but Ginkgo is a testing framework.  And tests fail!  Let's delve into how Ginkgo handles failure.

You typically use a matcher library, like [Gomega](https://github.com/onsi/gomega) to make assertions in your spec.  When a Gomega assertion fails, Gomega generates a failure message and passes it to Ginkgo to signal that the spec has failed.  It does this via Ginkgo's global `Fail` function.  Of course, you're allowed to call this function directly yourself:

```go
It("can read books", func() {
  if book.Title == "Les Miserables" && user.Age <= 3 {
    Fail("User is too young for this book")
  }
  user.Read(book)
})
```

whether in a setup or subject node, whenever `Fail` is called Ginkgo will mark the spec as failed and record and display the message passed to `Fail`.

But there's more.  The `Fail` function **panics** when it is called.  This allows Ginkgo to stop the current closure in its tracks - no subsequent assertions or code in the closure will run.  Ginkgo is quite opinionated about this behavior - if an assertion has failed then the current spec is not in an expected state and subsequent assertions will likely fail.  This fast-fail approach is especially useful when running slow complex integration tests.  It cannot be disabled.

When a failure occurs in a `BeforeEach`, `JustBeforeEach`, or `It` closure Ginkgo halts execution of the current spec and cleans up by invoking any registered `AfterEach` or `JustAfterEach` closures (and any registered `DeferCleanup` closures if applicable).  This is important to ensure the spec state is cleaned up.  

Ginkgo orchestrates this behavior by rescuing the panic thrown by `Fail` and unwinding the spec.  However, if your spec launches a **goroutine** that calls `Fail` (or, equivalently, invokes a failing Gomega assertion), there's no way for Ginkgo to rescue the panic that `Fail` throws.  This will cause the suite to panic and no subsequent specs will run.  To get around this you must rescue the panic using `defer GinkgoRecover()`.  Here's an example:

```go
It("panics in a goroutine", func() {
  var c chan interface{}
  go func() {
    defer GinkgoRecover()
    Fail("boom")
    close(c)
  }()
  <-c
})
```

You must remember to follow this pattern when making assertions in goroutines - however, if uncaught, Ginkgo's panic will include a helpful error to remind you to add `defer GinkgoRecover()` to your goroutine.

When a failure occurs Ginkgo marks the current spec as failed and moves on to the next spec.  If, however, you'd like to stop the entire suite when the first failure occurs you can run `ginkgo --fail-fast`.

One last thing before we move on.  When a failure occurs, Ginkgo records and presents the location of the failure to help you pinpoint where to look to debug your specs.  This is typically the line where the call to `Fail` was performed (or, if you're using Gomega, the line where the Gomega assertion failed).  Sometimes, however, you need to control the reported location.  For example, consider the case where you are using a helper function:

```go
/* === INVALID === */
func EnsureUserCanRead(book Book, user User) {
  if book.Title == "Les Miserables" && user.Age <= 3 {
    Fail("user is too young for this book") //A
  }
}

It("can read books", func() {
  EnsureUserCanRead(book, user) //B
  user.Read(book)
})
```

Now, if the `EnsureUserCanRead` helper fails the location presented to the user will point to `//A`.  Ideally, however we'd prefer that Ginkgo report `//B`.

There are a few ways to solve for this.  The first is to pass `Fail` an `offset` like so:

```go
func EnsureUserCanRead(book Book, user User) {
  if book.Title == "Les Miserables" && user.Age <= 3 {
    Fail("user is too young for this book", 1)
  }
}
```

This will tell Ginkgo to skip a stack frame when calculating the offset.  In this particular case Ginkgo will report the location that called `EnsureUserCanRead`: i.e. `//B`.

This works... however managing offset can quickly get unwieldy.  For example, say we wanted to compose helpers:

```go
func EnsureUserCanCheckout(book Book, user User) {
  EnsureUserCanRead(book, user)
  EnsureUserHasAccessTo(book, user)
}
```

in _this_ case, we'd need the offset that `EnsureUserCanRead` passes to `Fail` to be `2` instead of `1`.

Instead of managing offsets you can use `GinkgoHelper()`:

```go
func EnsureUserCanRead(book Book, user User) {
  GinkgoHelper()
  if book.Title == "Les Miserables" && user.Age <= 3 {
    Fail("user is too young for this book") //note the optional offset is gone
  }
}

func EnsureUserCanCheckout(book Book, user User) {
  GinkgoHelper()
  EnsureUserCanRead(book, user)
  EnsureUserHasAccessTo(book, user)
}
```

Any function in which `GinkgoHelper()` is called is tracked by Ginkgo and ignored when a failure location is being computed.  This allows you to build reusable test helpers and trust that the location presented to the user will always be in the spec that called the helper, and not the helper itself.

### Logging Output
As outlined above, when a spec fails - say via a failed Gomega assertion - Ginkgo will pass the failure message passed to the `Fail`  handler.  Often times the failure message generated by Gomega gives you enough information to understand and resolve the spec failure.

But there are several contexts, particularly when running large complex integration suites, where additional debugging information is necessary to understand the root cause of a failed spec.  You'll typically only want to see this information if a spec has failed - and hide it if the spec succeeds.

Ginkgo provides a globally available `io.Writer` called `GinkgoWriter` that solves for this usecase.  `GinkgoWriter` aggregates everything written to it while a spec is running and only emits to stdout if the test fails.

`GinkgoWriter` includes three convenience methods:

- `GinkgoWriter.Print(a ...interface{})` is equivalent to `fmt.Fprint(GinkgoWriter, a...)`
- `GinkgoWriter.Println(a ...interface{})` is equivalent to `fmt.Fprintln(GinkgoWriter, a...)`
- `GinkgoWriter.Printf(format string, a ...interface{})` is equivalent to `fmt.Fprintf(GinkgoWriter, format, a...)`

You can also attach additional `io.Writer`s for `GinkgoWriter` to tee to via `GinkgoWriter.TeeTo(writer)`.  Any data written to `GinkgoWriter` will immediately be sent to attached tee writers.  All attached Tee writers can be cleared with `GinkgoWriter.ClearTeeWriters()`.

Finally - when running in verbose mode via `ginkgo -v` anything written to `GinkgoWriter` will be immediately streamed to stdout.  This can help shorten the feedback loop when debugging a complex spec.

If [logr](https://github.com/go-logr/logr) is used for logging in a project the globally available `GinkgoLogr` provides a logger implementation. Any logging on `GinkgoLogr` is forwarded to `GinkgoWriter`.

### Documenting Complex Specs: By
As a rule, you should try to keep your subject and setup closures short and to the point.  Sometimes this is not possible, particularly when testing complex workflows in integration-style tests.  In these cases your test blocks begin to hide a narrative that is hard to glean by looking at code alone.  Ginkgo provides `By` to help in these situations.  Here's an example:

```go
var _ = Describe("Browsing the library", func() {
  BeforeEach(func() {
    By("Fetching a token and logging in")

    authToken, err := authClient.GetToken("gopher", "literati")
    Expect(err).NotTo(HaveOccurred())

    Expect(libraryClient.Login(authToken)).To(Succeed())
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

    By("Checking a book out")
    Expect(libraryClient.CheckOut(book)).To(Succeed())
    books, err = aisle.GetBooks()
    Expect(err).NotTo(HaveOccurred())
    Expect(books).To(HaveLen(6))
    Expect(books).NotTo(ContainElement(book))
  })
})
```

The string passed to `By` is attached to the spec and can be displayed by Ginkgo when needed.  If a test succeeds you won't see any output beyond Ginkgo's green dot.  If a test fails, however, you will see each step printed out up to the step immediately preceding the failure.  Running with `ginkgo -v` always emits all steps.

`By` takes an optional function of type `func()`.  When passed such a function `By` will immediately call the function.  This allows you to organize your `It`s into groups of steps.  

`By` doesn't affect the structure of your specs - it's primarily syntactic sugar to help you document long and complex specs.  Ginkgo has additional mechanisms to break specs up into more granular subunits with guaranteed ordering - we'll discuss [Ordered containers](#ordered-containers) in detail later.

### Mental Model: Spec Timelines
Several events can occur during the lifecycle of a Ginkgo spec.  You've seen a few of these already: various setup and subject nodes start and end; data is written to the `GinkgoWriter`; `By` annotations are generated; failures occur.  And there are several more that you'll see introduced later in these docs (e.g. [`ReportEntries`](#attaching-data-to-reports) and [Progess Reports](#getting-visibility-into-long-running-specs) are attached to specs; [flaky specs](#repeating-spec-runs-and-managing-flaky-specs) might be retried).

By default, when a spec passes Ginkgo does not emit any of this information.  When a failure occurs, however, Ginkgo emits a **timeline** view of the spec.  This includes all the events and `GinkgoWriter` output associated with a spec in ther order they were generated and provides the context needed to debug the spec and understand the nature and context of the failure.

You can view the timeline for all specs (whether passed or failed) by running `ginkgo -v` or `ginkgo -vv`.

### Table Specs
We'll round out this chapter on [Writing Specs](#writing-specs) with one last topic.  Ginkgo provides an expressive DSL for writing table driven specs.  This DSL is a simple wrapper around concepts you've already met - container nodes like `Describe` and subject nodes like `It`.

Let's write a table spec to describe the Author name functions we tested earlier:

```go
DescribeTable("Extracting the author's first and last name",
  func(author string, isValid bool, firstName string, lastName string) {
    book := &books.Book{
      Title: "My Book",
      Author: author,
      Pages: 10,
    }
    Expect(book.IsValid()).To(Equal(isValid))
    Expect(book.AuthorFirstName()).To(Equal(firstName))
    Expect(book.AuthorLastName()).To(Equal(lastName))
  },
  Entry("When author has both names", "Victor Hugo", true, "Victor", "Hugo"),
  Entry("When author has one name", "Hugo", true, "", "Hugo"),
  Entry("When author has a middle name", "Victor Marie Hugo", true, "Victor", "Hugo"),
  Entry("When author has no name", "", false, "", ""),
)
```

`DescribeTable` takes a string description, a **spec closure** to run for each table entry, and a set of entries.  Each `Entry` takes a string description, followed by a list of parameters.  `DescribeTable` will generate a spec for each `Entry` and when the specs run, the `Entry` parameters will be passed to the spec closure and must match the types expected by the spec closure.

You'll be notified with a clear message at runtime if the parameter types don't match the spec closure signature.

#### Mental Model: Table Specs are just Syntactic Sugar
`DescribeTable` is simply providing syntactic sugar to convert its Ls into a set of standard Ginkgo nodes.  During the [Tree Construction Phase](#mental-model-how-ginkgo-traverses-the-spec-hierarchy) `DescribeTable` is generating a single container node that contains one subject node per table entry.  The description for the container node will be the description passed to `DescribeTable` and the descriptions for the subject nodes will be the descriptions passed to the `Entry`s.  During the Run Phase, when specs run, each subject node will simply invoke the spec closure passed to `DescribeTable`, passing in the parameters associated with the `Entry`.

To put it another way, the table test above is equivalent to:

```go
Describe("Extracting the author's first and last name", func() {
  It("When author has both names", func() {
    book := &books.Book{
      Title: "My Book",
      Author: "Victor Hugo",
      Pages: 10,
    }
    Expect(book.IsValid()).To(Equal(true))
    Expect(book.AuthorFirstName()).To(Equal("Victor"))
    Expect(book.AuthorLastName()).To(Equal("Hugo"))
  })

  It("When author has one name", func() {
    book := &books.Book{
      Title: "My Book",
      Author: "Hugo",
      Pages: 10,
    }
    Expect(book.IsValid()).To(Equal(true))
    Expect(book.AuthorFirstName()).To(Equal(""))
    Expect(book.AuthorLastName()).To(Equal("Hugo"))
  })

  It("When author has a middle name", func() {
    book := &books.Book{
      Title: "My Book",
      Author: "Victor Marie Hugo",
      Pages: 10,
    }
    Expect(book.IsValid()).To(Equal(true))
    Expect(book.AuthorFirstName()).To(Equal("Victor"))
    Expect(book.AuthorLastName()).To(Equal("Hugo"))
  })  

  It("When author has no name", func() {
    book := &books.Book{
      Title: "My Book",
      Author: "",
      Pages: 10,
    }
    Expect(book.IsValid()).To(Equal(false))
    Expect(book.AuthorFirstName()).To(Equal(""))
    Expect(book.AuthorLastName()).To(Equal(""))
  })  
})
```

As you can see - the table spec can capture this sort of repetitive testing much more concisely!

Since `DescribeTable` is simply generating a container node you can nest it within other containers and surround it with setup nodes like so:

```go
Describe("book", func() {
  var book *books.Book

  BeforeEach(func() {
    book = &books.Book{
      Title: "Les Miserables",
      Author: "Victor Hugo",
      Pages: 2783,
    }
    Expect(book.IsValid()).To(BeTrue())
  })

  DescribeTable("Extracting the author's first and last name",
    func(author string, isValid bool, firstName string, lastName string) {
      book.Author = author
      Expect(book.IsValid()).To(Equal(isValid))
      Expect(book.AuthorFirstName()).To(Equal(firstName))
      Expect(book.AuthorLastName()).To(Equal(lastName))
    },
    Entry("When author has both names", "Victor Hugo", true, "Victor", "Hugo"),
    Entry("When author has one name", "Hugo", true, "", "Hugo"),
    Entry("When author has a middle name", "Victor Marie Hugo", true, "Victor", "Hugo"),
    Entry("When author has no name", "", false, "", ""),
  )

})
```

the `BeforeEach` closure will run before each table entry spec and set up a fresh copy of `book` for the spec closure to manipulate and assert against.

The fact that `DescribeTable` is constructed during the Tree Construction Phase can trip users up sometimes.  Specifically, variables declared in container nodes have not been initialized yet during the Tree Construction Phase.  Because of this, the following will not work:

```go
/* === INVALID === */
Describe("book", func() {
  var shelf map[string]*books.Book //Shelf is declared here

  BeforeEach(func() {
    shelf = map[string]*books.Book{ //...and initialized here
      "Les Miserables": &books.Book{Title: "Les Miserables", Author: "Victor Hugo", Pages: 2783},
      "Fox In Socks": &books.Book{Title: "Fox In Socks", Author: "Dr. Seuss", Pages: 24},
    }
  })

  DescribeTable("Categorizing books",
    func(book *books.Book, category books.Category) {
      Expect(book.Category()).To(Equal(category))
    },
    Entry("Novels", shelf["Les Miserables"], books.CategoryNovel),
    Entry("Short story", shelf["Fox in Socks"], books.CategoryShortStory),
  )
})
```

These specs will fail.  When `DescribeTable` and `Entry` are invoked during the Tree Construction Phase `shelf` will have been declared but uninitialized.  So `shelf["Les Miserables"]` will return a `nil` pointer and the spec will fail.

To get around this we must move access of the `shelf` variable into the body of the spec closure so that it can run at the appropriate time during the Run Phase.  We can do this like so:

```go
Describe("book", func() {
  var shelf map[string]*books.Book //Shelf is declared here

  BeforeEach(func() {
    shelf = map[string]*books.Book{ //...and initialized here
      "Les Miserables": &books.Book{Title: "Les Miserables", Author: "Victor Hugo", Pages: 2783},
      "Fox In Socks": &books.Book{Title: "Fox In Socks", Author: "Dr. Seuss", Pages: 24},
    }
  })

  DescribeTable("Categorizing books",
    func(key string, category books.Category) {
      Expect(shelf[key]).To(Equal(category))
    },
    Entry("Novels", "Les Miserables", books.CategoryNovel),
    Entry("Novels", "Fox in Socks", books.CategoryShortStory),
  )
})
```

We're now accessing the `shelf` variable in the spec closure during the Run Phase and can trust that it has been correctly instantiated by the setup node closure.

Be sure to check out the [Table Patterns](#table-specs-patterns) section of the [Ginkgo and Gomega Patterns](#ginkgo-and-gomega-patterns) chapter to learn about a few more table-based patterns.

#### Generating Entry Descriptions
In the examples we've shown so far, we are explicitly passing in a description for each table entry.  Recall that this description is used to generate the description of the resulting spec's Subject node.  That means it's important as it conveys the intent of the spec and is printed out in case the spec fails.

There are times, though, when adding a description manually can be tedious, repetitive, and error prone.  Consider this example:

```go
var _ = Describe("Math", func() {
  DescribeTable("addition",
    func(a, b, c int) {
      Expect(a+b).To(Equal(c))
    },
    Entry("1+2=3", 1, 2, 3),
    Entry("-1+2=1", -1, 2, 1),
    Entry("0+0=0", 0, 0, 0),
    Entry("10+100=101", 10, 100, 110), //OOPS TYPO
  )
})
```

Mercifully, Ginkgo's table DSL provides a few mechanisms to programmatically generate entry descriptions.

**`nil` Descriptions**

First - Entries can have their descriptions auto-generated by passing `nil` for the `Entry` description:

```go
var _ = Describe("Math", func() {
  DescribeTable("addition",
    func(a, b, c int) {
      Expect(a+b).To(Equal(c))
    },
    Entry(nil, 1, 2, 3),
    Entry(nil, -1, 2, 1),
    Entry(nil, 0, 0, 0),
    Entry(nil, 10, 100, 110),
  )
})
```

This will generate entries named after the spec parameters.  In this case we'd have `Entry: 1, 2, 3`, `Entry: -1, 2, 1`, `Entry: 0, 0, 0`, `Entry: 10, 100, 110`.

**Custom Description Generator**

Second - you can pass a table-level Entry **description closure** to render entries with `nil` description:

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
    Entry(nil, 0, 0, 0),
    Entry(nil, 10, 100, 110),
  )
})
```

This will generate entries named `1 + 2 = 3`, `-1 + 2 = 1`, `0 + 0 = 0`, and `10 + 100 = 110`.

The description closure must return a `string` and must accept the same parameters passed to the spec closure.

**`EntryDescription()` format string**

There's also a convenience decorator called `EntryDescription` to specify Entry descriptions as format strings:

```go
var _ = Describe("Math", func() {
  DescribeTable("addition",
    func(a, b, c int) {
      Expect(a+b).To(Equal(c))
    },
    EntryDescription("%d + %d = %d")
    Entry(nil, 1, 2, 3),
    Entry(nil, -1, 2, 1),
    Entry(nil, 0, 0, 0),
    Entry(nil, 10, 100, 110),
  )
})
```

This will have the same effect as the description above.

**Per-Entry Descriptions**

In addition to `nil` and strings you can also pass a string-returning closure or an `EntryDescription` as the first argument to `Entry`.  Doing so will cause the entry's description to be generated by the passed-in closure or `EntryDescription` format string.

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
    Entry(EntryDescription("%[3]d = %[1]d + %[2]d"), 10, 100, 110)
    Entry(func(a, b, c int) string {fmt.Sprintf("%d = %d", a + b, c)}, 4, 3, 7)
  )
})
```

Will generate entries named: `1 + 2 = 3`, `-1 + 2 = 1`, `zeros`, `110 = 10 + 100`, and `7 = 7`.

### Alternatives to Dot-Importing Ginkgo

As shown throughout this documentation, Ginkgo users are encouraged to dot-import the Ginkgo DSL into their test suites to effectively extend the Go language with Ginkgo's expressive building blocks:

```go
import . "github.com/onsi/ginkgo/v2"
```

Some users prefer to avoid dot-importing dependencies into their code in order to keep their global namespace clean and predictable.  You can, of course, do this with Ginkgo - we recommend using a simple shorthand like `g`:

```go
import g "github.com/onsi/ginkgo/v2"
```

now you can write tests as before, albeit with a slight stutter:

```go
var _ = g.Describe("Books", func() {
  g.BeforeEach(func() { ... })

  g.It("works as before", func() {
    g.By("you just need to repeat g. everywhere")
  })
})
```

Alternatively, you can choose to dot-import only _portions_ of Ginkgo's DSL into the global namespace.  The packages under `github.com/onsi/ginkgo/v2/dsl` organize the various pieces of Ginkgo into a series of subpackages.  You can choose to mix-and-match which of these are dot-imported vs namespaced.  For example, you can dot-import the core DSL (which provides the various setup, container, and subject nodes) while namespace importing the decorators DSL:

```go
import (
  . "github.com/onsi/ginkgo/v2/dsl/core"  
  "github.com/onsi/ginkgo/v2/dsl/decorators"  
)

var _ = It("gives you the core DSL", decorators.Label("and namespaced decorators"), func() {
  ...
})
```

The available DSL packages are:

| Package | Contents |
|-------|--------|
| `github.com/onsi/ginkgo/v2/dsl/core` | The core DSL including all container, setup, and subject nodes (`Describe`, `Context`, `BeforeEach`, `BeforeSuite`, `It`, etc...) as well as the most commonly used functions (`RunSpecs`, `Skip`, `Fail`, `By`, `GinkgoT`) | 
| `github.com/onsi/ginkgo/v2/decorators` | The decorator DSL includes all Ginkgo's decorators (e.g. `Label`, `Ordered`, `Serial`, etc...) |
| `github.com/onsi/ginkgo/v2/reporting` | The reporting DSL includes all reporting-related nodes and types (e.g. `Report`, `CurrentSpecReport`, `ReportAfterEach`, `AddReportEntry`) |
| `github.com/onsi/ginkgo/v2/table` | The table DSL includes all table-related types and functions (e.g. `DescribeTable`, `Entry`, `EntryDescription`) |

The DSL packages simply import and then re-export pieces of the Ginkgo DSL provided by `github.com/onsi/ginkgo/v2` so there are no differences in behavior or interoperability if you use the standard dot-import for Ginkgo or pull in the various DSL packages in piecemeal.

## Running Specs

The previous chapter covered the basics of [Writing Specs](#writing-specs) in Ginkgo.  We explored how Ginkgo lets you use container nodes, subject nodes, and setup nodes to construct hierarchical spec trees; and how Ginkgo transforms those trees into a list of specs to run.

In this chapter we'll shift our focus from the Tree Construction Phase to the Run Phase and dive into the various capabilities Ginkgo provides for manipulating the spec list and controlling how specs run.

To start, let's continue to flesh out our mental model for Ginkgo.

### Mental Model: Ginkgo Assumes Specs are Independent

We've already seen how Ginkgo generates a spec tree and converts it to a flat list of specs.  If you need a refresher, skim through the [Mental Model: How Ginkgo Traverses the Spec Hierarchy](#mental-model-how-ginkgo-traverses-the-spec-hierarchy) section up above.

Lists are powerful things.  They can be sorted.  They can be randomized.  They can be filtered.  They can be distributed to multiple workers.  Ginkgo supports all of these manipulations of the spec list enabling you to randomize, filter, and parallelize your test suite with minimal effort.

To unlock these powerful capabilities Ginkgo makes an important, foundational, assumption about the specs in your suite:

**Ginkgo assumes specs are independent**.

Because individual Ginkgo specs do not depend on each other, it is possible to run them in any order; it is possible to run subsets of them; it is even possible to run them simultaneously in parallel.  Ensuring your specs are independent is foundational to writing effective Ginkgo suites that make the most of Ginkgo's capabilities.

In the next few sections we'll unpack how Ginkgo randomizes specs and supports running specs in parallel.  As we do, we'll cover principles that - if followed - will help you write specs that are independent from each other.

### Spec Randomization

By default, Ginkgo will randomize the order in which the specs in a suite run.  This is done intentionally.  By randomizing specs, Ginkgo can help suss out spec pollution - accidental dependencies between specs - throughout a suite's development.

Ginkgo's default behavior is to only randomize the order of top-level containers -- the specs *within* those containers continue to run in the order in which they are specified in the test files.  This is helpful when developing specs as it mitigates the cognitive overload of having specs within a container continuously change the order in which they run during a debugging session.

When running on CI, or before committing code, it's good practice to instruct Ginkgo to randomize **all** specs in a suite.  You do this with the `--randomize-all` flag:

```bash
ginkgo --randomize-all
```

Ginkgo uses the current time to seed the randomization and prints out the seed near the beginning of the suite output.  If you notice intermittent spec failures that you think may be due to spec pollution, you can use the seed from a failing suite to exactly reproduce the spec order for that suite.  To do this pass the `--seed=SEED` flag:

```bash
ginkgo --seed=17
```

Because Ginkgo randomizes specs you should make sure that each spec runs from a clean independent slate.  Principles like ["Declare in container nodes, initialize in setup nodes"](#avoid-spec-pollution-dont-initialize-variables-in-container-nodes) help you accomplish this: when variables are initialized in setup nodes each spec is guaranteed to get a fresh, correctly initialized, state to operate on.  For example:

```go
/* === INVALID === */
Describe("Bookmark", func() {
  book := &books.Book{
    Title:  "Les Miserables",
    Author: "Victor Hugo",
    Pages:  2783,
  }

  It("has no bookmarks by default", func() {
    Expect(book.Bookmarks()).To(BeEmpty())
  })

  It("can add bookmarks", func() {
    book.AddBookmark(173)
    Expect(book.Bookmarks()).To(ContainElement(173))
  })
})
```

This suite only passes if the "has no bookmarks" spec runs before the "can add bookmarks" spec.  Instead, you should initialize the book variable in a setup node:

```go
Describe("Bookmark", func() {
  var book *books.Book

  BeforeEach(func() {
    book = &books.Book{
      Title:  "Les Miserables",
      Author: "Victor Hugo",
      Pages:  2783,
    }    
  })

  It("has no bookmarks by default", func() {
    Expect(book.Bookmarks()).To(BeEmpty())
  })

  It("can add bookmarks", func() {
    book.AddBookmark(173)
    Expect(book.Bookmarks()).To(ContainElement(173))
  })
})
```

In addition to avoiding accidental spec pollution you should make sure to avoid _intentional_ spec pollution!  Specifically, you should ensure that the correctness of your suite does not rely on the order in which specs run.

For example:

```go
/* === INVALID === */
Describe("checking out a book", func() {
  var book *books.Book
  var err error

  It("can fetch a book from a library", func() {
    book, err = libraryClient.FindByTitle("Les Miserables")
    Expect(err).NotTo(HaveOccurred())
    Expect(book.Title).To(Equal("Les Miserables"))
  })

  It("can check out the book", func() {
    Expect(library.CheckOut(book)).To(Succeed())
  })

  It("no longer has the book in stock", func() {
    book, err = libraryClient.FindByTitle("Les Miserables")
    Expect(err).To(MatchError(books.NOT_IN_STOCK))
    Expect(book).To(BeNil())
  })
})
```

These specs are not independent - the assume that they run in order.  This means they can't be randomized or parallelized with respect to each other.

You can fix these specs by creating a single `It` to test the behavior of checking out a book:

```go
Describe("checking out a book", func() {
  It("can perform a checkout flow", func() {
    By("fetching a book")
    book, err := libraryClient.FindByTitle("Les Miserables")
    Expect(err).NotTo(HaveOccurred())
    Expect(book.Title).To(Equal("Les Miserables"))

    By("checking out the book")
    Expect(library.CheckOut(book)).To(Succeed())


    By("validating the book is no longer in stock")
    book, err = libraryClient.FetchByTitle("Les Miserables")
    Expect(err).To(MatchError(books.NOT_IN_STOCK))
    Expect(book).To(BeNil())
  })
})
```

Ginkgo also provides an alternative that we'll discuss later - you can use [Ordered Containers](#ordered-containers) to tell Ginkgo when the specs in a container must _always_ be run in order.

Finally, if your specs need to _generate_ random numbers you can seed your pseudo-random number generator with the same seed used to seed Ginkgo's randomization.  This will help ensure that specifying the random seed fully determines the pseudo-random aspects of your suite.  You can get access to the random seed in the spec using `GinkgoRandomSeed()`

### Spec Parallelization

As spec suites grow in size and complexity they have a tendency to get slower.  Thankfully the vast majority of modern computers ship with multiple CPU cores.  Ginkgo helps you use those cores to speed up your suites by running specs in parallel.  This is _especially_ useful when running large, complex, and slow integration suites where the only means to speed things up is to embrace parallelism.

To run a Ginkgo suite in parallel you simply pass the `-p` flag to `ginkgo`:

```bash
ginkgo -p
```

this will automatically detect the optimal number of test processes to spawn based on the number of cores on your machine.  You can, instead, specify this number manually via `-procs=N`:

```bash
ginkgo -procs=N
```

And that's it!  Ginkgo will automatically run your specs in parallel and take care of collating the results into a single coherent output stream.

At this point, though, you may be scratching your head.  _How_ does Ginkgo support parallelism given the use of shared closure variables we've seen throughout?  Consider the example from above:

```go
Describe("Bookmark", func() {
  var book *books.Book

  BeforeEach(func() {
    book = &books.Book{
      Title:  "Les Miserables",
      Author: "Victor Hugo",
      Pages:  2783,
    }    
  })

  It("has no bookmarks by default", func() {
    Expect(book.Bookmarks()).To(BeEmpty())
  })

  It("can add bookmarks", func() {
    book.AddBookmark(173)
    Expect(book.Bookmarks()).To(ContainElement(173))
  })
})
```

Both "Bookmark" specs are interrogating and mutating the same shared `book` variable.  Running the two specs in parallel would lead to an obvious data race over `book` and undefined, seemingly random, behavior.

#### Mental Model: How Ginkgo Runs Parallel Specs

Ginkgo ensures specs running in parallel are fully isolated from one another.  It does this by running the specs in _different processes_.  Because Ginkgo specs are assumed to be fully independent they can be harvested out to run on different worker processes - each process has its own memory space and there is, therefore, no risk for shared variable data races.

Here's what happens under the hood when you run `ginkgo -p`:

First, the Ginkgo CLI compiles a single test binary (via `go test -c`).  It then invokes `N` copies of the test binary.

Each of these processes then enters the Tree Construction Phase and all processes generate an identical spec tree and, therefore, an identical list of specs to run.  The processes then enter the Run Phase and start running their specs.  They coordinate via the Ginkgo CLI (which acts a server) to figure out the next spec to run, and report to the CLI as specs finish running.  The CLI then takes care of generating a single coherent output stream of the running specs.  In essence, this is a simple map-reduce system with the CLI playing the role of a centralized server.

With few exceptions, the different test processes do not communicate with one another and for most spec suites you, the developer, do not need to worry about which spec is running on which process.  This makes it easy to parallelize your suites and get some major performance gains.

There are, however, contexts where you _do_ need to be aware of which process a given spec is running on.  In particular, there are several patterns for building effective parallelizable integration suites that need this information. We will explore such patterns in much more detail in the [Patterns chapter](#patterns-for-parallel-integration-specs) - feel free to jump straight there if you're interested!  For now we'll simply introduce some of the building blocks that Ginkgo provides for implementing these patterns.

#### Discovering Which Parallel Process a Spec is Running On

Ginkgo numbers the running parallel processes from `1` to `N`.  A spec can get the index of the Ginkgo process it is running on via `GinkgoParallelProcess()`.  This can be useful in contexts where specs need to share a globally available external resource but need to access a specific shard, namespace, or instance of the resource so as to avoid spec pollution.  For example:

```go
Describe("Storing books in an external database", func() {
  BeforeEach(func() {
    namespace := fmt.Sprintf("namespace-%d", GinkgoParallelProcess())
    Expect(dbClient.SetNamespace(namespace)).To(Succeed())
    DeferCleanup(dbClient.ClearNamespace, namespace)
  })

  It("returns empty when there are no books", func() {
    Expect(dbClient.Books()).To(BeEmpty())
  })

  Context("when a book is in the database", func() {
    var book *books.Book
    BeforeEach(func() {
      book = &books.Book{
        Title:  "Les Miserables",
        Author: "Victor Hugo",
        Pages:  2783,
      }
      Expect(dbClient.Store(book)).To(Succeed())
    })

    It("can fetch the book", func() {
      Expect(dbClient.Books()).To(ConsistOf(book))
    })

    It("can update the book", func() {
      book.Author = "Victor Marie Hugo"
      Expect(dbClient.Store(book)).To(Succeed())
      Expect(dbClient.Books()).To(ConsistOf(book))
    })

    It("can delete the book", func() {
      Expect(dbClient.Delete(book)).To(Succeed())
      Expect(dbClient.Books()).To(BeEmpty())      
    })
  })
})
```

Without sharding access to the database these specs would step on each other's toes and result in non-deterministic flaky behavior. By implementing sharded access to the database (e.g. `dbClient.SetNamespace` could instruct the client to prepend the `namespace` string to any keys stored in a key-value database) this suite can be trivially parallelized.  And by extending the "declare in container nodes, initialize in setup nodes" principle to apply to state stored _external_ to the suite we are able to ensure that each spec runs from a known clean shard of the database.

Such a suite will continue to be parallelizable as it grows - enabling faster runtimes with less flakiness than would otherwise be possible in a serial-only suite.

In addition to `GinkgoParallelProcess()`, Ginkgo provides access to the total number of running processes.  You can get this from `GinkgoConfiguration()`, which returns the state of Ginkgo's configuration, like so:

```go
suiteConfig, _ := GinkgoConfiguration()
totalProcesses := suiteConfig.ParallelTotal
```

#### Parallel Suite Setup and Cleanup: SynchronizedBeforeSuite and SynchronizedAfterSuite

Our example above assumed the existence of a single, globally shared, running database.  How might we have set up such a database?

You typically spin up external resources like this in the `BeforeSuite` in your suite bootstrap file.  We saw this example earlier:

```go
var dbClient *db.Client
var dbRunner *db.Runner

var _ = BeforeSuite(func() {
  dbRunner := db.NewRunner()
  Expect(dbRunner.Start()).To(Succeed())

  dbClient = db.NewClient()
  Expect(dbClient.Connect(dbRunner.Address())).To(Succeed())
})

var _ = AfterSuite(func() {
  Expect(dbClient.Cleanup()).To(Succeed())
  Expect(dbRunner.Stop()).To(Succeed())
})
```

However, since `BeforeSuite` runs on _every_ parallel process this would result in `N` independent databases spinning up.  Sometimes that's exactly what you want - as it provides maximal isolation for the running specs and is a natural way to shard data access.  Sometimes, however, spinning up multiple external processes is too resource intensive or slow and it is more efficient to share access to a single resource.

Ginkgo supports this usecase with `SynchronizedBeforeSuite` and `SynchronizedAfterSuite`.  Here are the full signatures for the two:

```go
func SynchronizedBeforeSuite(
  process1 func() []byte,
  allProcesses func([]byte),
)

func SynchronizedAfterSuite(
  allProcesses func(),
  process1 func(),
)
```

Let's dig into `SynchronizedBeforeSuite` (henceforth `SBS`) first.  `SBS` runs at the beginning of the Run Phase - before any specs have run but after the spec tree has been parsed and constructed.

`SBS` allows us to set up state in one process, and pass information to all the other processes.  Concretely, the `process1` function runs **only** on parallel process #1.  All other parallel processes pause and wait for `process1` to complete.  Upon completion `process1` returns arbitrary data as a `[]byte` slice and this data is then passed to all parallel processes which then invoke the `allProcesses` function in parallel, passing in the `[]byte` slice.  Note that the passing of a `[]byte` slice from `process1` to `allProcesses` is optional.  `SynchronizedBeforeSuite` also supports the following signature:

```go
func SynchronizedBeforeSuite(
  process1 func(),
  allProcesses func(),
)
```

Similarly, `SynchronizedAfterSuite` is split into two functions.  The first, `allProcesses`, runs on all processes after they finish running specs.  The second, `process1`, only runs on process #1 - and only _after_ all other processes have finished and exited.

We can use this behavior to set up shared external resources like so:

```go
var dbClient *db.Client
var dbRunner *db.Runner

var _ = SynchronizedBeforeSuite(func() []byte {
  //runs *only* on process #1
  dbRunner := db.NewRunner()
  Expect(dbRunner.Start()).To(Succeed())
  return []byte(dbRunner.Address())
}), func(address []byte) {
  //runs on *all* processes
  dbClient = db.NewClient()
  Expect(dbClient.Connect(string(address))).To(Succeed())
  dbClient.SetNamespace(fmt.Sprintf("namespace-%d", GinkgoParallelProcess()))
})

var _ = SynchronizedAfterSuite(func() {
  //runs on *all* processes
  Expect(dbClient.Cleanup()).To(Succeed())  
}, func() {
  //runs *only* on process #1
  Expect(dbRunner.Stop()).To(Succeed())
})
```

This code will spin up a single database and ensure that every parallel Ginkgo process connects to the database and sets up an appropriately sharded namespace.  Ginkgo does all the work of coordinating across these various closures and passing information back and forth - and all the complexity of the parallel setup in the test suite is now contained in the `Synchronized*` setup nodes.

By the way, we can clean all this up further using `DeferCleanup`.  `DeferCleanup` is context aware and so knows that any cleanup code registered in a `BeforeSuite`/`SynchronizedBeforeSuite` should run at the end of the suite:

```go
var dbClient *db.Client

var _ = SynchronizedBeforeSuite(func() []byte {
  //runs *only* on process #1
  dbRunner := db.NewRunner()
  Expect(dbRunner.Start()).To(Succeed())
  DeferCleanup(dbRunner.Stop)
  return []byte(dbRunner.Address())
}), func(address []byte) {
  //runs on *all* processes
  dbClient = db.NewClient()
  Expect(dbClient.Connect(string(address))).To(Succeed())
  dbClient.SetNamespace(fmt.Sprintf("namespace-%d", GinkgoParallelProcess()))
  DeferCleanup(dbClient.Cleanup)
})
```

#### The ginkgo CLI vs go test
One last word before we close out the topic of Spec Parallelization.  Ginkgo's process-based server-client parallelization model should make clear why you need to use the `ginkgo` CLI to run parallel specs instead of `go test`.  While Ginkgo suites are fully compatible with `go test` there _are_ some features, most notably parallelization, that require the use of the` ginkgo` CLI.

We recommend embracing the `ginkgo` CLI as part of your toolchain and workflow.  It's designed to make the process of writing and iterating on complex spec suites as painless as possible.  Consider, for example, the `watch` subcommand:

```bash
ginkgo watch -p
```

is all you need to have Ginkgo rerun your suite - in parallel -  whenever it detects a change in the suite or any of its dependencies.  Run that in a terminal while you build out your code and get immediate feedback as you evolve your suite!

### Mental Model: Spec Decorators
We've emphasized throughout this chapter that Ginkgo _assumes_ specs are fully independent.  This assumption enables spec randomization and spec parallelization.

There are some contexts, however, when spec independence is simply too difficult to achieve.  The cost of ensuring specs are independent may be too high.  Or there may be external constraints beyond your control.  When this is the case, Ginkgo allows you to explicitly control how specific specs in your suite must be run.

We'll get into that in the next two sections.  But first we'll need to introduce **Spec Decorators**.

So far we've seen that container nodes and subject nodes have the following signature:

```go
Describe("description", <closure>)
It("description", <closure>)
```

In actuality, the signatures for these functions is actually:

```go
Describe("description", args ...interface{})
It("description", args ...interface{})
```

and Ginkgo provides a number of additional types that can be passed in to container and subject nodes.  We call these types Spec Decorators as they decorate the spec with additional metadata.  This metadata can modify the behavior of the spec at run time.  A comprehensive [reference of all decorators](#decorator-reference) is maintained in these docs.

Some Spec Decorators only apply to a specific node.  For example the `Offset` or `CodeLocation` decorators allow you to adjust the location of a node reported by Ginkgo (this is useful when building shared libraries that generate their own Ginkgo nodes).

Most Spec Decorators, however, get applied to the specs that include the decorated node.  For example, the `Serial` decorator (which we'll see in the next section) instructs Ginkgo to ensure that any specs that include the `Serial` node should only run in series and never in parallel.

So, if `Serial` is applied to a container like so:

```go
Describe("Never in parallel please", Serial, func() {
  It("tests one behavior", func() {
    
  })

  It("tests another behavior", func() {
    
  })
})
```

Then both specs generated by the subject nodes in this container will be marked as `Serial`.  If we transfer the `Serial` decorator to one of the subject nodes, however:

```go
Describe("Never in parallel please",  func() {
  It("tests one behavior", func() {
    
  })

  It("tests another behavior", Serial, func() {
    
  })
})
```

now, only the spec with the "tests another behavior" subject node will be marked Serial.

Another way of capturing this behavior is to say that most Spec Decorators apply hierarchically.  If a container node is decorated with a decorator then the decorator applies to all its child nodes.

One last thing - spec decorators can also decorate [Table Specs](#table-specs):

```go
DescribeTable("Table", Serial, ...)
Entry("Entry", FlakeAttempts(3), ...)
```

will all work just fine.  You can put the decorators anywhere after the description strings.

The [reference](#decorator-reference) clarifies how decorator inheritance works for each decorator and which nodes can accept which decorators.

### Serial Specs

When you run `ginkgo -p` Ginkgo spins up multiple processes and distributes **all** your specs across those processes.  As such, any spec must be able to run in parallel with any other spec.

Sometimes, however, you simply _must_ enforce that a spec runs in series.  Perhaps it is a performance benchmark spec that cannot run in parallel with any other work.  Perhaps it is a spec that is known to exercise an edge case that places some external resource into a known-bad state and, therefore, must be run independently of all other specs.  Perhaps it is simply a spec that is just so resource intensive that it must run alone to avoid exhibiting flaky behavior.

Whatever the reason, Ginkgo allows you to decorate container and subject nodes with `Serial`:

```go

Describe("Something expensive", Serial, func() {
  It("is a resource hog that can't run in parallel", func() {
    ...
  })

  It("is another resource hog that can't run in parallel", func() {
    ...
  })
})
```

Ginkgo will guarantee that these specs will never run in parallel with other specs.

Under the hood Ginkgo does this by running `Serial` at the **end** of the suite on parallel process #1.  When it detects the presence of `Serial` specs, process #1 will wait for all other processes to exit before running the `Serial` specs.

### Ordered Containers

By default Ginkgo does not guarantee the order in which specs run.  As we've seen, `ginkgo --randomize-all` will shuffle the order of all specs and `ginkgo -p` will distribute all specs across multiple workers.  Both operations mean that the order in which specs run cannot be guaranteed.

There are contexts, however, when you must guarantee the order in which a set of specs run.  For example, you may be testing a complex flow of behavior and would like to break your spec up into multiple units instead of having one enormous `It`.  Or you may have to perform some expensive setup for a set of specs and only want to perform that setup **once** _before_ the specs run.

Ginkgo provides `Ordered` containers to solve for these usecases.  Specs in `Ordered` containers are guaranteed to run in the order in which they appear.  Let's pull out an example from before; recall that the following is invalid:

```go
/* === INVALID === */
Describe("checking out a book", func() {
  var book *books.Book
  var err error

  It("can fetch a book from a library", func() {
    book, err = libraryClient.FetchByTitle("Les Miserables")
    Expect(err).NotTo(HaveOccurred())
    Expect(book.Title).To(Equal("Les Miserables"))
  })

  It("can check out the book", func() {
    Expect(library.CheckOut(book)).To(Succeed())
  })

  It("no longer has the book in stock", func() {
    book, err = libraryClient.FetchByTitle("Les Miserables")
    Expect(err).To(MatchError(books.NOT_IN_STOCK))
    Expect(book).To(BeNil())
  })
})
```

These specs break the "declare in container nodes, initialize in setup nodes" principle.  When randomizing specs or running in parallel Ginkgo will not guarantee that these specs run in order.  Because the specs are mutating the same shared set of variables they will behave in non-deterministic ways when shuffled.  In fact, when running in parallel, specs on different parallel processes will be accessing completely different local copies of the closure variables!

When we introduced this example we recommended condensing the tests into a single `It` and using `By` to document the test.  `Ordered` containers provide an alternative that some users might prefer, stylistically:

```go
Describe("checking out a book", Ordered, func() {
  var book *books.Book
  var err error

  It("can fetch a book from a library", func() {
    book, err = libraryClient.FetchByTitle("Les Miserables")
    Expect(err).NotTo(HaveOccurred())
    Expect(book.Title).To(Equal("Les Miserables"))
  })

  It("can check out the book", func() {
    Expect(library.CheckOut(book)).To(Succeed())
  })

  It("no longer has the book in stock", func() {
    book, err = libraryClient.FetchByTitle("Les Miserables")
    Expect(err).To(MatchError(books.NOT_IN_STOCK))
    Expect(book).To(BeNil())
  })
})
```

Here we've decorated the `Describe` container as `Ordered`.  Ginkgo will guarantee that specs in an `Ordered` container will run sequentially, in the order they are written.  Specs in an `Ordered` container may run in parallel with respect to _other_ specs, but they will always run sequentially on the same parallel process.  This allows specs in `Ordered` containers to rely on mutating local closure state.

The `Ordered` decorator can only appear on a container node.  Any container nodes nested within a container node will automatically be considered `Ordered` and there is no way to mark a node within an `Ordered` container as "not `Ordered`".

> Ginkgo did not include support for `Ordered` containers for quite some time.  As you can see `Ordered` containers make it possible to circumvent the "Declare in container nodes, initialize in setup nodes" principle; and they make it possible to write dependent specs  This comes at a cost, of course - specs in `Ordered` containers cannot be fully parallelized which can result in slower suite runtimes.  Despite these cons, pragmatism prevailed and `Ordered` containers were introduced in response to real-world needs in the community.  Nonetheless, we recommend using `Ordered` containers only when needed.

#### Setup in Ordered Containers: BeforeAll and AfterAll

You can include all the usual setup nodes in an `Ordered` container however and they continue to operate in the same way.  `BeforeEach` will run before every spec and `AfterEach` will run after every spec.  This applies to all setup nodes in a spec's hierarchy.  So `BeforeEach`/`AfterEach` nodes that are present outside the `Ordered` container will still run before and after each spec in the container.

There are, however, two new setup node variants that can be used within `Ordered` containers: `BeforeAll` and `AfterAll`.

`BeforeAll` closures will run exactly once before any of the specs within the `Ordered` container.  `AfterAll` closures will run exactly once after the last spec has finished running.  Here's an extension of our earlier example that illustrates how these nodes might be used:

```go
Describe("checking out a book", Ordered, func() {
  var libraryClient *library.Client
  var book *books.Book
  var err error

  BeforeAll(func() {
    libraryClient = library.NewClient()
    Expect(libraryClient.Connect()).To(Succeed())
  })

  It("can fetch a book from a library", func() {
    book, err = libraryClient.FetchByTitle("Les Miserables")
    Expect(err).NotTo(HaveOccurred())
    Expect(book.Title).To(Equal("Les Miserables"))
  })

  It("can check out the book", func() {
    Expect(library.CheckOut(book)).To(Succeed())
  })

  It("no longer has the book in stock", func() {
    book, err = libraryClient.FetchByTitle("Les Miserables")
    Expect(err).To(MatchError(books.NOT_IN_STOCK))
    Expect(book).To(BeNil())
  })

  AfterAll(func() {
    Expect(libraryClient.Disconnect()).To(Succeed())
  })
})
```

here we only set up the `libraryClient` once before all the specs run, and then tear it down once all the specs complete.

`BeforeAll` and `AfterAll` nodes can only be introduced within an `Ordered` container.  `BeforeAll` and `AfterAll` can also be nested within containers that appear in `Ordered` containers - in such cases they will run before/after the specs in that nested container.

As always, you can also use `DeferCleanup`.  Since `DeferCleanup` is context aware, it will detect when it is called in a `BeforeAll` and behave like an `AfterAll` at the same nesting level.  The following is equivalent to the example above:

```go
BeforeAll(func() {
  libraryClient = library.NewClient()
  Expect(libraryClient.Connect()).To(Succeed())  
  DeferCleanup(libraryClient.Disconnect)
})
```

#### Setup around Ordered Containers: the OncePerOrdered Decorator

It's a common pattern to have setup and cleanup code at the outer-most level of a suite that is intended to ensure that every spec runs from with a clean slate.  For example, we may be testing our library service and want to ensure that each spec begins with the same library setup.  We might write something like this at the top level of our suite file:

```go
BeforeEach(func() {
    libraryClient = library.NewClient()
    Expect(libraryClient.Connect()).To(Succeed())

    snapshot := libraryClient.TakeSnapshot()
    DeferCleanup(libraryClient.RestoreSnapshot, snapshot)
})
```

now, every spec will be guaranteed to start with the same initial state and we are free to write our specs without worrying about spec pollution.

This behavior, however, will cause specs in Ordered containers to break.  Consider this set of specs:

```go
Describe("checking out a book", Ordered, func() {
  var book *books.Book
  var err error

  BeforeAll(func() {
    libraryClient.AddBook( &books.Book{
      Title:  "Les Miserables",
      Author: "Victor Hugo",
      Pages:  2783,
    })
  })

  It("can fetch a book from a library", func() {
    book, err = libraryClient.FetchByTitle("Les Miserables")
    Expect(err).NotTo(HaveOccurred())
    Expect(book.Title).To(Equal("Les Miserables"))
  })

  It("can check out the book", func() {
    Expect(library.CheckOut(book)).To(Succeed())
  })

  It("no longer has the book in stock", func() {
    book, err = libraryClient.FetchByTitle("Les Miserables")
    Expect(err).To(MatchError(books.NOT_IN_STOCK))
    Expect(book).To(BeNil())
  })
})
```

Because our outer-most `BeforeEach` runs before _every_ spec, the specs in this ordered container will fail.  Specifically the _first_ spec will pass but subsequent specs will fail as the `BeforeEach` cleans up state between them.

Ginkgo provides a `OncePerOrdered` decorator that can be applied to the `BeforeEach`, `JustBeforeEach`, `AfterEach`, and `JustAfterEach` setup nodes to solve for this usecase.  The `OncePerOrdered` decorator changes the semantics of these `*Each` setup nodes from "run around each spec" to "run around each independent unit".  Individual specs and specs that are in unordered containers constitute independent units and so the `*Each` nodes run around each spec.  However specs in `Ordered` containers behave like a single unit - so `*Each` setup nodes with the `OncePerOrdered` decorator will only run once before the unit begins and/or after the unit completes.  In this way a `BeforeEach` with `OncePerOrdered` that runs before. an Ordered container is semantically equivalent to a `BeforeAll` within that container.

By decorating our outermost `BeforeEach` with `OncePerOrdered`:

```go
BeforeEach(OncePerOrdered, func() {
    libraryClient = library.NewClient()
    Expect(libraryClient.Connect()).To(Succeed()

    snapshot := libraryClient.TakeSnapshot()
    DeferCleanup(libraryClient.RestoreSnapshot, snapshot)
})
```

we retain the existing behavior for the entire suite _and_ get the `BeforeAll`-like behavior we need for our `Ordered` container.

The `OncePerOrdered` decorator modifies the behavior of the `BeforeEach` setup node _only_ for Ordered containers at the same or lower nesting level as the setup node.  Adding a `OncePerOrdered` `BeforeEach` setup node _inside_ an `Ordered` container results in a setup node that behaves like a normal `BeforeEach` - it will run for every spec in the container.  However a container nested _within_ the container will trigger the `OncePerOrdered` behavior and the `BeforeEach` will run just once for the specs within the nested container.

Lastly, the `OncePerOrdered` container cannot be applied to the `ReportBeforeEach` and `ReportAfterEach` nodes discussed below.  In Ginkgo reporting always happens at the granularity of the individual spec.

#### Failure Handling in Ordered Containers

Normally, when a spec fails Ginkgo moves on to the next spec.  This is possible because Ginkgo assumes, by default, that all specs are independent.  However `Ordered` containers explicitly opt in to a different behavior.  Spec independence cannot be guaranteed in `Ordered` containers, so Ginkgo treats failures differently.

When a spec in an `Ordered` container fails all subsequent specs are skipped. Ginkgo will then run any `AfterAll` node closures to clean up after the specs.

You can override this behavior by decorating an `Ordered` container with `ContinueOnFailure`.  This is useful in cases where `Ordered` is being used to provide shared expensive set up for a collection of specs.  When `ContinueOnFailure` is set, Ginkgo will continue running specs even if an earlier spec in the `Ordered` container has failed.  If, however a `BeforeAll` or `OncePerOrdered` `BeforeEach` node has failed then Ginkgo will skip all subsequent specs as the setup for the collection specs is presumed to have failed.

`ContinueOnFailure` can only be applied to the outermost `Ordered` container.  It is an error to apply it to a nested container.

#### Combining Serial and Ordered

To sum up: specs decorated with `Serial` are guaranteed to run in series and never in parallel with other specs.  Specs in `Ordered` containers are guaranteed to run in order sequentially on the same parallel process but may be parallelized with specs in other containers.

You can combine both decorators to have specs in `Ordered` containers run serially with respect to all other specs.  To do this, you must apply the `Serial` decorator to the same container that has the `Ordered` decorator.  You cannot declare a spec within an `Ordered` container as `Serial` independently.

### Filtering Specs

There are several contexts where you may only want to run a _subset_ of specs in a suite.  Perhaps some specs are slow and only need to be run on CI or before a commit.  Perhaps you're only working on a subset of the code and want to run the relevant subset of the specs, or even just one spec.  Perhaps a spec is under development and isn't ready to run yet.  Perhaps a spec should always be skipped if a certain condition is met.

Ginkgo supports all these usecases (and more) through a wide variety of mechanisms to organize and filter specs.  Let's dig into them.

#### Pending Specs
You can mark individual specs, or containers of specs, as `Pending`.  This is used to denote that a spec or its code is under development and should not be run.  None of the other filtering mechanisms described in this chapter can override a `Pending` spec and cause it to run.

Here are all the ways you can mark a spec as `Pending`:

```go
// With the Pending decorator:
Describe("these specs aren't ready for primetime", Pending, func() { ... })
It("needs work", Pending, func() { ... })
It("placeholder", Pending) //note: pending specs don't require a closure
DescribeTable("under development", Pending, func() { ... }, ...)
Entry("this one isn't working yet", Pending)

// By prepending `P` or `X`:
PDescribe("these specs aren't ready for primetime", func() { ... })
XDescribe("these specs aren't ready for primetime", func() { ... })
PIt("needs work", func() { ... })
XIt("placeholder")
PDescribeTable("under development", func() {...}, ...)
XEntry("this one isn't working yet")
```

Ginkgo will never run a pending spec.  If all other specs in the suite pass the suite will be considered successful.  You can, however, run `ginkgo --fail-on-pending` to have Ginkgo fail the suite if it detects any pending specs.  This can be useful on CI if you want to enforce a policy that pending specs should not be committed to source control.

Note that pending specs are declared at compile time.  You cannot mark a spec as pending dynamically at runtime.  For that, keep reading...

#### Skipping Specs
If you need to skip a spec at runtime you can use Ginkgo's `Skip(...)` function.  For example, say we want to skip a spec if some condition is not met.  We could:

```go
It("should do something, if it can", func() {
  if !someCondition {
    Skip("Special condition wasn't met.")
  }
  ...
})
```

This will cause the current spec to skip.  Ginkgo will immediately end execution (`Skip`, just like `Fail`, throws a panic to halt execution of the current spec) and mark the spec as skipped.  The message passed to `Skip` will be included in the spec report.  Note that `Skip` **does not** fail the suite.  Even skipping all the specs in the suite will not cause the suite to fail.  Only an explicitly failure will do so.

You can call `Skip` in any subject or setup nodes.  If called in a `BeforeEach`, `Skip` will skip the current spec.  If called in a `BeforeAll`, `Skip` will skip all specs in the `Ordered` container (however, skipping an individual spec in an `Ordered` container does not skip subsequent specs).  If called in a `BeforeSuite`, `Skip` will skip the entire suite.

You cannot call `Skip` in a container node - `Skip` only applies during the Run Phase, not the Tree Construction Phase.

#### Focused Specs
Ginkgo allows you to `Focus` individual specs, or containers of specs.  When Ginkgo detects focused specs in a suite it skips all other specs and _only_ runs the focused specs.

Here are all the ways you can mark a spec as focused:

```go
// With the Focus decorator:
Describe("just these specs please", Focus, func() { ... })
It("just me please", Focus, func() { ... })
DescribeTable("run this table", Focus, func() { ... }, ...)
Entry("run just this entry", Focus)

// By prepending `F`:
FDescribe("just these specs please", func() { ... })
FIt("just me please", func() { ... })
FDescribeTable("run this table", func() { ... }, ...)
FEntry("run just this entry", ...)
```

doing so instructs Ginkgo to only run the focused specs.  To run all specs, you'll need to go back and remove all the `F`s and `Focus` decorators.

You can nest focus declarations.  Doing so follows a simple rule: if a child node is marked as focused, any of its ancestor nodes that are marked as focused will be unfocused.  This behavior was chosen as it most naturally maps onto the developers intent when iterating on a spec suite.  For example:

```go
FDescribe("some specs you're debugging", func() {
  It("might be failing", func() { ... })
  It("might also be failing", func() { ... })
})
```

will run both specs.  Let's say you discover that the second spec is the one failing and you want to rerun it rapidly as you iterate on the code.  Just `F` it:

```go
FDescribe("some specs you're debugging", func() {
  It("might be failing", func() { ... })
  FIt("might also be failing", func() { ... })
})
```

now only the second spec will run because of Ginkgo's focus rules.

We refer to the focus filtering mechanism as "Programmatic Focus" as the focus declarations are "programmed in" at compile time.  Programmatic focus can be super helpful when developing or debugging a test suite, however it can be a real pain to accidentally commit a focused spec. So...

When Ginkgo detects that a passing test suite has programmatically focused tests it causes the suite to exit with a non-zero status code.  The logs will show that the suite succeeded, but will also include a message that says that programmatic specs were detected.  The non-zero exit code will be caught by most CI systems and flagged, allowing developers to go back and unfocus the specs they committed. 

You can unfocus _all_ specs in a suite by running `ginkgo unfocus`.  This simply strips off any `F`s off of `FDescribe`, `FContext`, `FIt`, etc... and removes `Focus` decorators.

#### Spec Labels
`Pending`, `Skip`, and `Focus` provide ad-hoc mechanisms for filtering suites.  For particularly large and complex suites, however, you may need a more structured mechanism for organizing and filtering specs.  For such usecases, Ginkgo provides labels.

Labels are simply textual tags that can be attached to Ginkgo container and subject nodes via the `Label` decorator.  Here are the ways you can attach labels to a node:

```go
It("is labelled", Label("first label", "second label"), func() { ... })
It("is labelled", Label("first label"), Label("second label"), func() { ... })
```

Labels can container arbitrary strings but cannot contain any of the characters in the set: `"&|!,()/"`.  The labels associated with a spec is the union of all the labels attached to the spec's container nodes and subject nodes. For example:

```go
Describe("Storing books", Label("integration", "storage"), func() {
  It("can save entire shelves of books to the central library", Label("network", "slow", "library storage"), func() {
    // has labels [integration, storage, network, slow, library storage]
  })

  It("cannot delete books from the central library", Label("network", "library storage"), func() {
    // has labels [integration, storage, network, library storage]    
  })

  It("can check if a book is stored in the central library", Label("network", "slow", "library query"), func() {
    // has labels [integration, storage, network, slow, library query]    
  })

  It("can save books locally", Label("local"), func() {
    // has labels [integration, storage, local]    
  })

  It("can delete books locally", Label("local"), func() {
    // has labels [integration, storage, local]        
  })
})
```

The labels associated with a spec are included in any generated reports and are emitted alongside the spec whenever it fails.

The real power, of labels, however, is around filtering.  You can filter by label using via the `ginkgo --label-filter=QUERY` flag.  Ginkgo will accept and parse a simple filter query language with the following operators and rules:

- The `&&` and `||` logical binary operators representing AND and OR operations.
- The `!` unary operator representing the NOT operation.
- The `,` binary operator equivalent to `||`.
- The `()` for grouping expressions.
- All other characters will match as label literals.  Label matches are **case insensitive** and trailing and leading whitespace is trimmed.
- Regular expressions can be provided using `/REGEXP/` notation.

To build on our example above, here are some label filter queries and their behavior:

| Query | Behavior |
| --- | --- |
| `ginkgo --label-filter="integration"` | Match any specs with the `integration` label |
| `ginkgo --label-filter="!slow"` | Avoid any specs labelled `slow` |
| `ginkgo --label-filter="network && !slow"` | Run specs labelled `network` that aren't `slow` |
| `ginkgo --label-filter=/library/` | Run specs with labels matching the regular expression `library` - this will match the three library-related specs in our example.

You can list the labels used in a given package using the `ginkgo labels` subcommand.  This does a simple/naive scan of your test files for calls to `Label` and returns any labels it finds.

You can iterate on different filters quickly with `ginkgo --dry-run -v --label-filter=FILTER`.  This will cause Ginkgo to tell you which specs it will run for a given filter without actually running anything.

If you want to have finer-grained control within a test about what code to run/not-run depending on what labels match/don't match the filter you can perform a manual check against the label-filter passed into Ginkgo like so:

```go
It("can save books remotely", Label("network", "slow", "library query") {
  if Label("performance").MatchesLabelFilter(GinkgoLabelFilter()) {
    exp := gmeasure.NewExperiment()
    // perform some benchmarking with exp...
  }
  // rest of the saving books test
})
```

here `GinkgoLabelFilter()` returns the configured label filter passed in via `--label-filter`.  With a setup like this you could run `ginkgo --label-filter="network && !performance"` - this would select the `"can save books remotely"` spec but not run the benchmarking code in the spec.  Of course, this could also have been modeled as a separate spec with the `performance` label.

Finally, in addition to specifying Labels on subject and container nodes you can also specify suite-wide labels by decorating the `RunSpecs` command with `Label`:

```go
func TestBooks(t *testing.T) {
  RegisterFailHandler(Fail)
  RunSpecs(t, "Books Suite", Label("books", "this-is-a-suite-level-label"))
}
```

Suite-level labels apply to the entire suite making it easy to filter out entire suites using label filters.


#### Location-Based Filtering

Ginkgo allows you to filter specs based on their source code location from the command line.  You do this using the `ginkgo --focus-file` and `ginkgo --skip-file` flags.  Ginkgo will only run specs that are in files that _do_ match the `--focus-file` filter *and* _don't_ match the `--skip-file` filter.  You can provide multiple `--focus-file` and `--skip-file` flags.  The `--focus-file`s will be ORed together and the `--skip-file`s will be ORed together.

The argument passed to `--focus-file`/`--skip-file` is a file filter and takes one of the following forms:

- `FILE_REGEX` - will match specs in files who's absolute path matches the FILE_REGEX.  So `ginkgo --focus-file=foo` will match specs in files like `foo_test.go` or `/foo/bar_test.go`.
- `FILE_REGEX:LINE` - will match specs in files that match FILE_REGEX where at least one node in the spec is constructed at line number `LINE`.
- `FILE_REGEX:LINE1-LINE2` - will match specs in files that match FILE_REGEX where at least one node in the spec is constructed at a line within the range of `[LINE1:LINE2)`.

You can specify multiple comma-separated `LINE` and `LINE1-LINE2` arguments in a single `--focus-file/--skip-file` (e.g. `--focus-file=foo:1,2,10-12` will apply filters for line 1, line 2, and the range [10-12)).  To specify multiple files, pass in multiple `--focus-file` or `--skip-file` flags.

To filter a spec based on its line number you must use the exact line number where one of the spec's nodes (e.g. `It()`) is called.  You can't use a line number that is "close" to the node, or within the node's closure.

#### Description-Based Filtering

Finally, Ginkgo allows you to filter specs based on the description strings that appear in their subject nodes and/or container hierarchy nodes.  You do this using the `ginkgo --focus=REGEXP` and `ginkgo --skip=REGEXP` flags.

When these flags are provided Ginkgo matches the passed-in regular expression against the fully concatenated description of each spec.  For example the spec tree:

```go
Describe("Studying books", func() {
  Context("when the book is long", func() {
    It("can be read over multiple sessions", func() {
      
    })
  })
})
```

will generate a spec with description `"Studying books when the book is long can be read over multiple sessions"`.

When `--focus` and/or `--skip` are provided Ginkgo will _only_ run specs with descriptions that match the focus regexp **and** _don't_ match the skip regexp.  You can provide `--focus` and `--skip` multiple times.  The `--focus` filters will be ORed together and the `--skip` filters will be ORed together.  For example, say you have the following specs:

```go
It("likes dogs", func() {...})
It("likes purple dogs", func() {...})
It("likes cats", func() {...})
It("likes dog fish", func() {...})
It("likes cat fish", func() {...})
It("likes fish", func() {...})
```

then `ginkgo --focus=dog --focus=fish --skip=cat --skip=purple` will only run `"likes dogs"`, `"likes dog fish"`, and `"likes fish"`.

The description-based `--focus` and `--skip` flags were Ginkgo's original command-line based filtering mechanism and will continue to be supported - however we recommend using labels when possible as the label filter language is more flexible and easier to reason about.

#### Combining Filters

To sum up, we've seen that Ginkgo supports the following mechanisms for organizing and filtering specs:

- Specs that are marked as `Pending` at compile-time never run.
- At run-time, specs can be individually skipped by calling `Skip()`
- Specs that are programmatically focused with the `Focus` decorator at compile-time run to the exclusion of other specs.
- Specs can be labelled with the `Label()` decorator.  `ginkgo --label-filter=QUERY` will apply a label filter query and only run specs that pass the filter.
- `ginkgo --focus-file=FILE_FILTER/--skip-file=FILE_FILTER` will filter specs based on their source code location.
- `ginkgo --focus=REGEXP/--skip=REGEXP` will filter specs based on their descriptions.

These mechanisms can all be used in concert.  They combine with the following rules:

- `Pending` specs are always pending and can never be coerced to run by another filtering mechanism.
- Specs that invoke `Skip()` will always be skipped regardless of other filtering mechanisms.
- The CLI based filters (`--label-filter`, `--focus-file/--skip-file`, `--focus/--skip`) **always** override any programmatic focus.
- When multiple CLI filters are provided they are all ANDed together.  The spec must satisfy the label filter query **and** any location-based filters **and** any description based filters.

### Repeating Spec Runs and Managing Flaky Specs

Ginkgo wants to help you write reliable, deterministic, tests.  Flaky specs - i.e. specs that fail _sometimes_ in non-deterministic or difficult to reason about ways - can be incredibly frustrating to debug and can erode faith in the value of a spec suite.

Ginkgo provides a few mechanisms to help you suss out and debug flaky specs.  If you suspect a flaky spec you can rerun a suite repeatedly until it fails via:

```bash
ginkgo --until-it-fails
```

This will compile the suite once and then run it repeatedly, forever, until a failure is detected.  This flag pairs well with `--randomize-all` and `-p` to try and suss out failures due to accidental spec dependencies.

Since `--until-it-fails` runs indefinitely, until a failure is detected, it is not appropriate for CI environments.  If you'd like to help ensure that flaky specs don't creep into your codebase you can use:

```bash
ginkgo --repeat=N
```

to have Ginkgo repeat your test suite up to `N` times or until a failure occurs, whichever comes first.  This is especially valuable in CI environments.

One quick note on `--repeat`: when you invoke `ginkgo --repeat=N` Ginkgo will run your suite a total of `1+N` times.  In this way, `ginkgo --repeat=N` is similar to `go test --count=N+1` **however** `--count` is one of the few `go test` flags that is **not** compatible with Ginkgo suites.  Please use `ginkgo --repeat=N` instead.

Both `--until-it-fails` and `--repeat` help you identify flaky specs early.  Doing so will help you debug flaky specs while the context that introduced them is fresh.

A more granular approach to repeating specs is by decorating individual subject or container nodes with the MustPassRepeatedly(N) decorator:

```go
Describe("Storing books", func() {
  It("can save books to the central library", MustPassRepeatedly(3), func() {
    // this spec has been marked and will be retried up to 3 times
  })

  It("can save books locally", func() {
    // this spec has not been marked and will not be retired
  })
})
```

However,  There are times when the cost of preventing and/or debugging flaky specs simply is simply too high and specs simply need to be retried.  While this should never be the primary way of dealing with flaky specs, Ginkgo is pragmatic about this reality and provides a mechanism for retrying specs.

You can retry all specs in a suite via:

```bash
ginkgo --flake-attempts=N
```

Now, when a spec fails Ginkgo will not automatically mark the suite as failed.  Instead it will attempt to rerun the spec up to `N` times.  If the spec succeeds during a retry, Ginkgo moves on and marks the suite as successful but reports that the spec needed to be retried.

A more granular approach is also provided for this functionality with the use of the `FlakeAttempts(N)` decorator:

```go
Describe("Storing books", func() {
  It("can save books to the central library", FlakeAttempts(3), func() {
    // this spec has been marked as flaky and will be retried up to 3 times
  })

  It("can save books locally", func() {
    // this spec must always pass on the first try
  })
})
```

Ginkgo's retry behavior generally works as you'd expect with most specs, however there is some complexity when `FlakeAttempts` is applied to `Ordered` containers.  In brief, Ginkgo generally guarantees that `BeforeAll` and `AfterAll` node closures only run once - but `FlakeAttempts` can modify this behavior.  If a failure occurs within a subject node in an `Ordered` container (i.e. in an `It`) then Ginkgo will rerun that `It` but not the `BeforeAll` or `AfterAll`.  However, if a failure occurs in a `BeforeAll` Ginkgo will immediately run the `AfterAll` (to clean up) then rerun the `BeforeAll`.

Stepping back - it bears repeating: you should use `FlakeAttempts` judiciously.  The best approach to managing flaky spec suites is to debug flakes early and resolve them.  More often than not they are telling you something important about your architecture.  In a world of competing priorities and finite resources, however, `FlakeAttempts` provides a means to explicitly accept the technical debt of flaky specs and move on.

### Getting Visibility Into Long-Running Specs
Ginkgo is often used to build large, complex, integration suites and it is a common - if painful - experience for these suites to run slowly.  Ginkgo provides numerous mechanisms that enable developers to get visibility into what part of a suite is running and where, precisely, a spec may be lagging or hanging.

Ginkgo can provide a **Progress Report** of what is currently running in response to the `SIGINFO` and `SIGUSR1` signals.  The Progress Report includes information about which node is currently running and the exact line of code that it is currently executing, along with any relevant goroutines that were launched by the spec.  The report also includes the 10 most recent lines written to the `GinkgoWriter`.  A developer waiting for a stuck spec can get this information immediately by sending either the `SIGINFO` or `SIGUSR1` signal (on MacOS/BSD systems, `SIGINFO` can be sent via `^T` - making it especially convenient; if you're on linux you'll need to send `SIGUSR1` to the actual test process spawned by `ginkgo` - not the `ginkgo` cli process itself).

These Progress Reports can also show you a preview of the running source code, but only if Ginkgo can find your source files.  If need be you can tell Ginkgo where to look for source files by specifying `--source-root`.

Finally - you can instruct Ginkgo to provide  Progress Reports automatically whenever a node takes too long to complete.  You do this by passing the `--poll-progress-after=INTERVAL` flag to specify how long Ginkgo should wait before emitting a progress report.  Once this interval is passed Ginkgo can periodically emit Progress Reports - the interval between these reports is controlled via the `--poll-progress-interval=INTERVAL` flag.  By default `--poll-progress-after` is set to `0` and so Ginkgo does not emit Progress Reports.  

You can override the global setting of `poll-progess-after` and `poll-progress-interval` on a per-node basis by using the `PollProgressAfter(INTERVAL)` and `PollProgressInterval(INTERVAL)` decorators.  A value of `0` will explicitly turn off Progress Reports for a given node regardless of the global setting.

All Progress Reports generated by Ginkgo - whether interactively via `SIGINFO/SIGUSR1` or automatically via the `PollProgressAfter` configuration - also appear in Ginkgo's [machine-readable reports](#generating-machine-readable-reports).

In addition to these formal Progress Reports, Ginkgo tracks whenever a node begins and ends.  These node `> Enter` and `< Exit` events are usually only logged in the spec's timeline when running with `-vv`, however you can turn them on for other verbosity modes using the `--show-node-events` flag.

#### Attaching Additional Information to Progress Reports

**This section describes an experimental feature and the public-facing interface may change in a future minor version of Ginkgo**

Ginkgo also allows you to attach Progress Report providers to provide additional information when a progress report is generated.  For example, these could query the system under test for diagnostic information about its internal state and report back.  You attach these providers via `AttachProgressReporter`.  For example:

```go
AttachProgressReporter(func() string {
  libraryState := library.GetStatusReport()
  return fmt.Sprintf("%s: %s", library.ClientID, libraryState.Summary)
})
```

`AttachProgressReporter` returns a `cancel` func that you can call to unregister the progress reporter.  This allow you to do things like:

```go
BeforeEach(func() {
  library = libraryClient.ConnectAs("Jean ValJean")
  
  //we attach a progress reporter and can trust that it will be cleaned up after the spec runs
  DeferCleanup(AttachProgressReporter(func() string {
    libraryState := library.GetStatusReport()
    return fmt.Sprintf("%s: %s", library.ClientID, libraryState.Summary)
  }))
})
```

Note that the functions called by `AttachProgressReporter` must not block.  Ginkgo currently has a hard-coded 5 second limit.  If all attached progress reporters take longer than 5 seconds to report back, Ginkgo will move on so as to prevent the suite from blocking.


### Spec Timeouts and Interruptible Nodes

Sometimes specs get stuck.  Perhaps a network call is running slowly; or a newly introduced bug has caused an asynchronous process the test is relying on to hang.  It's important, in such cases, to be able to set a deadline for a given spec or node and require the spec/node to complete before the deadline has elapsed.

Ginkgo supports this through a collection of timeout-related decorators and the notion of **Interruptible Nodes**.

#### Interruptible Nodes and SpecContext

We've seen [how Ginkgo handles failures](#mental-model-how-ginkgo-handles-failure) when an explicit (or implicit, if using a matcher library) call to `Fail` takes place: `Fail` raises a panic to indicate a failure and immediately exit the current node.  Such failures emanate from _within_ a node's running goroutines.

However, in the context of a timeout the cause of failure comes from _outside_ a node's running goroutine.  Once a deadline has passed Ginkgo can mark a spec as failed, but also needs a mechanism to notify the current node's running goroutine that it is timed to stop trying and exit.  Ginkgo supports this through the notion of an Interruptible Node.

A node is considered interruptible if it has a callback that takes either a `SpecContext` or `context.Context` object:

```go
It("can save books", func(ctx SpecContext) {
      book := &books.Book{
        Title:  "Les Miserables",
        Author: "Victor Hugo",
        Pages:  2783,
      }

      Expect(libraryClient.SaveBook(ctx, book)).To(Succeed())
      Expect(libraryClient.ListBooks(ctx)).To(ContainElement(book))
})
```

when such a node is detected Ginkgo will automatically supply a `SpecContext` object.  This `SpecContext` object satisfies the `context.Context` interface and can be used anywhere a `context.Context` object is used.  When a spec times out or is interrupted by the user (see below) Ginkgo will cancel the `SpecContext` to signal to the spec that it is time to exit. In the case above, it is assumed that `libraryClient` knows how to return once `ctx` is cancelled.

Only setup and subjects nodes can be interruptible.  Container nodes cannot be interrupted.

As a more explicit example, here's a (contrived) example to illustrate a timeout in action:

```go
It("likes to sleep in", func(ctx context.Context) {
  select {
  case <-ctx.Done():
    return
  case <-time.After(time.Hour)
  }
}, NodeTimeout(time.Second))
```

rather than hang for an hour, this spec will exit (and be marked as failed due to a timeout), soon after the one second NodeTimeout deadline elapses.  When the deadline elapses Ginkgo takes a [Progress Report](#getting-visibility-into-long-running-specs) snapshot to document where, exactly, the goroutine was stuck when the timeout occurred.  Because it is important to take the snapshot just before the context is cancelled, Ginkgo manages the timing of the cancellation directly and does not rely on a `context.WithDeadline()`-flavored context.  As a result calling `ctx.Deadline()` will not return the deadline of the node in question - however you can trust that `ctx.Done()` will be closed on time.

Note that you are allowed to pass in either `SpecContext` or the more canonical `context.Context` as shown in this example.  The `SpecContext` object has a few additional methods attached to it and serves as an extension point for third-party libraries (including Gomega).  You are free to wrap `SpecContext` however you wish (e.g. via `context.WithValue(ctx, "key", "value")`) - Ginkgo will continue to cancel the resulting context at the correct time and third-party libraries will still have access to the full-blown `SpecContext` object as it is stored as a value within the context with the `"GINKGO_SPEC_CONTEXT"` key.

#### The SpecTimeout and NodeTimeout Decorators

We saw a quick preview of the `NodeTimeout` decorator above.  This applies a timeout deadline to a single node and can be applied to any interruptible node.  Once the `NodeTimeout` elapses, Ginkgo will cancel the interruptible node's context.

`SpecTimeout` is similar to `NodeTimeout` but can only decorate `It` nodes and acts as a deadline for the lifecycle of the spec.  That is, all nodes associated with the spec need to complete before `SpecTimeout` expires.  Note that individual nodes within the spec can also have a `NodeTimeout` - however that timeout can only ever be more stringent than the deadline implied by `SpecTimeout`.  Here's a simple example:

```go
Describe("interacting with the library", func() {
  BeforeEach(func(ctx SpecContext) {
      libraryClient = library.NewClient()
      Expect(libraryClient.Connect(ctx)).To(Succeed())
  }, NodeTimeout(time.Millisecond * 500))

  It("can save books", func(ctx SpecContext) {
        book := &books.Book{
          Title:  "Les Miserables",
          Author: "Victor Hugo",
          Pages:  2783,
        }

        Expect(libraryClient.SaveBook(ctx, book)).To(Succeed())
        Expect(libraryClient.ListBooks(ctx)).To(ContainElement(book))
  }, SpecTimeout(time.Second*2))

  AfterEach(func(ctx SpecContext) {
      Expect(libraryClient.Cleanup(ctx, "books")).To(Succeed())
  }, NodeTimeout(time.Second))
})
```

here the total runtime of `BeforeEach`, `It`, and `AfterEach` must be less than the `SpecTimeout` of 2 seconds.  In addition, the `BeforeEach` callback must exit within 500ms and the `AfterEach` after 1 second.

When a `SpecTimeout` expires the current node is interrupted (i.e. it's context is cancelled) and Ginkgo proceeds to run any associated clean up nodes (i.e. any `AfterEach`, `AfterAll`, and `DeferCleanup` nodes) subject to their own `NodeTimeout`s.  This is because cleanup is considered an essential part of the spec lifecycle and must not be skipped if possible.  Thus the `SpecTimeout` is not a strict guarantee on the runtime of a spec but rather a threshold at which the spec will be considered failed.

Currently, `SpecTimeout` and `NodeTimeout` cannot be applied to container nodes.

#### Mental Model: The Life-cycle of Interruptions and the GracePeriod Decorator

Interruptible nodes and the `SpecTimeout`/`NodeTimeout` decorators allow you to enforce deadlines at a granular per-spec/per-node level.  But what happens when a node fails to return after its `SpecContext` is cancelled.  What happens if it's _really_ stuck?

When a node times out Ginkgo cancels its `SpecContext` and then waits for it to exit for a period of time called the **Grace Period**.  If the node exits within the Grace Period Ginkgo will continue with the relevant portions of the spec (specifically, Ginkgo will behave as if a failure occurred and skip any subsequent setup or subject nodes and, instead, simply run through the cleanup nodes).  If, however, the node does not exit within the Grace Period, Ginkgo will allow the node to _leak_ and proceed with the relevant portion of the spec.

A leaked node continues to run in the background - and this can, potentially, be a source of confusion for future specs as a leaked node can interact with Ginkgo's global callbacks (e.g. `Fail`, or `AddReportEntry`) and pollute the currently running spec.  For this reason it's important to write specs that respond to cancelled contexts and exit as soon as possible.  Nonetheless, Ginkgo takes the opinion that it is better to potentially leak a node and continue with the suite than to allow the suite to hang forever.  When a node is leaked due to a timeout and elapsed Grace Period Ginkgo will emit a message stating that the node has leaked along with a [Progress Report](#getting-visibility-into-long-running-specs) that shows the currently running code in the leaked goroutine.

The Grace Period can be configured on a per-node basis using the `GracePeriod` decorator (which can be applied to any interruptible node) and/or globally with the `--grace-period=<DURATION>` cli flag.

One final, somewhat complex, note on timeouts and the Grace Period.  As mentioned above (and as you'll see below) when a `SpecTimeout` or user-initiated interrupt occurs Ginkgo will interrupt the current node by cancelling its context, and then run any relevant cleanup nodes.  These cleanup nodes **must** run to ensure specs clean up after themselves, however they are now running in a setting where the spec is out of time and needs to wind down as soon as possible.  To facilitate this Ginkgo applies a timeout to each of these remaining nodes as follows:

- If the remaining node is interruptible and has a `NodeTimeout`, Ginkgo uses that `NodeTimeout` to set a deadline for the node.  If the deadline expires then a Grace Period applies (either the node's `GracePeriod` or the global `--grace-period`) before Ginkgo leaks the node and moves on.
- If the remaining node is interruptible and **does not** have a `NodeTimeout`, Ginkgo uses the Grace Period to set a deadline for the node.  If the deadline expires then a second Grace Period applies before Ginkgo leaks the node and moves on.
- If the remaining node is **not** interruptible, Ginkgo will give the node a single Grace Period to complete and exit.  In this case since it cannot be interrupted Ginkgo will simply leak the node after one Grace Period.

#### Using SpecContext with Gomega's Eventually

Gomega provides `Eventually` to allow you to poll an object or function repeatedly until a Gomega matcher is satisfied.  `Eventually` integrates cleanly with interruptible nodes by accepting a `SpecContext`/`context.Context` parameter.  This allows you, for example, to enforce a single timeout across a set of polling assertions:

```go
Describe("interacting with the library", func() {
  BeforeEach(func(ctx SpecContext) {
      libraryClient = library.NewClient()
      // we use eventually here to keep trying until we succeed (e.g. perhaps the server is still spinning up)
      Eventually(func() error {
        return libraryClient.Connect(ctx)
      }).WithContext(ctx).Should(Succeed())
  }, NodeTimeout(time.Millisecond * 500))

  It("can save books", func(ctx SpecContext) {
        numBooks := libraryClient.CountBooks(ctx)

        book := &books.Book{
          Title:  "Les Miserables",
          Author: "Victor Hugo",
          Pages:  2783,
        }

        Expect(libraryClient.SaveBook(ctx, book)).To(Succeed())

        // perhaps the library is a distributed system that only converges eventually
        Eventually(func() ([]*books.Book, error) {
          return libraryClient.ListBooksByAuthor(ctx, "Victor Hugo")
        }).WithContext(ctx).Should(ContainElement(book))
        Eventually(func() int {
          return libraryClient.CountBooks(ctx)
        }).WithContext(ctx).Should(Equal(numBooks + 1))
  }, SpecTimeout(time.Second*2))

  AfterEach(func(ctx SpecContext) {
      Expect(libraryClient.Cleanup(ctx, "books")).To(Succeed())

      // let's make sure we eventually clean up
      Eventually(func() int {
        return libraryClient.CountBooks(ctx)
      }).WithContext(ctx).Should(Equal(0))
  }, NodeTimeout(time.Second))
})
```

now, if any of the node contexts are cancelled (either due to a timeout or an interruption) `Eventually` will exit immediately with an appropriate failure.  We've written out this example in full to show how the context is passed _both_ to `Eventually` via `.WithContext(ctx)` _and_ to the various client methods that take a context.  For example:

```go
Eventually(func() ([]*books.Book, error) {
  return libraryClient.ListBooksByAuthor(ctx, "Victor Hugo")
}).WithContext(ctx).Should(ContainElement(book))
```  

This is important as the cancellation of the context needs to cause `ListBooksByAuthor` to exit _and_ `Eventually` to stop retrying.  This is a common-enough pattern that Gomega provides some short hand.  If you pass `Eventually` a function that takes a `context.Context` as its first parameter, Gomega will pass in the context attached via `.WithContext()` automatically.  This allows us to turn statements like this:

```go
Eventually(func() error {
  return libraryClient.Connect(ctx)
}).WithContext(ctx).Should(Succeed())
```

into:

```go
Eventually(libraryClient.Connect).WithContext(ctx).Should(Succeed())
```

This also works well with Gomega's `.WithArguments(...)` method which allows us to turn statements like this:
```go
Eventually(func() ([]*books.Book, error) {
  return libraryClient.ListBooksByAuthor(ctx, "Victor Hugo")
}).WithContext(ctx).Should(ContainElement(book))
```  

into:
```go
Eventually(libraryClient.ListBooksByAuthor).WithContext(ctx).WithArguments("Victor Hugo").Should(ContainElement(book))
```  

all told this allows us to rewrite our example as:

```go
Describe("interacting with the library", func() {
  BeforeEach(func(ctx SpecContext) {
      libraryClient = library.NewClient()
      // we use eventually here to keep trying until we succeed (e.g. perhaps the server is still spinning up)
      Eventually(libraryClient.Connect).WithContext(ctx).Should(Succeed())
  }, NodeTimeout(time.Millisecond * 500))

  It("can save books", func(ctx SpecContext) {
        numBooks := libraryClient.CountBooks(ctx)

        book := &books.Book{
          Title:  "Les Miserables",
          Author: "Victor Hugo",
          Pages:  2783,
        }

        Expect(libraryClient.SaveBook(ctx, book)).To(Succeed())

        // perhaps the library is a distributed system that only converges eventually
        Eventually(libraryClient.ListBooksByAuthor).WithContext(ctx).WithArguments("Victor Hugo").Should(ContainElement(book))
        Eventually(libraryClient.CountBooks).WithContext(ctx).Should(Equal(numBooks + 1))
  }, SpecTimeout(time.Second*2))

  AfterEach(func(ctx SpecContext) {
      Expect(libraryClient.Cleanup(ctx, "books")).To(Succeed())
      // let's make sure we eventually clean up
      Eventually(libraryClient.CountBooks).WithContext(ctx).Should(Equal(0))
  }, NodeTimeout(time.Second))
})
```

which is much cleaner!

Lastly, there's another reason you'll want to pass the `SpecContext` to `Eventually`.  Gomega uses the extension point provided by `SpecContext` to provide additional information whenever a [Progress Report](#getting-visibility-into-long-running-specs) is requested.  This allows you to get deeper visibility into the state of a running `Eventually` simply by requesting a Progress Report (either by sending a `SIGINFO`/`SIGUSR1` or by using the `PollProgressAfter` decorator).  For example, imagine this assertion:

```go
Eventually(libraryClient.ListBooksByAuthor).WithContext(ctx).WithArguments("Victor Hugo").Should(ContainElement(book))
```

is stuck waiting.  A generated Progress Report would show the current state of the `Eventually` assertion which would include the failure message associated with the `ContainElement` matcher:

```
Expected
    <[]*Book | len:3, cap:3>: [
        "The Hunchback of Notre Dame",
        "Notre Dame de Paris",
        "L'Homme Qui Rit",
    ]
to contain element matching
    <*Book>: Les Miserables
```

#### Interruptible Node Function Signatures: A Quick Reference

Most Ginkgo nodes can be made interruptible.  **Setup** and **Subject** nodes typically take a simple `func() {}` but can be made interruptible like so:

```go
BeforeEach(func(ctx SpecContext) {
  ...
})

It("is interruptible", func (ctx context.Context) {
  ...
})

AfterEach(func(ctx context.Context) {
  ...
})
```

Note that both `context.Context` and `SpecContext` are valid.

In addition, the **Suite Setup** nodes can be made interruptible.  In the case of `BeforeSuite`, `AfterSuite`, and `SynchronizedAfterSuite` this is similar to the setup and subject nodes above:

```go
BeforeSuite(func(ctx SpecContext) {
  ...
})

AfterSuite(func(ctx SpecContext) {
  ...
})

SynchronizedAfterSuite(func(ctx SpecContext) {
  ...
}, func(ctx context.Context) {
  ...
})
```

Note that the `SynchronizedAfterSuite` takes two functions - the first runs on all processes, the second only on process 1.  Each of these can be optionally passed a context, making them independently interruptible (or not, if no context is passed in).

`SynchronizedBeforeSuite` also support independently interruptible functions.  [Recall](#parallel-suite-setup-and-cleanup-synchronizedbeforesuite-and-synchronizedaftersuite) that the two callbacks associated with `SynchronizedBeforeSuite` can optionally return and receive a `[]byte` array to facilitate communication between the primary process and the other parallel processes.  This optionality expands the set of possible interruptible signatures.  For example:

```go
SynchronizedBeforeSuite(func(ctx SpecContext) {
  ...
}, func(ctx SpecContext) {
  ...
})

SynchronizedBeforeSuite(func(ctx SpecContext) []byte {
  return []byte{"data"}
}, func(ctx SpecContext, b []byte) {
  ...
})
```
are all valid interruptible signatures.  Of course you can specify `context.Context` instead and can mix-and-match interruptibility between the two functions.

Currently the **Reporting** nodes (`ReportAfterEach`, `ReportAfterSuite`, and `ReportBeforeEach`) cannot be made interruptible and do not accept callbacks that receive a `SpecContext`.  This may change in a future release of Ginkgo (in a backward compatible way).

As for **Container** nodes, since these run during the Tree Construction Phase they cannot be made interruptible and so do not accept functions that expect a context.  And since the `By` annotation is simply syntactic sugar enabling more detailed spec documentation, any callbacks passed to `By` cannot be independently marked as interruptible (you should, instead, use the `context` passed into the node that you're calling `By` from).

Finally, there *are* two other Ginkgo constructs that can be made interruptible and their flexibility warrants some specific coverage in this section: `DeferCleanup` and `DescribeTable`.

Recall that [`DeferCleanup`](#cleaning-up-our-cleanup-code-defercleanup) effectively generates a dynamic `After*` node for your spec.  It's important to note that the lifecycle of this generated node is different from the lifecycle of the node in which `DeferCleanup` was called.  Consider, our earlier example:

```go
Describe("interacting with the library", func() {
  BeforeEach(func(ctx SpecContext) {
      libraryClient = library.NewClient()
      Expect(libraryClient.Connect(ctx)).To(Succeed())
  }, NodeTimeout(time.Millisecond * 500))

  It("can save books", func(ctx SpecContext) {
    ...
  }, SpecTimeout(time.Second*2))

  AfterEach(func(ctx SpecContext) {
      Expect(libraryClient.Cleanup(ctx, "books")).To(Succeed())
  }, NodeTimeout(time.Second))
})
```

We can tidy things up by replacing the `AfterEach` with a `DeferCleanup` in the `BeforeEach`:

```go
/* === INVALID === */
Describe("interacting with the library", func() {
  BeforeEach(func(ctx SpecContext) {
      libraryClient = library.NewClient()
      Expect(libraryClient.Connect(ctx)).To(Succeed())
      DeferCleanup(func() {
        libraryClient.Cleanup(ctx, "books")
      })
  }, NodeTimeout(time.Millisecond * 500))

  It("can save books", func(ctx SpecContext) {
    ...
  }, SpecTimeout(time.Second*2))
})
```

however we've committed a subtle error.  We've captured the `BeforeEach` `SpecContext` and passed it in to the `DeferCleanup` function.  However the `DeferCleanup` function will only run _after_ the `BeforeEach` completes (and its `SpecContext` has been cancelled) - as a result `libraryClient.Cleanup` will always receive a cancelled context.

Moreover, we want to preserve the fact that our `BeforeEach` has a 500ms timeout whereas our clean up code has a separate 1 second timeout.

The correct way to write this is to make the `DeferCleanup` node interruptible and decorate it with its own `NodeTimeout`:

```go
Describe("interacting with the library", func() {
  BeforeEach(func(ctx SpecContext) {
      libraryClient = library.NewClient()
      Expect(libraryClient.Connect(ctx)).To(Succeed())
      DeferCleanup(func(ctx SpecContext) {
        libraryClient.Cleanup(ctx, "books")
      }, NodeTimeout(time.Second))
  }, NodeTimeout(time.Millisecond * 500))

  It("can save books", func(ctx SpecContext) {
    ...
  }, SpecTimeout(time.Second*2))
})
```

(as before, we could have used `context.Context` instead of `SpecContext`).  This is looking better but we can do more.  Recall that `DeferCleanup` can take additional parameters at invocation to pass along to its function.  If the first argument that its function expects is a context, `DeferCleanup` will automatically treat the function as interruptible and provide it with a `SpecContext`.  This allows us to write:

```go
Describe("interacting with the library", func() {
  BeforeEach(func(ctx SpecContext) {
      libraryClient = library.NewClient()
      Expect(libraryClient.Connect(ctx)).To(Succeed())
      DeferCleanup(libraryClient.Cleanup, "books", NodeTimeout(time.Second))
  }, NodeTimeout(time.Millisecond * 500))

  It("can save books", func(ctx SpecContext) {
    ...
  }, SpecTimeout(time.Second*2))
})
```

As an aside, if you _don't_ want Ginkgo to inject `SpecContext` you can, instead, provide your own context.  Here, for example, we avoid making the `DeferCleanup` interruptible by passing in our own context:

```go
      DeferCleanup(libraryClient.Cleanup, "books", NodeTimeout(time.Second)) //interruptible
      DeferCleanup(libraryClient.Cleanup, context.Background(), "books") //*not* interruptible
```

The heuristic here is simple: if the function passed to `DeferCleanup` takes a `context.Context` as its first argument and a context is passed in as the first parameter to `DeferCleanup` then the function is not interruptible and the passed-in context is used.  Otherwise the function is considered interruptible and a `SpecContext` is passed-in instead.  If, instead, the first argument to the function is specifically a `SpecContext` then the function is always considered interruptible regardless of what the subsequent parameters are.

`DescribeTable` behaves similarly.  You can make the `It`s generated by your table interruptible by passing a `SpecContext` or `context.Context` as the first argument to the table function:

```go
DescribeTable("shelf counts", 
  func(ctx SpecContext, shelf string, count int) { // or context.Context instead
    Expect(libraryClient.Count(ctx, shelf)).To(Equal(count))
  },
  Entry("books on shelf A", "A", 17, NodeTimeout(time.Second)),
  Entry("books on shelf B", "B", 20, NodeTimeout(time.Second)),
  ...
)
```

Note that the `NodeTimeout` decorators go on the individual entries.

If you also want to specify a [custom entry description generator](#generating-entry-descriptions) you can pass in a function that takes the non-`SpecContext` parameters and returns `string`:


```go
DescribeTable("shelf counts", 
  func(ctx SpecContext, shelf string, count int) { // or context.Context instead
    Expect(libraryClient.Count(ctx, shelf)).To(Equal(count))
  },
  func(shelf string, _ int) string {
    return fmt.Sprintf("books on shelf %s", shelf)
  }
  Entry("books on shelf A", "A", 17, NodeTimeout(time.Second)),
  Entry("books on shelf B", "B", 20, NodeTimeout(time.Second)),
  ...
)
```

As with `DeferCleanup`, Ginkgo will detect if the entry parameter list provides a context.  Doing so will avoid treating the function as interruptible and use the provided context instead.  For example:

```go
DescribeTable("contrived context-value example", 
  func(ctx context.Context, result string) { //but **NOT** SpecContext
    Expect(libraryClient.Encabulate(ctx)).To(Equal(result))
  },
  Entry("with a generic context", context.Background(), "Nothin"),
  Entry("with a context with a magical value", context.WithValue(context.Background(), "magic", "word"), "Geminio"),
  ...
)
```

#### SpecContext and Progress Reports

`SpecContext` provides an extension point that enables consumers to attach additional information to Progress Reports that Ginkgo generates.  This is accomplished by calling `ctx.AttachProgressReporter(f)` where `f` has the signature `func() string`.  Once attached, the function will be called whenever a Progress Report needs to be generated (e.g. due to a user request via `SIGINFO`/`SIGUSR1` or via an interrupt or timeout).  `ctx.AttachProgressReporter` returns a detach function with signature `func()` that can be called to detach the attached progress reporter.  Because these progress reporters are attached to the passed-in `SpecContext` they only remain attached for the lifecycle of the context: i.e. the current node.

While users of Ginkgo can provide their own custom progress reporters the intent behind this extension point is to allow deeper integration between Ginkgo and third-party libraries, specifically Gomega.  Whenever Gomega's `Eventually` is passed a `SpecContext` it automatically registers a progress reporter.  This reporter will provide the latest state of the `Eventually` matcher - enabling users to get insight into where and why an `Eventually` might be stuck simply by asking for a Progress Report.

### Interrupting, Aborting, and Timing Out Suites

We've seen how nodes can be marked as interruptible and focused on how Ginkgo can apply deadlines to individual nodes and interrupt them when a timeout expires.  Ginkgo also provides a few, related, mechanisms for interrupting a _suite_ before all specs have naturally completed. 

First, you can signal to a suite that it must stop running by sending a `SIGINT` or `SIGTERM` signal to the running ginkgo process (or just hit `^C`).

Second, you can also specify a timeout on a suite (or set of suites) via:

```bash
ginkgo --timeout=duration
```

where `duration` is a parseable go duration string (the default is `1h` -- one hour).  When running multiple suites Ginkgo will ensure that the total runtime of _all_ the suites does not exceed the specified timeout.

Finally, you can abort a suite from within the suite by calling `Abort(<reason>)`.  This will immediately end the suite and is the programmatic equivalent of sending an interrupt signal to the test process.

All three mechanisms have same effects.  If the currently running node is interruptible, then Ginkgo will:

- Emit a [Progress Report](#getting-visibility-into-long-running-specs) for the current spec as possible.
- Interrupt the current node by cancelling its SpecContext...
- ...then wait up to the Grace Period for the node to exit.  If it does not, then Ginkgo will leak the node and proceed.
- Ginkgo will then run any clean-up and reporting nodes (`AfterEach`, `JustAfterEach`, `AfterAll`, `DeferCleanup`, `ReportAfterEach` code, etc.) for the current spec...
- ...and skip any subsequent specs.
- Ginkgo will then run any `AfterSuite` and `ReportAfterSuite` nodes.
- And finally, it will exit, marking the suite as failed.

If the currently running node is **not** interruptible then Ginkgo will simply leak the node and proceed with the cleanup nodes.

Once a suite is interrupted by one of these mechanisms any subsequent cleanup nodes that run will be subject to the following timeout behavior:

- If the cleanup node is interruptible and has a `NodeTimeout`, Ginkgo uses that `NodeTimeout` to set a deadline for the node.  If the deadline expires then a Grace Period applies (either the node's `GracePeriod` or the global `--grace-period`) before Ginkgo leaks the node and moves on.
- If the cleanup node is interruptible and **does not** have a `NodeTimeout`, Ginkgo uses the Grace Period to set a deadline for the node.  If the deadline expires then a second Grace Period applies before Ginkgo leaks the node and moves on.
- If the cleanup node is **not** interruptible, Ginkgo will give the node a single Grace Period to complete and exit.  In this case since it cannot be interrupted Ginkgo will simply leak the node after one Grace Period.

In short, Ginkgo does its best to cleanup and emit as much information as possible about the suite before shutting down... while also ensuring that the suite doesn't hang forever should a cleanup node get stuck.

A single interrupt (e.g. `SIGINT`/`SIGTERM`) interrupts the current running node and proceeds to perform cleanup.  If you want to skip cleanup you can send a second interrupt - this will still run reporting nodes in an effort to ensure the generated reports are not corrupted.  If you want to skip the reporting nodes and bail immediately, send a third interrupt signal.

If you want to get information about what is currently running in a suite _without_ interrupting it, check out the [Getting Visibility Into Long-Running Specs](#getting-visibility-into-long-running-specs) section above.

### Running Multiple Suites

So far we've covered writing and running specs in individual suites.  Of course, the `ginkgo` CLI also supports running multiple suites with a single invocation on the command line.  We'll close out this chapter on running specs by covering how Ginkgo runs multiple suites.

When you run `ginkgo` the Ginkgo CLI first looks for a spec suite in the current directory.  If it finds one it runs `go test -c` to compile the suite and generate a `.test` binary.  It then invokes the binary directly, passing along any necessary flags to correctly configure it.  In the case of parallel specs, the CLI will configure and spin up multiple copies of the binary and act as a server to coordinate running specs in parallel.

You can have `ginkgo` run multiple spec suites by pointing it at multiple package locations (i.e. directories) like so:

```bash
ginkgo <flags> path/to/package-1 path/to/package-2 ...
```

Ginkgo will enter each of these directory and look for a spec suite.  If it finds one it will compile the suite and run it.  Note that you need to include any `ginkgo` flags **before** the list of packages.

You can also have `ginkgo` recursively find and run all spec suites within the current directory:

```bash
ginkgo -r

- or, equivalently,

ginkgo <flags> ./...
```

Now Ginkgo will walk the file tree and search for spec suites.  It will compile any it finds and run them.

When there are multiple suites to run Ginkgo attempts to compile the suites in parallel but **always** runs them sequentially.  You can control the number of parallel compilation workers using the `ginkgo --compilers=N` flag, by default Ginkgo runs as many compilers as you have cores.

Ginkgo provides a few additional configuration flags when running multiple suites.

You can ask Ginkgo to skip certain packages via:

```bash
ginkgo -r --skip-package=list,of,packages
```

`--skip-package` takes a comma-separated list of package names.  If any part of the package's **path** matches one of the entries in this list that package is skipped: it is not compiled and it is not run.

By default, Ginkgo runs suites in the order it finds them.  You can have Ginkgo randomize the order in which suites run withL

```bash
ginkgo -r --randomize-suites
```

Finally, Ginkgo's default behavior when running multiple suites is to stop execution after the first suite that fails.  (Note that Ginkgo will run _all_ the specs in that suite unless `--fail-fast` is specified.)  You can alter this behavior and have Ginkgo run _all_ suites regardless of failure with:

```bash
ginkgo -r --keep-going
```

As you can see, Ginkgo provides several CLI flags for controlling how specs are run.  Be sure to check out the [Recommended Continuous Integration Configuration](#recommended-continuous-integration-configuration) section of the patterns chapter for pointers on which flags are best used in CI environments.

## Reporting and Profiling Suites
The previous two chapters covered how Ginkgo specs are written and how Ginkgo specs run.  This chapter is all about output.  We'll cover how Ginkgo reports on spec suites and how Ginkgo can help you profile your spec suites.

### Controlling Ginkgo's Output
Ginkgo emits a real-time report of the progress of your spec suite to the console while running your specs.  A green dot is emitted for each successful spec and a red `F`, along with failure information and the spec's [timeline](#mental-model-spec-timelines), is emitted for each unsuccessful spec.

There are several CLI flags that allow you to tweak this output:

#### Controlling Verbosity
Ginkgo has four verbosity settings: succinct (the default when running multiple suites), normal (the default when running a single suite), verbose, and very-verbose.

You can opt into succinct mode with `ginkgo --succinct`, verbose mode with `ginkgo -v` and very-verbose mode with `ginkgo -vv`.

These settings control the amount of information emitted with each spec.  By default (i.e. succinct and normal) Ginkgo only emits detailed information about specs that fail.  That includes the location of the spec/failure and a timeline that includes any captured `GinkgoWriter` content alongside a series of relevant spec events.

The two verbose settings are most helpful when debugging spec suites.  They make Ginkgo emit the detailed timeline information for _every_ spec regardless of failure or success.  When running in series with `-v` or `-vv` mode Ginkgo will stream out the timeline in real-time while specs are running. A real-time stream isn't possible when running in parallel (the [streams would be interleaved](https://www.youtube.com/watch?v=jyaLZHiJJnE)); instead Ginkgo emits all this information about each spec right after it completes.

Very-verbose mode contains additional information over verbose mode.  In particular, `-vv` timelines indicate when individual nodes start and end and also include the full failure descriptions for _every_ failure encountered by the spec.  Verbose mode does not include the node start/end events (though this can be turned on with `--show-node-events`) and does not include detailed failure information for anything other than the first (primary) failure.  (Additional/subseuqent failures typically occur in clean-up nodes and are not as relevant as the primary failure that occurs in a subject or setup node).

When you [filter specs](#filtering-specs) using Ginkgo's various filtering mechanism Ginkgo usually emits a single cyan `S` for each skipped spec.  If you run with the very-verbose setting, however, Ginkgo will emit the description and location information of every skipped spec.  This can be useful if you need to debug your filter queries and can be paired with `--dry-run`.

#### Other Settings
Here are a grab bag of other settings:

You can disable Ginkgo's color output by running `ginkgo --no-color`.

By default, Ginkgo only emits full stack traces when a spec panics.  When a normal assertion failure occurs, Ginkgo simply emits the line at which the failure occurred.  You can, instead, have Ginkgo always emit the full stack trace by running `ginkgo --trace`.

### Reporting Infrastructure
Ginkgo's console output is great when running specs on the console or quickly grokking a CI run.  Of course, there are several contexts where generating a machine-readable report is crucial.  Ginkgo provides first-class CLI support for generating and aggregating reports in a number of machine-readable formats _and_ an extensible reporting infrastructure to enable additional formats and custom reporting.  We'll dig into these topics in the next few sections.

### Generating machine-readable reports
Ginkgo natively supports generating and aggregating reports in a number of machine-readable formats - and these reports can be generated and managed by simply passing `ginkgo` command line flags.

A JSON-formatted report that faithfully captures all available information about a Ginkgo spec run can be generated via:

```bash
ginkgo --json-report=report.json
```

The resulting JSON file encodes an array of `types.Report`.  Each entry in that array lists detailed information about an individual spec suite and includes a list of `types.SpecReport` that captures detailed information about each spec.  These types are documented in [godoc](https://pkg.go.dev/github.com/onsi/ginkgo/v2/types).

When possible, we recommend building tooling on top of Ginkgo's JSON format and using Ginkgo's `types` package directly to access the suite and spec reports.  The structs in the package include several helper functions to interpret the report.

Ginkgo also supports generating JUnit reports with 

```bash
ginkgo --junit-report=report.xml
```

The JUnit report is compatible with the JUnit specification, however Ginkgo specs carry much more metadata than can be easily mapped onto the JUnit spec so some information is lost and/or a bit harder to decode than using Ginkgo's native JSON format.

Ginkgo also supports Teamcity reports with `ginkgo --teamcity-report=report.teamcity` though, again, the Teamcity spec makes it difficult to capture all the spec metadata.

All the machine-readable reports include the full `-vv` version of the timeline for all specs.  This allows you to run Ginkgo in CI with the normal verbosity setting but still get all the detailed information in the machine-readable format.

Of course, you can generate multiple formats simultaneously by passing in multiple flags:

```bash
ginkgo --json-report=report.json --junit-report=report.xml
```

By default, when any of these command-line report flags are provided Ginkgo will generate a single report file, per format, at the passed-in file name.  If Ginkgo is running multiple suites (e.g. `ginkgo -r --json-report=report.json`) then _all_ the suite reports will be encoded in the single report file.

If you'd rather generate separate reports for each suite, you can pass in the `--keep-separate-reports` flag like so: `ginkgo -r --json-report=report.json --keep-separate-reports`.  This will generate an individual report named `report.json` in each suite/package directory,

If you'd like to have all reports end up in a single directory.  Set `--output-dir=<dir>`:

When generating combined reports with: `ginkgo -r --json-report=report.json --output-dir=<dir>` Ginkgo will create the `<dir>` directory (if necessary), and place `report.json` there.

When generating separate reports with: `ginkgo -r --json-report=report.json --output-dir=<dir> --keep-separate-reports` Ginkgo will create the `<dir>` directory (if necessary), and place a report file per package in the directory.  These reports will be namespaced with the name of the package: `PACKAGE_NAME_report.json`.


### Generating reports programmatically

The JSON and JUnit reports described above can be easily generated from the command line - there's no need to make any changes to your suite.

Ginkgo's reporting infrastructure does, however, provide several mechanisms for writing custom reporting code in your spec suites (or, in a supporting package).  We'll explore these mechanisms next.

#### Getting a report for the current spec

At any point during the Run Phase you can get an information-rich up-to-date copy of the current spec's report by running `CurrentSpecReport()`.

There are several uses for this data.  For example, you can write code that performs additional, potentially expensive, diagnostics after a spec runs - but only if the spec has failed:

```go
Describe("Manipulating books at the central library", func() {
  It("can fetch all books", func() {
    Expect(libraryClient.FetchBooks()).NotTo(BeEmpty())
  })

  It("can fetch a specific book", func() {
    book, err := libraryClient.FetchBook("Les Miserables")
    Expect(err).NotTo(HaveOccurred())
    Expect(book.AuthorLastName()).To(Equal("Hugo"))    
  })

  It("can update a book", func() {
    book, err := libraryClient.FetchBook("Les Miserables")
    Expect(err).NotTo(HaveOccurred())
    book.Author = "Victor Marie Hugo"
    Expect(libraryClient.SaveBook(book)).To(Succeed())
  })

  AfterEach(func() {
    if CurrentSpecReport().Failed() {
      GinkgoWriter.Println(libraryClient.DebugLogs())
    }
  })
})
```

In this example, the `AfterEach` closure is using `CurrentSpecReport()` to discover whether or not the current spec has failed.  If it has debug information is fetched from the library server and emitted to the `GinkgoWriter`.

Given `CurrentSpecReport()` you can imagine generating custom report information with something like a top-level `AfterEach`.  For example, let's say we want to write report information to a local file using a custom format _and_ send updates to a remote server.  You might try something like:

```go
/* === INVALID === */
var report *os.File
BeforeSuite(func() {
  report = os.Create("report.custom")
  DeferCleanup(report.Close)
})

AfterEach(func() {
  report := CurrentSpecReport()
  customFormat := fmt.Sprintf("%s | %s", report.State, report.FullText())
  fmt.Fprintln(report, customFormat)
  client.SendReport(customFormat)
})
```

At first glance it looks like this could work.  However, there are a number of problems with this approach:

First of all, the `AfterEach` will _only_ be called if the spec in question runs.  It will never be called for skipped or pending specs and we'll miss reporting on those specs!

Second, the approach we're taking to generate a custom report file will work when running in serial, but not in parallel.  In parallel, multiple test processes will race over writing to `report.custom` and you'll end up with a mess.

Ginkgo's reporting infrastructure provides an alternative solution for this use case.  A special category of setup nodes called **Reporting Nodes**.

#### Reporting Nodes - ReportAfterEach and ReportBeforeEach

Ginkgo provides three reporting-focused nodes `ReportAfterEach`, `ReportAfterSuite`, and `ReportBeforeEach`.

`ReportAfterEach` behaves similarly to a standard `AfterEach` node and can be declared anywhere an `AfterEach` node can be declared.  `ReportAfterEach` takes a closure that accepts a single [`SpecReport`](https://pkg.go.dev/github.com/onsi/ginkgo/v2/types#SpecReport) argument.  For example, we could implement a top-level ReportAfterEach that emits information about every spec to a remote server:

```go
ReportAfterEach(func(report SpecReport) {
  customFormat := fmt.Sprintf("%s | %s", report.State, report.FullText())
  client.SendReport(customFormat)
})
```

`ReportAfterEach` has several unique properties that distinguish it from `AfterEach`.  Most importantly, `ReportAfterEach` closures are **always** called - even if the spec has failed, is marked pending, or is skipped.  This ensures reports that rely on `ReportAfterEach` are complete.

In addition, `ReportAfterEach` closures are called after a spec completes.  i.e. _after_ all `AfterEach` closures have run.  This gives them access to the complete final state of the spec.  Note that if a failure occurs in a `ReportAfterEach` your the spec will be marked as failed.  Subsequent `ReportAfterEach` closures will see the failed state, but not the closure in which the failure occurred.

`ReportAfterEach` is useful if you need to stream or emit up-to-date information about the suite as it runs. Ginkgo also provides `ReportBeforeEach` which is called before the test runs and receives a preliminary `types.SpecReport` - the state of this report will indicate whether the test will be skipped or is marked pending.

You should be aware that when running in parallel, each parallel process will be running specs and their `ReportAfterEach`es.  This means that multiple `ReportAfterEach` blocks can be running concurrently on independent processes.  Given that, code like this won't work:

```go
/* === INVALID === */
var reportFile *os.File
BeforeSuite(func() {
  reportFile = os.Create("report.custom")
})

ReportAfterEach(func(report SpecReport) {
  fmt.Fprintf(reportFile, "%s | %s\n", report.FullText(), report.State)
})
```

you'll end up with multiple processes writing to the same file and the output will be a mess.  There is a better approach for this usecase...

#### Reporting Nodes - ReportBeforeSuite and ReportAfterSuite
`ReportBeforeSuite` and `ReportAfterSuite` nodes behave similarly to `BeforeSuite` and `AfterSuite` and can be placed at the top-level of your suite (typically in the suite bootstrap file).  `ReportBeforeSuite` and `ReportAfterSuite` nodes take a closure that accepts a single [`Report`]((https://pkg.go.dev/github.com/onsi/ginkgo/v2/types#Report)) argument:

```go
var _ = ReportBeforeSuite(func(report Report) {
  // process report
})

var _ = ReportAfterSuite("custom report", func(report Report) {
  // process report
})
```

`Report` contains all available information about the suite.  For `ReportAfterSuite` this will include individual `SpecReport` entries for each spec that ran in the suite, and the overall status of the suite (whether it passed or failed).  Since `ReportBeforeSuite` runs before the suite starts - it does not contain any spec reports, however the count of the number of specs that _will_ be run can be extracted from `report.PreRunStats.SpecsThatWillBeRun`.

The closure passed to `ReportBeforeSuite` is called exactly once at the beginning of the suite before any `BeforeSuite` nodes or specs run have run.  The closure passed to `ReportAfterSuite` is called exactly once at the end of the suite after any `AfterSuite` nodes have run.

Finally, and most importantly, when running in parallel both `ReportBeforeSuite` and `ReportAfterSuite` **only run on process #1**.  Gingko guarantess that no other processes will start running their specs until after `ReportBeforeSuite` on process #1 has completed.  Similarly, Ginkgo will only run `ReportAfterSuite` on process #1 after all other processes have finished and exited.  Ginkgo provides a sinle `Report` that aggregates the `SpecReports` from all processes.  This allows you to perform any custom suite reporting in one place after all specs have run and not have to worry about aggregating information across multiple parallel processes.

Given all this, we can rewrite our invalid `ReportAfterEach` example from above into a valid `ReportAfterSuite` example:

```go
ReportAfterSuite("custom report", func(report Report) {
  f := os.Create("report.custom")
  for _, specReport := range report.SpecReports {
    fmt.Fprintf(f, "%s | %s\n", report.FullText(), specReport.State)
  }
  f.Close()
})
```

Now each suite will generate exactly one report with all the specs appropriately formatted whether running in series or in parallel.

### Attaching Data to Reports
Ginkgo supports attaching arbitrary data to individual spec reports.  These are called `ReportEntries` and appear in the various report-related data structures (e.g. `Report` in `ReportAfterSuite` and `SpecReport` in `ReportAfterEach`) as well as the machine-readable reports generated by `--json-report`, `--junit-report`, etc.  `ReportEntries` are also emitted to the console by Ginkgo's reporter and you can specify a visibility policy to control when this output is displayed.

You attach data to a spec report via

```go
AddReportEntry(name string, args ...interface{})
```

`AddReportEntry` can be called from any setup or subject node closure.  When called, `AddReportEntry` generates `ReportEntry` and attaches it to the current running spec.  `ReportEntry` includes the passed in `name` as well as the time and source location at which `AddReportEntry` was called.  Users can also attach a single object of arbitrary type to the `ReportEntry` by passing it into `AddReportEntry` - this object is wrapped and stored under `ReportEntry.Value` and is always included in the suite's JSON report.

You can access the report entries attached to a spec by getting the `CurrentSpecReport()` or registering a `ReportAfterEach()` - the returned report will include the attached `ReportEntries`.  You can fetch the value associated with the `ReportEntry` by calling `entry.GetRawValue()`.  When called in-process this returns the object that was passed to `AddReportEntry`.  When called after hydrating a report from JSON `entry.GetRawValue()` will include a parsed JSON `interface{}` - if you want to hydrate the JSON yourself into an object of known type you can `json.Unmarshal([]byte(entry.Value.AsJSON), &object)`.

#### Supported Args
`AddReportEntry` supports the `Offset` and `CodeLocation` decorators.  These will control the source code location associated with the generated `ReportEntry`.  You can also pass in a `time.Time` argument to override the timestamp associated with the `ReportEntry` - this can be helpful if you want to ensure a consistent timestamp between your code and the `ReportEntry`.

You can also pass in a `ReportEntryVisibility` enum to control the report's visibility.  This is discussed in more detail below.

If you pass multiple arguments of the same type (e.g. two `Offset`s), the last argument in wins.  This does mean you cannot attach an object with one of the types discussed in this section as the `ReportEntry.Value`.  To get by this you'll need to define a custom type.  For example, if you want the `Value` to be a `time.Time` timestamp you can use a custom type such as

`type Timestamp time.Time`

#### Controlling Output
By default, Ginkgo's console reporter will emit any `ReportEntry` attached to a spec.  It will emit the `ReportEntry` name, location, and time.  If the `ReportEntry` value is non-nil it will also emit a representation of the value.  If the value implements `fmt.Stringer` or `types.ColorableStringer` then `value.String()` or `value.ColorableString()` (which takes precedence) is used to generate the representation, otherwise Ginkgo uses `fmt.Sprintf("%#v", value)`. 

You can modify this default behavior by passing in one of the `ReportEntryVisibility` enum to `AddReportEntry`:

- `ReportEntryVisibilityAlways`: the default behavior - the `ReportEntry` is always emitted.
- `ReportEntryVisibilityFailureOrVerbose`: the `ReportEntry` is only emitted if the spec fails or the tests are run with `-v` (similar to `GinkgoWriter`s behavior).
- `ReportEntryVisibilityNever`: the `ReportEntry` is never emitted though it appears in any generated machine-readable reports (e.g. by setting `--json-report`).

The console reporter passes the string representation of the `ReportEntry.Value` through Ginkgo's `formatter`.  This allows you to generate colorful console output using the color codes documented in `github.com/onsi/ginkgo/v2/formatter/formatter.go`.  For example:

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

Lastly, it is possible to pass a pointer into `AddReportEntry`.  Ginkgo will compute the string representation of the passed in pointer at the last possible moment - so any changes to the object _after_ it is reported will be captured in the final report.  This is useful for building libraries on top of `AddReportEntry` - users can simply register objects when they're created and any subsequent mutations will appear in the generated report.  You can see an example of this in the [Benchmarking Code](#benchmarking-code) pattern section of the patterns chapter.

### Profiling your Suites
Go supports a rich set of profiling features to gather information about your running test suite.  Ginkgo exposes all of these and manages them for you when you are running multiple suites and/or parallel suites.

Ginkgo supports `--race` to analyze race conditions, `--cover` to compute code coverage, `--vet` to evaluate and vet your code, `--cpuprofile` to profile CPU performance, `--memprofile` to profile memory usage, `--blockprofile` to profile blocking goroutines, and `--mutexprofile` to profile locking around mutexes.

`ginkgo -race` runs the race detector and emits any detected race conditions as the suite runs.  If any are detected the suite is marked as failed.

`ginkgo -vet` allows you to configure the set of checks that are applied when your code is compiled.  `ginkgo` defaults to the set of default checks that `go test` uses and you can specify additional checks by passing a comma-separated list to `--vet`.  The set of available checks can be found by running `go doc cmd/vet`.

#### Computing Coverage
`ginkgo -cover` will compute and emit code coverage.  When running multiple suites Ginkgo will emit coverage for each suite and then emit a composite coverage across all running suites.  As with `go test` the default behavior for a given suite is to measure the coverage it provides for the code in the suite's package - however you can extend coverage to additional packages using `--coverpkg`.  You can provide a comma-separated list of package names (as they appear in `import` statements) or a relative path.  You can also use `...` for recursion.  For example, say we have a package called "github.com/foo/bar".  The following are equivalent:

```bash
ginkgo -coverpkg=./... -r
ginkgo -coverpkg=github.com/foo/bar/... -r
```

and will have the effect of calculating coverage for **all** code in the package by **all** specs in the package.

You can also specify the `--covermode` to be one of `set` ("was this code called at all?"), `count` (how many times was it called?) and `atomic` (same as count, but threadsafe and expensive).  If you run `ginkgo --race --cover` the `--covermode` is automatically set to `atomic`.

When run with `--cover`, Ginkgo will generate a single `coverprofile.out` file that captures the coverage statistics of all the suites that ran.  You can change the name of this file by specifying `-coverprofile=filename`.  If you would like to keep separate coverprofiles for each suite use the `--keep-separate-coverprofiles` option.

Ginkgo also honors the `--output-dir` flag when generating coverprofiles.  If you specify `--output-dir` the generated coverprofile will be placed in the requested directory.  If you also specify `--keep-separate-coverprofiles` individual package coverprofiles will be placed in the requested directory and namespaced with a prefix that contains the name of the package in question.

Finally, when running a suite that has [programatically focused specs](#focused-specs) (i.e. specs with the `Focus` decorator or with nodes prefixed with an `F`) Ginkgo exits the suite early with a non-zero exit code.  This interferes with `go test`'s profiling code and prevents profiles from being generated.  Ginkgo will tell you this has happened.  If you want to profile just a subset of your suite you'll need to use a different [mechanism](#filtering-specs) to filter your specs.

#### Other Profiles
Running `ginkgo` with any of `--cpuprofile=X`, `--memprofile=X`, `--blockprofile=X`, and `--mutexprofile=X` will generate corresponding profile files for suite that runs.  Doing so will also preserve the test binary generated by Ginkgo to enable users to use `go tool pprof <BINARY> <PROFILE>` to analyze the profile.

By default, the test binary and various profile files are stored in the individual directories of any suites that Ginkgo runs.  If you specify `--output-dir`, however, then these assets are moved to the requested directory and namespaced with a prefix that contains the name of the package in question.

As with coverage computation, these profiles will not generate a file if a suite includes programatically focused specs (see the discussion [above](#computing-coverage)).

## Ginkgo and Gomega Patterns
So far we've introduced and described the majority of Ginkgo's capabilities and building blocks.  Hopefully the previous chapters have helped give you a mental model for how Ginkgo specs are written and run.

In this chapter we'll switch gears and illustrate common patterns for how Ginkgo's building blocks can be put together to solve for real-world problems.  Since Ginkgo and Gomega are so often paired this chapter will assume that you are using both together - as you'll see, the combination can unlock some powerful, and expressive, testing patterns.

### Recommended Continuous Integration Configuration

When running in CI you must make sure that the version of the `ginkgo` CLI you are using matches the version of Ginkgo in your `go.mod` file.  You can ensure this by invoking the `ginkgo` command via `go run`:

`go run github.com/onsi/ginkgo/v2/ginkgo`

Once you have `ginkgo` running on CI, you'll want to pick and choose the optimal set of flags for your test runs.  We recommend the following set of flags when running in a continuous integration environment:

```bash
go run github.com/onsi/ginkgo/v2/ginkgo -r --procs=N --compilers=N --randomize-all --randomize-suites --fail-on-pending --keep-going --cover --coverprofile=cover.profile --race --trace --json-report=report.json --timeout=TIMEOUT --poll-progress-after=Xs --poll-progress-interval=Ys
```

Here's why:

- `-r` will recursively find and run all suites in the current directory.
- `-procs=N` will run each suite in parallel.  This can substantially speed up suites and you should experiment with different values of `N`.  Note that it is not recommended that you run specs in parallel with `-p` on CI.  Some CI services run on shared machines that will report (e.g.) `32` cores but will not actually give an individual account access to all those compute resources!
- `--compilers=N` will control how many cores to use to compile suites in parallel.  You may need to set this explicitly to avoid accidentally trying to use all `32` cores on that CI machine!
- `--randomize-all` and `--randomize-suites` will randomize all specs and randomize the order in which suites run.  This will help you suss out spec pollution early!
- `--keep-going` will instruct Ginkgo to keep running suites, even after a suite fails.  This can help you get a set of all failures instead of stopping after the first failed suite.
- `--cover` and `--coverprofile=cover.profile` will compute coverage scores and generate a single coverage file for all your specs.
- `--race` will run the race detector.
- `--trace` will instruct Ginkgo to generate a stack trace for all failures (instead of simply including the location where the failure occurred).  This isn't usually necessary but can be helpful in CI environments where you may not have access to a fast feedback loop to iterate on and debug code.
- `--json-report=report.json` will generate a JSON formatted report file.  You can store these off and use them later to get structured access to the suite and spec results.  Alternatively (or in addition) you can use `--junit-report=report.xml` to generate JUnit-formatted reports; these are compatible with several existing CI systems.
- `--timeout` allows you to specify a timeout for the `ginkgo` run.  The default duration is one hour, which may or may not be enough!
- `--poll-progress-after` and `--poll-progress-interval` will allow you to learn where long-running specs are getting stuck.  Choose a values for `X` and `Y` that are appropriate to your suite.  A long-running integration suite, for example, might set `X` to `120s` and `Y` to `30s` - whereas a quicker set of unit tests might not need this setting.  Note that if you precompile suites and run them from a different directory relative to your source code, you may also need to set `--source-root` to enable Ginkgo to emit source code lines when generating progress reports.

### Supporting Custom Suite Configuration

There are contexts where you may want to change some aspects of a suite's behavior based on user-provided configuration.  There are two widely adopted means of doing this: environment variables and command-line flags.

We'll explore both these options in this section by building out a concrete usecase.  Let's imagine a suite that is intended to ensure that a service is up and running correctly (these are sometimes referred to as smoketest suites).  We want to be able to point our suite at an arbitrary server address/port.  We also want to configure how our suite runs depending on the environment we're smoketesting - we'll want to be minimally invasive for `PRODUCTION` environments, but can perform a more thorough check for `STAGING` environments.

Here's a sketch of what this might look like.

#### Supporting Custom Suite Configuration: Environment Variables
Setting and parsing environment variables is fairly straightforward.  We'll configure the server address with a `SMOKETEST_SERVER_ADDR` environment variable and we'll configure the environment with a `SMOKETEST_ENV` variable.

Our suite might look like:

```go
// This is the testing hook in our bootstrap file
func TestSmokeTest(t *testing.T) {
  RegisterFailHandler(Fail)
  RunSpecs(t, "Smoketest Suite")
}

var client *client.Client
var _ = BeforeSuite(func() {
  // Some basic validations
  Expect(os.Getenv("SMOKETEST_SERVER_ADDR")).NotTo(BeZero(), "Please make sure SMOKETEST_SERVER_ADDR is set correctly.")
  Expect(os.Getenv("SMOKETEST_ENV")).To(Or(Equal("PRODUCTION"), Equal("STAGING")), "SMOKETEST_ENV must be set to PRODUCTION or STAGING.")

  //set up a client 
  client = client.NewClient(os.Getenv("SMOKETEST_SERVER_ADDR"))
})

var _ = Describe("Smoketests", func() {
  Describe("Minimally-invasive", func() {
    It("can connect to the server", func() {
      Eventually(client.Connect).Should(Succeed())
    })

    It("can get a list of books", func() {
      Expect(client.ListBooks()).NotTo(BeEmpty())
    })
  })

  if os.Getenv("SMOKETEST_ENV") == "STAGING" {
    Describe("Ensure basic CRUD operations", func() {
      It("can create, updated, and delete a book", func() {
        book := &books.Book{
          Title: "This Book is a Test",
          Author: "Ginkgo",
          Pages: 17,
        }
        Expect(client.Store(book)).To(Succeed())
        Expect(client.FetchByTitle("This Book is a Test")).To(Equal(book))
        Expect(client.Delete(book)).To(Succeed())
        Expect(client.FetchByTitle("This Book is a Test")).To(BeNil())
      })
    })
  }
})
```

users could then run:

```bash
SMOKETEST_SERVER_ADDR="127.0.0.1:3000" SMOKETEST_ENV="STAGING" ginkgo
```

to run all three specs against a local server listening on port `3000`.  If the user fails to correctly provide the configuration environment variables, the `BeforeSuite` checks will fail and `Gomega` will emit the description strings (e.g. "Please make sure SMOKETEST_SERVER_ADDR is set correctly.") to help the user know what they missed.

As you can see, environment variables are convenient and easily accessible from anywhere in the suite.  We use them during the Run Phase to configure the client.  But we also use them at the Tree Construction Phase to control which specs are included in the suite.  There are some clearer ways to accomplish the latter so keep reading!

#### Supporting Custom Configuration: Custom Command-Line Flags
An alternative to environment variables is to provide custom command-line flags to the suite.  These take a bit more setting up but have the benefit of being a bit more self-documenting and structured.

The tricky bits here are:

1. Injecting your command line flags into Go's `flags` list before the test process parses flags.
2. Understanding when in the spec lifecycle the parsed flags are available.
3. Remembering to pass the flags in correctly.

Here's a fleshed out example:

```go
var serverAddr, smokeEnv string

// Register your flags in an init function.  This ensures they are registered _before_ `go test` calls flag.Parse().
func init() {
  flag.StringVar(&serverAddr, "server-addr", "", "Address of the server to smoke-check")
  flag.StringVar(&smokeEnv, "environment", "", "Environment to smoke-check")
}

// This is the testing hook in our bootstrap file
func TestSmokeTest(t *testing.T) {
  RegisterFailHandler(Fail)
  RunSpecs(t, "Smoketest Suite")
}

var client *client.Client
var _ = BeforeSuite(func() {
  // Some basic validations - at this point the flags have been parsed so we can access them
  Expect(serverAddr).NotTo(BeZero(), "Please make sure --server-addr is set correctly.")
  Expect(smokeEnv).To(Or(Equal("PRODUCTION"), Equal("STAGING")), "--environment must be set to PRODUCTION or STAGING.")

  //set up a client 
  client = client.NewClient(serverAddr)
})

var _ = Describe("Smoketests", func() {
  Describe("Minimally-invasive", func() {
    It("can connect to the server", func() {
      Eventually(client.Connect).Should(Succeed())
    })

    It("can get a list of books", func() {
      Expect(client.ListBooks()).NotTo(BeEmpty())
    })
  })

  if smokeEnv == "STAGING" {
    Describe("Ensure basic CRUD operations", func() {
      It("can create, updated, and delete a book", func() {
        book := &books.Book{
          Title: "This Book is a Test",
          Author: "Ginkgo",
          Pages: 17,
        }
        Expect(client.Store(book)).To(Succeed())
        Expect(client.FetchByTitle("This Book is a Test")).To(Equal(book))
        Expect(client.Delete(book)).To(Succeed())
        Expect(client.FetchByTitle("This Book is a Test")).To(BeNil())
      })
    })
  }
})
```

We would invoke this suite with

```bash
ginkgo -- --server-addr="127.0.0.1:3000" --environment="STAGING"
```

note the `--` separating the arguments `ginkgo` from the arguments passed down to the suite.  You would put Ginkgo's arguments to the left of `--`.  For example, to run in parallel:

```bash
ginkgo -p -- --server-addr="127.0.0.1:3000" --environment="STAGING"
```

One more note before we move on.  As shown in this example, parsed flag variables are available both during the Run Phase (e.g. when we call `client.NewClient(serverAddr)`) _and_ during the Tree Construction Phase (e.g. when we guard the `CRUD` specs with `if smokeEnv == "STAGING"`).  However flag variables are _not_ available at the **top-level** of the suite.

Here's a trivial, but instructive, example.  Say we wanted to add the value of `environment` to the name the top-level `Describe`:

```go
...

var describeName = fmt.Sprintf("Smoketests - %s", smokeEnv)
var _ = Describe(describeName, func() {
  ...
})

...
```

Counter-intuitively, this will always yield `"Smoketests - "`.  The reason is that `fmt.Sprintf` is being called as go is traversing the top-level identifiers in the suite.  At this point, `init` functions are being _defined_ but have not yet been invoked.  So (a) we haven't actually registered our flags yet and, more importantly, (b) `go test` hasn't _parsed_ the flags yet.  Our `smokeEnv` variable is therefore empty.  There's no way around this - in general you should avoid trying to access configuration information at the top-level.  However, if you must then you will need to use use environment variables instead of flags.

#### Overriding Ginkgo's command-line configuration in the suite

The previous two examples used an `if` guard to control whether specs were included in the spec tree based on user-provided configuration.  This approach _works_ but can be a bit confusing - specs that are "skipped" in this way never appear in any generated reports, and the total number of specs in the suite depends on configuration.  It would be cleaner and clearer to leverage Ginkgo's filtering mechanisms.  You could, for example, use `Skip`:

```go
var _ = Describe("Smoketests", func() {
  Describe("Minimally-invasive", func() {
    It("can connect to the server", func() {
      ...
    })

    It("can get a list of books", func() {
      ...
    })
  })

  Describe("Ensure basic CRUD operations", func() {
    BeforeEach(func(){
      if environment != "STAGING" {
        Skip("CRUD spec only runs on staging")
      }
    })

    It("can create, updated, and delete a book", func() {
      ...
    })
  })
})
```

this works just fine - however as the suite grows you may see that `environment` check start to spread throughout the suite.  You could, instead, use Ginkgo's label mechanisms.  Here we're explicitly labeling specs with their allowed environments:

```go
var _ = Describe("Smoketests", func() {
  Describe("Minimally-invasive", Label("PRODUCTION", "STAGING"), func() {
    It("can connect to the server", func() {
      ...
    })

    It("can get a list of books", func() {
      ...
    })
  })

  Describe("Ensure basic CRUD operations", Label("STAGING"), func() {
    It("can create, updated, and delete a book", func() {
      ...
    })
  })
})
```

We could then use Ginkgo's expressive filter queries to control which specs do/don't run.  However that would require us to change our contract with the user.  They'll now need to run:

```bash
ginkgo --label-filter="STAGING" -- --server-addr="127.0.0.1"
```

this isn't great.  Ideally we'd maintain the same contract and allow the user to express their intent through the existing semantics of "environment" and take care of managing the label-filter in the suite.

You can accomplish this in Ginkgo by overriding Ginkgo's configuration _before_ running the specs.  Here's our fully-worked example showing how:

```go
var serverAddr, smokeEnv string

// Register your flags in an init function.  This ensures they are registered _before_ `go test` calls flag.Parse().
func init() {
  flag.StringVar(&serverAddr, "server-addr", "", "Address of the server to smoke-check")
  flag.StringVar(&smokeEnv, "environment", "", "Environment to smoke-check")
}

// This is the testing hook in our bootstrap file
func TestSmokeTest(t *testing.T) {
  RegisterFailHandler(Fail)

  //we're moving the validation up here since we're about to use the flag variables before entering the RunPhase
  //thankfully Gomega can run within normal `testing` tests, we simply create a new Gomega by wrapping `testing.T`
  g := NewGomegaWithT(t)
  g.Expect(serverAddr).NotTo(BeZero(), "Please make sure --server-addr is set correctly.")
  g.Expect(smokeEnv).To(Or(Equal("PRODUCTION"), Equal("STAGING")), "--environment must be set to PRODUCTION or STAGING.")

  //we're now guaranteed to have validated configuration variables
  //let's update Ginkgo's configuration using them
  //first we grab Ginkgo's current configuration
  suiteConfig, _ := GinkgoConfiguration() //the second argument is the reporter configuration which we won't be adjusting

  //now we modify the label-filter
  if suiteConfig.LabelFilter == "" {
    suiteConfig.LabelFilter = smokeEnv
  }  else {
    // if the user has specified a label-filter we extend it:
    suiteConfig.LabelFilter = "(" + suiteConfig.LabelFilter + ") && " + smokeEnv 
  }

  // finally, we pass the modified configuration in to RunSpecs
  RunSpecs(t, "Smoketest Suite", suiteConfig)
}

var client *client.Client
var _ = BeforeSuite(func() {
  client = client.NewClient(serverAddr)
})

var _ = Describe("Smoketests", func() {
  Describe("Minimally-invasive", Label("PRODUCTION", "STAGING"), func() {
    It("can connect to the server", func() {
      Eventually(client.Connect).Should(Succeed())
    })

    It("can get a list of books", func() {
      Expect(client.ListBooks()).NotTo(BeEmpty())
    })
  })

  Describe("Ensure basic CRUD operations", Label("STAGING"), func() {
    It("can create, updated, and delete a book", func() {
      book := &books.Book{
        Title: "This Book is a Test",
        Author: "Ginkgo",
        Pages: 17,
      }
      Expect(client.Store(book)).To(Succeed())
      Expect(client.FindByTitle("This Book is a Test")).To(Equal(book))
      Expect(client.Delete(book)).To(Succeed())
      Expect(client.FindByTitle("This Book is a Test")).To(BeNil())
    })
  })
})
```

In this way we can provide alternative, more semantically appropriate, interfaces to consumers of our suite and build on top of Ginkgo's existing building blocks.

### Dynamically Generating Specs

There are several patterns for dynamically generating specs with Ginkgo.  You can use a simple loop to generate specs.  For example:

```go
Describe("Storing and retrieving books by category", func() {
  for _, category := range []books.Category{books.CategoryNovel, books.CategoryShortStory, books.CategoryBiography} {
    category := category
    It(fmt.Sprintf("can store and retrieve %s books", category), func() {
      book := &books.Book{
        Title: "This Book is a Test",
        Author: "Ginkgo",
        Category: category,
      }
      Expect(library.Store(book)).To(Succeed())
      DeferCleanup(library.Delete, book)
      Expect(library.FindByCategory(category)).To(ContainElement(book))      
    })
  }
})
```

This will generate several `It`s - one for each category.  Note that you must assign a copy of the loop variable to a local variable (that's what `category := category` is doing) - otherwise the `It` closure will capture the mutating loop variable and all the specs will run against the last element in the loop.  It is idiomatic to give the local copy the same name as the loop variable.

Of course, this particular example might be better written as a [table](#table-specs)!

There are contexts where external information needs to be loaded in order to figure out which specs to dynamically generate.  For example, let's say we maintain a `json` file that lists a set of fixture books that we want to test storing/retrieving from the library.  There are many ways to approach writing such a test - but let's say we want to maximize parallelizability of our suite and so want to generate a separate `It` for each book fixture.

Many Ginkgo users attempt the following approach.  It's a common gotcha:

```go
/* === INVALID === */
var fixtureBooks []*books.Book

var _ = BeforeSuite(func() {
  fixtureBooks = LoadFixturesFrom("./fixtures/books.json")
  Expect(fixtureBooks).NotTo(BeEmpty())
})

Describe("Storing and retrieving the book fixtures", func() {
  for _, book := range fixtureBooks {
    book := book
    It(fmt.Sprintf("can store and retrieve %s", book.Title), func() {
      Expect(library.Store(book)).To(Succeed())
      DeferCleanup(library.Delete, book)
      Expect(library.FindByTitle(book.Title)).To(Equal(book))            
    })
  }
})
```

This will not work.  The fixtures are loaded in the `BeforeSuite` closure which runs during the **Run Phase**... _after_ the **Tree Construction Phase** where we loop over `fixtureBooks`.  If you need to perform work that influences the structure of the spec tree you must do it  _before_ or _during_ the Tree Construction Phase.  In this case, it is idiomatic to place the relevant code in the `Test` function in the bootstrap file:

```go
var fixtureBooks []*books.Book

func TestBooks(t *testing.T) {
  RegisterFailHandler(Fail)

  // perform work that needs to be done before the Tree Construction Phase here
  // note that we wrap `t` with a new Gomega instance to make assertions about the fixtures here.
  g := NewGomegaWithT(t)
  fixtureBooks = LoadFixturesFrom("./fixtures/books.json")
  g.Expect(fixtureBooks).NotTo(BeEmpty())

  // finally, we pass the modified configuration in to RunSpecs
  RunSpecs(t, "Books Suite")
}

Describe("Storing and retrieving the book fixtures", func() {
  for _, book := range fixtureBooks {
    book := book
    It(fmt.Sprintf("can store and retrieve %s", book.Title), func() {
      Expect(library.Store(book)).To(Succeed())
      DeferCleanup(library.Delete, book)
      Expect(library.FindByTitle(book.Title)).To(Equal(book))            
    })
  }
})
```

### Shared Behaviors
It's common to want to extract subsets of spec behavior for reuse - these are typically called "Shared Behaviors".  

It is often the case that within a particular suite there will be a number of different `Context`s that assert the exact same behavior, in that they have identical `It`s within them.  The only difference between these `Context`s is the set up done in their respective `BeforeEach`s.  Rather than repeat the `It`s for these `Context`s, you can extract the code into a shared-scope closure and avoid repeating yourself.  For example:

```go
Describe("Storing books in the library", func() {
  var book *books.Book{}

  Describe("the happy path", func() {
    BeforeEach(func() {
      book = &books.Book{
        Title:  "Les Miserables",
        Author: "Victor Hugo",
        Pages:  2783,
      }
    })

    It("validates that the book can be stored", func() {
      Expect(library.IsStorable(book)).To(BeTrue())
    })

    It("can store the book", func() {
      Expect(library.Store(book)).To(Succeed())
    })
  })

  Describe("failure modes", func() {
    AssertFailedBehavior := func() {
      It("validates that the book can't be stored", func() {
        Expect(library.IsStorable(book)).To(BeFalse())
      })

      It("fails to store the book", func() {
        Expect(library.Store(book)).To(MatchError(books.ErrStoringBook))
      })
    }

    Context("when the book has no title", func() {
      BeforeEach(func() {
        book = &books.Book{
          Author: "Victor Hugo",
          Pages:  2783,
        }
      })

      AssertFailedBehavior()
    })

    Context("when the book has no author", func() {
      BeforeEach(func() {
        book = &books.Book{
          Title: "Les Miserables",
          Pages:  2783,
        }
      })

      AssertFailedBehavior()
    })

    Context("when the book is nil", func() {
      BeforeEach(func() {
        book = nil
      })

      AssertFailedBehavior()
    })    
  })
})
```

Since `AssertFailedBehavior` is defined in the same stack of closures as the other nodes, it has access to the shared `book` variable.  Note that the `AssertFailedBehavior` function is called within the body of the `Context` container block.  This will happen during The Tree Construction phase and result in a spec tree that includes the `It`s defined in the `AssertFailedBehavior` function for each context.  

### Table Specs Patterns

We introduced Ginkgo's support for Table Specs in an [earlier section](#table-specs).  Here we'll just outline a couple of useful patterns.

Tables specs allow you to specify a spec function that takes arbitrary parameters and entries to feed parameters to the function.  This works well when you've got a small handful of parameters but can become unwieldy with more parameters.  For example:

```go
var book *books.Book
BeforeEach(func() {
  book = LoadFixture("les-miserables.json")
})
DescribeTable("Repaginating Books",
  func(fontSize int, lineHeight float64, pageWidth float64, pageHeight float64, expectedPages int) {
    book.SetFontSize(fontSize)
    book.SetLineHeight(lineHeight)
    book.SetPageDimensions(pageWidth, pageHeight)
    Expect(book.RecomputePages()).To(BeNumerically("~", expectedPages, 30))
  },
  func(fontSize int, lineHeight float64, pageWidth float64, pageHeight float64, expectedPages int) string {
    return fmt.Sprintf("FontSize: %d, LineHeight: %.2f, Page:%.2fx%.2f => %d", fontSize, lineHeight, pageWidth, pageHeight, expectedPages)
  }
  Entry(nil, 12, 1.2, 8.5, 11, 2783),
  Entry(nil, 14, 1.3, 8.5, 11, 3120),
  Entry(nil, 10, 1.2, 8.5, 11, 2100),
  Entry(nil, 12, 2.0, 8.5, 11, 6135),
  Entry(nil, 12, 1, 5, 6, 12321),
)
```

These entries are inscrutable!  A common pattern in this case is to define a type to capture the entry information:

```go
var book *books.Book
type BookFormatting struct {
  FontSize int
  LineHeight float64
  PageWidth float64
  PageHeight float64
}

BeforeEach(func() {
  book = LoadFixture("les-miserables.json")
})
DescribeTable("Repaginating Books",
  func(formatting BookFormatting, expectedPages int) {
    book.SetFontSize(formatting.FontSize)
    book.SetLineHeight(formatting.LineHeight)
    book.SetPageDimensions(formatting.PageWidth, formatting.PageHeight)
    Expect(book.RecomputePages()).To(BeNumerically("~", expectedPages, 30))
  },
  func(formatting BookFormatting, expectedPages int) string {
    return fmt.Sprintf("FontSize: %d, LineHeight: %.2f, Page:%.2fx%.2f => %d", 
      formatting.fontSize, formatting.lineHeight, 
      formatting.pageWidth, formatting.pageHeight,
      expectedPages)
  }
  Entry(nil, BookFormatting{FontSize: 12, LineHeight: 1.2, PageWidth:8.5, PageHeight:11}, 2783),
  Entry(nil, BookFormatting{FontSize: 14, LineHeight: 1.3, PageWidth:8.5, 11}, 3120),
  Entry(nil, BookFormatting{FontSize: 10, LineHeight: 1.2, PageWidth:8.5, 11}, 2100),
  Entry(nil, BookFormatting{FontSize: 12, LineHeight: 2.0, PageWidth:8.5, 11}, 6135),
  Entry(nil, BookFormatting{FontSize: 12, LineHeight: 1, PageWidth:5, PageHeight:6}, 12321),
)
```

This is longer but certainly easier to read!

Another Table Spec pattern involves the reuse of table of Entries.  If you have multiple cases to run against the same set of entries you can save of the entries in a `[]TableEntry` slice and then pass the slice to multiple `DescribeTable` functions.  For example:

```go

var InvalidBookEntries = []TableEntry{
  Entry("Empty book", &books.Book{}),
  Entry("Only title", &books.Book{Title: "Les Miserables"}),
  Entry("Only author", &books.Book{Author: "Victor Hugo"}),
  Entry("Missing pages", &books.Book{Title: "Les Miserables", Author: "Victor Hugo"}),
}

DescribeTable("Storing invalid books always errors", func(book *books.Book) {
  Expect(library.Store(book)).To(MatchError(books.ErrInvalidBook))
}, InvalidBookEntries)

DescribeTable("Reading invalid books always errors", func(book *books.Book) {
  Expect(user.Read(book)).To(MatchError(books.ErrInvalidBook))
}, InvalidBookEntries)

```

### Patterns for Asynchronous Testing

It is common, especially in integration suites, to be testing behaviors that occur asynchronously (either within the same process or, in the case of distributed systems, outside the current test process in some combination of external systems).  Ginkgo and Gomega provide the building blocks you need to write effective asynchronous specs efficiently.

Rather than an exhaustive/detailed review we'll simply walk through some common patterns.  Throughout you'll see that you should generally use Ginkgo's interruptible nodes with timeouts alongside use Gomega's `Eventually` and `Consistently` to make [asynchronous assertions](https://onsi.github.io/gomega/#making-asynchronous-assertions).

Both `Eventually` and `Consistently` perform asynchronous assertions by polling the provided input.  In the case of `Eventually`, Gomega polls the input repeatedly until the matcher is satisfied - once that happens the assertion exits successfully and execution continues.  If the matcher is never satisfied `Eventually` will time out with a useful error message.  Both the timeout and polling interval are [configurable](https://onsi.github.io/gomega/#eventually).

In the case of `Consistently`, Gomega polls the input repeatedly and asserts the matcher is satisfied every time.  `Consistently` only exits early if a failure occurs - otherwise it continues polling until the specified interval elapses.  This is often the only way to assert that something "does not happen" in an asynchronous system.

`Eventually` and `Consistently` can accept three types of input.  You can pass in bare values and assert that some aspect of the value changes eventually.  This is most commonly done with Go channels or Gomega's 
[`gbytes`](https://onsi.github.io/gomega/#gbytes-testing-streaming-buffers) and [`gexec`](https://onsi.github.io/gomega/#gexec-testing-external-processes) packages.  You can also pass in functions and assert that their return values `Eventually` or `Consistently` satisfy a matcher - we'll cover those later.  Lastly, you can pass in functions that take a `Gomega` argument - these allow you to make assertions within the function and are a way to assert that a series of assertions _eventually_ succeeds.  We'll cover _that_ later as well.  Let's look at these various input types through the lens of some concrete use-cases.

#### Testing an in-process Asynchronous Service.
Let's imagine an in-process asynchronous service that can prepare books for publishing and emit updates to a buffer.  Since publishing is expensive the publish service returns a channel that will include the published book bits and runs the actual publishing process in a separate Goroutine.  We could test such a service like so:

```go
Describe("Publishing books", func() {
  var book *books.Book
  BeforeEach(func() {
    book = loadBookWithContent("les_miserables.fixture")
    Expect(book).NotTo(BeNil())
  })

  It("can publish a book, emitting information as it goes", func(ctx SpecContext) {
    buffer := gbytes.NewBuffer() //gbytes provides a thread-safe buffer that works with the `gbytes.Say` matcher
    
    // we begin publishing the book.  This kicks off a goroutine and returns a channel
    // Publish takes a `context.Context` and so we pass in our `ctx` to clean up correctly in case the spec timeout elapses
    c := publisher.Publish(ctx, book, buffer)

    // gbytes.Say allows us to assert on output to a stream
    // we pass in the SpecContext to give this block of `Eventually's` a shared time horizon for completing: the 30 second SpecTimeout
    // we don't _have_ to pass in a SpecContext.  If we don't, then each `Eventually` will have its own, individual, timeout.
    Eventually(ctx, buffer).Should(gbytes.Say(`Publishing "Les Miserables...`))
    Eventually(ctx, buffer).Should(gbytes.Say(`Published page 1/2783`))
    Eventually(ctx, buffer).Should(gbytes.Say(`Published page 2782/2783`))
    Eventually(ctx, buffer).Should(gbytes.Say(`Publish complete!`))

    // rather than call <-c which could block the spec forever we use Eventually to poll the channel and
    // store any received values in a pointer
    // we pass in the SpecContext _and_ specify a timeout of 1 second:
    // at this point we expect `Publish()` to exit fairly quickly and should not need to wait for longer than 1s!
    var result publisher.PublishResult
    Eventually(ctx, c).WithTimeout(time.Second).Should(Receive(&result))

    //we make some *synchronous* assertions on the result
    Expect(result.Title).To(Equal("Les Miserables"))
    Expect(result.EpubSize).To(BeNumerically(">", 10))
    Expect(result.EpubContent).To(ContainSubstring("I've ransomed you from fear and hatred, and now I give you back to God."))

    //we expect the publisher to close the channel when it's done
    Eventually(ctx, c).WithTimeout(time.Second).Should(BeClosed())
  }, SpecTimeout(time.Second*30)) //this spec has 30 seconds to complete
})
```

As you can see Gomega allows us to make some pretty complex asynchronous assertions pretty easily!

#### Testing Local Processes
Launching and testing an external process is actually quite similar to testing an in-process asynchronous service (the example above).  You typically leverage Gomega's [`gexec`](https://onsi.github.io/gomega/#gexec-testing-external-processes) and [`gbytes`](https://onsi.github.io/gomega/#gbytes-testing-streaming-buffers) packages.  Let's imagine our book-publishing service was a actually a command-line tool we wanted to test:

```go
//We compile the publisher in a BeforeSuite so its available to our specs
//Not that this step can be skipped if the publisher binary is already precompiled
var publisherPath string
BeforeSuite(func() {
  var err error
  publisherPath, err = gexec.Build("path/to/publisher")
  Expect(err).NotTo(HaveOccurred())
  DeferCleanup(gexec.CleanupBuildArtifacts)  
})

Describe("Publishing books", func() {
  It("can publish a book, emitting information as it goes", func(ctx SpecContext) {
    //First, we create a command to invoke the publisher and pass appropriate args
    cmd := exec.CommandContext(ctx, publisherPath, "-o=les-miserables.epub", "les-miserables.fixture")

    //Now we launch the command with `gexec`.  This returns a session that wraps the running command.  
    //We also tell `gexec` to tee any stdout/stderr output from the process to `GinkgoWriter` - this will
    //ensure we get all the process output if the spec fails.
    session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
    Expect(err).NotTo(HaveOccurred())

    //At this point the process is running in the background
    //In addition to teeing to GinkgoWriter gexec will capture any stdout/stderr output to
    //gbytes buffers.  This allows us to make assertions against its stdout output using `gbytes.Say`
    Eventually(ctx, session).Should(gbytes.Say(`Publishing "Les Miserables...`))
    Eventually(ctx, session).Should(gbytes.Say(`Published page 1/2783`))
    Eventually(ctx, session).Should(gbytes.Say(`Published page 2782/2783`))
    Eventually(ctx, session).Should(gbytes.Say(`Publish complete!`))

    //We can also assert the session has exited 
    Eventually(ctx, session).WithTimeout(time.Second).Should(gexec.Exit(0)) //with exit code 0

    //At this point we should have the `les-miserables.epub` artifact
    Expect("les-miserables.epub").To(BeAnExistingFile())

    result, err := epub.Load("les-miserables.epub")
    Expect(err).NotTo(HaveOccurred())

    //we make some synchronous assertions on the result
    Expect(result.Title).To(Equal("Les Miserables"))
    Expect(result.EpubSize).To(BeNumerically(">", 10))
    Expect(result.EpubContent).To(ContainSubstring("I've ransomed you from fear and hatred, and now I give you back to God."))
  }, Time.Second * 30)
})
```

#### Testing Blocking Functions
It's common in Go for functions to block and perform complex operations synchronously - and leave the work of spawning goroutines and managing thread-safety to the user.  You can test such patterns easily with Gomega.  For example, let's test a flow that performs a few expensive operations and assert that everything finishes eventually.

```go
Describe("Change book font-size", func() {
  var book *books.Book
  BeforeEach(func() {
    book = loadBookWithContent("les_miserables.fixture")
    Expect(book).NotTo(BeNil())
  })
  
  It("can repaginate books without losing any content", func() {
    done := make(chan interface{})
    go func() {
      defer GinkgoRecover()

      content := book.RawContent()
      Expect(book.Pages).To(Equal(2783))

      //this might be quite expensive and will block...
      err := book.SetFontSize(28)
      Expect(err).NotTo(HaveOccurred())

      Expect(book.Pages).To(BeNumerically(">", 2783))
      Expect(book.RawContent()).To(Equal(content))

      close(done)
    }()

    // now we wait for the `done` channel to close.  Note that we neither pass in a context nor set an explicit timeout
    // in this case `Eventually` `will use Gomega's default global timeout (1 second, unless overriden by the user)
    Eventually(done).Should(BeClosed())
  })  
})
```

This use of a `done` channel is idiomatic and guards the spec against potentially hanging forever.  More typically, blocking functions like `SetFontSize` accept a `context.Context` to manage cancellation.  In that case we can simply write:


```go
Describe("Change book font-size", func() {
  var book *books.Book
  BeforeEach(func() {
    book = loadBookWithContent("les_miserables.fixture")
    Expect(book).NotTo(BeNil())
  })
  
  It("can repaginate books without losing any content", func(ctx SpecContext) {
    content := book.RawContent()
    Expect(book.Pages).To(Equal(2783))

    //this might be quite expensive and will block...
    err := book.SetFontSize(ctx, 28)
    Expect(err).NotTo(HaveOccurred())

    Expect(book.Pages).To(BeNumerically(">", 2783))
    Expect(book.RawContent()).To(Equal(content))
  }, SpecTimeout(time.Second))  
})
```


#### Testing External Systems
When integration testing an external system, particularly a distributed system, you'll often find yourself needing to wait for the external state to converge and become eventually consistent.  Gomega makes it easy to poll and validate that the system under test eventually exhibits the desired behavior.  This is typically done by passing functions in to `Eventually` and `Consistently`.

For example, let's imagine testing how an external library service handles notifying users about holds on their books.  Here's what a fully worked example might look like:

```go
var library *library.Client
var _ = BeforeSuite(func() {
  var err error
  library, err = library.NewClient(os.Getenv("LIBRARY_SERVICE"))
  Expect(err).NotTo(HaveOccurred())

  Eventually(library.Ping).Should(Succeed())
})

var _ = Describe("Getting notifications about holds", func() {
  var book *books.Book
  var sarah, jane *user.User
  BeforeEach(func(ctx SpecContext) {
    book = &books.Book{
      Title: "My test book",
      Author: "Ginkgo",
      Pages: 17,
    }

    Expect(library.Store(ctx, book)).To(Succeed())
    // we'll want to delete the book after the spec ends.  `library` has a `Delete` function with signature `Delete(context.Context, *book.Book)`. 
    // DeferCleanup will detect this signature and automatically pass a `SpecContext` (configured with a one second timeout thanks to the `NodeTimeout` decorator)
    // in as the first parameter.  `book` will be passed in as the second parameter.
    DeferCleanup(library.Delete, book, NodeTimeout(time.Second))

    sarah = user.NewUser(ctx, "Sarah", "integration-test-account+sarah@gmail.com")
    jane = user.NewUser(ctx, "Jane", "integration-test-account+jane@gmail.com")
    
    By("Sarah checks the book out")
    Expect(sarah.CheckOut(ctx, library, book)).To(Succeed())
  }, NodeTimeout(time.Second*10))

  It("notifies the user when their hold is ready", func(ctx SpecContext) {
    By("Jane can't check the book out so she places a hold")
    Expect(jane.CheckOut(ctx, library, book)).To(MatchError(books.ErrNoAvailableCopies))
    Expect(jane.PlaceHold(ctx, library, book)).To(Succeed())

    By("when Sarah returns the book")
    Expect(sarah.Return(ctx, library, book)).To(Succeed())

    By("Jane eventually gets notified that her book is available in the library app...")
    Eventually(func(ctx SpecContext) ([]user.Notification, error) {
      return jane.FetchNotifications(ctx, library)
    }).WithContext(ctx).Should(ContainElement(user.Notification{Title: book.Title, State: book.ReadyForPickup}))

    By("...and in her email...")
    Eventually(func(ctx SpecContext) ([]string, error) {
      messages, err := gmail.Fetch(ctx, jane.EmailAddress)
      if err != nil {
        return nil, err
      }
      subjects := []string{}
      for _, message := range messages {
        subjects = append(subjects, message.Subject)
      }
      return subjects, nil
    }).WithContext(ctx).Should(ContainElement(fmt.Sprintf(`"%s" is available for pickup`, book.Title)))

    Expect(jane.CheckOut(ctx, library, book)).To(Succeed())
  }, SpecTimeout(time.Second * 30))
})
```

As you can see we are able to clearly test both synchronous concerns (blocking calls to the library service that return immediately) with asynchronous concerns (out-of-band things that happen after a library call has been made).  The DSL allows us to clearly express our intent and capture the flow of this spec with relatively little noise.

`Eventually` has a few more tricks that we can leverage to clean this code up a bit.  Since `Eventually` accepts functions we can simply replace this:

```go
Eventually(func(ctx SpecContext) ([]user.Notification, error) {
  return jane.FetchNotifications(ctx, library)
}).WithContext(ctx).Should(ContainElement(user.Notification{Title: book.Title, State: book.ReadyForPickup}))
```

with this:

```go
Eventually(jane.FetchNotifications).WithContext(ctx).WithArguments(library).Should(ContainElement(user.Notification{Title: book.Title, State: book.ReadyForPickup}))
```

Note that `Eventually` automatically asserts a niladic error as it polls the `FetchNotifications` function.  Also note that we are passing in a reference to the method on the `jane` instance - not invoking it.  `Eventually(jane.FetchNotifications())` would not work - you must pass in `Eventually(jane.FetchNotifications)`!

`Eventually` can _also_ accept functions that take a `Gomega` parameter.  These functions are then passed a local `Gomega` that can be used to make assertions _inside_ the function as it is polled.  `Eventually` will retry the function if an assertion fails.  This would allow us to replace:

```go
Eventually(func(ctx SpecContext) ([]string, error) {
  messages, err := gmail.Fetch(ctx, jane.EmailAddress)
  if err != nil {
    return nil, err
  }
  subjects := []string{}
  for _, message := range messages {
    subjects = append(subjects, message.Subject)
  }
  return subjects, nil
}).WithContext(ctx).Should(ContainElement(fmt.Sprintf(`"%s" is available for pickup`, book.Title)))
```

with

```go
Eventually(func(g Gomega, ctx SpecContext) []string { //note: g Gomega must go first
  messages, err := gmail.Fetch(ctx, jane.EmailAddress)
  g.Expect(err).NotTo(HaveOccurred())
  subjects := []string{}
  for _, message := range messages {
    subjects = append(subjects, message.Subject)
  }
  return subjects, nil
}).WithContext(ctx).Should(ContainElement(fmt.Sprintf(`"%s" is available for pickup`, book.Title)))
```

we can even push the entire assertion into the polled function:

```go
Eventually(func(g Gomega, ctx SpecContext) {
  messages, err := gmail.Fetch(ctx, jane.EmailAddress)
  g.Expect(err).NotTo(HaveOccurred())
  subjects := []string{}
  for _, message := range messages {
    subjects = append(subjects, message.Subject)
  }
  g.Expect(subjects).To(ContainElement(fmt.Sprintf(`"%s" is available for pickup`, book.Title)))
  return subjects, nil
}).WithContext(ctx).Should(Succeed())
```

this approach highlights a special-case use of the `Succeed()` matcher with `Eventually(func(g Gomega) {})` - `Eventually` will keep retrying the function until no failures are detected.

> You may be wondering why we need to pass in a dedicated `Gomega` instance to the polled function.  That's because the default global-level assertions are implicitly tied to Ginkgo's `Fail` handler.  The first failed assertion in an `Eventually` would cause the spec to fail with no possibility to retry.  By passing in a fresh `Gomega` instance, `Eventually` can monitor for failures itself and control the final failure/success state of the assertion it is governing.

Finally, since we're on the topic of simplifying things, we can make use of the fact that `ContainElement` can take a matcher to compose it with the `WithTransform` matcher and get rid of the `subjects` loop:

```go
Eventually(func(g Gomega, ctx SpecContext) {
  messages, err := gmail.Fetch(jane.EmailAddress)
  g.Expect(err).NotTo(HaveOccurred())
  expectedSubject := fmt.Sprintf(`"%s" is available for pickup`, book.Title)
  subjectGetter := func(m gmail.Message) string { return m.Subject }
  g.Expect(messages).To(ContainElement(WithTransform(subjectGetter, Equal(expectedSubject))))
  return messages, nil
}).WithContext(ctx).Should(Succeed())
```

### Patterns for Parallel Integration Specs
One of Ginkgo's strengths centers around building and running large complex integration suites.  Integration suites are spec suites that exercise multiple related components to validate the behavior of the integrated system as a whole.  They are notorious for being difficult to write, susceptible to random failure, and painfully slow.  They also happen to be incredibly valuable, particularly when building large complex distributed systems.

The [Patterns for Asynchronous Testing](#patterns-for-asynchronous-testing) section above goes into depth about patterns for testing asynchronous systems like these.  This section will cover patterns for ensuring such specs can run in parallel.  Make sure to read the [Spec Parallelization](#spec-parallelization) section to build a mental model for how Ginkgo supports parallelization first - it's important to understand that parallel specs are running in **separate** processes and are coordinated via the Ginkgo CLI.

#### Managing External Processes in Parallel Suites

We covered how to use `gexec` and `gbytes` to compile, launch, and test external processes in the [Testing Local Processes](#testing-local-processes) portion of the asynchronous testing section.  We'll extend the example there to cover how to design such a test to work well in parallel.

First recall that we used a `BeforeSuite` to compile our `publisher` binary:

```go
var publisherPath string
BeforeSuite(func() {
  var err error
  publisherPath, err = gexec.Build("path/to/publisher")
  Expect(err).NotTo(HaveOccurred())
  DeferCleanup(gexec.CleanupBuildArtifacts)  
})
```

This code will work fine in parallel as well (under the hood `gexec.Build` places build artifacts in a randomly-generated temporary directory - this is why you need to call `gexec.CleanupBuildArtifacts` to clean 
up); but it's inefficient and all your parallel processes will spend time up front compiling multiple copies of the same binary.  Instead, we can use `SynchronizedBeforeSuite` to perform the compilation step just once:

```go
var publisherPath string
SynchronizedBeforeSuite(func() []byte {
  path, err := gexec.Build("path/to/publisher")
  Expect(err).NotTo(HaveOccurred())
  DeferCleanup(gexec.CleanupBuildArtifacts)
  return []byte(path)
}, func(path []byte) {
  publisherPath = string(path)
})
```

Now only process #1 will compile the publisher.  All other processes will wait until it's done.  Once complete it will pass the path to the compiled artifact to all other processes.  Note that the `DeferCleanup` in the `SynchronizedBeforeSuite` will have the same runtime semantics as a `SynchronizedAfterSuite` so `gexec` will not cleanup after itself until _all_ processes have finished running.

Now any spec running on any process can simply launch it's own instance of the `publisher` process via `gexec` and make assertions on its output with `gbytes`.  The only thing to be aware of is potential interactions between the multiple publisher processes if they happen to access some sort of shared singleton resources...  Keep reading!

#### Managing External Resources in Parallel Suites: Files

The filesystem is a shared singleton resource.  Each parallel process in a parallel spec run will have access to the same shared filesystem.  As such it is important to avoid spec pollution caused by accidental collisions.  For example, consider the following publisher specs:

```go
Describe("Publishing books", func() {
  It("can publish a complete epub", func(ctx SpecContext) {
    cmd := exec.CommandContext(ctx, publisherPath, "-o=out.epub", "les-miserables.fixture")
    session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
    Expect(err).NotTo(HaveOccurred())
    Eventually(ctx, session).Should(gexec.Exit(0)) //with exit code 0

    result, err := epub.Load("out.epub")
    Expect(err).NotTo(HaveOccurred())
    Expect(result.EpubPages).To(Equal(2783))
  }, SpecTimeout(time.Second*30))

  It("can publish a preview that contains just the first chapter", func(ctx SpecContext) {    
    cmd := exec.CommandContext(ctx, publisherPath, "-o=out.epub", "--preview", "les-miserables.fixture")
    session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
    Expect(err).NotTo(HaveOccurred())
    Eventually(ctx, session).Should(gexec.Exit(0)) //with exit code 0

    result, err := epub.Load("out.epub")
    Expect(err).NotTo(HaveOccurred())
    Expect(result.EpubPages).To(BeNumerically("<", 2783))
    Expect(result.EpubContent).To(ContainSubstring("Chapter 1"))
    Expect(result.EpubContent).NotTo(ContainSubstring("Chapter 2"))
  })
})
```

these specs will always run fine in series - but can fail in subtle and confusing ways when run in parallel!  Since both publish to the same `out.epub` file simultaneously collisions are possible.

There are multiple ways to approach this.  Perhaps the obvious way would be to manually ensure a different output name for each spec:

```go
Describe("Publishing books", func() {
  It("can publish a complete epub", func(ctx SpecContext) {
    cmd := exec.CommandContext(ctx, publisherPath, "-o=complete.epub", "les-miserables.fixture")
    ...
  })

  It("can publish a preview that contains just the first chapter", func(ctx SpecContext) {    
    cmd := exec.CommandContext(ctx, publisherPath, "-o=preview.epub", "--preview", "les-miserables.fixture")
    ...
  })
})
```

that's... _ok_.  But it's asking for trouble by putting the namespacing burden on the user.

A better alternative would be to carve out a separate namespace for each spec.  For example, we could create a temporary directory:

```go
var tmpDir string
BeforeEach(func() {
  tmpDir = GinkgoT().TempDir()
})

Describe("Publishing books", func() {
  It("can publish a complete epub", func(ctx SpecContext) {
    path := filepath.Join(tmpDir, "out.epub")
    cmd := exec.CommandContext(ctx, publisherPath, "-o="+path, "les-miserables.fixture")
    ...
  })

  It("can publish a preview that contains just the first chapter", func(ctx SpecContext) {
    path := filepath.Join(tmpDir, "out.epub")
    cmd := exec.CommandContext(ctx, publisherPath, "-o="+path, "--preview", "les-miserables.fixture")
    ...
  })
})

```
(here we're using `GinkgoT().TempDir()` to access Ginkgo's implementation of `t.TempDir()` which cleans up after itself - there's no magic here.  You could have simply called `os.MkdirTemp` and cleaned up afterwards yourself.)

This approach works fine but has the sometimes unfortunate side-effect of placing your files in a random location which can make debugging a bit more tedious.

Another approach - and the one used by Ginkgo's own integration suite - is to use the current parallel process index to shard the filesystem:

```go
var pathTo func(path string) string

BeforeEach(func() {
  //shard based on our current process index.
  //this starts at 1 and goes up to N, the number of parallel processes.
  dir := fmt.Sprintf("./tmp-%d", GinkgoParallelProcess())
  os.MkdirAll(dir)
  DeferCleanup(os.RemoveAll, dir)
  pathTo = func(path string) string { return filepath.Join(dir, path)}
})

Describe("Publishing books", func() {
  It("can publish a complete epub", func(ctx SpecContext) {
    path := pathTo("out.epub")
    cmd := exec.CommandContext(ctx, publisherPath, "-o="+path, "les-miserables.fixture")
    ...
  })

  It("can publish a preview that contains just the first chapter", func(ctx SpecContext) {
    path := pathTo("out.epub")
    cmd := exec.CommandContext(ctx, publisherPath, "-o="+path, "--preview", "les-miserables.fixture")
    ...
  })
})
```

this will create a namespaced local temp directory and provides a convenience function for specs to access paths to the directory.  The directory is cleaned up after each spec.

One nice thing about this approach is our ability to preserve the artifacts in the temporary directory in case of failure.  A common pattern when debugging is to use `--fail-fast` to indicate that the suite should stop running as soon as the first failure occurs.  We can key off of that config to change  the behavior of our cleanup code:

```go
var pathTo func(path string) string

BeforeEach(func() {
  //shard based on our current process index.
  //this starts at 1 and goes up to N, the number of parallel processes.
  dir := fmt.Sprintf("./tmp-%d", GinkgoParallelProcess())
  os.MkdirAll(dir)

  DeferCleanup(func() {
    suiteConfig, _ := GinkgoConfiguration()
    if CurrentSpecReport().Failed() && suiteConfig.FailFast {
      GinkgoWriter.Printf("Preserving artifacts in %s\n", dir)
      return
    }
    Expect(os.RemoveAll(dir)).To(Succeed())
  })

  pathTo = func(path string) string { return filepath.Join(dir, path)}
})
```

now, the temporary directory will be preserved in the event of spec failure, but only if `--fail-fast` is configured.

#### Managing External Resources in Parallel Suites: Ports
Another shared singleton resources is the set of available ports on the local machine.  If you need to be able to explicitly specify a port to use during a spec (e.g. you're spinning up an external process that needs to be told what port to listen on) you'll need to be careful how you carve up the available set of ports.  For example, the following would not work:

```go
var libraryAddr string

BeforeSuite(func() {
  libraryAddr := "127.0.0.1:50000"
  library.Serve(listenAddr)
  client = library.NewClient(listenAddr)
})
```

when running in parallel each process will attempt to listen on port 50000 and a race with only one winner will ensue.  You could, instead, have the server you're spinning up figure out a free port to use and report it back - but that is not always possible in the case where a service must be explicitly configured.

Instead, you can key off of the current parallel process index to give each process a unique port.  In this case we could:

```go
var libraryAddr string

BeforeSuite(func() {
  libraryAddr := fmt.Sprintf("127.0.0.1:%d", 50000 + GinkgoParallelProcess())
  library.Serve(listenAddr)
  client = library.NewClient(listenAddr)
})
```

now each process will have its own unique port.

#### Patterns for Testing against Databases
Stateful services that store data in external databases benefit greatly from a robust comprehensive test suite.  Unfortunately, many testers shy away from full-stack testing that includes the database for fear of slowing their suites down.  Fake/mock databases only get you so far, however.  In this section we outline patterns for spinning up real databases and testing against them in ways that are parallelizable and, therefore, able to leverage the many cores in modern machines to keep our full-stack tests fast.

The core challenge with stateful testing is to ensure that specs do not pollute one-another.  This applies in the serial context where a one spec can change the state of the database in a way that causes a subsequent spec to fail.  This also applies in the parallel context where multiple specs can write to the same database at the same time in contradictory ways.  Thankfully there are patterns that make mitigating these sorts of pollution straightforward and transparent to the user writing specs.

Throughout these examples we have a `DBRunner` library that can spin up instances of a database and a `DBClient` library that can connect to that instance and perform actions.  We aren't going to pick any particular database technology as these patterns apply across most of them.

##### A Database for Every Spec

Here's an incredibly expensive but sure-fire way to make sure each spec has a clean slate of data:

```go
var client *DBClient.Client
var _ = BeforeEach(func() {
  db, err := DBRunner.Start()
  Expect(err).NotTo(HaveOccurred())
  DeferCleanup(db.Stop)

  client = DBClient.New(db)
  Expect(client.Connect()).To(Succeed())
  DeferCleanup(client.Disconnect)

  client.InitializeSchema()
})
```

Now, each spec will get a fresh running database, with a clean initialized schema, to use.  This will work - but will probably be quite slow, even when running in parallel.

##### A Database for Every Parallel Process

Instead, a more common pattern is to spin up a database for each parallel process and reset its state between specs.

```go
var client *DBClient.Client
var snapshot *DBClient.Snapshot
var _ = BeforeSuite(func() {
  db, err := DBRunner.Start()
  Expect(err).NotTo(HaveOccurred())
  DeferCleanup(db.Stop)

  client = DBClient.New(db)
  Expect(client.Connect()).To(Succeed())
  DeferCleanup(client.Disconnect)

  client.InitializeSchema()
  snapshot, err = client.TakeSnapshot()
  Expect(err).NotTo(HaveOccurred())
})

var _ = BeforeEach(func() {
  Expect(client.RestoreSnapshot(snapshot)).To(Succeed())
})
```

here we've assumed the `client` can take and restore a snapshot of the database.  This could be as simple as truncating tables in a SQL database or clearing out a root key in a hierarchical key-value store.  Such methods are usually quite _fast_ - certainly fast enough to warrant full-stack testing over mock/fake-heavy testing.

With this approach each parallel process has its own dedicated database so there is no chance for cross-spec pollution when running in parallel.  Within each parallel process the dedicated database is cleared out between specs so there's no chance for spec pollution from one spec to the next.

This all works if you have the ability to spin up a local copy of the database.  But there are times when you must rely on an external stateful singleton resource and need to test against it.  We'll explore patterns for testing those next.

#### Patterns for Testing against Singletons
There are times when your spec suite must run against a stateful shared singleton system.  Perhaps it is simply too expensive to spin up multiple systems (e.g. each "system" is actually a memory-hungry cluster of distributed systems; or, perhaps, you are testing against a real-life instance of a service and can't spin up another instance).

In such cases the recommended pattern for ensuring your specs are parallelizable is to embrace sharding the external service by the parallel process index.  Exactly how this is done will depend on the nature of the system.

Here are some examples to give you a sense for how to approach this:

- If you're testing against a shared hierarchical key-value store (in which the keys are represented as `/paths/to/values` - e.g. S3, etcd) you can write your specs and code to accept a configurable root key such that all values are stored under `/{ROOT}/path/to/value`.  The suite can then configure `ROOT = fmt.Sprintf("test-%d", GinkgoParallelProcess())`
- If you're testing an external multi-tenant service you can have your suite create a unique tenant per parallel process.  Perhaps something like `service.CreateUser(fmt.Sprintf("test-user-%d", GinkgoParallelProcess()))`
- If you're testing an external service that supports namespace you can request a dedicated namespace per parallel process (e.g. a dedicated Cloud Foundry org and space, or a dedicated Kubernetes namespace).

The details will be context dependent - but generally speaking you should be able to find a way to shard access to the singleton system by `GinkgoParallelProcess()`.  You'll also need to figure out how to reset the shard between specs to ensure that each spec has a clean slate to operate from.

#### Some Subtle Parallel Testing Gotchas

We'll round out the parallel testing patterns with a couple of esoteric gotchas.

There's a somewhat obscure issue where an external process that outlives the current spec suite can cause the spec suite to hang mysteriously.  If you've hit that issue read through this [GitHub issue](https://github.com/onsi/gomega/issues/473) - there's likely a stdout/stderr pipe that's sticking around preventing Go's `cmd.Wait()` from returning.

When you spin up a process yourself you should generally have it pipe its output to `GinkgoWriter`.  If you pipe to `os.Stdout` and/or `os.Stderr` and the process outlives the current spec you'll cause Ginkgo's output interceptor to hang.  Ginkgo will actually catch this and print out a long error message telling you what to do.  You can learn more on the associated [GitHub issue](https://github.com/onsi/ginkgo/issues/851)

### Benchmarking Code

Go's built-in `testing` package provides support for running `Benchmark`s.  Earlier versions of Ginkgo subject-node variants that were able to mimic Go's `Benchmark` tests.  As of Ginkgo 2.0 these nodes are no longer available.  Instead, Ginkgo users can benchmark their code using Gomega's substantially more flexible `gmeasure` package.  If you're interested, check out the `gmeasure` [docs](https://onsi.github.io/gomega/#gmeasure-benchmarking-code).  Here we'll just provide a quick example to show how `gmeasure` integrates into Ginkgo's reporting infrastructure.

`gmeasure` is structured around the metaphor of Experiments.  With `gmeasure` you create ``Experiments` that can record multiple named `Measurements`.  Each named `Measurement` can record multiple values (either `float64` or `duration`).  `Experiments` can then produce reports to show the statistical distribution of their `Measurements` and different `Measurements`, potentially from different `Experiments` can be ranked and compared.  `Experiments` can also be cached using an `ExperimentCache` - this can be helpful to avoid rerunning expensive experiments _and_ to save off "gold-master" experiments to compare against to identify potential regressions in performance - orchestrating all that is left to the user.

Here's an example where we profile how long it takes to repaginate books:

```go

Describe("Repaginating Books", func() {
  var book *books.Book
  BeforeEach(func() {
    book = LoadFixture("les-miserables.json")
  })

  // this is a spec that validates the behavior is correct
  // note that we can mix validation specs alongside performance specs
  It("can repaginate books", func() {
    Expect(book.CurrentFontSize()).To(Equal(12))
    originalPages := book.Pages

    book.SetFontSize(10)
    Expect(book.RecomputePages()).To(BeNumerically(">", originalPages))
  })

  // this is our performance spec.  we mark it as Serial to ensure it does not run in
  // parallel with other specs (which could affect performance measurements)
  // we also label it with "measurement" - this is optional but would allow us to filter out
  // measurement-related specs more easily
  It("repaginates books efficiently", Serial, Label("measurement"), func() {
    //we create a new experiment
    experiment := gmeasure.NewExperiment("Repaginating Books")

    //Register the experiment as a ReportEntry - this will cause Ginkgo's reporter infrastructure
    //to print out the experiment's report and to include the experiment in any generated reports
    AddReportEntry(experiment.Name, experiment)

    //we sample a function repeatedly to get a statistically significant set of measurements
    experiment.Sample(func(idx int) {
      book = LoadFixture("les-miserables.json") //always start with a fresh copy
      book.SetFontSize(10)

      //measure how long it takes to RecomputePages() and store the duration in a "repagination" measurement
      experiment.MeasureDuration("repagination", func() {
        book.RecomputePages()
      })
    }, gmeasure.SamplingConfig{N:20, Duration: time.Minute}) //we'll sample the function up to 20 times or up to a minute, whichever comes first.
  })
})
```

Now when this spec runs Ginkgo will print out a report detailing the experiment:

```bash
Will run 1 of 1 specs
------------------------------
• [2.029 seconds]
Repaginating Books repaginates books efficiently [measurement]
/path/to/books_test.go:19

  Begin Report Entries >>
  Repaginating Books - /path/to/books_test.go:21 @ 11/04/21 13:42:57.936
    Repaginating Books
    Name          | N  | Min   | Median | Mean  | StdDev | Max
    ==========================================================================
    repagination [duration] | 20 | 5.1ms | 104ms  | 101.4ms | 52.1ms | 196.4ms
  << End Report Entries
```

This is helpful - but the real value in a performance suite like this would be to capture possible performance regressions.  There are multiple ways of doing this - you could use an [Experiment Cache](https://onsi.github.io/gomega/#caching-experiments) and make the suite [configurable](#supporting-custom-suite-configuration) such that a baseline experiment is stored to disk when the suite is so configured.  Then, when the suite runs, it simply loads the baseline from the cache and compares it to the measured duration.  Ginkgo's own performance suite does this.

Alternatively you can just hard-code an expected value after running the experiment and make an appropriate assertion.  For example:

```go
It("repaginates books efficiently", Serial, Label("measurement"), func() {
  experiment := gmeasure.NewExperiment("Repaginating Books")
  AddReportEntry(experiment.Name, experiment)

  experiment.Sample(func(idx int) {
    book = LoadFixture("les-miserables.json")
    book.SetFontSize(10)

    experiment.MeasureDuration("repagination", func() {
      book.RecomputePages()
    })
  }, gmeasure.SamplingConfig{N:20, Duration: time.Minute})

  //we get the median repagination duration from the experiment we just ran
  repaginationStats := experiment.GetStats("repagination")
  medianDuration := repaginationStats.DurationFor(gmeasure.StatMedian)

  //and assert that it hasn't changed much from ~100ms
  Expect(medianDuration).To(BeNumerically("~", 100*time.Millisecond, 50*time.Millisecond))
})
```

now the spec will fail if the pagination time ever changes drastically from its measured value.  Of course the actual runtime will depend on the machine and test environment you're running on - so some caveats will apply.  Nonetheless an upper bound spec such as:

```go
Expect(medianDuration).To(BeNumerically("<", 300*time.Millisecond))
```

could still be a useful smoketest to catch any major regressions early in the development cycle.

### Building Custom Matchers
As you've seen throughout this documentation, Gomega allows you to write expressive assertions.  You can build on Gomega's building blocks to construct custom matchers tuned to the semantics of your codebase.

One way to do this is by implementing Gomega's `GomegaMatcher` interface.

A simpler, alternative, however, is to simply compose matchers together in a simple function.  For example, let's write a matcher that asserts that our book is valid, has a given title, author, and page-count.  Rather than repeat this all the time:

```go
Expect(book.IsValid()).To(BeTrue())
Expect(book.Title).To(Equal("Les Miserables"))
Expect(book.Author).To(Equal("Victor Hugo"))
Expect(book.Pages).To(Equal(2783))
```

we can implement a function that returns a composite Gomega matcher:

```go
func BeAValidBook(title string, author string, pages int) types.GomegaMatcher {
  return And(
    WithTransform(func(book *books.Book) bool {
      return book.IsValid()
    }, BeTrue()),
    HaveField("Title", Equal(title)),
    HaveField("Author", Equal(author)),
    HaveField("Pages", Equal(pages)),
  )
}
```

this function uses Gomega's `And` matcher to require that the four passed-in matchers are satisfied.  It then uses `WithTransform` to accept the passed-in book and call it's `IsValid()` method, then asserts the returned value is `true`.  It then uses the `HaveField` matcher to make assertions on the fields within the `Book` struct.

Now we can write:

```go
Expect(book).To(BeAValidBook("Les Miserables", "Victor Hugo", 2783))
```

We can go one step further and use typed parameters to pick and choose which pieces of `Book` we want to test with our matcher.  This is a bit contrived for our simple example but can be quite useful in more complex domains:

```go

type Title string
type Author string
type Pages int

func BeAValidBook(params ...interface{}) types.GomegaMatcher {
  matchers := []types.GomegaMatcher{
    WithTransform(func(book *books.Book) bool {
      return book.IsValid()
    }, BeTrue())
  }

  if len(params) > 0 {
    for _, param := range params {
      switch v := param.(type) {
      case Title:
        matchers = append(matchers, HaveField("Title", Equal(v)))
      case Author:
        matchers = append(matchers, HaveField("Author", Equal(v)))
      case Pages:
        matchers = append(matchers, HaveField("Pages", Equal(v)))
      default:
        Fail("Unknown type %T in BeAValidBook() \n", v)
      }
    }
  }

  return And(matchers...)
}
```

Now we can do things like:

```go
Expect(book).To(BeAValidBook()) //simply asserts IsValid() is true
Expect(book).To(BeAValidBook(Title("Les Miserables")))
Expect(book).To(BeAValidBook(Author("Victor Hugo")))
Expect(book).To(BeAValidBook(Title("Les Miserables"), Pages(2783)))
```

The failure messages generated by composed matchers are generally good enough to capture the reason for the failure.  However if you want more fine-control over the message, or if you want more complex logic in your matcher you can use [`gcustom`](https://onsi.github.io/gomega/#gcustom-a-convenient-mechanism-for-buildling-custom-matchers) to build custom matchers using a simple function and templates - to learn more check out the [`gucstom` docs](https://onsi.github.io/gomega/#gcustom-a-convenient-mechanism-for-buildling-custom-matchers) and [godoc](https://pkg.go.dev/github.com/onsi/gomega/gcustom). 

## Decorator Reference
We've seen a number of Decorators detailed throughout this documentation.  This reference collects them all in one place.

#### Node Decorators Overview
Ginkgo's container nodes, subject nodes, and setup nodes all accept decorators.  Decorators are specially typed arguments passed into the node constructors.  They can appear anywhere in the `args ...interface{}` list in the constructor signatures:

```go
func Describe(text string, args ...interface{})
func It(text string, args ...interface{})
func BeforeEach(args ...interface{})
```

Ginkgo will vet the passed in decorators and exit with a clear error message if it detects any invalid configurations. 

Moreover, Ginkgo also supports passing in arbitrarily nested slices of decorators.  Ginkgo will unroll these slices and process the flattened list.  This makes it easier to pass around groups of decorators.  For example, this is valid:

```go
markFlaky := []interface{}{Label("flaky"), FlakeAttempts(3)}

var _ = Describe("a bunch of flaky controller tests", markFlaky, Label("controller"), func() {
  ...
}
```
The resulting tests will be decorated with `FlakeAttempts(3)` and the two labels `flaky` and `controller`.

#### The Serial Decorator
The `Serial` decorator applies to container nodes and subject nodes only.  It is an error to try to apply the `Serial` decorator to a setup node.

`Serial` allows the user to mark specs and containers of specs as only eligible to run in serial.  Ginkgo will guarantee that these specs never run in parallel with other specs.

If a container is marked as `Serial` then all the specs defined in that container will be marked as `Serial`.

You cannot mark specs and containers as `Serial` if they appear in an `Ordered` container.  Instead, mark the `Ordered` container as `Serial`.

#### The Ordered Decorator
The `Ordered` decorator applies to container nodes only.  It is an error to try to apply the `Ordered` decorator to a setup or subject node.

`Ordered` allows the user to [mark containers of specs as ordered](#ordered-containers).  Ginkgo will guarantee that the container's specs will run in the order they appear in and will never run in parallel with one another (though they may run in parallel with other specs unless the `Serial` decorator is also applied to the `Ordered` container).

When a spec in an `Ordered` container fails, all subsequent specs in the ordered container are skipped.  Only `Ordered` containers can contain `BeforeAll` and `AfterAll` setup nodes.

#### The ContinueOnFailure Decorator
The `ContinueOnFailure` decorator applies to outermost `Ordered` container nodes only.  It is an error to try to apply the `ContinueOnFailure` decorator to anything other than an `Ordered` container - and that `Ordered` container must not have any parent `Ordered` containers.

When an `Ordered` container is decorated with `ContinueOnFailure` then the failure of one spec in the container will not prevent other specs from running.  This is useful in cases where `Ordered` containers are being used to have share common (expensive) setup for a collection of specs but the specs, themselves, don't rely on one another.

#### The OncePerOrdered Decorator
The `OncePerOrdered` decorator applies to setup nodes only.  It is an error to try to apply the `OncePerOrdered` decorator to a container or subject node.

Normally, setup nodes like `BeforeEach` run for every spec in a suite.  When decorated with `OncePerOrdered`, however, `BeforeEach` will treat any `Ordered` container at a deeper nesting level as a single executable unit and run once before the container begins (mimicking the semantics of `BeforeAll`).  The usecases for this are covered in more detail in the [Setup around Ordered Containers: the OncePerOrdered Decorator](#setup-around-ordered-containers-the-onceperordered-decorator) section of the docs.

#### The Label Decorator
The `Label` decorator applies to container nodes and subject nodes only.  It is an error to try to apply the `Label` decorator to a setup node.  You can also apply the `Label` decorator to your `RunSpecs` invocation to annotate the entire suite with a label.

`Label` allows the user to annotate specs and containers of specs with labels.  The `Label` decorator takes a variadic set of strings allowing you to apply multiple labels simultaneously.  Labels are arbitrary strings that do not include the characters `"&|!,()/"`.  Specs can have as many labels as you'd like and the set of labels for a given spec is the union of all the labels of the container nodes and the subject node.

Labels can be used to control which subset of tests to run.  This is done by providing the `--label-filter` flag to the `ginkgo` CLI.  More details can be found at [Spec Labels](#spec-labels).

#### The Focus and Pending Decorators
The `Focus` and `Pending` decorators apply to container nodes and subject nodes only.  It is an error to try to `Focus` or `Pending` a setup node.

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

The `Focus` and `Pending` decorators are propagated through the test hierarchy as described in [Pending Specs](#pending-specs) and [Focused Specs](#focused-specs)

#### The Offset Decorator
The `Offset(uint)` decorator applies to all decorable nodes.  The `Offset(uint)` decorator allows the user to change the stack-frame offset used to compute the location of the test node.  This is useful when building shared test behaviors.  For example:

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

`Offset`s only apply to the node that they decorate.  Setting the `Offset` for a container node does not affect the `Offset`s computed in its child nodes.

If multiple `Offset`s are provided on a given node, only the last one is used.

Lastly, since introducing `Offset` Ginkgo has introduced `GinkgoHelper()` which marks the current function as a test helper who's location should be skipped when determining the location for a node.  We generally recommend using `GinkgoHelper()` instead of `Offset()` to manage how locations are computed.  The above example could be rewritten as

```go
SharedBehaviorIt := func() {
  GinkgoHelper()
  It("does something common and complicated", func() {
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

#### The CodeLocation Decorator
In addition to `Offset`, users can decorate nodes with a `types.CodeLocation`.  `CodeLocation`s are the structs Ginkgo uses to capture location information.  You can, for example, set a custom location using `types.NewCustomCodeLocation(message string)`.  Now when the location of the node is emitted the passed in `message` will be printed out instead of the usual `file:line` location.

Passing a `types.CodeLocation` decorator in has the same semantics as passing `Offset` in: it only applies to the node in question.

#### The FlakeAttempts Decorator
The `FlakeAttempts(uint)` decorator applies to container and subject nodes.  It is an error to apply `FlakeAttempts` to a setup node.

`FlakeAttempts` allows the user to flag specs trees as potentially flaky.  Ginkgo will retry the spec up to the number of times specified in `FlakeAttempts` until they pass.  For example:

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

With this setup, `"is flaky"` and `"is also flaky"` will run up to 3 times.  `"is _really_ flaky"` will run up to 5 times.  `"is _not_ flaky"` will run only once.  Note that if multiple `FlakeAttempts` appear in a spec's hierarchy, the most deeply nested `FlakeAttempts` wins.  If multiple `FlakeAttempts` are passed into a given node, the last one wins.

If `ginkgo --flake-attempts=N` is set the value passed in by the CLI will override all the decorated values.  Every spec in the test suite will now run up to `N` times.

#### The MustPassRepeatedly Decorator
The `MustPassRepeatedly(uint)` decorator applies to container and subject nodes.  It is an error to apply `MustPassRepeatedly` to a setup node.

the `MustPassRepeatedly` flag allows the user to repeatedly run specs in a controlled manner.  Ginkgo will repeatedly run specs up to the number of times specified in `MustPassRepeatedly` or until they fail.  For example:

```go
Describe("repeated specs", MustPassRepeatedly(3), func() {
  It("is repeated", func() {
    ...
  })

  It("is also repeated", func() {
    ...
  })

  It("is repeated even more", MustPassRepeatedly(5) func() {
    ...
  })

  It("is repeated less", MustPassRepeatedly(1), func() {
    ...
  })
})
```

With this setup, `"is repeated"` and `"is also repeated"` will run up to 3 times.  `"is repeated even more"` will run up to 5 times.  `"is repeated less"` will run only once.  Note that if multiple `MustPassRepeatedly` appear in a spec's hierarchy, the most deeply nested `MustPassRepeatedly` wins.  If multiple `MustPassRepeatedly` are passed into a given node, the last one wins.

The `ginkgo --repeat=N` value passed in by the CLI has no relation with the `MustPassRepeatedly` decorator. If the `--repeat` CLI flag is used and a container or subject node also contains the `MustPassRepeatedly` decorator, then the spec will run up to `N*R` times, where `N` is the values passed to the `--repeat` CLI flag and `R` is the value passed to the MustPassRepeatedly decorator.

If the `MustPassRepeatedly` decorator is set, it will override the `ginkgo --flake-attempts=N` CLI config. The specs that do not contain the `MustPassRepeatedly(R)` decorator will still run up to `N` times, in accordance to the `ginkgo --flake-attempts=N` CLI config.

#### The SuppressProgressOutput Decorator

When running with `ginkgo -v -progress` Ginkgo will emit information about each node just before it runs.   This information goes to the `GinkgoWriter` and straight to the console if using `-v`.  There are contexts when this can be overly noisy.  In particular, `ReportBeforeEach` and `ReportAfterEach` nodes always run, even when a spec is skipped.  This can make Ginkgo's output noise when running with `-v -progress` as each `Report*Each` node will be announced, even for skipped specs.

The `SuppressProgressOutput` decorator allows you to disable progress reporting for a given node:

```go
ReportAfterEach(func(report SpecReport) {
   //...
}, SuppressProgressReporting)
```

#### The PollProgressAfter and PollProgressInterval Decorators

As described in the [Getting Visibility Into Long-Running Specs](#getting-visibility-into-long-running-specs) section, the globally specified values for `--poll-progress-after` and `--poll-progress-interval` can be overridden on a particular node using the `PollProgressAfter(INTERVAL)` and `PollProgressInterval(INTERVAL)` decorators.  Here, `INTERVAL` is a `time.Duration` and when specified Ginkgo will start emitting Progress Reports for the node after a duration of `PollProgressAfter` and will repeatedly emit a Progress Report at an interval of `PollProgressInterval`.  To turn off progress reporting for a given node, set `PollProgressAfter` to `0`.

Both of these decorators can only be used on subject and setup nodes, not container nodes.

#### The SpecTimeout, NodeTimeout, and GracePeriod Decorators

As described in the [Spec Timeouts and Interruptible Nodes](#spec-timeouts-and-interruptible-nodes) section, Ginkgo allows you to decorate interruptible nodes with individual `NodeTimeout`s and spec-wide `SpecTimeout`s.  `NodeTimeout` takes a `time.Duration` and applies to any interruptible node (i.e. a node with a function that accepts a `SpecContext`).  `SpecTimeout` also takes a `time.Duration` but applies only to `It` subject nodes.  Whereas `NodeTimeout` specified a deadline for an individual node, `SpecTimeout` specifies a deadline for all nodes associated with an individual spec.

Once interrupted, Ginkgo waits for a Grace Period before abandoning a node and moving on.  A global Grace Period can be specified via the `--grace-period=DURATION` cli flag and overridden by the `GracePeriod` decorator on a per-node basis.  `GracePeriod` takes a `time.Duration` and can only be applied to interruptible nodes.

Currently none of these decorators can be applied to container nodes.

## Ginkgo CLI Overview

This chapter provides a quick overview and tour of the Ginkgo CLI.  For comprehensive details about all of the Ginkgo CLI's flags, run `ginkgo help`.  To get information about Ginkgo's implicit `run` command (i.e. what you get when you just run `ginkgo`) run `ginkgo help run`.

The Ginkgo CLI is the recommended and supported tool for running Ginkgo suites.  While you _can_ run Ginkgo suites with `go test` you must use the CLI to run suites in parallel and to aggregate profiles.  There are also a (small) number of `go test` flags that Ginkgo does not support - an error will be emitted if you attempt to use these (for example, `go test -count=N`, use `ginkgo -repeat=N` instead).

In addition to Ginkgo's own flags, the `ginkgo` CLI also supports passing through (nearly) all `go test` flags and `go build` flags.  These are documented under `ginkgo help run` and `ginkgo help build` (which provides a detailed list of available `go build` flags).  If you think Ginkgo's missing anything, please open an [issue](https://github.com/onsi/ginkgo/issues/new).

### Running Specs

By default:

```bash
ginkgo
```

Will run the suite in the current directory.

You can run multiple suites by passing them in as arguments:

```bash
ginkgo path/to/suite path/to/other/suite
```

or by running:

```bash
ginkgo -r
#or
ginkgo ./...
```

which will recurse through the current file tree and run any suites it finds.

To pass additional arguments or custom flags down to your suite use `--` to separate your arguments from arguments intended for `ginkgo`:

```bash
ginkgo -- <PASS-THROUGHS>
```

Finally, note that any Ginkgo flags must appear _before_ the list of packages.  Putting it all together:

```bash
ginkgo <GINKGO-FLAGS> <PACKAGES> -- <PASS-THROUGHS>
```

By default Ginkgo is running the `run` subcommand.  So all these examples can also be written as `ginkgo run <GINKGO-FLAGS> <PACKAGES> -- <PASS-THROUGHS>`.  To get help about Ginkgo's run flags you'll need to run `ginkgo help run`.

### Precompiling Suites

It is often convenient to precompile suites and distribute them as binaries.  You can do this with `ginkgo build`:

```bash
ginkgo build path/to/suite /path/to/other/suite
```

This will produce precompiled binaries called `package-name.test`.  You can then run `ginkgo package-name.test` _or_ `./package-name.test` to invoke the binary without going through a compilation step.

Since the `ginkgo` CLI is a [necessary component when running specs in parallel](#spec-parallelization) to run precompiled specs in parallel you must:

```bash
ginkgo -p ./path/to/suite.test
```

As with the rest of the go tool chain, you can cross-compile and target different platforms using the standard `GOOS` and `GOARCH` environment variables.  For example: 

```bash
GOOS=linux GOARCH=amd64 ginkgo build path/to/package
```

will build a linux binary.

Finally, the `build` command accepts a subset of the flags of the `run` command.  This is because some flags apply at compile time whereas others apply at run-time only.  This can be a bit confusing with the `go test` toolchain but Ginkgo tries to make things clearer by carefully controlling the availability of flags across the two commands.

### Watching for Changes

To help enable a fast feedback loop during development, Ginkgo provides a `watch` subcommand that watches suites and their dependencies for changes.  When a change is detected `ginkgo watch` will automatically rerun the suite.

`ginkgo watch` accepts most of `ginkgo run`'s flags.  So, you can do things like:

```bash
ginkgo watch -r -p
```

to monitor all packages, recursively, for changes and run them in parallel when changes are detected.

For each monitored package, Ginkgo also monitors that package's dependencies.  By default `ginkgo watch` monitors a package's immediate dependencies.  You can adjust this using the `-depth` flag.  Set `-depth` to `0` to disable monitoring dependencies and set `-depth` to something greater than `1` to monitor deeper down the dependency graph.


### Generators

As discussed above, Ginkgo provides a pair of generator functions to help you bootstrap a suite and add a spec file to it:

```bash
ginkgo bootstrap
```

will generate a file named `PACKAGE_suite_test.go` and

```bash
ginkgo generate <SUBJECT>
```

will generate a file named `SUBJECT_test.go` (or `PACKAGE_test.go` if `<SUBJECT>` is not provided).  Both generators support custom templates using `--template` 
and the option to provide extra custom data to be rendered into the template, besides the default values, using `--template-data`. The custom data should be a well structured JSON file. When loaded into the template the custom data will be available to access from the global key `.CustomData`. For example,
with a JSON file
```json
{ "suitename": "E2E",
  "labels": ["fast", "parallel", "component"]}
```
The custom data can be accessed like so:
`{{ .CustomData.suitename }}` or `{{ range .CustomData.labels }} {{.}} {{ end }}`

Take a look at the [Ginkgo's CLI code](https://github.com/onsi/ginkgo/tree/master/ginkgo/generators) to see what's available in the template.

### Creating an Outline of Specs

If you want to see an outline of the Ginkgo specs in an individual file, you can use the `ginkgo outline` command:

```bash
ginkgo outline book_test.go
```

This generates an outline in a comma-separated-values (CSV) format. Column headers are on the first line, followed by Ginkgo containers, specs, and other identifiers, in the order they appear in the file:

  Name,Text,Start,End,Spec,Focused,Pending,Labels
  Describe,Book,124,973,false,false,false,""
  BeforeEach,,217,507,false,false,false,""
  Describe,Categorizing book length,513,970,false,false,false,""
  Context,With more than 300 pages,567,753,false,false,false,""
  It,should be a novel,624,742,true,false,false,""
  Context,With fewer than 300 pages,763,963,false,false,false,""
  It,should be a short story,821,952,true,false,false,""

The columns are:

- Name (string): The name of a container, spec, or other identifier in the core DSL.
- Text (string): The description of a container or spec. (If it is not a literal, it is undefined in the outline.)
- Start (int): Position of the first character in the container or spec.
- End (int): Position of the character immediately after the container or spec.
- Spec (bool): True, if the identifier is a spec.
- Focused (bool): True, if focused. (Conforms to the rules in [Focused Specs](#focused-specs).)
- Pending (bool): True, if pending. (Conforms to the rules in [Pending Specs](#pending-specs).)
- Labels (string): If labels are assigned to nodes then will be shown as double quoted comma separated values. (Conforms to the rules in [Spec Labels](#spec-labels).)

You can set a different output format with the `-format` flag. Accepted formats are `csv`, `indent`, and `json`. The `ident` format is like `csv`, but uses indentation to show the nesting of containers and specs. Both the `csv` and `json` formats can be read by another program, e.g., an editor plugin that displays a tree view of Ginkgo tests in a file, or presents a menu for the user to quickly navigate to a container or spec.

`ginkgo outline` is intended for integration with third-party libraries and applications.  If you simply want to know how a suite will run without running it try `ginkgo -v --dry-run` instead.

### Other Subcommands

To unfocus any programmatically focused specs in the current directory or subdirectories, run:

```bash
ginkgo unfocus
```

To get a list of `Label`s used in a suite run

```bash
ginkgo labels
```

`labels` (naively) parses your spec files and looks for calls to the `Label` decorator.

To get the current version of the `ginkgo` CLI run:

```bash
ginkgo version
```

## Third-Party Integrations

### Using Third-party Libraries

Most third-party Go `testing` integrations (e.g. matcher libraries, mocking libraries) take and wrap a `*testing.T` to provide functionality.  Unfortunately there is no formal interface for `*testing.T` however Ginkgo provides a function, `GinkgoT()` that returns a struct that implements all the methods that `*testing.T` implements.  Most libraries accept the `*testing.T` object via an interface and you can usually simply pass in `GinkgoT()` and expect the library to work.

For example, you can choose to use [testify](https://github.com/stretchr/testify) instead of Gomega like so:

```go
package foo_test

import (
  . "github.com/onsi/ginkgo/v2"

  "github.com/stretchr/testify/assert"
)

var _ = Describe(func("foo") {
  It("should testify to its correctness", func(){
    assert.Equal(GinkgoT(), foo{}.Name(), "foo")
  })
})
```

Similarly if you're using [Gomock](https://code.google.com/p/gomock/) you can simply pass `GinkgoT()` to your controller:


```go
import (
  "code.google.com/p/gomock/gomock"

  . "github.com/onsi/ginkgo/v2"
  . "github.com/onsi/gomega"
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


  It("should consume things", func() {
    mockThing.EXPECT().OmNom()
    consumer.Consume()
  })
})
```

Since `GinkgoT()` implements `Cleanup()` (using `DeferCleanup()` under the hood) Gomock will automatically register a call to `mockCtrl.Finish()` when the controller is created.

When using Gomock you may want to run `ginkgo` with the `-trace` flag to print out stack traces for failures which will help you trace down where, in your code, invalid calls occurred.

`GinkgoT()` also provides additional methods that are Ginkgo-specific.  This allows rich third-party integrations to be built on top of Ginkgo - with GinkgoT() serving as a single connection point.

### IDE Support
Ginkgo works best from the command-line, and [`ginkgo watch`](#watching-for-changes) makes it easy to rerun tests on the command line whenever changes are detected.

There are a set of [completions](https://github.com/onsi/ginkgo-sublime-completions) available for [Sublime Text](https://www.sublimetext.com/) (just use [Package Control](https://sublime.wbond.net/) to install `Ginkgo Completions`) and for [VS Code](https://code.visualstudio.com/) (use the extensions installer and install vscode-ginkgo).  There is also a VS Code extension to run specs from VSCode called [Ginkgo Test Explorer](https://github.com/joselitofilho/ginkgoTestExplorer).

IDE authors can set the `GINKGO_EDITOR_INTEGRATION` environment variable to any non-empty value to enable coverage to be displayed for focused specs. By default, Ginkgo will fail with a non-zero exit code if specs are focused to ensure they do not pass in CI.

### The ginkgolinter
The [ginkgolinter](https://github.com/nunnatsa/ginkgolinter) enforces several patterns of using ginkgo and gomega. It can run as an independent executable or as part of the [golangci-lint](https://golangci-lint.run/) linter. See the ginkgolinter [READMY](https://github.com/nunnatsa/ginkgolinter#readme) for more details.

{% endraw  %}

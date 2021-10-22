---
layout: default
title: Ginkgo
---
{% raw  %}
[Ginkgo](https://github.com/onsi/ginkgo) is a Go testing framework built to help you efficiently write expressive and comprehensive tests.  It is best paired with the [Gomega](https://github.com/onsi/gomega) matcher library.  When combined, Ginkgo and Gomega provide a rich and expressive DSL ([Domain-specific Language](https://en.wikipedia.org/wiki/Domain-specific_language)) for writing tests.

The narrative docs you are reading here are supplemented by the [godoc](https://pkg.go.dev/github.com/onsi/ginkgo) API-level docs.  We suggest starting here to build a mental model for how Ginkgo works and understand how the Ginkgo DSL can be used to solve real-world testing scenarios.  These docs are written assuming you are familiar with Go and the Go toolchain and that you are using Ginkgo V2 (V1 is no longer supported - see [here](https://onsi.github.com/ginkgo/MIGRATING_TO_V2) for the migration guide).

## Why Ginkgo?

This section captures some of Ginkgo's history and motivation.  If you just want to dive straight in, feel free to [jump ahead](#getting-started)!

Like all software projects, Ginkgo was written at a particular time and place to solve a particular set of problems.

The first commit to Ginkgo was made by [@onsi](https://github.com/onsi/) on August 19th, 2013 (to put that timeframe in perspective, it's roughly three months before [Go 1.2](https://golang.org/doc/go1.2) was released!)  Ginkgo came together in the highly collaborative environment fostered by Pivotal, a software company and consultancy that advocated for outcome-oriented software development built by balanced teams that embrace test-driven development.

Specifically, Pivotal was one of the lead contributors to Cloud Foundry.  A sprawling distributed system, originally written in Ruby, that was slowly migrating towards the emerging distributed systems language of choice: Go.  At the time (and, arguably, to this day) the landscape of Go's testing infrastructure was somewhat anemic.  For engineers coming from the rich ecosystems of testing frameworks such as [Jasmine](https://jasmine.github.io), [rspec](https://rspec.info), and [Cedar](https://github.com/cedarbdd/cedar) there was a need for a comprehensive testing framework with a mature set of matchers in Go.

The need was twofold: organizational and technical. As a growing organization Pivotal needed a shared testing framework to be used across its many teams writing Go.  Engineers jumping from one team to another needed to be able to hit the ground running; we needed fewer testing bikesheds and more shared testing patterns.  And our test-driven development culture put a premium on tests as first-class citizens: they needed to be easy to write, easy to read, and easy to maintain.

Moreover, the _nature_ of the code being built - complex distributed systems - required a testing framework that could provide for the needs unique to unit-testing and integration-testing such a system.  We needed to make testing [asynchronous behavior](https://onsi.github.io/gomega/#making-asynchronous-assertions) ubiqutous and straightforward.  We needed to have [parallelizable integration tests](#parallelism) to ensure our test run-times didn't get out of control.  We needed a test framework that helped us [suss out](#spec-permutation) flaky tests and fix them.

This was the context that led to Ginkgo.  Over the years the Go testing ecosystem has grown and evolved - sometimes [bringing](https://go.dev/blog/subtests) it [closer](https://golang.org/doc/go1.17#testing) to Ginkgo.  Throughout, the community's reactions to Ginkgo have been... interesting.  Some enjoy the expressive framework and rich set of matchers - for many the DSL is familiar and the `ginkgo` CLI is productive.  Others have found the DSL off-putting, arguing that Ginkgo is not "the Go way."  That's OK; the world is plenty large enough for options to abound :)

Happy Testing!

---

## Getting Started

In this section we  cover installing Ginkgo, Gomega, and the `ginkgo` CLI.  We bootstrap a Ginkgo suite, write our first spec, and run it.

### Installing Ginkgo

Ginkgo uses [go modules](https://go.dev/blog/using-go-modules).  To add Ginkgo to your project, assuming you have a `go.mod` file setup, just `go get` it:

```bash
$> go get github.com/onsi/ginkgo/ginkgo
$> go get github.com/onsi/gomega/...
```

This fetches Ginkgo and installs the `ginkgo` executable under `$GOBIN` - you'll want that on your `$PATH`.  It also fetches the core Gomega matcher library and its set of supporting libraries.

You should now be able to run `ginkgo version` at the command line and see the Ginkgo CLI emit a version number.

### Your First Ginkgo Suite

Ginkgo hooks into Go's existing `testing` infrastructure.  That means that Ginkgo specs live in `*_test.go` files, just like standard go tests.  However, instead of using `func TestX(t *testing.T) {}` to write your tests you use the Ginkgo and Gomega DSLs.  

We call a collection of Ginkgo specs in a given package a **Ginkgo suite**; and we use the word **spec** to talk about individual Ginkgo tests contained in the suite.  Though they're functionally interchangeable, we'll use the word "spec" instead of "test" to make a distinction between Ginkgo tests and traditional `testing` tests.

In most Ginkgo suites there is only one `TestX` function - the entry point for Ginkgo.  Let's bootstrap a Ginkgo suite to see what that looks like.

### Bootstrapping a Suite

Say you have a package named `books` that you'd like to add a Ginkgo suite to.  To bootstrap the suite run:

```bash
$> cd path/to/books
$> ginkgo bootstrap
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

Under the hood, `ginkgo` is simply calling `go test`.  While you _can_ run `go test` instead of the `ginkgo` CLI, Ginkgo has several capabilities that can only be accessed via `ginkgo`.  We generally recommend users embrace the `ginkgo` CLI and treat it as a first-class member of their testing toolchain.

Alright, we've successfully set up and run our first suite.  Of course that suite is empty, which isn't very interesting.  Let's add some specs.

#### Adding Specs to a Suite
While you can add all your specs directly into `books_suite_test.go` you'll generally prefer to place your specs in separate files.  This is particularly true if you have packages with multiple files that need to be tested.  Let's say our `book` package includes a `book.go` model and we'd like to test its behavior.  We can generate a test file like so:

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

Finally, we use **subject nodes** like `It` to write a spec that makes assertions about the subject under test.  In this case, we are ensuring that `book.Category()` returns the correct category ``enum` based on the length of the book.  We make these assertions using Gomega's `Equal` matcher and `Expect` syntax.  You can learn much more about [Gomega here](https://onsi.github.io/gomega/#making-assertions) - the examples used throughout these docs should be self-explanatory.

The container, setup, and subject nodes form a **spec tree**.  Ginkgo uses the tree to construct a flattened list of specs where each spec can have multiple setup nodes but will only have one subject node.

Because there are two subject nodes, Ginkgo will identify two specs to run.  For each spec, Ginkgo will run the closures attached to any associated setup nodes and then run the closure attached to the subject node.  In order to share state between the setup node and subject node we define closure variables like `lesMis` and `foxInSocks`.  This is a common pattern and the main way that tests are organized in Ginkgo.

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

Success!  We've written and run our first Ginkgo suite.  From here we can grow our test suite as we iterate on our code.

The next sections will delve into the many mechanisms Ginkgo provides for writing and running specs.

## Writing Specs

Ginkgo makes it easy to write expressive specs that describe the behavior of your code in an organized manner.  We've seen that Ginkgo suites are hierarchical collections of specs comprised of container nodes, setup nodes, and subject nodes organized into a spec tree.  In this section we dig into the various kinds of nodes available in Ginkgo and their properties.

### Spec Subjects: `It`
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

Ginkgo provides an alias for `It` called `Specify`.  `Specify` is functionally identical to `It` but can help your specs read more naturally.

### Extracting Common Setup: `BeforeEach`
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

When Ginkgo runs a spec it runs through all the `BeforeEach` closures that appear in that spec's hierarchy from the outer-most to the inner-most.  For the `both names` specs, Ginkgo will run the outermost `BeforeEach` closure before the subject node closure.  For the `one name` specs, Ginkgo will run the outermost `BeforeEach` closure and then the innermost `BeforeEach` closure which sets `book.Author = "Hugo"`.

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
                Expect(book.AuthorLastName()).To(Equal("Victor"))
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

Once the spec tree is constructed Ginkgo scans the tree to generate a flattened list of specs.  Each spec will have exactly one subject node and all its associated setup nodes.  At this point Ginkgo can randomize and filter the spec list (you'll learn more about that later) and is ready to run through the specs.

During the Run Phase Ginkgo runs through each spec in the flattened spec list  sequentially.  When running a spec Ginkgo invokes the setup and subject nodes in the correct order and tracks any failed assertions.  Container node closures are _never_ invoked during the run phase.

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
    book := &book.Book{ // No!
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

/* === DO THIS INSTEAD === */
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

### Separating Creation and Configuration: `JustBeforEach`

Let's get back to your growing Book suite and explore a few more Ginkgo nodes.  So far we've met the `BeforeEach` setup node, let's introduce it's closely related cousin: `JustBeforeEach`.

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
/* INVALID */
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

As with `BeforeEach` you can have multiple `JustBeforeEach` nodes at different levels of container nesting.  Ginkgo will first run all the `BeforeEach` closures from the outside in, then all the `JustBeforeEach` closures from the outside in.  While powerful and flexible overuse of `JustBeforeEach` (and nest `JustBeforeEach`es in particular!) can lead to confusing suites to be sure to use `JustBeforeEach` judiciously!d

### Spec Cleanup: `AfterEach` and `DeferCleanup`

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

#### Cleaning up our Cleanup code: `DeferCleanup`

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

As shown above `DeferCleanup` can be passed a function that takes no arguments and returns no value.  You can also pass a function that returns a single value.  `DeferCleanup` interprets this value as an error and fails the spec if the error is non-nil - a common go pattern.  This allows us to rewrite our example as:

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

#### Separating Diagnostics Collection and Teardown: `JustAfterEach`

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

### Suite Setup and Cleanup: `BeforeSuite` and `AfterSuite`

The setup nodes we've explored so far have all applied at the spec level.  They run Before**Each** or After**Each** spec in their associated container node.

It is common, however, to need to perform setup and cleanup at the level of the Ginkgo suite.  This is setup that should be performed just once - before any specs run, and cleanup that should be performed just once, when all the specs have finished.  Such code is particularly common in integration tests that need to prepare environments or spin up external resources.

Ginkgo supports suite-level setup and cleanup through two specialized **suite setup** nodes: `BeforeSuite` and `AfterSuite`.  These suite setup nodes **must** be called at the top-level of the suite and cannot be nested in containers.  Also there can be at most one `BeforeSuite` node and one `AfterSuite` node per suite.  It is idiomatic to place the suite setup nodes in the Ginkgo bootstrap suite file.

Let's continue to build out our book tests.  Books can be stored and retrieved from an external database and we'd like to test this behavior.  To do that, we'll need to spin up a database and set up a client to access it.  We can do that `BeforeEach` spec - but doing so would be prohibitively expensive and slow.  Instead, it would be more efficient to spin up the database just once when the suite starts.  Here's how we'd do it in our `books_suite_test.go` file:

```go
package books_test

import (
    . "github.com/onsi/ginkgo"
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

Ginkgo will run our `BeforeSuite` closure at the beginning of the [run phase](Mental Model: How Ginkgo Traverses the Spec Hierarchy) - i.e. after the spec tree has been constructed but before any specs have run.  This closure will instantiate a new `*db.Runner` - this is hypothetical code that knows how to spin up an instance of a database - and ask the runner to `Start()` a database.

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

> We won't get into it here but make sure to keep reading to understand how Ginkgo manages [suite parallelism](#parallelism) and provides [SynchronizedBeforeSuite and SynchronizedAfterSuite](#parallel-suite-setup-and-cleanup-codesynchronizedbeforesuitecode-and-codesynchronizedaftersuitecode) suite setup nodes.

### Mental Model: How Ginkgo Handles Failure
So far we've focused on how Ginkgo specs are constructed using nested nodes and how node closures are called in order when specs run.

...but Ginkgo is a testing framework.  And tests fail!  Let's delve into how Ginkgo handles failure.

You typically use a matcher library, like [Gomega](https://github.com/onsi/gomega) to make assertions in your spec.  When a Gomega assertion fails, Gomega generates a failure message and passes it to Ginkgo to signal that the spec has failed.  It does this via Ginkgo's global `Fail` function.  Of course, you're allowed to call this function directly yourself:

```
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

You must remember follow this pattern when making assertions in goroutines - however, if uncaught, Ginkgo's panic will include a helpful error to remind you to add `defer GinkgoRecover()` to your goroutine.

### Logging Output
As outlined above, when a spec fails - say via a failed Gomega assertion - Ginkgo will the failure message passed to the `Fail`  handler.  Often times the failure message generated by Gomega gives you enough information to understand and resolve the spec failure.

But there are several contexts, particularly when running large complex integration suites, where additional debugging information is necessary to understand the root cause of a failed spec.  You'll typically only want to see this information if a spec has failed - and hide it if the spec succeeds.

Ginkgo provides a globally available `io.Writer` called `GinkgoWriter` that solves for this usecase.  `GinkgoWriter` aggregates everything written to it while a spec is running and only emits to stdout if the test fails or is interrupted (via `^C`).

`GinkgoWriter` includes three convenience methods:

- `GinkgoWriter.Print(a ...interface{})` is equivalent to `fmt.Fprint(GinkgoWriter, a...)`
- `GinkgoWriter.Println(a ...interface{})` is equivalent to `fmt.Fprintln(GinkgoWriter, a...)`
- `GinkgoWriter.Printf(format string, a ...interface{})` is equivalent to `fmt.Fprintf(GinkgoWriter, format, a...)`

You can also attach additional `io.Writer`s for `GinkgoWriter` to tee to via `GinkgoWriter.TeeTo(writer)`.  Any data written to `GinkgoWriter` will immediately be sent to attached tee writers.  All attached Tee writers can be cleared with `GinkgoWriter.ClearTeeWriters()`.

Finally - when running in verbose mode via `ginkgo -v` anything written to `GinkgoWriter` will be immediately streamed to stdout.  This can help shorten the feedback loop when debugging a complex spec.

### Documenting Complex Specs: `By`

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

The string passed to `By` is emitted via the [`GinkgoWriter`](#logging-output).  If a test succeeds you won't see any output beyond Ginkgo's green dot.  If a test fails, however, you will see each step printed out up to the step immediately preceding the failure.  Running with `ginkgo -v` always emits all steps.

`By` takes an optional function of type `func()`.  When passed such a function `By` will immediately call the function.  This allows you to organize your `It`s into groups of steps but is purely optional.  

We haven't discussed [Report Entries](#attaching-data-to-reports) yet but we'll also mention that `By` also adds a `ReportEntry` to the running spec.  This ensures that the steps outlined in `By` appear in the structure JSON and JUnit reports that Ginkgo can generate.  If passed a function `By` will measure the runtime of the function and attach the resulting duration to the report as well.

`By` doesn't affect the structure of your specs - it's simply syntactic sugar to help you document long and complex specs.  Ginkgo has additional mechanisms to break specs up into more granular subunits with guaranteed ordering - we'll discuss [Ordered containers](#ordered-containers) in detail later.

### Table Specs

We'll round out this chapter on [Writing Specs](#writing-specs) with one last topic.  Ginkgo provides an expressive DSL for writing table driven specs.  This DSL is a simple wrapper around concepts you've already met - container nodes like `Describe` and subject nodes like `It`.

Let's write a table spec to describe the Author name functions we tested earlier:

```go
DescribeTable("Extracting the author's first and last name",
    func(author string, isValid bool, firstName string, lastName string) {
        book := &books.Book{
            Title: "My Book"
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

`DescribeTable` takes a string description, a **spec closure** to run for each table entry, and a set of entries.  Each `Entry` takes a string description, followed by a list of parameters.  `DescribeTable` will generate a spec for each `Entry` and when the specs run, the `Entry` parameters will be passed to the spec closure and must match the types expected by the the spec closure.

You'll be notified with a clear message at runtime if the parameter types don't match the spec closure signature.

#### Mental Model: Table Specs are just Synctatic Sugar
`DescribeTable` is simply providing syntactic sugar to convert its inputs into a set of standard Ginkgo nodes.  During the [Tree Construction Phase](#mental-model-how-ginkgo-traverses-the-spec-hierarchy) `DescribeTable` is generating a single container node that contains one subject node per table entry.  The description for the container node will be the description passed to `DescribeTable` and the descriptions for the subject nodes will be the descriptions passed to the `Entry`s.  During the Run Phase, when specs run, each subject node will simply invoke the spec closure passed to `DescribeTable`, passing in the parameters associated with the `Entry`.

To put it another way, the table test above is equivalent to:

```go
Describe("Extracting the author's first and last name", func() {
    It("When author has both names", func() {
        book := &books.Book{
            Title: "My Book"
            Author: "Victor Hugo",
            Pages: 10,
        }
        Expect(book.IsValid()).To(Equal(true))
        Expect(book.AuthorFirstName()).To(Equal("Victor"))
        Expect(book.AuthorLastName()).To(Equal("Hugo"))
    })

    It("When author has one name", func() {
        book := &books.Book{
            Title: "My Book"
            Author: "Hugo",
            Pages: 10,
        }
        Expect(book.IsValid()).To(Equal(true))
        Expect(book.AuthorFirstName()).To(Equal(""))
        Expect(book.AuthorLastName()).To(Equal("Hugo"))
    })

    It("When author has a middle name", func() {
        book := &books.Book{
            Title: "My Book"
            Author: "Victor Marie Hugo",
            Pages: 10,
        }
        Expect(book.IsValid()).To(Equal(true))
        Expect(book.AuthorFirstName()).To(Equal("Victor"))
        Expect(book.AuthorLastName()).To(Equal("Hugo"))
    })    

    It("When author has no name", func() {
        book := &books.Book{
            Title: "My Book"
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
/* INVALID */
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
        Entry("Novels", shelf["Fox in Socks"], books.CategoryShortStory),
    )
})
```

These specs will fail.  When `DescribeTable` and `Entry` are invoked during the Tree Construction Phase `shelf` will have been declared but uninitialized.  So `shelf["Les Miserables"]` will return a `nil` pointer and the spec will fail.

To get around this we must move access of the `shelf` variable into the body of the spec closure so that it can run at the appropriate time during the Run Phase.  We can do this like so:

```go
/* INVALID */
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

we're now accessing the `shelf` variable in the spec closure during the Run Phase and can trust that it has been correctly instantiated by the setup node closure.

Be sure to check out the [Table Patterns](#table-patterns) section of the [Ginkgo and Gomega Patterns](#ginkgo-and-gomega-patterns) chapter to learn about a few more table-based patterns.

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

Mercifully, Ginkgo's table DSL provides a few mechanisms to programatically generate entry descriptions.

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

## Running Specs
  ### Spec Permutation
  ### Parallelism
  #### Parallel Suite Setup and Cleanup: `SynchronizedBeforeSuite` and `SynchronizedAfterSuite`
  #### `ginkgo` CLI vs `go test`
  ### Pending Specs
  ### Skipping Specs
  (include skipping in BeforeSuite)
  ### Filtering Specs
    #### Programmatic Filtering
    (include tables!)
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
  ### Asynchronous Testing
  ### Managing External Processes in Parallel Test Suites
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

Ginkgo works best from the command-line, and [`ginkgo watch`](#watching-for-changes) makes it easy to rerun tests on the command line whenever changes are detected.

There are a set of [completions](https://github.com/onsi/ginkgo-sublime-completions) available for [Sublime Text](https://www.sublimetext.com/) (just use [Package Control](https://sublime.wbond.net/) to install `Ginkgo Completions`) and for [VSCode](https://code.visualstudio.com/) (use the extensions installer and install vscode-ginkgo).

IDE authors can set the `GINKGO_EDITOR_INTEGRATION` environment variable to any non-empty value to enable coverage to be displayed for focused specs. By default, Ginkgo will fail with a non-zero exit code if specs are focused to ensure they do not pass in CI.


{% endraw  %}
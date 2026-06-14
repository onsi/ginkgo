---
name: writing-specs
description: Author good Ginkgo specs — container nodes (Describe/Context/When), subject nodes (It/Specify), setup/cleanup nodes (BeforeEach, JustBeforeEach, AfterEach, DeferCleanup, BeforeSuite/AfterSuite), the "declare in container, initialize in BeforeEach" rule, separating creation from configuration, reusable test helpers with GinkgoHelper()/GinkgoHelperGo(), and By/GinkgoWriter output. Use when writing or reviewing specs or extracting a test helper. Covers the tree-construction-time pitfalls (no assertions/init/loop-capture in container bodies).
---

# Writing Ginkgo specs

Assumes Ginkgo is wired into the suite (`ginkgo:setup`) and you know the two-phase model (`ginkgo:overview`). Docs: <https://onsi.github.io/ginkgo/#writing-specs>.

## The shape of a spec

```go
var _ = Describe("Books", func() {
	var book *books.Book          // declare here...

	BeforeEach(func() {
		book = &books.Book{        // ...initialize here, fresh per spec
			Title:  "Les Miserables",
			Author: "Victor Hugo",
			Pages:  2783,
		}
		Expect(book.IsValid()).To(BeTrue()) // assertions in setup are fine
	})

	Describe("categorizing", func() {
		Context("with more than 300 pages", func() {
			It("is a novel", func() {
				Expect(book.Category()).To(Equal(books.CategoryNovel))
			})
		})
	})
})
```

- **Containers** (`Describe`/`Context`/`When`) organize; they're identical — pick the one that reads as a sentence. **Subjects** (`It`/`Specify`) hold the assertions; one spec runs per `It`.
- The spec's full name is the concatenation of every container text plus the `It` text — write them to read as a phrase.

## The rule that prevents most bugs: declare in container, initialize in setup

Container bodies run **once, at tree-construction time** (`ginkgo:overview`). So:

- **Declare** shared variables in the container body (`var book *books.Book`).
- **Initialize** them in `BeforeEach` so each spec gets a clean copy.

```go
// WRONG — runs once at construction; every spec shares & pollutes one book
var _ = Describe("Books", func() {
	book := &books.Book{Pages: 2783}   // initialized at construction time
	It("mutates", func() { book.Pages = 0 })
	It("expects 2783", func() { Expect(book.Pages).To(Equal(2783)) }) // flaky under randomization
})
```

The same trap explains: **no assertions in container bodies** (they fire at construction, with no spec active) and **no expensive/stateful work in container bodies**. If you see logic directly inside a `Describe`/`Context`/`When` body, it almost always belongs in a `BeforeEach`.

## Setup and cleanup nodes — and their ordering

| Node | Runs |
|---|---|
| `BeforeEach` | Before each spec, **outer→inner**. The workhorse. |
| `JustBeforeEach` | After all `BeforeEach`es, just before the `It`. |
| `JustAfterEach` | Just after the `It`, before any `AfterEach`. |
| `AfterEach` | After each spec, **inner→outer** (reverse). |
| `BeforeSuite`/`AfterSuite` | Once, around the whole suite (top-level only). |
| `BeforeAll`/`AfterAll` | Once per `Ordered` container (→ `ginkgo:ordering-and-flakes`). |

**`JustBeforeEach` separates creation from configuration.** Let nested `BeforeEach`es *configure* inputs into declared variables, and do the single *creation* step in `JustBeforeEach` — so each context overrides just the inputs it cares about:

```go
var jsonString string
BeforeEach(func() { jsonString = `{"id":1,"name":"Sally"}` })      // base config
JustBeforeEach(func() { user, err = NewUser(jsonString) })          // creation, runs last
Context("with malformed JSON", func() {
	BeforeEach(func() { jsonString = `{"oops"` })                  // override one input
	It("errors", func() { Expect(err).To(HaveOccurred()) })
})
```

Use it deliberately — deeply nested `JustBeforeEach`es get hard to follow.

## Cleanup: prefer DeferCleanup, and restore rather than clear

`DeferCleanup` registers teardown *next to* the setup that needs it, and runs in LIFO order (like `defer`). It works in any setup/subject node and adapts to scope (called in `BeforeSuite`, it cleans up after the suite; in `BeforeEach`, after the spec):

```go
BeforeEach(func() {
	original := os.Getenv("MODE")
	os.Setenv("MODE", "test")
	DeferCleanup(os.Setenv, "MODE", original) // captured args passed at cleanup time
})
```

- `DeferCleanup` accepts `func()`, `func() error` (a non-nil error fails the spec), captured arguments, and a `func(ctx SpecContext)` form (→ `ginkgo:timeouts-and-async`).
- **Restore original state; don't blindly clear it.** `os.Unsetenv` after the test wrongly assumes the var started unset — capture and restore instead (as above).
- `DeferCleanup` is a function call, not a node — it's the one cleanup mechanism you may use *inside* setup/subject closures. You may **not** define nodes (`It`, `BeforeEach`, …) inside a running closure.

## Output: GinkgoWriter and By

- **`GinkgoWriter`** buffers logs and only prints them when a spec **fails** (or always under `-v`) — so passing specs stay quiet. Use `GinkgoWriter.Printf(...)`, or `GinkgoWriter.TeeTo(w)` to also stream live. → `ginkgo:debugging-failures`.
- **`By("...")`** annotates steps in a long spec; the annotations surface on failure (and under `-v`) to show how far the spec got. It records into the spec's timeline.

```go
It("processes an order", func() {
	By("submitting the cart")
	// ...
	By("charging the card")
	// ...
})
```

## Failures, in brief

A failed Gomega assertion calls `Fail`, which **panics**; Ginkgo recovers it, marks the spec failed, and still runs cleanup. So code after a failed assertion in the same closure does not run. Use `Skip("reason")` to skip a spec at runtime (→ `ginkgo:filtering`). For failures inside goroutines and async polling, see `ginkgo:timeouts-and-async`.

## Test helpers — keep failure locations honest with GinkgoHelper()

Extract repeated setup or assertions into plain Go functions. The catch: **a `Fail` (or failed Gomega assertion) inside a helper reports the helper's own line** — useless for knowing *which call* failed. Mark the helper with **`GinkgoHelper()`** and Ginkgo skips that frame when computing the failure location, pointing at the spec that called it instead:

```go
func expectValidBook(b *books.Book) {
	GinkgoHelper()                    // this frame is ignored in failure locations
	Expect(b).NotTo(BeNil())
	Expect(b.IsValid()).To(BeTrue()) // a failure here is reported at the CALLER
}

It("accepts a good book", func() {
	expectValidBook(book)             // ← failures point here, not inside the helper
})
```

- **`GinkgoHelper()` composes.** Mark every helper in a chain (`expectValidBook` → `expectStorable` → …) and the reported location is always the spec that kicked it off.
- Prefer it over the older manual frame-counting: `Fail(msg, offset)` or the `Offset(n)` decorator (→ `ginkgo:decorators`). Those break the moment helpers call helpers (you'd have to bump every offset).

**A helper that fails from a goroutine** uses `GinkgoHelperGo` — it runs your code on a new goroutine, **already implies `defer GinkgoRecover()`** (→ `ginkgo:timeouts-and-async`), and gives you a `helperFail` to use for the *helper's own* failures (so they still report at the call site); caller-supplied assertions report inline:

```go
func EnsureSprockets(n int, fn func(int)) {
	GinkgoHelper()
	GinkgoHelperGo(func(helperFail func(string, ...int)) { // implies defer GinkgoRecover()
		if n == 0 {
			helperFail("sprockets must not be zero") // reported at the EnsureSprockets call site
		}
		fn(n)                                        // caller's assertions report inline
	})
}
```

(With Gomega, a helper's own assertions can run through `g := gomega.NewGomega(helperFail); g.Expect(...)`.)

## Where to go next

- Parameterize or generate specs → `ginkgo:tables-and-dynamic-specs`
- Decorate nodes (`Serial`, `Label`, timeouts, …) → `ginkgo:decorators`
- Run and filter what you wrote → `ginkgo:running`, `ginkgo:filtering`

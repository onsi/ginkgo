---
name: tables-and-dynamic-specs
description: Parameterize and generate Ginkgo specs — DescribeTable/Entry table-driven specs, Entry descriptions (string, nil, closure, EntryDescription), PEntry/FEntry and per-Entry decorators, DescribeTableSubtree, generating specs in a construction-time loop, loading fixtures in TestXxx before RunSpecs, and shared-behavior closures. Use when you have repetitive specs differing only by inputs, want data-driven or generated specs, or are extracting reusable It blocks across Contexts.
---

# Table specs and dynamically generated specs

Ginkgo gives you a DSL for table-driven specs plus idioms for generating specs from loops and data. All of it is **syntactic sugar that runs during the tree-construction phase** (`ginkgo:overview`) — the gotchas all follow from that.

Perfer `DescrtibeTable`/`DescribeTableSubtree` with shared configuration over multiple repeated `It`s.

Docs: <https://onsi.github.io/ginkgo/#table-specs>.

## DescribeTable + Entry

`DescribeTable(desc, specFunc, ...Entry)` generates one container holding one `It` per `Entry`. `Entry(desc, params...)` — its params are passed to `specFunc` at run time and **must match `specFunc`'s signature** (you get a clear runtime message if they don't).

```go
DescribeTable("Extracting the author's first and last name",
	func(author string, isValid bool, firstName, lastName string) {
		book := &books.Book{Title: "My Book", Author: author, Pages: 10}
		Expect(book.IsValid()).To(Equal(isValid))
		Expect(book.AuthorFirstName()).To(Equal(firstName))
		Expect(book.AuthorLastName()).To(Equal(lastName))
	},
	Entry("both names", "Victor Hugo", true, "Victor", "Hugo"),
	Entry("one name", "Hugo", true, "", "Hugo"),
	Entry("no name", "", false, "", ""),
)
```

A `DescribeTable` is just a container, so nest it inside `Describe`/`Context` and surround it with `BeforeEach` — setup runs fresh before each entry's spec.

### THE gotcha: Entry params are evaluated at construction time

`Entry(...)` arguments are evaluated when the tree is built — **before any `BeforeEach` has run**. So an `Entry` cannot read a variable initialized in `BeforeEach`; it will see the zero value (a `nil` map/pointer).

```go
var shelf map[string]*books.Book
BeforeEach(func() { shelf = loadShelf() }) // runs at RUN time

// WRONG — shelf is nil when Entry is evaluated at construction time
DescribeTable("category", func(b *books.Book, c books.Category) { ... },
	Entry("novel", shelf["Les Miserables"], books.CategoryNovel), // nil pointer!
)

// RIGHT — pass a key, dereference shelf inside the spec closure (run time)
DescribeTable("category", func(key string, c books.Category) {
	Expect(shelf[key].Category()).To(Equal(c))
},
	Entry("novel", "Les Miserables", books.CategoryNovel),
)
```

## The four ways to describe an Entry

| Mechanism | How |
|---|---|
| Explicit string | `Entry("both names", ...)` |
| **`nil`** | `Entry(nil, 1, 2, 3)` → auto-named from params: `Entry: 1, 2, 3` |
| Table-level description closure | pass `func(a,b,c int) string {...}` as 3rd arg to `DescribeTable`; renders every `nil` entry |
| `EntryDescription(fmt)` | `EntryDescription("%d + %d = %d")` as 3rd arg to `DescribeTable`, or per-entry |

The description closure must return `string` and accept the **same params** as `specFunc`. Per-entry, the first arg may itself be a closure or `EntryDescription` (overrides the table default):

```go
DescribeTable("addition", func(a, b, c int) { Expect(a + b).To(Equal(c)) },
	EntryDescription("%d + %d = %d"),                              // table default
	Entry(nil, 1, 2, 3),                                          // "1 + 2 = 3"
	Entry("zeros", 0, 0, 0),                                      // explicit
	Entry(EntryDescription("%[3]d = %[1]d + %[2]d"), 10, 100, 110), // per-entry override
	Entry(func(a, b, c int) string { return fmt.Sprintf("%d = %d", a+b, c) }, 4, 3, 7),
)
```

## Decorating entries

`Entry` and `DescribeTable` accept every decorator (→ `ginkgo:decorators`): `Entry("flaky case", FlakeAttempts(3), ...)`, `Entry(..., Label("slow"))`. Focus/pending shortcuts: `FEntry` / `PEntry` / `XEntry` (and `FDescribeTable` / `PDescribeTable`). **`PEntry` needs no params**; focus/pending obey the same precedence as everywhere else (→ `ginkgo:filtering`).

## DescribeTableSubtree — many Its per row

When you want a whole subtree (multiple `It`s, their own setup) per entry, use `DescribeTableSubtree`. Its body **runs at construction time, once per entry, inside a fresh container** — you must place `It`s inside it or no specs are generated:

```go
DescribeTableSubtree("handling requests",
	func(url string, code int, message string) {
		var resp *http.Response
		BeforeEach(func() {
			var err error
			resp, err = http.Get(url)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(resp.Body.Close)
		})
		It("returns the status code", func() { Expect(resp.StatusCode).To(Equal(code)) })
		It("returns the message", func() {
			body, _ := io.ReadAll(resp.Body)
			Expect(string(body)).To(Equal(message))
		})
	},
	Entry("default", "example.com/response", http.StatusOK, "hello world"),
	Entry("missing", "example.com/missing", http.StatusNotFound, "wat?"),
)
```

## Patterns

- **Struct-per-row** for many params — inscrutable positional entries like `Entry(nil, 12, 1.2, 8.5, 11, 2783)` are unreadable. Define a type and pass it: `Entry(nil, BookFormatting{FontSize: 12, LineHeight: 1.2, ...}, 2783)`.
- **Reusable `[]TableEntry`** — share one entry set across tables: `var InvalidBooks = []TableEntry{ Entry("empty", &books.Book{}), ... }`, then `DescribeTable("storing errors", storeFn, InvalidBooks)` and `DescribeTable("reading errors", readFn, InvalidBooks)`. Or feed the slice to `DescribeTableSubtree` to attach multiple specs per entry.

### Loading fixture data: do it in TestXxx, NOT BeforeSuite

If the spec **structure** depends on external data, that data must be available *during* tree construction. `BeforeSuite` runs in the *run* phase — too late; a loop reading a `BeforeSuite`-populated slice generates **zero specs**. Load it in the `TestXxx` bootstrap function before `RunSpecs`:

```go
var fixtureBooks []*books.Book

func TestBooks(t *testing.T) {
	RegisterFailHandler(Fail)
	g := NewGomegaWithT(t) // wrap t to assert before RunSpecs
	fixtureBooks = LoadFixturesFrom("./fixtures/books.json")
	g.Expect(fixtureBooks).NotTo(BeEmpty())
	RunSpecs(t, "Books Suite")
}

var _ = Describe("fixtures", func() {
	for _, book := range fixtureBooks { // populated before construction — works
		book := book
		It("stores "+book.Title, func() { Expect(library.Store(book)).To(Succeed()) })
	}
})
```

This works because `TestBooks` runs before tree construction, so `fixtureBooks` is populated when the loop runs, and because the function passed to `Describe` is not invoked until tree construction time. Don't load it in `BeforeSuite` — it's too late, and the loop generates no specs.

Docs: <https://onsi.github.io/ginkgo/#dynamically-generating-specs>.

## Shared behaviors

To reuse identical `It`s across `Context`s that differ only in setup, put the `It`s in a closure and call it inside each `Context` body (it runs at construction time, adding those specs to each context). Because the closure is defined in the same scope, it closes over the shared variable that each `Context`'s `BeforeEach` configures:

```go
AssertFailedBehavior := func() {
	It("can't be stored", func() { Expect(library.IsStorable(book)).To(BeFalse()) })
	It("fails to store", func() { Expect(library.Store(book)).To(MatchError(books.ErrStoringBook)) })
}
Context("when the book has no title", func() {
	BeforeEach(func() { book = &books.Book{Author: "Victor Hugo", Pages: 2783} })
	AssertFailedBehavior()
})
Context("when the book is nil", func() {
	BeforeEach(func() { book = nil })
	AssertFailedBehavior()
})
```

Docs: <https://onsi.github.io/ginkgo/#shared-behaviors>, <https://onsi.github.io/ginkgo/#table-specs-patterns>.

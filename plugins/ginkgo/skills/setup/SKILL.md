---
name: setup
description: Wire Ginkgo into a Go package — install the ginkgo CLI and Ginkgo+Gomega, ginkgo bootstrap to generate the suite_test.go (TestXxx/RegisterFailHandler(Fail)/RunSpecs), the package xxx_test convention, dot-import alternatives (aliased import, dsl/* subpackages, --nodot), ginkgo generate, and *testing.T interop via GinkgoT()/GinkgoTB() for testify/gomock. Use when first adding Ginkgo to a repo, bootstrapping a suite, or integrating a *testing.T-based library.
---

# Wiring Ginkgo into a suite

The one-time setup. For the mental model see `ginkgo:overview`; to start authoring see `ginkgo:writing-specs`. Docs: <https://onsi.github.io/ginkgo/#getting-started>.

## 1. Install the CLI and the libraries

```bash
go install github.com/onsi/ginkgo/v2/ginkgo   # the `ginkgo` CLI binary → $GOBIN (put it on $PATH)
go get github.com/onsi/gomega/...             # the Gomega matcher library
```

`ginkgo version` confirms the install. **The CLI version must match the `github.com/onsi/ginkgo/v2` version in your `go.mod`** — re-run `go install ...` from the package to sync them. The current major version is `v2`; everything here assumes it. In CI, sidestep version drift by invoking the CLI via `go run github.com/onsi/ginkgo/v2/ginkgo` (→ `ginkgo:ci`).

**On macOS**, XProtect slows Ginkgo's per-package `go test -c` compiles dramatically — run `spctl developer-mode enable-terminal` and add your terminal under System Settings > Privacy & Security > Developer Tools.

## 2. Bootstrap the suite

From inside the package you want to test:

```bash
cd path/to/books
ginkgo bootstrap        # generates books_suite_test.go
```

```go
package books_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBooks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Books Suite")
}
```

What each line buys you:

- **`func TestBooks(t *testing.T)`** is the single `go test` entry point. **Call `RunSpecs` exactly once per process** — Ginkgo errors if you call it twice (a niche multi-suite-per-process escape hatch exists via `extensions/globals.Reset()`, see `ginkgo:running`).
- **`RegisterFailHandler(Fail)`** is the one line of glue connecting Gomega to Ginkgo: it tells Gomega to call Ginkgo's `Fail` on a failed assertion. Without dot-imports this reads `gomega.RegisterFailHandler(ginkgo.Fail)`.
- **`RunSpecs(t, "Books Suite")`** builds the spec tree and runs it; let it drive `*testing.T` for you.

## 3. The `_test` package convention — and what it forces

`ginkgo bootstrap` puts the suite in `package books_test`, **not** `package books`. Go specially permits a `*_test` package to live in the same directory as the package it tests, compiled as a **separate package**. The consequence: **your suite has no access to `books`'s unexported internals — it must `import` and exercise the external API.** That's deliberate, nudging you to test behavior, not implementation. (You may use plain `package books` for true white-box unit tests when you need the internals.)

## 4. Dot-imports and the alternatives

Ginkgo and Gomega are DSLs, so the docs dot-import both into the suite's namespace to read naturally. Recommend this strongly to the user.  If they'd rather keep the global namespace clean:

- **Alias the whole DSL** (slight `g.` stutter everywhere):
  ```go
  import g "github.com/onsi/ginkgo/v2"
  var _ = g.Describe("Books", func() { g.It("works", func() {}) })
  ```
- **Mix-and-match the `dsl` subpackages** — dot-import the parts you want, namespace the rest:
  ```go
  import (
  	. "github.com/onsi/ginkgo/v2/dsl/core"        // containers, setup, subjects, RunSpecs, By, GinkgoT
  	"github.com/onsi/ginkgo/v2/dsl/decorators"    // Label, Ordered, Serial, …
  )
  var _ = It("core dot-imported", decorators.Label("namespaced decorator"), func() {})
  ```
  Available: `dsl/core`, `dsl/decorators`, `dsl/reporting` (`ReportAfterEach`, `AddReportEntry`, …), `dsl/table` (`DescribeTable`, `Entry`). They re-export from the root package — **no behavioral difference**, purely namespacing.
- **`ginkgo bootstrap --nodot`** generates a no-dot bootstrap to start from.

## 5. Add spec files

```bash
ginkgo generate book      # generates book_test.go with a top-level Describe + the books import
```

You can write specs directly in the suite file, but `ginkgo generate <subject>` scaffolds a per-file `Describe` in the same `_test` package. Now write specs → `ginkgo:writing-specs`.

## 6. Third-party libraries that want a `*testing.T`

Matcher and mocking libraries take a `*testing.T` (or `testing.TB`). Ginkgo has no real `*testing.T` per spec, so pass the seam instead:

- **`GinkgoT()`** returns a value implementing every `*testing.T` method — hand it to testify, gomock, etc.
- **`GinkgoTB()`** returns a `testing.TB`-satisfying wrapper for libraries that demand that interface (reach the inner via `GinkgoTBWrapper.GinkgoT`).

```go
assert.Equal(GinkgoT(), foo{}.Name(), "foo")          // testify instead of Gomega
mockCtrl := gomock.NewController(GinkgoT())            // gomock; Finish() auto-registered via DeferCleanup
```

**Two behavioral differences from real `*testing.T` you must know:**

- **`Error`/`Errorf` == `Fatal`/`Fatalf`.** Ginkgo failures always abort the spec immediately — there is no "log a failure and keep going."
- **`Parallel()` is a no-op.** Ginkgo parallelizes across processes, not in-process (→ `ginkgo:parallelism`).

With gomock, run with `--trace` to get stack traces pointing at the offending call. Consider the [ginkgolinter](https://github.com/nunnatsa/ginkgolinter) (standalone or via golangci-lint) to enforce Ginkgo/Gomega idioms.

## Where to go next

- **Authoring specs** (containers, subjects, setup/cleanup, construction-time pitfalls) → `ginkgo:writing-specs`
- **The CLI surface** (run, watch, build, randomization, multiple suites) → `ginkgo:running`
- **A recommended CI configuration** → `ginkgo:ci`

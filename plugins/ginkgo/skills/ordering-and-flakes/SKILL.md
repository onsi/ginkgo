---
name: ordering-and-flakes
description: Control spec ordering and manage flaky specs — Serial, Ordered containers with BeforeAll/AfterAll/ContinueOnFailure, OncePerOrdered, SpecPriority, plus FlakeAttempts/--flake-attempts, MustPassRepeatedly, --repeat, and --until-it-fails. Use when specs must run in a fixed order, you need once-per-group setup, you're combining Serial+Ordered, a spec is flaky, or you want to hunt order-dependence with --until-it-fails -p --randomize-all.
---

# Ordering and flakes

Two topics that share a root cause: specs that aren't truly independent (`ginkgo:overview`). Ordering decorators let you *declare* a dependency on purpose; flake controls let you *manage or expose* one. Docs: <https://onsi.github.io/ginkgo/#serial-specs>, <https://onsi.github.io/ginkgo/#ordered-containers>, <https://onsi.github.io/ginkgo/#repeating-spec-runs-and-managing-flaky-specs>.

## Serial: opt out of parallelism

`Serial` (container or subject) guarantees a spec never runs in parallel with anything else. Under the hood Ginkgo runs serial specs **last, on process #1**, after all other processes have exited. Reach for it sparingly.  OK for benchmarks, resource hogs, or specs that put an external resource into a known-bad state - but not OK as a band-aid for order-dependence or poorly isolated code/specs.

```go
Describe("Something expensive", Serial, func() {
	It("is a resource hog that can't share the box", func() { ... })
})
```

→ `ginkgo:decorators`, `ginkgo:parallelism`.

## Ordered: run specs in definition order

`Ordered` (container only) guarantees its child specs run **sequentially, in written order, on one process** — so they may legitimately mutate shared closure state. They can still parallelize against specs in *other* containers; only the inner order is pinned.

```go
Describe("checking out a book", Ordered, func() {
	var libraryClient *library.Client
	var book *books.Book

	BeforeAll(func() {                         // once, before the first spec
		libraryClient = library.NewClient()
		Expect(libraryClient.Connect()).To(Succeed())
		DeferCleanup(libraryClient.Disconnect) // context-aware: behaves like AfterAll
	})

	It("can fetch a book", func() { book, _ = libraryClient.FetchByTitle("Les Miserables") })
	It("can check it out", func() { Expect(library.CheckOut(book)).To(Succeed()) })
	It("is then out of stock", func() { ... })

	AfterAll(func() { ... })                   // once, after the last spec
})
```

- **`BeforeAll`/`AfterAll` are legal *only* inside an `Ordered` container** (or a container nested within one). That's the whole point of `Ordered` — once-per-group expensive setup.
- **A failing spec skips the rest of the group by default** (then `AfterAll` still runs). Spec independence is gone, so Ginkgo won't pretend the later specs are meaningful.
- **`ContinueOnFailure`** (outermost `Ordered` only — error on a nested container) overrides that: keep running later specs after a failure. Use it when `Ordered` is just shared setup, not a dependent flow. A failed `BeforeAll` still skips everything — the setup is presumed broken.
- Nested containers inside an `Ordered` container are automatically `Ordered`; there's no way to un-order them.

### OncePerOrdered: outer *Each setup that respects groups

A top-level `BeforeEach` that resets state runs before *every* spec — which breaks an `Ordered` group's first-spec-sets-up-rest contract (spec 1 passes, the rest get wiped). Decorate it with `OncePerOrdered` so it runs **once per independent unit**: still per-spec for normal specs, but once around each `Ordered` container (like a `BeforeAll`).

```go
BeforeEach(OncePerOrdered, func() {
	libraryClient = library.NewClient()
	snapshot := libraryClient.TakeSnapshot()
	DeferCleanup(libraryClient.RestoreSnapshot, snapshot)
})
```

Applies to `BeforeEach`/`JustBeforeEach`/`AfterEach`/`JustAfterEach` only — **not** `ReportBeforeEach`/`ReportAfterEach` (reporting is always per-spec).

### The Serial-inside-Ordered gotcha

**You cannot mark an individual spec inside an `Ordered` container `Serial`.** To make a whole ordered group run serially, put both decorators on the *container*: `Describe("...", Ordered, Serial, func(){...})`.

### Ordered vs one big It

`Ordered` and a single `It` with `By` steps (`ginkgo:writing-specs`) solve the same "dependent flow" problem. Prefer one `It` + `By` by default — it stays fully parallelizable and independent. Reach for `Ordered` when you want each step reported as its own spec, or genuinely need `BeforeAll` once-per-group setup. Either way, **use `Ordered` only when needed** — it gives up internal parallelism.

### SpecPriority

`SpecPriority(int)` (default `0`, higher runs earlier, may be negative) is a scheduling *hint* — e.g. "boulders first, then sand" in big suites. Honored per-spec under `--randomize-all`, otherwise per outermost container. **Anti-pattern: do not use it to impose a full deterministic order** — that's what `Ordered` is for, and it defeats randomization's whole purpose. → `ginkgo:decorators`.

## Flaky specs

**The stance: retries are explicit technical debt. Default to none — debug the flake; it's usually telling you something about your architecture.** Ginkgo gives you tools to both *expose* flakes and (reluctantly) *tolerate* them.

### Expose them

```bash
ginkgo --until-it-fails -p --randomize-all   # loop forever until something breaks — best flake hunter
ginkgo --repeat=N                            # run 1+N times, stop at first failure (CI-safe)
```

`--until-it-fails` compiles once and reruns indefinitely, so it's for local hunting, not CI; `--repeat=N` is the CI-safe bounded version. **Note `--repeat=N` runs `1+N` times** — and Ginkgo is **not** compatible with `go test --count`; use `--repeat`. → `ginkgo:ci`. Granular per-spec version: `MustPassRepeatedly(N)` (must pass N times in a row).

### Tolerate them (judiciously, as a last resort)

```bash
ginkgo --flake-attempts=3        # retry any failing spec up to 3x; a later pass = flaky, not failed
```

Per-spec: `FlakeAttempts(N)` decorator (CLI overrides it). On a pass-after-retry the suite stays green but reports the spec as flaky.

**FlakeAttempts × Ordered:** a failure in an `It` reruns just that `It` (not `BeforeAll`/`AfterAll`); a failure in a `BeforeAll` runs the `AfterAll` to clean up, then reruns the `BeforeAll`.

## Related

- Why independence matters in the first place → `ginkgo:overview`.
- Diagnosing *why* a spec failed/flaked from reports → `ginkgo:debugging-failures`.
- The full decorator signatures → `ginkgo:decorators`.

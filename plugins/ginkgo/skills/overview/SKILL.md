---
name: overview
description: The Ginkgo mental model for writing Go tests — the one idea that explains everything (Ginkgo builds a spec tree at construction time, then runs it) and its consequences for how you write specs, plus spec independence and the node taxonomy. Use this first when you start working with Ginkgo in a project, or to decide how to approach a Ginkgo task. Routes to the other ginkgo:* skills.
---

# Ginkgo: the mental model

Ginkgo is an expressive BDD-style testing framework for Go, paired with the [Gomega](https://onsi.github.io/gomega/) matcher library. You build suites out of nested **container** nodes (`Describe`/`Context`/`When`) and **subject** nodes (`It`), with **setup** nodes (`BeforeEach`/`AfterEach`/…) supplying state. You drive it with the `ginkgo` CLI.  Strongly prefer `ginkgo` over `go test`.

Read the canonical narrative docs at <https://onsi.github.io/ginkgo/> — they are the source of truth. This skill is the orientation; the other skills go deep.

## The one idea: tree construction, then running

**Ginkgo runs your suite in two distinct phases.** Internalizing this explains nearly every Ginkgo gotcha:

1. **Tree-construction phase.** Ginkgo invokes *every container body exactly once* to discover the structure of your suite. It collects — but does **not** run — the closures you pass to setup and subject nodes. The result is a tree it flattens into a list of specs.
2. **Run phase.** Ginkgo walks the flattened spec list (randomized, possibly in parallel) and, for each spec, runs its setup closures, then its one subject closure, then its cleanup closures.

**Container bodies run at construction time. Setup/subject closures run later, at run time.** The consequences you must internalize:

- **No assertions in container bodies.** They'd run once during construction, with no spec active — not as part of any test. Put assertions in `It` or `BeforeEach`.
- **No initialization in container bodies.** A variable set in a `Describe` body is set *once*, shared across every spec, and mutated by whichever spec runs first. **Declare in the container, initialize in `BeforeEach`** so every spec gets a pristine copy. → `ginkgo:writing-specs`.
- **Loops that build specs run at construction time** — that's how you generate specs dynamically, but it also means closure-captured loop variables and any data the loop reads must be available *then*, not in `BeforeSuite`. → `ginkgo:tables-and-dynamic-specs`.

## Spec independence — the assumption everything rests on

**Ginkgo assumes every spec is independent.** This is what lets it randomize order (`--randomize-all`), run specs in parallel across processes (`-p`), and run any subset (`--focus`/labels). The rule that keeps you honest: *declare in container, initialize in setup*. Deliberate cross-spec dependencies break randomization and parallelism — when you genuinely need ordering, say so explicitly with an `Ordered` container (→ `ginkgo:ordering-and-flakes`) rather than relying on definition order.

## The Parallelism Model

Ginkgo's parallelism model is process-based: each process runs a subset of the specs, with no shared memory.

## The node taxonomy at a glance

- **Containers** — `Describe`, `Context`, `When` (identical; pick the one that reads best). Organize the tree.
- **Subjects** — `It`, `Specify` (identical). The actual test; contains the assertions.
- **Setup** — `BeforeEach`/`AfterEach` (around every spec), `JustBeforeEach`/`JustAfterEach` (split configuration from creation), `BeforeAll`/`AfterAll` (once per `Ordered` container), `BeforeSuite`/`AfterSuite` (once per suite), `SynchronizedBeforeSuite`/`SynchronizedAfterSuite` (for common setup, then per-process setup when parallel). `DeferCleanup` registers teardown inline.
- **Reporting** — `ReportAfterEach`/`ReportAfterSuite` etc. observe results. → `ginkgo:reporting`.
- **Decorators** — modifiers attached to any node: `Serial`, `Ordered`, `Label`, `Focus`, `FlakeAttempts`, `SpecTimeout`, … → `ginkgo:decorators`.

Failures work by panic: `Fail` (which Gomega calls for you) panics; Ginkgo recovers it, records the failure, and runs cleanup. **A goroutine that might fail needs `defer GinkgoRecover()`** or its panic crashes the suite. → `ginkgo:timeouts-and-async`.

## Where to go next

- **Wiring Ginkgo into a project** (bootstrap, `RunSpecs`, the CLI, dot-import alternatives) → `ginkgo:setup`
- **Authoring specs** (containers, subjects, setup/cleanup, the construction-time pitfalls) → `ginkgo:writing-specs`
- **Table-driven and generated specs** (`DescribeTable`, shared behaviors) → `ginkgo:tables-and-dynamic-specs`
- **Looking up a decorator** → `ginkgo:decorators`
- **Running a subset** (focus, skip, labels) → `ginkgo:filtering`
- **Running in parallel / integration suites** → `ginkgo:parallelism`
- **Ordered containers and flaky-spec management** → `ginkgo:ordering-and-flakes`
- **Timeouts, interruptible nodes, and async assertions** → `ginkgo:timeouts-and-async`
- **The CLI, randomization, and running multiple suites** → `ginkgo:running`
- **A recommended CI configuration** → `ginkgo:ci`
- **Reports, profiling, and attaching data** → `ginkgo:reporting`
- **A suite failed and you want to know why** → `ginkgo:debugging-failures`

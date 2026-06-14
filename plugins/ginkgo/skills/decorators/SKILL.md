---
name: decorators
description: One-line reference for every Ginkgo decorator, grouped by what it does, with the node types each can decorate — Serial, Ordered, ContinueOnFailure, OncePerOrdered, Label, Focus, Pending, FlakeAttempts, MustPassRepeatedly, NodeTimeout, SpecTimeout, GracePeriod, PollProgressAfter/Interval, SuppressProgressReporting, SpecPriority, SemVerConstraint, AroundNode, Offset/CodeLocation, EntryDescription. Use to look up a decorator's exact name, semantics, and where it's legal.
---

# Ginkgo decorator reference

Terse lookup. Decorators are passed alongside a node's text/closure: `It("...", Serial, Label("slow"), func(){...})`. Most attach to **containers and/or subjects**; some are setup-only or interruptible-only. Decorators on a container apply to all nodes within it; the **innermost/most-specific wins** when they conflict. Full docs: <https://onsi.github.io/ginkgo/#decorator-reference>.

## Execution / ordering
- `Serial` (container, subject) — never runs in parallel; Ginkgo runs serial specs last, on process #1. → `ginkgo:parallelism`.
- `Ordered` (container only) — child specs run in definition order, never reordered; enables `BeforeAll`/`AfterAll`. → `ginkgo:ordering-and-flakes`.
- `ContinueOnFailure` (outermost `Ordered` container only) — a failing spec doesn't skip the rest of the ordered group.
- `OncePerOrdered` (setup nodes) — makes a `BeforeEach`/`AfterEach` run *once* around an `Ordered` container (like a `BeforeAll`) instead of per-spec.
- `SpecPriority(int)` (container, subject) — hint to run higher-priority specs earlier; use sparingly, not to impose a total order.

## Selection / filtering  (→ `ginkgo:filtering`)
- `Label("a", "b")` (container, subject) — tag specs for `--label-filter`; labels union down the hierarchy. No `&|!,()/"` chars.
- `Focus` (container, subject) — run only focused specs; **don't commit it** (CI exits non-zero). Constructors: `FDescribe`, `FContext`, `FIt`, `FEntry`.
- `Pending` (container, subject) — never run (and never compiled into the run set); no closure needed. Constructors: `PDescribe`, `XDescribe`, `PIt`, `XIt`, `PEntry`.
- `SemVerConstraint(">= 2.0.0", ...)` / `ComponentSemVerConstraint("redis", ">= 8.0.0")` (container, subject) — run only when `--sem-ver-filter` satisfies the constraint; children may narrow, not widen.

## Flakiness / repetition  (→ `ginkgo:ordering-and-flakes`)
- `FlakeAttempts(n)` (container, subject) — retry a failing spec up to `n` times; a later pass marks it flaky, not failed. CLI `--flake-attempts=n` overrides.
- `MustPassRepeatedly(n)` (container, subject) — the spec must pass `n` times in a row to count as passed.

## Timeouts / interruption  (interruptible nodes — those taking `ctx SpecContext`; → `ginkgo:timeouts-and-async`)
- `NodeTimeout(d)` (subject, setup) — deadline for a single interruptible node.
- `SpecTimeout(d)` (`It` only) — deadline for the whole spec (all its nodes combined).
- `GracePeriod(d)` (subject, setup) — how long Ginkgo waits for an interrupted node to return before leaking it; overrides `--grace-period`.

## Progress reporting  (→ `ginkgo:debugging-failures`)
- `PollProgressAfter(d)` (subject, setup) — emit a progress report if the node runs longer than `d`; `0` disables.
- `PollProgressInterval(d)` (subject, setup) — interval between progress reports once they start.
- `SuppressProgressReporting` (subject, setup) — silence progress reports for this node.

## Wrapping / advanced
- `AroundNode(fn)` (any node, incl. `RunSpecs`) — wrap a node's execution; forms a stack on a container applied to *every* descendant node (unlike `BeforeEach`, which runs once per spec). Forms: a bare `func()`, a context transformer `func(context.Context) context.Context`, or a full wrapper `func(ctx, func(context.Context))`.

## Location / identity  (rarely needed by hand)
- `Offset(n)` (any node) — skip `n` stack frames when computing the reported code location. Prefer `GinkgoHelper()` inside helper funcs instead.
- `CodeLocation(loc)` (any node) — supply a `types.CodeLocation` explicitly.

## Tables
- `EntryDescription("format %d")` — auto-format `Entry` descriptions from their parameters; usable table-wide or per-entry. → `ginkgo:tables-and-dynamic-specs`.
- `Entry`/`It` also accept `Label`, `Focus`, `Pending`, `FlakeAttempts`, etc. — every decorator above works on table entries.

## Reporting
- `ReportEntryVisibility{Always,FailureOrVerbose,Never}` — passed to `AddReportEntry` to control console emission. → `ginkgo:reporting`.

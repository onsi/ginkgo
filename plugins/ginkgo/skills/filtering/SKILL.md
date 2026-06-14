---
name: filtering
description: Run a subset of a Ginkgo suite — Pending/PIt/XIt, runtime Skip, programmatic Focus/FIt (and ginkgo unfocus), Label with the --label-filter query language and label sets, suite-level labels, SemVerConstraint/--sem-ver-filter, --focus/--skip and --focus-file/--skip-file, the filtering precedence rules, and --fail-on-pending/--fail-on-empty. Use when you want to run, skip, focus, label, or version-gate specs, or debug why specs were or weren't selected.
---

# Filtering specs: running a subset

Ginkgo offers many ways to run fewer specs — ad-hoc (`Pending`/`Skip`/`Focus`) and structured (labels, semver, file/description filters). They combine by a strict precedence (see bottom). Decorator specifics live in `ginkgo:decorators`; label-driven CI in `ginkgo:ci`. Docs: <https://onsi.github.io/ginkgo/#filtering-specs>.

## Pending — compile-time skip, uncoercible

Marks a spec/container as under development. **Nothing can override `Pending` and make it run** — not focus, not labels.

```go
It("needs work", Pending, func() { ... })
It("placeholder", Pending)            // pending specs need no closure
PDescribe("not ready", func() { ... }) // == Describe(..., Pending); X-prefix is identical
PIt(...) / XIt(...) / PEntry(...)
```

Pending specs don't fail the suite. `ginkgo --fail-on-pending` makes them fail CI — a policy that pending specs shouldn't be committed. **Pending is compile-time only**; you cannot make a spec pending at runtime — for that, use `Skip`.

## Skip — runtime skip

Call `Skip("reason")` from any subject or setup node to skip during the run phase. It panics to halt the spec (like `Fail`) and records the reason; it **does not fail the suite** (even skipping every spec passes).

```go
It("if it can", func() {
	if !someCondition { Skip("special condition wasn't met") }
	...
})
```

Scope matters: in a `BeforeEach` it skips the current spec; in a `BeforeAll` it skips **all** specs in the `Ordered` container; in a `BeforeSuite` it skips the **entire suite**. **You cannot call `Skip` in a container body** — it only applies during the run phase.

## Focus — programmatic, for iterating (don't commit it)

When any spec is focused, Ginkgo runs **only** focused specs.

```go
FIt("just me", func() { ... })   // or It(..., Focus, ...)
FDescribe(...) / FContext(...) / FEntry(...)
```

**Child focus unfocuses focused ancestors.** `F` an inner `It` inside an `FDescribe` and only that `It` runs — matches how you narrow while debugging.

**The danger: committing focus breaks CI quietly-loudly.** A *passing* suite with programmatic focus **exits non-zero** (logs say it succeeded but flags focused specs) so CI catches it. Strip them all with `ginkgo unfocus` (removes `F` prefixes and `Focus` decorators).

## Labels and --label-filter

`Label("a", "b")` (→ `ginkgo:decorators`) tags container/subject nodes; a spec's labels are the **union** down its hierarchy. **No `&|!,()/"` chars in a label.** Filter with `ginkgo --label-filter=QUERY`:

| Operator | Meaning |
|---|---|
| `&&` | AND |
| `||` or `,` | OR |
| `!` | NOT |
| `()` | grouping |
| `/regexp/` | regex match against labels |

Bare words match label literals — **case-insensitive**, whitespace-trimmed.

```bash
ginkgo --label-filter="network && !slow"
ginkgo --label-filter="integration, smoke"   # comma == ||
ginkgo --label-filter=/library/              # regex
```

**Consistent labelling is a best-practice for maintainable filtering**.

### Label sets — KEY:VALUE

A label `KEY:VALUE` adds `VALUE` to set `KEY` (multiple values accumulate, e.g. `Label("API:Library", "API:Geo")` → `API = {Library, Geo}`). Filter with `KEY: OP <arg>` where arg is a single value or `{V1, V2}`:

| Op | Matches when set `KEY`… |
|---|---|
| `isEmpty` | has no `KEY:*` labels |
| `containsAny` | contains any arg element |
| `containsAll` | contains all arg elements |
| `consistsOf` | equals exactly the args |
| `isSubsetOf` | is a subset of args — **note: an empty set always matches** |

```bash
ginkgo --label-filter="integration && !slow && Readiness: isSubsetOf {Beta, RC}"
ginkgo --label-filter="API: consistsOf {Library, Geo}"
```

Set values are literals (no regex). Prefer sets over regexes for structured filtering.

- **Suite-level labels**: `RunSpecs(t, "Books Suite", Label("books"))` apply to the whole suite — filter out entire suites.
- **Runtime check**: `Label("perf").MatchesLabelFilter(GinkgoLabelFilter())` lets a spec branch on the active `--label-filter` (`GinkgoLabelFilter()` returns it).
- **Discover/iterate**: `ginkgo labels` lists labels in a package; `ginkgo --dry-run -v --label-filter=Q` shows what a filter selects without running.

## SemVer filtering

`SemVerConstraint(">= 3.2.0")` (and `ComponentSemVerConstraint("redis", ">= 8.0.0")`) gate specs by version; run with `--sem-ver-filter`. Constraints follow [Masterminds/semver](https://github.com/Masterminds/semver) (`~`, `^`, comma-ANDed). Hierarchical: child constraints **narrow** (intersect), never widen. **Specs with no constraint always run** when `--sem-ver-filter` is set.

```bash
ginkgo --sem-ver-filter="2.1.1, redis=8.2.0"
```

## CLI file and description filters

- `--focus=REGEXP` / `--skip=REGEXP` — match the **full concatenated spec description** (all container texts + `It` text). Multiple `--focus` are ORed; multiple `--skip` ORed; runs specs matching focus AND not matching skip. (Labels are usually preferable.)
- `--focus-file=FILTER` / `--skip-file=FILTER` — by source location. Forms: `FILE_REGEX`, `FILE_REGEX:LINE`, `FILE_REGEX:LINE1-LINE2` (half-open). Comma-separate lines/ranges in one flag; repeat the flag for multiple files. **`LINE` must be the exact line a node (e.g. `It()`) is called on**, not "near" it.

## Precedence — how filters combine

1. **`Pending`** — never runs; uncoercible by anything.
2. **`Skip()`** — always skipped at runtime, regardless of other filters.
3. **Programmatic `Focus`** — always applies and forces a **non-zero exit**. CLI filters only narrow the focused subset.
4. **CLI filters** (`--label-filter`, `--focus-file`/`--skip-file`, `--focus`/`--skip`) are **ANDed** — a spec must satisfy all of them.

## Catching filtering mistakes

- `ginkgo --fail-on-pending` — fail if any pending specs exist.
- `ginkgo --fail-on-empty` — fail if **zero** specs ran (catches a typo'd `--label-filter` or over-aggressive skip). e.g. `ginkgo --fail-on-empty --label-filter mytypo ./...`.
- `--silence-skips` — suppress the `S` skip markers in noisy output.

# Ginkgo plugin for Claude Code

A set of [Claude Code](https://claude.com/claude-code) skills that help an agent (and you) write, run, and debug Go tests with [Ginkgo](https://onsi.github.io/ginkgo/) and [Gomega](https://onsi.github.io/gomega/) — **in your own project**.

## Install

The Ginkgo repo doubles as the marketplace:

```
/plugin marketplace add onsi/ginkgo
/plugin install ginkgo@ginkgo
```

(or non-interactively: `claude plugin marketplace add onsi/ginkgo` then `claude plugin install ginkgo@ginkgo`)

## What you get

All skills are namespaced under `ginkgo:` and activate when you're working in a Go repo with a Ginkgo suite. Start with `ginkgo:overview` — it carries the mental model and routes to the rest.

| Skill | Use it when |
|---|---|
| `ginkgo:overview` | You want the mental model — the tree-construction-then-run model and spec independence that explain every Ginkgo gotcha (read me first). |
| `ginkgo:setup` | You're wiring Ginkgo into a repo: installing the CLI, `ginkgo bootstrap`, `RegisterFailHandler(Fail)`/`RunSpecs`, dot-import alternatives, `*testing.T` interop. |
| `ginkgo:writing-specs` | You're authoring specs: containers, subjects, setup/cleanup nodes, `DeferCleanup`, and the construction-time pitfalls. |
| `ginkgo:tables-and-dynamic-specs` | You have repetitive or data-driven specs: `DescribeTable`/`Entry`, generated specs, shared behaviors. |
| `ginkgo:decorators` | You need a one-line reference for a decorator and where it's legal. |
| `ginkgo:filtering` | You want to run a subset: focus, skip, pending, labels and `--label-filter`, file/description filters. |
| `ginkgo:parallelism` | You're running in parallel or building integration suites: `-p`, `SynchronizedBeforeSuite`, sharding by process. |
| `ginkgo:ordering-and-flakes` | Specs must run in order (`Ordered`/`Serial`) or you're managing a flaky spec. |
| `ginkgo:timeouts-and-async` | A spec hangs or you're testing async behavior: interruptible nodes, timeouts, `Eventually`, goroutines. |
| `ginkgo:running` | You're running suites: the CLI, randomization, multiple suites, watching, previewing, custom config. |
| `ginkgo:ci` | You're setting up or hardening CI: the recommended flag set and the exit-code safeguards. |
| `ginkgo:reporting` | You need report files, programmatic reporters, `AddReportEntry`, or profiling. |
| `ginkgo:debugging-failures` | A suite failed and you want to know why — JSON-first diagnosis with `jq`. Also `/ginkgo:debugging-failures`. |

## Versioning

These skills track the Ginkgo library. The narrative docs at <https://onsi.github.io/ginkgo/> are the source of truth.

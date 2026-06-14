---
name: debugging-failures
description: Diagnose a failing Ginkgo suite as an agent — always run with --json-report into a predictable temp/gitignored location, read the terminal verdict line, then use jq to extract structured failure details (name, message, file:line, panic value, captured logs). Covers the panicked-vs-failed trap, panic locations pointing into the Go runtime, parallel output interleaving, reproducing with --seed, and progress reports for hangs. Use when a suite failed and you need to know why. Also invokable as /ginkgo:debugging-failures.
---

# Debugging Ginkgo failures (especially as an agent)

Ginkgo's terminal output is good for a human; for *programmatic* diagnosis, prefer a hybrid: **terminal for the verdict, JSON + `jq` for the details.** This is empirically the most reliable, lowest-token workflow. Docs: <https://onsi.github.io/ginkgo/#reporting-and-profiling-suites>.

## The workflow: one run, JSON-first diagnosis

```bash
# One invocation gives you BOTH the human verdict on the terminal AND a structured report.
# --json-report does NOT suppress console output, so you don't pay for two runs.
ginkgo -r -p --no-color --json-report=report.json --output-dir=.ginkgo-report
```

1. **Read the terminal tail** for the one-line verdict: `FAIL! -- 5 Passed | 6 Failed | 1 Pending | 1 Skipped`. Cheap, immediate.
2. **If there are failures, query the JSON with `jq`** (filters below) to get exactly the failing specs — stable, addressable, order-independent.
3. **Clean up** the report dir when done.

Why hybrid and not one or the other:
- **Terminal-only** is brittle to parse (free-form text) and, **under `-p`, interleaves failures in nondeterministic order**.
- **JSON-only** is wasteful if you dump it: the full `report.json` is ~9× the tokens of the terminal output. A `jq` failures-only extraction is a fraction of either. The discipline is *filter, don't dump*.

## Where the JSON lands (use --output-dir deliberately)

- `--json-report=report.json` alone writes `report.json` to the **current directory** — which scatters into package dirs under `-r` and risks getting committed.
- **Prefer `--output-dir=DIR`** to collect everything in one known place (auto-created). With `-r`, all suites **merge into one `DIR/report.json`** — ideal: one file, one set of `jq` filters.
- Use a **gitignored or temp** location: a repo-local `.ginkgo-report/` (add it to `.gitignore`) or an absolute temp path (`--output-dir=/tmp/ginkgo-report`). Don't leave `report.json` in the working tree.
- `--keep-separate-reports` (with `-r`) writes one `PACKAGE_report.json` per package instead of merging — only reach for it when you must attribute failures per package.

## The jq filters (tested against the real schema)

The report is an **array of suite reports**, each with a `.SpecReports[]` array.

**Verdict / counts per suite:**
```bash
jq -r '.[] | "\(.SuiteDescription): \([.SpecReports[].State]|group_by(.)|map("\(length) \(.[0])")|join(", "))"' .ginkgo-report/report.json
```

**All failures — name, location, message** (the workhorse). **Select `failed` *and* `panicked` *and* `timedout`** — see the trap below:
```bash
jq -r '.[].SpecReports[]
  | select(.State=="failed" or .State=="panicked" or .State=="timedout")
  | "[\(.State|ascii_upcase)] \((.ContainerHierarchyTexts+[.LeafNodeText])|join(" > "))\n  \(.Failure.Location.FileName):\(.Failure.Location.LineNumber)\n  \(.Failure.Message)"' .ginkgo-report/report.json
```

**Panics — the real user-code line** (the headline location lies; see below):
```bash
jq -r '.[].SpecReports[] | select(.State=="panicked")
  | "PANIC: \(.Failure.ForwardedPanic)\n\(.Failure.Location.FullStackTrace)"' .ginkgo-report/report.json
```

**Drill into one spec by name, with its captured logs:**
```bash
jq -r '.[].SpecReports[] | select(.LeafNodeText|test("SUBSTRING"))
  | "STATE: \(.State)\n\(.Failure.Message)\n--- GinkgoWriter ---\n\(.CapturedGinkgoWriterOutput // "(none)")"' .ginkgo-report/report.json
```

**Slowest specs:**
```bash
jq '.[0].SpecReports | map([(.ContainerHierarchyTexts+[.LeafNodeText])|join(" "), .RunTime/1e9]) | sort_by(.[1]) | reverse | .[0:10]' .ginkgo-report/report.json
```

## Schema facts you need

- `State`: `passed | failed | panicked | pending | skipped | timedout | interrupted | aborted`.
- Spec name = `ContainerHierarchyTexts` (array) + `LeafNodeText`.
- `Failure.Message` — the failure text (Gomega's full expected/got block for assertions).
- `Failure.Location.{FileName,LineNumber,FullStackTrace}`.
- `Failure.ForwardedPanic` — the panic value (only meaningful for `panicked`).
- `Failure.FailureNodeType` — e.g. `BeforeEach` vs `It`, so you know a *setup* node failed.
- `CapturedGinkgoWriterOutput` — `GinkgoWriter` logs (omitted if empty).
- `SpecEvents[]` — `By(...)` steps and node enter/exit (does **not** include raw GinkgoWriter text).
- `ParallelProcess` — which worker ran the spec (`-p`).

## The three traps that bite agents

1. **`panicked` is not `failed`.** `select(.State=="failed")` silently misses panics. Always include `"panicked"` (and `"timedout"`).
2. **A panic's `Location.FileName` points into the Go runtime**, not your code (e.g. `runtime_faststr.go`). The real line is in `Failure.ForwardedPanic` + `Failure.Location.FullStackTrace`.
3. **`--fail-fast` truncates the report** to whatever ran before the stop. For a full failure picture, *don't* use it — Ginkgo keeps going within a suite by default; use `--keep-going` with `-r` so one failing package doesn't stop the others.

## Reproduce, then narrow

- Every run prints its **random seed**. Reproduce the exact order with `ginkgo --seed=N`.
- Narrow to the failing spec with `--focus="<regex over the full name>"` or `--focus-file=file:line` (→ `ginkgo:filtering`). Avoid committing `FIt`.
- Re-run the one spec with `-v` (or `-vv --trace`) for the full inline **timeline** — `By` steps, `GinkgoWriter` output, and the failure point in execution order. Reserve this verbosity for a specific elusive spec; it's noise across a whole suite.

## When it's a hang, not a failure

A stuck spec emits no failure. Get a snapshot of what's running without stopping the suite: send **`SIGINFO` (`Ctrl+\` is `SIGQUIT`; macOS `Ctrl+T` = `SIGINFO`)** or **`SIGUSR1` on Linux** — Ginkgo prints the current node, goroutine stacks, and the last `GinkgoWriter` lines (a *progress report*). Make stuck specs auto-report with `--poll-progress-after=120s --poll-progress-interval=30s` (or per-node `PollProgressAfter`). Progress reports also appear in the JSON. → `ginkgo:timeouts-and-async` for making specs interrupt cleanly.

## Other formats

`--junit-report` / `--teamcity-report` / `--gojson-report` exist but lose Ginkgo metadata — use them only to feed an external CI system. For diagnosis, the native `--json-report` is richer. → `ginkgo:reporting`.

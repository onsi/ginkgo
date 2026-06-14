---
name: reporting
description: Generate, consume, and enrich Ginkgo reports — console verbosity (-v/-vv/--trace/--no-color/--succinct), machine-readable reports (--json-report/--junit-report with --output-dir/--keep-separate-reports), programmatic reporting nodes (ReportAfterEach, ReportAfterSuite, CurrentSpecReport), AddReportEntry with ReportEntryVisibility, and profiling (--cover/--race/--cpuprofile/--memprofile). Use when you need a report file, custom suite-level reporting, attaching data to a spec, controlling console output, or profiling a suite.
---

# Reporting and profiling

How to *produce* reports and feed data into them. To *read* a failing report as an agent (the `jq` failure-diagnosis workflow), see `ginkgo:debugging-failures`. Docs: <https://onsi.github.io/ginkgo/#generating-machine-readable-reports> and <https://onsi.github.io/ginkgo/#attaching-data-to-reports>.

## Console output

| Flag | Effect |
|---|---|
| `--succinct` | Minimal (default when running multiple suites). |
| (none) | Normal — detail only for *failed* specs (default, single suite). |
| `-v` | Full timeline for *every* spec; streams live in series. |
| `-vv` | `-v` plus node `> Enter`/`< Exit` events and *every* failure (not just the primary). |
| `--trace` | Full stack trace on every failure, not just panics. |
| `--no-color` / `GINKGO_NO_COLOR=TRUE` | Disable color. |
| `--github-output` | Format console output for GitHub Actions. |

**In parallel (`-p`), `-v`/`-vv` timelines can't stream** — the streams would interleave, so Ginkgo emits each spec's full timeline *after that spec completes*. → CI flags in `ginkgo:ci`.

## Machine-readable reports

```bash
ginkgo -r --json-report=report.json --output-dir=./.reports
```

- **`--json-report` is native and richest** — an array of `types.Report`, each with `SpecReports[]` of `types.SpecReport` ([godoc](https://pkg.go.dev/github.com/onsi/ginkgo/v2/types)); build tooling against the `types` package.
- `--junit-report=x.xml`, `--teamcity-report=x`, `--gojson-report=x.go.json` exist for external CI but **lose Ginkgo metadata** — use them to feed a system, not to diagnose. (JUnit maps a `Label("owner:XYZ")` to the `Owner` attribute.)
- `--output-dir=DIR` collects all files in one place; under `-r` all suites **merge into one report file**. `--keep-separate-reports` writes one `PACKAGE_report.json` per package instead.
- **Every machine-readable report embeds the full `-vv` timeline regardless of console verbosity.** Run the suite quiet and mine the file — you don't need `-vv` on the console to get rich detail in the report.

## Programmatic reporting nodes

`CurrentSpecReport()` works in any run-phase closure — use it for in-spec conditional diagnostics:

```go
AfterEach(func() {
  if CurrentSpecReport().Failed() {
    GinkgoWriter.Println(libraryClient.DebugLogs())   // expensive diagnostics, only on failure
  }
})
```

For *custom report files*, don't roll your own in `AfterEach` — use reporting nodes:

- **`ReportAfterEach(func(r SpecReport){...})` / `ReportBeforeEach`** — run for **every** spec including skipped/pending (unlike `AfterEach`), after all `AfterEach`es. A failure here fails the spec. **In parallel they run on each process concurrently — never write a shared file from them** (you'll get interleaved garbage).
- **`ReportBeforeSuite` / `ReportAfterSuite("name", func(r Report){...})`** — top-level, and **run only on process #1**. `ReportAfterSuite` receives the `Report` **aggregated across all parallel processes** — this is the safe place for custom suite-level report files:

```go
ReportAfterSuite("custom report", func(report Report) {
  f, _ := os.Create("report.custom")
  for _, sr := range report.SpecReports {
    fmt.Fprintf(f, "%s | %s\n", sr.FullText(), sr.State)
  }
  f.Close()
})
```

Reporting nodes can be made interruptible by also accepting `SpecContext` → `ginkgo:timeouts-and-async`.

## Attaching data to a spec: AddReportEntry

```go
It("is reported", func() {
  AddReportEntry("Publish stats", stats, ReportEntryVisibilityFailureOrVerbose)
})
```

`AddReportEntry(name, value, ...args)` attaches arbitrary data to the current spec; it lands in `SpecReport.ReportEntries`, in all machine-readable reports, and (per visibility) on the console. Visibility enum:

- `ReportEntryVisibilityAlways` (default) — always printed.
- `ReportEntryVisibilityFailureOrVerbose` — only on failure or `-v` (like `GinkgoWriter`).
- `ReportEntryVisibilityNever` — never on console, but **still in the JSON report**.

`value` is JSON-round-tripped into the report (retrieve via `entry.GetRawValue()` / `json.Unmarshal(entry.Value.AsJSON, ...)`). A `fmt.Stringer`/`types.ColorableStringer` controls console rendering; pass a *pointer* to capture later mutations. `AddReportEntry` decorators (`Offset`, `CodeLocation`, `ReportEntryVisibility`) → `ginkgo:decorators`.

**Recap (owned by `ginkgo:writing-specs`):** `GinkgoWriter` buffers log output and only prints it on failure (or always under `-v`); `By("step")` annotates the timeline. Use these for ad-hoc logging; use `AddReportEntry` for structured data you want in the report file.

## Profiling

```bash
ginkgo -r --cover --coverpkg=./... --race --cpuprofile=cpu.out --output-dir=./.profiles
```

- Coverage: `--cover`, `--coverprofile=NAME` (filename only — use `--output-dir` for a path), `--coverpkg=./...` to cover beyond the suite's own package, `--keep-separate-coverprofiles` per-package. `--race` forces `--covermode=atomic`.
- Profiles: `--cpuprofile` / `--memprofile` / `--blockprofile` / `--mutexprofile` (Ginkgo keeps the test binary so you can `go tool pprof BINARY PROFILE`). `--output-dir` collects and package-namespaces all profile/coverage assets.
- **Focused specs break profiling.** A suite with `Focus`/`FIt`/`FDescribe` exits early with a non-zero code, which stops `go test` from writing the profile. **Subset with filters (`--focus=`, `--label-filter=`), not `Focus`** → `ginkgo:filtering`.

## See also

- Reading a failing report (`jq` filters, the panicked-vs-failed trap) → `ginkgo:debugging-failures`
- The full CI flag set and version-matching → `ginkgo:ci`
- Decorators (`Label`, `ReportEntryVisibility`, `Offset`) → `ginkgo:decorators`

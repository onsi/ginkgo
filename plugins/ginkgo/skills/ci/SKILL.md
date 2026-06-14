---
name: ci
description: Configure Ginkgo for continuous integration — the recommended CLI flag set and the rationale for each flag (-r -p --randomize-all --randomize-suites --fail-on-pending --fail-on-empty --keep-going --cover --race --trace --json-report --timeout --poll-progress-after/-interval), invoking via go run to pin the CLI to go.mod, the exit-code safeguards that catch committed Focus/Pending and empty filters, collecting report and coverage artifacts with --output-dir, and CI-friendly output (--github-output/--force-newlines/--no-color). Use when setting up or hardening a CI pipeline for a Ginkgo suite.
---

# Ginkgo in CI

A CI invocation should maximize signal — surface flakes and spec pollution, catch mistakes that pass locally, and emit machine-readable artifacts — while collecting *every* failure in one run. This builds on the CLI (`ginkgo:running`) and report formats (`ginkgo:reporting`). Full rationale: <https://onsi.github.io/ginkgo/#recommended-continuous-integration-configuration>.

## The recommended flag set

Invoke via `go run` so the CLI version always tracks the `github.com/onsi/ginkgo/v2` in your `go.mod` — no separately-installed binary to drift (→ `ginkgo:setup`).

```bash
go run github.com/onsi/ginkgo/v2/ginkgo \
  -r -p --randomize-all --randomize-suites \
  --fail-on-pending --fail-on-empty --keep-going \
  --cover --coverprofile=cover.profile --race --trace \
  --json-report=report.json --output-dir=.ginkgo-report \
  --timeout=TIMEOUT --poll-progress-after=Xs --poll-progress-interval=Ys
```

| Flag | Why |
|---|---|
| `-r` | recursively find and run every suite |
| `-p` | run each suite in parallel (→ `ginkgo:parallelism`; set `--procs=N`/`--compilers=N` if CPU detection is wrong) |
| `--randomize-all` / `--randomize-suites` | shuffle all specs and the suite order to surface spec pollution (→ `ginkgo:running`) |
| `--fail-on-pending` | fail if any `Pending` specs were committed |
| `--fail-on-empty` | fail if no specs ran (usually a malformed filter) |
| `--keep-going` | don't stop at the first failed suite — collect all failures |
| `--cover --coverprofile=cover.profile` | compute coverage into one merged profile (→ `ginkgo:reporting`) |
| `--race` | run with the race detector |
| `--trace` | full stack traces on failure (worth it without a local feedback loop) |
| `--json-report=report.json` | structured results for diagnosis and downstream tools (→ `ginkgo:debugging-failures`) |
| `--timeout=TIMEOUT` | cap the whole run (default 1h — often not enough) |
| `--poll-progress-after`/`--poll-progress-interval` | emit progress reports for stuck specs (→ `ginkgo:timeouts-and-async`) |

## The exit-code safeguards — CI's real value

Three flags turn "looks green" into "is actually trustworthy." They make CI fail on mistakes that pass silently on a developer's machine:

- **`--fail-on-pending`** — a committed `Pending`/`PIt`/`XIt` becomes a CI failure, so dev-time placeholders can't rot into the suite.
- **`--fail-on-empty`** — a typo'd `--label-filter` (or an over-aggressive `--skip`) that selects *zero* specs fails instead of falsely passing.
- **Committed programmatic focus already fails CI for free.** A *passing* suite that contains `FIt`/`FDescribe`/`Focus` **exits non-zero** — Ginkgo's built-in guard against shipping focus. Don't suppress it; fix it with `ginkgo unfocus` (→ `ginkgo:filtering`).

## Collecting artifacts

Point reports and profiles at one directory and upload it as a build artifact:

- `--output-dir=.ginkgo-report` collects the JSON report (and, under `-r`, merges all suites into one `report.json`) plus coverage/profiles into one place. → `ginkgo:reporting` for `--keep-separate-reports` and the other formats.
- `--junit-report=junit.xml` *additionally* if your CI system renders JUnit — but keep `--json-report` as the source of truth; JUnit loses Ginkgo metadata.
- An agent or human then diagnoses failures straight from `report.json` with `jq` → `ginkgo:debugging-failures`.

## CI-friendly console output

- **`--github-output`** — formats the console log for GitHub Actions readability (grouped, annotated).
- **`--force-newlines`** — flush output line-by-line for CI systems that only flush on newline.
- **`--no-color`** (or `GINKGO_NO_COLOR=TRUE`) — drop ANSI codes from logs that don't render them.

## Flakes and timeouts in CI

- **Collect, don't bail.** `--keep-going` (across suites) plus Ginkgo's default keep-going *within* a suite gives you the full failure picture in one run. Avoid `--fail-fast` in CI — it truncates the report (→ `ginkgo:debugging-failures`).
- **Don't paper over flakes.** `--flake-attempts=N` exists, but treat it as explicit, temporary tech debt rather than a standing CI flag — fix the root cause (→ `ginkgo:ordering-and-flakes`).
- **Set `--timeout` deliberately** (default 1h) and add `--poll-progress-after`/`--poll-progress-interval` so a stuck spec emits a progress report (current node + goroutine stacks) instead of silently eating the budget. For long suites, `120s`/`30s` are reasonable; skip them for fast unit suites. If you precompile and run elsewhere, set `--source-root` so progress reports can show source.

## See also

- The CLI surface and randomization details → `ginkgo:running`
- Report formats, programmatic reporters, and profiling → `ginkgo:reporting`
- Reading the JSON report after a failed CI run → `ginkgo:debugging-failures`

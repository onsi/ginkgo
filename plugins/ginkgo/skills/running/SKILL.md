---
name: running
description: Run Ginkgo suites with the ginkgo CLI — run, -r, -p, --dry-run, watch, build (precompiled .test binaries), generate, outline, unfocus, labels, version; spec randomization (--randomize-all/--randomize-suites/--seed); running multiple suites (--keep-going/--skip-package/--compilers); previewing (--dry-run, PreviewSpecs); and parameterizing a suite via env vars or init()-registered flags after -- plus GinkgoConfiguration() overrides. Use when running suites locally, precompiling, watching for changes, or parameterizing a run from the command line. For a CI configuration see ginkgo:ci.
---

# Running Ginkgo

How to drive the suite once it's wired up (`ginkgo:setup`). For a production CI configuration, see `ginkgo:ci`. Docs: <https://onsi.github.io/ginkgo/#ginkgo-cli-overview>.

## The CLI surface

`ginkgo` is the recommended runner — **you can run suites with `go test`, but only the CLI does parallelism and profile aggregation**, and a few `go test` flags (e.g. `-count=N`, use `--repeat=N`) are unsupported. `ginkgo` is shorthand for `ginkgo run`; it passes through nearly all `go test`/`go build` flags.

```bash
ginkgo                                  # run the suite in the current directory
ginkgo path/to/suite path/to/other      # run several named suites
ginkgo -r                               # recurse the tree, run every suite found (== ginkgo ./...)
ginkgo -p                               # run in parallel (→ ginkgo:parallelism)
ginkgo -v --dry-run                     # build the tree and print what WOULD run, without running
ginkgo <GINKGO-FLAGS> <PACKAGES> -- <PASS-THROUGHS>   # the full grammar
```

**Ginkgo flags must come before the package list; suite-bound args go after `--`.** Subcommands:

| Command | Does |
|---|---|
| `ginkgo build path/...` | Precompile each suite to a `package.test` binary (accepts a subset of `run` flags — compile-time only) |
| `ginkgo watch -r -p` | Re-run suites when they or their dependencies change (`-depth` tunes dependency monitoring) |
| `ginkgo bootstrap` / `ginkgo generate <subject>` | Scaffold a suite / a spec file (→ `ginkgo:setup`); both take `--template`/`--template-data` |
| `ginkgo outline <file>` | Static (AST-parsed) spec outline as `csv`/`indent`/`json` — no compile needed |
| `ginkgo unfocus` | Strip all programmatic `F`-focus from the tree (→ `ginkgo:filtering`) |
| `ginkgo labels` | List the `Label`s used in a suite (naive parse) |
| `ginkgo version` | Print the CLI version |

**Precompiled binaries** run via `ginkgo package.test` or `./package.test` — but **to run a precompiled suite in parallel you must go back through the CLI**: `ginkgo -p ./package.test`. Cross-compile with the usual `GOOS`/`GOARCH`.

## Randomization

Ginkgo randomizes spec order to surface spec pollution. **By default it only shuffles top-level containers** (specs within a container keep file order, easing debugging).

- `--randomize-all` — shuffle *every* spec. Use it in CI (→ `ginkgo:ci`).
- `--randomize-suites` — shuffle the *order suites run* in (multi-suite runs).
- `--seed=N` — pin the seed (printed near the top of every run) to **reproduce a failing order exactly**. Specs that generate randomness should seed off `GinkgoRandomSeed()` so the seed fully determines the run.

Randomization is only sound because **Ginkgo assumes specs are independent** (→ `ginkgo:overview`); when ordering is genuinely required, use an `Ordered` container (→ `ginkgo:ordering-and-flakes`), never definition order.

## Running multiple suites

With `-r` or multiple package args, Ginkgo **compiles suites in parallel but always runs them sequentially**.

- `--compilers=N` — number of parallel compile workers (default: one per core).
- `--skip-package=a,b,c` — comma-separated; skips any suite whose **path** contains a listed substring (not compiled, not run).
- `--keep-going` — run *all* suites even after one fails (default stops after the first failed suite), to collect every failure. (`--fail-fast` controls stopping *within* a suite.)

You can also run several suites **within one process** by calling `RunSpecs` more than once, resetting Ginkgo's global state between calls with `extensions/globals.Reset()` (a niche orchestration escape hatch — your package-level state is *not* reset, only Ginkgo's).

## Previewing specs

- **`ginkgo --dry-run -v`** compiles, builds the tree, and walks it honoring filters and dynamically generated specs — the most complete preview. **`--dry-run` can't combine with `-p`/`--procs`; it runs in series.** (`ginkgo outline` is faster but static, so it misses dynamic specs and filter/order effects.)
- **`PreviewSpecs`** in place of `RunSpecs` returns a full `Report` (run specs → `SpecStatePassed`, skipped → `SpecStateSkipped`) for programmatic inspection. Pass it the *same* description+decorators you'd give `RunSpecs`:
  ```go
  func TestMySuite(t *testing.T) {
  	if config, _ := GinkgoConfiguration(); config.DryRun {
  		report := PreviewSpecs("My Suite", Label("suite-label"))
  		_ = report // e.g. reporters.GenerateJUnitReport(report, "./preview.xml")
  	} else {
  		RunSpecs(t, "My Suite", Label("suite-label"))
  	}
  }
  ```

## Custom suite configuration

Two ways to parameterize a suite from the command line.

**Environment variables** — simplest, readable anywhere (Run *and* Tree-Construction phases):

```bash
SMOKETEST_SERVER_ADDR="127.0.0.1:3000" SMOKETEST_ENV="STAGING" ginkgo
```

**Custom flags** — register in `init()` (so they exist before `go test` calls `flag.Parse()`), then pass them **after `--`**:

```go
var serverAddr string
func init() { flag.StringVar(&serverAddr, "server-addr", "", "server to smoke-check") }
```
```bash
ginkgo -p -- --server-addr="127.0.0.1:3000" --environment="STAGING"
```

**Gotcha: parsed flags are available in `BeforeSuite`, in setup/subject closures, and in container bodies (Tree-Construction) — but NOT at the suite's top level.** A top-level `var name = fmt.Sprintf("...%s", smokeEnv)` runs while `init()`s are merely being *defined*, before flags are parsed, so it sees the zero value. If you must read config at the top level, use an env var instead.

To translate user-facing config into Ginkgo's own filtering, **grab and mutate the config before `RunSpecs`** via `GinkgoConfiguration()`:

```go
suiteConfig, _ := GinkgoConfiguration()
suiteConfig.LabelFilter = smokeEnv            // e.g. drive label filtering from --environment
RunSpecs(t, "Smoketest Suite", suiteConfig)
```

This keeps a clean user contract while leaning on Ginkgo's label filtering (→ `ginkgo:filtering`) instead of `if`-guarding specs out of the tree.

## Where to go next

- A hardened CI invocation (the recommended flag set + rationale) → `ginkgo:ci`
- Report formats, profiling, and attaching data → `ginkgo:reporting`
- Diagnosing a failing run → `ginkgo:debugging-failures`

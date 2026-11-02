---
name: parallelism
description: Run Ginkgo suites in parallel — ginkgo -p / --procs, the separate-process (not goroutine) model, SynchronizedBeforeSuite/SynchronizedAfterSuite vs BeforeSuite, GinkgoParallelProcess() for sharding ports/tmpdirs/databases, building a binary once via gexec, piping child-process output to GinkgoWriter, and what N processes actually cost (compilation vs teardown, and the fixed ~1s --race adds per suite — GORACE=atexit_sleep_ms=0). Use when parallelizing a suite, speeding up integration tests, auditing why a parallel run is slow, fixing parallel-only flakes/races, sharding external resources, or choosing between BeforeSuite and SynchronizedBeforeSuite.
---

# Running Ginkgo in parallel

Parallelism is where Ginkgo's "every spec is independent" assumption (`ginkgo:overview`) earns its keep. If your specs are independent, going parallel is one flag. If they aren't, parallelism is how you find out. Docs: <https://onsi.github.io/ginkgo/#spec-parallelization>, <https://onsi.github.io/ginkgo/#patterns-for-parallel-integration-specs>.

## The flag

```bash
ginkgo -p            # auto-detect optimal process count from CPU cores
ginkgo --procs=4     # pin the count explicitly
ginkgo watch -p      # re-run in parallel on every file change
```

**`-p`/`--procs` require the `ginkgo` CLI — `go test` cannot parallelize a Ginkgo suite** (it has no server to coordinate processes). Most other features work under `go test`; parallelism is the headline exception. → `ginkgo:running`.

## The mental model: separate processes, not goroutines

This is the one thing to internalize. Ginkgo compiles the suite once (`go test -c`), then launches **N copies of the test binary as separate OS processes**. Each process runs the full tree-construction phase independently (so all N build the identical spec list), then pulls specs to run from the CLI, which acts as a server and merges everything into one coherent output stream.

- **Separate processes means separate memory.** Each process has its own copy of every package-level var and closure variable. There is **no shared memory**, so shared-closure specs like `var book` don't race across processes — they each get their own `book`. (Within a process specs still run one at a time.)
- The flip side: **a `BeforeSuite` runs on every process**, so anything it creates is created N times.
- You usually don't care which process runs a spec — except for the integration patterns below, which deliberately shard shared external resources by process index.

## What the N processes cost

Going parallel is cheap, but not free — and the cost is rarely where people look for it.

- **Compilation is once per *suite*, not once per process.** Ginkgo builds the test binary a single time and runs N copies of it; under `-r` the CLI also overlaps compiling the next suite with running the current one. It is usually not the bottleneck it's assumed to be.
- **Teardown is once per process — but they overlap**, so a suite pays roughly *one* process teardown, not N.
- **Under `--race` that one teardown is ~1s.** ThreadSanitizer sleeps for a second at process exit (`atexit_sleep_ms`, its default) so races reported by goroutines still running at exit aren't lost. Every process pays it however trivial the suite: a 2-spec package pays exactly what a 2000-spec package pays.

**That flatness is the tell.** The sleep lands in the gap *between* suites under `-r`, where it looks exactly like compilation — but a build cost can't be flat across package size, and a per-process cost must be. If you're auditing a slow `-p --race` run, precompile with `ginkgo build -r` so compilation is excluded by construction rather than by inference. To remove the sleep:

```bash
GORACE=atexit_sleep_ms=0 ginkgo -r -p --race
```

This slightly weakens detection at the very *end* of a suite — a reasonable trade when your races are found by your specs rather than at teardown, but a deliberate one; write the reason down where you set it. → `ginkgo:ci` for the whole CI flag set.

## Suite setup when N copies is wrong: Synchronized*Suite

`BeforeSuite` runs on **every** process → N independent resources. Sometimes that's ideal (max isolation). When a resource is expensive or must be shared, use `SynchronizedBeforeSuite`: set it up **once** on process #1, then distribute connection info to **all** processes.

```go
var dbClient *db.Client

var _ = SynchronizedBeforeSuite(func() []byte {
	// runs ONLY on process #1; all others wait
	dbRunner := db.NewRunner()
	Expect(dbRunner.Start()).To(Succeed())
	DeferCleanup(dbRunner.Stop)              // runs like SynchronizedAfterSuite: after ALL procs finish
	return []byte(dbRunner.Address())        // []byte passed to every process
}, func(address []byte) {
	// runs on ALL processes with the data from process #1
	dbClient = db.NewClient()
	Expect(dbClient.Connect(string(address))).To(Succeed())
	dbClient.SetNamespace(fmt.Sprintf("namespace-%d", GinkgoParallelProcess()))
	DeferOnFailure(dbClient.DumpNamespace)   // suite-wide failure diagnostics belong here (every process), not in the proc-#1 func
})
```

| | `BeforeSuite` | `SynchronizedBeforeSuite` |
|---|---|---|
| process #1 | runs | first func runs, returns `[]byte` |
| all processes | runs (N independent resources) | second func runs with that `[]byte` |
| use for | per-process resources | one shared resource, info fanned out |

`SynchronizedAfterSuite(allProcesses, process1)` mirrors it: the first func runs on every process as it finishes; the second runs **only on process #1, after every other process has finished its suite** (Ginkgo waits for them to report back, not for their processes to exit). The `[]byte` return is optional — a `func()`/`func()` form exists too.

## Sharding shared resources by process index

`GinkgoParallelProcess()` returns this process's index (`1..N`); `GinkgoConfiguration()` exposes `ParallelTotal`. Use them to carve up any singleton resource so processes don't collide — the "declare in container, initialize in setup" principle extended to *external* state.

```go
// ports
addr := fmt.Sprintf("127.0.0.1:%d", 50000+GinkgoParallelProcess())

// tmp dirs
dir := fmt.Sprintf("./tmp-%d", GinkgoParallelProcess())
os.MkdirAll(dir, 0755); DeferCleanup(os.RemoveAll, dir)

// db / kv / multi-tenant namespaces
client.SetNamespace(fmt.Sprintf("test-%d", GinkgoParallelProcess()))
```

**Specs that pass in series but fail mysteriously under `-p` are almost always colliding on a shared singleton** (a fixed port, a hard-coded `out.epub`, one DB key). Shard it by process index. `GinkgoT().TempDir()` is a zero-config alternative for files (auto-cleaned, but lands in a random location).

## Integration patterns

- **Build a binary once, share the path.** Compiling in `BeforeSuite` recompiles N times. Compile in `SynchronizedBeforeSuite` on process #1 via `gexec.Build`, return the path, and let every process launch its own instance:
  ```go
  var publisherPath string
  var _ = SynchronizedBeforeSuite(func() []byte {
  	path, err := gexec.Build("path/to/publisher")
  	Expect(err).NotTo(HaveOccurred())
  	DeferCleanup(gexec.CleanupBuildArtifacts)
  	return []byte(path)
  }, func(path []byte) { publisherPath = string(path) })
  ```
- **Databases — pick a strategy by cost:** a fresh DB per spec (`BeforeEach`) is bulletproof but slow; a DB per process spun up in `BeforeSuite` with snapshot/restore between specs is the common sweet spot; a single shared singleton sharded by `GinkgoParallelProcess()` namespace works when you can't spin up your own.
- **Pipe child-process stdout/stderr to `GinkgoWriter`, never `os.Stdout`/`os.Stderr`.** If a process you spawn outlives the spec and its output is wired to `os.Stdout`, **it hangs Ginkgo's output interceptor**. `gexec.Start(cmd, GinkgoWriter, GinkgoWriter)` is the safe default. → `ginkgo:writing-specs` (GinkgoWriter).

## Related

- A spec that genuinely can't run in parallel → mark it `Serial` (runs last on process #1). → `ginkgo:decorators`, `ginkgo:ordering-and-flakes`.
- Specs that must run in a fixed order → `Ordered` containers, not definition order. → `ginkgo:ordering-and-flakes`.
- Launching processes and asserting on their async output (`Eventually`, `gexec.Exit`) → `ginkgo:timeouts-and-async`.
- Surfacing the spec that failed only under parallelism → `ginkgo:debugging-failures`.

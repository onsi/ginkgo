---
name: timeouts-and-async
description: Make Ginkgo specs interruptible and test asynchronous behavior — SpecContext/context.Context cancellable nodes, NodeTimeout/SpecTimeout/GracePeriod, the --timeout flag, Abort and SIGINT behavior, Gomega Eventually/Consistently (the func(g Gomega) form, .WithContext), and the defer GinkgoRecover() rule for goroutines. Use when a spec hangs or times out, polls for eventual consistency, tests channels/streams/processes, launches goroutines, or needs a deadline.
---

# Timeouts, interruptible nodes, and async testing

Builds on the failure mental model in `ginkgo:overview` (failure is a panic Ginkgo recovers; a goroutine that fails needs `defer GinkgoRecover()`). Docs: <https://onsi.github.io/ginkgo/#spec-timeouts-and-interruptible-nodes> and <https://onsi.github.io/ginkgo/#patterns-for-asynchronous-testing>.

## Interruptible nodes: take a context, honor cancellation

A setup or subject node becomes **interruptible** simply by accepting a `SpecContext` (or plain `context.Context`) — Ginkgo supplies one automatically:

```go
It("likes to sleep in", func(ctx context.Context) {
  select {
  case <-ctx.Done(): return        // honor cancellation and exit promptly
  case <-time.After(time.Hour): ...
  }
}, NodeTimeout(time.Second))
```

On timeout or interrupt, Ginkgo **cancels the context** to tell the node to stop. Pass `ctx` down into every blocking call (`libraryClient.SaveBook(ctx, book)`, `exec.CommandContext(ctx, ...)`) so the cancellation actually propagates. Only setup/subject nodes are interruptible — **container nodes are not** (they run at construction time).

- **`ctx.Deadline()` does NOT report the node's deadline.** Ginkgo manages cancellation timing itself (to snapshot a progress report first), so it does not use a `WithDeadline` context. Trust `<-ctx.Done()`, not `Deadline()`.
- `SpecContext` satisfies `context.Context`; you may wrap it (`context.WithValue`) and Ginkgo still cancels the result on time.

## The timeout decorators (full reference → `ginkgo:decorators`)

| Decorator | Scope |
|---|---|
| `NodeTimeout(d)` | Deadline for **one** interruptible node. |
| `SpecTimeout(d)` | Deadline for the whole spec lifecycle — **`It` only**. |
| `GracePeriod(d)` | How long Ginkgo waits after cancelling before leaking the node. |

- **`SpecTimeout` can only be more lenient than a node's `NodeTimeout`** — it caps the sum of all nodes (`BeforeEach`+`It`+`AfterEach`); a per-node `NodeTimeout` inside it is always more stringent.
- **Cleanup still runs after a timeout.** When `SpecTimeout` fires, Ginkgo cancels the current node, then runs `AfterEach`/`AfterAll`/`DeferCleanup` (each under its own `NodeTimeout`/grace). The timeout is a "mark failed" threshold, not a hard kill.
- **A node that ignores cancellation is "leaked," not killed.** After the grace period Ginkgo gives up waiting and moves on. Leaking is deliberately preferred over hanging forever — **but a leaked goroutine keeps running and can call `Fail`/`AddReportEntry` and pollute a *later* spec.** Always make blocking code respond to `ctx.Done()`.
- `DeferCleanup`/`DescribeTable` entries are interruptible too — give the cleanup/entry func a `ctx` first arg; don't capture and reuse the *parent* node's `ctx` (it's already cancelled by cleanup time).

## Interrupting the whole suite

- **`ginkgo --timeout=DURATION`** — suite-wide budget across *all* suites (default `1h`).
- **`Abort("reason")`** — end the suite immediately from within a spec (programmatic interrupt).
- **`SIGINT`/`SIGTERM` (`^C`)** — interrupt: cancel the current interruptible node, run its cleanup + reporting nodes, skip the rest, exit failed. **Escalation: a 2nd interrupt skips cleanup (still runs reporting); a 3rd bails immediately.** To inspect a suite *without* stopping it, send `SIGINFO`/`SIGUSR1` for a progress report → `ginkgo:debugging-failures`.

## Async assertions: Eventually / Consistently

Use Gomega's `Eventually` (polls until the matcher passes or it times out) and `Consistently` (polls and requires the matcher to hold the whole interval — the way to assert something *doesn't* happen). Three input shapes: bare values (channels, `gbytes`), functions returning `(value[, error])`, and functions taking a `Gomega`.

```go
It("publishes a book", func(ctx SpecContext) {
  buffer := gbytes.NewBuffer()
  c := publisher.Publish(ctx, book, buffer)              // pass ctx so it cancels cleanly
  Eventually(ctx, buffer).Should(gbytes.Say(`Publish complete!`))
  var result publisher.PublishResult
  Eventually(ctx, c).WithTimeout(time.Second).Should(Receive(&result)) // poll, don't <-c
}, SpecTimeout(time.Second*30))
```

**Propagate the spec deadline into the poll** with `.WithContext(ctx)` or the positional `Eventually(ctx, ...)` — now a node timeout/interrupt makes the `Eventually` exit immediately instead of running its own clock. Gomega also **auto-injects** the context and `.WithArguments(...)` into a polled function whose first params are `(ctx)`, so `Eventually(client.Connect).WithContext(ctx).Should(Succeed())` works (pass the method *reference*, not `client.Connect()`).

### Assertions inside the polled function — use `g`, never global `Expect`

Pass a function taking `func(g Gomega, ...)` and assert with **`g.Expect`** so a failed poll *retries* instead of failing the spec outright:

```go
Eventually(func(g Gomega, ctx SpecContext) {   // g Gomega must be first
  messages, err := gmail.Fetch(ctx, jane.EmailAddress)
  g.Expect(err).NotTo(HaveOccurred())
  g.Expect(messages).To(ContainElement(WithTransform(subjectOf, Equal(want))))
}).WithContext(ctx).Should(Succeed())          // Succeed() = "no failures in the func"
```

**Using the global `Expect` inside an `Eventually` defeats the retry** — the first failure calls `Ginkgo`'s `Fail` and the spec dies with no second attempt. The local `g` lets `Eventually` catch and re-poll.

## Goroutines: the two rules that bite

```go
It("repaginates", func() {
  done := make(chan any)
  go func() {
    defer GinkgoRecover()                 // RULE 1
    Expect(book.SetFontSize(28)).To(Succeed())
    close(done)
  }()
  Eventually(done).Should(BeClosed())     // RULE 2: poll, don't block on <-done
})
```

1. **Any goroutine that may call `Fail` or a Gomega assertion needs `defer GinkgoRecover()`.** Ginkgo can't recover a panic raised on a goroutine it didn't start — without this, one failed assertion crashes the *entire suite*. (Ginkgo's crash message reminds you.)
2. **Don't block the spec on a channel fed by a failing goroutine.** If the goroutine fails before `close(done)`, a bare `<-done` blocks until the node *times out* instead of reporting the real failure. `Eventually(done).Should(BeClosed())` lets the failure surface immediately.

## See also

- Decorator reference (`NodeTimeout`, `SpecTimeout`, `GracePeriod`, `PollProgressAfter`) → `ginkgo:decorators`
- Progress reports for a hanging spec, and diagnosing a timeout from the JSON report → `ginkgo:debugging-failures`
- `--timeout` and friends in a CI config → `ginkgo:ci`

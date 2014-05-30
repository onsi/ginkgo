## Head
New Features:

- Test suites with a programmatic focus (`FIt`, `FDescribe`, etc...) exit with non-zero status code, evne when they pass.  This allows CI systems to detect accidental commits of focused test suites.
- `ginkgo -p` runs the testsuite in parallel with an auto-detected number of nodes.
- `ginkgo -tags=TAG_LIST` passes a list of tags down to the `go build` command.

Bug Fixes:

- `ginkgo boostrap` and `ginkgo generate` no longer fail when dealing with `hyphen-separated-packages`.

## 1.0.0 (5/24/214)
New Features:

- Add `GinkgoParallelNode()` - shorthand for `config.GinkgoConfig.ParallelNode`

Improvements:

- When compilation fails, the compilation output is rewritten to present a correct *relative* path.  Allows ⌘-clicking in iTerm open the file in your text editor.
- `--untilItFails` and `ginkgo watch` now generate new random seeds between test runs, unless a particular random seed is specified.

Bug Fixes:

- `-cover` now generates a correctly combined coverprofile when running with in parallel with multiple `-node`s.
- Print out the contents of the `GinkgoWriter` when `BeforeSuite` or `AfterSuite` fail.
- Fix all remaining race conditions in Ginkgo's test suite.

## 1.0.0-beta (4/14/2014)
Breaking changes:

- `thirdparty/gomocktestreporter` is gone.  Use `GinkgoT()` instead
- Modified the Reporter interface 
- `watch` is now a subcommand, not a flag.

DSL changes:

- `BeforeSuite` and `AfterSuite` for setting up and tearing down test suites.
- `AfterSuite` is triggered on interrupt (`^C`) as well as exit.
- `SynchronizedBeforeSuite` and `SynchronizedAfterSuite` for setting up and tearing down singleton resources across parallel nodes.

CLI changes:

- `watch` is now a subcommand, not a flag
- `--nodot` flag can be passed to `ginkgo generate` and `ginkgo bootstrap` to avoid dot imports.  This explicitly imports all exported identifiers in Ginkgo and Gomega.  Refreshing this list can be done by running `ginkgo nodot`
- Additional arguments can be passed to specs.  Pass them after the `--` separator
- `--skipPackage` flag takes a regexp and ignores any packages with package names passing said regexp.
- `--trace` flag prints out full stack traces when errors occur, not just the line at which the error occurs.

Misc:

- Start using semantic versioning
- Start maintaining changelog

Major refactor:

- Pull out Ginkgo's internal to `internal`
- Rename `example` everywhere to `spec`
- Much more!

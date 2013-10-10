![Ginkgo: A Golang BDD Testing Framework](http://onsi.github.io/ginkgo/images/ginkgo.png)

[![Build Status](https://travis-ci.org/onsi/ginkgo.png)](https://travis-ci.org/onsi/ginkgo)


Jump straight to the [docs](http://onsi.github.io/ginkgo/) to learn more.

## Feature List

- Ginkgo uses Go's `testing` package and can live alongside your existing `testing` tests.  It's easy to [bootstrap](http://onsi.github.io/ginkgo/#bootstrapping_a_suite) and start writing your [first tests](http://onsi.github.io/ginkgo/#adding_specs_to_a_suite)

- Structure your BDD-style tests expressively:
    - Nestable [`Describe` and `Context` container blocks](http://onsi.github.io/ginkgo/#organizing_specs_with_containers__and_)
    - [`BeforeEach` and `AfterEach` blocks](http://onsi.github.io/ginkgo/#extracting_common_setup_) for setup and teardown
    - [`It` blocks](http://onsi.github.io/ginkgo/#individual_specs_) that hold your assertions
    - [`JustBeforeEach` blocks](http://onsi.github.io/ginkgo/#separating_creation_and_configuration_) that separate creation from configuration (also known as the subject action pattern).

- A comprehensive test runner that lets you:
    - Mark specs as [pending](http://onsi.github.io/ginkgo/#pending_specs)
    - [Focus](http://onsi.github.io/ginkgo/#focused_specs) individual specs, and groups of specs, either programmatically or on the command line
    - Run your tests in [random order](http://onsi.github.io/ginkgo/#spec_permutation), and then reuse random seeds to replicate the same order.
    - Break up your test suite into parallel processes for straightforward [test parallelization](http://onsi.github.io/ginkgo/#parallel_specs)

- Built-in support for testing [asynchronicity](http://onsi.github.io/ginkgo/#asynchronous_tests)

- Built-in support for [benchmarking](http://onsi.github.io/ginkgo/#benchmark_tests) your code.  Control the number of benchmark samples as you gather runtimes and other, arbitrary, bits of numerical information about your code. 

- `ginkgo`: a command line interface with plenty of handy command line arguments for [running your tests](http://onsi.github.io/ginkgo/#running_tests) and [generating](http://onsi.github.io/ginkgo/#generators) test files.

    The `ginkgo` CLI is convenient, but purely optional -- Ginkgo works just fine with `go test`

- A modular architecture that lets you easily:
    - Write [custom reporters](http://onsi.github.io/ginkgo/#writing_custom_reporters)
    - [Adapt an existing matcher library (or write your own!)](http://onsi.github.io/ginkgo/#using_other_matcher_libraries) to work with Ginkgo

## [Gomega](http://github.com/onsi/gomega): Ginkgo's Preferred Matcher Library

Learn more about Gomega [here](http://onsi.github.io/gomega/)

## License

Ginkgo is MIT-Licensed

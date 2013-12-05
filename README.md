![Ginkgo: A Golang BDD Testing Framework](http://onsi.github.io/ginkgo/images/ginkgo.png)

[![Build Status](https://travis-ci.org/onsi/ginkgo.png)](https://travis-ci.org/onsi/ginkgo)

Jump to the [docs](http://onsi.github.io/ginkgo/) to learn more.  To start rolling your Ginkgo tests *now* [keep reading](#set_me_up)!

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
    - Write [custom reporters](http://onsi.github.io/ginkgo/#writing_custom_reporters) (for example, Ginkgo comes with a [JUnit XML reporter](http://onsi.github.io/ginkgo/#generating_junit_xml_output))
    - [Adapt an existing matcher library (or write your own!)](http://onsi.github.io/ginkgo/#using_other_matcher_libraries) to work with Ginkgo

## [Gomega](http://github.com/onsi/gomega): Ginkgo's Preferred Matcher Library

Learn more about Gomega [here](http://onsi.github.io/gomega/)

## Set Me Up!

You'll need Golang v1.1+ (Ubuntu users: you probably have Golang v1.0 -- you'll need to upgrade!)

```bash

go get github.com/onsi/ginkgo/ginkgo  # installs the ginkgo CLI
go get github.com/onsi/gomega         # fetches the matcher library

cd path/to/package/you/want/to/test

ginkgo bootstrap # set up a new ginkgo suite
ginkgo generate  # will create a sample test file.  edit this file and add your tests then...

go test # to run your tests

ginkgo  # also runs your tests

```

## License

Ginkgo is MIT-Licensed

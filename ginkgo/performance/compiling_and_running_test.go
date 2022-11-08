package performance_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

var _ = Describe("Compiling and Running a single test package", func() {
	var cache gmeasure.ExperimentCache

	BeforeEach(func() {
		if os.Getenv("PERF") == "" {
			Skip("PERF environment not set, skipping")
		}

		var err error
		cache, err = gmeasure.NewExperimentCache("./compiling-and-running-single-cache")
		Ω(err).ShouldNot(HaveOccurred())

		// we mount everything outside the Ginkgo parent directory to make sure GOMODULES doesn't get confused by the go.mod in Ginkgo's root
		pfm = NewPerformanceFixtureManager(fmt.Sprintf("../../../ginkgo_perf_tmp_%d", GinkgoParallelProcess()))
		gmcm = NewGoModCacheManager(fmt.Sprintf("../../../ginkgo_perf_cache_%d", GinkgoParallelProcess()))
		if !DEBUG {
			DeferCleanup(pfm.Cleanup)
			DeferCleanup(gmcm.Cleanup)
		}
	})

	Describe("Experiments", func() {
		BeforeEach(func() {
			pfm.MountFixture("performance")
		})

		It("runs a series of experiments with various scenarios", func() {
			SampleScenarios(cache, 8, 1, true,
				ScenarioSettings{Fixture: "performance", NumSuites: 1, ConcurrentCompilers: 1, ConcurrentRunners: 1},
				ScenarioSettings{Fixture: "performance", NumSuites: 1, UseGoTestDirectly: true, ConcurrentGoTests: 1},
				ScenarioSettings{Fixture: "performance", NumSuites: 1, UseGoTestDirectly: true, GoTestCompileThenRunSerially: true},
			)
		})
	})

	Describe("Analysis", func() {
		It("analyzes the various scenarios to identify winners", func() {
			AnalyzeCache(cache)
		})
	})
})

var _ = Describe("Compiling and Running multiple tests", func() {
	var cache gmeasure.ExperimentCache

	BeforeEach(func() {
		if os.Getenv("PERF") == "" {
			Skip("PERF environment not set, skipping")
		}

		var err error
		cache, err = gmeasure.NewExperimentCache("./compiling-and-running-multiple-cache")
		Ω(err).ShouldNot(HaveOccurred())

		// we mount everything outside the Ginkgo parent directory to make sure GOMODULES doesn't get confused by the go.mod in Ginkgo's root
		pfm = NewPerformanceFixtureManager(fmt.Sprintf("../../../ginkgo_perf_tmp_%d", GinkgoParallelProcess()))
		gmcm = NewGoModCacheManager(fmt.Sprintf("../../../ginkgo_perf_cache_%d", GinkgoParallelProcess()))
		if !DEBUG {
			DeferCleanup(pfm.Cleanup)
			DeferCleanup(gmcm.Cleanup)
		}
	})

	Describe("Experiments", func() {
		BeforeEach(func() {
			pfm.MountFixture("performance")
		})

		It("runs a series of experiments with various scenarios", func() {
			SampleScenarios(cache, 8, 1, true,
				ScenarioSettings{Fixture: "performance", NumSuites: 5, ConcurrentCompilers: 1, ConcurrentRunners: 1, Recurse: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, ConcurrentCompilers: 2, ConcurrentRunners: 1, Recurse: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, ConcurrentCompilers: 4, ConcurrentRunners: 1, Recurse: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, ConcurrentCompilers: 2, ConcurrentRunners: 1, CompileFirstSuiteSerially: true, Recurse: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, ConcurrentCompilers: 4, ConcurrentRunners: 1, CompileFirstSuiteSerially: true, Recurse: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, ConcurrentCompilers: 2, ConcurrentRunners: 2, Recurse: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, ConcurrentCompilers: 4, ConcurrentRunners: 2, Recurse: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, ConcurrentCompilers: 2, ConcurrentRunners: 4, CompileFirstSuiteSerially: true, Recurse: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, ConcurrentCompilers: 4, ConcurrentRunners: 4, CompileFirstSuiteSerially: true, Recurse: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, UseGoTestDirectly: true, ConcurrentGoTests: 1, Recurse: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, UseGoTestDirectly: true, ConcurrentGoTests: 2, Recurse: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, UseGoTestDirectly: true, ConcurrentGoTests: 4, Recurse: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, UseGoTestDirectly: true, ConcurrentGoTests: 8, Recurse: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, UseGoTestDirectly: true, GoTestCompileThenRunSerially: true, Recurse: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, UseGoTestDirectly: true, GoTestRecurse: true, Recurse: true},
			)
		})
	})

	Describe("Analysis", func() {
		It("analyzes the various scenarios to identify winners", func() {
			AnalyzeCache(cache)
		})
	})
})

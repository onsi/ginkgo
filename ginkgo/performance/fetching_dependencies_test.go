package performance_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

var _ = Describe("Fetching Dependencies", func() {
	var cache gmeasure.ExperimentCache

	BeforeEach(func() {
		var err error
		cache, err = gmeasure.NewExperimentCache("./fetching-dependencies-cache")
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
			SampleScenarios(cache, 8, 1, false,
				ScenarioSettings{Fixture: "performance", NumSuites: 5, ConcurrentCompilers: 1, ConcurrentRunners: 1, Recurse: true, ClearGoModCache: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, ConcurrentCompilers: 2, ConcurrentRunners: 1, Recurse: true, ClearGoModCache: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, ConcurrentCompilers: 4, ConcurrentRunners: 1, Recurse: true, ClearGoModCache: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, ConcurrentCompilers: 1, ConcurrentRunners: 1, GoModDownloadFirst: true, Recurse: true, ClearGoModCache: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, ConcurrentCompilers: 2, ConcurrentRunners: 1, GoModDownloadFirst: true, Recurse: true, ClearGoModCache: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, ConcurrentCompilers: 4, ConcurrentRunners: 1, GoModDownloadFirst: true, Recurse: true, ClearGoModCache: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, ConcurrentCompilers: 2, ConcurrentRunners: 1, CompileFirstSuiteSerially: true, Recurse: true, ClearGoModCache: true},
				ScenarioSettings{Fixture: "performance", NumSuites: 5, ConcurrentCompilers: 4, ConcurrentRunners: 1, CompileFirstSuiteSerially: true, Recurse: true, ClearGoModCache: true},
			)
		})
	})

	Describe("Analysis", func() {
		It("analyzes the various fetching dependencies scenarios to identify winners", func() {
			AnalyzeCache(cache)
		})
	})
})

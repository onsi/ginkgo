package performance_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/gmeasure"
)

func LoadOrCreate(cache gmeasure.ExperimentCache, name string, version int) (*gmeasure.Experiment, bool) {
	experiment := cache.Load(name, version)
	if experiment != nil {
		return experiment, true
	}
	return gmeasure.NewExperiment(name), false
}

var _ = Describe("Running a large test suite", Ordered, Serial, func() {
	var cache gmeasure.ExperimentCache
	var REGENERATE_BENCHMARK = os.Getenv("BENCH") != ""
	const BENCHMARK_VERSION = 2
	const N = 10

	var runtimes = []gmeasure.Stats{}

	BeforeAll(func() {
		if os.Getenv("PERF") == "" {
			Skip("PERF environment not set, skipping")
		}

		var err error
		cache, err = gmeasure.NewExperimentCache("./large-suite-cache")
		Ω(err).ShouldNot(HaveOccurred())

		pfm = NewPerformanceFixtureManager(fmt.Sprintf("./ginkgo_perf_tmp_%d", GinkgoParallelProcess()))
		if !DEBUG {
			DeferCleanup(pfm.Cleanup)
		}
		pfm.MountFixture("large_suite")

		session := startGinkgo(pfm.PathTo("large_suite"), "build")
		Eventually(session).Should(gexec.Exit(0))
		Expect(pfm.PathTo("large_suite", "large_suite.test")).To(BeAnExistingFile())
	})

	var nameFor = func(nodes int, protocol string, interceptor string) string {
		if nodes == 1 {
			return "serial"
		}
		return "parallel" + "-" + protocol + "-" + interceptor
	}

	DescribeTable("scenarios",
		func(nodes int, protocol string, interceptor string) {
			var experiment *gmeasure.Experiment
			name := nameFor(nodes, protocol, interceptor)

			if REGENERATE_BENCHMARK {
				experiment = gmeasure.NewExperiment(name + "-benchmark")
			} else {
				benchmark := cache.Load(name+"-benchmark", BENCHMARK_VERSION)
				Ω(benchmark).ShouldNot(BeNil())
				runtimes = append(runtimes, benchmark.GetStats("runtime"))
				experiment = gmeasure.NewExperiment(name)
			}
			AddReportEntry(experiment.Name, experiment)

			env := []string{}
			if nodes > 1 {
				env = append(env, "GINKGO_PARALLEL_PROTOCOL="+protocol)
			}

			experiment.SampleDuration("runtime", func(idx int) {
				fmt.Printf("Running %s %d/%d\n", name, idx+1, N)
				session := startGinkgoWithEnv(
					pfm.PathTo("large_suite"),
					env,
					fmt.Sprintf("--procs=%d", nodes),
					fmt.Sprintf("--output-interceptor-mode=%s", interceptor),
					"large_suite.test",
				)
				Eventually(session).Should(gexec.Exit(0))
			}, gmeasure.SamplingConfig{N: N})
			runtimes = append(runtimes, experiment.GetStats("runtime"))

			fmt.Printf("Profiling %s\n", name)
			session := startGinkgoWithEnv(
				pfm.PathTo("large_suite"),
				env,
				fmt.Sprintf("--procs=%d", nodes),
				fmt.Sprintf("--output-interceptor-mode=%s", interceptor),
				"--cpuprofile=CPU.profile",
				"--blockprofile=BLOCK.profile",
				"large_suite.test",
			)
			Eventually(session).Should(gexec.Exit(0))

			if REGENERATE_BENCHMARK {
				cache.Save(experiment.Name, BENCHMARK_VERSION, experiment)
			}
		},
		nameFor,
		Entry(nil, 1, "", ""),
		Entry(nil, 2, "RPC", "DUP"),
		Entry(nil, 2, "RPC", "SWAP"),
		Entry(nil, 2, "RPC", "NONE"),
		Entry(nil, 2, "HTTP", "DUP"),
		Entry(nil, 2, "HTTP", "SWAP"),
		Entry(nil, 2, "HTTP", "NONE"),
	)

	It("analyzes the experiments", func() {
		if REGENERATE_BENCHMARK {
			Skip("no analysis when generating benchmark")
		}
		AddReportEntry("Ranking", gmeasure.RankStats(gmeasure.LowerMedianIsBetter, runtimes...))
	})
})

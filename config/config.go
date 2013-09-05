package config

import (
	"flag"
	"time"
)

type GinkgoConfigType struct {
	RandomSeed        int64
	RandomizeAllSpecs bool
	FocusString       string
	ParallelNode      int
	ParallelTotal     int
}

var GinkgoConfig = GinkgoConfigType{}

type DefaultReporterConfigType struct {
	NoColor           bool
	SlowSpecThreshold float64
	NoisyPendings     bool
}

var DefaultReporterConfig = DefaultReporterConfigType{}

func init() {
	flag.Int64Var(&(GinkgoConfig.RandomSeed), "ginkgo.seed", time.Now().Unix(), "The seed used to randomize the spec suite.")
	flag.BoolVar(&(GinkgoConfig.RandomizeAllSpecs), "ginkgo.randomizeAllSpecs", false, "If set, ginkgo will randomize all specs together.  By default, ginkgo only randomizes the top level Describe/Context groups.")
	flag.StringVar(&(GinkgoConfig.FocusString), "ginkgo.focus", "", "If set, ginkgo will only run specs that match this regular expression.")
	flag.IntVar(&(GinkgoConfig.ParallelNode), "ginkgo.parallel.node", 1, "This worker node's (one-indexed) node number.  For running specs in parallel.")
	flag.IntVar(&(GinkgoConfig.ParallelTotal), "ginkgo.parallel.total", 1, "The total number of worker nodes.  For running specs in parallel.")

	flag.BoolVar(&(DefaultReporterConfig.NoColor), "ginkgo.noColor", false, "If set, suppress color output in default reporter.")
	flag.Float64Var(&(DefaultReporterConfig.SlowSpecThreshold), "ginkgo.slowSpecThreshold", 5.0, "(in seconds) Specs that take longer to run than this threshold are flagged as slow by the default reporter (default: 5 seconds).")
	flag.BoolVar(&(DefaultReporterConfig.NoisyPendings), "ginkgo.noisyPendings", true, "If set, default reporter will shout about pending tests.")

	flag.Parse()
}

/*
Ginkgo accepts a number of configuration options.

These are documented [here](http://onsi.github.io/ginkgo/#the_ginkgo_cli)

You can also learn more via

	ginkgo help

or (I kid you not):

	go test -asdf
*/
package config

import (
	"flag"
	"time"

	"fmt"
)

const VERSION = "0.9"

type GinkgoConfigType struct {
	RandomSeed        int64
	RandomizeAllSpecs bool
	FocusString       string
	SkipString        string
	ParallelNode      int
	ParallelTotal     int
	SkipMeasurements  bool
	FailOnPending     bool
}

var GinkgoConfig = GinkgoConfigType{}

type DefaultReporterConfigType struct {
	NoColor           bool
	SlowSpecThreshold float64
	NoisyPendings     bool
	Succinct          bool
	Verbose           bool
}

var DefaultReporterConfig = DefaultReporterConfigType{}

func processPrefix(prefix string) string {
	if prefix != "" {
		prefix = prefix + "."
	}
	return prefix
}

func Flags(prefix string, includeParallelFlags bool) {
	prefix = processPrefix(prefix)
	flag.Int64Var(&(GinkgoConfig.RandomSeed), prefix+"seed", time.Now().Unix(), "The seed used to randomize the spec suite.")
	flag.BoolVar(&(GinkgoConfig.RandomizeAllSpecs), prefix+"randomizeAllSpecs", false, "If set, ginkgo will randomize all specs together.  By default, ginkgo only randomizes the top level Describe/Context groups.")
	flag.BoolVar(&(GinkgoConfig.SkipMeasurements), prefix+"skipMeasurements", false, "If set, ginkgo will skip any measurement specs.")
	flag.BoolVar(&(GinkgoConfig.FailOnPending), prefix+"failOnPending", false, "If set, ginkgo will mark the test suite as failed if any specs are pending.")
	flag.StringVar(&(GinkgoConfig.FocusString), prefix+"focus", "", "If set, ginkgo will only run specs that match this regular expression.")
	flag.StringVar(&(GinkgoConfig.SkipString), prefix+"skip", "", "If set, ginkgo will only run specs that do not match this regular expression.")

	if includeParallelFlags {
		flag.IntVar(&(GinkgoConfig.ParallelNode), prefix+"parallel.node", 1, "This worker node's (one-indexed) node number.  For running specs in parallel.")
		flag.IntVar(&(GinkgoConfig.ParallelTotal), prefix+"parallel.total", 1, "The total number of worker nodes.  For running specs in parallel.")
	}

	flag.BoolVar(&(DefaultReporterConfig.NoColor), prefix+"noColor", false, "If set, suppress color output in default reporter.")
	flag.Float64Var(&(DefaultReporterConfig.SlowSpecThreshold), prefix+"slowSpecThreshold", 5.0, "(in seconds) Specs that take longer to run than this threshold are flagged as slow by the default reporter (default: 5 seconds).")
	flag.BoolVar(&(DefaultReporterConfig.NoisyPendings), prefix+"noisyPendings", true, "If set, default reporter will shout about pending tests.")
	flag.BoolVar(&(DefaultReporterConfig.Verbose), prefix+"v", false, "If set, default reporter print out all specs as they begin.")
	flag.BoolVar(&(DefaultReporterConfig.Succinct), prefix+"succinct", false, "If set, default reporter prints out succinct code location information.")
}

func BuildFlagArgs(prefix string, ginkgo GinkgoConfigType, reporter DefaultReporterConfigType) []string {
	prefix = processPrefix(prefix)
	result := make([]string, 0)

	if ginkgo.RandomSeed > 0 {
		result = append(result, fmt.Sprintf("--%sseed=%d", prefix, ginkgo.RandomSeed))
	}

	if ginkgo.RandomizeAllSpecs {
		result = append(result, fmt.Sprintf("--%srandomizeAllSpecs", prefix))
	}

	if ginkgo.SkipMeasurements {
		result = append(result, fmt.Sprintf("--%sskipMeasurements", prefix))
	}

	if ginkgo.FailOnPending {
		result = append(result, fmt.Sprintf("--%sfailOnPending", prefix))
	}

	if ginkgo.FocusString != "" {
		result = append(result, fmt.Sprintf("--%sfocus=%s", prefix, ginkgo.FocusString))
	}

	if ginkgo.SkipString != "" {
		result = append(result, fmt.Sprintf("--%sskip=%s", prefix, ginkgo.SkipString))
	}

	if ginkgo.ParallelNode != 0 {
		result = append(result, fmt.Sprintf("--%sparallel.node=%d", prefix, ginkgo.ParallelNode))
	}

	if ginkgo.ParallelTotal != 0 {
		result = append(result, fmt.Sprintf("--%sparallel.total=%d", prefix, ginkgo.ParallelTotal))
	}

	if reporter.NoColor {
		result = append(result, fmt.Sprintf("--%snoColor", prefix))
	}

	if reporter.SlowSpecThreshold > 0 {
		result = append(result, fmt.Sprintf("--%sslowSpecThreshold=%.5f", prefix, reporter.SlowSpecThreshold))
	}

	if !reporter.NoisyPendings {
		result = append(result, fmt.Sprintf("--%snoisyPendings=false", prefix))
	}

	if reporter.Verbose {
		result = append(result, fmt.Sprintf("--%sv", prefix))
	}

	if reporter.Succinct {
		result = append(result, fmt.Sprintf("--%ssuccinct", prefix))
	}

	return result
}

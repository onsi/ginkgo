package ginkgo

import (
	"flag"
	"time"
)

type GinkoConfigType struct {
	RandomSeed        *int64
	RandomizeAllSpecs *bool
}

type defaultReporterConfigType struct {
	noColor           *bool
	slowSpecThreshold *float64
	noisyPendings     *bool
}

var GinkgoConfig = GinkoConfigType{
	RandomSeed:        flag.Int64("ginkgo.seed", time.Now().Unix(), "The seed used to randomize the spec suite."),
	RandomizeAllSpecs: flag.Bool("ginkgo.randomizeAllSpecs", false, "If set, ginkgo will randomize all specs together.  By default, ginkgo only randomizes the top level Describe/Context groups."),
}

var defaultReporterConfig = defaultReporterConfigType{
	noColor:           flag.Bool("ginkgo.noColor", false, "If set, suppress color output in default reporter."),
	slowSpecThreshold: flag.Float64("ginkgo.slowSpecThreshold", 5.0, "(in seconds) Specs that take longer to run than this threshold are flagged as slow by the default reporter (default: 5 seconds)."),
	noisyPendings:     flag.Bool("ginkgo.noisyPendings", true, "If set, default reporter will shout about pending tests."),
}

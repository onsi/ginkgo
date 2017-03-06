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
	"fmt"
	"time"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

const VERSION = "1.2.0"

type GinkgoConfigType struct {
	RandomSeed         int64
	RandomizeAllSpecs  bool
	RegexScansFilePath bool
	FocusString        string
	SkipString         string
	SkipMeasurements   bool
	FailOnPending      bool
	FailFast           bool
	FlakeAttempts      int
	EmitSpecProgress   bool
	DryRun             bool

	ParallelNode  int
	ParallelTotal int
	SyncHost      string
	StreamHost    string
}

var GinkgoConfig = GinkgoConfigType{}

type DefaultReporterConfigType struct {
	SlowSpecThreshold float64
	NoisyPendings     bool
	Succinct          bool
	Verbose           bool
	FullTrace         bool

	color   bool
	noColor bool

	// Colorize caches whether or not we want to colorize output
	Colorize *bool
}

// ColorizeOutput determines if output should be colorized. The user
// can specify they want color (--color) or not (--noColor). Further-
// more, we detect if the output to which we are printing is a TTY.
// If the output is a TTY, the default is to use color but a user can
// still request no color; if the output is not a TTY, the default is
// to not use color but the user can still request it.
func (c *DefaultReporterConfigType) ColorizeOutput() bool {
	if c.Colorize == nil {
		var colorize bool
		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			colorize = !c.noColor
		} else {
			colorize = c.color
		}
		c.Colorize = &colorize
	}

	return *c.Colorize
}

var DefaultReporterConfig = DefaultReporterConfigType{}

func processPrefix(prefix string) string {
	if prefix != "" {
		prefix = prefix + "."
	}
	return prefix
}

func Flags(flagSet *flag.FlagSet, prefix string, includeParallelFlags bool) {
	prefix = processPrefix(prefix)
	flagSet.Int64Var(&(GinkgoConfig.RandomSeed), prefix+"seed", time.Now().Unix(), "The seed used to randomize the spec suite.")
	flagSet.BoolVar(&(GinkgoConfig.RandomizeAllSpecs), prefix+"randomizeAllSpecs", false, "If set, ginkgo will randomize all specs together.  By default, ginkgo only randomizes the top level Describe/Context groups.")
	flagSet.BoolVar(&(GinkgoConfig.SkipMeasurements), prefix+"skipMeasurements", false, "If set, ginkgo will skip any measurement specs.")
	flagSet.BoolVar(&(GinkgoConfig.FailOnPending), prefix+"failOnPending", false, "If set, ginkgo will mark the test suite as failed if any specs are pending.")
	flagSet.BoolVar(&(GinkgoConfig.FailFast), prefix+"failFast", false, "If set, ginkgo will stop running a test suite after a failure occurs.")

	flagSet.BoolVar(&(GinkgoConfig.DryRun), prefix+"dryRun", false, "If set, ginkgo will walk the test hierarchy without actually running anything.  Best paired with -v.")

	flagSet.StringVar(&(GinkgoConfig.FocusString), prefix+"focus", "", "If set, ginkgo will only run specs that match this regular expression.")
	flagSet.StringVar(&(GinkgoConfig.SkipString), prefix+"skip", "", "If set, ginkgo will only run specs that do not match this regular expression.")

	flagSet.BoolVar(&(GinkgoConfig.RegexScansFilePath), prefix+"regexScansFilePath", false, "If set, ginkgo regex matching also will look at the file path (code location).")

	flagSet.IntVar(&(GinkgoConfig.FlakeAttempts), prefix+"flakeAttempts", 1, "Make up to this many attempts to run each spec. Please note that if any of the attempts succeed, the suite will not be failed. But any failures will still be recorded.")

	flagSet.BoolVar(&(GinkgoConfig.EmitSpecProgress), prefix+"progress", false, "If set, ginkgo will emit progress information as each spec runs to the GinkgoWriter.")

	if includeParallelFlags {
		flagSet.IntVar(&(GinkgoConfig.ParallelNode), prefix+"parallel.node", 1, "This worker node's (one-indexed) node number.  For running specs in parallel.")
		flagSet.IntVar(&(GinkgoConfig.ParallelTotal), prefix+"parallel.total", 1, "The total number of worker nodes.  For running specs in parallel.")
		flagSet.StringVar(&(GinkgoConfig.SyncHost), prefix+"parallel.synchost", "", "The address for the server that will synchronize the running nodes.")
		flagSet.StringVar(&(GinkgoConfig.StreamHost), prefix+"parallel.streamhost", "", "The address for the server that the running nodes should stream data to.")
	}

	flagSet.BoolVar(&(DefaultReporterConfig.color), prefix+"color", false, "If set, enable color output in default reporter even when not printing to a TTY.")
	flagSet.BoolVar(&(DefaultReporterConfig.noColor), prefix+"noColor", false, "If set, suppress color output in default reporter.")
	flagSet.Float64Var(&(DefaultReporterConfig.SlowSpecThreshold), prefix+"slowSpecThreshold", 5.0, "(in seconds) Specs that take longer to run than this threshold are flagged as slow by the default reporter.")
	flagSet.BoolVar(&(DefaultReporterConfig.NoisyPendings), prefix+"noisyPendings", true, "If set, default reporter will shout about pending tests.")
	flagSet.BoolVar(&(DefaultReporterConfig.Verbose), prefix+"v", false, "If set, default reporter print out all specs as they begin.")
	flagSet.BoolVar(&(DefaultReporterConfig.Succinct), prefix+"succinct", false, "If set, default reporter prints out a very succinct report")
	flagSet.BoolVar(&(DefaultReporterConfig.FullTrace), prefix+"trace", false, "If set, default reporter prints out the full stack trace when a failure occurs")
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

	if ginkgo.FailFast {
		result = append(result, fmt.Sprintf("--%sfailFast", prefix))
	}

	if ginkgo.DryRun {
		result = append(result, fmt.Sprintf("--%sdryRun", prefix))
	}

	if ginkgo.FocusString != "" {
		result = append(result, fmt.Sprintf("--%sfocus=%s", prefix, ginkgo.FocusString))
	}

	if ginkgo.SkipString != "" {
		result = append(result, fmt.Sprintf("--%sskip=%s", prefix, ginkgo.SkipString))
	}

	if ginkgo.FlakeAttempts > 1 {
		result = append(result, fmt.Sprintf("--%sflakeAttempts=%d", prefix, ginkgo.FlakeAttempts))
	}

	if ginkgo.EmitSpecProgress {
		result = append(result, fmt.Sprintf("--%sprogress", prefix))
	}

	if ginkgo.ParallelNode != 0 {
		result = append(result, fmt.Sprintf("--%sparallel.node=%d", prefix, ginkgo.ParallelNode))
	}

	if ginkgo.ParallelTotal != 0 {
		result = append(result, fmt.Sprintf("--%sparallel.total=%d", prefix, ginkgo.ParallelTotal))
	}

	if ginkgo.StreamHost != "" {
		result = append(result, fmt.Sprintf("--%sparallel.streamhost=%s", prefix, ginkgo.StreamHost))
	}

	if ginkgo.SyncHost != "" {
		result = append(result, fmt.Sprintf("--%sparallel.synchost=%s", prefix, ginkgo.SyncHost))
	}

	if ginkgo.RegexScansFilePath {
		result = append(result, fmt.Sprintf("--%sregexScansFilePath", prefix))
	}

	if reporter.ColorizeOutput() {
		result = append(result, fmt.Sprintf("--%scolor", prefix))
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

	if reporter.FullTrace {
		result = append(result, fmt.Sprintf("--%strace", prefix))
	}

	return result
}

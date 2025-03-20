package performance_test

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/ginkgo/internal"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/gmeasure"
)

var pathToGinkgo string
var DEBUG bool
var pfm PerformanceFixtureManager
var gmcm GoModCacheManager

func init() {
	flag.BoolVar(&DEBUG, "debug", false, "keep assets around after test run")
}

func TestPerformance(t *testing.T) {
	SetDefaultEventuallyTimeout(30 * time.Second)
	format.TruncatedDiff = false
	RegisterFailHandler(Fail)
	RunSpecs(t, "Performance Suite", Label("performance"))
}

var _ = SynchronizedBeforeSuite(func() []byte {
	pathToGinkgo, err := gexec.Build("../../ginkgo")
	Ω(err).ShouldNot(HaveOccurred())
	return []byte(pathToGinkgo)
}, func(computedPathToGinkgo []byte) {
	pathToGinkgo = string(computedPathToGinkgo)
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
})

/*
GoModCacheManager sets up a new GOMODCACHE and knows how to clear it
This allows us to bust the go mod cache.
*/
type GoModCacheManager struct {
	Path         string
	OldCachePath string
}

func NewGoModCacheManager(path string) GoModCacheManager {
	err := os.MkdirAll(path, 0700)
	Ω(err).ShouldNot(HaveOccurred())
	absPath, err := filepath.Abs(path)
	Ω(err).ShouldNot(HaveOccurred())
	oldCachePath := os.Getenv("GOMODCACHE")
	os.Setenv("GOMODCACHE", absPath)
	return GoModCacheManager{
		Path:         path,
		OldCachePath: oldCachePath,
	}

}

func (m GoModCacheManager) Clear() {
	cmd := exec.Command("go", "clean", "-modcache")
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Ω(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
}

func (m GoModCacheManager) Cleanup() {
	m.Clear()
	if m.OldCachePath == "" {
		os.Unsetenv("GOMODCACHE")
	} else {
		os.Setenv("GOMODCACHE", m.OldCachePath)
	}
}

/* PerformanceFixtureManager manages fixture data */
type PerformanceFixtureManager struct {
	TmpDir string
}

func NewPerformanceFixtureManager(tmpDir string) PerformanceFixtureManager {
	err := os.MkdirAll(tmpDir, 0700)
	Ω(err).ShouldNot(HaveOccurred())
	return PerformanceFixtureManager{
		TmpDir: tmpDir,
	}
}

func (f PerformanceFixtureManager) Cleanup() {
	Ω(os.RemoveAll(f.TmpDir)).Should(Succeed())
}

func (f PerformanceFixtureManager) MountFixture(fixture string, subPackage ...string) {
	src := filepath.Join("_fixtures", fixture+"_fixture")
	dst := filepath.Join(f.TmpDir, fixture)

	if len(subPackage) > 0 {
		src = filepath.Join(src, subPackage[0])
		dst = filepath.Join(dst, subPackage[0])
	}

	f.copyIn(src, dst)
}

func (f PerformanceFixtureManager) copyIn(src string, dst string) {
	Expect(os.MkdirAll(dst, 0777)).To(Succeed())

	files, err := os.ReadDir(src)
	Expect(err).NotTo(HaveOccurred())

	for _, file := range files {
		srcPath := filepath.Join(src, file.Name())
		dstPath := filepath.Join(dst, file.Name())
		if file.IsDir() {
			f.copyIn(srcPath, dstPath)
			continue
		}

		srcContent, err := os.ReadFile(srcPath)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(os.WriteFile(dstPath, srcContent, 0666)).Should(Succeed())
	}
}

func (f PerformanceFixtureManager) PathTo(pkg string, target ...string) string {
	if len(target) == 0 {
		return filepath.Join(f.TmpDir, pkg)
	}
	components := append([]string{f.TmpDir, pkg}, target...)
	return filepath.Join(components...)
}

/* GoModDownload runs go mod download for a given package */
func GoModDownload(fixture string) {
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = pfm.PathTo(fixture)
	sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Ω(err).ShouldNot(HaveOccurred())
	Eventually(sess).Should(gexec.Exit(0))
}

/* ScenarioSettings configures the test scenario */
type ScenarioSettings struct {
	Fixture         string
	NumSuites       int
	Recurse         bool
	ClearGoModCache bool

	ConcurrentCompilers       int
	ConcurrentRunners         int
	CompileFirstSuiteSerially bool
	GoModDownloadFirst        bool

	UseGoTestDirectly            bool
	ConcurrentGoTests            int
	GoTestCompileThenRunSerially bool
	GoTestRecurse                bool
}

func (s ScenarioSettings) Name() string {
	out := []string{s.Fixture}
	if s.UseGoTestDirectly {
		if s.GoTestCompileThenRunSerially {
			out = append(out, "go test -c; then run (serially)")
		} else if s.GoTestRecurse {
			out = append(out, "go test ./...")
		} else {
			out = append(out, "go test")
			if s.ConcurrentGoTests == 1 {
				out = append(out, "serially")
			} else {
				out = append(out, fmt.Sprintf("run concurrently [%d]", s.ConcurrentGoTests))
			}
		}
	} else {
		if s.ConcurrentCompilers == 1 {
			out = append(out, "compile serially")
		} else {
			out = append(out, fmt.Sprintf("compile concurrently [%d]", s.ConcurrentCompilers))
		}
		if s.ConcurrentRunners == 1 {
			out = append(out, "run serially")
		} else {
			out = append(out, fmt.Sprintf("run concurrently [%d]", s.ConcurrentRunners))
		}
		if s.CompileFirstSuiteSerially {
			out = append(out, "will compile first suite serially")
		}
		if s.GoModDownloadFirst {
			out = append(out, "will go mod download first")
		}
	}
	return strings.Join(out, " - ")
}

func SampleScenarios(cache gmeasure.ExperimentCache, numSamples int, cacheVersion int, runGoModDownload bool, scenarios ...ScenarioSettings) {
	// we randomize the sample set of scenarios to try to avoid any systematic effects that emerge
	// during the run (e.g. change in internet connection speed, change in computer performance)

	experiments := map[string]*gmeasure.Experiment{}
	runs := []ScenarioSettings{}
	for _, scenario := range scenarios {
		name := scenario.Name()
		if experiment := cache.Load(name, cacheVersion); experiment != nil {
			AddReportEntry(name, experiment, Offset(1), ReportEntryVisibilityFailureOrVerbose)
			continue
		}
		experiments[name] = gmeasure.NewExperiment(name)
		AddReportEntry(name, experiments[name], Offset(1), ReportEntryVisibilityFailureOrVerbose)
		for i := 0; i < numSamples; i++ {
			runs = append(runs, scenario)
		}
	}
	rand.New(rand.NewSource(GinkgoRandomSeed())).Shuffle(len(runs), func(i, j int) {
		runs[i], runs[j] = runs[j], runs[i]
	})

	if len(runs) > 0 && runGoModDownload {
		GoModDownload("performance")
	}

	for idx, run := range runs {
		fmt.Printf("%d - %s\n", idx, run.Name())
		RunScenario(experiments[run.Name()].NewStopwatch(), run, gmeasure.Annotation(fmt.Sprintf("%d", idx+1)))
	}

	for name, experiment := range experiments {
		cache.Save(name, cacheVersion, experiment)
	}
}

func AnalyzeCache(cache gmeasure.ExperimentCache) {
	headers, err := cache.List()
	Ω(err).ShouldNot(HaveOccurred())

	experiments := []*gmeasure.Experiment{}
	for _, header := range headers {
		experiments = append(experiments, cache.Load(header.Name, header.Version))
	}

	for _, measurement := range []string{"first-output", "total-runtime"} {
		stats := []gmeasure.Stats{}
		for _, experiment := range experiments {
			stats = append(stats, experiment.GetStats(measurement))
		}
		AddReportEntry(measurement, gmeasure.RankStats(gmeasure.LowerMedianIsBetter, stats...))
	}
}

func RunScenario(stopwatch *gmeasure.Stopwatch, settings ScenarioSettings, annotation gmeasure.Annotation) {
	if settings.ClearGoModCache {
		gmcm.Clear()
	}

	if settings.GoModDownloadFirst {
		GoModDownload(settings.Fixture)
		stopwatch.Record("mod-download", annotation)
	}

	if settings.UseGoTestDirectly {
		RunScenarioWithGoTest(stopwatch, settings, annotation)
	} else {
		RunScenarioWithGinkgoInternals(stopwatch, settings, annotation)
	}
}

/* CompileAndRun uses the Ginkgo CLIs internals to compile and run tests with different possible settings governing concurrency and ordering */
func RunScenarioWithGinkgoInternals(stopwatch *gmeasure.Stopwatch, settings ScenarioSettings, annotation gmeasure.Annotation) {
	cliConfig := types.NewDefaultCLIConfig()
	cliConfig.Recurse = settings.Recurse
	suiteConfig := types.NewDefaultSuiteConfig()
	reporterConfig := types.NewDefaultReporterConfig()
	reporterConfig.Succinct = true
	goFlagsConfig := types.NewDefaultGoFlagsConfig()

	suites := internal.FindSuites([]string{pfm.PathTo(settings.Fixture)}, cliConfig, true)
	Ω(suites).Should(HaveLen(settings.NumSuites))

	compile := make(chan internal.TestSuite, len(suites))
	compiled := make(chan internal.TestSuite, len(suites))
	completed := make(chan internal.TestSuite, len(suites))
	firstOutputOnce := sync.Once{}

	for compiler := 0; compiler < settings.ConcurrentCompilers; compiler++ {
		go func() {
			for suite := range compile {
				if !suite.State.Is(internal.TestSuiteStateCompiled) {
					subStopwatch := stopwatch.NewStopwatch()
					suite = internal.CompileSuite(suite, goFlagsConfig, false)
					subStopwatch.Record("compile-test: "+suite.PackageName, annotation)
					Ω(suite.CompilationError).Should(BeNil())
				}
				compiled <- suite
			}
		}()
	}

	if settings.CompileFirstSuiteSerially {
		compile <- suites[0]
		suites[0] = <-compiled
	}

	for runner := 0; runner < settings.ConcurrentRunners; runner++ {
		go func() {
			for suite := range compiled {
				firstOutputOnce.Do(func() {
					stopwatch.Record("first-output", annotation, gmeasure.Style("{{cyan}}"))
				})
				subStopwatch := stopwatch.NewStopwatch()
				suite = internal.RunCompiledSuite(suite, suiteConfig, reporterConfig, cliConfig, goFlagsConfig, []string{})
				subStopwatch.Record("run-test: "+suite.PackageName, annotation)
				Ω(suite.State).Should(Equal(internal.TestSuiteStatePassed))
				completed <- suite
			}
		}()
	}

	for _, suite := range suites {
		compile <- suite
	}

	completedSuites := []internal.TestSuite{}
	for suite := range completed {
		completedSuites = append(completedSuites, suite)
		if len(completedSuites) == len(suites) {
			close(completed)
			close(compile)
			close(compiled)
		}
	}

	stopwatch.Record("total-runtime", annotation, gmeasure.Style("{{green}}"))
	internal.Cleanup(goFlagsConfig, completedSuites...)
}

func RunScenarioWithGoTest(stopwatch *gmeasure.Stopwatch, settings ScenarioSettings, annotation gmeasure.Annotation) {
	defer func() {
		stopwatch.Record("total-runtime", annotation, gmeasure.Style("{{green}}"))
	}()

	if settings.GoTestRecurse {
		cmd := exec.Command("go", "test", "-count=1", "./...")
		cmd.Dir = pfm.PathTo(settings.Fixture)
		sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())
		Eventually(sess).Should(gbytes.Say(`.`)) //should say _something_ eventually!
		stopwatch.Record("first-output", annotation, gmeasure.Style("{{cyan}}"))
		Eventually(sess).Should(gexec.Exit(0))
		return
	}

	cliConfig := types.NewDefaultCLIConfig()
	cliConfig.Recurse = settings.Recurse
	suites := internal.FindSuites([]string{pfm.PathTo(settings.Fixture)}, cliConfig, true)
	Ω(suites).Should(HaveLen(settings.NumSuites))
	firstOutputOnce := sync.Once{}

	if settings.GoTestCompileThenRunSerially {
		for _, suite := range suites {
			subStopwatch := stopwatch.NewStopwatch()
			cmd := exec.Command("go", "test", "-c", "-o=out.test")
			cmd.Dir = suite.AbsPath()
			sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Ω(err).ShouldNot(HaveOccurred())
			Eventually(sess).Should(gexec.Exit(0))
			subStopwatch.Record("compile-test: "+suite.PackageName, annotation).Reset()

			firstOutputOnce.Do(func() {
				stopwatch.Record("first-output", annotation, gmeasure.Style("{{cyan}}"))
			})
			cmd = exec.Command("./out.test")
			cmd.Dir = suite.AbsPath()
			sess, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Ω(err).ShouldNot(HaveOccurred())
			Eventually(sess).Should(gexec.Exit(0))
			subStopwatch.Record("run-test: "+suite.PackageName, annotation)

			Ω(os.Remove(filepath.Join(suite.AbsPath(), "out.test"))).Should(Succeed())
		}
	} else {
		run := make(chan internal.TestSuite, len(suites))
		completed := make(chan internal.TestSuite, len(suites))
		for runner := 0; runner < settings.ConcurrentGoTests; runner++ {
			go func() {
				for suite := range run {
					subStopwatch := stopwatch.NewStopwatch()
					cmd := exec.Command("go", "test", "-count=1")
					cmd.Dir = suite.AbsPath()
					sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Ω(err).ShouldNot(HaveOccurred())
					Eventually(sess).Should(gbytes.Say(`.`)) //should say _something_ eventually!
					firstOutputOnce.Do(func() {
						stopwatch.Record("first-output", annotation, gmeasure.Style("{{cyan}}"))
					})
					Eventually(sess).Should(gexec.Exit(0))
					subStopwatch.Record("run-test: "+suite.PackageName, annotation)
					completed <- suite
				}
			}()
		}
		for _, suite := range suites {
			run <- suite
		}
		numCompleted := 0
		for _ = range completed {
			numCompleted += 1
			if numCompleted == len(suites) {
				close(completed)
				close(run)
			}
		}
	}
}

func ginkgoCommand(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command(pathToGinkgo, args...)
	cmd.Dir = dir

	return cmd
}

func startGinkgo(dir string, args ...string) *gexec.Session {
	cmd := ginkgoCommand(dir, args...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Ω(err).ShouldNot(HaveOccurred())
	return session
}

func startGinkgoWithEnv(dir string, env []string, args ...string) *gexec.Session {
	cmd := ginkgoCommand(dir, args...)
	cmd.Env = append(os.Environ(), env...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Ω(err).ShouldNot(HaveOccurred())
	return session
}

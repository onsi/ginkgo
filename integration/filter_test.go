package integration_test

import (
        "path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
)

var _ = Describe("Filter", func() {
	BeforeEach(func() {
		fm.MountFixture("filter")
	})

	It("honors the focus, skip, focus-file and skip-file flags", func() {
		session := startGinkgo(fm.PathTo("filter"),
			"--focus=dog", "--focus=fish",
			"--skip=cat",
			"--focus-file=sprocket", "--focus-file=widget:1-24", "--focus-file=_b:24-42",
			"--skip-file=_c",
			"--json-report=report.json",
			"--label-filter=TopLevelLabel && !SLOW && !(Feature: containsAny Alpha)",
		)
		Eventually(session).Should(gexec.Exit(0))
		specs := Reports(fm.LoadJSONReports("filter", "report.json")[0].SpecReports)

		passedSpecs := []string{
			"SprocketA dog fish",
			"SprocketB dog", "SprocketB dog fish",
			"WidgetA dog", "WidgetA dog fish",
			"WidgetB dog fish",
			// lines in _b > 24 are in --focus-file
			"More WidgetB dog", "More WidgetB dog fish",
		}

		skippedSpecs := []string{
			// nugget files are not in focus-file
			"NuggetA cat", "NuggetA dog", "NuggetA cat fish", "NuggetA dog fish",
			"NuggetB cat", "NuggetB dog", "NuggetB cat fish", "NuggetB dog fish",
			// cat is not in -focus
			"SprocketA cat", "SprocketB cat", "WidgetA cat", "WidgetB cat", "More WidgetB cat",
			// fish is in -focus but cat is in -skip
			"SprocketA cat fish", "SprocketB cat fish", "WidgetA cat fish", "WidgetB cat fish", "More WidgetB cat fish",
			// Tests with Feature:Alpha
			"WidgetB fish",
			// Tests labelled 'slow'
			"WidgetB dog",
			"SprocketB fish",
			// _c is in -skip-file
			"SprocketC cat", "SprocketC dog", "SprocketC cat fish", "SprocketC dog fish",
			// lines in widget > 24 are not in --focus-file
			"More WidgetA cat", "More WidgetA dog", "More WidgetA cat fish", "More WidgetA dog fish",
		}
		pendingSpecs := []string{
			"SprocketA pending dog",
		}

		Ω(specs).Should(HaveLen(len(passedSpecs) + len(skippedSpecs) + len(pendingSpecs)))

		for _, text := range passedSpecs {
			Ω(specs.FindByFullText(text)).Should(HavePassed(), text)
		}
		for _, text := range skippedSpecs {
			Ω(specs.FindByFullText(text)).Should(HaveBeenSkipped(), text)
		}
		for _, text := range pendingSpecs {
			Ω(specs.FindByFullText(text)).Should(BePending(), text)
		}
	})

	It("ignores empty filter flags", func() {
		session := startGinkgo(fm.PathTo("filter"),
			"--focus=", "--skip=",
			"--json-report=report.json",
		)
		Eventually(session).Should(gexec.Exit(0))
		specs := Reports(fm.LoadJSONReports("filter", "report.json")[0].SpecReports)
		for _, spec := range specs {
			Ω(spec).Should(SatisfyAny(HavePassed(), BePending()))
		}
	})

	It("errors if the file-filter format is wrong", func() {
		session := startGinkgo(fm.PathTo("filter"), "--focus-file=foo:bar", "--skip-file=")
		Eventually(session).Should(gexec.Exit(1))
		Ω(session).Should(gbytes.Say("Invalid File Filter"))
		Ω(session).Should(gbytes.Say("Invalid File Filter"))
	})

	Describe("Listing labels", func() {
		BeforeEach(func() {
			fm.MountFixture("labels")
		})

		It("can list labels", func() {
			session := startGinkgo(fm.TmpDir, "labels", "-r")
			Eventually(session).Should(gexec.Exit(0))
			Ω(session).Should(gbytes.Say(`filter: \["Feature:Alpha", "Feature:Beta", "TopLevelLabel", "slow"\]`))
			Ω(session).Should(gbytes.Say(`labels: \["beluga", "bird", "cat", "chicken", "cow", "dog", "giraffe", "koala", "monkey", "otter", "owl", "panda"\]`))
			Ω(session).Should(gbytes.Say(`nolabels: No labels found`))
			Ω(session).Should(gbytes.Say(`onepkg: \["beluga", "bird", "cat", "chicken", "cow", "dog", "giraffe", "koala", "monkey", "otter", "owl", "panda"\]`))
		})
	})

	Describe("Semantic Version Filtering", func() {
		BeforeEach(func() {
			fm.MountFixture("semver")
		})

		It("filters specs based on semantic version constraints", func() {
			session := startGinkgo(filepath.Join(fm.TmpDir, "semver"),
				"--sem-ver-filter=2.2.0",
				"--json-report=report.json",
			)
			Eventually(session).Should(gexec.Exit(0))
			specs := Reports(fm.LoadJSONReports("semver", "report.json")[0].SpecReports)
			passedSpecs := []string{
				"should run without constraints",
				"should run with version in range [2.0.0, ~)",
				"should run with version in range [2.0.0, 3.0.0)",
				"should run with version in range [2.0.0, 4.0.0)",
				"should run with version in range [2.0.0, 5.0.0)",
				"should inherit container constraint",
				"should narrow down the constraint",
				"shouldn't expand the constraint",
				"should run without constraints by table driven",
				"should run with version in range [2.0.0, ~) by table driven",
			}
			skippedSpecs := []string{
				"shouldn't run with version in a conflict range",
				"shouldn't combine with a conflict constraint",
				"shouldn't run with version in a conflict range by table driven",
			}
			Ω(specs).Should(HaveLen(len(passedSpecs) + len(skippedSpecs)))
			for _, passed := range passedSpecs {
				Ω(specs.Find(passed)).Should(HavePassed())
			}
			for _, skipped := range skippedSpecs {
				Ω(specs.Find(skipped)).Should(HaveBeenSkipped())
			}
		})

		It("filters specs with hierarchy based on semantic version constraints", func() {
			session := startGinkgo(filepath.Join(fm.TmpDir, "semver", "spechierarchy"),
				"--sem-ver-filter=2.2.0",
				"--json-report=report.json",
			)
			Eventually(session).Should(gexec.Exit(0))
			specs := Reports(fm.LoadJSONReports(filepath.Join("semver", "spechierarchy"), "report.json")[0].SpecReports)
			passedSpecs := []string{
				"should inherit spec constraint",
			}
			skippedSpecs := []string{
				"should narrow down spec constraint",
			}
			Ω(specs).Should(HaveLen(len(passedSpecs) + len(skippedSpecs)))
			for _, passed := range passedSpecs {
				Ω(specs.Find(passed)).Should(HavePassed())
			}
			for _, skipped := range skippedSpecs {
				Ω(specs.Find(skipped)).Should(HaveBeenSkipped())
			}
		})
	})
})

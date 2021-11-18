package integration_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
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
			"--label-filter=TopLevelLabel && !SLOW",
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
		Eventually(session).Should(gexec.Exit(types.GINKGO_FOCUS_EXIT_CODE))
		specs := Reports(fm.LoadJSONReports("filter", "report.json")[0].SpecReports)
		for _, spec := range specs {
			if strings.HasPrefix(spec.FullText(), "SprocketC") {
				Ω(spec).Should(HavePassed())
			} else {
				Ω(spec).Should(Or(HaveBeenSkipped(), BePending()))
			}
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
			Ω(session).Should(gbytes.Say(`filter: \["TopLevelLabel", "slow"\]`))
			Ω(session).Should(gbytes.Say(`labels: \["beluga", "bird", "cat", "chicken", "cow", "dog", "giraffe", "koala", "monkey", "otter", "owl", "panda"\]`))
			Ω(session).Should(gbytes.Say(`nolabels: No labels found`))
			Ω(session).Should(gbytes.Say(`onepkg: \["beluga", "bird", "cat", "chicken", "cow", "dog", "giraffe", "koala", "monkey", "otter", "owl", "panda"\]`))
		})
	})
})

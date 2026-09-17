package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRelease(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Release Tool Suite")
}

var _ = Describe("the release tool", func() {
	const versionGo = "package types\n\nconst VERSION = \"2.32.2\"\n"

	// The released history below the Unreleased section - Prepare must never touch it.
	const history = "## 2.32.2\n\n### Fixes\n- a fix [abc1234]\n\n## 2.32.1\n\n### Fixes\n- another fix [def5678]\n\n"

	Describe("versions", func() {
		It("reads VERSION", func() {
			Expect(CurrentVersion(versionGo)).To(Equal("2.32.2"))
		})

		It("refuses a version.go without exactly one VERSION line", func() {
			_, err := CurrentVersion("package types\n")
			Expect(err).To(HaveOccurred())
			_, err = CurrentVersion(versionGo + versionGo)
			Expect(err).To(HaveOccurred())
		})

		It("bumps patch and minor", func() {
			Expect(NextVersion("2.32.2", "patch")).To(Equal("2.32.3"))
			Expect(NextVersion("2.32.2", "minor")).To(Equal("2.33.0"))
			Expect(NextVersion("2.9.9", "minor")).To(Equal("2.10.0"))
		})

		It("refuses anything else", func() {
			_, err := NextVersion("2.32.2", "major")
			Expect(err).To(HaveOccurred())
			_, err = NextVersion("2.32.2-rc.1", "patch")
			Expect(err).To(HaveOccurred())
		})

		It("rewrites only the VERSION line", func() {
			rewritten, err := SetVersion(versionGo, "2.33.0")
			Expect(err).NotTo(HaveOccurred())
			Expect(rewritten).To(Equal(strings.Replace(versionGo, `"2.32.2"`, `"2.33.0"`, 1)))
			Expect(CurrentVersion(rewritten)).To(Equal("2.33.0"))
		})
	})

	Describe("Prepare", func() {
		It("releases Unreleased as the new version under a fresh Unreleased, dropping empty subsections", func() {
			changelog := "## Unreleased\n\n### Features\n\n### Fixes\n\n- fixed a thing\n  across two lines\n\n### Maintenance\n\n" + history
			Expect(Prepare(changelog, "2.32.3")).To(Equal(
				freshUnreleased +
					"## 2.32.3\n\n### Fixes\n\n- fixed a thing\n  across two lines\n\n" +
					history))
		})

		It("keeps every non-empty subsection and entries outside any subsection", func() {
			changelog := "## Unreleased\n\n- loose entry\n\n### Features\n\n- a feature\n\n### Fixes\n\n- a fix\n\n" + history
			Expect(Prepare(changelog, "2.33.0")).To(Equal(
				freshUnreleased +
					"## 2.33.0\n\n- loose entry\n\n### Features\n\n- a feature\n\n### Fixes\n\n- a fix\n\n" +
					history))
		})

		It("treats a #### heading as content, not as an empty subsection", func() {
			changelog := "## Unreleased\n\n### Features\n\n#### The CLI\n\n### Fixes\n\n" + history
			Expect(Prepare(changelog, "2.33.0")).To(Equal(
				freshUnreleased +
					"## 2.33.0\n\n### Features\n\n#### The CLI\n\n" +
					history))
		})

		It("leaves text above Unreleased byte-for-byte alone", func() {
			changelog := "# Changelog\n\nSome preamble.\n\n## Unreleased\n\n- entry\n\n" + history
			prepared, err := Prepare(changelog, "2.32.3")
			Expect(err).NotTo(HaveOccurred())
			Expect(prepared).To(Equal("# Changelog\n\nSome preamble.\n\n" + freshUnreleased + "## 2.32.3\n\n- entry\n\n" + history))
		})

		It("keeps a blank line before the next release when the last entry had none", func() {
			changelog := "## Unreleased\n\n### Fixes\n\n- entry\n### Features\n" + history
			Expect(Prepare(changelog, "2.32.3")).To(Equal(
				freshUnreleased + "## 2.32.3\n\n### Fixes\n\n- entry\n\n" + history))
		})

		It("works when Unreleased is the only section", func() {
			Expect(Prepare("## Unreleased\n\n- entry\n", "2.33.0")).To(Equal(freshUnreleased + "## 2.33.0\n\n- entry\n"))
		})

		It("stops when Unreleased holds only blank lines and ### headings", func() {
			for _, unreleased := range []string{
				freshUnreleased,
				"## Unreleased\n\n",
				"## Unreleased\n",
				"## Unreleased\n### Features\n   \n### Fixes\n",
			} {
				_, err := Prepare(unreleased+history, "2.32.3")
				Expect(err).To(MatchError(ErrEmptyUnreleased), unreleased)
			}
		})

		It("leaves the real CHANGELOG.md's released history byte-for-byte alone", func() {
			onDisk, err := os.ReadFile(filepath.Join("..", "..", changelogPath))
			Expect(err).NotTo(HaveOccurred())
			_, _, history, err := splitSection(string(onDisk), "Unreleased")
			Expect(err).NotTo(HaveOccurred())
			Expect(history).To(HavePrefix("## "))

			changelog := "## Unreleased\n\n### Features\n\n- a feature\n\n### Fixes\n\n" + history
			prepared, err := Prepare(changelog, "9.9.9")
			Expect(err).NotTo(HaveOccurred())
			Expect(prepared).To(Equal(freshUnreleased + "## 9.9.9\n\n### Features\n\n- a feature\n\n" + history))
		})

		It("stops when there is no Unreleased section", func() {
			_, err := Prepare(history, "2.32.3")
			Expect(err).To(MatchError(ContainSubstring("no ## Unreleased section")))
		})

		It("stops when the version is already released", func() {
			_, err := Prepare("## Unreleased\n\n- entry\n\n"+history, "2.32.2")
			Expect(err).To(MatchError(ContainSubstring("already has a ## 2.32.2 section")))
		})
	})

	Describe("Notes", func() {
		It("returns the section body without its surrounding blank lines", func() {
			prepared, err := Prepare("## Unreleased\n\n### Fixes\n\n- a fix\n\n### Features\n\n"+history, "2.32.3")
			Expect(err).NotTo(HaveOccurred())
			Expect(Notes(prepared, "2.32.3")).To(Equal("### Fixes\n\n- a fix\n"))
		})

		It("fails for a missing or empty section", func() {
			_, err := Notes(history, "9.9.9")
			Expect(err).To(HaveOccurred())
			_, err = Notes("## Unreleased\n\n## 2.33.0\n\n## 2.32.2\n\n- x\n", "2.33.0")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("stamping the plugin manifests", func() {
		It("rewrites only the version field", func() {
			manifest := "{\n  \"name\": \"ginkgo\",\n  \"version\": \"2.32.2\",\n  \"license\": \"MIT\"\n}\n"
			stamped, ok := StampManifest(manifest, "2.33.0")
			Expect(ok).To(BeTrue())
			Expect(stamped).To(Equal(strings.Replace(manifest, `"2.32.2"`, `"2.33.0"`, 1)))
		})

		It("leaves a manifest with no version field alone - it tracks the commit SHA", func() {
			manifest := "{\n  \"name\": \"ginkgo\"\n}\n"
			stamped, ok := StampManifest(manifest, "2.33.0")
			Expect(ok).To(BeFalse())
			Expect(stamped).To(Equal(manifest))
		})

		It("finds every plugin manifest this repo ships, and no other json", func() {
			manifests, err := FindPluginManifests(filepath.Join("..", ".."))
			Expect(err).NotTo(HaveOccurred())
			Expect(manifests).To(ContainElement(filepath.Join("..", "..", "plugins", "ginkgo", ".claude-plugin", "plugin.json")))
			for _, manifest := range manifests {
				Expect(filepath.Base(manifest)).To(Equal("plugin.json"))
			}
		})

		It("keeps every shipped manifest's version in lockstep with VERSION", func() {
			manifests, err := FindPluginManifests(filepath.Join("..", ".."))
			Expect(err).NotTo(HaveOccurred())
			Expect(manifests).NotTo(BeEmpty())

			source, err := os.ReadFile(filepath.Join("..", "..", versionPath))
			Expect(err).NotTo(HaveOccurred())
			version, err := CurrentVersion(string(source))
			Expect(err).NotTo(HaveOccurred())

			for _, manifest := range manifests {
				contents, err := os.ReadFile(manifest)
				Expect(err).NotTo(HaveOccurred())
				if _, ok := StampManifest(string(contents), version); !ok {
					continue
				}
				Expect(string(contents)).To(ContainSubstring(`"version": %q`, version), manifest)
			}
		})
	})
})

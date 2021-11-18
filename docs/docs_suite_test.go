package docs_test

import (
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal/test_helpers"
	. "github.com/onsi/gomega"
)

func TestDocs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Docs Suite")
}

var anchors test_helpers.Anchors

var _ = BeforeSuite(func() {
	var err error
	anchors, err = test_helpers.LoadAnchors(test_helpers.DOCS, "../")
	Ω(err).ShouldNot(HaveOccurred())
})

var _ = Describe("Validating internal links", func() {
	var entries = []TableEntry{
		Entry("Narrative Documentation", "index.md"),
		Entry("V2 Migration Documentation", "MIGRATING_TO_V2.md"),
		Entry("Repo Readme", "README.md"),
	}

	DescribeTable("Ensuring no headings have any markdown formatting characters in them", func(name string) {
		headings, err := test_helpers.LoadMarkdownHeadings(test_helpers.DOCS.DocWithName(name).Path("../"))
		Ω(err).ShouldNot(HaveOccurred())
		failed := false
		for _, heading := range headings {
			if strings.ContainsAny(heading, "`*_~#") {
				failed = true
				GinkgoWriter.Printf("%s: '%s'\n", name, heading)
			}
		}
		if failed {
			Fail("Identified invalid headings")
		}
	}, entries)

	DescribeTable("Ensuring all anchors resolve", func(name string) {
		links, err := test_helpers.LoadMarkdownLinks(test_helpers.DOCS.DocWithName(name).Path("../"))
		Ω(err).ShouldNot(HaveOccurred())
		Ω(links).ShouldNot(BeEmpty())
		failed := false
		for _, link := range links {
			if !anchors.IsResolvable(name, link) {
				failed = true
				GinkgoWriter.Printf("%s: '%s'\n", name, link)
			}
		}
		if failed {
			Fail("Identified invalid links")
		}
	}, entries)
})

var _ = Describe("Validating godoc links", func() {
	It("validates that all links in the core dsl package are good", func() {
		fset := token.NewFileSet()
		entries, err := os.ReadDir("../")
		Ω(err).ShouldNot(HaveOccurred())
		parsedFiles := []*ast.File{}
		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasSuffix(name, ".go") {
				continue
			}
			parsed, err := parser.ParseFile(fset, filepath.Join("../", name), nil, parser.ParseComments)
			Ω(err).ShouldNot(HaveOccurred())
			parsedFiles = append(parsedFiles, parsed)
		}

		p, err := doc.NewFromFiles(fset, parsedFiles, "github.com/onsi/ginkgo/v2")
		Ω(err).ShouldNot(HaveOccurred())

		var b strings.Builder
		b.WriteString(p.Doc)
		b.WriteString("\n")
		for _, elem := range p.Consts {
			b.WriteString(elem.Doc)
			b.WriteString("\n")
		}
		for _, elem := range p.Types {
			b.WriteString(elem.Doc)
			b.WriteString("\n")
		}
		for _, elem := range p.Vars {
			b.WriteString(elem.Doc)
			b.WriteString("\n")
		}
		for _, elem := range p.Funcs {
			b.WriteString(elem.Doc)
			b.WriteString("\n")
		}

		doc := b.String()
		urlRegexp := regexp.MustCompile(`https*[\w:/#\-\.]*`)
		links := urlRegexp.FindAllString(doc, -1)

		failed := false
		for _, link := range links {
			if !anchors.IsResolvable("", link) {
				failed = true
				GinkgoWriter.Printf("Godoc: '%s'\n", link)
			}
		}
		if failed {
			Fail("Identified invalid links")
		}
	})
})

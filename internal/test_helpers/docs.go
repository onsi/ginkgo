package test_helpers

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var punctuationRE = regexp.MustCompile(`[^\w\-\s]`)
var linkRE = regexp.MustCompile(`\]\(([\w:/#\-\.]*)\)`)

type Doc struct {
	Name string
	URLs []string
	path string
}

func (doc Doc) Path(root string) string {
	return filepath.Join(root, doc.path)
}

type Docs []Doc

func (d Docs) DocWithName(name string) Doc {
	for _, doc := range d {
		if doc.Name == name {
			return doc
		}
	}
	return Doc{}
}

func (d Docs) DocWithURL(url string) Doc {
	for _, doc := range d {
		for _, u := range doc.URLs {
			if u == url {
				return doc
			}
		}
	}
	return Doc{}
}

var DOCS = Docs{
	{"index.md", []string{"https://onsi.github.io/ginkgo/"}, "docs/index.md"},
	{"MIGRATING_TO_V2.md", []string{"https://onsi.github.io/ginkgo/MIGRATING_TO_V2"}, "docs/MIGRATING_TO_V2.md"},
	{"README.md", []string{"https://github.com/onsi/ginkgo", "https://github.com/onsi/ginkgo/blob/master/README.md"}, "README.md"},
}

type Anchors struct {
	Docs       Docs
	DocAnchors map[string][]string
}

func (a Anchors) IsResolvable(docName string, link string) bool {
	var anchorSet []string
	var expectedAnchor string
	if strings.HasPrefix(link, "#") {
		anchorSet = a.DocAnchors[docName]
		expectedAnchor = strings.TrimPrefix(link, "#")
	} else {
		components := strings.Split(link, "#")
		doc := a.Docs.DocWithURL(components[0])
		if doc.Name == "" {
			//allow external links
			return true
		}
		if len(components) == 1 {
			//allow links to the doc with no anchor
			return true
		}
		expectedAnchor = components[1]
		anchorSet = a.DocAnchors[doc.Name]
	}

	for _, anchor := range anchorSet {
		if anchor == expectedAnchor {
			return true
		}
	}

	return false
}

func LoadAnchors(docs Docs, rootPath string) (Anchors, error) {
	out := Anchors{
		Docs:       docs,
		DocAnchors: map[string][]string{},
	}
	for _, doc := range docs {
		anchors, err := loadMarkdownHeadingAnchors(doc.Path(rootPath))
		if err != nil {
			return Anchors{}, err
		}
		out.DocAnchors[doc.Name] = anchors
	}
	return out, nil
}

func loadMarkdownHeadingAnchors(filename string) ([]string, error) {
	headings, err := LoadMarkdownHeadings(filename)
	if err != nil {
		return nil, err
	}

	var anchors []string
	for _, heading := range headings {
		heading = punctuationRE.ReplaceAllString(heading, "")
		heading = strings.ToLower(heading)
		heading = strings.ReplaceAll(heading, " ", "-")
		anchors = append(anchors, heading)
	}

	return anchors, nil
}

func LoadMarkdownHeadings(filename string) ([]string, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	headings := []string{}
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		line = strings.TrimLeft(line, " ")
		line = strings.TrimRight(line, " ")
		if !strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimLeft(line, "# ")
		headings = append(headings, line)
	}
	return headings, nil
}

func LoadMarkdownLinks(filename string) ([]string, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	links := []string{}
	lines := strings.Split(string(b), "\n")

	for _, line := range lines {
		matches := linkRE.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if match[1] != "" {
				links = append(links, match[1])
			}
		}
	}

	return links, nil
}

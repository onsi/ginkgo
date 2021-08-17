package test_helpers

import (
	"os"
	"regexp"
	"strings"
)

var punctuationRE = regexp.MustCompile(`[^\w\-\s]`)

func LoadMarkdownHeadingAnchors(filename string) []string {
	b, err := os.ReadFile(filename)
	if err != nil {
		return []string{}
	}

	var anchors []string
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		line = strings.TrimLeft(line, " ")
		line = strings.TrimRight(line, " ")
		if !strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimLeft(line, "# ")
		line = punctuationRE.ReplaceAllString(line, "")
		line = strings.ToLower(line)
		line = strings.ReplaceAll(line, " ", "-")
		anchors = append(anchors, line)
	}

	return anchors
}

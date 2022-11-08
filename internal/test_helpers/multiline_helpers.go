package test_helpers

import (
	"fmt"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gcustom"
)

func MultilineTextHelper(s string) string {
	lines := strings.Split(s, "\n")
	out := "\nstrings.Join([]string{\n"
	for _, l := range lines {
		out = out + fmt.Sprintf("    %#v,\n", l)
	}
	out += `}, "\n")`
	return out
}

func MatchLines(expected ...any) gcustom.CustomGomegaMatcher {
	data := map[string]any{}
	return gcustom.MakeMatcher(func(actual string) (bool, error) {
		data["RenderedActual"] = MultilineTextHelper(actual)
		if len(expected) == 0 && len(actual) == 0 {
			return true, nil
		}
		lines := strings.Split(actual, "\n")
		for idx, expectation := range expected {
			if idx >= len(lines) {
				data["Failure"] = "More Expectations than lines..."
				return false, nil
			}
			var matcher OmegaMatcher
			if expectedString, isString := expectation.(string); isString {
				matcher = Equal(expectedString)
			} else {
				matcher = expectation.(OmegaMatcher)
			}
			matches, err := matcher.Match(lines[idx])
			if err != nil {
				data["Failure"] = fmt.Sprintf("At line %d:\n%s", idx+1, err.Error())
				return false, nil
			}
			if !matches {
				data["Failure"] = fmt.Sprintf("At line %d:\n%s", idx+1, matcher.FailureMessage(lines[idx]))
				return false, nil
			}
		}
		if len(expected) < len(lines) {
			data["Failure"] = fmt.Sprintf("Missing expectations for last %d lines", len(lines)-len(expected))
			return false, nil
		}
		return true, nil
	}).WithTemplate("Failed {{.To}} MatchLines for:\n{{.Data.RenderedActual}}\n\n{{.Data.Failure}}").WithTemplateData(data)
}

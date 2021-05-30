package test_helpers

import (
	"fmt"
	"strings"
)

func MultilneTextHelper(s string) string {
	lines := strings.Split(s, "\n")
	out := "\nstrings.Join([]string{\n"
	for _, l := range lines {
		out = out + fmt.Sprintf("    %#v,\n", l)
	}
	out += `}, "\n")`
	return out
}

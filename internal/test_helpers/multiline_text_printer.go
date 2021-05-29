package test_helpers

import (
	"fmt"
	"strings"
)

func PrintMultilneText(s string) {
	lines := strings.Split(s, "\n")
	out := "\nstrings.Join([]string{\n"
	for _, l := range lines {
		out = out + fmt.Sprintf("    %#v,\n", l)
	}
	out += `}, "\n")`
	fmt.Println(out)
}

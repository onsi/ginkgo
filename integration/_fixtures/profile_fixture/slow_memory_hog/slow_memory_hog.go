package slow_memory_hog

import "strings"

func SomethingExpensive(n int) string {
	pirate := []string{"r"}
	for i := 0; i < n; i++ {
		pirate = doubleArrrr(pirate)
	}
	return strings.Join(pirate, "\n")
}

func doubleArrrr(in []string) []string {
	out := make([]string, len(in)*2)
	for j := 0; j < len(in); j++ {
		out[j] = in[j]
		out[len(in)+j] = in[j]
	}
	return out
}

package main

import (
	"fmt"
	"os/exec"
)

func unfocusSpecs() {
	unfocus("Describe")
	unfocus("Context")
	unfocus("It")
	unfocus("Measure")
}

func unfocus(component string) {
	fmt.Printf("Removing F%s...\n", component)
	cmd := exec.Command("gofmt", fmt.Sprintf("-r=F%s -> %s", component, component), "-w", ".")
	out, _ := cmd.CombinedOutput()
	if string(out) != "" {
		println(string(out))
	}
}

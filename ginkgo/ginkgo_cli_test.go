package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Command Documentation Links", func() {
	It("all commands with doc links point to valid documentation", func() {
		commands := GenerateCommands()
		for _, command := range commands {
			if command.DocLink != "" {
				Ω(anchors.DocAnchors["index.md"]).Should(ContainElement(command.DocLink))
			}
		}
	})
})

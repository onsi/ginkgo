package internal

import (
	"strings"

	"github.com/onsi/ginkgo/types"
)

type Spec struct {
	Nodes Nodes
	Skip  bool
}

func (s Spec) Text() string {
	texts := []string{}
	for _, node := range s.Nodes {
		if node.Text != "" {
			texts = append(texts, node.Text)
		}
	}
	return strings.Join(texts, " ")
}

func (s Spec) FirstNodeWithType(nodeTypes ...types.NodeType) Node {
	return s.Nodes.FirstNodeWithType(nodeTypes...)
}

func (s Spec) FlakeAttempts() int {
	flakeAttempts := 0
	for _, node := range s.Nodes {
		if node.FlakeAttempts > 0 {
			flakeAttempts = node.FlakeAttempts
		}
	}

	return flakeAttempts
}

type Specs []Spec

func (s Specs) HasAnySpecsMarkedPending() bool {
	for _, spec := range s {
		if spec.Nodes.HasNodeMarkedPending() {
			return true
		}
	}

	return false
}

func (s Specs) CountWithoutSkip() int {
	n := 0
	for _, spec := range s {
		if !spec.Skip {
			n += 1
		}
	}
	return n
}

package internal

import (
	"strings"

	"github.com/onsi/ginkgo/types"
)

type Spec struct {
	Nodes Nodes
	Skip  bool
}

func (t Spec) Text() string {
	texts := []string{}
	for _, node := range t.Nodes {
		if node.Text != "" {
			texts = append(texts, node.Text)
		}
	}
	return strings.Join(texts, " ")
}

func (t Spec) FirstNodeWithType(nodeTypes ...types.NodeType) Node {
	return t.Nodes.FirstNodeWithType(nodeTypes...)
}

type Specs []Spec

func (t Specs) HasAnySpecsMarkedPending() bool {
	for _, test := range t {
		if test.Nodes.HasNodeMarkedPending() {
			return true
		}
	}

	return false
}

func (t Specs) CountWithoutSkip() int {
	n := 0
	for _, test := range t {
		if !test.Skip {
			n += 1
		}
	}
	return n
}

func (t Specs) PartitionByFirstNodeWithType(nodeTypes ...types.NodeType) []Specs {
	indexById := map[uint]int{}
	partition := []Specs{}
	for _, test := range t {
		id := test.FirstNodeWithType(nodeTypes...).ID
		if id == 0 {
			continue
		}
		idx, found := indexById[id]
		if !found {
			partition = append(partition, Specs{})
			idx = len(partition) - 1
			indexById[id] = idx
		}
		partition[idx] = append(partition[idx], test)
	}

	return partition
}

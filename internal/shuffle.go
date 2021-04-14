package internal

import (
	"math/rand"
	"sort"

	"github.com/onsi/ginkgo/types"
)

func ShuffleSpecs(specs Specs, suiteConfig types.SuiteConfig) Specs {
	/*
		Ginkgo has sophisticated suport for randomizing specs.  Specs are guaranteed to have the same
		order for a given seed across test runs.

		By default only top-level containers and specs are shuffled - this makes for a more intuitive debugging
		experience - specs within a given container run in the order they appear in the file.

		Developers can set -randomizeAllSpecs to shuffle _all_ specs.
	*/
	r := rand.New(rand.NewSource(suiteConfig.RandomSeed))

	// We shuffle by partitioning specs by the id of the first node of a given type, then shuffling that partition
	// When -randomizeAllSpecs is set we partition the specs by the id of the `It` - i.e. each spec has a unique id and this is equivalent to sorting all specs
	// Otherwise, we partition by the id of the first container (or It if there is no container).  This preserves the spec grouping by top-level container.
	nodeTypesToShuffle := types.NodeTypesForContainerAndIt
	if suiteConfig.RandomizeAllSpecs {
		nodeTypesToShuffle = []types.NodeType{types.NodeTypeIt}
	}

	partition := specs.PartitionByFirstNodeWithType(nodeTypesToShuffle...)

	// To ensure a consistent shuffle, the partition slice must have the same order across spec runs.
	// To do this we stable sort by codelocation of the node of type that was used to form the partition
	sort.SliceStable(partition, func(i, j int) bool {
		a := partition[i][0].FirstNodeWithType(nodeTypesToShuffle...)
		b := partition[j][0].FirstNodeWithType(nodeTypesToShuffle...)
		return a.CodeLocation.String() < b.CodeLocation.String()
	})

	//Shuffle the partition
	shuffledSpecs := Specs{}
	permutation := r.Perm(len(partition))
	for _, j := range permutation {
		shuffledSpecs = append(shuffledSpecs, partition[j]...)
	}

	return shuffledSpecs
}

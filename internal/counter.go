package internal

import (
	"github.com/onsi/ginkgo/internal/parallel_support"
	"github.com/onsi/ginkgo/types"
)

func MakeNextIndexCounter(suiteConfig types.SuiteConfig) func() (int, error) {
	if suiteConfig.ParallelTotal > 1 {
		client := parallel_support.NewClient(suiteConfig.ParallelHost)
		return func() (int, error) {
			return client.FetchNextCounter()
		}
	} else {
		return MakeIncrementingIndexCounter()
	}
}

func MakeIncrementingIndexCounter() func() (int, error) {
	idx := -1
	return func() (int, error) {
		idx += 1
		return idx, nil
	}
}

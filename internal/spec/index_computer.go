package spec

import (
	"math"
)

func ParallelizedIndexRange(length int, parallelTotal int, parallelNode int) (startIndex int, count int) {
	if length == 0 {
		return 0, 0
	}

	count = int(math.Floor((float64(length) / float64(parallelTotal)) + 0.5))
	if count == 0 {
		count = 1
	}

	startIndex = (parallelNode - 1) * count
	if startIndex >= length {
		startIndex = length
		count = 0
	}

	if parallelNode == parallelTotal {
		count = length - startIndex
	}

	return
}

package ginkgo

func parallelizedIndexRange(length int, parallelTotal int, parallelNode int) (startIndex int, count int) {
	count = length / parallelTotal
	if count < 1 {
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

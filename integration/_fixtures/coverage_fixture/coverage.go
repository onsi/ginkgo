package coverage_fixture

import (
	_ "github.com/onsi/ginkgo/v2/integration/_fixtures/coverage_fixture/external_coverage"
)

func A() string {
	return "A"
}

func B() string {
	return "B"
}

func C() string {
	return "C"
}

func D() string {
	return "D"
}

func E() string {
	return "tested by additional"
}

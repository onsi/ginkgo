package types

import (
	"time"
)

type Benchmarker interface {
	Time(name string, body func(), info ...interface{}) (elapsedTime time.Duration)
	RecordValue(name string, value float64, info ...interface{})
}

type GinkgoTestDescription struct {
	ComponentTexts []string
	FullTestText   string
	TestText       string

	IsMeasurement bool

	FileName   string
	LineNumber int
}

type GinkgoTestingT interface {
	Fail()
}

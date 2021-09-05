package nodot_test

import (
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

func TestNodot(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Nodot Suite")
}

// Declarations for Ginkgo DSL
type (
	Done        ginkgo.Done
	Benchmarker ginkgo.Benchmarker
)

var (
	GinkgoWriter                          = ginkgo.GinkgoWriter
	GinkgoParallelNode                    = ginkgo.GinkgoParallelNode
	GinkgoT                               = ginkgo.GinkgoT
	CurrentGinkgoTestDescription          = ginkgo.CurrentGinkgoTestDescription
	RunSpecs                              = ginkgo.RunSpecs
	RunSpecsWithDefaultAndCustomReporters = ginkgo.RunSpecsWithDefaultAndCustomReporters
	RunSpecsWithCustomReporters           = ginkgo.RunSpecsWithCustomReporters
	Fail                                  = ginkgo.Fail
	GinkgoRecover                         = ginkgo.GinkgoRecover
	Describe                              = ginkgo.Describe
	FDescribe                             = ginkgo.FDescribe
	PDescribe                             = ginkgo.PDescribe
	XDescribe                             = ginkgo.XDescribe
	Context                               = ginkgo.Context
	FContext                              = ginkgo.FContext
	PContext                              = ginkgo.PContext
	XContext                              = ginkgo.XContext
	It                                    = ginkgo.It
	FIt                                   = ginkgo.FIt
	PIt                                   = ginkgo.PIt
	XIt                                   = ginkgo.XIt
	Measure                               = ginkgo.Measure
	FMeasure                              = ginkgo.FMeasure
	PMeasure                              = ginkgo.PMeasure
	XMeasure                              = ginkgo.XMeasure
	BeforeSuite                           = ginkgo.BeforeSuite
	AfterSuite                            = ginkgo.AfterSuite
	SynchronizedBeforeSuite               = ginkgo.SynchronizedBeforeSuite
	SynchronizedAfterSuite                = ginkgo.SynchronizedAfterSuite
	BeforeEach                            = ginkgo.BeforeEach
	JustBeforeEach                        = ginkgo.JustBeforeEach
	JustAfterEach                         = ginkgo.JustAfterEach
	AfterEach                             = ginkgo.AfterEach
)

// Declarations for Gomega DSL
var (
	RegisterFailHandler                   = gomega.RegisterFailHandler
	RegisterTestingT                      = gomega.RegisterTestingT
	InterceptGomegaFailures               = gomega.InterceptGomegaFailures
	Ω                                     = gomega.Ω
	Expect                                = gomega.Expect
	ExpectWithOffset                      = gomega.ExpectWithOffset
	Eventually                            = gomega.Eventually
	EventuallyWithOffset                  = gomega.EventuallyWithOffset
	Consistently                          = gomega.Consistently
	ConsistentlyWithOffset                = gomega.ConsistentlyWithOffset
	SetDefaultEventuallyTimeout           = gomega.SetDefaultEventuallyTimeout
	SetDefaultEventuallyPollingInterval   = gomega.SetDefaultEventuallyPollingInterval
	SetDefaultConsistentlyDuration        = gomega.SetDefaultConsistentlyDuration
	SetDefaultConsistentlyPollingInterval = gomega.SetDefaultConsistentlyPollingInterval
)

// Declarations for Gomega Matchers
var (
	Equal                = gomega.Equal
	BeEquivalentTo       = gomega.BeEquivalentTo
	BeNil                = gomega.BeNil
	BeTrue               = gomega.BeTrue
	BeFalse              = gomega.BeFalse
	HaveOccurred         = gomega.HaveOccurred
	MatchError           = gomega.MatchError
	BeClosed             = gomega.BeClosed
	Receive              = gomega.Receive
	MatchRegexp          = gomega.MatchRegexp
	ContainSubstring     = gomega.ContainSubstring
	MatchJSON            = gomega.MatchJSON
	BeEmpty              = gomega.BeEmpty
	HaveLen              = gomega.HaveLen
	BeZero               = gomega.BeZero
	ContainElement       = gomega.ContainElement
	ConsistOf            = gomega.ConsistOf
	HaveKey              = gomega.HaveKey
	HaveKeyWithValue     = gomega.HaveKeyWithValue
	BeNumerically        = gomega.BeNumerically
	BeTemporally         = gomega.BeTemporally
	BeAssignableToTypeOf = gomega.BeAssignableToTypeOf
	Panic                = gomega.Panic
)

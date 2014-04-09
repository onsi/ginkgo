package leafnodes

import (
	"encoding/json"
	"github.com/onsi/ginkgo/internal/failer"
	"github.com/onsi/ginkgo/types"
	"io/ioutil"
	"net/http"
	"time"
)

type compoundAfterSuiteNode struct {
	runnerA *runner
	runnerB *runner

	ginkgoNode       int
	totalGinkgoNodes int
	syncHost         string

	outcome types.SpecState
	failure types.SpecFailure
	runTime time.Duration
}

func NewCompoundAfterSuiteNode(bodyA interface{}, bodyB interface{}, codeLocation types.CodeLocation, timeout time.Duration, failer *failer.Failer, ginkgoNode int, totalGinkgoNodes int, syncHost string) SuiteNode {
	return &compoundAfterSuiteNode{
		runnerA:          newRunner(bodyA, codeLocation, timeout, failer, types.SpecComponentTypeAfterSuite, 0),
		runnerB:          newRunner(bodyB, codeLocation, timeout, failer, types.SpecComponentTypeAfterSuite, 0),
		ginkgoNode:       ginkgoNode,
		totalGinkgoNodes: totalGinkgoNodes,
		syncHost:         syncHost,
	}
}

func (node *compoundAfterSuiteNode) Run() bool {
	node.outcome, node.failure = node.runnerA.run()

	if node.ginkgoNode == 1 {
		if node.totalGinkgoNodes > 1 {
			node.waitUntilOtherNodesAreDone()
		}

		outcome, failure := node.runnerB.run()

		if node.outcome == types.SpecStatePassed {
			node.outcome, node.failure = outcome, failure
		}
	}

	return node.outcome == types.SpecStatePassed
}

func (node *compoundAfterSuiteNode) Passed() bool {
	return node.outcome == types.SpecStatePassed
}

func (node *compoundAfterSuiteNode) Summary() *types.SetupSummary {
	return &types.SetupSummary{
		ComponentType: node.runnerA.nodeType,
		CodeLocation:  node.runnerA.codeLocation,
		State:         node.outcome,
		RunTime:       node.runTime,
		Failure:       node.failure,
	}
}

func (node *compoundAfterSuiteNode) waitUntilOtherNodesAreDone() {
	for {
		if node.canRun() {
			return
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func (node *compoundAfterSuiteNode) canRun() bool {
	resp, err := http.Get(node.syncHost + "/AfterSuiteCanRun")
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	resp.Body.Close()

	r := AfterSuiteCanRun{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return false
	}

	return r.CanRun
}

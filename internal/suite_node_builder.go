package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

var SUITE_NODE_POLLING_INTERVAL = 50 * time.Millisecond

/*
The family of SuiteNodes - i.e. BeforeSuite, AfterSuite, SynchronizedBeforeSuite, and SynchronizedAfterSuite
have a set of properties and behaviors that cannot be determined until PhaseRun.

Specifically, the Synchronized* family need information about parallelism (config.ParallelTotal, config.ParallelNode, config.ParallelHost)
that can only be obtained after the config object has been populated and this only happens once the test suite has begun.
Of course the suite DSl functions are typicaly called in PhaseBuildTopLevel.  Given that, we use
this factory approach to save off information about the suite node at PhaseBuildTopLevel and then construct
the actual nodes when we have the config information in PhaseRun.

Note that SuiteNodeBuilder takes care of packaging all the parallelisum behavior into the single node.Body().  This pushes
all the complexity of the suite node's behavior into this builder class.
*/
type SuiteNodeBuilder struct {
	NodeType     types.NodeType
	CodeLocation types.CodeLocation

	BeforeSuiteBody                     func()
	SynchronizedBeforeSuiteNode1Body    func() []byte
	SynchronizedBeforeSuiteAllNodesBody func([]byte)

	AfterSuiteBody                     func()
	SynchronizedAfterSuiteAllNodesBody func()
	SynchronizedAfterSuiteNode1Body    func()
}

func (s SuiteNodeBuilder) BuildNode(config config.GinkgoConfigType, failer *Failer) Node {
	node := Node{
		ID:           UniqueNodeID(),
		NodeType:     s.NodeType,
		CodeLocation: s.CodeLocation,
	}
	switch s.NodeType {
	case types.NodeTypeBeforeSuite:
		node.Body = s.BeforeSuiteBody
	case types.NodeTypeSynchronizedBeforeSuite:
		node.Body = s.buildSynchronizedBeforeSuiteBody(config, failer)
	case types.NodeTypeAfterSuite:
		node.Body = s.AfterSuiteBody
	case types.NodeTypeSynchronizedAfterSuite:
		node.Body = s.buildSynchronizedAfterSuiteBody(config, failer)
	default:
		return Node{}
	}

	return node
}

func (s SuiteNodeBuilder) buildSynchronizedBeforeSuiteBody(config config.GinkgoConfigType, failer *Failer) func() {
	if config.ParallelTotal == 1 {
		return func() {
			data := s.SynchronizedBeforeSuiteNode1Body()
			s.SynchronizedBeforeSuiteAllNodesBody(data)
		}
	}

	if config.ParallelNode == 1 {
		return func() {
			result := func() (result types.RemoteBeforeSuiteData) {
				defer func() {
					if e := recover(); e != nil {
						failer.Panic(types.NewCodeLocation(2), e)
					}
				}()

				result.State = types.RemoteBeforeSuiteStateFailed
				result.Data = s.SynchronizedBeforeSuiteNode1Body()
				result.State = types.RemoteBeforeSuiteStatePassed
				return
			}()

			resp, err := http.Post(config.ParallelHost+"/BeforeSuiteState", "application/json", bytes.NewBuffer(result.ToJSON()))
			if err != nil || resp.StatusCode != http.StatusOK {
				failer.Fail("SynchronizedBeforeSuite failed to send data to other nodes", s.CodeLocation)
				return
			}
			resp.Body.Close()

			if result.State == types.RemoteBeforeSuiteStatePassed {
				s.SynchronizedBeforeSuiteAllNodesBody(result.Data)
			}
		}
	} else {
		return func() {
			var result types.RemoteBeforeSuiteData
			for {
				result = types.RemoteBeforeSuiteData{}
				err := s.pollEndpoint(config.ParallelHost+"/BeforeSuiteState", &result)
				if err != nil {
					failer.Fail("SynchronizedBeforeSuite Server Communication Issue:\n"+err.Error(), s.CodeLocation)
					return
				}
				if result.State != types.RemoteBeforeSuiteStatePending {
					break
				}
				time.Sleep(SUITE_NODE_POLLING_INTERVAL)
			}

			switch result.State {
			case types.RemoteBeforeSuiteStatePassed:
				s.SynchronizedBeforeSuiteAllNodesBody(result.Data)
			case types.RemoteBeforeSuiteStateFailed:
				failer.Fail("SynchronizedBeforeSuite on Node 1 failed", s.CodeLocation)
			case types.RemoteBeforeSuiteStateDisappeared:
				failer.Fail("SynchronizedBeforeSuite on Node 1 disappeared before it could report back", s.CodeLocation)
			}
		}
	}
}

func (s SuiteNodeBuilder) buildSynchronizedAfterSuiteBody(config config.GinkgoConfigType, failer *Failer) func() {
	if config.ParallelTotal == 1 {
		return func() {
			s.SynchronizedAfterSuiteAllNodesBody()
			s.SynchronizedAfterSuiteNode1Body()
		}
	}

	if config.ParallelNode > 1 {
		return func() {
			s.SynchronizedAfterSuiteAllNodesBody()
		}
	} else {
		return func() {
			s.SynchronizedAfterSuiteAllNodesBody()

			for {
				afterSuiteData := types.RemoteAfterSuiteData{}
				err := s.pollEndpoint(config.ParallelHost+"/AfterSuiteState", &afterSuiteData)
				if err != nil {
					failer.Fail("SynchronizedAfterSuite Server Communication Issue:\n"+err.Error(), s.CodeLocation)
					break
				}
				if afterSuiteData.CanRun {
					break
				}
				time.Sleep(SUITE_NODE_POLLING_INTERVAL)
			}

			s.SynchronizedAfterSuiteNode1Body()
		}
	}
}

func (s SuiteNodeBuilder) pollEndpoint(endpoint string, data interface{}) error {
	resp, err := http.Get(endpoint)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	err = json.Unmarshal(body, data)
	if err != nil {
		return err
	}

	return nil
}

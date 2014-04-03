package main

import (
	"github.com/onsi/ginkgo/ginkgo/testrunner"
	"github.com/onsi/ginkgo/ginkgo/testsuite"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type TestRunnerManager struct {
	activeRunners []*testrunner.TestRunner
	commandFlags  *RunAndWatchCommandFlags
	lock          *sync.Mutex
}

func NewTestRunnerManager(commandFlags *RunAndWatchCommandFlags) *TestRunnerManager {
	manager := &TestRunnerManager{
		commandFlags: commandFlags,
		lock:         &sync.Mutex{},
	}
	manager.registerSignalHandler()
	return manager
}

func (m *TestRunnerManager) MakeAndRegisterTestRunner(suite *testsuite.TestSuite) *testrunner.TestRunner {
	m.lock.Lock()
	defer m.lock.Unlock()

	runner := testrunner.New(suite, m.commandFlags.NumCPU, m.commandFlags.ParallelStream, m.commandFlags.Race, m.commandFlags.Cover)
	m.activeRunners = append(m.activeRunners, runner)
	return runner
}

func (m *TestRunnerManager) UnregisterRunner(runner *testrunner.TestRunner) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for i, registeredRunner := range m.activeRunners {
		if registeredRunner == runner {
			m.activeRunners[i] = m.activeRunners[len(m.activeRunners)-1]
			m.activeRunners = m.activeRunners[0 : len(m.activeRunners)-1]
			break
		}
	}
}

func (m *TestRunnerManager) registerSignalHandler() {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

		select {
		case sig := <-c:
			m.lock.Lock()
			for _, runner := range m.activeRunners {
				runner.CleanUp(sig)
			}
			m.lock.Unlock()
			os.Exit(1)
		}
	}()
}

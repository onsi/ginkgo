package test_helpers

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal/parallel_support"
	"github.com/onsi/ginkgo/v2/reporters"
	. "github.com/onsi/gomega"
)

func SetUpServerAndClient(numNodes int) (parallel_support.Server, parallel_support.Client, map[int]chan any) {
	server, err := parallel_support.NewServer(numNodes, reporters.NoopReporter{})
	Ω(err).ShouldNot(HaveOccurred())
	server.Start()
	client := parallel_support.NewClient(server.Address())
	Eventually(client.Connect).Should(BeTrue())

	exitChannels := map[int]chan any{}
	for node := 1; node <= numNodes; node++ {
		c := make(chan any)
		exitChannels[node] = c
		server.RegisterAlive(node, func() bool {
			select {
			case <-c:
				return false
			default:
				return true
			}
		})
	}

	DeferCleanup(server.Close)
	DeferCleanup(client.Close)

	return server, client, exitChannels
}

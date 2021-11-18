package block_contest_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/integration/_fixtures/profile_fixture/block_contest"
)

func TestBlockContest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BlockContest Suite")
}

var _ = Describe("quick channel reads", func() {
	for i := 0; i < 10; i++ {
		It("reads a channel quickly", func() {
			c := make(chan bool)
			go func() {
				block_contest.ReadTheChannel(c)
			}()
			time.Sleep(5 * time.Millisecond)
			c <- true
		})
	}
})

var _ = Describe("slow channel reads", func() {
	It("gets stuck for a bit", func() {
		c := make(chan bool)
		go func() {
			block_contest.SlowReadTheChannel(c)
		}()
		time.Sleep(500 * time.Millisecond)
		c <- true
	})
})

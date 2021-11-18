package lock_contest_test

import (
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/integration/_fixtures/profile_fixture/lock_contest"
)

func TestLockContest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LockContest Suite")
}

var _ = Describe("quick lock blocks", func() {
	for i := 0; i < 10; i++ {
		It("reads a lock quickly", func() {
			c := make(chan bool)
			l := &sync.Mutex{}
			l.Lock()
			go func() {
				lock_contest.WaitForLock(l)
				close(c)
			}()
			time.Sleep(5 * time.Millisecond)
			l.Unlock()
			<-c
		})
	}
})

var _ = Describe("slow lock block", func() {
	It("gets stuck for a bit", func() {
		c := make(chan bool)
		l := &sync.Mutex{}
		l.Lock()
		go func() {
			lock_contest.SlowWaitForLock(l)
			close(c)
		}()
		time.Sleep(500 * time.Millisecond)
		l.Unlock()
		<-c
	})
})

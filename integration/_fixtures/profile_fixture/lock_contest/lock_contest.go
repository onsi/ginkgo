package lock_contest

import (
	"sync"
)

func WaitForLock(l *sync.Mutex) {
	l.Lock()
	l.Unlock()
}

func SlowWaitForLock(l *sync.Mutex) {
	l.Lock()
	l.Unlock()
}

package util

import (
	"sync"
	"time"
)

var locks sync.Map

func TryLockFor(key any, dur time.Duration) bool {
	lock, ok := locks.Load(key)

	if ok {
		return lock.(*sync.Mutex).TryLock()
	}

	lock = &sync.Mutex{}
	ok = lock.(*sync.Mutex).TryLock()

	if ok {
		locks.Store(key, lock)

		go func() {
			time.Sleep(dur)
			lock.(*sync.Mutex).Unlock()
			locks.Delete(key)
		}()
	}

	return ok
}

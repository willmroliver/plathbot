package util

import (
	"sync"
	"time"
)

var locks = map[any]time.Time{}
var mut = &sync.Mutex{}

func TryLockFor(key any, dur time.Duration) bool {
	mut.Lock()
	defer mut.Unlock()

	lock, ok := locks[key]

	if ok && lock.After(time.Now()) {
		return false
	}

	locks[key] = time.Now().Add(dur)

	return true
}

var breaker = true

func InitLockerTidy(every time.Duration) {
	breaker = true

	go func() {
		for breaker {
			time.Sleep(every)
			for key, lock := range locks {
				if lock.Before(time.Now()) {
					delete(locks, key)
				}
			}
		}
	}()
}

func HaultLockerTidy() {
	breaker = false
}

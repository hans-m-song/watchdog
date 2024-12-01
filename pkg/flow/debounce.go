package flow

import (
	"sync"
	"time"
)

type Debouncer func(func())

func NewDebouncer(delay time.Duration) Debouncer {
	var timer *time.Timer
	var mu sync.Mutex

	return func(f func()) {
		mu.Lock()
		defer mu.Unlock()

		if timer != nil {
			timer.Stop()
		}

		timer = time.AfterFunc(delay, f)
	}
}

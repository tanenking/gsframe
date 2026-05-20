package helper

import (
	"sync"
	"time"
)

func Debounce(f func(), wait_milli int) func() {
	var timer *time.Timer
	var l sync.Mutex
	return func() {
		l.Lock()
		defer l.Unlock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(time.Duration(wait_milli)*time.Millisecond, f)
	}
}

package gsinf

import (
	"time"
)

type ITimer interface {
	Stop() bool
}
type ITimingWheel interface {
	Start()
	Stop()
	AfterFunc(d time.Duration, f func(udata ...interface{}), udata ...interface{}) ITimer
	ScheduleFunc(interval time.Duration, f func(udata ...interface{}), udata ...interface{}) (t ITimer)
}

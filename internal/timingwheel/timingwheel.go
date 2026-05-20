package timingwheel

import (
	"errors"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/timingwheel/delayqueue"
)

type TimingWheel struct {
	tick      int64 // in milliseconds
	wheelSize int64

	interval    int64 // in milliseconds
	currentTime int64 // in milliseconds
	buckets     []*bucket
	queue       *delayqueue.DelayQueue

	overflowWheel unsafe.Pointer // type: *TimingWheel

	exitC     chan struct{}
	waitGroup waitGroupWrapper
}

func NewTimingWheel(tick time.Duration, wheelSize int64) gsinf.ITimingWheel {
	tickMs := int64(tick / time.Millisecond)
	if tickMs <= 0 {
		panic(errors.New("tick must be greater than or equal to 1ms"))
	}

	startMs := timeToMs(time.Now().UTC())

	return newTimingWheel(
		tickMs,
		wheelSize,
		startMs,
		delayqueue.New(int(wheelSize)),
	)
}

func newTimingWheel(tickMs int64, wheelSize int64, startMs int64, queue *delayqueue.DelayQueue) *TimingWheel {
	buckets := make([]*bucket, wheelSize)
	for i := range buckets {
		buckets[i] = newBucket()
	}
	return &TimingWheel{
		tick:        tickMs,
		wheelSize:   wheelSize,
		currentTime: truncate(startMs, tickMs),
		interval:    tickMs * wheelSize,
		buckets:     buckets,
		queue:       queue,
		exitC:       make(chan struct{}),
	}
}

func (tw *TimingWheel) add(t *Timer) bool {
	currentTime := atomic.LoadInt64(&tw.currentTime)
	if t.expiration < currentTime+tw.tick {
		return false
	} else if t.expiration < currentTime+tw.interval {
		virtualID := t.expiration / tw.tick
		b := tw.buckets[virtualID%tw.wheelSize]
		b.Add(t)

		if b.SetExpiration(virtualID * tw.tick) {
			tw.queue.Offer(b, b.Expiration())
		}

		return true
	} else {
		overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
		if overflowWheel == nil {
			atomic.CompareAndSwapPointer(
				&tw.overflowWheel,
				nil,
				unsafe.Pointer(newTimingWheel(
					tw.interval,
					tw.wheelSize,
					currentTime,
					tw.queue,
				)),
			)
			overflowWheel = atomic.LoadPointer(&tw.overflowWheel)
		}
		return (*TimingWheel)(overflowWheel).add(t)
	}
}

func (tw *TimingWheel) addOrRun(t *Timer) {
	if !tw.add(t) {
		constants.Go(func() { t.task(t.udata...) })
	}
}

func (tw *TimingWheel) advanceClock(expiration int64) {
	currentTime := atomic.LoadInt64(&tw.currentTime)
	if expiration >= currentTime+tw.tick {
		currentTime = truncate(expiration, tw.tick)
		atomic.StoreInt64(&tw.currentTime, currentTime)

		overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
		if overflowWheel != nil {
			(*TimingWheel)(overflowWheel).advanceClock(currentTime)
		}
	}
}

func (tw *TimingWheel) Start() {
	tw.waitGroup.Wrap(func() {
		tw.queue.Poll(tw.exitC, func() int64 {
			return timeToMs(time.Now().UTC())
		})
	})

	tw.waitGroup.Wrap(func() {
		for {
			select {
			case elem := <-tw.queue.C:
				b := elem.(*bucket)
				tw.advanceClock(b.Expiration())
				b.Flush(tw.addOrRun)
			case <-tw.exitC:
				return
			}
		}
	})
}

func (tw *TimingWheel) Stop() {
	close(tw.exitC)
	tw.waitGroup.Wait()
}

func (tw *TimingWheel) AfterFunc(d time.Duration, f func(udata ...interface{}), udata ...interface{}) gsinf.ITimer {
	t := &Timer{
		expiration: timeToMs(time.Now().UTC().Add(d)),
		task:       f,
		udata:      udata,
	}
	tw.addOrRun(t)
	return t
}

//	type Scheduler interface {
//		Next(time.Time) time.Time
//	}
type Scheduler struct {
	Interval time.Duration
}

func (s *Scheduler) Next(prev time.Time) time.Time {
	return prev.Add(s.Interval)
}
func (tw *TimingWheel) ScheduleFunc(interval time.Duration, f func(udata ...interface{}), udata ...interface{}) (t gsinf.ITimer) {
	return tw.scheduleFunc(&Scheduler{Interval: interval}, f, udata...)
}
func (tw *TimingWheel) scheduleFunc(s *Scheduler, f func(udata ...interface{}), udata ...interface{}) (t *Timer) {
	expiration := s.Next(time.Now().UTC())
	if expiration.IsZero() {
		return
	}

	t = &Timer{
		expiration: timeToMs(expiration),
		task: func(udata ...interface{}) {
			expiration := s.Next(msToTime(t.expiration))
			if !expiration.IsZero() {
				t.expiration = timeToMs(expiration)
				tw.addOrRun(t)
			}

			f(udata...)
		},
		udata: udata,
	}
	tw.addOrRun(t)

	return
}

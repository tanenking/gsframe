package constants

import (
	"fmt"
	"os"
	"runtime/debug"
	"sync/atomic"
)

// func gocaller() string {
// 	_, file, line, _ := runtime.Caller(2)
// 	return fmt.Sprintf("%s:%d", file, line)
// }

var poolChan = make(chan func())
var poolGoCount int32 = 0
var poolSize int32 = 50000
var panicCount int32 = 0

func init() {
	// poolSize = int32(runtime.NumCPU() * 2)
}

func Try(fun func()) {
	defer AutoRecover()()
	fun()
}

func TryHandle(fun func(), handler func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			atomic.AddInt32(&panicCount, 1)
			os.Stderr.WriteString(fmt.Sprintf(`%+v`, err))
			debug.PrintStack()
			if handler != nil {
				handler(err)
			}
		}
	}()
	fun()
}

func Go(fn func()) {
	pc := poolGoCount + 1
	select {
	case poolChan <- fn:
		return
	default:
	}

	AppExitWait()
	go func() {
		pc = atomic.AddInt32(&poolGoCount, 1)
		Try(fn)
		for pc <= poolSize {
			select {
			case <-ExitChannel:
				pc = poolSize + 1
			case nfn := <-poolChan:
				Try(nfn)
			}
		}
		atomic.AddInt32(&poolGoCount, -1)
		AppExitDone()
	}()
}

func Go2(fn func(params ...interface{}), params ...interface{}) {
	pc := poolGoCount + 1
	select {
	case poolChan <- func() { fn(params...) }:
		return
	default:
	}

	AppExitWait()
	go func() {
		pc = atomic.AddInt32(&poolGoCount, 1)
		Try(func() { fn(params...) })
		for pc <= poolSize {
			select {
			case <-ExitChannel:
				pc = poolSize + 1
			case nfn := <-poolChan:
				Try(nfn)
			}
		}
		atomic.AddInt32(&poolGoCount, -1)
		AppExitDone()
	}()
}

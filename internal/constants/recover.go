package constants

import (
	"fmt"
	"os"
	"runtime/debug"
	"sync/atomic"
)

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func AutoRecover() func() {
	return func() {
		if err := recover(); err != nil {
			atomic.AddInt32(&panicCount, 1)
			os.Stderr.WriteString(fmt.Sprintf(`%+v`, err))
			debug.PrintStack()
		}
	}
}

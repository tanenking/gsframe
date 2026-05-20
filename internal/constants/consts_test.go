package constants_test

import (
	"fmt"
	"testing"
	"time"
)

func TestGoTry(t *testing.T) {
}

func TestGoGo(t *testing.T) {
	var poolChan = make(chan func())
	var ExitChannel1 = make(chan struct{})
	go func() {
		for {
			select {
			case <-ExitChannel1:
				return
			case nfn := <-poolChan:
				nfn()
			}
		}
	}()
	time.Sleep(time.Second * 3)
	poolChan <- func() { fmt.Println(`test`) }
	time.Sleep(time.Second * 3)
	ExitChannel1 <- struct{}{}
	time.Sleep(time.Second * 3)
}

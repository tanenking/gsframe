package constants

import "sync"

var ExitChannel chan struct{}
var ExitWaitGroup sync.WaitGroup

func init() {
	ExitChannel = make(chan struct{})
	ExitWaitGroup = sync.WaitGroup{}
}

func AppExitWait() {
	ExitWaitGroup.Add(1)
}
func AppExitDone() {
	ExitWaitGroup.Done()
}

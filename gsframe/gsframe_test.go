package gsframe_test

import (
	"fmt"
	"testing"
	"time"
)

func TestMake(t *testing.T) {
	Data := make([]byte, 0, 100)
	fmt.Println(Data)
}

func TestChannle(t *testing.T) {
	go func() {
		ch := make(chan int, 10)
		ch <- 1
		ch <- 99
		close(ch)
		for {
			select {
			case data, ok := <-ch:
				if !ok {
					fmt.Println(`!ok`)
					break
				}
				fmt.Println(data)
			default:
				return
			}
		}
	}()
	time.Sleep(time.Second * 5)
	fmt.Println(`finish`)
}

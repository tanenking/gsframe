//go:build windows
// +build windows

package logx

import (
	"syscall"
)

var (
	//windows console color
	proc        interface{}
	closeHandle interface{}
	// procSetStdHandle interface{}
)

var (
	p *syscall.LazyProc
	c *syscall.LazyProc
)

func colorPrint(s string, i int) {
	handle, _, _ := p.Call(uintptr(syscall.Stdout), uintptr(i))
	// print(s, "\n")
	c.Call(handle)
}
func initKernel32() {
	kernel32 := syscall.NewLazyDLL(`kernel32.dll`)
	proc = kernel32.NewProc(`SetConsoleTextAttribute`)
	closeHandle = kernel32.NewProc(`CloseHandle`)
	// procSetStdHandle = kernel32.NewProc("SetStdHandle")

	p = proc.(*syscall.LazyProc)
	c = closeHandle.(*syscall.LazyProc)
}

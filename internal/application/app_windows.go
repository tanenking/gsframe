//go:build windows
// +build windows

package application

import (
	"syscall"
	"unsafe"
)

func setTitle(title string) {
	kernel32, _ := syscall.LoadLibrary(`kernel32.dll`)
	sct, _ := syscall.GetProcAddress(kernel32, `SetConsoleTitleW`)
	syscall.Syscall(sct, 1, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))), 0, 0)
	// p, e := syscall.UTF16PtrFromString(title)
	// if e != nil {
	// 	return
	// }
	// syscall.SyscallN(sct, 1, uintptr(unsafe.Pointer(p)), 0, 0)
	syscall.FreeLibrary(kernel32)
}

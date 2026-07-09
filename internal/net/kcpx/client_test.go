package kcpx

import (
	"fmt"
	"testing"

	"github.com/tanenking/gsframe/gsinf"
)

func TestXxx(t *testing.T) {
	var header = (int64(gsinf.KcpControlCMD) << int64(32)) | int64(1)
	fmt.Println(header)
	cmdid := int32(header >> 32)
	cmd := int32(header & 0xffffffff)
	fmt.Println(cmdid, cmd)
}

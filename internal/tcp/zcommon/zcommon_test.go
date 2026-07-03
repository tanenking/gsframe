package zcommon_test

import (
	"fmt"
	"testing"

	"github.com/tanenking/gsframe/internal/tcp/zcommon"
)

func fff(s []byte) {
}
func TestXxx123(t *testing.T) {
	b := make([]byte, 10)
	fff(b)
	fmt.Println(b)
}

func TestXxx(t *testing.T) {
	msg1 := zcommon.MessagePoop.Get().(*zcommon.Message)
	defer func() {
		zcommon.MessagePoop.Put(msg1)
	}()
	msg1.ID = `id`
	msg1.RequestSeq = 23
	msg1.Data = []byte(`maliang`)
	msg1.DataLen = uint32(len(msg1.Data))
	pkg1, _ := zcommon.Pack(msg1)

	msg2 := zcommon.MessagePoop.Get().(*zcommon.Message)
	defer func() {
		zcommon.MessagePoop.Put(msg2)
	}()
	msg2.ID = `iddf`
	msg2.RequestSeq = 4
	msg2.Data = []byte(`leon`)
	msg2.DataLen = uint32(len(msg2.Data))
	pkg2, _ := zcommon.Pack(msg2)

	msg := zcommon.MessagePoop.Get().(*zcommon.Message)
	zcommon.UnpackFromBytes(pkg1, uint32(len(pkg1)), msg)
	fmt.Println(msg)
	zcommon.MessagePoop.Put(msg)

	msg = zcommon.MessagePoop.Get().(*zcommon.Message)
	zcommon.UnpackFromBytes(pkg2, uint32(len(pkg2)), msg)
	fmt.Println(msg)
	zcommon.MessagePoop.Put(msg)
}

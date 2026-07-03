package common

import "sync"

var bytePool = sync.Pool{
	New: func() any {
		b := make([]byte, 1024)
		b = b[:0]
		return b
	},
}

// func CreateByteBuffer(maxlen int) []byte {
// 	var bs = bytePool.Get().([]byte)
// 	if cap(bs) < maxlen {
// 		bs = make([]byte, 0, maxlen)
// 	}
// 	bs = bs[:0]
// 	return bs
// }
// func DeleteByteBuffer(bs []byte) {
// 	bytePool.Put(bs)
// }

type ByteBuffer struct {
	Data []byte
}

var byteBufferPool = sync.Pool{
	New: func() any {
		b := &ByteBuffer{Data: make([]byte, 1024)}
		b.Data = b.Data[:0]
		return b
	},
}

func CreateByteBuffer(maxlen int) *ByteBuffer {
	var bb = byteBufferPool.Get().(*ByteBuffer)
	if cap(bb.Data) < maxlen {
		bb.Data = make([]byte, 0, maxlen)
	}
	bb.Data = bb.Data[:0]
	return bb
}
func DeleteByteBuffer(bb *ByteBuffer) {
	byteBufferPool.Put(bb)
}

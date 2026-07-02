package common

import "sync"

var bytePool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 1024)
		return b
	},
}

func CreateByteBuffer(maxlen int) []byte {
	var bs = bytePool.Get().([]byte)
	if cap(bs) < maxlen {
		bs = make([]byte, 0, maxlen)
	}
	bs = bs[:0]
	return bs
}
func DeleteByteBuffer(bs []byte) {
	bytePool.Put(bs)
}

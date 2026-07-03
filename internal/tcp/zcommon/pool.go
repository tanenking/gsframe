package zcommon

import "sync"

var RequestPool = sync.Pool{
	New: func() interface{} {
		b := &Request{}
		return b
	},
}

var MessagePoop = sync.Pool{
	New: func() any {
		b := &Message{}
		b.Data = make([]byte, 0, GlobalObject.MaxPacketSize)
		return b
	},
}

var BytePool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, GlobalObject.MaxPacketSize)
		return b
	},
}

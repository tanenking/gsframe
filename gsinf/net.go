package gsinf

import "encoding/binary"

type IMessage interface {
	GetHeader() int64
	GetMsgID() string
	GetData() []byte
	GetByteOrder() binary.ByteOrder
}

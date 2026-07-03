package common

import (
	"bytes"
	"encoding/binary"
	"sync"
)

// Message 消息
type Message struct {
	Header    int64
	ID        string //消息的ID
	Data      []byte //消息的内容
	byteorder binary.ByteOrder
}

var messagePool = sync.Pool{
	New: func() any {
		return &Message{
			Data: make([]byte, 0, 4096),
		}
	},
}

// NewMsgPackage 创建一个Message消息包
func CreateMessage(byteorder binary.ByteOrder) *Message {
	var msg = messagePool.Get().(*Message)
	msg.Data = msg.Data[:0]
	msg.byteorder = byteorder
	return msg
}

func DeleteMessage(msg *Message) {
	messagePool.Put(msg)
}

// GetDataLen 获取消息数据段长度
func (msg *Message) GetDataLen() uint32 {
	return uint32(len(msg.Data))
}

// GetData 获取消息内容
func (msg *Message) GetHeader() int64 {
	return msg.Header
}

// GetMsgID 获取消息ID
func (msg *Message) GetMsgID() string {
	return msg.ID
}

// GetData 获取消息内容
func (msg *Message) GetData() []byte {
	return msg.Data
}

func (msg *Message) GetByteOrder() binary.ByteOrder {
	return msg.byteorder
}

// SetMsgID 设计消息ID
func (msg *Message) SetMsgID(msgID string) {
	msg.ID = msgID
}
func (msg *Message) ToBytes(outbs *ByteBuffer) error {
	var nameSize = uint8(len(msg.ID))
	var totalLen int32 = 0
	totalLen = int32(binary.Size(totalLen) + binary.Size(msg.Header) + binary.Size(nameSize) + int(nameSize) + len(msg.Data))
	if cap(outbs.Data) < int(totalLen) {
		outbs.Data = make([]byte, 0, totalLen)
	}
	//创建一个写入缓冲区
	dataBuff := bytes.NewBuffer(outbs.Data)

	//写入4字节消息总长度
	if err := binary.Write(dataBuff, msg.byteorder, totalLen); err != nil {
		return err
	}
	//写入8字节header
	if err := binary.Write(dataBuff, msg.byteorder, msg.Header); err != nil {
		return err
	}
	//写入id字节长度
	if err := binary.Write(dataBuff, msg.byteorder, nameSize); err != nil {
		return err
	}
	if nameSize > 0 {
		//再写入id
		if err := binary.Write(dataBuff, msg.byteorder, []byte(msg.ID)); err != nil {
			return err
		}
	}
	//写入data数据
	if len(msg.Data) > 0 {
		if err := binary.Write(dataBuff, msg.byteorder, msg.Data); err != nil {
			return err
		}
	}
	return nil
}
func (msg *Message) FromBytes(inbs []byte) error {
	//创建一个从输入二进制数据的ioReader
	dataBuff := bytes.NewReader(inbs)
	var totalLen int32 = 0
	//读取4字节消息总长度
	if err := binary.Read(dataBuff, msg.byteorder, &totalLen); err != nil {
		return err
	}
	//读取8字节header
	if err := binary.Read(dataBuff, msg.byteorder, &msg.Header); err != nil {
		return err
	}
	//读id字节长度
	var nameSize uint8 = 0
	if err := binary.Read(dataBuff, msg.byteorder, &nameSize); err != nil {
		return err
	}
	//读id
	if nameSize > 0 {
		name := CreateByteBuffer(256)
		defer DeleteByteBuffer(name)
		if err := binary.Read(dataBuff, msg.byteorder, name.Data); err != nil {
			return err
		}
		msg.ID = string(name.Data)
	} else {
		msg.ID = ``
	}

	var dataLen = len(inbs) - binary.Size(totalLen) - binary.Size(msg.Header) - binary.Size(nameSize) - int(nameSize)
	if cap(msg.Data) < dataLen {
		msg.Data = make([]byte, 0, dataLen)
	}
	msg.Data = msg.Data[:0]
	if dataLen > 0 {
		//读data
		if err := binary.Read(dataBuff, msg.byteorder, msg.Data); err != nil {
			return err
		}
	}
	return nil
}

func ReadMessageTotalLength(bs []byte, byteorder binary.ByteOrder) int32 {
	var totalLen int32 = 0
	if len(bs) < binary.Size(totalLen) {
		return totalLen
	}
	if byteorder == binary.LittleEndian {
		return int32(binary.LittleEndian.Uint32(bs[:binary.Size(totalLen)]))
	}
	return int32(binary.BigEndian.Uint32(bs[:binary.Size(totalLen)]))
}

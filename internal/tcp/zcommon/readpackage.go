package zcommon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type ReadPackage struct {
	ReadSize  uint32 //已读取的长度
	TotalSize uint32 //消息总长度
	Data      []byte //消息体(size + namesize + name + msg)
}

func NewReadPackage() *ReadPackage {
	r := &ReadPackage{
		ReadSize:  0,
		TotalSize: 0,
		Data:      make([]byte, GlobalObject.MaxPacketSize),
	}
	return r
}

func (r *ReadPackage) Success() bool {
	return r.ReadSize >= r.TotalSize
}

func (r *ReadPackage) Clear() {
	r.ReadSize = 0
	r.TotalSize = 0
}

func (r *ReadPackage) ReadFromConn(rd io.Reader) (err error) {

	if rd == nil {
		err = fmt.Errorf("conn was nil")
		return
	}

	var rcount int
	headSize := uint32(binary.Size(r.TotalSize))
	if r.ReadSize < headSize {
		//先读4个字节的头信息
		rcount, err = rd.Read(r.Data[r.ReadSize:headSize])
		if err != nil {
			return
		}
		r.ReadSize += uint32(rcount)
	}
	if r.ReadSize >= headSize {
		if r.TotalSize <= 0 {
			//解析头信息
			b := bytes.NewReader(r.Data[:headSize])
			err = binary.Read(b, ByteOrder, &r.TotalSize)
			if err != nil {
				return
			}
			if r.TotalSize <= 0 || (GlobalObject.MaxPacketSize > 0 && r.TotalSize > GlobalObject.MaxPacketSize) {
				err = fmt.Errorf("err TotalSize = %d", r.TotalSize)
				return
			}
		}
		if r.TotalSize > 0 {
			rcount, err = rd.Read(r.Data[r.ReadSize:r.TotalSize])
			if err != nil {
				return
			}
			r.ReadSize += uint32(rcount)
		}
	}
	return
}

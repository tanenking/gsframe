package zcommon

import "github.com/tanenking/gsframe/gsinf"

// Request 请求
type Request struct {
	Conn gsinf.IConnection //已经和客户端建立好的 链接
	Msg  *Message          //客户端请求的数据
}

// GetConnection 获取请求连接信息
func (r *Request) GetConnection() gsinf.IConnection {
	return r.Conn
}

// GetData 获取请求消息的数据
func (r *Request) GetData() []byte {
	return r.Msg.GetData()
}

// GetMsgID 获取请求的消息的ID
func (r *Request) GetMsgRequestNo() int32 {
	return r.Msg.GetRequestSeq()
}

// GetMsgID 获取请求的消息的ID
func (r *Request) GetMsgID() string {
	return r.Msg.GetMsgID()
}

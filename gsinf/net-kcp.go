package gsinf

import (
	"encoding/binary"
	"time"
)

// //////////////////////////////////////////////////////////////
type IKcpServer interface {
	GetConnection(connId int32) IKcpConnection
	GetConnectionCount() int32
	SendGroup(groupName string, header int64, msgID string, data []byte)
}

type IKcpConnection interface {
	ClientIP() string
	GetConnID() int32
	Stop()
	IsValid() bool
	InGroup(groupName string) bool
	AddGroup(groupName string)
	DeleteGroup(groupName string)
	GetGroupList() []string
	SetProperty(key string, value interface{})
	GetProperty(key string) (interface{}, error)
	RemoveProperty(key string)
	Send(header int64, msgID string, data []byte) error
}

type IKcpClient interface {
	Send(header int64, msgID string, data []byte) error
}

type IKcpClientMessageCallback interface {
	Handle(msg IMessage) //处理conn业务的方法
}

type KcpClientConfig struct {
	// IP:Port
	Address string
	//连接池大小
	PoolSize int
	//go缓冲区大小
	GoReadWriteBufferSize int32
	//tcp缓冲区大小
	TcpReadWriteBufferSize int32
	//数据包的最大值
	MaxPacketSize int32
	//发送消息的缓冲最大长度
	WriteMessageBufferLen int32
	//读超时时间
	ReadTimeout time.Duration
	//写超时时间
	WriteTimeout time.Duration
	//
	NoDelay bool
	//流模式
	StreamMode bool
	//字节序
	ByteOrder binary.ByteOrder
	//消息处理回调
	MessageCallback IKcpClientMessageCallback
}

////////////////////////////////////////////////////////////////////////

type IKcpConnectionMessageCallback interface {
	PreHandle(conn IKcpConnection, msg IMessage) bool //在处理conn业务之前的钩子方法
	Handle(conn IKcpConnection, msg IMessage)         //处理conn业务的方法
	PostHandle(conn IKcpConnection, msg IMessage)     //处理conn业务之后的钩子方法
}

type KcpServerConfig struct {
	//监听端口
	Port int
	//go缓冲区大小
	GoReadWriteBufferSize int32
	//tcp缓冲区大小
	TcpReadWriteBufferSize int32
	//数据包的最大值
	MaxPacketSize int32
	//发送消息的缓冲最大长度
	WriteMessageBufferLen int32
	//读超时时间
	ReadTimeout time.Duration
	//写超时时间
	WriteTimeout time.Duration
	//心跳超时时间
	HeartTimeoutSec int64
	//
	NoDelay bool
	//流模式
	StreamMode bool
	//字节序
	ByteOrder binary.ByteOrder

	//连接创建后的回调
	OnConnectionCreate func(conn IKcpConnection)
	//连接终止前的回调
	OnConnectionStop func(conn IKcpConnection)

	//ip黑名单验证, 用于创建连接判断是否拒绝连接
	IPBlackValidate func(ip string) bool

	//消息处理回调
	MessageCallback IKcpConnectionMessageCallback
}

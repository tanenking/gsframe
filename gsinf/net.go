package gsinf

import (
	"encoding/binary"
	"time"

	"golang.org/x/time/rate"
)

type IMessage interface {
	GetHeader() int64
	GetMsgID() string
	GetData() []byte
	GetByteOrder() binary.ByteOrder
}

// //////////////////////////////////////////////////////////////
type IWebSocketServer interface {
	GetConnection(connId int32) IWebSocketConnection
	GetConnectionCount() int32
}

type IWebSocketConnection interface {
	ClientIP() string
	GetConnID() int32
	Stop()
	IsValid() bool
	InGroup(groupName string) bool
	AddGroup(groupName string)
	DeleteGroup(groupName string)
	GetGroupList() []string
	WaitLimiterToken() error
	SetProperty(key string, value interface{})
	GetProperty(key string) (interface{}, error)
	RemoveProperty(key string)
	Send(header int64, msgID string, data []byte) error
}

type IWebSocketMessageCallback interface {
	PreHandle(conn IWebSocketConnection, msg IMessage) bool //在处理conn业务之前的钩子方法
	Handle(conn IWebSocketConnection, msg IMessage)         //处理conn业务的方法
	PostHandle(conn IWebSocketConnection, msg IMessage)     //处理conn业务之后的钩子方法
}

type WebSocketServerConfig struct {
	//监听端口
	Port int
	//go缓冲区大小
	GoReadWriteBufferSize int32
	//tcp缓冲区大小
	TcpReadWriteBufferSize int32
	//数据包的最大值
	MaxPacketSize int32
	//当前服务器主机允许的最大链接个数
	MaxConn int32
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
	//限流器
	LimiterLimit rate.Limit
	//限流器超时时间
	LimiterTimeout time.Duration
	//限流器桶大小
	LimiterBucketCount int32
	//字节序
	ByteOrder binary.ByteOrder

	//连接创建后的回调
	OnConnectionCreate func(conn IWebSocketConnection)
	//连接终止前的回调
	OnConnectionStop func(conn IWebSocketConnection)

	//ip黑名单验证, 用于创建连接判断是否拒绝连接
	IPBlackValidate func(ip string) bool

	//消息处理回调
	MessageCallback IWebSocketMessageCallback
}

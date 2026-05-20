package gsinf

import (
	"context"
	"net"
	"time"
)

// 定义连接接口
type IConnection interface {
	Start()                   //启动连接，让当前连接开始工作
	Stop()                    //停止连接，结束当前连接状态M
	Context() context.Context //返回ctx，用于用户自定义的go程获取连接退出状态

	//GetTCPConnection() *net.TCPConn //从当前连接获取原始的socket TCPConn
	GetConnID() uint32    //获取当前连接ID
	RemoteAddr() net.Addr //获取远程客户端地址信息
	ClientIP() string
	ClientAddress() string //ip:port

	SendBuffMsg(ctx context.Context, msgID string, data []byte) error //直接将Message数据发送给远程的TCP客户端(有缓冲)

	SetProperty(key string, value interface{})   //设置链接属性
	GetProperty(key string) (interface{}, error) //获取链接属性
	RemoveProperty(key string)                   //移除链接属性

	GetLimiterToken() error //限流器
	IsValid() bool          //是否有效连接
	IsClosing() bool
	InGroup(groupName string) bool
	AddGroup(groupName string)
	DeleteGroup(groupName string)
	GetGroupList() []string
}

/*
连接管理抽象层
*/
type IConnManager interface {
	Add(conn IConnection)                   //添加链接
	Remove(conn IConnection)                //删除连接
	Get(connID uint32) (IConnection, error) //利用ConnID获取链接
	Len() int                               //获取当前连接
	ClearConn()                             //删除并停止所有链接
}

/*
IRequest 接口：
实际上是把客户端请求的链接信息 和 请求的数据 包装到了 Request里
*/
type IRequest interface {
	GetConnection() IConnection //获取请求连接信息
	GetData() []byte            //获取请求消息的数据
	GetMsgRequestNo() int32     //获取消息请求序列号
	GetMsgID() string           //获取请求的消息ID
}

/*
路由接口， 这里面路由是 使用框架者给该链接自定的 处理业务方法
路由里的IRequest 则包含用该链接的链接信息和该链接的请求数据信息
*/
type IRouter interface {
	PreHandle(request IRequest) bool //在处理conn业务之前的钩子方法
	Handle(request IRequest)         //处理conn业务的方法
	PostHandle(request IRequest)     //处理conn业务之后的钩子方法
}

type Router struct{}

func (Router) PreHandle(request IRequest) bool { return true }
func (Router) Handle(request IRequest)         { panic(`not imp handle`) }
func (Router) PostHandle(request IRequest)     {}

type IConnPreValidate interface {
	BlackListValidate(ip string) bool //白名单验证
}

// 定义服务接口
type IServer interface {
	Start()                                      //启动服务器方法
	Stop()                                       //停止服务器方法
	Serve()                                      //开启业务服务方法
	RegisterRouter(msgID string, router IRouter) //路由功能: 给当前服务注册一个路由业务方法，供客户端链接处理使用
	RegisterGlobalRouter(router IRouter)         //路由功能: 没有指定的消息,都通过这个处理
	GetConnMgr() IConnManager                    //得到链接管理
	SetOnConnStart(func(IConnection))            //设置该Server的连接创建时 Hook函数
	SetOnConnStop(func(IConnection))             //设置该Server的连接断开时的 Hook函数
	SetOnConnPreValidate(IConnPreValidate)
	CallOnConnStart(conn IConnection) //调用连接OnConnStart Hook函数
	CallOnConnStop(conn IConnection)  //调用连接OnConnStop Hook函数
	CallOnPreValidate(conn IConnection) bool
	IsClosing() bool //是否关闭中
	Close()          //关闭
}

type TcpGlobalConfig_t struct {
	TCPServer IServer //当前全局Server对象
	Host      string  //当前服务器主机IP
	TCPPort   int     //当前服务器主机监听端口号

	MaxPacketSize uint32 //都需数据包的最大值
	MaxConn       int    //当前服务器主机允许的最大链接个数
	MaxMsgChanLen uint32 //SendBuffMsg发送消息的缓冲最大长度

	DirectProtoMsg bool

	ReadTimeout  time.Duration //读超时时间
	WriteTimeout time.Duration //写超时时间
}

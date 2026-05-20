package znet

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/logx"
	"github.com/tanenking/gsframe/internal/tcp/zcommon"
)

// Server 接口实现，定义一个Server服务类
type Server struct {
	//tcp4 or other
	IPVersion string
	//服务绑定的IP地址
	IP string
	//服务绑定的端口
	Port int
	//当前Server的消息管理模块，用来绑定MsgID和对应的处理方法
	msgHandler *zcommon.MsgHandle
	//当前Server的链接管理器
	ConnMgr gsinf.IConnManager
	//该Server的连接创建时Hook函数
	OnConnStart func(conn gsinf.IConnection)
	//该Server的连接断开时的Hook函数
	OnConnStop    func(conn gsinf.IConnection)
	OnPreValidate gsinf.IConnPreValidate
	//是否关闭中
	Closing bool
}

// NewServer 创建一个服务器句柄
func NewServer() gsinf.IServer {
	zcommon.PrintLogo()

	zcommon.Validate()

	s := &Server{
		IPVersion:     "tcp4",
		IP:            zcommon.GlobalObject.Host,
		Port:          zcommon.GlobalObject.TCPPort,
		msgHandler:    zcommon.NewMsgHandle(),
		ConnMgr:       zcommon.NewConnManager(),
		OnPreValidate: nil,
	}

	return s
}

//============== 实现 gsinf.IServer 里的全部接口方法 ========

func (s *Server) IsClosing() bool { //是否关闭中
	return s.Closing
}

func (s *Server) Close() { //是否关闭中
	s.Closing = true
}

// Start 开启网络服务
func (s *Server) Start() {
	fmt.Printf("[START] Server name: %s,listenner at IP: %s, Port %d is starting\n", constants.ServiceType, s.IP, s.Port)

	//开启一个go去做服务端Linster业务
	constants.Go(func() {
		//1 获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			logx.DebugF("resolve tcp addr err: %+v", err)
			return
		}

		//2 监听服务器地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			panic(err)
		}

		//已经监听成功
		logx.DebugF("start Zinx server  %s %s", constants.ServiceType, " success, now listenning...")

		//TODO server 应该有一个自动生成ID的方法
		var cID atomic.Uint32
		cID.Store(0)

		//3 启动server网络连接业务
		for {
			time.Sleep(time.Millisecond * 20)
			//3.1 阻塞等待客户端建立连接请求
			conn, err := listener.AcceptTCP()
			if err != nil {
				logx.DebugF("Accept err %+v", err)
				continue
			}
			if s.Closing || constants.GetSystemStatus() == gsinf.SystemStatus_Maintain {
				//正在关闭中,不接收新连接了
				continue
			}
			logx.DebugF("Get conn remote addr = %s", conn.RemoteAddr().String())

			//3.2 设置服务器最大连接控制,如果超过最大连接，那么则关闭此新的连接
			if s.ConnMgr.Len() >= zcommon.GlobalObject.MaxConn {
				conn.Close()
				continue
			}

			//3.3 处理该新连接请求的 业务 方法， 此时应该有 handler 和 conn是绑定的
			id := cID.Add(1)
			if cID.Load() >= 0x7fffffff {
				cID.Store(0)
			}
			dealConn := NewConnection(s, conn, id, s.msgHandler)
			if !s.CallOnPreValidate(dealConn) {
				logx.ErrorF("%s 连接前置验证失败, 丢弃连接", dealConn.ClientIP())
				conn.Close()
				return
			}

			//将新创建的Conn添加到链接管理中
			s.GetConnMgr().Add(dealConn)
			//3.4 启动当前链接的处理业务
			constants.Go(func() { dealConn.Start() })
		}
	})
}

// Stop 停止服务
func (s *Server) Stop() {
	s.ConnMgr.ClearConn()
}

// Serve 运行服务
func (s *Server) Serve() {
	s.Start()

	//TODO Server.Serve() 是否在启动服务的时候 还要处理其他的事情呢 可以在这里添加

	//阻塞,否则主Go退出， listenner的go将会退出
	select {}
}

// RegisterRouter 路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
func (s *Server) RegisterRouter(msgID string, router gsinf.IRouter) {
	s.msgHandler.RegisterRouter(msgID, router)
}

// 路由功能: 没有指定的消息,都通过这个处理
func (s *Server) RegisterGlobalRouter(router gsinf.IRouter) {
	s.msgHandler.RegisterGlobalRouter(router)
}

// GetConnMgr 得到链接管理
func (s *Server) GetConnMgr() gsinf.IConnManager {
	return s.ConnMgr
}

// SetOnConnStart 设置该Server的连接创建时Hook函数
func (s *Server) SetOnConnStart(hookFunc func(gsinf.IConnection)) {
	s.OnConnStart = hookFunc
}

// SetOnConnStop 设置该Server的连接断开时的Hook函数
func (s *Server) SetOnConnStop(hookFunc func(gsinf.IConnection)) {
	s.OnConnStop = hookFunc
}

func (s *Server) SetOnConnPreValidate(validate gsinf.IConnPreValidate) {
	s.OnPreValidate = validate
}

// CallOnConnStart 调用连接OnConnStart Hook函数
func (s *Server) CallOnConnStart(conn gsinf.IConnection) {
	if s.OnConnStart != nil {
		logx.DebugF("---> CallOnConnStart....%d", conn.GetConnID())
		s.OnConnStart(conn)
	}
}

// CallOnConnStop 调用连接OnConnStop Hook函数
func (s *Server) CallOnConnStop(conn gsinf.IConnection) {
	if s.OnConnStop != nil {
		logx.DebugF("---> CallOnConnStop....%d", conn.GetConnID())
		s.OnConnStop(conn)
	}
}

func (s *Server) CallOnPreValidate(conn gsinf.IConnection) bool {
	if s.OnPreValidate == nil {
		return true
	}

	return !s.OnPreValidate.BlackListValidate(conn.ClientIP())
}

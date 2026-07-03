package zwspb

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/logger"
	"github.com/tanenking/gsframe/internal/tcp/zcommon"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

func (s *Server) handshake(c *gin.Context) {
	if websocket.IsWebSocketUpgrade(c.Request) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, c.Writer.Header())
		if err != nil {
			logger.Log().Error("upgrade err -> : %v", err)
			return
		}
		if s.Closing || constants.GetSystemStatus() == gsinf.SystemStatus_Maintain {
			//正在关闭中,不接收新连接了
			return
		}

		//3.2 设置服务器最大连接控制,如果超过最大连接，那么则关闭此新的连接
		if s.ConnMgr.Len() >= zcommon.GlobalObject.MaxConn {
			logger.Log().Error("当前连接数量超过最大值,放弃新连接")
			conn.Close()
		}

		//TODO server 应该有一个自动生成ID的方法

		//3.3 处理该新连接请求的 业务 方法， 此时应该有 handler 和 conn是绑定的
		id := cID.Add(1)
		if cID.Load() >= 0x7fffffff {
			cID.Store(0)
		}
		dealConn := NewConnection(s, conn, id, s.msgHandler, c.ClientIP())
		if !s.CallOnPreValidate(dealConn) {
			logger.Log().Error("%s 连接前置验证失败, 丢弃连接", dealConn.ClientAddress())
			conn.Close()
			return
		}
		_ = conn.SetReadDeadline(time.Now().Add(zcommon.GlobalObject.ReadTimeout))
		_ = conn.SetWriteDeadline(time.Now().Add(zcommon.GlobalObject.WriteTimeout))

		// 可选：设置Pong回调（刷新读超时，保持空闲连接存活）
		conn.SetPongHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(zcommon.GlobalObject.ReadTimeout))
			return nil
		})

		//将新创建的Conn添加到链接管理中
		s.GetConnMgr().Add(dealConn)
		//3.4 启动当前链接的处理业务
		constants.Go(func() { dealConn.Start() })
	} else {
		logger.Log().Error("不是websocket请求")
	}
}

// Start 开启网络服务
func (s *Server) Start() {
	fmt.Printf("[START] Server name: %s,listenner at IP: %s, Port %d is starting\n", constants.ServiceType, s.IP, s.Port)

	//开启一个go去做服务端Linster业务
	constants.Go(func() {
		g := gin.Default()
		g.GET("/", s.handshake)

		addr := fmt.Sprintf(":%d", s.Port)
		httpServer := &http.Server{
			Addr:    addr,
			Handler: g,
		}

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log().Error("ListenAndServe %v", err)
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
	// s.msgHandler.RegisterRouter(msgID, router)
	panic(`zwspb just only proto, not have msgId`)
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
		logger.Log().Debug("---> CallOnConnStart....%d", conn.GetConnID())
		s.OnConnStart(conn)
	}
}

// CallOnConnStop 调用连接OnConnStop Hook函数
func (s *Server) CallOnConnStop(conn gsinf.IConnection) {
	if s.OnConnStop != nil {
		logger.Log().Debug("---> CallOnConnStop....%d", conn.GetConnID())
		s.OnConnStop(conn)
	}
}

func (s *Server) CallOnPreValidate(conn gsinf.IConnection) bool {
	if s.OnPreValidate == nil {
		return true
	}

	return !s.OnPreValidate.BlackListValidate(conn.ClientIP())
}

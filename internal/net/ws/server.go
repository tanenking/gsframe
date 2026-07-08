package ws

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/logger"
	"github.com/tanenking/gsframe/internal/net/common"
)

type server struct {
	closed      atomic.Int32
	connections sync.Map //int32 => *connection
	idgenerate  int32
	ccount      int32

	groupMsgSeq  uint16
	groupMsgSync sync.Mutex
	groupMsgList [common.MaxGroupMsgCount]*common.GroupMessage
}

func CreateServer(_config *gsinf.WebSocketServerConfig) gsinf.IWebSocketServer {
	config = _config
	validateConfig()
	_server := &server{
		groupMsgList: [common.MaxGroupMsgCount]*common.GroupMessage{},
	}
	_server.groupMsgList[_server.groupMsgSeq] = &common.GroupMessage{C: make(chan struct{})}
	_server.start()
	return _server
}

func (r *server) GetConnection(connId int32) gsinf.IWebSocketConnection {
	if connId <= 0 {
		return nil
	}
	val, ok := r.connections.Load(connId)
	if !ok {
		return nil
	}
	return val.(*connection)
}

func (r *server) GetConnectionCount() int32 {
	return r.ccount
}

func (r *server) SendGroup(groupName string, header int64, msgID string, data []byte) {
	r.groupMsgSync.Lock()
	gmsg := r.groupMsgList[r.groupMsgSeq]
	r.groupMsgSeq++
	if r.groupMsgSeq >= common.MaxGroupMsgCount {
		r.groupMsgSeq = 0
	}
	r.groupMsgList[r.groupMsgSeq] = &common.GroupMessage{C: make(chan struct{})}
	r.groupMsgSync.Unlock()

	gmsg.GroupName = groupName
	gmsg.Header = header
	gmsg.MsgID = msgID
	if cap(gmsg.MsgData) < len(data) {
		gmsg.MsgData = make([]byte, 0, len(data))
	}
	gmsg.MsgData = gmsg.MsgData[:len(data)]
	copy(gmsg.MsgData, data)

	close(gmsg.C)
}

func (r *server) start() {
	//开启一个go去做服务端Linster业务
	constants.Go(func() {
		defer constants.AutoRecover()()

		g := gin.Default()
		g.GET("/", r.handshake)

		addr := fmt.Sprintf(":%d", config.Port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			panic(err)
		}
		defer listener.Close()

		config.Port = listener.Addr().(*net.TCPAddr).Port

		logger.Log().Info("[START] Server websocket server name: %s,listenner at Port %d is starting\n", constants.ServiceType, config.Port)

		httpServer := &http.Server{
			Handler: g,
		}
		err = httpServer.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	})
}

func (r *server) getFreeConnection() *connection {
	return connectionPool.Get().(*connection)
}

func (r *server) freeConnection(conn *connection) {
	if conn == nil {
		return
	}
	atomic.AddInt32(&r.ccount, -1)
	r.connections.Delete(conn.connId)
	connectionPool.Put(conn)
}

func (r *server) handshake(c *gin.Context) {
	if websocket.IsWebSocketUpgrade(c.Request) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, c.Writer.Header())
		if err != nil {
			logger.Log().Error("upgrade err -> : %v", err)
			return
		}
		if r.closed.Load() > 0 || constants.GetSystemStatus() == gsinf.SystemStatus_Maintain {
			//正在关闭中,不接收新连接了
			conn.Close()
			return
		}

		if config.IPBlackValidate != nil {
			if config.IPBlackValidate(c.ClientIP()) {
				logger.Log().Error("ip 黑名单 %s", c.ClientIP())
				conn.Close()
				return
			}
		}

		var dealConn = r.getFreeConnection()
		if dealConn == nil {
			logger.Log().Error("ws新建connection失败")
			conn.Close()
			return
		}
		atomic.AddInt32(&r.ccount, 1)
		connId := atomic.AddInt32(&r.idgenerate, 1)
		r.connections.Store(connId, dealConn)

		dealConn.init(r, conn, connId)

		constants.Go(func() { dealConn.start() })
	} else {
		logger.Log().Error("不是websocket请求")
	}
}

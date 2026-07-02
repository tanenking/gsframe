package ws

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/logger"
	"github.com/tanenking/gsframe/internal/net/common"
)

type server struct {
	sync.Mutex
	closed         atomic.Int32
	connections    []*connection
	cindex         int32
	ccount         int32
	connectionPool chan *connection

	groupMsgSeq  uint16
	groupMsgSync sync.Mutex
	groupMsgList [common.MaxGroupMsgCount]*common.GroupMessage
}

func CreateServer(_config *gsinf.WebSocketServerConfig) gsinf.IWebSocketServer {
	config = _config
	_server := &server{
		connections:    make([]*connection, _config.MaxConn),
		connectionPool: make(chan *connection, _config.MaxConn),
		groupMsgList:   [common.MaxGroupMsgCount]*common.GroupMessage{},
	}
	_server.start()
	return _server
}

func (r *server) GetConnection(connId int32) gsinf.IWebSocketConnection {
	if connId <= 0 || connId > config.MaxConn {
		return nil
	}
	idx := connId - 1
	return r.connections[idx]
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

		logger.Log().Info("[START] Server name: %s,listenner at Port %d is starting\n", constants.ServiceType, config.Port)

		httpServer := &http.Server{
			Handler: g,
		}
		err = httpServer.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	})
}

func (r *server) findNextFreeIndex() int32 {
	for i := 0; i < int(config.MaxConn); i++ {
		if r.connections[r.cindex] == nil {
			return r.cindex
		}
		r.cindex = (r.cindex + 1) % config.MaxConn
	}
	return -1
}

func (r *server) getFreeConnection() *connection {
	select {
	case conn, ok := <-r.connectionPool:
		if !ok {
			return nil
		}
		return conn
	case <-time.After(100 * time.Millisecond):
		// 池空且未超时，尝试新建
		if r.ccount < config.MaxConn {
			conn := newConnection()
			r.ccount++
			return conn
		}
		return nil
	}
}

func (r *server) freeConnection(conn *connection) {
	select {
	case r.connectionPool <- conn:
		// 归还成功
	default:
		// 池已满, 放弃这个连接
	}
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
			return
		}

		if config.IPBlackValidate != nil {
			if config.IPBlackValidate(c.ClientIP()) {
				logger.Log().Error("ip 黑名单 %s", c.ClientIP())
				conn.Close()
				return
			}
		}

		r.Lock()
		defer r.Unlock()

		if r.ccount >= config.MaxConn {
			logger.Log().Error("当前连接数量超过最大值,放弃新连接")
			conn.Close()
			return
		}

		var index = r.findNextFreeIndex()
		if index < 0 {
			logger.Log().Error("当前连接数量超过最大值,放弃新连接")
			conn.Close()
			return
		}

		var dealConn = r.getFreeConnection()
		if dealConn == nil {
			logger.Log().Error("新建connection失败")
			conn.Close()
			return
		}
		r.connections[index] = dealConn

		dealConn.init(r, conn, index+1)

		constants.Go(func() { dealConn.start() })
	} else {
		logger.Log().Error("不是websocket请求")
	}
}

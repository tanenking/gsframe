package kcpx

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/logger"
	"github.com/xtaci/kcp-go/v5"
)

type server struct {
	closed      atomic.Int32
	connections sync.Map //int32 => *connection
	idgenerate  int32
	ccount      int32

	// groupMsgSeq  uint16
	// groupMsgSync sync.Mutex
	// groupMsgList [common.MaxGroupMsgCount]*common.GroupMessage
}

func CreateServer(_config *gsinf.KcpServerConfig) gsinf.IKcpServer {
	config = _config
	validateConfig()
	_server := &server{
		// groupMsgList: [common.MaxGroupMsgCount]*common.GroupMessage{},
	}
	// _server.groupMsgList[_server.groupMsgSeq] = &common.GroupMessage{C: make(chan struct{})}
	_server.start()
	return _server
}

func (r *server) GetConnection(connId int32) gsinf.IKcpConnection {
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

// func (r *server) SendGroup(groupName string, header int64, msgID string, data []byte) {
// 	r.groupMsgSync.Lock()
// 	gmsg := r.groupMsgList[r.groupMsgSeq]
// 	r.groupMsgSeq++
// 	if r.groupMsgSeq >= common.MaxGroupMsgCount {
// 		r.groupMsgSeq = 0
// 	}
// 	r.groupMsgList[r.groupMsgSeq] = &common.GroupMessage{C: make(chan struct{})}
// 	r.groupMsgSync.Unlock()

// 	gmsg.GroupName = groupName
// 	gmsg.Header = header
// 	gmsg.MsgID = msgID
// 	if cap(gmsg.MsgData) < len(data) {
// 		gmsg.MsgData = make([]byte, 0, len(data))
// 	}
// 	gmsg.MsgData = gmsg.MsgData[:len(data)]
// 	copy(gmsg.MsgData, data)

// 	close(gmsg.C)
// }

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

func (r *server) start() {
	constants.Go(func() {
		defer constants.AutoRecover()()

		addr := fmt.Sprintf(":%d", config.Port)
		udpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			panic(err)
		}

		listener, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			panic(err)
		}
		defer listener.Close()

		config.Port = listener.LocalAddr().(*net.UDPAddr).Port

		logger.Log().Info("[START] Server kcp server name: %s,listenner at Port %d is starting\n", constants.ServiceType, config.Port)

		kcplistener, err := kcp.ServeConn(nil, 0, 0, listener)
		if err != nil {
			panic(err)
		}
		defer kcplistener.Close()

		for {
			conn, err := kcplistener.AcceptKCP()
			if err != nil {
				logger.Log().Error("kcp accept err: %s", err.Error())
				continue
			}
			r.accept(conn)
		}
	})
}
func (r *server) accept(conn *kcp.UDPSession) {
	if r.closed.Load() > 0 || constants.GetSystemStatus() == gsinf.SystemStatus_Maintain {
		//正在关闭中,不接收新连接了
		conn.Close()
		return
	}
	remoteAddress := conn.RemoteAddr().(*net.UDPAddr)
	if config.IPBlackValidate != nil {
		if config.IPBlackValidate(remoteAddress.IP.String()) {
			logger.Log().Error("ip 黑名单 %s", remoteAddress.IP.String())
			conn.Close()
			return
		}
	}

	var dealConn = r.getFreeConnection()
	if dealConn == nil {
		logger.Log().Error("kcp新建connection失败")
		conn.Close()
		return
	}
	atomic.AddInt32(&r.ccount, 1)
	connId := atomic.AddInt32(&r.idgenerate, 1)
	r.connections.Store(connId, dealConn)

	dealConn.init(r, conn, connId)

	constants.Go(func() { dealConn.start() })
}
